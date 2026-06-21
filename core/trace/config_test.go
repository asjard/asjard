package trace

import (
	"os"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestDefaultTraceConfigAndDisabledInit(t *testing.T) {
	conf := GetConfig()
	require.False(t, conf.Enabled)
	require.Equal(t, time.Second, conf.Timeout.Duration)
	require.NoError(t, Init())
}
