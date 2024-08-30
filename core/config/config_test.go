package config

import "testing"

func TestConfiger(t *testing.T) {
	testConfiger(t, &Configs{
		cfgs: make(map[string]*Value),
	})
}

func TestSourcesConfiger(t *testing.T) {
	testSourcesConfiger(t, &SourcesConfig{
		sources: make(map[string]SourceConfiger),
	})
}

func TestSourceConfiger(t *testing.T) {
	testSourceConfiger(t, &SourceConfigs{
		cfgs: make(map[string][]*Value),
	})
}

func testConfiger(t *testing.T, configer Configer) {
	t.Run("Get", func(t *testing.T) {
		cases := []struct {
			key   string
			value any
		}{
			{key: "test_get_key", value: "test_get_key_value"},
		}
		for _, caze := range cases {
			configer.Set(caze.key, &Value{
				Sourcer: &testSource{},
				Value:   caze.value,
			})
		}
		for _, caze := range cases {
			v, ok := configer.Get(caze.key)
			if !ok {
				t.Errorf("key %s not found", caze.key)
				t.FailNow()
			}
			if v.Value != caze.value {
				t.Errorf("key %s value not equal, act: %v, want: %v", caze.key, v.Value, caze.value)
				t.FailNow()
			}
		}
	})
	t.Run("GetAll", func(t *testing.T) {})
	t.Run("GetWithPrefixs", func(t *testing.T) {})
	t.Run("Set", func(t *testing.T) {})
	t.Run("Del", func(t *testing.T) {})
}

func testSourcesConfiger(t *testing.T, _ SourcesConfiger) {
	t.Run("Get", func(t *testing.T) {})
	t.Run("Set", func(t *testing.T) {})
	t.Run("Del", func(t *testing.T) {})
}

func testSourceConfiger(t *testing.T, _ SourceConfiger) {
	t.Run("Get", func(t *testing.T) {})
	t.Run("Set", func(t *testing.T) {})
	t.Run("Del", func(t *testing.T) {})
}
