package xasynq

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/core/status"
	"google.golang.org/grpc/codes"
)

const (
	// AsynqReadEntityInterceptorName is the unique identifier for this interceptor.
	AsynqReadEntityInterceptorName = "asynqReadEntity"
)

func init() {
	// The init function is typically used to register the interceptor to the
	// global framework so it can be enabled via configuration.
	// server.AddInterceptor(AsynqReadEntityInterceptorName, NewAsynqReadEntityInterceptor, Protocol)
}

// AsynqReadEntity is the struct that implements the ServerInterceptor interface.
type AsynqReadEntity struct{}

// Name returns the identifier of the interceptor.
func (r *AsynqReadEntity) Name() string {
	return AsynqReadEntityInterceptorName
}

// NewAsynqReadEntityInterceptor is the factory function to create a new instance.
func NewAsynqReadEntityInterceptor() (server.ServerInterceptor, error) {
	return &AsynqReadEntity{}, nil
}

// Interceptor contains the core logic for the middleware.
// It wraps the business logic handler to perform pre-processing on the task payload.
func (r *AsynqReadEntity) Interceptor() server.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *server.UnaryServerInfo, handler server.UnaryHandler) (resp any, err error) {
		// Type assertion to ensure the context is the specialized xasynq.Context.
		rtx, ok := ctx.(*Context)
		if ok {
			// Retrieve the raw bytes from the task.
			payload := rtx.Payload()

			// Attempt to unmarshal the JSON payload into the 'req' object,
			// which is the struct pointer passed by the generated server code.
			if err := json.Unmarshal(payload, &req); err != nil {
				logger.Error("asynq read entity fail", "payload", string(payload), "err", err)
				// Return a standardized 'InvalidArgument' error if the JSON is malformed.
				return nil, status.Errorf(codes.InvalidArgument, "invalid request")
			}
		} else {
			// This is a safety check: the interceptor should only be used with the xasynq protocol.
			logger.Error("readEntity ctx must be *xasynq.Context", "current", fmt.Sprintf("%T", ctx))
		}

		// Pass the populated 'req' struct to the next handler in the chain (or the final business logic).
		return handler(ctx, req)
	}
}
