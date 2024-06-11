package server

import (
	"context"
	"fmt"

	"github.com/asjard/asjard/core/config"
)

// UnaryServerInfo consists of various information about a unary RPC on
// server side. All per-rpc information may be mutated by the interceptor.
type UnaryServerInfo struct {
	// Server is the service implementation the user provides. This is read-only.
	Server any
	// FullMethod is the full RPC method string, i.e., /package.service/method.
	FullMethod string
	// 协议
	Protocol string
}

// UnaryHandler defines the handler invoked by UnaryServerInterceptor to complete the normal
// execution of a unary RPC.
//
// If a UnaryHandler returns an error, it should either be produced by the
// status package, or be one of the context errors. Otherwise, gRPC will use
// codes.Unknown as the status code and err.Error() as the status message of the
// RPC.
type UnaryHandler func(ctx context.Context, req any) (any, error)

// UnaryServerInterceptor provides a hook to intercept the execution of a unary RPC on the server. info
// contains all the information of this RPC the interceptor can operate on. And handler is the wrapper
// of the service method implementation. It is the responsibility of the interceptor to invoke handler
// to complete the RPC.
type UnaryServerInterceptor func(ctx context.Context, req any, info *UnaryServerInfo, handler UnaryHandler) (resp any, err error)

// ServerInterceptor 服务拦截器需要实现的方法
type ServerInterceptor interface {
	// 拦截器名称
	Name() string
	// 拦截器
	Interceptor() UnaryServerInterceptor
}

type NewServerInterceptor func() ServerInterceptor

var newServerInterceptors []NewServerInterceptor

// AddInterceptor 添加拦截器
func AddInterceptor(newInterceptor NewServerInterceptor) {
	newServerInterceptors = append(newServerInterceptors, newInterceptor)
}

// TODO 添加默认拦截器
func getServerInterceptors(protocol string) []UnaryServerInterceptor {
	var interceptors []UnaryServerInterceptor
	// 自定义拦截器
	for _, interceptorName := range config.GetStrings(fmt.Sprintf("servers.%s.interceptors", protocol),
		config.GetStrings("servers.interceptors", []string{})) {
		for _, newInterceptor := range newServerInterceptors {
			interceptor := newInterceptor()
			if interceptor.Name() == interceptorName {
				interceptors = append(interceptors, interceptor.Interceptor())
			}
		}
	}
	return interceptors
}

func getChainUnaryInterceptors(protocol string) UnaryServerInterceptor {
	interceptors := getServerInterceptors(protocol)
	var chainedInt UnaryServerInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = chainUnaryInterceptors(interceptors)
	}
	return chainedInt
}

func chainUnaryInterceptors(interceptors []UnaryServerInterceptor) UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *UnaryServerInfo, handler UnaryHandler) (any, error) {
		return interceptors[0](ctx, req, info, getChainUnaryHandler(interceptors, 0, info, handler))
	}
}

func getChainUnaryHandler(interceptors []UnaryServerInterceptor, curr int, info *UnaryServerInfo, finalHandler UnaryHandler) UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(ctx context.Context, req any) (any, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}
