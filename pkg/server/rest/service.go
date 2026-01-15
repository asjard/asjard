package rest

import (
	"github.com/asjard/asjard/core/server"
)

// methodHandler defines the standard signature for a REST request handler.
// It wraps the business logic with an Asjard Context and supports Unary Interceptors
// for cross-cutting concerns like logging, tracing, or authentication.
type methodHandler func(ctx *Context, srv any, interceptor server.UnaryServerInterceptor) (any, error)

// ServiceDesc represents an RPC service's specification in the REST world.
// It contains the metadata necessary to register a full service (e.g., "UserService").
type ServiceDesc struct {
	ServiceName string       // The unique identifier for the service (e.g., "api.v1.user").
	Name        string       // Human-readable display name for the service.
	Desc        string       // Brief description of the service's purpose.
	HandlerType any          // A pointer to the interface type, used for reflection validation.
	ErrPage     string       // Optional custom error page URL for this specific service.
	Writer      Writer       // Custom response writer (e.g., for specialized XML or binary outputs).
	OpenAPI     []byte       // Marshaled OpenAPI v3 specification data for this service.
	Methods     []MethodDesc // The list of individual endpoints (methods) within this service.
}

// MethodDesc represents the detailed specification for a single RPC method.
// This is used by the router to map physical HTTP requests to Go logic.
type MethodDesc struct {
	// MethodName is the internal name of the function (e.g., "GetUser").
	MethodName string
	// Method is the HTTP Verb (e.g., "GET", "POST", "PUT").
	Method string
	// Path is the URL pattern for the endpoint (e.g., "/api/v1/user/:id").
	Path string
	// Handler is the actual function that executes the request logic.
	Handler methodHandler
	// Name is a human-friendly name for this specific endpoint.
	Name string
	// Desc provides a detailed description of what this endpoint does.
	Desc string
	// WriterName specifies a registered response writer to use (e.g., "json", "proto").
	WriterName string
}
