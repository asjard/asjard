package utils

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		datas := []struct {
			input  string
			output JSONDuration
			ok     bool
		}{
			{input: `1`, output: JSONDuration{Duration: time.Duration(1)}, ok: true},
			{input: `"1s"`, output: JSONDuration{Duration: time.Second}, ok: true},
			{input: `"1`, ok: false},
			{input: `{"a":1}`, ok: false},
		}
		for _, data := range datas {
			var output JSONDuration
			err := json.Unmarshal([]byte(data.input), &output)
			if (err == nil) != data.ok || (data.ok && output.Duration != data.output.Duration) {
				t.Errorf("input %s not equal, current %s, want: %s", data.input,
					output.String(), data.output.String())
				t.Error(err)
				t.FailNow()
			}
		}
	})
	t.Run("Marshal", func(t *testing.T) {
		datas := []struct {
			input  JSONDuration
			output string
			ok     bool
		}{
			{input: JSONDuration{Duration: time.Duration(1)}, output: `"1ns"`, ok: true},
			{input: JSONDuration{Duration: time.Second}, output: `"1s"`, ok: true},
		}
		for _, data := range datas {
			b, err := json.Marshal(data.input)
			if (err == nil) != data.ok || (data.ok && data.output != string(b)) {
				t.Errorf("marshal %v fail, current: %s, want: %s", data.input, string(b), data.output)
				t.Error(err)
				t.FailNow()
			}
		}
	})
}

func TestStrings(t *testing.T) {
	t.Run("Unmarshal", func(t *testing.T) {
		datas := []struct {
			input  string
			output JSONStrings
			ok     bool
		}{
			{input: `[1,2,3]`, output: []string{"1", "2", "3"}, ok: true},
			{input: `["1","2","3"]`, output: []string{"1", "2", "3"}, ok: true},
			{input: `["a","b","c"]`, output: []string{"a", "b", "c"}, ok: true},
			{input: `"1,2,3"`, output: []string{"1", "2", "3"}, ok: true},
			{input: `"a,b,c"`, output: []string{"a", "b", "c"}, ok: true},
			{input: `"`, ok: false},
			{input: `"a`, ok: false},
			{input: `""`, ok: true},
		}
		for _, data := range datas {
			var output JSONStrings
			err := json.Unmarshal([]byte(data.input), &output)
			if (err == nil) != data.ok {
				t.Errorf("unmarshal %s fail %v", data.input, err)
				t.FailNow()
			}
			if !data.ok {
				continue
			}
			if len(output) != len(data.output) {
				t.Errorf("input %s length not equal, current %d, want: %d", data.input,
					len(output), len(data.output))
				t.FailNow()
			}
			for index, v := range output {
				if v != data.output[index] {
					t.Errorf("input %s index %d value not equal, current %s, want %s",
						data.input, index, v, data.output[index])
					t.FailNow()
				}
			}
		}
	})
	t.Run("Marshal", func(t *testing.T) {
		datas := []struct {
			input  JSONStrings
			output string
			ok     bool
		}{
			{input: []string{"a"}, output: `"a"`, ok: true},
			{input: []string{"a", "b"}, output: `"a,b"`, ok: true},
		}
		for _, data := range datas {
			b, err := json.Marshal(data.input)
			if (err == nil) != data.ok || (data.ok && string(b) != data.output) {
				t.Errorf("test %v fail, current: %s, want: %s", data.input, string(b), data.output)
				t.Error(err)
				t.FailNow()
			}
		}
	})
}
