package registry

import (
	"fmt"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/constant"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// Registry .
type Registry struct {
	cache *cache

	// 当前服务实例
	currentInstance *server.Instance

	// 注册中心列表
	registers []Register
	// 服务发现中心列表
	discovers []Discovery
}

var registryManager *Registry

// 初始化服务发现与注册中心
func init() {
	registryManager = &Registry{}
}

// Init 服务注册中心初始化
// 只发现服务，不注册服务，等服务启动后再注册服务
func Init() error {
	registryManager.currentInstance = server.GetInstance()
	registryManager.cache = newCache(registryManager.healthCheck)
	for _, newRegister := range newRegisters {
		register, err := newRegister()
		if err != nil {
			return err
		}
		registryManager.registers = append(registryManager.registers, register)
	}
	for _, newDiscover := range newDiscoverys {
		discover, err := newDiscover()
		if err != nil {
			return err
		}
		registryManager.discovers = append(registryManager.discovers, discover)
	}
	return registryManager.discove()
}

func (r *Registry) registe() error {
	// 延迟注册
	if delay := config.GetString(constant.ConfigRegistryDelayRegiste, ""); delay != "" {
		return r.delayRegiste(delay)
	}
	return r.doRegiste()
}

func (r *Registry) delayRegiste(delay string) error {
	duration, err := time.ParseDuration(delay)
	if err != nil {
		return fmt.Errorf("parse registry.delayRegiste fail[%s]", err.Error())
	}
	// 延迟注册
	go func(duration time.Duration) {
		t := time.After(duration)
		<-t
		r.doRegiste()
	}(duration)
	return nil
}

// 注册当前服务到注册中心
func (r *Registry) doRegiste() error {
	for _, register := range r.registers {
		if err := register.Registe(r.currentInstance); err != nil {
			return err
		}
	}
	return nil
}

// 注册中心心跳
// 当开启了心跳后，心跳时间向所有注册中心发起心跳
func (r *Registry) heartbeat() error {
	duration, err := time.ParseDuration(config.GetString(constant.ConfigRegistryHeartbeatInterval, "5s"))
	if err != nil {
		return err
	}
	go func(duration time.Duration) {
		ticker := time.NewTicker(duration)
		for {
			<-ticker.C
			r.doHeartbeat()
		}
	}(duration)
	return nil
}

// 向注册中心发起心跳表示本服务还存活
// TODO 添加超时逻辑
func (r *Registry) doHeartbeat() {
	// for _, register := range r.registers {
	// register.Heartbeat(r.currentInstance)
	// }
}

// 从注册中心删除本服务
func (r *Registry) remove() error {
	for _, register := range r.registers {
		register.Remove(r.currentInstance)
	}
	return nil
}

// 自动发现服务
func (r *Registry) discove() error {
	if !config.GetBool(constant.CofigRegistryAutoDiscove, false) {
		logger.Warn("registry.autoDiscove not enabled")
		return nil
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

func (r *Registry) healthCheck(discoverName string, instance *server.Instance) error {
	for _, discover := range r.discovers {
		if discover.Name() == discoverName {
			// return discover.HealthCheck(instance)
		}
	}
	return fmt.Errorf("service '%s(%s)' health check discover '%s' not found",
		instance.Name, instance.ID, discoverName)
}

// 服务变化更新
func (r *Registry) watch(et *Event) {
	switch et.Type {
	case EventTypeCreate, EventTypeUpdate:
		r.update(et)
	case EventTypeDelete:
		r.delete(et)
	}
}

// 更新服务
func (r *Registry) update(event *Event) {
	r.cache.update([]*Instance{event.Instance})
}

// 删除服务
func (r *Registry) delete(event *Event) {
	r.cache.delete(event.Instance)
}

func (r *Registry) pick(options *Options) []*Instance {
	return r.cache.pick(options)
}

func (r *Registry) removeListener(name string) {
	r.cache.removeListener(name)
}
