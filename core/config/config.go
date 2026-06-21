package config

import (
	"maps"
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
	keys []string
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
	maps.Copy(cfgs, c.cfgs)
	return cfgs
}

// GetAllWithPrefixs filters global configurations by prefix and strips the prefix from keys.
func (c *Configs) GetAllWithPrefixs(prefixs ...string) map[string]*Value {
	c.ensureKeyIndex()
	c.m.RLock()
	defer c.m.RUnlock()
	capacity := 0
	for _, prefix := range prefixs {
		start, end := prefixKeyRange(c.keys, prefix)
		capacity += end - start
	}
	cfgs := make(map[string]*Value, capacity)
	// Apply prefixes from general to specific so later chain entries
	// deterministically override values normalized to the same key.
	for _, p := range prefixs {
		start, end := prefixKeyRange(c.keys, p)
		for _, key := range c.keys[start:end] {
			cfgs[strings.TrimPrefix(key, p+constant.ConfigDelimiter)] = c.cfgs[key]
		}
	}
	return cfgs
}

// Set updates the global configuration map with thread-safety.
func (c *Configs) Set(key string, value *Value) {
	c.m.Lock()
	defer c.m.Unlock()
	if _, ok := c.cfgs[key]; !ok {
		c.keys = insertSortedKey(c.keys, key)
	}
	c.cfgs[key] = value
}

// Del removes a key from the global configuration map.
func (c *Configs) Del(key string) {
	c.m.Lock()
	defer c.m.Unlock()
	if _, ok := c.cfgs[key]; ok {
		c.keys = removeSortedKey(c.keys, key)
	}
	delete(c.cfgs, key)
}

func (c *Configs) ensureKeyIndex() {
	c.m.RLock()
	indexed := len(c.keys) == len(c.cfgs)
	c.m.RUnlock()
	if indexed {
		return
	}
	c.m.Lock()
	if len(c.keys) != len(c.cfgs) {
		c.keys = c.keys[:0]
		for key := range c.cfgs {
			c.keys = append(c.keys, key)
		}
		sort.Strings(c.keys)
	}
	c.m.Unlock()
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
	c.m.RLock()
	configs, ok := c.sources[sourceName]
	c.m.RUnlock()
	if !ok {
		c.m.Lock()
		configs, ok = c.sources[sourceName]
		if !ok {
			configs = &SourceConfigs{cfgs: make(map[string][]*Value)}
			c.sources[sourceName] = configs
		}
		c.m.Unlock()
	}
	return configs.Set(key, value)
}

// Del removes a configuration from the specified source.
func (c *SourcesConfig) Del(sourceName, key, ref string, priority int) {
	c.m.RLock()
	configs, ok := c.sources[sourceName]
	c.m.RUnlock()
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
	newValues, setted := upsertSortedValue(values, value)
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

func prefixKeyRange(keys []string, prefix string) (int, int) {
	start := sort.SearchStrings(keys, prefix)
	end := start + sort.Search(len(keys)-start, func(i int) bool {
		return !strings.HasPrefix(keys[start+i], prefix)
	})
	return start, end
}

func insertSortedKey(keys []string, key string) []string {
	index := sort.SearchStrings(keys, key)
	keys = append(keys, "")
	copy(keys[index+1:], keys[index:])
	keys[index] = key
	return keys
}

func removeSortedKey(keys []string, key string) []string {
	index := sort.SearchStrings(keys, key)
	if index == len(keys) || keys[index] != key {
		return keys
	}
	copy(keys[index:], keys[index+1:])
	keys[len(keys)-1] = ""
	return keys[:len(keys)-1]
}

func upsertSortedValue(values []*Value, value *Value) ([]*Value, bool) {
	winnerChanged := len(values) == 0 || value.Priority >= values[0].Priority
	result := make([]*Value, 0, len(values)+1)
	for _, current := range values {
		if current.Priority != value.Priority {
			result = append(result, current)
		}
	}
	index := sort.Search(len(result), func(i int) bool {
		return result[i].Priority < value.Priority
	})
	result = append(result, nil)
	copy(result[index+1:], result[index:])
	result[index] = value
	return result, winnerChanged
}
