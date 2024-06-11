package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/client"
	"google.golang.org/grpc"
)

// CircuitBreaker 断路器
type CircuitBreaker struct{}

func init() {
	client.AddInterceptor(NewCircuitBreaker)
}

// NewCircuitBreaker 拦截器初始化
func NewCircuitBreaker() client.ClientInterceptor {
	// hystrix.Configure(cmds map[string]hystrix.CommandConfig)
	return &CircuitBreaker{}
}

// Name 拦截器名称
func (CircuitBreaker) Name() string {
	return "circuitBreaker"
}

// Interceptor 拦截器实现
func (CircuitBreaker) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface, invoker client.UnaryInvoker) error {
		return invoker(ctx, method, req, reply, cc)
	}
}
