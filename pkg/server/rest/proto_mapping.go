package rest

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/asjard/asjard/utils/cast"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func protoForm(ptr proto.Message, form map[string][]string) error {
	msg := ptr.ProtoReflect()
	fields := msg.Descriptor().Fields()

	for i := 0; i < fields.Len(); i++ {
		field := fields.Get(i)
		kd := field.Kind()
		// 只解析基础类型
		if kd == protoreflect.MessageKind || kd == protoreflect.GroupKind {
			continue
		}

		key := field.TextName()

		values, exists := form[key]
		if !exists {
			continue
		}

		switch {
		case field.IsList(): // 重复字段
			list := msg.Mutable(field).List()
			for _, s := range values {
				val, err := stringToValue(s, field)
				if err != nil {
					return fmt.Errorf("field %s: %w", key, err)
				}
				list.Append(val)
			}
		default: // 单值字段
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

func stringToValue(s string, field protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	switch field.Kind() {
	case protoreflect.StringKind:
		return protoreflect.ValueOfString(s), nil

	case protoreflect.BytesKind:
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
		return strToProtoFloat32Value(s)

	case protoreflect.BoolKind:
		return strToProtoBoolValue(s)

	case protoreflect.EnumKind:
		return strToProtoEnumValue(s, field)

	default:
		return protoreflect.Value{}, fmt.Errorf("unsupported type: %s", field.Kind())
	}
}

func strToProtoBytesValue(s string) (protoreflect.Value, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("base64 decode error: %w", err)
	}
	return protoreflect.ValueOfBytes(data), nil
}
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
	v, err := cast.ToBoolE(s)
	if err != nil {
		return protoreflect.Value{}, fmt.Errorf("invalid bool: %w", err)
	}
	return protoreflect.ValueOfBool(v), nil
}

func strToProtoEnumValue(s string, field protoreflect.FieldDescriptor) (protoreflect.Value, error) {
	// 尝试按名称解析
	if enumVal := field.Enum().Values().ByName(protoreflect.Name(s)); enumVal != nil {
		return protoreflect.ValueOfEnum(enumVal.Number()), nil
	}
	// 尝试按数值解析
	if v, err := strconv.ParseInt(s, 10, 32); err == nil {
		return protoreflect.ValueOfEnum(protoreflect.EnumNumber(v)), nil
	}
	return protoreflect.Value{}, fmt.Errorf("invalid enum value: %s", s)
}
