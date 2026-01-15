/*
Package mem implements an in-memory configuration source,
fulfilling the core/config/Source interface.
*/
package mem

import (
	"sync"

	"github.com/asjard/asjard/core/config"
)

const (
	// Name is the unique identifier for the memory-based configuration source.
	Name = "mem"
	// Priority is 99. In the framework's hierarchy, a higher numerical value
	// usually represents a lower precedence.
	// This makes memory the "baseline" or "default" override source.
	Priority = 99
)

// Mem manages a thread-safe map of configuration keys and values in the local process memory.
type Mem struct {
	options *config.SourceOptions
	configs map[string]any
	cm      sync.RWMutex // Protects the configs map from concurrent read/write access.
}

func init() {
	// Register the memory source.
	// Note: It registers with Priority*-1 to adjust its standing in the source registry.
	config.AddSource(Name, Priority*-1, New)
}

// New initializes a new in-memory configuration provider.
func New(options *config.SourceOptions) (config.Sourcer, error) {
	return &Mem{
		configs: make(map[string]any),
		options: options,
	}, nil
}

// GetAll returns a snapshot of all current in-memory configurations.
// It wraps the raw values into the framework's config.Value structure.
func (m *Mem) GetAll() map[string]*config.Value {
	configs := make(map[string]*config.Value)
	configMap := m.getAll()
	for key, value := range configMap {
		configs[key] = &config.Value{
			Sourcer: m,
			Value:   value,
		}
	}
	return configs
}

// Set adds or updates a configuration key in memory and triggers a callback event.
// This allows other components to react to runtime configuration changes immediately.
func (m *Mem) Set(key string, value any) error {
	m.set(key, value)
	// Notify the configuration manager of the update.
	m.options.Callback(&config.Event{
		Type: config.EventTypeUpdate,
		Key:  key,
		Value: &config.Value{
			Sourcer: m,
			Value:   value,
		},
	})
	return nil
}

// Disconnect is a no-op as there are no external connections to manage.
func (m *Mem) Disconnect() {}

// Priority returns the precedence level of this source.
func (m *Mem) Priority() int {
	return Priority
}

// Name returns "mem".
func (m *Mem) Name() string {
	return Name
}

// get retrieves a value for a specific key using a read-lock.
func (m *Mem) get(key string) (any, bool) {
	m.cm.RLock()
	v, ok := m.configs[key]
	m.cm.RUnlock()
	return v, ok
}

// getAll creates a copy of the current configuration map to avoid data races.
func (m *Mem) getAll() map[string]any {
	configs := make(map[string]any)
	m.cm.RLock()
	for key, value := range m.configs {
		configs[key] = value
	}
	m.cm.RUnlock()
	return configs
}

// set updates the internal map using a write-lock.
func (m *Mem) set(key string, value any) {
	m.cm.Lock()
	m.configs[key] = value
	m.cm.Unlock()
}
