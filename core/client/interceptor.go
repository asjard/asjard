package client

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/constant"
)

// ClientInterceptor 客户端拦截器需要实现的方法
type ClientInterceptor interface {
	// 拦截器名称
	Name() string
	// 拦截器
	Interceptor() UnaryClientInterceptor
}

// NewClientInterceptor 客户端拦截器初始化方法
type NewClientInterceptor func() (ClientInterceptor, error)

var (
	newClientInterceptors = make(map[string]map[string]NewClientInterceptor)
	ncm                   sync.RWMutex
)

// AddInterceptor 添加客户端拦截器
func AddInterceptor(name string, newInterceptor NewClientInterceptor, supportProtocols ...string) {
	ncm.Lock()
	defer ncm.Unlock()
	if len(supportProtocols) == 0 {
		supportProtocols = []string{constant.AllProtocol}
	}
	for _, protocol := range supportProtocols {
		if _, ok := newClientInterceptors[protocol]; !ok {
			newClientInterceptors[protocol] = map[string]NewClientInterceptor{
				name: newInterceptor,
			}
		} else {
			newClientInterceptors[protocol][name] = newInterceptor
		}
	}
}

// UnaryInvoker is called by UnaryClientInterceptor to complete RPCs.
type UnaryInvoker func(ctx context.Context, method string, req, reply any, cc ClientConnInterface) error

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
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error

func getChainUnaryInterceptors(protocol string, conf Config) (UnaryClientInterceptor, error) {
	interceptors, err := getClientInterceptors(protocol, conf)
	if err != nil {
		return nil, err
	}
	var chainedInt UnaryClientInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		chainedInt = chainUnaryInterceptors(interceptors)
	}
	return chainedInt, nil
}

// 添加默认拦截器
func getClientInterceptors(protocol string, conf Config) ([]UnaryClientInterceptor, error) {
	var interceptors []UnaryClientInterceptor
	ncm.RLock()
	defer ncm.RUnlock()
	newInterceptors := make(map[string]NewClientInterceptor)
	for name, newInterceptor := range newClientInterceptors[protocol] {
		newInterceptors[name] = newInterceptor
	}
	for name, newInterceptor := range newClientInterceptors[constant.AllProtocol] {
		newInterceptors[name] = newInterceptor
	}
	// 顺序需要按照配置执行
	for _, interceptorName := range conf.Interceptors {
		if newInterceptor, ok := newInterceptors[interceptorName]; ok {
			interceptor, err := newInterceptor()
			if err != nil {
				return interceptors, err
			}
			interceptors = append(interceptors, interceptor.Interceptor())
		}
	}
	return interceptors, nil
}

func chainUnaryInterceptors(interceptors []UnaryClientInterceptor) UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error {
		return interceptors[0](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, 0, invoker))
	}
}

func getChainUnaryInvoker(interceptors []UnaryClientInterceptor, curr int, finalInvoker UnaryInvoker) UnaryInvoker {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, method string, req, reply any, cc ClientConnInterface) error {
		return interceptors[curr+1](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, curr+1, finalInvoker))
	}
}
