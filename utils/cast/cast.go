package cast

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// ToBoolE casts an interface to a bool type.
//
//gocyclo:ignore
func ToBoolE(i interface{}) (bool, error) {
	i = indirect(i)

	switch b := i.(type) {
	case bool:
		return b, nil
	case nil:
		return false, nil
	case int:
		return b != 0, nil
	case int64:
		return b != 0, nil
	case int32:
		return b != 0, nil
	case int16:
		return b != 0, nil
	case int8:
		return b != 0, nil
	case uint:
		return b != 0, nil
	case uint64:
		return b != 0, nil
	case uint32:
		return b != 0, nil
	case uint16:
		return b != 0, nil
	case uint8:
		return b != 0, nil
	case float64:
		return b != 0, nil
	case float32:
		return b != 0, nil
	case time.Duration:
		return b != 0, nil
	case string:
		return strToBool(i.(string))
	case json.Number:
		v, err := cast.ToInt64E(b)
		if err == nil {
			return v != 0, nil
		}
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", i, i)
	default:
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", i, i)
	}
}

// ToStringSliceE casts an interface to a []string type.
//
//gocyclo:ignore
func ToStringSliceE(i interface{}, delimiter string) ([]string, error) {
	var a []string

	switch v := i.(type) {
	case []interface{}:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case []string:
		return v, nil
	case []int8:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case []int:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case []int32:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case []int64:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case []float32:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case []float64:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case string:
		return strToStrs(v, delimiter), nil
	case []error:
		for _, err := range i.([]error) {
			a = append(a, err.Error())
		}
		return a, nil
	case interface{}:
		str, err := cast.ToStringE(v)
		if err != nil {
			return a, fmt.Errorf("unable to cast %#v of type %T to []string", i, i)
		}
		return strToStrs(str, delimiter), nil
	default:
		return a, fmt.Errorf("unable to cast %#v of type %T to []string", i, i)
	}
}

// 字符串转列表, 如果分隔符为空则以空格分隔
// 分隔符不为空按照分隔符分隔，并移除空格
func strToStrs(str, delimiter string) []string {
	if strings.TrimSpace(delimiter) == "" {
		return strings.Fields(str)
	}
	var noSpaceStrs []string
	for _, v := range strings.Split(str, delimiter) {
		noSpaceStr := strings.TrimSpace(v)
		if noSpaceStr != "" {
			noSpaceStrs = append(noSpaceStrs, noSpaceStr)
		}
	}
	return noSpaceStrs
}

// 非false列表中的均为true
func strToBool(str string) (bool, error) {
	lower := strings.ToLower(str)
	switch lower {
	case "0", "f", "false", "n", "no", "off":
		return false, nil
	}
	return true, nil
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
