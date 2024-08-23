package xetcd

import (
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	config.Set("asjard.stores.etcd.clients.default.endpoints", "127.0.0.1:2379")
	config.Set("asjard.stores.etcd.clients.another.endpoints", "127.0.0.1:2379")
	time.Sleep(50 * time.Millisecond)
	if err := clientManager.Start(); err != nil {
		panic(err)
	}
	m.Run()
	clientManager.Stop()
}

func TestNewClients(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		client, err := Client()
		assert.Nil(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config.GetStrings("asjard.stores.etcd.clients.default.endpoints", []string{}), client.Endpoints())
	})
	t.Run("another", func(t *testing.T) {
		client, err := Client(WithClientName("another"))
		assert.Nil(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, config.GetStrings("asjard.stores.etcd.clients.another.endpoints", []string{}), client.Endpoints())
	})

	t.Run("new", func(t *testing.T) {
		config.Set("asjard.stores.etcd.clients.new.endpoints", "127.0.0.1:2379")
		time.Sleep(200 * time.Millisecond)
		_, err := Client(WithClientName("new"))
		if err != nil {
			t.Error(err.Error())
			t.FailNow()
		}
		assert.Nil(t, err)
	})
	t.Run("shutdown", func(t *testing.T) {
		clientManager.Stop()
		_, err := Client()
		assert.NotNil(t, err)
	})
}
