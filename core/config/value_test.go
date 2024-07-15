package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValue(t *testing.T) {
	datas := []struct {
		input Value
		str   string
	}{
		{input: Value{}},
		{input: Value{Sourcer: &testSource{}}},
	}
	for _, data := range datas {
		assert.NotEmpty(t, data.input.String())
	}
}
