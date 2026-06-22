package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChainUnaryClientInterceptorsAppendsOptionInterceptor(t *testing.T) {
	var order []string
	globalInterceptor := func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error {
		order = append(order, "global:before")
		err := invoker(ctx, method, req, reply, cc)
		order = append(order, "global:after")
		return err
	}
	optionInterceptor := func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error {
		order = append(order, "option:before")
		err := invoker(ctx, method, req, reply, cc)
		order = append(order, "option:after")
		return err
	}

	interceptor := ChainUnaryInterceptors(globalInterceptor, optionInterceptor)
	require.NotNil(t, interceptor)

	err := interceptor(context.Background(), "/test.Service/Call", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc ClientConnInterface) error {
		order = append(order, "invoke")
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, []string{
		"global:before",
		"option:before",
		"invoke",
		"option:after",
		"global:after",
	}, order)
}

func TestChainUnaryClientInterceptorsUsesOptionInterceptorWhenGlobalNil(t *testing.T) {
	called := false
	optionInterceptor := func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error {
		called = true
		return invoker(ctx, method, req, reply, cc)
	}

	interceptor := ChainUnaryInterceptors(nil, optionInterceptor)
	require.NotNil(t, interceptor)

	err := interceptor(context.Background(), "/test.Service/Call", nil, nil, nil, func(ctx context.Context, method string, req, reply any, cc ClientConnInterface) error {
		return nil
	})
	require.NoError(t, err)
	require.True(t, called)
}
