/*
 *
 * Copyright 2020 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"fmt"
	"strings"

	"github.com/asjard/asjard/pkg/protobuf/httppb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	contextPackage = protogen.GoImportPath("context")
	restPackage    = protogen.GoImportPath("github.com/asjard/asjard/pkg/server/rest")
	serverPackage  = protogen.GoImportPath("github.com/asjard/asjard/core/server")
	// restPackage    = protogen.GoImportPath("google.golang.org/grpc")
	// codesPackage   = protogen.GoImportPath("google.golang.org/grpc/codes")
	// statusPackage  = protogen.GoImportPath("google.golang.org/grpc/status")
)

type serviceGenerateHelperInterface interface {
	formatFullMethodSymbol(service *protogen.Service, method *protogen.Method) string
	genFullMethods(g *protogen.GeneratedFile, service *protogen.Service)
	generateClientStruct(g *protogen.GeneratedFile, clientName string)
	generateNewClientDefinitions(g *protogen.GeneratedFile, service *protogen.Service, clientName string)
	generateUnimplementedServerType(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service)
	generateServerFunctions(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, serverType string, serviceDescVar string)
	formatHandlerFuncName(service *protogen.Service, hname string) string
}

type serviceGenerateHelper struct{}

func (serviceGenerateHelper) formatFullMethodSymbol(service *protogen.Service, method *protogen.Method) string {
	return fmt.Sprintf("%s_%s_FullMethodName", service.GoName, method.GoName)
}

func (serviceGenerateHelper) genFullMethods(g *protogen.GeneratedFile, service *protogen.Service) {
	if len(service.Methods) == 0 {
		return
	}
	g.P()
}

func (serviceGenerateHelper) generateClientStruct(g *protogen.GeneratedFile, clientName string) {
	g.P("type ", unexport(clientName), " struct {")
	// g.P("cc ", restPackage.Ident("ClientConnInterface"))
	g.P("}")
	g.P()
}

func (serviceGenerateHelper) generateNewClientDefinitions(g *protogen.GeneratedFile, service *protogen.Service, clientName string) {
	g.P("return &", unexport(clientName), "{cc}")
}

func (serviceGenerateHelper) generateUnimplementedServerType(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
}

func (serviceGenerateHelper) generateServerFunctions(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service, serverType string, serviceDescVar string) {
	// Server handler implementations.
	handlerNames := make([]string, 0, len(service.Methods))
	for _, method := range service.Methods {
		if httpOptions, ok := proto.GetExtension(method.Desc.Options(), httppb.E_Http).([]*httppb.Http); ok && len(httpOptions) != 0 {
			hname := genServerMethod(gen, file, g, method, serverType, func(hname string) string {
				return hname
			})
			handlerNames = append(handlerNames, hname)
		} else {
			handlerNames = append(handlerNames, "")
		}
	}
	// genServiceDesc(file, g, serviceDescVar, serverType, service, handlerNames)
}

func (serviceGenerateHelper) formatHandlerFuncName(service *protogen.Service, hname string) string {
	return hname
}

var helper serviceGenerateHelperInterface = serviceGenerateHelper{}

// FileDescriptorProto.package field number
const fileDescriptorProtoPackageFieldNumber = 2

// FileDescriptorProto.syntax field number
const fileDescriptorProtoSyntaxFieldNumber = 12

// generateFile generates a _grpc.pb.go file containing gRPC service definitions.
func generateFile(gen *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + ".d.ts"
	g := gen.NewGeneratedFile(filename, file.GoImportPath)
	// Attach all comments associated with the syntax field.
	genLeadingComments(g, file.Desc.SourceLocations().ByPath(protoreflect.SourcePath{fileDescriptorProtoSyntaxFieldNumber}))
	g.P("// Code generated by protoc-gen-ts. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-ts v", version)
	g.P("// - protoc             ", protocVersion(gen))
	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	// Attach all comments associated with the package field.
	genLeadingComments(g, file.Desc.SourceLocations().ByPath(protoreflect.SourcePath{fileDescriptorProtoPackageFieldNumber}))
	// g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(gen, file, g)
	return g
}

func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

// generateFileContent generates the gRPC service definitions, excluding the package statement.
func generateFileContent(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}
	g.P()
	for _, message := range file.Messages {
		genMessage(g, message)
	}
	// for _, service := range file.Services {
	// 	genService(gen, file, g, service)
	// }
}

func genMessage(g *protogen.GeneratedFile, message *protogen.Message) {
	genComments(g, message.Comments)
	g.P("// -----------")
	g.P("export interface ", message.GoIdent, "{")
	for _, field := range message.Fields {
		g.P("//", field.Desc.Name(), ":", field.Desc.Kind(), field.Desc.)
		g.P("    ", field.Desc.Name(), ": ", tsKindString(field.Desc.Kind()), ";")

	}
	g.P("}")
}

func tsKindString(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.StringKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Uint32Kind,
		protoreflect.Int64Kind, protoreflect.Uint64Kind:
		return "number"
	case protoreflect.MessageKind:
		return kind.String()
	}
	return ""
}

func genComments(g *protogen.GeneratedFile, comments protogen.CommentSet) {
	if comments.Leading != "" {
		// Add empty comment line to attach this service's comments to
		// the godoc comments previously output for all services.
		g.P("//")
		g.P(strings.TrimSpace(comments.Leading.String()))
	}
}

func genService(gen *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	// Full methods constants.
	helper.genFullMethods(g, service)

	serverType := service.GoName + "Server"
	serviceDescVar := service.GoName + "RestServiceDesc"
	helper.generateServerFunctions(gen, file, g, service, serverType, serviceDescVar)
}

func genServerMethod(_ *protogen.Plugin, _ *protogen.File, g *protogen.GeneratedFile, method *protogen.Method, serverType string, hnameFuncNameFormatter func(string) string) string {
	service := method.Parent
	hname := fmt.Sprintf("_%s_%s_RestHandler", service.GoName, method.GoName)

	g.P("func ", hnameFuncNameFormatter(hname), "(ctx *", restPackage.Ident("Context"), ", srv any, interceptor ", serverPackage.Ident("UnaryServerInterceptor"), ") (any, error) {")
	g.P("in := new(", method.Input.GoIdent, ")")
	g.P("if interceptor == nil {")
	g.P("return srv.(", serverType, ").", method.GoName, "(ctx, in)")
	g.P("}")
	g.P("info := &", serverPackage.Ident("UnaryServerInfo"), "{")
	g.P("Server: srv,")
	g.P("FullMethod: \"", service.Desc.FullName(), ".", method.Desc.Name(), "\",")
	g.P("Protocol: ", restPackage.Ident("Protocol"), ",")
	g.P("}")
	g.P("handler := func(ctx ", contextPackage.Ident("Context"), ",req any)(any, error) {")
	g.P("return srv.(", serverType, ").", method.GoName, "(ctx, in)")
	g.P("}")
	g.P("return interceptor(ctx, in, info, handler)")
	g.P("}")
	return hname
}

func genLeadingComments(g *protogen.GeneratedFile, loc protoreflect.SourceLocation) {
	for _, s := range loc.LeadingDetachedComments {
		g.P(protogen.Comments(s))
		g.P()
	}
	if s := loc.LeadingComments; s != "" {
		g.P(protogen.Comments(s))
		g.P()
	}
}

const deprecationComment = "// Deprecated: Do not use."

func unexport(s string) string { return strings.ToLower(s[:1]) + s[1:] }
