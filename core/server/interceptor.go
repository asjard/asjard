package server

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
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

type NewServerInterceptor func() (ServerInterceptor, error)

var (
	newServerInterceptors = make(map[string]map[string]NewServerInterceptor)
	nsm                   sync.RWMutex
)

// AddInterceptor 添加拦截器
func AddInterceptor(name string, newInterceptor NewServerInterceptor, supportProtocols ...string) {
	nsm.Lock()
	if len(supportProtocols) == 0 {
		supportProtocols = []string{constant.AllProtocol}
	}
	for _, protocol := range supportProtocols {
		if _, ok := newServerInterceptors[protocol]; !ok {
			newServerInterceptors[protocol] = map[string]NewServerInterceptor{
				name: newInterceptor,
			}
		} else {
			newServerInterceptors[protocol][name] = newInterceptor
		}
	}
	nsm.Unlock()
}

// 获取协议拦截器
func getServerInterceptors(protocol string) ([]UnaryServerInterceptor, error) {
	logger.Debug("get server intereptors", "protocol", protocol)
	var interceptors []UnaryServerInterceptor
	nsm.RLock()
	defer nsm.RUnlock()
	newInterceptors := make(map[string]NewServerInterceptor)
	for name, newInterceptor := range newServerInterceptors[protocol] {
		newInterceptors[name] = newInterceptor
	}
	for name, newInterceptor := range newServerInterceptors[constant.AllProtocol] {
		newInterceptors[name] = newInterceptor
	}
	conf := GetConfigWithProtocol(protocol)
	// 自定义拦截器
	for _, interceptorName := range conf.Interceptors {
		if newInerceptor, ok := newInterceptors[interceptorName]; ok {
			interceptor, err := newInerceptor()
			if err != nil {
				return interceptors, err
			}
			interceptors = append(interceptors, interceptor.Interceptor())
		}
	}
	return interceptors, nil
}

func getChainUnaryInterceptors(protocol string) (UnaryServerInterceptor, error) {
	interceptors, err := getServerInterceptors(protocol)
	if err != nil {
		return nil, err
	}
	var chainedInt UnaryServerInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = chainUnaryInterceptors(interceptors)
	}
	return chainedInt, nil
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
