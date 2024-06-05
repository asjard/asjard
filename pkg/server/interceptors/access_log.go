package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/server"
)

// AccessLog access日志拦截器
type AccessLog struct{}

func init() {
	server.AddInterceptor(NewAccessLogInterceptor)
}

// NewAccessLogInterceptor .
func NewAccessLogInterceptor() server.ServerInterceptor {
	return &AccessLog{}
}

// Name 日志拦截器名称
func (AccessLog) Name() string {
	return "accessLog"
}

// Interceptor 拦截器实现
// 垮协议拦截器
func (AccessLog) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		return handler(ctx, req)
	}
}
