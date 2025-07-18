/*
 *
 * Copyright 2024 ASJARD authors.
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

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// FileDescriptorProto.package field number
	fileDescriptorProtoPackageFieldNumber = 2
	// FileDescriptorProto.syntax field number
	fileDescriptorProtoSyntaxFieldNumber = 12
)

type TsGenerator struct {
	plugin     *protogen.Plugin
	gen        *protogen.GeneratedFile
	file       *protogen.File
	openapiVar string
}

func NewGwGenerator(plugin *protogen.Plugin, file *protogen.File) *TsGenerator {
	return &TsGenerator{
		plugin: plugin,
		file:   file,
	}
}

func (g *TsGenerator) Run() *protogen.GeneratedFile {
	if len(g.file.Enums) == 0 {
		return nil
	}
	filenamePrefix := strings.Join(strings.Split(g.file.GeneratedFilenamePrefix, "/")[1:], "/")
	g.gen = g.plugin.NewGeneratedFile(filenamePrefix+".enum.pb.ts", g.file.GoImportPath)

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

	g.gen.P("// @ts-ignore")
	g.gen.P("/* eslint-disable */")
	g.gen.P()

	// Attach all comments associated with the package field.
	g.genLeadingComments(g.file.Desc.SourceLocations().ByPath(protoreflect.SourcePath{fileDescriptorProtoPackageFieldNumber}))

	g.genFileContent()
	return g.gen
}

func (g *TsGenerator) genFileContent() {
	// msgLen := len(g.file.Messages)
	enumLen := len(g.file.Enums)
	if enumLen != 0 {
		for _, em := range g.file.Enums {
			g.genEnum(em)
		}
	}
	// if msgLen != 0 || enumLen != 0 {
	// 	g.gen.P("declare namespace ", g.file.Desc.FullName(), " {")
	// }
	// for _, em := range g.file.Enums {
	// 	g.genEnum(em, false)
	// }
	// for _, message := range g.file.Messages {
	// 	g.genMessage(message)
	// }
	// if msgLen != 0 || enumLen != 0 {
	// 	g.gen.P("}")
	// for _, em := range g.file.Enums {
	// 	g.genEnum(em, true)
	// }
	// }
}

func (g *TsGenerator) genService(service *protogen.Service) {
	for _, method := range service.Methods {
		g.genServiceMethod(service, method)
	}
}

func (g *TsGenerator) genServiceMethod(service *protogen.Service, method *protogen.Method) {
}

func (g *TsGenerator) genLeadingComments(loc protoreflect.SourceLocation) {
	for _, s := range loc.LeadingDetachedComments {
		g.gen.P(protogen.Comments(s))
		g.gen.P()
	}
	if s := loc.LeadingComments; s != "" {
		g.gen.P(protogen.Comments(s))
		g.gen.P()
	}
}

func (g *TsGenerator) genComment(comments protogen.CommentSet) {
	if comments.Leading != "" {
		g.gen.P(strings.TrimSpace(comments.Leading.String()))
	}
	if comments.Trailing != "" {
		g.gen.P(strings.TrimSpace(comments.Trailing.String()))
	}
}

func (g *TsGenerator) protocVersion() string {
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

func (g *TsGenerator) genMessage(message *protogen.Message) {
	g.genComment(message.Comments)
	g.gen.P("  type ", message.GoIdent, " = {")
	for _, field := range message.Fields {
		g.genComment(field.Comments)
		if field.Desc.IsList() {
			g.gen.P(field.Desc.Name(), "?: ", g.tsKindString(field), "[];")
		} else {
			g.gen.P(field.Desc.Name(), "?: ", g.tsKindString(field), ";")
		}
	}
	g.gen.P("  };")
	g.gen.P()
}

func (g *TsGenerator) genEnum(em *protogen.Enum) {
	g.genComment(em.Comments)
	g.gen.P("export  enum ", em.GoIdent, " {")
	for _, value := range em.Values {
		g.genComment(value.Comments)
		g.gen.P("    ", value.Desc.Name(), " = '", value.Desc.Name(), "',")
	}
	g.gen.P("  }")
	g.gen.P("")

	g.genComment(em.Comments)
	g.gen.P("export const ValueEnum", em.GoIdent, "={")
	for idx, value := range em.Values {
		if idx == 0 {
			continue
		}
		g.genComment(value.Comments)
		g.gen.P(value.Desc.Name(), ":{")
		name := strings.Split(string(value.Desc.Name()), "_")
		if len(name) > 1 {
			g.gen.P("text:'", strings.Join(name[1:], "_"), "',")
		} else {
			g.gen.P("text:'", name, "',")
		}
		g.gen.P("},")
	}
	g.gen.P("}")
}

func (g *TsGenerator) tsKindString(field *protogen.Field) string {
	kind := field.Desc.Kind()
	switch kind {
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.StringKind, protoreflect.Int64Kind, protoreflect.DoubleKind:
		return "string"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Uint32Kind,
		protoreflect.Fixed32Kind, protoreflect.Sfixed32Kind, protoreflect.FloatKind:
		return "number"
	case protoreflect.EnumKind:
		return string(field.Enum.Desc.FullName())
	case protoreflect.MessageKind:
		if field.Message.Desc.IsMapEntry() {
			return "{}"
		} else {
			return string(field.Message.Desc.FullName())
		}
	}
	return ""
}
