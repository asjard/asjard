package config

import (
	"sync"
	"testing"
	"time"
)

const (
	testSourceName     = "testSource"
	testSourcePriority = 0
)

type testSource struct {
	cb      func(*Event)
	configs map[string]any
	cm      sync.RWMutex
}

func newTestSource() (Sourcer, error) {
	return &testSource{
		configs: map[string]any{
			"testInt":      1,
			"testStr":      "test data",
			"testFloat":    0.01,
			"testDuration": "10m",
			"testBool":     true,
		},
	}, nil
}

// GetAll .
func (s *testSource) GetAll() map[string]*Value {
	configs := make(map[string]*Value)
	s.cm.RLock()
	for key, value := range s.configs {
		configs[key] = &Value{
			Sourcer: s,
			Value:   value,
		}
	}
	s.cm.RUnlock()
	return configs
}

func (s *testSource) Set(key string, value any) error {
	s.cm.Lock()
	s.configs[key] = value
	s.cm.Unlock()
	s.cb(&Event{
		Type: EventTypeUpdate,
		Key:  key,
		Value: &Value{
			Sourcer: s,
			Value:   value,
		},
	})
	return nil
}

func (s *testSource) Watch(callback func(event *Event)) error {
	s.cb = callback
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

func TestGetString(t *testing.T) {
	t.Run("GetString", func(t *testing.T) {
		datas := []struct {
			key    string
			value  any
			expect string
		}{
			{key: "testStr", value: "test_str", expect: "test_str"},
			{key: "testStrInt", value: 1, expect: "1"},
			{key: "testStrFloat", value: 0.01, expect: "0.01"},
			{key: "testStrBool", value: true, expect: "true"},
		}
		for _, data := range datas {
			if err := Set(data.key, data.value); err != nil {
				t.Errorf("set data fail %s", err.Error())
				t.FailNow()
			}
		}
		for _, data := range datas {
			out := GetString(data.key, "")
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
func TestGetInt(t *testing.T) {

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
		{key: "testBoolStr7", value: "2", expect: false},
		{key: "testBoolStr8", value: "0", expect: false},
		{key: "testBoolInt", value: 1, expect: true},
		{key: "testBoolInt1", value: 2, expect: true},
	}
	for _, data := range datas {
		if err := Set(data.key, data.value); err != nil {
			t.Errorf("set data fail %s", err.Error())
			t.FailNow()
		}
	}
	for _, data := range datas {
		out := GetBool(data.key, !data.expect)
		if out != data.expect {
			t.Errorf("test key %s, value %v fail, out %v want %v", data.key, data.value, out, data.expect)
			t.FailNow()
		}
	}
}
