/*
Package consul implements service registration and discovery using HashiCorp Consul.
It allows the application to participate in a service mesh by announcing its
endpoints and watching for other services.
*/
package consul

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/registry"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/core/server"
	"github.com/asjard/asjard/pkg/stores/consul"
	"github.com/asjard/asjard/utils"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

const (
	// NAME is the unique identifier for this registry implementation.
	NAME                 = "consul"
	AddressTypeListen    = "listen"
	AddressTypeAdvertise = "advertise"
)

// Consul manages the lifecycle of the Consul client and service watches.
type Consul struct {
	client           *api.Client
	conf             *Config
	exit             chan struct{}              // Used to stop the TTL heartbeat goroutine.
	discoveryOptions *registry.DiscoveryOptions // Stores callback and filtering logic for discovery.
}

// Config defines the settings for the Consul registry.
type Config struct {
	Client  string             `json:"client"`  // Named client from the store package.
	Timeout utils.JSONDuration `json:"timeout"` // Heartbeat interval and TTL.
	Tags    utils.JSONStrings  `json:"tags"`    // Global tags added to all registered services.
}

var (
	// Ensure the Consul struct satisfies the Registry and Discovery interfaces.
	_ registry.Register  = &Consul{}
	_ registry.Discovery = &Consul{}

	defaultConfig = Config{
		Client:  consul.DefaultClientName,
		Timeout: utils.JSONDuration{Duration: 5 * time.Second},
	}

	newConsul *Consul
	newOnce   sync.Once
)

func init() {
	// Register the factory functions with the global registry manager.
	registry.AddRegister(NAME, NewRegister)
	registry.AddDiscover(NAME, NewDiscovery)
}

// NewRegister initializes the Consul client for service registration.
func NewRegister() (registry.Register, error) {
	return New(nil)
}

// NewDiscovery initializes the Consul client and starts the background service watch.
func NewDiscovery(options *registry.DiscoveryOptions) (registry.Discovery, error) {
	discovery, err := New(options)
	if err != nil {
		return nil, err
	}
	// Start the global catalog watch.
	if err := newServiceWatch(discovery); err != nil {
		return nil, err
	}
	return discovery, nil
}

// New is a helper to ensure the Consul client is created only once (Singleton).
func New(options *registry.DiscoveryOptions) (*Consul, error) {
	var err error
	newOnce.Do(func() {
		consulRegistry := &Consul{
			exit: make(chan struct{}),
		}
		err = consulRegistry.loadConfig()
		if err != nil {
			return
		}
		// Fetch the concrete consul client based on configured name.
		consulRegistry.client, err = consul.Client(consul.WithClientName(consulRegistry.conf.Client))
		if err != nil {
			return
		}
		newConsul = consulRegistry
	})
	if err != nil {
		return nil, err
	}
	if options != nil {
		newConsul.discoveryOptions = options
	}
	return newConsul, nil
}

// Registe publishes the service endpoints and metadata to Consul.
// It also starts a background goroutine to maintain the TTL heartbeat.
func (c *Consul) Registe(service *server.Service) error {
	appDetail, err := json.Marshal(&service.APP)
	if err != nil {
		return err
	}
	endpoints, err := json.Marshal(&service.Endpoints)
	if err != nil {
		return err
	}

	// TTL is set slightly higher than the heartbeat interval to prevent flapping.
	ttl := c.conf.Timeout.Duration + time.Second
	meta := map[string]string{
		"app":        service.App,
		"app_detail": string(appDetail),
		"endpoints":  string(endpoints),
	}

	// Merge instance-specific metadata.
	for k, v := range service.APP.Instance.MetaData {
		meta[k] = v
	}

	tags := c.conf.Tags
	for protocol := range service.Endpoints {
		tags = append(tags, protocol)
	}

	registration := &api.AgentServiceRegistration{
		ID:   service.Instance.ID,
		Name: service.Instance.Name,
		Meta: meta,
		Tags: tags,
		Check: &api.AgentServiceCheck{
			CheckID: service.Instance.ID,
			Name:    service.Instance.Name,
			TTL:     ttl.String(),
			Status:  "passing",
			// Automatically cleanup the service if heartbeats stop for too long.
			DeregisterCriticalServiceAfter: c.conf.Timeout.String(),
		},
	}

	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return err
	}

	// TTL Heartbeat Goroutine.
	go func() {
		for {
			select {
			case <-c.exit:
				return
			case <-time.After(c.conf.Timeout.Duration):
				if err := c.client.Agent().UpdateTTL(service.Instance.ID, "", "passing"); err != nil {
					logger.Error("consul update ttl fail", "instance", service.Instance.ID, "err", err)
				}

			}
		}
	}()
	return nil
}

