package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/server"
)

// RateLimiter 服务端限速拦截器
// 无需实现redis版本的限速器
// 限速的目的是为了保护服务,以免服务过载
// 如需精确限制访问速度，请参考quota拦截器
type RateLimiter struct{}

func NewRateLimiterInterceptor() server.ServerInterceptor {
	return &RateLimiter{}
}

// Interceptor 拦截器实现
func (RateLimiter) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		return handler(ctx, req)
	}
}

// Name 拦截器名称
func (RateLimiter) Name() string {
	return "ratelimiter"
}
