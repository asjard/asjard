package interceptors

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/server"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type validatableRequest struct{ err error }

func (r validatableRequest) IsValid(string, string) error { return r.err }

func TestValidateInterceptor(t *testing.T) {
	wantErr := errors.New("invalid")
	called := false
	interceptor := (&Validate{}).Interceptor()
	_, err := interceptor(context.Background(), validatableRequest{err: wantErr}, &server.UnaryServerInfo{FullMethod: "/test"}, func(context.Context, any) (any, error) {
		called = true
		return nil, nil
	})
	require.ErrorIs(t, err, wantErr)
	require.False(t, called)
	resp, err := interceptor(context.Background(), validatableRequest{}, &server.UnaryServerInfo{FullMethod: "/test"}, func(context.Context, any) (any, error) {
		called = true
		return "ok", nil
	})
	require.NoError(t, err)
	require.Equal(t, "ok", resp)
}

func TestPanicInterceptor(t *testing.T) {
	_, err := (&Panic{}).Interceptor()(context.Background(), "request", &server.UnaryServerInfo{FullMethod: "/test", Protocol: "test"}, func(context.Context, any) (any, error) {
		panic("boom")
	})
	require.Error(t, err)
}

func TestBasicNamesAndMetricsPassThrough(t *testing.T) {
	require.Equal(t, PanicInterceptorName, (&Panic{}).Name())
	require.Equal(t, ValidateInterceptorName, (&Validate{}).Name())
	require.Equal(t, MetricsInterceptorName, (Metrics{}).Name())
	created, err := NewMetricsInterceptor()
	require.NoError(t, err)
	resp, err := created.Interceptor()(context.Background(), nil, &server.UnaryServerInfo{FullMethod: "/test", Protocol: "grpc"}, func(context.Context, any) (any, error) {
		return "ok", nil
	})
	require.NoError(t, err)
	require.Equal(t, "ok", resp)
}
