package client

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// ClientOptions  客户端参数
type ClientOptions struct {
	// 解析器
	Resolver resolver.Builder
	// 负载均衡
	Balancer balancer.Builder
	// 拦截器
	Interceptor UnaryClientInterceptor
}
