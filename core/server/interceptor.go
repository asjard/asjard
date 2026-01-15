package server

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
)

// UnaryServerInfo contains metadata about a single (unary) RPC call.
// This is passed to interceptors to provide context about the server and method being called.
type UnaryServerInfo struct {
	// Server is the underlying service implementation.
	Server any
	// FullMethod is the path to the RPC (e.g., "/user.UserService/GetUser").
	FullMethod string
	// Protocol identifies the transport (e.g., "grpc", "rest").
	Protocol string
}

// UnaryHandler is the signature of the final business logic or the next step in the chain.
type UnaryHandler func(ctx context.Context, req any) (any, error)

// UnaryServerInterceptor is a middleware function that wraps the request execution.
// It can modify the context/request before execution or the response/error after.
type UnaryServerInterceptor func(ctx context.Context, req any, info *UnaryServerInfo, handler UnaryHandler) (resp any, err error)

// ServerInterceptor is the interface for pluggable middleware components.
type ServerInterceptor interface {
	// Name returns the unique identifier for the interceptor (e.g., "logger").
	Name() string
	// Interceptor returns the actual function that performs the wrapping.
	Interceptor() UnaryServerInterceptor
}

// NewServerInterceptor is a factory function type to initialize an interceptor.
type NewServerInterceptor func() (ServerInterceptor, error)

var (
	// newServerInterceptors stores interceptors mapped by protocol -> name -> factory.
	newServerInterceptors = make(map[string]map[string]NewServerInterceptor)
	nsm                   sync.RWMutex
)

// AddInterceptor registers an interceptor. If no protocols are specified,
// it is applied to "AllProtocol" (global).
func AddInterceptor(name string, newInterceptor NewServerInterceptor, supportProtocols ...string) {
	nsm.Lock()
	if len(supportProtocols) == 0 {
		supportProtocols = []string{constant.AllProtocol}
	}
	for _, protocol := range supportProtocols {
		if _, ok := newServerInterceptors[protocol]; !ok {
			newServerInterceptors[protocol] = map[string]NewServerInterceptor{
				name: newInterceptor,
			}
		} else {
			newServerInterceptors[protocol][name] = newInterceptor
		}
	}
	nsm.Unlock()
}

// getServerInterceptors retrieves and initializes the interceptors configured for a protocol.
// It merges protocol-specific interceptors with global ones based on the server configuration.
func getServerInterceptors(protocol string) ([]UnaryServerInterceptor, error) {
	logger.Debug("get server intereptors", "protocol", protocol)
	var interceptors []UnaryServerInterceptor
	nsm.RLock()
	defer nsm.RUnlock()

	// Temporary map to collect relevant factory functions.
	newInterceptors := make(map[string]NewServerInterceptor)
	for name, newInterceptor := range newServerInterceptors[protocol] {
		newInterceptors[name] = newInterceptor
	}
	for name, newInterceptor := range newServerInterceptors[constant.AllProtocol] {
		newInterceptors[name] = newInterceptor
	}

	conf := GetConfigWithProtocol(protocol)
	// Build the slice based on the order defined in the configuration.
	for _, interceptorName := range conf.Interceptors {
		if newInerceptor, ok := newInterceptors[interceptorName]; ok {
			interceptor, err := newInerceptor()
			if err != nil {
				return interceptors, err
			}
			interceptors = append(interceptors, interceptor.Interceptor())
		}
	}
	return interceptors, nil
}

// getChainUnaryInterceptors converts a slice of interceptors into a single
// chained interceptor.
func getChainUnaryInterceptors(protocol string) (UnaryServerInterceptor, error) {
	interceptors, err := getServerInterceptors(protocol)
	if err != nil {
		return nil, err
	}
	var chainedInt UnaryServerInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		// More than one? Chain them together recursively.
		chainedInt = chainUnaryInterceptors(interceptors)
	}
	return chainedInt, nil
}

// chainUnaryInterceptors kicks off the recursive wrapping of interceptors.
func chainUnaryInterceptors(interceptors []UnaryServerInterceptor) UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *UnaryServerInfo, handler UnaryHandler) (any, error) {
		// Start with the first interceptor and pass a handler that points to the rest of the chain.
		return interceptors[0](ctx, req, info, getChainUnaryHandler(interceptors, 0, info, handler))
	}
}

// getChainUnaryHandler returns a UnaryHandler that, when called, executes the next interceptor in the slice.
func getChainUnaryHandler(interceptors []UnaryServerInterceptor, curr int, info *UnaryServerInfo, finalHandler UnaryHandler) UnaryHandler {
	// Base case: if we are at the last interceptor, the "next" step is the actual business logic.
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	// Return a handler that triggers the next interceptor in the sequence.
	return func(ctx context.Context, req any) (any, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}
