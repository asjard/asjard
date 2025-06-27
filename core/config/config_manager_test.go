package config

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	testSourceName     = "testSource"
	testSourcePriority = 0
)

type testSource struct {
	options *SourceOptions
	configs map[string]any
	cm      sync.RWMutex
}

func newTestSource(options *SourceOptions) (Sourcer, error) {
	return &testSource{
		options: options,
		configs: map[string]any{
			"asjard.config.setDefaultSource": testSourceName,
			"testInt":                        1,
			"testStr":                        "test data",
			"testFloat":                      0.01,
			"testDuration":                   "10m",
			"testBool":                       true,
			"test_del_ref":                   "test_del_ref_value",
		},
	}, nil
}

// GetAll .
func (s *testSource) GetAll() map[string]*Value {
	configs := make(map[string]*Value)
	s.cm.RLock()
	for key, value := range s.configs {
		ref := ""
		if key == "test_del_ref" {
			ref = "test_del_ref"
		}
		configs[key] = &Value{
			Sourcer: s,
			Value:   value,
			Ref:     ref,
		}
	}
	s.cm.RUnlock()
	return configs
}

func (s *testSource) Set(key string, value any) error {
	s.cm.Lock()
	s.configs[key] = value
	s.cm.Unlock()
	if key == "test_del" {
		s.options.Callback(&Event{
			Type: EventTypeDelete,
			Key:  key,
			Value: &Value{
				Sourcer: s,
				Value:   value,
			},
		})
	} else if key == "test_del_ref" {
		s.options.Callback(&Event{
			Type: EventTypeDelete,
			Value: &Value{
				Sourcer: s,
				Value:   value,
				Ref:     "test_del_ref",
			},
		})
	} else {
		s.options.Callback(&Event{
			Type: EventTypeUpdate,
			Key:  key,
			Value: &Value{
				Sourcer: s,
				Value:   value,
			},
		})
	}
	return nil
}

func (s *testSource) Disconnect() {}

func (s *testSource) Priority() int {
	return testSourcePriority
}
func (s *testSource) Name() string {
	return testSourceName
}

func initTestConfig() {
	if err := AddSource(testSourceName, testSourcePriority, newTestSource); err != nil {
		panic(err)
	}

	if err := Load(-1); err != nil {
		panic(err)
	}
}

func TestMain(m *testing.M) {
	initTestConfig()
	m.Run()
}

func TestDeleteEvent(t *testing.T) {
	t.Run("DelByKey", func(t *testing.T) {
		if err := Set("test_del", "test_del_value"); err != nil {
			t.Error(err)
			t.FailNow()
		}
		if out := GetString("test_del", ""); out != "" {
			t.Errorf("test del fail, current: %s, want: ", out)
			t.FailNow()
		}
	})
	t.Run("DelByRef", func(t *testing.T) {
		if err := Set("test_del_ref", "test_del_value"); err != nil {
			t.Error(err)
			t.FailNow()
		}
		if out := GetString("test_del_ref", ""); out != "" {
			t.Errorf("test del fail, current: %s, want: ", out)
			t.FailNow()
		}
	})
}

func TestAddDuplicateSource(t *testing.T) {
	t.Run("SameName", func(t *testing.T) {
		assert.NotNil(t, AddSource(testSourceName, 1111, newTestSource))
	})
	t.Run("SamePriority", func(t *testing.T) {
		assert.NotNil(t, AddSource("not_exist_source_name", 0, newTestSource))
	})

	t.Run("DiffSource", func(t *testing.T) {
		assert.NoError(t, AddSource("not_exist_source_name", 1111, newTestSource))
	})
}

