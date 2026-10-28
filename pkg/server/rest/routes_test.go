package rest

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAddRoute(t *testing.T) {

}

func TestDefaultConfig(t *testing.T) {
	conf := DefaultConfig(Protocol)
	require.Equal(t, Protocol, "rest")
	require.NotEmpty(t, conf.Cors.AllowMethods)
}
