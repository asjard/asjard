package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/status"
)

const (
	// PanicInterceptorName is the unique identifier for this interceptor.
	PanicInterceptorName = "panic"
)

func init() {
	// Register the panic recovery interceptor globally for all server protocols.
	client.AddInterceptor(PanicInterceptorName, NewPanic)
}

// Panic represents the recovery interceptor component.
type Panic struct{}

// NewPanic initializes the panic interceptor.
func NewPanic() (client.ClientInterceptor, error) {
	return &Panic{}, nil
}

// Interceptor returns the middleware function that handles panic recovery.
func (*Panic) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) (err error) {
		// Use defer to ensure the recovery logic runs even if the handler panics.
		defer func() {
			if rcv := recover(); rcv != nil {
				// 1. Collect diagnostic information about the crash.
				args := []any{
					"err", rcv, // The actual panic object/message.
					"req", req, // The request payload that triggered the panic.
					"method", method, // The endpoint being called.
					"protocol", cc.Protocol(), // gRPC or REST.
					"service", cc.ServiceName(),
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
		return invoker(ctx, method, req, reply, cc)
	}
}

// Name returns the interceptor's unique name.
func (*Panic) Name() string {
	return PanicInterceptorName
}