// Remove stops the heartbeat and deregisters the service from the Consul agent.
func (c *Consul) Remove(service *server.Service) {
	c.exit <- struct{}{}
	if err := c.client.Agent().ServiceDeregister(service.Instance.ID); err != nil {
		logger.Error("remove instance fail", "err", err)
	}
}

func (c *Consul) Name() string { return NAME }

// GetAll fetches all active instances currently known by the local Consul agent.
func (c *Consul) GetAll() ([]*registry.Instance, error) {
	serviceMap, err := c.getAgentServices()
	if err != nil {
		return []*registry.Instance{}, err
	}
	instances := make([]*registry.Instance, 0, len(serviceMap))
	for _, service := range serviceMap {
		instances = append(instances, &registry.Instance{
			DiscoverName: NAME,
			Service:      service,
		})
	}
	return instances, nil
}

// getAgentServices queries Consul for all services belonging to this application.
func (c *Consul) getAgentServices() (map[string]*server.Service, error) {
	// Filter results based on the "app" metadata key.
	agentServices, err := c.client.Agent().ServicesWithFilter("Meta.app==" + runtime.GetAPP().App)
	if err != nil {
		logger.Error("get all instance from consul fail", "err", err)
		return nil, err
	}
	serviceMap := make(map[string]*server.Service)
	for serviceId, agentService := range agentServices {
		endpoints := agentService.Meta["endpoints"]
		appDetail := agentService.Meta["app_detail"]
		if _, ok := serviceMap[serviceId]; !ok {
			var service server.Service
			if err := json.Unmarshal([]byte(appDetail), &service.APP); err != nil {
				logger.Error("consul unmarshal appDetail fail", "app_detail", appDetail, "err", err)
				return nil, err
			}
			if err := json.Unmarshal([]byte(endpoints), &service.Endpoints); err != nil {
				logger.Error("consul unmarshal endpoints fail", "endpoints", endpoints, "err", err)
				return nil, err
			}
			serviceMap[serviceId] = &service
		}
	}
	return serviceMap, nil
}

func (c *Consul) loadConfig() error {
	conf := defaultConfig
	if err := config.GetWithUnmarshal("asjard.registry.consul", &conf); err != nil {
		return err
	}
	c.conf = &conf
	return nil
}

// serviceWatch manages watches for the entire service catalog.
type serviceWatch struct {
	services map[string]*watch.Plan // Tracks active instance watchers per service name.
	sm       sync.RWMutex
	c        *Consul
}

// newServiceWatch starts watching for any service names appearing in the Consul catalog.
func newServiceWatch(c *Consul) error {
	watcher := &serviceWatch{
		services: make(map[string]*watch.Plan),
		c:        c,
	}
	pl, err := watch.Parse(map[string]any{
		"type": "services",
	})
	if err != nil {
		return err
	}
	pl.Handler = watcher.serviceHandler
	go func() {
		if err := pl.RunWithClientAndHclog(watcher.c.client, nil); err != nil {
			logger.Error("consul watch service fail", "err", err)
		}
	}()
	return nil
}

