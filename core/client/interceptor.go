package client

import (
	"context"
	"fmt"

	"github.com/asjard/asjard/core/config"
	"google.golang.org/grpc"
)

// ClientInterceptor 客户端拦截器需要实现的方法
type ClientInterceptor interface {
	// 拦截器名称
	Name() string
	// 拦截器
	Interceptor() UnaryClientInterceptor
}

// NewClientInterceptor 客户端拦截器初始化方法
type NewClientInterceptor func() ClientInterceptor

var newClientInterceptors []NewClientInterceptor

// AddInterceptor 添加客户端拦截器
func AddInterceptor(newInterceptor NewClientInterceptor) {
	newClientInterceptors = append(newClientInterceptors, newInterceptor)
}

// UnaryInvoker is called by UnaryClientInterceptor to complete RPCs.
type UnaryInvoker func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface) error

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
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface, invoker UnaryInvoker) error

func getChainUnaryInterceptors(protocol string) UnaryClientInterceptor {
	interceptors := getClientInterceptors(protocol)
	var chainedInt UnaryClientInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = chainUnaryInterceptors(interceptors)
	}
	return chainedInt
}

// TODO 添加默认拦截器
func getClientInterceptors(protocol string) []UnaryClientInterceptor {
	var interceptors []UnaryClientInterceptor
	for _, interceptorName := range config.GetStrings(fmt.Sprintf("clients.%s.interceptors", protocol),
		config.GetStrings("clients.interceptors", []string{})) {
		for _, newInterceptor := range newClientInterceptors {
			interceptor := newInterceptor()
			if interceptor.Name() == interceptorName {
				interceptors = append(interceptors, interceptor.Interceptor())
			}
		}
	}
	return interceptors
}

func chainUnaryInterceptors(interceptors []UnaryClientInterceptor) UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface, invoker UnaryInvoker) error {
		return interceptors[0](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, 0, invoker))
	}
}

func getChainUnaryInvoker(interceptors []UnaryClientInterceptor, curr int, finalInvoker UnaryInvoker) UnaryInvoker {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface) error {
		return interceptors[curr+1](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, curr+1, finalInvoker))
	}
}
