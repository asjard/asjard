package interceptors

import (
	"context"
	"strings"

	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/status"
	"github.com/asjard/asjard/pkg/client/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

const (
	// HeaderRequestChain stores the call path. Format: {protocol}://{app}/{serviceName}/{method}
	HeaderRequestChain = "x-request-chain"
	// HeaderRequestDest represents the request destination identifier.
	HeaderRequestDest = "x-request-dest"
	// HeaderRequestApp represents the requesting application name.
	HeaderRequestApp = "x-request-app"
	// CycleChainInterceptorName is the unique identifier for this interceptor.
	CycleChainInterceptorName = "cycleChainInterceptor"
)

// CycleChainInterceptor detects circular dependencies in the service call graph.
// It relies on metadata being injected into the context by the load balancer
// or previous interceptors.
type CycleChainInterceptor struct {
}

func init() {
	// Register the interceptor specifically for gRPC client protocols.
	client.AddInterceptor(CycleChainInterceptorName, NewCycleChainInterceptor, grpc.Protocol)
}

// NewCycleChainInterceptor creates a new instance of the cycle detection interceptor.
func NewCycleChainInterceptor() (client.ClientInterceptor, error) {
	return &CycleChainInterceptor{}, nil
}

// Name returns the interceptor's registration name.
func (CycleChainInterceptor) Name() string {
	return CycleChainInterceptorName
}

// Interceptor provides the logic to track the call chain and prevent loops.
// It currently focuses on gRPC-to-gRPC hops within a distributed trace.
func (s CycleChainInterceptor) Interceptor() client.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc client.ClientConnInterface, invoker client.UnaryInvoker) error {
		// Only apply logic to gRPC client connections.
		if _, ok := cc.(*grpc.ClientConn); !ok {
			return invoker(ctx, method, req, reply, cc)
		}

		// Extract metadata from the context to check the existing call chain.
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			// Fallback to incoming context if outgoing is empty (start of a new hop).
			md, ok = metadata.FromIncomingContext(ctx)
			if !ok {
				md = metadata.New(map[string]string{})
			}
		}

		// Format the current method into a standardized string: grpc://Service.Method
		currentRequestMethod := "grpc://" + strings.ReplaceAll(strings.Trim(method, "/"), "/", ".")

		// 1. Detect Cycles:
		// Check if the current method has already appeared in the upstream chain.
		if requestChains, ok := md[HeaderRequestChain]; ok {
			for _, requestMethod := range requestChains {
				if requestMethod == currentRequestMethod {
					// Found a match! Append it one last time for visibility and return an error.
					requestChains = append(requestChains, currentRequestMethod)
					return status.Errorf(codes.Canceled, "cycle call, chains: %s", strings.Join(requestChains, " -> "))
				}
			}
		}

		// 2. Propagate Chain:
		// Append the current method to the chain and pass it down to the next service.
		md[HeaderRequestChain] = append(md[HeaderRequestChain], currentRequestMethod)

		// Create a new context containing the updated outgoing metadata.
		return invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc)
	}
}
