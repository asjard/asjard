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
	NAME                 = "consul"
	AddressTypeListen    = "listen"
	AddressTypeAdvertise = "advertise"
)

type Consul struct {
	client           *api.Client
	conf             *Config
	exit             chan struct{}
	discoveryOptions *registry.DiscoveryOptions
}

type Config struct {
	Client  string             `json:"client"`
	Timeout utils.JSONDuration `json:"timeout"`
}

var (
	_             registry.Register  = &Consul{}
	_             registry.Discovery = &Consul{}
	defaultConfig                    = Config{
		Client:  consul.DefaultClientName,
		Timeout: utils.JSONDuration{Duration: 5 * time.Second},
	}

	newConsul *Consul
	newOnce   sync.Once
)

func init() {
	registry.AddRegister(NAME, NewRegister)
	registry.AddDiscover(NAME, NewDiscovery)
}

// NewRegister 服务注册初始化
func NewRegister() (registry.Register, error) {
	return New(nil)
}

// NewDiscovery 服务发现初始化
func NewDiscovery(options *registry.DiscoveryOptions) (registry.Discovery, error) {
	discovery, err := New(options)
	if err != nil {
		return nil, err
	}
	if err := newServiceWatch(discovery); err != nil {
		return nil, err
	}
	return discovery, nil
}

// New consul初始化
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

func (c *Consul) Registe(service *server.Service) error {
	appDetail, err := json.Marshal(&service.APP)
	if err != nil {
		return err
	}
	endpoints, err := json.Marshal(&service.Endpoints)
	if err != nil {
		return err
	}
	ttl := c.conf.Timeout.Duration + time.Second
	registration := &api.AgentServiceRegistration{
		ID:   service.Instance.ID,
		Name: service.Instance.Name,
		Meta: map[string]string{
			"app":        service.App,
			"app_detail": string(appDetail),
			"endpoints":  string(endpoints),
		},
		Check: &api.AgentServiceCheck{
			CheckID:                        service.Instance.ID,
			Name:                           service.Instance.Name,
			TTL:                            ttl.String(),
			Status:                         "passing",
			DeregisterCriticalServiceAfter: c.conf.Timeout.String(),
		},
	}
	if err := c.client.Agent().ServiceRegister(registration); err != nil {
		return err
	}
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

func (c *Consul) Remove(service *server.Service) {
	c.exit <- struct{}{}
	if err := c.client.Agent().ServiceDeregister(service.Instance.ID); err != nil {
		logger.Error("remove instance fail", "err", err)
	}
}

func (c *Consul) Name() string { return NAME }

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

func (c *Consul) getAgentServices() (map[string]*server.Service, error) {
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

type serviceWatch struct {
	services map[string]*watch.Plan
	sm       sync.RWMutex
	c        *Consul
}

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
			if !ok {
				s.instanceWatch(service)
			}
		}
		s.sm.Lock()
		for service, plan := range s.services {
			if _, ok := d[service]; !ok {
				time.Sleep(time.Second)
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

type instanceWatch struct {
	instances map[string]uint64
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

func (s *instanceWatch) handler(_ uint64, data any) {
	switch d := data.(type) {
	case []*api.ServiceEntry:
		for _, entry := range d {
			s.im.Lock()
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