//gocyclo:ignore
func TestGetWithUnmarshal(t *testing.T) {
	t.Run("JsonUnmarshalWithParam", func(t *testing.T) {
		content := `test_param_prefix:
  clients:
    default:
      address: 127.0.0.1:6379
      db: 1
      username: user
      password: !!str 123
      options:
        clientName: testClientName
    cache: ${test_param_prefix.clients.default}
    statistic:
      address: ${test_param_prefix.clients.default.address}
      options: ${test_param_prefix.clients.default.options}`
		propsMap, err := ConvertToProperties(".yml", []byte(content))
		if err != nil {
			t.Error(err)
			t.FailNow()
		}
		for key, value := range propsMap {
			if err := Set(key, value); err != nil {
				t.Error(err)
				t.FailNow()
			}
		}
		type testParamsPrefixOptions struct {
			ClientName string `json:"clientName"`
		}
		type testParamsPrefixConfig struct {
			Address  string                  `json:"address"`
			DB       int                     `json:"db"`
			Username string                  `json:"username"`
			Password string                  `json:"password"`
			Options  testParamsPrefixOptions `json:"options"`
		}
		out := make(map[string]*testParamsPrefixConfig)
		if err := GetWithUnmarshal("test_param_prefix.clients", &out); err != nil {
			t.Error(err)
			t.FailNow()
		}
		// default 不为空
		defaultConf, ok := out["default"]
		if !ok {
			t.Error("default config not found")
			t.FailNow()
		}
		if defaultConf.Address == "" ||
			defaultConf.DB == 0 ||
			defaultConf.Username == "" ||
			defaultConf.Password == "" ||
			defaultConf.Options.ClientName == "" {
			t.Errorf("default config is empty, default: %+v", defaultConf)
			t.FailNow()
		}
		// cache不为空，且值和default相等
		cacheConf, ok := out["cache"]
		if !ok {
			t.Error("cache config not found")
			t.FailNow()
		}
		if cacheConf.Address != defaultConf.Address ||
			cacheConf.DB != defaultConf.DB ||
			cacheConf.Username != defaultConf.Username ||
			cacheConf.Password != defaultConf.Password ||
			cacheConf.Options.ClientName != defaultConf.Options.ClientName {
			t.Errorf("cache config not equal with default config, cache: %+v, default: %+v", cacheConf, defaultConf)
			t.FailNow()
		}
		// statistic不为空，且只有address和options.ClientName和default相等
		statisticConf, ok := out["statistic"]
		if !ok {
			t.Error("statistic config not found")
			t.FailNow()
		}
		if statisticConf.Address != defaultConf.Address ||
			statisticConf.Options.ClientName != defaultConf.Options.ClientName ||
			statisticConf.DB != 0 ||
			statisticConf.Username != "" ||
			statisticConf.Password != "" {
			t.Errorf("unexpect statistic config, statistic: %+v, default: %+v", statisticConf, defaultConf)
			t.FailNow()
		}
	})
	t.Run("JsonUnmarshal", func(t *testing.T) {
		datas := []struct {
			prefix string
			key    string
			value  int
		}{
			{prefix: "test_json_prefix", key: "a", value: 1},
			{prefix: "test_json_prefix", key: "b", value: 2},
		}
		for _, data := range datas {
			if err := Set(data.prefix+"."+data.key, data.value); err != nil {
				t.Error(err)
				t.FailNow()
			}
		}
		out := make(map[string]int)
		if err := GetWithUnmarshal("test_json_prefix", &out); err != nil {
			t.Error(err)
			t.FailNow()
		}
		for _, data := range datas {
			assert.Equal(t, data.value, out[data.key])
		}
	})

	t.Run("YamlUnmarshal", func(t *testing.T) {
		datas := []struct {
			prefix string
			key    string
			value  int
		}{
			{prefix: "test_yaml_prefix", key: "a", value: 1},
			{prefix: "test_yaml_prefix", key: "b", value: 2},
		}
		for _, data := range datas {
			if err := Set(data.prefix+"."+data.key, data.value); err != nil {
				t.Error(err)
				t.FailNow()
			}
		}
		out := make(map[string]int)
		if err := GetWithYamlUnmarshal("test_yaml_prefix", &out); err != nil {
			t.Error(err)
			t.FailNow()
		}
		for _, data := range datas {
			assert.Equal(t, data.value, out[data.key])
		}
	})

}

