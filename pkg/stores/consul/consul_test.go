package consul

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
	config.Set("asjard.stores.consul.clients.default.address", "127.0.0.1:8500")
	config.Set("asjard.stores.consul.clients.cipher.address", "MTI3LjAuMC4xOjg1MDA=")
	config.Set("asjard.stores.consul.clients.cipher.cipherName", "base64")
	time.Sleep(50 * time.Millisecond)
	if err := clientManager.Start(); err != nil {
		panic(err)
	}
	m.Run()
	clientManager.Stop()
}

func TestClient(t *testing.T) {
	t.Run("DefaultClient", func(t *testing.T) {
		client, err := Client()
		assert.Nil(t, err)
		assert.NotNil(t, client)
		_, err = client.Status().Leader()
		assert.Nil(t, err)
	})
	t.Run("cipher", func(t *testing.T) {
		client, err := Client(WithClientName("cipher"))
		assert.Nil(t, err)
		assert.NotNil(t, client)
		_, err = client.Status().Leader()
		assert.Nil(t, err)
	})
	t.Run("NotFoundClient", func(t *testing.T) {
		_, err := Client(WithClientName("NotFound"))
		assert.NotNil(t, err)
	})
	t.Run("AddNewClient", func(t *testing.T) {
		config.Set("asjard.stores.consul.clients.newAdd.address", "127.0.0.1:8500")
		time.Sleep(50 * time.Millisecond)
		client, err := Client(WithClientName("newAdd"))
		assert.Nil(t, err)
		assert.NotNil(t, client)
		_, err = client.Status().Leader()
		assert.Nil(t, err)
	})
}
