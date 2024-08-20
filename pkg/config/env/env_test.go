package env

import (
	"os"
	"strings"
	"testing"

	"github.com/asjard/asjard/core/config"
	"github.com/stretchr/testify/assert"
)

func TestEnv(t *testing.T) {
	testEnv := &Env{}
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
		var cbKey string
		var cbValue any
		testEnv.cb = func(event *config.Event) {
			cbValue = event.Value.Value
			cbKey = event.Key
		}

		assert.Nil(t, testEnv.Set(key, value))

		assert.Equal(t, key, cbKey)
		assert.Equal(t, value, cbValue)

		assert.Equal(t, value, os.Getenv(strings.ReplaceAll(key, ".", "_")))
	})
}
