package client

import (
	"context"
	"sync"

	"github.com/asjard/asjard/core/constant"
)

// ClientInterceptor defines the interface for a client-side interceptor.
// Implementing this allows a module to provide metadata and the actual interceptor logic.
type ClientInterceptor interface {
	// Name returns the unique identifier of the interceptor.
	Name() string
	// Interceptor returns the functional UnaryClientInterceptor.
	Interceptor() UnaryClientInterceptor
}

// NewClientInterceptor is a factory function type that initializes a ClientInterceptor.
type NewClientInterceptor func() (ClientInterceptor, error)

var (
	// newClientInterceptors stores interceptor factories mapped by protocol and then by name.
	newClientInterceptors = make(map[string]map[string]NewClientInterceptor)
	// ncm protects access to the interceptor registry.
	ncm sync.RWMutex
)

// AddInterceptor registers a client interceptor.
// If supportProtocols is empty, the interceptor is registered for all protocols.
func AddInterceptor(name string, newInterceptor NewClientInterceptor, supportProtocols ...string) {
	ncm.Lock()
	defer ncm.Unlock()
	if len(supportProtocols) == 0 {
		supportProtocols = []string{constant.AllProtocol}
	}
	for _, protocol := range supportProtocols {
		if _, ok := newClientInterceptors[protocol]; !ok {
			newClientInterceptors[protocol] = map[string]NewClientInterceptor{
				name: newInterceptor,
			}
		} else {
			newClientInterceptors[protocol][name] = newInterceptor
		}
	}
}

// UnaryInvoker is the completion function called by interceptors to proceed with the RPC.
type UnaryInvoker func(ctx context.Context, method string, req, reply any, cc ClientConnInterface) error

// UnaryClientInterceptor is a function that intercepts a unary RPC call.
// It can perform logic before and after the invoker is called, such as logging,
// tracing, or modifying request metadata.
type UnaryClientInterceptor func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error

// getChainUnaryInterceptors retrieves and chains interceptors based on protocol and configuration.
func getChainUnaryInterceptors(protocol string, conf Config) (UnaryClientInterceptor, error) {
	interceptors, err := getClientInterceptors(protocol, conf)
	if err != nil {
		return nil, err
	}
	var chainedInt UnaryClientInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
		// Create a single functional chain from the slice of interceptors.
		chainedInt = chainUnaryInterceptors(interceptors)
	}
	return chainedInt, nil
}

// getClientInterceptors builds an ordered slice of interceptors based on the Config.Interceptors list.
func getClientInterceptors(protocol string, conf Config) ([]UnaryClientInterceptor, error) {
	var interceptors []UnaryClientInterceptor
	ncm.RLock()
	defer ncm.RUnlock()

	// Merge protocol-specific and global interceptor factories.
	newInterceptors := make(map[string]NewClientInterceptor)
	for name, newInterceptor := range newClientInterceptors[protocol] {
		newInterceptors[name] = newInterceptor
	}
	for name, newInterceptor := range newClientInterceptors[constant.AllProtocol] {
		newInterceptors[name] = newInterceptor
	}

	// Instantiate interceptors in the order specified in the configuration.
	for _, interceptorName := range conf.Interceptors {
		if newInterceptor, ok := newInterceptors[interceptorName]; ok {
			interceptor, err := newInterceptor()
			if err != nil {
				return interceptors, err
			}
			interceptors = append(interceptors, interceptor.Interceptor())
		}
	}
	return interceptors, nil
}

// chainUnaryInterceptors creates a single interceptor out of a chain of many interceptors.
// This implements the recursive onion model for middleware execution.
func chainUnaryInterceptors(interceptors []UnaryClientInterceptor) UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc ClientConnInterface, invoker UnaryInvoker) error {
		// Start the execution at index 0.
		return interceptors[0](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, 0, invoker))
	}
}

// getChainUnaryInvoker recursively wraps the final invoker with the next interceptor in the chain.
func getChainUnaryInvoker(interceptors []UnaryClientInterceptor, curr int, finalInvoker UnaryInvoker) UnaryInvoker {
	if curr == len(interceptors)-1 {
		return finalInvoker
	}
	return func(ctx context.Context, method string, req, reply any, cc ClientConnInterface) error {
		// Point to the next interceptor in the slice.
		return interceptors[curr+1](ctx, method, req, reply, cc, getChainUnaryInvoker(interceptors, curr+1, finalInvoker))
	}
}
