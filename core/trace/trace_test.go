package trace

import (
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/stretchr/testify/require"
)

func TestOTLPHTTPInitializationIntegration(t *testing.T) {
	require.NoError(t, config.Set("asjard.trace.enabled", true))
	require.NoError(t, config.Set("asjard.trace.endpoint", "http://127.0.0.1:4318/v1/traces"))
	require.Eventually(t, func() bool { return GetConfig().Enabled }, 3*time.Second, 20*time.Millisecond)
	require.NoError(t, Init())
}
