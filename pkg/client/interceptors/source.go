package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/client"
	"google.golang.org/grpc"
)

const (
	// HeaderSourceServiceName 源服务名称
	HeaderSourceServiceName = "x-request-source"
	// HeaderSourceMethod 源服务方法
	HeaderSourceMethod = "x-request-method"
)

// SourceInterceptor 来源拦截器
type SourceInterceptor struct{}

func init() {
	client.AddInterceptor(NewSourceInterceptor)
}

// NewSourceInterceptor 初始化来源拦截器
func NewSourceInterceptor() client.ClientInterceptor {
	return &SourceInterceptor{}
}

// Name 拦截器名称
func (SourceInterceptor) Name() string {
	return "sourceInterceptor"
}

// Interceptor 拦截器
// 上下文中添加当前服务
// 如果出现循环服务则拦截
func (SourceInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc grpc.ClientConnInterface, invoker client.UnaryInvoker) error {
		return invoker(ctx, method, req, reply, cc)
	}
}
