package main

import (
	"strconv"
	"strings"

	"github.com/asjard/asjard/pkg/protobuf/validatepb"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func messageSchema(msg *protogen.Message, withRequired bool) string {
	if msg == nil {
		return `{"type":"object"}`
	}
	var b strings.Builder
	b.WriteString(`{"type":"object","properties":{`)
	var required []string
	for i, field := range msg.Fields {
		if i > 0 {
			b.WriteByte(',')
		}
		name := string(field.Desc.Name())
		b.WriteString(strconv.Quote(name))
		b.WriteByte(':')
		b.WriteString(fieldSchemaJSON(field))
		if withRequired && fieldRequired(field) {
			required = append(required, name)
		}
	}
	b.WriteByte('}')
	if withRequired && len(required) > 0 {
		b.WriteString(`,"required":[`)
		for i, name := range required {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(strconv.Quote(name))
		}
		b.WriteByte(']')
	}
	b.WriteByte('}')
	return b.String()
}

func fieldSchemaJSON(field *protogen.Field) string {
	typ := jsonSchemaType(field.Desc.Kind())
	var b strings.Builder
	b.WriteString(`{"type":`)
	switch {
	case field.Desc.IsMap():
		b.WriteString(`"object"`)
	case field.Desc.IsList():
		b.WriteString(`"array","items":{"type":`)
		b.WriteString(strconv.Quote(typ))
		b.WriteByte('}')
	default:
		b.WriteString(strconv.Quote(typ))
	}
	if desc := fieldDescription(field); desc != "" {
		b.WriteString(`,"description":`)
		b.WriteString(strconv.Quote(desc))
	}
	b.WriteByte('}')
	return b.String()
}

func fieldDescription(field *protogen.Field) string {
	return joinCommentAndRules(fieldComment(field), validateRulesText(field))
}

func joinCommentAndRules(comment, rules string) string {
	switch {
	case comment != "" && rules != "":
		return comment + " [" + rules + "]"
	case comment != "":
		return comment
	default:
		return rules
	}
}

func fieldComment(field *protogen.Field) string {
	return strings.Join(commentLines(field.Comments), " ")
}

func commentLines(commentSet protogen.CommentSet) []string {
	commentStr := ""
	if commentSet.Leading != "" {
		commentStr += strings.TrimSpace(commentSet.Leading.String())
	}
	if commentSet.Trailing != "" {
		if commentStr != "" {
			commentStr += "\n"
		}
		commentStr += strings.TrimSpace(commentSet.Trailing.String())
	}
	if commentStr == "" {
		return nil
	}
	raw := strings.Split(strings.TrimSuffix(commentStr, "\n"), "\n")
	lines := make([]string, 0, len(raw))
	for _, line := range raw {
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "//"))
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

func validateRulesText(field *protogen.Field) string {
	v := fieldValidate(field)
	if v == nil || len(v.GetRules()) == 0 {
		return ""
	}
	return strings.Join(v.GetRules(), ";")
}

func fieldRequired(field *protogen.Field) bool {
	v := fieldValidate(field)
	if v == nil {
		return false
	}
	for _, rule := range v.GetRules() {
		if ruleHasRequired(rule) {
			return true
		}
	}
	return false
}

func fieldValidate(field *protogen.Field) *validatepb.Validate {
	v, ok := proto.GetExtension(field.Desc.Options(), validatepb.E_Validate).(*validatepb.Validate)
	if !ok {
		return nil
	}
	return v
}

func ruleHasRequired(rule string) bool {
	for _, part := range strings.Split(rule, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if i := strings.Index(part, ":"); i >= 0 && !strings.Contains(part[:i], "=") {
			part = part[i+1:]
		}
		for _, tag := range strings.Split(part, ",") {
			name := strings.TrimSpace(tag)
			if eq := strings.Index(name, "="); eq >= 0 {
				name = name[:eq]
			}
			if name == "required" {
				return true
			}
		}
	}
	return false
}

func jsonSchemaType(kind protoreflect.Kind) string {
	switch kind {
	case protoreflect.BoolKind:
		return "boolean"
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint64Kind, protoreflect.Fixed64Kind,
		protoreflect.FloatKind, protoreflect.DoubleKind:
		return "number"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return "object"
	default:
		return "string"
	}
}
