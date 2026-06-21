package rest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/resolver"
)

func TestClientResolverContract(t *testing.T) {
	c := New()
	require.NotNil(t, c)
	require.NotNil(t, c.ParseServiceConfig(`{"loadBalancingPolicy":"round_robin"}`))
	require.NoError(t, c.UpdateState(resolver.State{}))
	c.NewAddress(nil)
	c.ReportError(errors.New("resolver error"))
}
