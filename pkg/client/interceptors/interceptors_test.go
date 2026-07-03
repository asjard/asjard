package interceptors

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	clientgrpc "github.com/asjard/asjard/pkg/client/grpc"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type validatableRequest struct{ err error }

func (r validatableRequest) IsValid(string, string) error { return r.err }

type fakeConn struct{ protocol, service string }

func (c fakeConn) ServiceName() string                                              { return c.service }
func (c fakeConn) Protocol() string                                                 { return c.protocol }
func (c fakeConn) Conn() grpc.ClientConnInterface                                   { return c }
func (fakeConn) Invoke(context.Context, string, any, any, ...grpc.CallOption) error { return nil }
func (fakeConn) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, nil
}

func TestValidateInterceptor(t *testing.T) {
	wantErr := errors.New("invalid")
	called := false
	invoker := func(context.Context, string, any, any, client.ClientConnInterface) error {
		called = true
		return nil
	}
	validate := (&Validate{}).Interceptor()
	require.ErrorIs(t, validate(context.Background(), "/test", validatableRequest{err: wantErr}, nil, fakeConn{}, invoker), wantErr)
	require.False(t, called)
	require.NoError(t, validate(context.Background(), "/test", validatableRequest{}, nil, fakeConn{}, invoker))
	require.True(t, called)
}

func TestPanicInterceptor(t *testing.T) {
	panicInterceptor := (&Panic{}).Interceptor()
	err := panicInterceptor(context.Background(), "/test", "request", nil, fakeConn{protocol: "test", service: "service"},
		func(context.Context, string, any, any, client.ClientConnInterface) error { panic("boom") })
	require.Error(t, err)
	require.Equal(t, PanicInterceptorName, (&Panic{}).Name())
}

func TestCycleChainInterceptor(t *testing.T) {
	interceptor := (CycleChainInterceptor{}).Interceptor()
	called := false
	require.NoError(t, interceptor(context.Background(), "/svc/Method", nil, nil, fakeConn{}, func(context.Context, string, any, any, client.ClientConnInterface) error {
		called = true
		return nil
	}))
	require.True(t, called)

	cc := &clientgrpc.ClientConn{}
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs(HeaderRequestChain, "grpc://svc.Method"))
	err := interceptor(ctx, "/svc/Method", nil, nil, cc, func(context.Context, string, any, any, client.ClientConnInterface) error {
		t.Fatal("cycle must stop invocation")
		return nil
	})
	require.Error(t, err)
}

func TestInterceptorNamesAndConstructors(t *testing.T) {
	constructors := []struct {
		name string
		fn   func() (client.ClientInterceptor, error)
	}{
		{PanicInterceptorName, NewPanic}, {ValidateInterceptorName, NewValidateInterceptor},
		{CycleChainInterceptorName, NewCycleChainInterceptor},
	}
	for _, tc := range constructors {
		got, err := tc.fn()
		require.NoError(t, err)
		require.Equal(t, tc.name, got.Name())
		require.NotNil(t, got.Interceptor())
	}
}
