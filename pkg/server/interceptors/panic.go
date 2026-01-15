package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
)

const (
	// PanicInterceptorName is the unique identifier for this interceptor.
	PanicInterceptorName = "panic"
)

func init() {
	// Register the panic recovery interceptor globally for all server protocols.
	server.AddInterceptor(PanicInterceptorName, NewPanic)
}

// Panic represents the recovery interceptor component.
type Panic struct{}

// NewPanic initializes the panic interceptor.
func NewPanic() (server.ServerInterceptor, error) {
	return &Panic{}, nil
}

// Interceptor returns the middleware function that handles panic recovery.
func (*Panic) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		// Use defer to ensure the recovery logic runs even if the handler panics.
		defer func() {
			if rcv := recover(); rcv != nil {
				// 1. Collect diagnostic information about the crash.
				args := []any{
					"err", rcv, // The actual panic object/message.
					"req", req, // The request payload that triggered the panic.
					"method", info.FullMethod, // The endpoint being called.
					"protocol", info.Protocol, // gRPC or REST.
					"stack", string(debug.Stack()), // The full goroutine stack trace.
				}

				// 2. Log the incident with Error level for alerting and debugging.
				logger.L(ctx).Error("request panic", args...)

				// 3. Mask the internal crash from the client by returning a
				// standardized 500 Internal Server Error.
				err = status.InternalServerError()
			}
		}()

		// Execute the next interceptor or the final business logic handler.
		return handler(ctx, req)
	}
}

// Name returns the interceptor's unique name.
func (*Panic) Name() string {
	return PanicInterceptorName
}
