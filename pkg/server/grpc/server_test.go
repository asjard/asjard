package grpc

import (
	"testing"

	"github.com/asjard/asjard/core/server"
	"github.com/stretchr/testify/require"
)

func TestGrpcServerContract(t *testing.T) {
	conf := Config{Config: server.Config{Enabled: true, Addresses: server.AddressConfig{Listen: "127.0.0.1:0"}}}
	created, err := MustNew(conf, &server.ServerOptions{})
	require.NoError(t, err)
	s := created.(*GrpcServer)
	require.Equal(t, Protocol, s.Protocol())
	require.True(t, s.Enabled())
	require.Equal(t, conf.Addresses, s.ListenAddresses())
	require.Error(t, s.AddHandler(struct{}{}))
	s.Stop()
}

func TestGrpcStartRequiresAddress(t *testing.T) {
	created, err := MustNew(Config{}, &server.ServerOptions{})
	require.NoError(t, err)
	require.Error(t, created.Start(make(chan error, 1)))
	created.Stop()
}
