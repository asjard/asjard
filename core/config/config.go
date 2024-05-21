package config

import (
	"sync"
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

// 配置源配置
type sourcesConfigs struct {
	sources map[string]*configs
	m       sync.RWMutex
}

func newSourcesConfigs() *sourcesConfigs {
	return &sourcesConfigs{
		sources: make(map[string]*configs),
	}
}

func (c *configs) get(key string) (*Value, bool) {
	c.m.RLock()
	value, ok := c.cfgs[key]
	c.m.RUnlock()
	return value, ok
}

func (c *configs) getAll() map[string]*Value {
	cfgs := make(map[string]*Value)
	c.m.RLock()
	for key, value := range c.cfgs {
		cfgs[key] = value
	}
	c.m.RUnlock()
	return cfgs
}

func (c *configs) set(key string, value *Value) {
	c.m.Lock()
	c.cfgs[key] = value
	c.m.Unlock()
}

func (c *configs) del(key, ref string) {
	c.m.Lock()
	if key != "" {
		delete(c.cfgs, key)
	}
	if ref != "" {
		for key, value := range c.cfgs {
			if value.Ref == ref {
				delete(c.cfgs, key)
			}
		}
	}
	c.m.Unlock()
}

func (c *sourcesConfigs) get(sourceName, key string) (*Value, bool) {
	cfgs, ok := c.getConfigs(sourceName)
	if !ok {
		return nil, false
	}
	return cfgs.get(key)
}

func (c *sourcesConfigs) getConfigs(sourceName string) (*configs, bool) {
	c.m.RLock()
	cfgs, ok := c.sources[sourceName]
	c.m.RUnlock()
	return cfgs, ok
}

func (c *sourcesConfigs) set(sourceName, key string, value *Value) {
	c.m.Lock()
	if _, ok := c.sources[sourceName]; !ok {
		c.sources[sourceName] = newConfigs()
	}
	c.sources[sourceName].set(key, value)
	c.m.Unlock()
}

func (c *sourcesConfigs) del(sourceName, key, ref string) {
	cfgs, ok := c.getConfigs(sourceName)
	if !ok {
		return
	}
	cfgs.del(key, ref)
}
