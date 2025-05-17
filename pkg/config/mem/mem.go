/*
Package mem 内存配置中心实现,对core/config/Source的实现
*/
package mem

import (
	"sync"

	"github.com/asjard/asjard/core/config"
)

const (
	// Name 内存配置源名称
	Name = "mem"
	// Priority 优先级
	Priority = 99
)

// Mem .
type Mem struct {
	options *config.SourceOptions
	configs map[string]any
	cm      sync.RWMutex
}

func init() {
	config.AddSource(Name, Priority*-1, New)
}

// New .s
func New(options *config.SourceOptions) (config.Sourcer, error) {
	return &Mem{
		configs: make(map[string]any),
		options: options,
	}, nil
}

// GetAll .
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

// Set .
func (m *Mem) Set(key string, value any) error {
	m.set(key, value)
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

// Disconnect .
func (m *Mem) Disconnect() {}

// Priority .
func (m *Mem) Priority() int {
	return Priority
}

// Name .
func (m *Mem) Name() string {
	return Name
}

func (m *Mem) get(key string) (any, bool) {
	m.cm.RLock()
	v, ok := m.configs[key]
	m.cm.RUnlock()
	return v, ok
}

func (m *Mem) getAll() map[string]any {
	configs := make(map[string]any)
	m.cm.RLock()
	for key, value := range m.configs {
		configs[key] = value
	}
	m.cm.RUnlock()
	return configs
}

func (m *Mem) set(key string, value any) {
	m.cm.Lock()
	m.configs[key] = value
	m.cm.Unlock()
}
