package grpc

import (
	"testing"

	"github.com/asjard/asjard/core/client"
	"github.com/stretchr/testify/require"
)

func TestClientAndConnectionMetadata(t *testing.T) {
	created := NewClient(&client.ClientOptions{})
	require.IsType(t, &Client{}, created)
	conn := ClientConn{serviceName: "users", protocol: Protocol}
	require.Equal(t, "users", conn.ServiceName())
	require.Equal(t, Protocol, conn.Protocol())
	require.Nil(t, conn.Conn())
}

func TestNewConnReusesTarget(t *testing.T) {
	created := NewClient(&client.ClientOptions{}).(*Client)
	t.Cleanup(created.Close)

	conn1, err := created.NewConn("passthrough:///127.0.0.1:1")
	require.NoError(t, err)
	conn2, err := created.NewConn("passthrough:///127.0.0.1:1")
	require.NoError(t, err)
	require.Same(t, conn1, conn2)

	conn3, err := created.NewConn("passthrough:///127.0.0.1:2")
	require.NoError(t, err)
	require.NotSame(t, conn1, conn3)
}
