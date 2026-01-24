package registry

import (
	"fmt"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// Registry is a composite interface that requires both Discovery and Registration capabilities.
type Registry interface {
	Discovery
	Register
}

// RegistryManager orchestrates the interaction between the local service,
// the service cache, and various external registry backends.
type RegistryManager struct {
	// cache is the local storage for all discovered service instances.
	cache *cache

	// currentService represents the metadata of the currently running application.
	currentService *server.Service
	// conf holds the global registry/discovery configuration settings.
	conf *Config

	// registers is the list of active backends where this service is published.
	registers []Register
	// discovers is the list of active backends being polled for upstream services.
	discovers []Discovery
}

var registryManager *RegistryManager

// Package-level initialization of the singleton manager.
func init() {
	registryManager = &RegistryManager{}
}

// Init prepares the manager by loading the current service context, configuration,
// and initializing the local discovery cache.
func Init() error {
	registryManager.currentService = server.GetService()
	registryManager.conf = GetConfig()
	// Pass healthCheck method as the callback for the cache's background prober.
	registryManager.cache = newCache(registryManager.conf, registryManager.healthCheck)
	return nil
}

// registe initializes all registered backends and begins the service publication process.
func (r *RegistryManager) registe() error {
	if !r.conf.AutoRegiste {
		return nil
	}
	// Instantiate all available registration plugins.
	for _, newRegister := range newRegisters {
		register, err := newRegister()
		if err != nil {
			return err
		}
		r.registers = append(r.registers, register)
	}

	// Support for "Warm-up" periods before making the service discoverable.
	if r.conf.DelayRegiste.Duration != 0 {
		return r.delayRegiste(r.conf.DelayRegiste.Duration)
	}
	return r.doRegiste()
}

// delayRegiste schedules the actual registration after a specific time duration.
func (r *RegistryManager) delayRegiste(duration time.Duration) error {
	go func(duration time.Duration) {
		t := time.After(duration)
		<-t
		r.doRegiste()
	}(duration)
	return nil
}

// doRegiste performs the actual call to external registry backends.
func (r *RegistryManager) doRegiste() error {
	for _, register := range r.registers {
		if err := register.Registe(r.currentService); err != nil {
			return err
		}
	}
	return nil
}

// heartbeat starts a background goroutine to periodically signal vitality to registers.
func (r *RegistryManager) heartbeat() error {
	go func(duration time.Duration) {
		ticker := time.NewTicker(duration)
		for {
			<-ticker.C
			r.doHeartbeat()
		}
	}(r.conf.HeartbeatInterval.Duration)
	return nil
}

// doHeartbeat sends the keep-alive signal to all registers.
func (r *RegistryManager) doHeartbeat() {
	// Implementation placeholder for heartbeat logic.
}

// remove handles the "un-registration" of the service, typically during shutdown.
func (r *RegistryManager) remove() error {
	if !r.conf.AutoRegiste {
		return nil
	}
	for _, register := range r.registers {
		register.Remove(r.currentService)
	}
	return nil
}

// discove initializes the discovery backends and performs the initial pull of service data.
func (r *RegistryManager) discove() error {
	if !r.conf.AutoDiscove {
		logger.Warn("registry.autoDiscove not enabled")
		return nil
	}

	// Initialize discovery providers with a callback to the 'watch' method.
	for name, newDiscover := range newDiscoverys {
		logger.Debug("add discover", "name", name)
		discover, err := newDiscover(NewDiscoveryOptions(WithDiscoveryCallback(r.watch)))
		if err != nil {
			return err
		}
		r.discovers = append(r.discovers, discover)
	}

	// Initial population of the local cache.
	for _, discover := range r.discovers {
		services, err := discover.GetAll()
		if err != nil {
			return err
		}
		logger.Debug("discover get all service", "name", discover.Name(), "services", services)
		r.cache.update(services)
	}
	return nil
}

// healthCheck facilitates active probing of discovered instances using the specific discovery source.
func (r *RegistryManager) healthCheck(discoverName string, instance *server.Service) error {
	for _, discover := range r.discovers {
		if discover.Name() == discoverName {
			// Proxy health check to the specific discovery implementation if supported.
		}
	}
	return fmt.Errorf("service '%s(%s)' health check discover '%s' not found",
		instance.Instance.Name, instance.Instance.ID, discoverName)
}

// watch acts as the central dispatcher for events coming from discovery providers.
func (r *RegistryManager) watch(et *Event) {
	switch et.Type {
	case EventTypeCreate, EventTypeUpdate:
		r.update(et)
	case EventTypeDelete:
		r.delete(et)
	}
}

// update synchronizes a created or modified instance into the local cache.
func (r *RegistryManager) update(event *Event) {
	r.cache.update([]*Instance{event.Instance})
}

// delete removes an instance from the local cache when it leaves the registry.
func (r *RegistryManager) delete(event *Event) {
	r.cache.delete(event.Instance)
}

// pick retrieves a filtered list of services from the local cache.
func (r *RegistryManager) pick(options *Options) []*Instance {
	return r.cache.pick(options)
}

// checks if there is at least one service instance currently available
// in the registry that matches the given criteria.
func (r *RegistryManager) isAvailable(options *Options) bool {
	return r.cache.isAvailable(options)
}

// removeListener stops watching for changes under a specific listener ID.
func (r *RegistryManager) removeListener(name string) {
	r.cache.removeListener(name)
}
