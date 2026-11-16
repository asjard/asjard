package grpc

import (
	"testing"

	"github.com/asjard/asjard/v2/core/client"
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
