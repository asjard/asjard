package config

import (
	"sort"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/constant"
)

// Configer defines the interface for global configuration storage and retrieval.
type Configer interface {
	// Get retrieves a value by its unique key.
	Get(key string) (*Value, bool)
	// GetAll returns a copy of all stored configurations.
	GetAll() map[string]*Value
	// GetAllWithPrefixs returns all configurations matching specific prefixes,
	// trimming the prefix from the keys in the resulting map.
	GetAllWithPrefixs(prefixs ...string) map[string]*Value
	// Set stores or updates a configuration value.
	Set(key string, value *Value)
	// Del removes a configuration value by its key.
	Del(key string)
}

// SourcesConfiger defines methods for managing configurations across multiple sources.
type SourcesConfiger interface {
	// Get retrieves a value from a specific configuration source.
	Get(sourceName, key string) (*Value, bool)
	// Set adds or updates a configuration within a specific source.
	Set(sourceName, key string, value *Value) bool
	// Del removes configurations from a specific source based on key, reference, or priority.
	Del(sourceName, key, ref string, priority int)
}

// SourceConfiger defines methods for managing configurations within a single specific source.
type SourceConfiger interface {
	// Get retrieves the highest priority value for a key in this source.
	Get(key string) (*Value, bool)
	// Set adds a value to the source, maintaining order by priority.
	Set(key string, value *Value) bool
	// Del removes values matching the key, reference, and priority criteria.
	Del(key, ref string, priority int)
}

// Configs represents the final, flattened global configuration state.
type Configs struct {
	cfgs map[string]*Value
	m    sync.RWMutex
}

var _ Configer = &Configs{}

// Get retrieves a configuration from the global map.
func (c *Configs) Get(key string) (*Value, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	value, ok := c.cfgs[key]
	return value, ok
}

// GetAll returns a thread-safe copy of all global configurations.
func (c *Configs) GetAll() map[string]*Value {
	c.m.RLock()
	defer c.m.RUnlock()
	cfgs := make(map[string]*Value, len(c.cfgs))
	for key, value := range c.cfgs {
		cfgs[key] = value
	}
	return cfgs
}

// GetAllWithPrefixs filters global configurations by prefix and strips the prefix from keys.
func (c *Configs) GetAllWithPrefixs(prefixs ...string) map[string]*Value {
	c.m.RLock()
	defer c.m.RUnlock()
	cfgs := make(map[string]*Value)
	for key, value := range c.cfgs {
		for _, p := range prefixs {
			if strings.HasPrefix(key, p) {
				// Trim the prefix and the delimiter to normalize the key.
				cfgs[strings.TrimPrefix(key, p+constant.ConfigDelimiter)] = value
			}
		}
	}
	return cfgs
}

// Set updates the global configuration map with thread-safety.
func (c *Configs) Set(key string, value *Value) {
	c.m.Lock()
	defer c.m.Unlock()
	c.cfgs[key] = value
}

// Del removes a key from the global configuration map.
func (c *Configs) Del(key string) {
	c.m.Lock()
	defer c.m.Unlock()
	delete(c.cfgs, key)
}

// SourcesConfig manages a collection of named configuration sources (e.g., file, env, consul).
type SourcesConfig struct {
	sources map[string]SourceConfiger
	m       sync.RWMutex
}

var _ SourcesConfiger = &SourcesConfig{}

// Get retrieves a key from a specific named configuration source.
func (c *SourcesConfig) Get(sourceName, key string) (*Value, bool) {
	c.m.RLock()
	configs, ok := c.sources[sourceName]
	c.m.RUnlock()
	if !ok {
		return nil, false
	}
	return configs.Get(key)
}

// Set updates a configuration in a specific source. If the source doesn't exist, it is initialized.
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

// Del removes a configuration from the specified source.
func (c *SourcesConfig) Del(sourceName, key, ref string, priority int) {
	c.m.Lock()
	defer c.m.Unlock()
	configs, ok := c.sources[sourceName]
	if ok {
		configs.Del(key, ref, priority)
	}
}

// SourceConfigs manages multiple values for the same key within a source, sorted by priority.
type SourceConfigs struct {
	cfgs map[string][]*Value
	m    sync.RWMutex
}

var _ SourceConfiger = &SourceConfigs{}

// Get returns the value with the highest priority (first element) for a specific key.
func (c *SourceConfigs) Get(key string) (*Value, bool) {
	c.m.RLock()
	defer c.m.RUnlock()
	values, ok := c.cfgs[key]
	if !ok || len(values) == 0 {
		return nil, false
	}
	return values[0], true
}

// Set inserts a value into the source and re-sorts the priorities.
// Returns true if the newly added value becomes the highest priority for the key.
func (c *SourceConfigs) Set(key string, value *Value) bool {
	c.m.Lock()
	defer c.m.Unlock()
	values, ok := c.cfgs[key]
	if !ok || len(values) == 0 {
		c.cfgs[key] = []*Value{value}
		return true
	}
	// Remove existing values with the same priority to allow override.
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

	// Check if this new value will take precedence.
	setted := value.Priority > newValues[0].Priority
	newValues = append(newValues, value)

	// Sort from highest to lowest priority.
	sort.Slice(newValues, func(i, j int) bool {
		return newValues[i].Priority > newValues[j].Priority
	})
	c.cfgs[key] = newValues
	return setted
}

// Del removes values from the source based on key, reference, or specific priority.
func (c *SourceConfigs) Del(key, ref string, priority int) {
	c.m.Lock()
	defer c.m.Unlock()
	// Case 1: Delete by key.
	if key != "" {
		values, ok := c.cfgs[key]
		if ok {
			// If priority is negative, clear all values for this key.
			if priority < 0 {
				delete(c.cfgs, key)
			} else {
				// Filter out the specific priority.
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
	// Case 2: Delete by reference (e.g., all keys associated with a specific file).
	if ref != "" {
		for key, values := range c.cfgs {
			newValues := make([]*Value, 0, len(values))
			for _, v := range values {
				// Skip values matching the reference and optional priority.
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
}
