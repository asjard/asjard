package status

import (
	"testing"

	"github.com/asjard/asjard/core/config"
	"google.golang.org/grpc/codes"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	m.Run()
}

func TestNewCode(t *testing.T) {
	datas := []struct {
		input  codes.Code
		output codes.Code
	}{
		{input: codes.Aborted, output: 10040910},
		{input: 100, output: 100500100},
		{input: 1001, output: 1001001},
		{input: 4001, output: 1004001},
	}
	for _, data := range datas {
		output := newCode(data.input)
		t.Log("input", data.input, "output", output)
		if output != data.output {
			t.Errorf("test %d fail, current: %d, want: %d", data.input, output, data.output)
			t.FailNow()
		}
	}
}

func TestParseCode(t *testing.T) {
	datas := []struct {
		input                               codes.Code
		code, systemCode, httpCode, errCode uint32
	}{
		{input: 1, httpCode: 499, errCode: 1},
		{input: 1004045, systemCode: 100, httpCode: 404, errCode: 5},
	}

	for _, data := range datas {
		systemCode, httpCode, errCode := parseCode(data.input)
		if systemCode != data.systemCode || httpCode != data.httpCode || errCode != data.errCode {
			t.Errorf("test %d fail, current: %d,%d,%d, want: %d,%d,%d", data.input,
				systemCode, httpCode, errCode,
				data.systemCode, data.httpCode, data.errCode)
			t.FailNow()
		}
	}
}
