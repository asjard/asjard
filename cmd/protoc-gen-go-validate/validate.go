package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/asjard/asjard/pkg/protobuf/validatepb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// FileDescriptorProto.package field number
	fileDescriptorProtoPackageFieldNumber = 2
	// FileDescriptorProto.syntax field number
	fileDescriptorProtoSyntaxFieldNumber = 12
)

const (
	validatorPackage        = protogen.GoImportPath("github.com/go-playground/validator")
	defaultValidatorPackage = protogen.GoImportPath("github.com/asjard/asjard/pkg/protobuf/validatepb")
	statusPackage           = protogen.GoImportPath("github.com/asjard/asjard/core/status")
	codesPackage            = protogen.GoImportPath("google.golang.org/grpc/codes")
)

type ValidateGenerator struct {
	plugin *protogen.Plugin
	file   *protogen.File
	gen    *protogen.GeneratedFile
}

func NewValidateGeneratr(plugin *protogen.Plugin, file *protogen.File) *ValidateGenerator {
	return &ValidateGenerator{
		plugin: plugin,
		file:   file,
	}
}

func (g *ValidateGenerator) Run() *protogen.GeneratedFile {
	if len(g.file.Messages) == 0 {
		return nil
	}
	g.gen = g.plugin.NewGeneratedFile(g.file.GeneratedFilenamePrefix+"_validate.pb.go", g.file.GoImportPath)

	g.genLeadingComments(g.file.Desc.SourceLocations().ByPath(protoreflect.SourcePath{fileDescriptorProtoSyntaxFieldNumber}))
	g.gen.P("// Code generated by ", name, ". DO NOT EDIT.")
	g.gen.P("// versions:")
	g.gen.P("// - ", name, " v", version)
	g.gen.P("// - protoc             ", g.protocVersion())
	if g.file.Proto.GetOptions().GetDeprecated() {
		g.gen.P("// ", g.file.Desc.Path(), " is a deprecated file.")
	} else {
		g.gen.P("// source: ", g.file.Desc.Path())
	}
	g.gen.P()

	// Attach all comments associated with the package field.
	g.genLeadingComments(g.file.Desc.SourceLocations().ByPath(protoreflect.SourcePath{fileDescriptorProtoPackageFieldNumber}))
	g.gen.P("package ", g.file.GoPackageName)
	g.gen.P()

	g.genFileContent()
	return g.gen
}

func (g *ValidateGenerator) genFileContent() {
	for _, message := range g.file.Messages {
		g.genMessage(message)
	}
}

func (g *ValidateGenerator) genMessageMessages(messages []*protogen.Message) {
	for _, message := range messages {
		if !message.Desc.IsMapEntry() {
			g.genMessage(message)
		}
	}
}

//gocyclo:ignore
func (g *ValidateGenerator) genMessage(message *protogen.Message) {
	g.genMessageMessages(message.Messages)
	g.genComment(message.Comments)
	g.gen.P("func (m *", message.GoIdent.GoName, ")IsValid(parentFieldName,fullMethod string ) error{")
	inited := false
	for _, field := range message.Fields {
		switch field.Desc.Kind() {
		case protoreflect.MessageKind:
			if field.Desc.IsList() {
				g.gen.P("for _, fm := range m.", field.GoName, "{")
				g.gen.P("if err := fm.IsValid(", defaultValidatorPackage.Ident("ValidateFieldName"), "(parentFieldName, \"", field.Desc.JSONName(), "\")", ", fullMethod); err != nil {")
				g.gen.P("return err")
				g.gen.P("}")
				g.gen.P("}")

			} else if !field.Desc.IsMap() && field.Oneof == nil {
				inited = g.genMessageValid(field, inited)
				g.gen.P("if m.", field.GoName, "!= nil {")
				g.gen.P("if err := m.", field.GoName, ".IsValid(", defaultValidatorPackage.Ident("ValidateFieldName"), "(parentFieldName, \"", field.Desc.JSONName(), "\")", ",fullMethod); err != nil {")
				g.gen.P("return err")
				g.gen.P("}")
				g.gen.P("}")
			} else if field.Oneof != nil {
				g.gen.P("if vl, ok := m.Get", field.Oneof.GoName, "().(*", message.GoIdent.GoName, "_", field.GoName, "); ok {")
				g.gen.P("if err := vl.", field.GoName, ".IsValid(", defaultValidatorPackage.Ident("ValidateFieldName"), "(parentFieldName, \"", field.Desc.JSONName(), "\")", ",fullMethod); err != nil {")
				g.gen.P("return err")
				g.gen.P("}")
				g.gen.P("}")
			}
		case protoreflect.GroupKind:
			g.gen.P("//--group kind--", field.GoName)
		case protoreflect.EnumKind:
			if !field.Desc.IsList() {
				g.gen.P("if _, ok := ", field.Enum.GoIdent, "_name[int32(m.", field.GoName, ")]; !ok {")
				g.gen.P("return ", statusPackage.Ident("Errorf"), "(", codesPackage.Ident("InvalidArgument"), `,"`, "invalid %s", `",`, defaultValidatorPackage.Ident("ValidateFieldName"), "(parentFieldName, \"", field.Desc.JSONName(), "\")", `)`)
				g.gen.P("}")
			}
		default:
			inited = g.genMessageValid(field, inited)
		}
	}
	g.gen.P("return nil")
	g.gen.P("}")
	g.gen.P("")
}

