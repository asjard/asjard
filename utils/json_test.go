package utils

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDuration(t *testing.T) {
	datas := []struct {
		input  string
		output JSONDuration
	}{
		{input: `1`, output: JSONDuration{Duration: time.Duration(1)}},
		{input: `"1s"`, output: JSONDuration{Duration: time.Second}},
	}
	for _, data := range datas {
		var output JSONDuration
		if err := json.Unmarshal([]byte(data.input), &output); err != nil {
			t.Error(err)
			t.FailNow()
		}
		if output.Duration != data.output.Duration {
			t.Errorf("input %s not equal, current %s, want: %s", data.input,
				output.String(), data.output.String())
			t.FailNow()
		}
	}
}

func TestStrings(t *testing.T) {
	datas := []struct {
		input  string
		output JSONStrings
	}{
		{input: `[1,2,3]`, output: []string{"1", "2", "3"}},
		{input: `["1","2","3"]`, output: []string{"1", "2", "3"}},
		{input: `["a","b","c"]`, output: []string{"a", "b", "c"}},
		{input: `"1,2,3"`, output: []string{"1", "2", "3"}},
		{input: `"a,b,c"`, output: []string{"a", "b", "c"}},
	}
	for _, data := range datas {
		var output JSONStrings
		if err := json.Unmarshal([]byte(data.input), &output); err != nil {
			t.Errorf("unmarshal %s fail %v", data.input, err)
			t.FailNow()
		}
		if len(output) != len(data.output) {
			t.Errorf("input %s lenght not equal, current %d, want: %d", data.input,
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
}
