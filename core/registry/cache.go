package registry

import (
	"sync"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/server"
)

// 服务健康方法
type healthCheckFunc func(discoverName string, instance *server.Instance) error

// ServiceInstance 带发现者的服务实例详情
type ServiceInstance struct {
	// 服务发现者
	DiscoverName string
	// 实例详情
	Instance *server.Instance
}

// 服务发现缓存
type cache struct {
	// 服务列表
	// TODO 需要更新存储结构
	services []*ServiceInstance

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
}

type serviceCache struct {
	// key: 服务名称
	// value: 相同服务名称的服务实例列表
	// 一个注册中心中相同的服务不应该重复
	// 不同注册中心可以出现相同的服务
	services map[string][]*server.Instance
	sm       sync.RWMutex
}

// 初始化一个本地缓存用以维护发现的服务实例
func newCache(hf healthCheckFunc) *cache {
	c := &cache{
		failureThreshold:  config.GetInt("registry.failureThreshold", 1),
		healthCheckFunc:   hf,
		failureThresholds: map[string]int{},
	}
	if config.GetBool("registry.healthCheck", true) {
		go c.healthCheck()
	}
	return c
}

func (instance *ServiceInstance) canPick(options *Options) bool {
	for _, pickFunc := range options.getPickFuncs() {
		if !pickFunc(instance) {
			return false
		}
	}
	return true
}

// 获取服务实例
func (c *cache) pick(options *Options) []*server.Instance {
	var services []*server.Instance
	for _, service := range c.services {
		if service.canPick(options) {
			services = append(services, service.Instance)
		}
	}
	return services
}

// 更新本地缓存中的服务实例
func (c *cache) update(discoverName string, instances []*server.Instance) {
	for _, svc := range c.services {
		for _, instance := range instances {
			if instance.ID == svc.Instance.ID {
				svc.DiscoverName = discoverName
				svc.Instance = instance
			}
		}
	}
}

// 从本地缓存中删除服务实例
func (c *cache) delete(instance *server.Instance) {
	for index, svc := range c.services {
		if svc.Instance.ID == instance.ID {
			c.services = append(c.services[:index], c.services[index+1:]...)
		}
	}
}

// 服务健康检查
func (c *cache) healthCheck() {
	duration, err := time.ParseDuration(config.GetString("registry.healthCheckInterval", "10s"))
	if err != nil {
		duration = 10 * time.Second
	}
	ticker := time.NewTicker(duration)
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

func (c *cache) newServiceCache(_ string) *serviceCache {
	serviceCache := &serviceCache{
		services: make(map[string][]*server.Instance),
	}
	// c.sm.Lock()
	// c.discoverServices[discoverName] = serviceCache
	// c.sm.Unlock()
	return serviceCache
}

func (c *serviceCache) addOrUpdate(instances []*server.Instance) {
	for _, instance := range instances {
		exist := false
		c.sm.RLock()
		services, ok := c.services[instance.Name]
		c.sm.RUnlock()
		if ok {
			for index, service := range services {
				if service.ID == instance.ID {
					logger.Debugf("update instance %s(%s)", instance.Name, instance.ID)
					exist = true
					c.sm.Lock()
					c.services[instance.Name][index] = instance
					c.sm.Unlock()
					break
				}
			}
		}
		if !ok || !exist {
			logger.Debugf("add instance %s(%s)", instance.Name, instance.ID)
			c.sm.Lock()
			c.services[instance.Name] = append(c.services[instance.Name], instance)
			c.sm.Unlock()
		}
	}
}

func (c *serviceCache) delete(instance *server.Instance) {
	c.sm.RLock()
	services, ok := c.services[instance.Name]
	c.sm.RUnlock()
	if ok {
		for index, service := range services {
			if service.ID == instance.ID {
				logger.Debugf("delete instance %s(%s)", service.Name, service.ID)
				// 删除该实例
				c.sm.Lock()
				c.services[instance.Name] = append(services[:index], services[index+1:]...)
				c.sm.Unlock()
				break
			}
		}
	}
}

func (c *serviceCache) healthCheck(discoverName string, hf healthCheckFunc) []*server.Instance {
	var notHealthInstances []*server.Instance
	c.sm.RLock()
	for _, instances := range c.services {
		for _, instance := range instances {
			if err := hf(discoverName, instance); err != nil {
				logger.Errorf("health check discover '%s' instance '%s(%s)' fail[%s]",
					discoverName, instance.Name, instance.ID, err.Error())
				notHealthInstances = append(notHealthInstances, instance)
			}
		}
	}
	c.sm.RUnlock()
	return notHealthInstances
}