func TestGetString(t *testing.T) {
	t.Run("GetString", func(t *testing.T) {
		datas := []struct {
			key     string
			value   any
			expect  string
			options []Option
		}{
			{key: "testStr", value: "test_str", expect: "test_str"},
			{key: "testStr", value: "Test_Str", expect: "test_str", options: []Option{WithToLower()}},
			{key: "testStrInt", value: 1, expect: "1"},
			{key: "testStrFloat", value: 0.01, expect: "0.01"},
			{key: "testStrBool", value: true, expect: "true"},
			{key: "testStrBool1", value: true, expect: "TRUE", options: []Option{WithToUpper()}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetString(data.key, "", data.options...)
			if out != data.expect {
				t.Errorf("test %v fail, out %s want %s", data.value, out, data.expect)
				t.FailNow()
			}
		}
	})
	t.Run("GetStrings", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect []string
		}{
			{key: "testStr", value: "test_str", expect: []string{"test_str"}},
			{key: "testStr1", value: "test_str,test_str1", expect: []string{"test_str", "test_str1"}},
			{key: "testStr2", value: []string{"test_str", "test_str1"}, expect: []string{"test_str", "test_str1"}},
			{key: "testStrInt", value: 1, expect: []string{"1"}},
			{key: "testStrFloat", value: 0.01, expect: []string{"0.01"}},
			{key: "testStrBool", value: true, expect: []string{"true"}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
		}
		for _, data := range datas {
			out := GetStrings(data.key, []string{})
			if len(out) != len(data.expect) {
				t.Errorf("test %v fail, len not equal, current: %d, want: %d", data.value, len(out), len(data.expect))
				t.FailNow()
			}
			for index, v := range out {
				if data.expect[index] != v {
					t.Errorf("test %v fail, out %v want %v", data.value, out, data.expect)
					t.FailNow()
				}
			}
		}
	})
}

func TestGetWithParam(t *testing.T) {
	assert.Nil(t, Set("test_param", "test_param_value"))
	assert.Nil(t, Set("test_get_by_param", "${test_param}"))
	assert.Equal(t, "test_param_value", GetString("test_get_by_param", ""))
}

func TestGetWithCipher(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		key := "test_base64_cipher"
		value := "test_base64_cipher_value"
		assert.Nil(t, Set(key, base64.StdEncoding.EncodeToString([]byte(value))))
		assert.Equal(t, value, GetString(key, "", WithCipher("base64")))
	})
	t.Run("autoEncrypt", func(t *testing.T) {
		key := "test_auto_decrypt"
		value := "test_auto_decrypt_value"
		encryptedValue := "encrypted_base64:" + base64.StdEncoding.EncodeToString([]byte(value))
		assert.Nil(t, Set(key, encryptedValue))
		// 自动解密
		assert.Equal(t, value, GetString(key, ""))
		// 禁用自动解密
		assert.Equal(t, encryptedValue, GetString(key, "", WithDisableAutoDecryptValue()))
	})

}

func TestSetWithCipher(t *testing.T) {
	key := "test_set_with_base64_cipher"
	value := "test_set_with_base64_cipher_value"
	assert.Nil(t, Set(key, value, WithCipher("base64")))
	assert.Equal(t, base64.StdEncoding.EncodeToString([]byte(value)), GetString(key, ""))
	assert.Equal(t, value, GetString(key, "", WithCipher("base64")))
}

