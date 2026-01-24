package registry

import "github.com/asjard/asjard/core/logger"

const (
	// LocalRegistryName is the identifier for the built-in local service discovery.
	LocalRegistryName = "local"
)

// Discovery defines the standard interface that any service discovery provider
// (e.g., ETCD, Consul) must implement to plug into the framework.
type Discovery interface {
	// GetAll retrieves the full list of available service instances from the remote registry.
	GetAll() ([]*Instance, error)
	// Name returns the unique identifier of the discovery implementation.
	Name() string
}

// CallbackFunc is the signature for functions that react to service topology changes.
type CallbackFunc func(event *Event)

// DiscoveryOptions encapsulates the configuration for initializing a Discovery provider.
type DiscoveryOptions struct {
	// Callback is triggered by the provider when it detects changes in the remote registry.
	Callback CallbackFunc
}

// DiscoveryOption is a functional argument for customizing DiscoveryOptions.
type DiscoveryOption func(options *DiscoveryOptions)

// NewDiscoveryFunc is a factory function type that creates a new Discovery instance.
type NewDiscoveryFunc func(options *DiscoveryOptions) (Discovery, error)

// WithDiscoveryCallback is a functional option to attach a watcher to the discovery process.
func WithDiscoveryCallback(callback CallbackFunc) func(options *DiscoveryOptions) {
	return func(options *DiscoveryOptions) {
		options.Callback = callback
	}
}

// NewDiscoveryOptions aggregates multiple functional options into a single options struct.
func NewDiscoveryOptions(opts ...DiscoveryOption) *DiscoveryOptions {
	options := &DiscoveryOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

var (
	// newDiscoverys maintains a registry of available discovery implementation factories.
	newDiscoverys = make(map[string]NewDiscoveryFunc)
)

// AddDiscover registers a new discovery provider factory (e.g., called by an 'etcd' driver init function).
func AddDiscover(name string, newFunc NewDiscoveryFunc) error {
	logger.Debug("add discover", "name", name)
	newDiscoverys[name] = newFunc
	return nil
}

// Discover initiates the discovery process through the global registry manager.
// It usually triggers the initial pull of services from all registered sources.
func Discover() error {
	return registryManager.discove()
}

// PickServices is the primary API used by Load Balancers or Clients.
// It returns a filtered list of service instances from the local cache based on criteria
// like service name, version, or labels.
func PickServices(opts ...Option) []*Instance {
	options := NewOptions(opts)
	if len(opts) == 0 {
		options = DefaultOptions()
	}
	return registryManager.pick(options)
}

// IsAvailable checks if there is at least one service instance currently available
// in the registry that matches the given criteria.
//
// If no options are provided, it uses the default configuration to perform the check.
//
// Example:
//
//	// Check if any instance of the current service is available
//	if registry.IsAvailable() {
//	    fmt.Println("Service is ready")
//	}
//
//	// Check with specific constraints (e.g., version or region)
//	ready := registry.IsAvailable(registry.WithVersion("1.0.0"), registry.WithRegion("us-east-1"))
func IsAvailable(opts ...Option) bool {
	options := NewOptions(opts)
	// If no options are provided, fallback to DefaultOptions to ensure
	// a consistent baseline for the availability check.
	if len(opts) == 0 {
		options = DefaultOptions()
	}
	return registryManager.isAvailable(options)
}

// RemoveListener unregisters a service-watch listener by name to stop receiving topology updates.
func RemoveListener(name string) {
	registryManager.removeListener(name)
}
