package grpc

import (
	"context"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// TestInterceptor .
type TestInterceptor struct{}

// TestInterceptor1 .
type TestInterceptor1 struct{}

var _ server.ServerInterceptor = &TestInterceptor{}
var _ server.ServerInterceptor = &TestInterceptor1{}

func init() {
	server.AddInterceptor(NewTestInterceptor)
}

// NewTestInterceptor .
func NewTestInterceptor() server.ServerInterceptor {
	return &TestInterceptor{}
}

// Name .
func (TestInterceptor) Name() string {
	return "testInterceptor"
}

// Interceptor .
func (TestInterceptor) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (any, error) {
		logger.Debugf("test grpc interceptor")
		return handler(ctx, req)
	}
}

// Name .
func (TestInterceptor1) Name() string {
	return "testInterceptor1"
}

// Interceptor .
func (TestInterceptor1) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (any, error) {
		logger.Debugf("test grpc interceptor1")
		return handler(ctx, req)
	}
}