func TestListener(t *testing.T) {
	t.Run("AddListener", func(t *testing.T) {
		key := "test_add_listener"
		value := "test_add_listener_value"
		received := make(chan struct{})
		AddListener(key, func(event *Event) {
			received <- struct{}{}
		})
		assert.Nil(t, Set(key, value))
		select {
		case <-received:
			break
		case <-time.After(time.Millisecond * 10):
			t.Error("after 100ms not received event")
			t.FailNow()
		}
		assert.Equal(t, value, GetString(key, ""))
		RemoveListener(key)
	})
	t.Run("AddPatternListener", func(t *testing.T) {
		key := "test_add_listener_pattern"
		value := "test_add_listener_pattern_value"
		received := make(chan struct{})
		AddPatternListener(key+".*", func(event *Event) {
			received <- struct{}{}
		})
		assert.Nil(t, Set(key, value))
		select {
		case <-received:
			break
		case <-time.After(time.Millisecond * 10):
			t.Error("after 10ms not received event")
			t.FailNow()
		}
		assert.Equal(t, value, GetString(key, ""))
	})
	t.Run("AddPrefixListener", func(t *testing.T) {
		key := "test_add_listener_prefix"
		value := "test_add_listener_prefix_value"
		received := make(chan struct{})
		AddPrefixListener(key, func(event *Event) {
			received <- struct{}{}
		})
		assert.Nil(t, Set(key, value))
		select {
		case <-received:
			break
		case <-time.After(time.Millisecond * 10):
			t.Error("after 10ms not received event")
			t.FailNow()
		}
		assert.Equal(t, value, GetString(key, ""))
	})
}

func TestGetBytes(t *testing.T) {
	datas := []struct {
		key    string
		value  any
		expect []byte
	}{
		{key: "testBytesInt", value: 1, expect: []byte("1")},
		{key: "testByteStr", value: "test_bytes", expect: []byte("test_bytes")},
		{key: "testBytes", value: []byte("test"), expect: []byte("test")},
	}
	for _, data := range datas {
		if err := Set(data.key, data.value); err != nil {
			t.Errorf("set data fail %s", err.Error())
			t.FailNow()
		}
		out := GetByte(data.key, []byte(""))
		if string(out) != string(data.expect) {
			t.Errorf("test %s fail, current: %s, want: %s", data.key, string(out), string(data.expect))
			t.FailNow()
		}
	}
}

type testYamlUnmarshal struct{}

func (testYamlUnmarshal) Unmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

func TestGetOption(t *testing.T) {
	t.Run("WithSource", func(t *testing.T) {
		key := "test_with_source"
		value := "test_with_source_value"
		assert.Nil(t, Set(key, value, WithSource(testSourceName)))
		assert.Equal(t, value, GetString(key, "", WithSource(testSourceName)))
	})
	t.Run("WithLocation", func(t *testing.T) {
		key := "test_with_location"
		date := time.Now().AddDate(0, 0, -2)
		assert.Nil(t, Set(key, date.Unix()))
		assert.Equal(t, date.Unix(), GetTime(key, time.Now(), WithLocation(time.Local)).Unix())

	})
	t.Run("WithUnmarshal", func(t *testing.T) {
		key := "test_with_unmarshal"
		value := `
---
a: 1`
		assert.Nil(t, Set(key, value))
		out := make(map[string]int)
		assert.Nil(t, GetAndUnmarshal(key, &out, WithUnmarshaler(&testYamlUnmarshal{})))
		if out["a"] != 1 {
			t.Error("test withUnmarshal fail")
			t.FailNow()
		}
	})
	t.Run("WithDelimiter", func(t *testing.T) {
		key := "test_with_delimiter"
		value := []string{"a", "b", "c", "d"}
		assert.Nil(t, Set(key, strings.Join(value, "|")))
		out := GetStrings(key, []string{}, WithDelimiter("|"))
		if len(out) != len(value) {
			t.Errorf("test with delimiter fail length not equal, current: %d, want: %d", len(out), len(value))
			t.FailNow()
		}
		for index, v := range out {
			if v != value[index] {
				t.Errorf("test with delimiter fail, current: %s, want: %s", v, value[index])
				t.FailNow()
			}
		}
	})
	t.Run("WithIgnoreCase", func(t *testing.T) {})
	t.Run("WithChain", func(t *testing.T) {
		keys := []string{"test_chain_1", "test_chain_2", "test_chain_3"}
		values := []string{"test_chain_1_value", "test_chain_2_value", "test_chain_3_value"}
		for index, k := range keys {
			assert.Nil(t, Set(k, values[index]))
		}
		assert.Equal(t, values[len(values)-1], GetString(keys[0], "", WithChain(keys[1:])))

	})
}

