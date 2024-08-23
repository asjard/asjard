package env

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	var cbKey string
	var cbValue any
	var m sync.RWMutex
	testEnv := &Env{
		options: &config.SourceOptions{
			Callback: func(event *config.Event) {
				m.Lock()
				defer m.Unlock()
				cbValue = event.Value.Value
				cbKey = event.Key
			},
		},
	}

	assert.Equal(t, Name, testEnv.Name())
	assert.Equal(t, Priority, testEnv.Priority())
	t.Run("TestGetAll", func(t *testing.T) {
		datas := []struct {
			key      string
			propsKey string
			value    string
		}{
			{key: "test_key", propsKey: "test.key", value: "test_value"},
			{key: "Test_key", propsKey: "Test.key", value: "test_value"},
			{key: "Test__key", propsKey: "Test..key", value: "test_value"},
		}
		for _, data := range datas {
			assert.Nil(t, os.Setenv(data.key, data.value), "set %s", data.key)
			envConfigs := testEnv.GetAll()
			v, ok := envConfigs[data.propsKey]
			assert.Equal(t, true, ok, data.key)
			assert.Equal(t, "test_value", v.Value, data.key)
		}
	})
	t.Run("TestSet", func(t *testing.T) {
		key := "test.set.key"
		value := "test_set_value"

		assert.Nil(t, testEnv.Set(key, value))
		time.Sleep(50 * time.Millisecond)

		m.RLock()
		defer m.RUnlock()
		assert.Equal(t, key, cbKey)
		assert.Equal(t, value, cbValue)

		assert.Equal(t, value, os.Getenv(strings.ReplaceAll(key, ".", "_")))
	})
}
