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
	t.Run("GetInts", func(t *testing.T) {

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
				t.Errorf("test %s fail, lenght not equal, current %d want %d", data.key, len(out), len(data.expect))
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
