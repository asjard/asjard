package cast

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestToBoolE(t *testing.T) {
	c := qt.New(t)

	var jf, jt, je json.Number
	_ = json.Unmarshal([]byte("0"), &jf)
	_ = json.Unmarshal([]byte("1"), &jt)
	_ = json.Unmarshal([]byte("1.0"), &je)
	tests := []struct {
		input  interface{}
		expect bool
		iserr  bool
	}{
		{0, false, false},
		{int64(0), false, false},
		{int32(0), false, false},
		{int16(0), false, false},
		{int8(0), false, false},
		{uint(0), false, false},
		{uint64(0), false, false},
		{uint32(0), false, false},
		{uint16(0), false, false},
		{uint8(0), false, false},
		{float64(0), false, false},
		{float32(0), false, false},
		{time.Duration(0), false, false},
		{jf, false, false},
		{nil, false, false},
		{"false", false, false},
		{"FALSE", false, false},
		{"False", false, false},
		{"f", false, false},
		{"F", false, false},
		{false, false, false},

		{"true", true, false},
		{"TRUE", true, false},
		{"True", true, false},
		{"t", true, false},
		{"T", true, false},
		{1, true, false},
		{int64(1), true, false},
		{int32(1), true, false},
		{int16(1), true, false},
		{int8(1), true, false},
		{uint(1), true, false},
		{uint64(1), true, false},
		{uint32(1), true, false},
		{uint16(1), true, false},
		{uint8(1), true, false},
		{float64(1), true, false},
		{float32(1), true, false},
		{time.Duration(1), true, false},
		{jt, true, false},
		{je, true, false},
		{true, true, false},
		{-1, true, false},
		{int64(-1), true, false},
		{int32(-1), true, false},
		{int16(-1), true, false},
		{int8(-1), true, false},

		// errors
		{"test", true, false},
		{testing.T{}, true, true},
	}

	for i, test := range tests {
		errmsg := qt.Commentf("i = %d, input = %v", i, test.input) // assert helper message
		v, err := ToBoolE(test.input)
		if test.iserr {
			c.Assert(err, qt.IsNotNil, errmsg)
			continue
		}

		c.Assert(err, qt.IsNil)
		c.Assert(v, qt.Equals, test.expect, errmsg)
	}
}

func TestToStringSliceE(t *testing.T) {
	c := qt.New(t)

	tests := []struct {
		input     interface{}
		expect    []string
		iserr     bool
		delimiter string
	}{
		{[]int{1, 2}, []string{"1", "2"}, false, ""},
		{[]int8{int8(1), int8(2)}, []string{"1", "2"}, false, ""},
		{[]int32{int32(1), int32(2)}, []string{"1", "2"}, false, ""},
		{[]int64{int64(1), int64(2)}, []string{"1", "2"}, false, ""},
		{[]float32{float32(1.01), float32(2.01)}, []string{"1.01", "2.01"}, false, ""},
		{[]float64{float64(1.01), float64(2.01)}, []string{"1.01", "2.01"}, false, ""},
		{[]string{"a", "b"}, []string{"a", "b"}, false, ""},
		{[]interface{}{1, 3}, []string{"1", "3"}, false, ""},
		{interface{}(1), []string{"1"}, false, ""},
		{[]error{errors.New("a"), errors.New("b")}, []string{"a", "b"}, false, ""},
		// errors
		{nil, nil, true, ""},
		{"1 2 3", []string{"1", "2", "3"}, false, ""},
		{"1|2|3", []string{"1", "2", "3"}, false, "|"},
		{"1|2| 3", []string{"1", "2", "3"}, false, "|"},
		{"1|2|3 ", []string{"1", "2", "3"}, false, "|"},
		{"1,2,3", []string{"1", "2", "3"}, false, ","},
		{"1, 2,3", []string{"1", "2", "3"}, false, ","},
		{testing.T{}, nil, true, ""},
	}

	for i, test := range tests {
		errmsg := qt.Commentf("i = %d", i) // assert helper message

		v, err := ToStringSliceE(test.input, test.delimiter)
		if test.iserr {
			c.Assert(err, qt.IsNotNil)
			continue
		}

		c.Assert(err, qt.IsNil)
		c.Assert(v, qt.DeepEquals, test.expect, errmsg)
	}
}
