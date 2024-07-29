package registry

import (
	"net/url"
	"sync"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

const (
	LocalDiscoverName = "localDiscover"
)

// LocalRegistry 本地服务发现
type LocalRegistry struct {
	cb func(event *Event)
	// key: serviceName
	instances               []*Instance
	localDiscoverConfPrefix string
	dm                      sync.RWMutex
}

func init() {
	// 添加本地服务发现
	AddDiscover(LocalDiscoverName, NewLocalDiscover)
}

// NewLocalDiscover .
func NewLocalDiscover() (Discovery, error) {
	localDiscover := &LocalRegistry{
		localDiscoverConfPrefix: constant.ConfigRegistryLocalDiscoverPrefix,
	}
	localDiscover.getAndWatch()
	return localDiscover, nil
}

// GetAll 获取所有服务列表
func (l *LocalRegistry) GetAll() ([]*Instance, error) {
	return l.instances, nil
}

// Watch 监听配置变化
func (l *LocalRegistry) Watch(callback func(event *Event)) {
	l.cb = callback
}

// Name 返回本地注册中心名称
func (l *LocalRegistry) Name() string {
	return LocalRegistryName
}

func (l *LocalRegistry) getAndWatch() {
	services := make(map[string][]string)
	if err := config.GetWithUnmarshal(l.localDiscoverConfPrefix,
		&services,
		config.WithWatch(l.watch)); err != nil {
		logger.Error("get registry.localDiscover fail",
			"err", err.Error())
	} else {
		l.instances = l.getInstances(services)
	}

}
func (l *LocalRegistry) watch(event *config.Event) {
	services := make(map[string][]string)
	if err := config.GetWithUnmarshal(l.localDiscoverConfPrefix, &services); err != nil {
		logger.Error("get local discover conf fail", "err", err)
	}
	instances := l.getInstances(services)
	for _, instance := range l.instances {
		l.cb(&Event{
			Type:     EventTypeDelete,
			Instance: instance,
		})
	}

	for _, instance := range instances {
		l.cb(&Event{
			Type:     EventTypeUpdate,
			Instance: instance,
		})

	}
	l.instances = instances
}

func (l *LocalRegistry) getInstances(services map[string][]string) []*Instance {
	var instances []*Instance
	for name, addresses := range services {
		service := server.NewService()
		service.Instance.Name = name
		for index := range addresses {
			u, err := url.Parse(addresses[index])
			if err == nil {
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
