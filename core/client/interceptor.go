package client

import (
	"context"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
)

// CallOption configures a Call before it starts or extracts information from
// a Call after it completes.
type CallOption interface {
	// before is called before the call is sent to any server.  If before
	// returns a non-nil error, the RPC fails with that error.
	before(*callInfo) error

	// after is called after the call has completed.  after cannot return an
	// error, so any failures should be reported via output parameters.
	// after(*callInfo, *csAttempt)
}

// callInfo contains all related configuration and information about an RPC.
type callInfo struct {
	compressorType        string
	failFast              bool
	maxReceiveMessageSize *int
	maxSendMessageSize    *int
	creds                 credentials.PerRPCCredentials
	contentSubtype        string
	// codec                 baseCodec
	maxRetryRPCBufferSize int
	onFinish              []func(err error)
}

func defaultCallInfo() *callInfo {
	return &callInfo{
		failFast:              true,
		maxRetryRPCBufferSize: 256 * 1024, // 256KB
	}
}

// UnaryInvoker is called by UnaryClientInterceptor to complete RPCs.
type UnaryInvoker func(ctx context.Context, method string, req, reply any, cc *resolver.ClientConn, opts ...CallOption) error

// UnaryClientInterceptor intercepts the execution of a unary RPC on the client.
// Unary interceptors can be specified as a DialOption, using
// WithUnaryInterceptor() or WithChainUnaryInterceptor(), when creating a
// ClientConn. When a unary interceptor(s) is set on a ClientConn, gRPC
// delegates all unary RPC invocations to the interceptor, and it is the
// responsibility of the interceptor to call invoker to complete the processing
// of the RPC.
//
// method is the RPC name. req and reply are the corresponding request and
// response messages. cc is the ClientConn on which the RPC was invoked. invoker
// is the handler to complete the RPC and it is the responsibility of the
// interceptor to call it. opts contain all applicable call options, including
// defaults from the ClientConn as well as per-call options.
//
// The returned error must be compatible with the status package.
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply any, cc *resolver.ClientConn, invoker UnaryInvoker, opts ...CallOption) error
