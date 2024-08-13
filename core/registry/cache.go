package registry

import (
	"sync"
	"time"

	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// 服务健康方法
type healthCheckFunc func(discoverName string, instance *server.Service) error

// Instance 带发现者的服务实例详情
type Instance struct {
	// 服务发现者
	DiscoverName string
	// 实例详情
	Service *server.Service
}

// 服务发现缓存
type cache struct {
	// 服务列表
	services map[string]*Instance
	sm       sync.RWMutex

	conf *Config

	// 健康检查方法
	healthCheckFunc healthCheckFunc
	// 连续失败次数
	// 当一个服务健康检查失败次数超过此阈值时从本地缓存列表中删除此服务
	failureThreshold int
	// 健康检查失败记录
	// key: discoverName + serviceName + serviceID
	// value: 失败次数
	failureThresholds map[string]int
	fm                sync.RWMutex
	listeners         map[string]*listener
	lm                sync.RWMutex
}

type listener struct {
	options  *Options
	callback func(*Event)
}

// 初始化一个本地缓存用以维护发现的服务实例
func newCache(conf *Config, hf healthCheckFunc) *cache {
	c := &cache{
		services:          make(map[string]*Instance),
		failureThreshold:  conf.FailureThreshold,
		healthCheckFunc:   hf,
		failureThresholds: map[string]int{},
		listeners:         map[string]*listener{},
		conf:              conf,
	}
	if conf.HealthCheck {
		go c.healthCheck()
	}
	return c
}

func (instance *Instance) canPick(options *Options) bool {
	for _, pickFunc := range options.getPickFuncs() {
		if !pickFunc(instance) {
			return false
		}
	}
	return true
}

// 获取服务实例
func (c *cache) pick(options *Options) []*Instance {
	var instances []*Instance
	c.sm.RLock()
	defer c.sm.RUnlock()
	for _, instance := range c.services {
		if instance.canPick(options) {
			instances = append(instances, instance)
		}
	}
	c.addListener(options)
	return instances
}

func (c *cache) addListener(options *Options) {
	if options.watchName != "" && options.watch != nil {
		c.lm.Lock()
		c.listeners[options.watchName] = &listener{
			options:  options,
			callback: options.watch,
		}
		c.lm.Unlock()
	}
}

func (c *cache) removeListener(listenerName string) {
	c.lm.Lock()
	for name := range c.listeners {
		if name == listenerName {
			delete(c.listeners, name)
		}
	}
	c.lm.Unlock()
}

// 更新本地缓存中的服务实例
func (c *cache) update(instances []*Instance) {
	c.sm.Lock()
	defer c.sm.Unlock()
	for _, instance := range instances {
		logger.Debug("update instance", "instance", instance.Service.Instance.ID)
		c.services[instance.Service.Instance.ID] = instance
		c.notify(EventTypeUpdate, instance)
	}
}

// 从本地缓存中删除服务实例
func (c *cache) delete(instance *Instance) {
	logger.Debug("delete instance",
		"instance", instance.Service.Instance.ID)
	c.sm.Lock()
	if svc, ok := c.services[instance.Service.Instance.ID]; ok {
		delete(c.services, instance.Service.Instance.ID)
		c.notify(EventTypeDelete, svc)
	}
	c.sm.Unlock()
}

func (c *cache) notify(eventType EventType, instance *Instance) {
	c.lm.RLock()
	for _, listener := range c.listeners {
		if instance.canPick(listener.options) {
			listener.options.watch(&Event{
				Type:     eventType,
				Instance: instance,
			})
		}
	}
	c.lm.RUnlock()
}

// 服务健康检查
func (c *cache) healthCheck() {
	ticker := time.NewTicker(c.conf.HealthCheckInterval.Duration)
	for {
		select {
		case <-ticker.C:
			c.doHealthCheck()
		}
	}
}

func (c *cache) doHealthCheck() {
	// c.sm.RLock()
	// for discoverName, service := range c.discoverServices {
	// 	notHealthInstances := service.healthCheck(discoverName, c.healthCheckFunc)
	// 	for _, instance := range notHealthInstances {
	// 		failKey := fmt.Sprintf("%s:%s:%s", discoverName, instance.Name, instance.ID)
	// 		threshold := c.getFailureThreshold(failKey)
	// 		if threshold >= c.failureThreshold {
	// 			// 移除该服务实例
	// 			service.delete(instance)
	// 			// 移除失败次数记录
	// 			delete(c.failureThresholds, failKey)
	// 		} else {
	// 			c.setFailureThreshold(failKey, threshold+1)
	// 		}
	// 	}
	// }
	// c.sm.RUnlock()
}

func (c *cache) getFailureThreshold(failKey string) int {
	c.fm.RLock()
	threshold, ok := c.failureThresholds[failKey]
	c.fm.RUnlock()
	if ok {
		return threshold
	}
	return 1
}

func (c *cache) setFailureThreshold(failKey string, threshold int) {
	c.fm.Lock()
	c.failureThresholds[failKey] = threshold
	c.fm.Unlock()
}