//gocyclo:ignore
func TestGetInt(t *testing.T) {
	t.Run("GetInt", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect int
		}{
			{key: "testInt", value: 100, expect: 100},
			{key: "testIntStr", value: "100", expect: 100},
			{key: "testIntFloat", value: 100.00, expect: 100},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
		}
		for _, data := range datas {
			out := GetInt(data.key, 0)
			if out != data.expect {
				t.Errorf("test %s fail, out %d want %d", data.key, out, data.expect)
				t.FailNow()
			}
		}
	})
	t.Run("GetInt64", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect int64
		}{
			{key: "testInt", value: 100, expect: 100},
			{key: "testIntStr", value: "100", expect: 100},
			{key: "testIntFloat", value: 100.00, expect: 100},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
		}
		for _, data := range datas {
			out := GetInt64(data.key, 0)
			if out != data.expect {
				t.Errorf("test %s fail, out %d want %d", data.key, out, data.expect)
				t.FailNow()
			}
		}
	})
	t.Run("GetInt32", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect int32
		}{
			{key: "testInt", value: 100, expect: 100},
			{key: "testIntStr", value: "100", expect: 100},
			{key: "testIntFloat", value: 100.00, expect: 100},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
		}
		for _, data := range datas {
			out := GetInt32(data.key, 0)
			if out != data.expect {
				t.Errorf("test %s fail, out %d want %d", data.key, out, data.expect)
				t.FailNow()
			}
		}
	})
	t.Run("GetInts", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect []int
		}{
			{key: "testInt", value: 100, expect: []int{100}},
			{key: "testIntStr", value: "100", expect: []int{100}},
			{key: "testIntFloat", value: 100.00, expect: []int{100}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetInts(data.key, []int{})
			if len(out) != len(data.expect) {
				t.Errorf("test %s len fail,current: %d, want: %d ", data.key, len(out), len(data.expect))
				t.FailNow()
			}
			for index, v := range out {
				if v != data.expect[index] {
					t.Errorf("test %s fail, out %d want %d", data.key, v, data.expect[index])
					t.FailNow()
				}
			}
		}
	})
	t.Run("GetInt64s", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect []int64
		}{
			{key: "testInt", value: 100, expect: []int64{100}},
			{key: "testIntStr", value: "100", expect: []int64{100}},
			{key: "testIntFloat", value: 100.00, expect: []int64{100}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetInt64s(data.key, []int64{})
			if len(out) != len(data.expect) {
				t.Errorf("test %s len fail,current: %d, want: %d ", data.key, len(out), len(data.expect))
				t.FailNow()
			}
			for index, v := range out {
				if v != data.expect[index] {
					t.Errorf("test %s fail, out %d want %d", data.key, v, data.expect[index])
					t.FailNow()
				}
			}
		}
	})
	t.Run("GetInt32s", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect []int32
		}{
			{key: "testInt", value: 100, expect: []int32{100}},
			{key: "testIntStr", value: "100", expect: []int32{100}},
			{key: "testIntFloat", value: 100.00, expect: []int32{100}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetInt32s(data.key, []int32{})
			if len(out) != len(data.expect) {
				t.Errorf("test %s len fail,current: %d, want: %d ", data.key, len(out), len(data.expect))
				t.FailNow()
			}
			for index, v := range out {
				if v != data.expect[index] {
					t.Errorf("test %s fail, out %d want %d", data.key, v, data.expect[index])
					t.FailNow()
				}
			}
		}
	})
}

