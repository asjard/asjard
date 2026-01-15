package utils

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/spf13/cast"
)

/*
JSONDuration wraps time.Duration to support flexible JSON unmarshaling.
It allows durations to be defined in JSON as either:
1. A number (nanoseconds, e.g., 1000000000)
2. A string (e.g., "1h30m", "300ms")
*/
type JSONDuration struct {
	time.Duration
}

// MarshalJSON returns the string representation of the duration (e.g., "1m0s").
func (d JSONDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

// UnmarshalJSON handles both numeric and string input for time.Duration.
func (d *JSONDuration) UnmarshalJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		// Treats numeric input as nanoseconds.
		d.Duration = time.Duration(value)
		return nil
	case string:
		// Parses human-readable strings like "10s" or "5m".
		var err error
		d.Duration, err = time.ParseDuration(value)
		return err
	default:
		return errors.New("invalid duration")
	}
}

/*
JSONStrings is a versatile string slice that supports being unmarshaled from
either a native JSON array ["a", "b"] or a comma-separated string "a,b".
*/
type JSONStrings []string

// MarshalJSON converts the slice back into a single comma-separated string wrapped in quotes.
func (s JSONStrings) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.Join(s, ",") + `"`), nil
}

// UnmarshalJSON detects if the input is a raw string or a JSON array and parses accordingly.
func (s *JSONStrings) UnmarshalJSON(b []byte) error {
	n := len(b)
	if n <= 2 {
		return nil
	}
	// Case 1: Input is a string (e.g., "item1,item2")
	if b[0] == '"' {
		*s = strings.Split(string(b[1:n-1]), ",")
		return nil
	} else if b[0] == '[' {
		// Case 2: Input is a standard JSON array (e.g., ["item1", "item2"])
		var out []any
		if err := json.Unmarshal(b, &out); err != nil {
			return err
		}
		result := make([]string, 0, len(out))
		for _, v := range out {
			// Uses cast.ToString to handle mixed types in the array gracefully.
			result = append(result, cast.ToString(v))
		}
		*s = result
		return nil
	}
	return errors.New("invliad strings")
}

// Contains checks if a specific string exists within the slice.
func (s JSONStrings) Contains(subStr string) bool {
	for _, item := range s {
		if item == subStr {
			return true
		}
	}
	return false
}

const (
	// DelFlag indicates an item should be removed from the base list.
	DelFlag = "-"
	// AppendFlag indicates an item should be added immediately after an existing one.
	AppendFlag = "+"
	// ReplaceFlag indicates an existing item should be swapped for a new one.
	ReplaceFlag = "="
	// SplitFlag separates the target item from the new value in instructions.
	SplitFlag = ":"
)

// Merge merges user-defined configuration changes (cs) into a base list (s).
// This is used to modify built-in framework defaults without redefining the whole list.
//
// Syntax:
// "-a"      => Remove "a"
// "+a:b"    => Append "b" after "a"
// "=a:b"    => Replace "a" with "b"
// "e"       => Simply add "e" to the end
//
// Example:
// Base:   ["a", "b", "c"]
// Change: ["-a", "+b:b1", "=c:cc", "d"]
// Result: ["b", "b1", "cc", "d"]
func (s JSONStrings) Merge(cs JSONStrings) JSONStrings {
	var ns JSONStrings
	// Iterate through the base list and apply specific modifiers found in cs.
	for _, v := range s {
		values := JSONStrings{v}
		for _, v1 := range cs {
			// Handle deletion: -target
			if v1 == DelFlag+v {
				values = JSONStrings{}
				continue
			}
			// Handle appending: +target:new
			if strings.HasPrefix(v1, AppendFlag+v+SplitFlag) {
				values = JSONStrings{v, strings.TrimPrefix(v1, AppendFlag+v+SplitFlag)}
				continue
			}
			// Handle replacement: =target:new
			if strings.HasPrefix(v1, ReplaceFlag+v+SplitFlag) {
				values = JSONStrings{strings.TrimPrefix(v1, ReplaceFlag+v+SplitFlag)}
				continue
			}
		}
		if len(values) != 0 {
			ns = append(ns, values...)
		}
	}
	// Append new items from cs that are not instructions (don't start with -, +, or =).
	for _, v := range cs {
		if strings.HasPrefix(v, DelFlag) ||
			strings.HasPrefix(v, AppendFlag) ||
			strings.HasPrefix(v, ReplaceFlag) {
			continue
		}
		ns = append(ns, v)
	}
	return ns.unique()
}

// unique removes duplicate strings while preserving the original order.
func (s JSONStrings) unique() JSONStrings {
	var result JSONStrings
	resultMap := make(map[string]struct{}, len(s))
	for _, v := range s {
		if _, ok := resultMap[v]; !ok {
			result = append(result, v)
		}
		resultMap[v] = struct{}{}
	}
	return result
}
