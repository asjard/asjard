package etcd

import (
	"sync"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/stretchr/testify/assert"
	"go.etcd.io/etcd/client/v3/mock/mockserver"
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
			"asjard.database.etcd.clients.default.endpoints": "localhost:0",
			"asjard.database.etcd.clients.another.endpoints": "localhost:1",
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
	mockserver.StartMockServers(1)
	if err := clientManager.Bootstrap(); err != nil {
		panic(err)
	}
	m.Run()
	clientManager.Shutdown()
}

func TestNewClients(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		client, err := Client()
		assert.Nil(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config.GetStrings("asjard.database.etcd.clients.default.endpoints", []string{}), client.Endpoints())
	})
	t.Run("another", func(t *testing.T) {
		client, err := Client(WithClientName("another"))
		assert.Nil(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config.GetStrings("asjard.database.etcd.clients.another.endpoints", []string{}), client.Endpoints())
	})

	t.Run("new", func(t *testing.T) {
		config.Set("asjard.database.etcd.clients.new.endpoints", "localhost:2")
		time.Sleep(2 * time.Second)
		_, err := Client(WithClientName("new"))
		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}
		assert.Nil(t, err)
	})
	t.Run("shutdown", func(t *testing.T) {
		clientManager.Shutdown()
		_, err := Client()
		assert.NotNil(t, err)
	})
}
