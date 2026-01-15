package server

// ServerOptions encapsulates the functional configurations for a server instance.
// These options are typically built by the framework using the configuration
// files and the registered interceptor plugins.
type ServerOptions struct {
	// Interceptor is the primary entry point for the server's middleware pipeline.
	// It usually contains a "chained" interceptor that wraps multiple individual
	// interceptors (logging, metrics, auth, etc.) into a single execution flow.
	// If nil, the server will execute the business logic directly without middleware.
	Interceptor UnaryServerInterceptor
}
