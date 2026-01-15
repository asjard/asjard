package xasynq

import (
	"strings"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
)

// handlerFunc defines the internal signature for Asynq task execution.
// It matches the standard Asjard handler pattern, allowing for interceptors
// (middleware) to be injected into the background task execution flow.
type handlerFunc func(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error)

// ServiceDesc contains the metadata for a task-processing service.
// This is typically generated from a .proto file or defined manually.
type ServiceDesc struct {
	// ServiceName is the full name of the service (e.g., "email.v1.EmailService").
	ServiceName string
	// HandlerType is used for reflection to ensure the provided implementation
	// matches the expected interface.
	HandlerType any
	// Methods is the list of individual task types this service can handle.
	Methods []MethodDesc
}

// MethodDesc represents the specification for a single background task handler.
type MethodDesc struct {
	// Pattern is the raw string identifier for the task (e.g., "/email.v1.EmailService/Send").
	Pattern string
	// Handler is the Go function that executes the task logic.
	Handler handlerFunc
}

// Pattern converts a full method name into a standardized Asynq task identifier.
// It transforms URL-style paths into colon-separated strings and applies
// framework-level resource key logic for global uniqueness.
//
// Example: "/api.v1.User/Create" -> "appname:asynq:api:v1:User:Create"
func Pattern(fulleMethodName string) string {
	// 1. Replace all forward slashes with colons for a flattened hierarchy.
	// 2. Remove any leading colons.
	pattern := strings.TrimPrefix(strings.ReplaceAll(fulleMethodName, "/", ":"), ":")

	// 3. Use the runtime ResourceKey generator to prefix the string with the
	// application name and protocol, ensuring this task doesn't conflict
	// with other apps using the same Redis instance.
	return runtime.GetAPP().ResourceKey(
		Protocol,
		pattern,
		runtime.WithoutService(true),
		runtime.WithDelimiter(":"),
	)
}
