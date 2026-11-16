package bootstrap

import (
	"testing"

	"github.com/asjard/asjard/v2/core/config"
	_ "github.com/asjard/asjard/v2/pkg/config/mem"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
}

func TestWatchLogger(t *testing.T) {
	assert.Nil(t, Bootstrap())
}
