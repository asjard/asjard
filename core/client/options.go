package client

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// ClientOptions  客户端参数
type ClientOptions struct {
	Resolver resolver.Builder
	Balancer balancer.Builder
}
