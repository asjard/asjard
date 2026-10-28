package xfasthttp

import (
	"testing"
	"time"

	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/stretchr/testify/require"
)

func TestDefaultRESTTimeouts(t *testing.T) {
	conf := DefaultConfig(rest.Protocol)
	require.Equal(t, 30*time.Second, conf.Options.ReadTimeout.Duration)
	require.Equal(t, 60*time.Second, conf.Options.IdleTimeout.Duration)
}
