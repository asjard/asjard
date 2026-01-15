package config

import (
	"sort"
	"strings"
	"sync"

	"github.com/asjard/asjard/core/constant"
)

// ConfigsWithSyncMap implements the Configer interface using sync.Map.
// It stores the final "effective" configuration values that have already
// undergone priority filtering.
type ConfigsWithSyncMap struct {
	// cfgs stores map[string]*Value
	cfgs sync.Map
}

// Ensure interface compliance at compile time.
var _ Configer = &ConfigsWithSyncMap{}

// Get retrieves a configuration value by its direct key.
func (c *ConfigsWithSyncMap) Get(key string) (*Value, bool) {
	v, ok := c.cfgs.Load(key)
	if ok {
		return v.(*Value), ok
	}
	return nil, ok
}

// GetAll returns a snapshot of all current configurations in a standard map.
func (c *ConfigsWithSyncMap) GetAll() map[string]*Value {
	cfgs := make(map[string]*Value)
	c.cfgs.Range(func(key, value any) bool {
		cfgs[key.(string)] = value.(*Value)
		return true
	})
	return cfgs
}

// GetAllWithPrefixs retrieves all configurations starting with specified prefixes.
// It trims the prefix and the delimiter from the returned keys.
func (c *ConfigsWithSyncMap) GetAllWithPrefixs(prefixs ...string) map[string]*Value {
	cfgs := make(map[string]*Value)
	c.cfgs.Range(func(key, value any) bool {
		k := key.(string)
		for _, prefix := range prefixs {
			if strings.HasPrefix(k, prefix) {
				// Example: prefix "app", key "app.name" -> returns "name"
				cfgs[strings.TrimPrefix(k, prefix+constant.ConfigDelimiter)] = value.(*Value)
			}
		}
		return true
	})
	return cfgs
}

// Set saves a configuration value to the map.
func (c *ConfigsWithSyncMap) Set(key string, value *Value) {
	c.cfgs.Store(key, value)
}

// Del removes a configuration entry from the map.
func (c *ConfigsWithSyncMap) Del(key string) {
	c.cfgs.Delete(key)
}

// SourcesConfigWithSyncMap manages configurations grouped by their source name.
// This layer allows the manager to track which source provided which value.
type SourcesConfigWithSyncMap struct {
	// sources stores map[sourceName]SourceConfiger (usually SourceConfigsWithSyncMap)
	sources sync.Map
}

var _ SourcesConfiger = &SourcesConfigWithSyncMap{}

// Get retrieves a specific key's value from a specific configuration source.
func (c *SourcesConfigWithSyncMap) Get(sourceName, key string) (*Value, bool) {
	sourceConfigs, ok := c.sources.Load(sourceName)
	if !ok {
		return nil, false
	}
	return sourceConfigs.(SourceConfiger).Get(key)
}

// Set adds or updates a configuration for a specific source.
// Returns true if the set operation resulted in a state change.
func (c *SourcesConfigWithSyncMap) Set(sourceName, key string, value *Value) bool {
	sourceConfigs, ok := c.sources.Load(sourceName)
	if !ok {
		// Initialize the source-specific container if it doesn't exist yet.
		n := &SourceConfigsWithSyncMap{}
		c.sources.Store(sourceName, n)
		return n.Set(key, value)
	}
	return sourceConfigs.(SourceConfiger).Set(key, value)
}

// Del removes configuration data from a specific source, either by key or by reference.
func (c *SourcesConfigWithSyncMap) Del(sourceName, key, ref string, priority int) {
	sourceConfigs, ok := c.sources.Load(sourceName)
	if ok {
		sourceConfigs.(SourceConfiger).Del(key, ref, priority)
	}
}

// SourceConfigsWithSyncMap stores configuration values for a single source.
// Since a single source (like a file source) might have internal priorities,
// it stores values in a sorted slice for each key.
type SourceConfigsWithSyncMap struct {
	// cfgs stores map[string][]*Value (sorted by priority descending)
	cfgs sync.Map
}

// Get retrieves the highest priority value for a key within this specific source.
func (c *SourceConfigsWithSyncMap) Get(key string) (*Value, bool) {
	v, ok := c.cfgs.Load(key)
	if !ok {
		return nil, false
	}
	values := v.([]*Value)
	if len(values) == 0 {
		return nil, false
	}
	// Values are kept sorted, so index 0 is always the winner.
	return values[0], true
}

// Set adds a value to the key's list and sorts it by priority.
// Returns true if the new value becomes the highest priority (the first element).
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

	// Filter out existing values with the same priority to avoid duplicates.
	newValues := make([]*Value, 0, len(values))
	for _, vl := range values {
		if vl.Priority != value.Priority {
			newValues = append(newValues, vl)
		}
	}

	if len(newValues) == 0 {
		c.cfgs.Store(key, []*Value{value})
		return true
	}

	// Determine if this new value will override the current leader.
	setted := value.Priority > newValues[0].Priority
	newValues = append(newValues, value)

	// Keep values sorted: highest priority at the beginning.
	sort.Slice(newValues, func(i, j int) bool {
		return newValues[i].Priority > newValues[j].Priority
	})

	c.cfgs.Store(key, newValues)
	return setted
}

// Del removes values from this source.
// If priority < 0, it removes all values for the given key/ref.
// If key is provided, it targets a specific configuration key.
// If ref is provided, it targets all keys associated with a reference (e.g., a specific file).
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
					// Keep values that DON'T match the priority being deleted.
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

	// Reference-based deletion: iterate through all keys to find matches.
	if ref != "" {
		c.cfgs.Range(func(key, value any) bool {
			values := value.([]*Value)
			newValues := make([]*Value, 0, len(values))
			for _, vl := range values {
				// Skip values that match the reference and priority.
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
