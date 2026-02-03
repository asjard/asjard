package rest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// mapForm maps a map of strings (form/query) to a pointer to a struct.
// It uses reflection to match map keys with struct field tags (JSON) or field names.
//
//gocyclo:ignore
func mapForm(ptr any, form map[string][]string) error {
	if len(form) == 0 {
		return nil
	}

	// Get the type and value of the struct being pointed to.
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)

		// Skip fields that cannot be modified.
		if !structField.CanSet() {
			continue
		}

		structFieldKind := structField.Kind()

		// 1. Determine the input field name to look for in the map.
		// Priority: `json` tag value -> Struct Field Name.
		inputFieldName := typeField.Tag.Get("json")
		if inputFieldName != "" {
			fieldNameList := strings.Split(inputFieldName, ",")
			if len(fieldNameList) > 0 {
				inputFieldName = strings.TrimSpace(fieldNameList[0])
			}
		}

		if inputFieldName == "" {
			inputFieldName = typeField.Name

			// Recursively handle nested structs if the current field is a struct
			// and no specific tag was found. Useful for flattened form data.
			if structFieldKind == reflect.Struct {
				err := mapForm(structField.Addr().Interface(), form)
				if err != nil {
					return err
				}
				continue
			}
		}

		// 2. Extract value from the form map.
		inputValue, exists := form[inputFieldName]
		if !exists {
			continue
		}

		numElems := len(inputValue)

		// 3. Handle Slices (multiple values for the same key).
		if structFieldKind == reflect.Slice && numElems > 0 {
			sliceOf := structField.Type().Elem().Kind()
			slice := reflect.MakeSlice(structField.Type(), numElems, numElems)
			for i := 0; i < numElems; i++ {
				if err := setWithProperType(sliceOf, inputValue[i], slice.Index(i)); err != nil {
					return err
				}
			}
			val.Field(i).Set(slice)
		} else {
			// 4. Handle Special Types: time.Time.
			if _, isTime := structField.Interface().(time.Time); isTime {
				if err := setTimeField(inputValue[0], typeField, structField); err != nil {
					return err
				}
				continue
			}

			// 5. Handle Pointers.
			if typeField.Type.Kind() == reflect.Ptr {
				instance := reflect.New(typeField.Type.Elem())
				ptr := instance.Interface()
				// Hack for string pointers to ensure valid JSON unmarshaling if quotes are missing.
				if typeField.Type == reflect.TypeOf((*string)(nil)) && !strings.HasPrefix(inputValue[0], `"`) {
					inputValue[0] = fmt.Sprintf(`"%s"`, inputValue[0])
				}
				if err := json.Unmarshal([]byte(inputValue[0]), ptr); err != nil {
					return err
				}
				structField.Set(reflect.ValueOf(ptr))
				continue
			}

			// 6. Handle Primitive Types (int, string, bool, etc.).
			if err := setWithProperType(typeField.Type.Kind(), inputValue[0], structField); err != nil {
				return err
			}
		}
	}
	return nil
}

// setWithProperType dispatches the string value to the correct setter based on the target Kind.
func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return fmt.Errorf("Unknown type '%s'", valueKind.String())
	}
	return nil
}

// Specific setter functions handle string parsing with bit-size awareness.

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return nil
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

// setTimeField parses time strings using struct tags: `time_format`, `time_utc`, and `time_location`.
func setTimeField(val string, structField reflect.StructField, value reflect.Value) error {
	timeFormat := structField.Tag.Get("time_format")
	if timeFormat == "" {
		timeFormat = time.RFC3339
	}

	if val == "" {
		value.Set(reflect.ValueOf(time.Time{}))
		return nil
	}

	l := time.Local
	if isUTC, _ := strconv.ParseBool(structField.Tag.Get("time_utc")); isUTC {
		l = time.UTC
	}

	if locTag := structField.Tag.Get("time_location"); locTag != "" {
		loc, err := time.LoadLocation(locTag)
		if err != nil {
			return err
		}
		l = loc
	}

	t, err := time.ParseInLocation(timeFormat, val, l)
	if err != nil {
		return err
	}

	value.Set(reflect.ValueOf(t))
	return nil
}