func (g ValidateGenerator) genMessageValid(field *protogen.Field, inited bool) bool {
	if validateRule, ok := proto.GetExtension(field.Desc.Options(), validatepb.E_Validate).(*validatepb.Validate); ok && validateRule != nil {
		methodRules := make(map[string][]string)
		var globalRules []string
		for _, rule := range validateRule.Rules {
			rules := strings.Split(rule, ";")
			if len(rules) != 0 {
				for _, rule := range rules {
					methodAndRule := strings.Split(rule, ":")
					if len(methodAndRule) == 2 {
						methodRules[methodAndRule[0]] = append(methodRules[methodAndRule[0]], methodAndRule[1])
					} else {
						globalRules = append(globalRules, rule)
					}
				}
			}
		}
		if !inited && (len(globalRules) != 0 || len(methodRules) != 0) {
			g.gen.P("v := ", defaultValidatorPackage.Ident("DefaultValidator"))
			inited = true
		}
		if len(globalRules) != 0 {
			g.genFieldValid(field, strings.Join(globalRules, ","), validateRule)
		}
		for method, rules := range methodRules {
			g.gen.P("if fullMethod != \"\" && fullMethod == ", strconv.Quote(method), "{")
			g.genFieldValid(field, strings.Join(rules, ","), validateRule)
			g.gen.P("}")
		}
	}
	return inited
}

func (g *ValidateGenerator) genFieldValid(field *protogen.Field, rule string, validate *validatepb.Validate) {
	g.gen.P("if err := v.Var(m.", field.GoName, ",", strconv.Quote(rule), "); err != nil {")
	errMsg := fmt.Sprintf("validation field '%%s' on '%s' fail", rule)
	if validate.ErrCode != 0 {
		g.gen.P("return ", statusPackage.Ident("Errorf"), "(", validate.ErrCode, ",", strconv.Quote(errMsg), ",", defaultValidatorPackage.Ident("ValidateFieldName"), "(parentFieldName, \"", field.Desc.JSONName(), "\")", ")")
	} else {
		g.gen.P("return ", statusPackage.Ident("Errorf"), "(", codesPackage.Ident("InvalidArgument"), ",", strconv.Quote(errMsg), ",", defaultValidatorPackage.Ident("ValidateFieldName"), "(parentFieldName, \"", field.Desc.JSONName(), "\")", ")")
	}
	g.gen.P("}")
}

func (g *ValidateGenerator) genLeadingComments(loc protoreflect.SourceLocation) {
	for _, s := range loc.LeadingDetachedComments {
		g.gen.P(protogen.Comments(s))
		g.gen.P()
	}
	if s := loc.LeadingComments; s != "" {
		g.gen.P(protogen.Comments(s))
		g.gen.P()
	}
}

func (g *ValidateGenerator) genComment(comments protogen.CommentSet) {
	if comments.Leading != "" {
		g.gen.P("// IsValid Params validate")
		g.gen.P(strings.TrimSpace(comments.Leading.String()))
	}
}

func (g *ValidateGenerator) protocVersion() string {
	v := g.plugin.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}
