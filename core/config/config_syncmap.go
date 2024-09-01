package config

import (
	"sort"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/constant"
)

// ConfigsWithSyncMap sync.Map实现全局配置
type ConfigsWithSyncMap struct {
	cfgs sync.Map
}

var _ Configer = &ConfigsWithSyncMap{}

// 根据key获取配置
func (c *ConfigsWithSyncMap) Get(key string) (*Value, bool) {
	v, ok := c.cfgs.Load(key)
	if ok {
		return v.(*Value), ok
	}
	return nil, ok
}

// 获取所有配置
func (c *ConfigsWithSyncMap) GetAll() map[string]*Value {
	cfgs := make(map[string]*Value)
	c.cfgs.Range(func(key, value any) bool {
		cfgs[key.(string)] = value.(*Value)
		return true
	})
	return cfgs
}

// 根据前缀获取所有配置
func (c *ConfigsWithSyncMap) GetAllWithPrefixs(prefixs ...string) map[string]*Value {
	cfgs := make(map[string]*Value)
	c.cfgs.Range(func(key, value any) bool {
		k := key.(string)
		for _, prefix := range prefixs {
			if strings.HasPrefix(k, prefix) {
				cfgs[strings.TrimPrefix(k, prefix+constant.ConfigDelimiter)] = value.(*Value)
			}
		}
		return true
	})
	return cfgs
}

// 设置配置
func (c *ConfigsWithSyncMap) Set(key string, value *Value) {
	c.cfgs.Store(key, value)
}

// 删除配置
func (c *ConfigsWithSyncMap) Del(key string) {
	c.cfgs.Delete(key)
}

// SourcesConfigWithSyncMap sync.Map实现的配置源配置
type SourcesConfigWithSyncMap struct {
	// map[sourceName]*configsSource
	sources sync.Map
}

var _ SourcesConfiger = &SourcesConfigWithSyncMap{}

// 配置源获取配置
func (c *SourcesConfigWithSyncMap) Get(sourceName, key string) (*Value, bool) {
	sourceConfigs, ok := c.sources.Load(sourceName)
	if !ok {
		return nil, false
	}
	return sourceConfigs.(SourceConfiger).Get(key)
}

// 配置源设置配置
func (c *SourcesConfigWithSyncMap) Set(sourceName, key string, value *Value) bool {
	sourceConfigs, ok := c.sources.Load(sourceName)
	if !ok {
		n := &SourceConfigsWithSyncMap{}
		c.sources.Store(sourceName, n)
		return n.Set(key, value)
	}
	return sourceConfigs.(SourceConfiger).Set(key, value)
}

// 配置源删除配置
func (c *SourcesConfigWithSyncMap) Del(sourceName, key, ref string, priority int) {
	sourceConfigs, ok := c.sources.Load(sourceName)
	if ok {
		sourceConfigs.(SourceConfiger).Del(key, ref, priority)
	}
}

// SourceConfigsWithSyncMap sync.Map实现配置源的配置
type SourceConfigsWithSyncMap struct {
	// map[key][]*Value
	cfgs sync.Map
}

// Get 获取配置源的配置
func (c *SourceConfigsWithSyncMap) Get(key string) (*Value, bool) {
	v, ok := c.cfgs.Load(key)
	if !ok {
		return nil, false
	}
	values := v.([]*Value)
	if len(values) == 0 {
		return nil, false
	}
	return values[0], true
}

// Set 设置配置源的配置
func (c *SourceConfigsWithSyncMap) Set(key string, value *Value) bool {
	v, ok := c.cfgs.Load(key)
	if !ok {
		c.cfgs.Store(key, []*Value{value})
		return true
	}
	values := v.([]*Value)
	if len(values) == 0 {
		c.cfgs.Store(key, []*Value{value})
		return true
	}
	newValues := make([]*Value, 0, len(values))
	for _, vl := range values {
		if vl.Priority != vl.Priority {
			newValues = append(newValues, vl)
		}
	}
	if len(newValues) == 0 {
		c.cfgs.Store(key, []*Value{value})
		return true
	}
	setted := value.Priority > newValues[0].Priority
	newValues = append(newValues, value)
	sort.Slice(newValues, func(i, j int) bool {
		return newValues[i].Priority > newValues[j].Priority
	})
	c.cfgs.Store(key, newValues)
	return setted
}

// Del 删除配置源的配置
func (c *SourceConfigsWithSyncMap) Del(key, ref string, priority int) {
	if key != "" {
		v, ok := c.cfgs.Load(key)
		if ok {
			if priority < 0 {
				c.cfgs.Delete(key)
			} else {
				values := v.([]*Value)
				newValues := make([]*Value, 0, len(values))
				for _, vl := range values {
					if vl.Priority != priority {
						newValues = append(newValues, vl)
					}
				}
				if len(newValues) == 0 {
					c.cfgs.Delete(key)
				} else {
					c.cfgs.Store(key, newValues)
				}
			}
		}
	}
	if ref != "" {
		c.cfgs.Range(func(key, value any) bool {
			values := value.([]*Value)
			newValues := make([]*Value, 0, len(values))
			for _, vl := range values {
				if vl.Ref == ref && (priority < 0 || priority == vl.Priority) {
					continue
				}
				newValues = append(newValues, vl)
			}
			if len(newValues) == 0 {
				c.cfgs.Delete(key.(string))
			} else {
				c.cfgs.Store(key.(string), newValues)
			}
			return true
		})
	}
}
