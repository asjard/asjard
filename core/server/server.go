package server

import (
	"fmt"
)

// Server is the interface that every transport protocol implementation must satisfy.
// It allows the framework to manage servers generically regardless of whether
// they are gRPC, HTTP, or custom TCP servers.
type Server interface {
	// AddHandler registers business logic (services/controllers) to the server.
	AddHandler(handler any) error
	// Start begins listening for incoming requests.
	// Any startup errors are sent to the provided startErr channel.
	Start(startErr chan error) error
	// Stop performs a graceful shutdown of the server.
	Stop()
	// Protocol returns the identifier for the server (e.g., "grpc", "rest").
	Protocol() string
	// ListenAddresses returns the binding and advertising network configuration.
	ListenAddresses() AddressConfig
	// Enabled indicates if the server is active based on current configuration.
	Enabled() bool
}

// NewServerFunc is a factory function signature used to instantiate a specific Server.
type NewServerFunc func(options *ServerOptions) (Server, error)

var (
	// newServerFuncs maintains a registry of available server implementations.
	newServerFuncs = make(map[string]NewServerFunc)
)

// Init initializes all servers that have been registered with the framework.
// For each protocol, it constructs the interceptor chain and passes it to the server factory.
func Init() ([]Server, error) {
	var servers []Server
	for protocol, newServer := range newServerFuncs {
		// 1. Generate the combined middleware chain for this specific protocol.
		interceptor, err := getChainUnaryInterceptors(protocol)
		if err != nil {
			return servers, err
		}

		// 2. Instantiate the server using its factory function and the generated options.
		server, err := newServer(&ServerOptions{
			Interceptor: interceptor,
		})
		if err != nil {
			return servers, err
		}

		servers = append(servers, server)
	}
	return servers, nil
}

// AddServer registers a new server implementation (e.g., a gRPC driver) into the global registry.
// This is typically called by specific protocol packages during their initialization phase.
func AddServer(protocol string, newServerFunc NewServerFunc) error {
	if _, ok := newServerFuncs[protocol]; ok {
		return fmt.Errorf("protocol %s server already exist", protocol)
	}
	newServerFuncs[protocol] = newServerFunc
	return nil
}
