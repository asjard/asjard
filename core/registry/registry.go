package registry

import (
	"fmt"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// Registry 服务注册和发现需要实现的接口
type Registry interface {
	Discovery
	Register
}

// RegistryManager 服务注册和发现管理
type RegistryManager struct {
	cache *cache

	// 当前服务
	currentService *server.Service
	// 配置
	conf *Config

	// 注册中心列表
	registers []Register
	// 服务发现中心列表
	discovers []Discovery
}

var registryManager *RegistryManager

// 初始化服务发现与注册中心
func init() {
	registryManager = &RegistryManager{}
}

// Init 服务注册发现初始化
func Init() error {
	registryManager.currentService = server.GetService()
	registryManager.conf = GetConfig()
	registryManager.cache = newCache(registryManager.conf, registryManager.healthCheck)
	return nil
}

func (r *RegistryManager) registe() error {
	if !r.conf.AutoRegiste {
		return nil
	}
	for name, newRegister := range newRegisters {
		for _, registerName := range r.conf.Registers {
			if name == registerName {
				register, err := newRegister()
				if err != nil {
					return err
				}
				r.registers = append(r.registers, register)
				break
			}
		}
	}
	// 延迟注册
	if r.conf.DelayRegiste.Duration != 0 {
		return r.delayRegiste(r.conf.DelayRegiste.Duration)
	}
	return r.doRegiste()
}

func (r *RegistryManager) delayRegiste(duration time.Duration) error {
	// 延迟注册
	go func(duration time.Duration) {
		t := time.After(duration)
		<-t
		r.doRegiste()
	}(duration)
	return nil
}

// 注册当前服务到注册中心
func (r *RegistryManager) doRegiste() error {
	for _, register := range r.registers {
		if err := register.Registe(r.currentService); err != nil {
			return err
		}
	}
	return nil
}

// 注册中心心跳
// 当开启了心跳后，心跳时间向所有注册中心发起心跳
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

// 向注册中心发起心跳表示本服务还存活
// TODO 添加超时逻辑
func (r *RegistryManager) doHeartbeat() {
	// for _, register := range r.registers {
	// register.Heartbeat(r.currentInstance)
	// }
}

// 从注册中心删除本服务
func (r *RegistryManager) remove() error {
	if !r.conf.AutoRegiste {
		return nil
	}
	for _, register := range r.registers {
		register.Remove(r.currentService)
	}
	return nil
}

// 自动发现服务
func (r *RegistryManager) discove() error {
	if !r.conf.AutoDiscove {
		logger.Warn("registry.autoDiscove not enabled")
		return nil
	}
	for name, newDiscover := range newDiscoverys {
		for _, discoverName := range r.conf.Discovers {
			if name == discoverName {
				discover, err := newDiscover()
				if err != nil {
					return err
				}
				r.discovers = append(r.discovers, discover)
				break
			}
		}
	}
	for _, discover := range r.discovers {
		services, err := discover.GetAll()
		if err != nil {
			return err
		}
		r.cache.update(services)
		discover.Watch(r.watch)
	}
	return nil
}

func (r *RegistryManager) healthCheck(discoverName string, instance *server.Service) error {
	for _, discover := range r.discovers {
		if discover.Name() == discoverName {
			// return discover.HealthCheck(instance)
		}
	}
	return fmt.Errorf("service '%s(%s)' health check discover '%s' not found",
		instance.Instance.Name, instance.Instance.ID, discoverName)
}

// 服务变化更新
func (r *RegistryManager) watch(et *Event) {
	switch et.Type {
	case EventTypeCreate, EventTypeUpdate:
		r.update(et)
	case EventTypeDelete:
		r.delete(et)
	}
}

// 更新服务
func (r *RegistryManager) update(event *Event) {
	r.cache.update([]*Instance{event.Instance})
}

// 删除服务
func (r *RegistryManager) delete(event *Event) {
	r.cache.delete(event.Instance)
}

func (r *RegistryManager) pick(options *Options) []*Instance {
	return r.cache.pick(options)
}

func (r *RegistryManager) removeListener(name string) {
	r.cache.removeListener(name)
}
