package pprof

import (
	"testing"

	"github.com/asjard/asjard/core/server"
	"github.com/stretchr/testify/require"
)

func TestPprofServerContract(t *testing.T) {
	conf := server.Config{Enabled: true, Addresses: server.AddressConfig{Listen: "127.0.0.1:0"}}
	created, err := MustNew(conf, &server.ServerOptions{})
	require.NoError(t, err)
	s := created.(*PprofServer)
	require.Equal(t, Protocol, s.Protocol())
	require.True(t, s.Enabled())
	require.Equal(t, conf.Addresses, s.ListenAddresses())
	require.NoError(t, s.AddHandler(struct{}{}))
	s.Stop()
}

func TestPprofStartRequiresAddress(t *testing.T) {
	s := &PprofServer{}
	require.Error(t, s.Start(make(chan error, 1)))
}