// serviceHandler is called when the list of services in Consul changes.
func (s *serviceWatch) serviceHandler(_ uint64, data any) {
	switch d := data.(type) {
	case map[string][]string:
		for service := range d {
			if service == "consul" {
				continue
			}
			s.sm.RLock()
			_, ok := s.services[service]
			s.sm.RUnlock()
			// If we see a new service, start a specific watcher for its healthy instances.
			if !ok {
				s.instanceWatch(service)
			}
		}
		// Cleanup watchers for services that have been removed from the catalog.
		s.sm.Lock()
		for service, plan := range s.services {
			if _, ok := d[service]; !ok {
				time.Sleep(time.Second) // Small delay to avoid race conditions.
				plan.Stop()
				delete(s.services, service)
			}
		}
		s.sm.Unlock()
	default:
		logger.Error("can not decide the wath type, must be map[string][]string")

	}
}

func (s *serviceWatch) instanceWatch(service string) {
	pl, err := newInstanceWatch(s.c, service)
	if err != nil {
		logger.Error("new instance watch fail", "err", err)
		return
	}
	s.sm.Lock()
	s.services[service] = pl
	s.sm.Unlock()
}

// instanceWatch manages long-polling watches for specific service instances.
type instanceWatch struct {
	instances map[string]uint64 // Tracks ModifyIndex to avoid redundant updates.
	im        sync.RWMutex
	c         *Consul
	service   string
}

func newInstanceWatch(c *Consul, service string) (*watch.Plan, error) {
	watcher := &instanceWatch{
		instances: make(map[string]uint64),
		c:         c,
		service:   service,
	}
	pl, err := watch.Parse(map[string]any{
		"type":    "service",
		"service": service,
	})
	if err != nil {
		return nil, err
	}
	pl.Handler = watcher.handler
	go func() {
		if err := pl.RunWithClientAndHclog(watcher.c.client, nil); err != nil {
			logger.Error("consul watch service instance fail", "service", service)
		}
	}()
	return pl, nil
}

// handler processes updates from a specific service's instance list.
func (s *instanceWatch) handler(_ uint64, data any) {
	switch d := data.(type) {
	case []*api.ServiceEntry:
		// 1. Check for new or updated instances.
		for _, entry := range d {
			s.im.Lock()
			// Only trigger callback if it's a new ID or the data has been modified in Consul.
			if modifyIndex, ok := s.instances[entry.Service.ID]; !ok || modifyIndex != entry.Service.ModifyIndex {
				s.instances[entry.Service.ID] = entry.Service.ModifyIndex
				var service server.Service
				if err := json.Unmarshal([]byte(entry.Service.Meta["endpoints"]), &service.Endpoints); err != nil {
					logger.Error("consul unmarshal appDetail fail", "endpoints", entry.Service.Meta["endpoints"], "err", err)
					continue
				}
				if err := json.Unmarshal([]byte(entry.Service.Meta["app_detail"]), &service.APP); err != nil {
					logger.Error("consul unmarshal app fail", "app_detail", entry.Service.Meta["app_detail"], "err", err)
					continue
				}
				// Trigger the framework's discovery callback (used by load balancers).
				s.c.discoveryOptions.Callback(&registry.Event{
					Type: registry.EventTypeCreate,
					Instance: &registry.Instance{
						DiscoverName: NAME,
						Service:      &service,
					},
				})
			}
			s.im.Unlock()
		}

		// 2. Check for deleted instances.
		s.im.Lock()
		for key := range s.instances {
			exist := false
			for _, entry := range d {
				if entry.Service.ID == key {
					exist = true
					break
				}
			}
			if !exist {
				// Instance is gone from Consul, notify the framework to remove it from the pool.
				s.c.discoveryOptions.Callback(&registry.Event{
					Type: registry.EventTypeDelete,
					Instance: &registry.Instance{
						DiscoverName: NAME,
						Service: &server.Service{
							APP: runtime.APP{
								Instance: runtime.Instance{
									ID: key,
								},
							},
						},
					},
				})
				delete(s.instances, key)
			}
		}
		s.im.Unlock()
	default:
		logger.Error("consul instance watch fail, invalid type, must be []*api.ServiceEntry", "service", s.service, "type", fmt.Sprintf("%T", data))
	}
}
