package interceptors

import (
	"context"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

const (
	// McpAuthInterceptorName is the unique identifier for this interceptor.
	McpAuthInterceptorName = "mcpAuth"
)

func init() {
	// Register the panic recovery interceptor globally for all server protocols.
	server.AddInterceptor(McpAuthInterceptorName, NewMcpAuth)
}

// McpAuth represents the recovery interceptor component.
type McpAuth struct{}

// NewPanic initializes the panic interceptor.
func NewMcpAuth() (server.ServerInterceptor, error) {
	return &McpAuth{}, nil
}

// Interceptor returns the middleware function that handles panic recovery.
func (p *McpAuth) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		logger.L(ctx).Debug("start server interceptor", "interceptor", p.Name(), "full_method", info.FullMethod, "protocol", info.Protocol)
		// Execute the next interceptor or the final business logic handler.
		return handler(ctx, req)
	}
}

// Name returns the interceptor's unique name.
func (*McpAuth) Name() string {
	return McpAuthInterceptorName
}
