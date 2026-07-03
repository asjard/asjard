package server

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceEndpoints(t *testing.T) {
	s := &Service{Endpoints: make(map[string]*Endpoint)}
	require.Error(t, s.AddEndpoint("", AddressConfig{}))
	require.NoError(t, s.AddEndpoint("grpc", AddressConfig{Listen: "127.0.0.1:9000", Advertise: "service:9000"}))
	require.Equal(t, []string{"127.0.0.1:9000"}, s.GetListenAddresses("grpc"))
	require.Equal(t, []string{"service:9000"}, s.GetAdvertiseAddresses("grpc"))
	require.Empty(t, s.GetListenAddresses("missing"))
	endpoint, ok := s.GetEndpoint("grpc")
	require.True(t, ok)
	require.Len(t, endpoint.Listen, 1)
}

func TestServiceConcurrentEndpoints(t *testing.T) {
	s := &Service{Endpoints: make(map[string]*Endpoint)}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			require.NoError(t, s.AddEndpoint("grpc", AddressConfig{Advertise: "service:9000"}))
			_ = s.GetAdvertiseAddresses("grpc")
		}()
	}
	wg.Wait()
	require.Len(t, s.GetAdvertiseAddresses("grpc"), 20)
}

func TestChainUnaryInterceptors(t *testing.T) {
	var calls []string
	makeInterceptor := func(name string) UnaryServerInterceptor {
		return func(ctx context.Context, req any, info *UnaryServerInfo, next UnaryHandler) (any, error) {
			calls = append(calls, name+"-before")
			resp, err := next(ctx, req)
			calls = append(calls, name+"-after")
			return resp, err
		}
	}
	chain := chainUnaryInterceptors([]UnaryServerInterceptor{makeInterceptor("one"), makeInterceptor("two")})
	resp, err := chain(context.Background(), "request", &UnaryServerInfo{}, func(context.Context, any) (any, error) {
		calls = append(calls, "handler")
		return "response", nil
	})
	require.NoError(t, err)
	require.Equal(t, "response", resp)
	require.Equal(t, []string{"one-before", "two-before", "handler", "two-after", "one-after"}, calls)

	wantErr := errors.New("stop")
	short := chainUnaryInterceptors([]UnaryServerInterceptor{func(context.Context, any, *UnaryServerInfo, UnaryHandler) (any, error) {
		return nil, wantErr
	}})
	_, err = short(context.Background(), nil, &UnaryServerInfo{}, func(context.Context, any) (any, error) {
		t.Fatal("handler must not be called")
		return nil, nil
	})
	require.ErrorIs(t, err, wantErr)
}