//gocyclo:ignore
func TestGetFloat(t *testing.T) {
	t.Run("GetFloat32", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect float32
		}{
			{key: "testInt", value: 100, expect: 100},
			{key: "testIntStr", value: "100", expect: 100},
			{key: "testIntFloat", value: 100.00, expect: 100},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
		}
		for _, data := range datas {
			out := GetFloat32(data.key, 0)
			if out != data.expect {
				t.Errorf("test %s fail, out %f want %f", data.key, out, data.expect)
				t.FailNow()
			}
		}
	})
	t.Run("GetFloat64", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect float64
		}{
			{key: "testInt", value: 100, expect: 100},
			{key: "testIntStr", value: "100", expect: 100},
			{key: "testIntFloat", value: 100.00, expect: 100},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
		}
		for _, data := range datas {
			out := GetFloat64(data.key, 0)
			if out != data.expect {
				t.Errorf("test %s fail, out %f want %f", data.key, out, data.expect)
				t.FailNow()
			}
		}
	})
	t.Run("GetFloat32s", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect []float32
		}{
			{key: "testInt", value: 100, expect: []float32{100}},
			{key: "testIntStr", value: "100", expect: []float32{100}},
			{key: "testIntFloat", value: 100.00, expect: []float32{100}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetFloat32s(data.key, []float32{})
			if len(out) != len(data.expect) {
				t.Errorf("test %s len fail,current: %d, want: %d ", data.key, len(out), len(data.expect))
				t.FailNow()
			}
			for index, v := range out {
				if v != data.expect[index] {
					t.Errorf("test %s fail, out %f want %f", data.key, v, data.expect[index])
					t.FailNow()
				}
			}
		}
	})
	t.Run("GetFloat64s", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect []float64
		}{
			{key: "testInt", value: 100, expect: []float64{100}},
			{key: "testIntStr", value: "100", expect: []float64{100}},
			{key: "testIntFloat", value: 100.00, expect: []float64{100}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetFloat64s(data.key, []float64{})
			if len(out) != len(data.expect) {
				t.Errorf("test %s len fail,current: %d, want: %d ", data.key, len(out), len(data.expect))
				t.FailNow()
			}
			for index, v := range out {
				if v != data.expect[index] {
					t.Errorf("test %s fail, out %f want %f", data.key, v, data.expect[index])
					t.FailNow()
				}
			}
		}
	})
}

func TestGetTime(t *testing.T) {
	if err := Set("test_time", time.Now()); err != nil {
		t.Error(err)
		t.FailNow()
	}
	out := GetTime("test_time", time.Now().AddDate(0, 0, 1))
	if out.After(time.Now()) {
		t.Errorf("get time fail, current: %s, now: %s", out.String(), time.Now())
		t.FailNow()
	}
}

func TestExist(t *testing.T) {
	if err := Set("test_exist_key", "exist"); err != nil {
		t.Error(err)
		t.FailNow()
	}
	assert.Equal(t, true, Exist("test_exist_key"))
	assert.Equal(t, false, Exist("test_not_found_key"))
}

func TestGetAndUnmarshal(t *testing.T) {
	t.Run("JsonUnmarshal", func(t *testing.T) {
		if err := Set("test_json_content", `{"a":1}`); err != nil {
			t.Error(err)
			t.FailNow()
		}
		out := make(map[string]int)
		if err := GetAndUnmarshal("test_json_content", &out); err != nil {
			t.Error(err)
			t.FailNow()
		}
		if out["a"] != 1 {
			t.Errorf("test json unmarshal fail, current: %d, want: %d", out["a"], 1)
			t.FailNow()
		}
	})
	t.Run("YamlUnmarshal", func(t *testing.T) {
		content := `
---
a: 1`
		if err := Set("test_yaml_content", content); err != nil {
			t.Error(err)
			t.FailNow()
		}
		out := make(map[string]int)
		if err := GetAndYamlUnmarshal("test_yaml_content", &out); err != nil {
			t.Error(err)
			t.FailNow()
		}
		if out["a"] != 1 {
			t.Errorf("test unmarshal yaml content fail, current: %d, want: %d", out["a"], 1)
			t.FailNow()
		}
	})

}

func TestGetDuration(t *testing.T) {

	datas := []struct {
		key    string
		value  any
		expect time.Duration
	}{
		{key: "testDuration", value: "10s", expect: time.Second * 10},
		{key: "testDurationMin", value: "10m", expect: time.Minute * 10},
		{key: "testDurationInt", value: 10, expect: time.Duration(10)},
	}
	for _, data := range datas {
		if err := Set(data.key, data.value); err != nil {
			t.Errorf("set data fail %s", err.Error())
			t.FailNow()
		}
	}
	for _, data := range datas {
		out := GetDuration(data.key, 0)
		if out != data.expect {
			t.Errorf("test %v fail, out %d want %d", data.value, out, data.expect)
			t.FailNow()
		}
	}
}

