package registry

import (
	"github.com/asjard/asjard/core/server"
)

// Register defines the standard lifecycle methods for service announcement.
// Any backend (e.g., Zookeeper, Eureka, ETCD) must implement this interface.
type Register interface {
	// Registe publishes the service's existence, endpoints, and metadata
	// to the remote registry.
	Registe(service *server.Service) error

	// Remove explicitly withdraws the service instance from the registry,
	// typically called during a graceful shutdown.
	Remove(service *server.Service)

	// Name returns the unique identifier for the registration backend (e.g., "consul").
	Name() string
}

// NewRegisterFunc is a factory function type used to initialize a specific
// Register implementation.
type NewRegisterFunc func() (Register, error)

// newRegisters acts as a plugin registry, storing factory methods for all
// available registration backends.
var (
	newRegisters = make(map[string]NewRegisterFunc)
)

// AddRegister is called by driver packages (in their init functions) to
// register themselves with the framework.
func AddRegister(name string, newFunc NewRegisterFunc) error {
	newRegisters[name] = newFunc
	return nil
}

// Registe triggers the global registration process.
// The registryManager will iterate through all enabled backends in newRegisters
// and call their specific Registe implementation.
func Registe() error {
	return registryManager.registe()
}

// Unregiste triggers the global removal process.
// This ensures the service is cleaned up from all registries to prevent
// "zombie" instances from receiving traffic after the process stops.
func Unregiste() error {
	return registryManager.remove()
}
