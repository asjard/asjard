package consul

import (
	"sync"
	"testing"

	"github.com/asjard/asjard/core/config"
)

const (
	testSourceName     = "testSource"
	testSourcePriority = 0
)

type testSource struct {
	cb      func(*config.Event)
	configs map[string]any
	cm      sync.RWMutex
}

func newTestSource() (config.Sourcer, error) {
	return &testSource{
		configs: map[string]any{
			"asjard.stores.consul.clients.default.address": "127.0.0.1:8500",
			"asjard.config.setDefaultSource":               testSourceName,
		},
	}, nil
}

func (s *testSource) GetAll() map[string]*config.Value {
	configs := make(map[string]*config.Value)
	for key, value := range s.configs {
		configs[key] = &config.Value{
			Sourcer: s,
			Value:   value,
		}
	}
	return configs
}

// 添加配置到配置源中
func (s *testSource) Set(key string, value any) error {
	s.cm.Lock()
	s.configs[key] = value
	s.cm.Unlock()
	s.cb(&config.Event{
		Type: config.EventTypeCreate,
		Key:  key,
		Value: &config.Value{
			Sourcer: s,
			Value:   value,
		},
	})
	return nil
}

// 监听配置变化,当配置源中的配置发生变化时,
// 通过此回调方法通知config_manager进行配置变更
func (s *testSource) Watch(callback func(event *config.Event)) error {
	s.cb = callback
	return nil
}

// 和配置中心断开连接
func (s *testSource) Disconnect() {}

// 配置中心的优先级
func (s *testSource) Priority() int { return testSourcePriority }

// 配置源名称
func (s *testSource) Name() string { return testSourceName }

func initTestConfig() {
	if err := config.AddSource(testSourceName, testSourcePriority, newTestSource); err != nil {
		panic(err)
	}
	if err := config.Load(-1); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	initTestConfig()
	if err := clientManager.Start(); err != nil {
		panic(err)
	}
	m.Run()
	clientManager.Stop()
}
