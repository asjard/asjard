package grpc

import (
	"context"
)

// GrpcClient grpc客户端
type GrpcClient struct{}

// Invoke 请求调用
func (c *GrpcClient) Invoke(ctx context.Context) {
	// grpc.WithPerRPCCredentials()
}
