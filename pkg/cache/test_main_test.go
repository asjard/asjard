package cache

import (
	"os"
	"testing"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
