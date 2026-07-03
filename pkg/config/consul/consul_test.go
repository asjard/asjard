package consul

import (
	"testing"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/runtime"
	"github.com/stretchr/testify/require"
)

func TestConsulSourceMetadataAndKeys(t *testing.T) {
	s := &Consul{
		options: &config.SourceOptions{},
		conf:    &Config{Client: "default", Delimiter: "/"},
		app: runtime.APP{App: "shop", Environment: "prod", Region: "cn", AZ: "a", Instance: runtime.Instance{
			ID: "id", Name: "api", Group: "service",
		}},
	}
	require.Equal(t, Name, s.Name())
	require.Equal(t, Priority, s.Priority())
	require.Empty(t, s.GetAll())
	require.NoError(t, s.Set("key", "value"))
	require.Equal(t, "shop/configs", s.prefix())
	require.Equal(t, "name", s.configKey("shop/configs/", "shop/configs/name"))
	require.Len(t, s.prefixs(), 15)
	s.Disconnect()
}
