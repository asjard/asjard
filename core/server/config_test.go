package server

import (
	"testing"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	m.Run()
}

func TestGetConfig(t *testing.T) {
	testAddress := "127.0.0.1:9909"
	testInterceptor := "testInterceptor"
	testDefaultHandler := "testDefaultHandler"
	config.Set("asjard.servers.interceptors", testInterceptor)
	config.Set("asjard.servers.defaultHandlers", testDefaultHandler)
	config.Set("asjard.servers.testProtocol.addresses.listen", testAddress)
	t.Run("GetConfigWithProtocol", func(t *testing.T) {
		conf := GetConfigWithProtocol("testProtocol")
		if conf.Addresses.Listen != testAddress {
			t.Errorf("get protocol listen address fail, want: %s, act: %s", testAddress, conf.Addresses.Listen)
			t.FailNow()
		}
	})
	t.Run("GetConfig", func(t *testing.T) {
		conf := GetConfig()
		if len(conf.Interceptors) == 0 {
			t.Error("get interceptors is empty")
			t.FailNow()
		}
		if conf.Interceptors[len(conf.Interceptors)-1] != testInterceptor {
			t.Error("testInterceptor not found")
			t.FailNow()
		}
		if len(conf.DefaultHandlers) == 0 {
			t.Error("get defaultHandlers is empty")
			t.FailNow()
		}
		if conf.DefaultHandlers[len(conf.DefaultHandlers)-1] != testDefaultHandler {
			t.Error("testDefaultHandler not found")
			t.FailNow()
		}
	})
}
