package config

import (
	"sort"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/constant"
)

// configs 配置维护
type configs struct {
	cfgs map[string]*Value
	m    sync.RWMutex
}

func newConfigs() *configs {
	return &configs{
		cfgs: make(map[string]*Value),
	}
}

func (c *configs) get(key string) (*Value, bool) {
	c.m.RLock()
	value, ok := c.cfgs[key]
	c.m.RUnlock()
	return value, ok
}

func (c *configs) getAll() map[string]*Value {
	c.m.RLock()
	cfgs := make(map[string]*Value, len(c.cfgs))
	for key, value := range c.cfgs {
		cfgs[key] = value
	}
	c.m.RUnlock()
	return cfgs
}

// 根据前缀获取所有配置,并删除前缀
func (c *configs) getAllWithPrefixs(prefixs ...string) map[string]*Value {
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

func (c *configs) set(key string, value *Value) {
	c.m.Lock()
	c.cfgs[key] = value
	c.m.Unlock()
}

func (c *configs) del(key string) {
	c.m.Lock()
	delete(c.cfgs, key)
	c.m.Unlock()
}

// 配置源配置
type sourcesConfigs struct {
	sources map[string]*sourceConfig
	m       sync.RWMutex
}

type sourceConfig struct {
	cfgs map[string][]*Value
	m    sync.RWMutex
}

func newSourcesConfigs() *sourcesConfigs {
	return &sourcesConfigs{
		sources: make(map[string]*sourceConfig),
	}
}

func (c *sourcesConfigs) get(sourceName, key string) (*Value, bool) {
	c.m.RLock()
	configs, ok := c.sources[sourceName]
	c.m.RUnlock()
	if !ok {
		return nil, false
	}
	return configs.get(key)
}

func (c *sourcesConfigs) set(sourceName, key string, value *Value) bool {
	c.m.Lock()
	defer c.m.Unlock()
	configs, ok := c.sources[sourceName]
	if !ok {
		c.sources[sourceName] = &sourceConfig{
			cfgs: map[string][]*Value{key: {value}},
		}
		return true
	}
	return configs.set(key, value)
}

func (c *sourcesConfigs) del(sourceName, key, ref string, priority int) {
	c.m.Lock()
	configs, ok := c.sources[sourceName]
	if ok {
		configs.del(key, ref, priority)
	}
	c.m.Unlock()
}

func (c *sourceConfig) get(key string) (*Value, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	values, ok := c.cfgs[key]
	if !ok || len(values) == 0 {
		return nil, false
	}
	return values[0], true
}

func (c *sourceConfig) set(key string, value *Value) bool {
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

func (c *sourceConfig) del(key, ref string, priority int) {
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
