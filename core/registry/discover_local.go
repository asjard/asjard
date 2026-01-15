package registry

import (
	"net/url"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
	"github.com/google/uuid"
)

const (
	// LocalDiscoverName is the internal identifier for this local implementation.
	LocalDiscoverName = "localDiscover"
)

// LocalRegistry implements the Discovery interface using local configuration.
// It maps static configuration keys to service instances.
type LocalRegistry struct {
	discoveryOptions        *DiscoveryOptions
	instances               []*Instance
	localDiscoverConfPrefix string
	dm                      sync.RWMutex
}

func init() {
	// Register this provider with the discovery factory system during package initialization.
	AddDiscover(LocalDiscoverName, NewLocalDiscover)
}

// NewLocalDiscover creates a new instance of the local discovery provider.
// It immediately starts loading and watching the local configuration.
func NewLocalDiscover(options *DiscoveryOptions) (Discovery, error) {
	localDiscover := &LocalRegistry{
		localDiscoverConfPrefix: constant.ConfigRegistryLocalDiscoverPrefix, // Usually "asjard.registry.localDiscover"
		discoveryOptions:        options,
	}
	localDiscover.getAndWatch()
	return localDiscover, nil
}

// GetAll returns the current list of service instances loaded from local config.
func (l *LocalRegistry) GetAll() ([]*Instance, error) {
	return l.instances, nil
}

// Name returns the standard name for this registry type.
func (l *LocalRegistry) Name() string {
	return LocalRegistryName
}

// getAndWatch performs the initial load of services and sets up a configuration observer.
func (l *LocalRegistry) getAndWatch() {
	services := make(map[string][]string)
	// Fetches configuration and registers 'l.watch' as a callback for changes.
	if err := config.GetWithUnmarshal(l.localDiscoverConfPrefix,
		&services,
		config.WithWatch(l.watch)); err != nil {
		logger.Error("get registry.localDiscover fail",
			"err", err.Error())
	} else {
		l.instances = l.getInstances(services)
	}
}

// watch is triggered whenever the local configuration file is modified.
// it effectively "reloads" the service list and notifies the manager of changes.
func (l *LocalRegistry) watch(event *config.Event) {
	services := make(map[string][]string)
	if err := config.GetWithUnmarshal(l.localDiscoverConfPrefix, &services); err != nil {
		logger.Error("get local discover conf fail", "err", err)
	}

	instances := l.getInstances(services)

	// Notify the registry manager to clear old instances.
	for _, instance := range l.instances {
		l.discoveryOptions.Callback(&Event{
			Type:     EventTypeDelete,
			Instance: instance,
		})
	}

	// Notify the registry manager to add the new/updated instances.
	for _, instance := range instances {
		l.discoveryOptions.Callback(&Event{
			Type:     EventTypeUpdate,
			Instance: instance,
		})
	}
	l.instances = instances
}

// getInstances converts a map of service names and address strings into structured Instance objects.
// Example input: {"user-service": ["http://127.0.0.1:8080", "grpc://127.0.0.1:9090"]}
func (l *LocalRegistry) getInstances(services map[string][]string) []*Instance {
	var instances []*Instance
	for name, addresses := range services {
		service := server.NewService()
		service.Instance.Name = name
		// Since these are static, we generate a random UUID for the instance ID.
		service.Instance.ID = uuid.NewString()

		for index := range addresses {
			u, err := url.Parse(addresses[index])
			if err == nil {
				// Parse the scheme (http, grpc, etc.) and the host:port.
				service.AddEndpoint(u.Scheme, server.AddressConfig{
					Listen: u.Host,
				})
			}
		}
		instances = append(instances, &Instance{
			DiscoverName: l.Name(),
			Service:      service,
		})
	}
	return instances
}
