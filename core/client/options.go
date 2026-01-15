package client

import (
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/resolver"
)

// ClientOptions contains the functional configuration used to initialize a ClientConnection.
// It encapsulates the core RPC components required for service discovery,
// traffic distribution, and request interception.
type ClientOptions struct {
	// Resolver is the gRPC builder responsible for translating a target URI
	// into a list of physical backend addresses (endpoints).
	Resolver resolver.Builder

	// Balancer is the RPC builder responsible for selecting a specific backend
	// address from the list provided by the resolver based on a defined strategy
	// (e.g., Round Robin, Locality-Aware).
	Balancer balancer.Builder

	// Interceptor defines a single unary interceptor or a chained interceptor
	// that wraps the RPC call to provide cross-cutting concerns like logging,
	// tracing, and security.
	Interceptor UnaryClientInterceptor
}
