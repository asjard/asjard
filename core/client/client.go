package client

import (
	"context"

	"google.golang.org/grpc"
)

// ClientConn 客户端需要实现的方法
type ClientConn interface {
	Invoke(ctx context.Context, method string, req any) (any, error)
}

// ClientConnInterface defines the functions clients need to perform unary and
// streaming RPCs.  It is implemented by *ClientConn, and is only intended to
// be referenced by generated code.
type ClientConnInterface interface {
	// Invoke performs a unary RPC and returns after the response is received
	// into reply.
	Invoke(ctx context.Context, method string, args any, reply any, opts ...CallOption) error
	// NewStream begins a streaming RPC.
	// NewStream(ctx context.Context, desc *StreamDesc, method string, opts ...CallOption) (ClientStream, error)
}

// NewClient .
func NewClient(protocol, serviceName string) {
	grpc.NewClient(serviceName)
}
