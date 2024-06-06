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
	// LocalRegistryName 本地注册发现中心名称
	LocalRegistryName = "local"
)

// Discovery 服务发现相关功能
type Discovery interface {
	// 获取所有服务实例
	GetAll() ([]*Instance, error)
	// 监听服务变化
	Watch(callbak func(event *Event))
	// 服务发现中心名称
	Name() string
}

// NewDiscoveryFunc 服务发现
type NewDiscoveryFunc func() (Discovery, error)

var newDiscoverys []NewDiscoveryFunc

func init() {
	// 添加本地服务发现
	AddDiscover(NewLocalDiscover)
}

// AddDiscover 添加服务发现组件
func AddDiscover(newFunc NewDiscoveryFunc) error {
	newDiscoverys = append(newDiscoverys, newFunc)
	return nil
}

// PickServices 获取服务列表
// 从本地缓存中获取符合要求的服务实例
func PickServices(opts ...Option) []*Instance {
	options := NewOptions(opts)
	if len(opts) == 0 {
		options = DefaultOptions()
	}
	return registryManager.pick(options)
}

// RemoveListener 移除监听
func RemoveListener(name string) {
	registryManager.removeListener(name)
}

// LocalRegistry 本地服务发现
type LocalRegistry struct {
	cb func(event *Event)
	// key: serviceName
	instances               []*Instance
	localDiscoverConfPrefix string
	dm                      sync.RWMutex
}

// NewLocalDiscover .
func NewLocalDiscover() (Discovery, error) {
	localDiscover := &LocalRegistry{
		localDiscoverConfPrefix: "registry.localDiscover",
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
		config.WithMatchWatch(l.localDiscoverConfPrefix+".*", l.watch)); err != nil {
		logger.Error("get registry.localDiscover fail",
			"err", err.Error())
	} else {
		l.instances = l.getInstances(services)
	}

}
func (l *LocalRegistry) watch(event *config.Event) {
	services := make(map[string][]string)
	config.GetWithUnmarshal(l.localDiscoverConfPrefix, &services)
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
		instance := server.NewInstance()
		instance.Name = name
		endpoints := make(map[string][]string)
		for index := range addresses {
			u, err := url.Parse(addresses[index])
			if err == nil {
				endpoints[u.Scheme] = append(endpoints[u.Scheme], u.Host)
			}
		}
		for protocol, addresses := range endpoints {
			if err := instance.AddEndpoints(protocol, map[string][]string{
				constant.ServerListenAddressName: addresses,
			}); err != nil {
				logger.Error(err.Error())
			}
		}
		instances = append(instances, &Instance{
			DiscoverName: l.Name(),
			Instance:     instance,
		})
	}
	return instances
}
