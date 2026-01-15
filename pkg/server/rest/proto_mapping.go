package rest

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/asjard/asjard/utils/cast"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// protoForm maps a map of string slices (e.g., URL queries or form data)
// to a Protobuf message using reflection.
func protoForm(ptr proto.Message, form map[string][]string) error {
	// Use Protobuf reflection to access the message's internal structure.
	msg := ptr.ProtoReflect()
	fields := msg.Descriptor().Fields()

	// Iterate through each field defined in the .proto file.
	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		kd := field.Kind()

		// Skip complex types like nested messages or groups.
		// These are typically handled via JSON body parsing rather than form/query params.
		if kd == protoreflect.MessageKind || kd == protoreflect.GroupKind {
			continue
		}

		// Use the field's text name (usually snake_case) as the lookup key in the form map.
		key := field.TextName()

		values, exists := form[key]
		if !exists {
			continue
		}

		switch {
		// Handle 'repeated' fields in Protobuf.
		case field.IsList():
			list := msg.Mutable(field).List()
			for _, s := range values {
				val, err := stringToValue(s, field)
				if err != nil {
					return fmt.Errorf("field %s: %w", key, err)
				}
				list.Append(val)
			}
		// Handle standard single-value fields.
		default:
			if len(values) > 0 {
				val, err := stringToValue(values[0], field)
				if err != nil {
					return fmt.Errorf("field %s: %w", key, err)
				}
				msg.Set(field, val)
			}
		}
	}
	return nil
}

// stringToValue converts a string input into a type-safe protoreflect.Value
// based on the field's underlying Protobuf kind.
func stringToValue(s string, field protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	switch field.Kind() {
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(s), nil

	case protoreflect.BytesKind:
		// Expects a Base64 encoded string for binary data.
		return strToProtoBytesValue(s)

	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return strToProtoInt32Value(s)

	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return strToProtoInt64Value(s)

	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		return strToProtoUint32Value(s)

	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return strToProtoUint64Value(s)

	case protoreflect.FloatKind:
		return strToProtoFloat32Value(s)

	case protoreflect.DoubleKind:
		// Map double (float64) to float32 value helper.
		return strToProtoFloat32Value(s)

	case protoreflect.BoolKind:
		return strToProtoBoolValue(s)

	case protoreflect.EnumKind:
		// Resolves enums by name or numeric value.
		return strToProtoEnumValue(s, field)

	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported type: %s", field.Kind())
	}
}

// strToProtoBytesValue handles base64 decoding for Protobuf 'bytes' fields.
func strToProtoBytesValue(s string) (protoreflect.Value, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("base64 decode error: %w", err)
	}
	return protoreflect.ValueOfBytes(data), nil
}

// Numeric conversion helpers ensure the string input fits the specific bit-size of the field.

func strToProtoInt32Value(s string) (protoreflect.Value, error) {
	v, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid int32: %w", err)
	}
	return protoreflect.ValueOfInt32(int32(v)), nil
}

func strToProtoUint32Value(s string) (protoreflect.Value, error) {
	v, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid uint32: %w", err)
	}
	return protoreflect.ValueOfUint32(uint32(v)), nil
}

func strToProtoInt64Value(s string) (protoreflect.Value, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid int64: %w", err)
	}
	return protoreflect.ValueOfInt64(v), nil
}

func strToProtoUint64Value(s string) (protoreflect.Value, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid uint64: %w", err)
	}
	return protoreflect.ValueOfUint64(v), nil
}

func strToProtoFloat32Value(s string) (protoreflect.Value, error) {
	v, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid float: %w", err)
	}
	return protoreflect.ValueOfFloat32(float32(v)), nil
}

func strToProtoFloat64Value(s string) (protoreflect.Value, error) {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid double: %w", err)
	}
	return protoreflect.ValueOfFloat64(v), nil
}

func strToProtoBoolValue(s string) (protoreflect.Value, error) {
	// Uses asjard/utils/cast to handle multiple boolean formats (1, true, t, etc).
	v, err := cast.ToBoolE(s)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid bool: %w", err)
	}
	return protoreflect.ValueOfBool(v), nil
}

// strToProtoEnumValue handles Enum resolution.
func strToProtoEnumValue(s string, field protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	// 1. Try to find the enum value by its string name (e.g., "ACTIVE").
	if enumVal := field.Enum().Values().ByName(protoreflect.Name(s)); enumVal != nil {
		return protoreflect.ValueOfEnum(enumVal.Number()), nil
	}
	// 2. Fallback: try to parse the string as a numeric enum ID (e.g., "1").
	if v, err := strconv.ParseInt(s, 10, 32); err == nil {
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(v)), nil
	}
	return protoreflect.Value{}, fmt.Errorf("invalid enum value: %s", s)
}