func TestGetBool(t *testing.T) {
	t.Run("GetBool", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect bool
		}{
			{key: "testBool", value: true, expect: true},
			{key: "testBool1", value: false, expect: false},
			{key: "testBoolStr", value: "true", expect: true},
			{key: "testBoolStr1", value: "True", expect: true},
			{key: "testBoolStr2", value: "TRUE", expect: true},
			{key: "testBoolStr3", value: "yes", expect: true},
			{key: "testBoolStr4", value: "Yes", expect: true},
			{key: "testBoolStr5", value: "YES", expect: true},
			{key: "testBoolStr6", value: "1", expect: true},
			{key: "testBoolStr7", value: "2", expect: true},
			{key: "testBoolStr8", value: "0", expect: false},
			{key: "testBoolInt", value: 1, expect: true},
			{key: "testBoolInt1", value: 2, expect: true},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetBool(data.key, !data.expect)
			if out != data.expect {
				t.Errorf("test key %s, value %v fail, out %v want %v", data.key, data.value, out, data.expect)
				t.FailNow()
			}
		}
	})
	t.Run("GetBools", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect []bool
		}{
			{key: "testBool", value: true, expect: []bool{true}},
			{key: "testBool1", value: false, expect: []bool{false}},
			{key: "testBool2", value: []bool{true}, expect: []bool{}},
			{key: "testBoolStr", value: "true", expect: []bool{true}},
			{key: "testBoolStr01", value: "true,false", expect: []bool{true, false}},
			{key: "testBoolStr1", value: "True", expect: []bool{true}},
			{key: "testBoolStr2", value: "TRUE", expect: []bool{true}},
			{key: "testBoolStr3", value: "yes", expect: []bool{true}},
			{key: "testBoolStr4", value: "Yes", expect: []bool{true}},
			{key: "testBoolStr5", value: "YES", expect: []bool{true}},
			{key: "testBoolStr6", value: "1", expect: []bool{true}},
			{key: "testBoolStr7", value: "2", expect: []bool{true}},
			{key: "testBoolStr8", value: "0", expect: []bool{false}},
			{key: "testBoolInt", value: 1, expect: []bool{true}},
			{key: "testBoolInt1", value: 2, expect: []bool{true}},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
			out := GetBools(data.key, []bool{})
			if len(out) != len(data.expect) {
				t.Errorf("test %s fail, length not equal, current %d want %d", data.key, len(out), len(data.expect))
				t.FailNow()
			}
			for index, v := range out {
				if v != data.expect[index] {
					t.Errorf("test %s fail, current %v, want %v", data.key, v, data.expect[index])
					t.FailNow()
				}
			}
		}
	})
}

// 异步更新配置
func asyncUpdateConfig(exit <-chan struct{}) {
	i := 0
	for {
		select {
		case _, ok := <-exit:
			if !ok {
				return
			}
			return
		case <-time.After(10 * time.Millisecond):
			Set(fmt.Sprintf("async_insert_%d", i), i)
			i++
		}

	}
}

func BenchmarkGet(b *testing.B) {
	exit := make(chan struct{})
	go asyncUpdateConfig(exit)
	b.Run("GetString", func(b *testing.B) {
		b.Run("Exist-Key", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				GetString("testStr", "")
			}
		})
		b.Run("Not-Exist-Key", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				GetString("bench_not_exist_key", "")
			}
		})
	})
	b.Run("GetBytes", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetByte("testStr", []byte{})
		}
	})
	b.Run("GetBool", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			GetBool("testStr", true)
		}
	})
	close(exit)
}

func BenchmarkSetGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// b.StopTimer()
		istr := strconv.Itoa(i)
		key := "bench_test_set_get_key" + istr
		value := "bench_test_set_get_value" + istr
		Set(key, value)
		// b.StartTimer()

		for GetString(key, "") != "" {
			break
		}
	}
}
