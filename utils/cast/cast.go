package cast

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cast"
)

// ToBoolE attempts to cast an interface to a boolean.
// It handles pointers automatically via indirect() and treats non-zero numbers
// and specific truthy strings as true.
//
//gocyclo:ignore
func ToBoolE(i interface{}) (bool, error) {
	// Dereference pointers to get the underlying value.
	i = indirect(i)

	switch b := i.(type) {
	case bool:
		return b, nil
	case nil:
		return false, nil
	// Integer types: 0 is false, everything else is true.
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
	// Float types: 0.0 is false.
	case float64:
		return b != 0, nil
	case float32:
		return b != 0, nil
	// Duration: 0 is false.
	case time.Duration:
		return b != 0, nil
	// String: Checked against a list of "false" keywords.
	case string:
		return strToBool(i.(string))
	// json.Number: Common when decoding dynamic JSON.
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

// ToStringSliceE converts an interface to a slice of strings.
// It supports converting slices of various types (int, float, error) into strings,
// and can split a single string into a slice based on a provided delimiter.
//
//gocyclo:ignore
func ToStringSliceE(i interface{}, delimiter string) ([]string, error) {
	var a []string

	switch v := i.(type) {
	// Handle generic slices by converting each element individually.
	case []interface{}:
		for _, u := range v {
			a = append(a, cast.ToString(u))
		}
		return a, nil
	case []string:
		return v, nil
	// Numeric slices: each number becomes a string.
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
	// Single string: split by delimiter.
	case string:
		return strToStrs(v, delimiter), nil
	// Error slice: extract message from each error.
	case []error:
		for _, err := range i.([]error) {
			a = append(a, err.Error())
		}
		return a, nil
	// Fallback for single values: convert to string, then split.
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

// strToStrs splits a string into a slice.
// If delimiter is empty, it splits by whitespace (strings.Fields).
// If delimiter is provided, it splits and trims whitespace from resulting elements.
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

// strToBool defines "falsy" strings.
// Anything not in this list is considered true (if it's a string).
func strToBool(str string) (bool, error) {
	lower := strings.ToLower(str)
	switch lower {
	case "0", "f", "false", "n", "no", "off":
		return false, nil
	}
	return true, nil
}

// indirect uses reflection to follow pointers to their base value.
// It will recursively dereference (e.g., ***int -> int) until it hits
// a non-pointer type or a nil pointer.
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}
