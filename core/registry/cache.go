package registry

import (
	"sync"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// healthCheckFunc defines the signature for verifying if a specific service instance is healthy.
type healthCheckFunc func(discoverName string, instance *server.Service) error

// Instance wraps a discovered service with the name of the discovery source (e.g., "etcd", "nacos").
type Instance struct {
	// DiscoverName is the name of the discovery component that found this instance.
	DiscoverName string
	// Service contains the actual connection details (IP, Port, Metadata).
	Service *server.Service
}

// cache maintains a local, thread-safe snapshot of all available services in the cluster.
type cache struct {
	// services stores map[instanceID]*Instance for fast O(1) lookups.
	services map[string]*Instance
	sm       sync.RWMutex // Protects the services map.

	conf *Config

	// healthCheckFunc is the logic used to ping/probe instances.
	healthCheckFunc healthCheckFunc

	// failureThreshold is the limit for consecutive failed health checks before removal.
	failureThreshold int

	// failureThresholds tracks the current failure count for specific instances.
	// key: discoverName + serviceName + serviceID
	failureThresholds map[string]int
	fm                sync.RWMutex // Protects failure count records.

	// listeners stores registered watchers interested in service changes.
	listeners map[string][]*listener
	lm        sync.RWMutex // Protects listener registration.
}

// listener wraps a watch callback with specific filtering options.
type listener struct {
	options  *Options
	callback func(*Event)
}

// newCache initializes the local discovery cache and starts background health checking if enabled.
func newCache(conf *Config, hf healthCheckFunc) *cache {
	c := &cache{
		services:          make(map[string]*Instance),
		failureThreshold:  conf.FailureThreshold,
		healthCheckFunc:   hf,
		failureThresholds: map[string]int{},
		listeners:         map[string][]*listener{},
		conf:              conf,
	}
	// Start the background health check loop if configured.
	if conf.HealthCheck {
		go c.healthCheck()
	}
	return c
}

// canPick evaluates if an instance satisfies the filtering criteria (e.g., version match, tag match).
func (instance *Instance) canPick(options *Options) bool {
	for _, pickFunc := range options.getPickFuncs() {
		if !pickFunc(instance) {
			return false
		}
	}
	return true
}

// pick filters the cache for instances matching the provided options and registers a listener if requested.
func (c *cache) pick(options *Options) []*Instance {
	c.sm.RLock()
	defer c.sm.RUnlock()
	instances := make([]*Instance, 0, len(c.services))
	for _, instance := range c.services {
		if instance.canPick(options) {
			instances = append(instances, instance)
		}
	}
	// Automatically register a watcher if the user provided watch criteria.
	c.addListener(options)
	return instances
}

func (c *cache) isAvailable(options *Options) bool {
	c.sm.RLock()
	defer c.sm.RUnlock()
	for _, instance := range c.services {
		if instance.canPick(options) {
			return true
		}
	}
	return false
}

// addListener registers a callback to be notified when services matching the criteria are added/removed.
func (c *cache) addListener(options *Options) {
	if options.watchName != "" && options.watch != nil {
		c.lm.Lock()
		c.listeners[options.watchName] = append(c.listeners[options.watchName], &listener{
			options:  options,
			callback: options.watch,
		})
		c.lm.Unlock()
	}
}

// removeListener unregisters a watcher by its unique name.
func (c *cache) removeListener(listenerName string) {
	c.lm.Lock()
	for name := range c.listeners {
		if name == listenerName {
			delete(c.listeners, name)
		}
	}
	c.lm.Unlock()
}

// update adds or refreshes service instances in the local cache and triggers notifications.
func (c *cache) update(instances []*Instance) {
	c.sm.Lock()
	defer c.sm.Unlock()
	for _, instance := range instances {
		logger.Debug("update instance",
			"instance_id", instance.Service.Instance.ID,
			"instance_name", instance.Service.Instance.Name,
			"registry", instance.DiscoverName)
		c.services[instance.Service.Instance.ID] = instance
		// Notify interested watchers about the update.
		c.notify(EventTypeUpdate, instance)
	}
}

// delete removes a service instance from the local cache and notifies watchers.
func (c *cache) delete(instance *Instance) {
	logger.Debug("delete instance",
		"instance", instance.Service.Instance.ID,
		"registry", instance.DiscoverName)
	c.sm.Lock()
	if svc, ok := c.services[instance.Service.Instance.ID]; ok {
		delete(c.services, instance.Service.Instance.ID)
		c.notify(EventTypeDelete, svc)
	}
	c.sm.Unlock()
}

// notify iterates through all relevant listeners and executes their callbacks if the instance matches their filters.
func (c *cache) notify(eventType EventType, instance *Instance) {
	c.lm.RLock()
	for _, listeners := range c.listeners {
		for _, listener := range listeners {
			if instance.canPick(listener.options) {
				listener.options.watch(&Event{
					Type:     eventType,
					Instance: instance,
				})
			}
		}
	}
	c.lm.RUnlock()
}

// healthCheck is a loop that periodically triggers the health probe mechanism.
func (c *cache) healthCheck() {
	ticker := time.NewTicker(c.conf.HealthCheckInterval.Duration)
	for {
		select {
		case <-ticker.C:
			c.doHealthCheck()
		}
	}
}

// doHealthCheck executes the probing logic (placeholder implementation commented out).
func (c *cache) doHealthCheck() {
	// Logic: Iterate through services, call healthCheckFunc,
	// update failure thresholds, and delete instances that exceed the threshold.
}

// getFailureThreshold retrieves the current number of consecutive failures for a key.
func (c *cache) getFailureThreshold(failKey string) int {
	c.fm.RLock()
	threshold, ok := c.failureThresholds[failKey]
	c.fm.RUnlock()
	if ok {
		return threshold
	}
	return 1
}

// setFailureThreshold updates the failure count for an instance.
func (c *cache) setFailureThreshold(failKey string, threshold int) {
	c.fm.Lock()
	c.failureThresholds[failKey] = threshold
	c.fm.Unlock()
}
