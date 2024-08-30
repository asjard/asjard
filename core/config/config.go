package config

import (
	"sort"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/constant"
)

// Configer 配置存储需要实现的方法
type Configer interface {
	// 根据key获取配置
	Get(key string) (*Value, bool)
	// 获取所有配置
	GetAll() map[string]*Value
	// 根据前缀获取所有配置
	GetAllWithPrefixs(prefixs ...string) map[string]*Value
	// 设置配置
	Set(key string, value *Value)
	// 删除配置
	Del(key string)
}

// SourcesConfiger 配置源配置需要实现的方法
type SourcesConfiger interface {
	// 配置源获取配置
	Get(sourceName, key string) (*Value, bool)
	// 配置源设置配置
	Set(sourceName, key string, value *Value) bool
	// 配置源删除配置
	Del(sourceName, key, ref string, priority int)
}

// SourceConfiger 配置源的配置需要实现的方法
type SourceConfiger interface {
	// 获取配置源的配置
	Get(key string) (*Value, bool)
	// 设置配置源的配置
	Set(key string, value *Value) bool
	// 删除配置源的配置
	Del(key, ref string, priority int)
}

// Configs 全局配置维护
type Configs struct {
	cfgs map[string]*Value
	m    sync.RWMutex
}

var _ Configer = &Configs{}

// Get 获取配置
func (c *Configs) Get(key string) (*Value, bool) {
	c.m.RLock()
	value, ok := c.cfgs[key]
	c.m.RUnlock()
	return value, ok
}

// GetAll 获取所有配置
func (c *Configs) GetAll() map[string]*Value {
	c.m.RLock()
	cfgs := make(map[string]*Value, len(c.cfgs))
	for key, value := range c.cfgs {
		cfgs[key] = value
	}
	c.m.RUnlock()
	return cfgs
}

// GetAllWithPrefixs 根据前缀获取所有配置,并删除前缀
func (c *Configs) GetAllWithPrefixs(prefixs ...string) map[string]*Value {
	c.m.RLock()
	cfgs := make(map[string]*Value)
	for key, value := range c.cfgs {
		for _, p := range prefixs {
			if strings.HasPrefix(key, p) {
				cfgs[strings.TrimPrefix(key, p+constant.ConfigDelimiter)] = value
			}
		}
	}
	c.m.RUnlock()
	return cfgs
}

// Set 设置配置
func (c *Configs) Set(key string, value *Value) {
	c.m.Lock()
	c.cfgs[key] = value
	c.m.Unlock()
}

// Del 删除配置
func (c *Configs) Del(key string) {
	c.m.Lock()
	delete(c.cfgs, key)
	c.m.Unlock()
}

// SourcesConfig 配置源配置
type SourcesConfig struct {
	sources map[string]SourceConfiger
	m       sync.RWMutex
}

var _ SourcesConfiger = &SourcesConfig{}

// Get 配置源获取配置
func (c *SourcesConfig) Get(sourceName, key string) (*Value, bool) {
	c.m.RLock()
	configs, ok := c.sources[sourceName]
	c.m.RUnlock()
	if !ok {
		return nil, false
	}
	return configs.Get(key)
}

// Set 配置源设置配置
func (c *SourcesConfig) Set(sourceName, key string, value *Value) bool {
	c.m.Lock()
	defer c.m.Unlock()
	configs, ok := c.sources[sourceName]
	if !ok {
		c.sources[sourceName] = &SourceConfigs{
			cfgs: map[string][]*Value{key: {value}},
		}
		return true
	}
	return configs.Set(key, value)
}

// Del 配置源删除配置
func (c *SourcesConfig) Del(sourceName, key, ref string, priority int) {
	c.m.Lock()
	configs, ok := c.sources[sourceName]
	if ok {
		configs.Del(key, ref, priority)
	}
	c.m.Unlock()
}

// SourceConfigs 配置源的配置
type SourceConfigs struct {
	cfgs map[string][]*Value
	m    sync.RWMutex
}

var _ SourceConfiger = &SourceConfigs{}

// Get 获取配置源配置
func (c *SourceConfigs) Get(key string) (*Value, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	values, ok := c.cfgs[key]
	if !ok || len(values) == 0 {
		return nil, false
	}
	return values[0], true
}

// Set 配置源设置配置
func (c *SourceConfigs) Set(key string, value *Value) bool {
	c.m.Lock()
	defer c.m.Unlock()
	values, ok := c.cfgs[key]
	if !ok || len(values) == 0 {
		c.cfgs[key] = []*Value{value}
		return true
	}
	// 同优先级, 后来的覆盖先来的
	newValues := make([]*Value, 0, len(values))
	for _, v := range values {
		if v.Priority != value.Priority {
			newValues = append(newValues, v)
		}
	}
	if len(newValues) == 0 {
		c.cfgs[key] = []*Value{value}
		return true
	}
	setted := value.Priority > newValues[0].Priority
	newValues = append(newValues, value)
	// 从大到小重新排序
	sort.Slice(newValues, func(i, j int) bool {
		return newValues[i].Priority > newValues[j].Priority
	})
	c.cfgs[key] = newValues
	return setted
}

// Del 配置源删除配置
func (c *SourceConfigs) Del(key, ref string, priority int) {
	c.m.Lock()
	if key != "" {
		values, ok := c.cfgs[key]
		if ok {
			if priority < 0 {
				delete(c.cfgs, key)
			} else {
				newValues := make([]*Value, 0, len(values))
				for _, v := range values {
					if v.Priority != priority {
						newValues = append(newValues, v)
					}
				}
				if len(newValues) == 0 {
					delete(c.cfgs, key)
				} else {
					c.cfgs[key] = newValues
				}
			}
		}
	}
	if ref != "" {
		for key, values := range c.cfgs {
			newValues := make([]*Value, 0, len(values))
			for _, v := range values {
				if v.Ref == ref && (priority < 0 || priority == v.Priority) {
					continue
				}
				newValues = append(newValues, v)
			}
			if len(newValues) == 0 {
				delete(c.cfgs, key)
			} else {
				c.cfgs[key] = newValues
			}
		}
	}
	c.m.Unlock()
}
