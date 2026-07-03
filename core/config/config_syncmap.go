package config

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/asjard/asjard/core/constant"
)

// ConfigsWithSyncMap implements the Configer interface using sync.Map.
// It stores the final "effective" configuration values that have already
// undergone priority filtering.
type ConfigsWithSyncMap struct {
	// cfgs stores map[string]*Value
	cfgs sync.Map
	// keys is a sorted index used to avoid a full sync.Map scan per prefix.
	keys         []string
	keysMu       sync.RWMutex
	keysVersion  atomic.Uint64
	indexVersion uint64
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
	c.ensureKeyIndex()
	c.keysMu.RLock()
	defer c.keysMu.RUnlock()
	capacity := 0
	for _, prefix := range prefixs {
		start, end := prefixKeyRange(c.keys, prefix)
		capacity += end - start
	}
	cfgs := make(map[string]*Value, capacity)
	// Apply prefixes from general to specific so later chain entries win.
	for _, prefix := range prefixs {
		start, end := prefixKeyRange(c.keys, prefix)
		for _, key := range c.keys[start:end] {
			if value, ok := c.cfgs.Load(key); ok {
				cfgs[strings.TrimPrefix(key, prefix+constant.ConfigDelimiter)] = value.(*Value)
			}
		}
	}
	return cfgs
}

// Set saves a configuration value to the map.
func (c *ConfigsWithSyncMap) Set(key string, value *Value) {
	if _, loaded := c.cfgs.LoadOrStore(key, value); loaded {
		c.cfgs.Store(key, value)
	} else {
		c.keysVersion.Add(1)
	}
}

// Del removes a configuration entry from the map.
func (c *ConfigsWithSyncMap) Del(key string) {
	if _, loaded := c.cfgs.LoadAndDelete(key); loaded {
		c.keysVersion.Add(1)
	}
}

func (c *ConfigsWithSyncMap) ensureKeyIndex() {
	version := c.keysVersion.Load()
	c.keysMu.RLock()
	current := c.indexVersion == version
	c.keysMu.RUnlock()
	if current {
		return
	}
	c.keysMu.Lock()
	defer c.keysMu.Unlock()
	version = c.keysVersion.Load()
	if c.indexVersion == version {
		return
	}
	keys := make([]string, 0)
	c.cfgs.Range(func(key, _ any) bool {
		keys = append(keys, key.(string))
		return true
	})
	sort.Strings(keys)
	c.keys = keys
	c.indexVersion = version
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
		sourceConfigs, _ = c.sources.LoadOrStore(sourceName, &SourceConfigsWithSyncMap{})
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
	// cfgs stores map[string]*sourceValues.
	cfgs sync.Map
	mu   sync.RWMutex
}

type sourceValues struct {
	mu     sync.RWMutex
	values []*Value
}

// Get retrieves the highest priority value for a key within this specific source.
func (c *SourceConfigsWithSyncMap) Get(key string) (*Value, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.cfgs.Load(key)
	if !ok {
		return nil, false
	}
	entry := v.(*sourceValues)
	entry.mu.RLock()
	defer entry.mu.RUnlock()
	values := entry.values
	if len(values) == 0 {
		return nil, false
	}
	// Values are kept sorted, so index 0 is always the winner.
	return values[0], true
}

// Set adds a value to the key's list and sorts it by priority.
// Returns true if the new value becomes the highest priority (the first element).
func (c *SourceConfigsWithSyncMap) Set(key string, value *Value) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	loaded, _ := c.cfgs.LoadOrStore(key, &sourceValues{})
	entry := loaded.(*sourceValues)
	entry.mu.Lock()
	defer entry.mu.Unlock()
	values, loadedWinner := upsertSortedValue(entry.values, value)
	entry.values = values
	return loadedWinner
}

// Del removes values from this source.
// If priority < 0, it removes all values for the given key/ref.
// If key is provided, it targets a specific configuration key.
// If ref is provided, it targets all keys associated with a reference (e.g., a specific file).
func (c *SourceConfigsWithSyncMap) Del(key, ref string, priority int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if key != "" {
		v, ok := c.cfgs.Load(key)
		if ok {
			entry := v.(*sourceValues)
			entry.mu.Lock()
			if priority < 0 {
				c.cfgs.Delete(key)
			} else {
				newValues := make([]*Value, 0, len(entry.values))
				for _, vl := range entry.values {
					// Keep values that DON'T match the priority being deleted.
					if vl.Priority != priority {
						newValues = append(newValues, vl)
					}
				}
				if len(newValues) == 0 {
					c.cfgs.Delete(key)
				} else {
					entry.values = newValues
				}
			}
			entry.mu.Unlock()
		}
	}

	// Reference-based deletion: iterate through all keys to find matches.
	if ref != "" {
		c.cfgs.Range(func(key, value any) bool {
			entry := value.(*sourceValues)
			entry.mu.Lock()
			defer entry.mu.Unlock()
			newValues := make([]*Value, 0, len(entry.values))
			for _, vl := range entry.values {
				// Skip values that match the reference and priority.
				if vl.Ref == ref && (priority < 0 || priority == vl.Priority) {
					continue
				}
				newValues = append(newValues, vl)
			}

			if len(newValues) == 0 {
				c.cfgs.Delete(key.(string))
			} else {
				entry.values = newValues
			}
			return true
		})
	}
}
