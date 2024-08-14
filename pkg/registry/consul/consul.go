package consul

import (
	"encoding/json"
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
)

const (
	NAME                 = "consul"
	AddressTypeListen    = "listen"
	AddressTypeAdvertise = "advertise"
)

type Consul struct {
	client     *api.Client
	cb         func(event *registry.Event)
	conf       *Config
	exit       chan struct{}
	serviceMap map[string]*server.Service
	sm         sync.RWMutex
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

func NewRegister() (registry.Register, error) {
	return New()
}

func NewDiscovery() (registry.Discovery, error) {
	discovery, err := New()
	if err != nil {
		return nil, err
	}
	go discovery.watch()
	return discovery, nil
}

func New() (*Consul, error) {
	var err error
	newOnce.Do(func() {
		consulRegistry := &Consul{
			exit:       make(chan struct{}),
			serviceMap: make(map[string]*server.Service),
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
	c.sm.Lock()
	c.serviceMap = serviceMap
	c.sm.Unlock()
	instances := make([]*registry.Instance, 0, len(serviceMap))
	for _, service := range serviceMap {
		instances = append(instances, &registry.Instance{
			DiscoverName: NAME,
			Service:      service,
		})
	}
	return instances, nil
}

func (c *Consul) Watch(callback func(event *registry.Event)) {
	c.cb = callback
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

func (c *Consul) watch() {
	for {
		select {
		case <-time.After(c.conf.Timeout.Duration):
			serviceMap, err := c.getAgentServices()
			if err == nil {
				c.sm.Lock()
				for serviceId, service := range serviceMap {
					if _, ok := c.serviceMap[serviceId]; !ok {
						if c.cb != nil {
							c.cb(&registry.Event{
								Type: registry.EventTypeCreate,
								Instance: &registry.Instance{
									DiscoverName: NAME,
									Service:      service,
								},
							})
						}
					}
				}
				for serviceId, service := range c.serviceMap {
					if _, ok := serviceMap[serviceId]; !ok {
						if c.cb != nil {
							c.cb(&registry.Event{
								Type: registry.EventTypeDelete,
								Instance: &registry.Instance{
									DiscoverName: NAME,
									Service:      service,
								},
							})
						}
					}
				}
				c.serviceMap = serviceMap
				c.sm.Unlock()
			}
		}
	}
}
