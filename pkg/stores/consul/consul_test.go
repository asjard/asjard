package consul

import (
	"testing"
	"time"

	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestClientOptionsAndAPIConfig(t *testing.T) {
	opts := defaultClientOptions()
	require.Equal(t, DefaultClientName, opts.clientName)
	WithClientName("named")(opts)
	require.Equal(t, "named", opts.clientName)

	conf := &ClientConnConfig{
		Address: "127.0.0.1:8500", Scheme: "http", PathPrefix: "/v1", Datacenter: "dc1",
		Username: "user", Password: "pass", Token: "token", WaitTime: utils.JSONDuration{Duration: time.Second},
	}
	got, err := (&ClientManager{}).newApiConfig(conf)
	require.NoError(t, err)
	require.Equal(t, conf.Address, got.Address)
	require.Equal(t, conf.Scheme, got.Scheme)
	require.Equal(t, conf.PathPrefix, got.PathPrefix)
	require.Equal(t, conf.Datacenter, got.Datacenter)
	require.Equal(t, conf.Token, got.Token)
	require.Equal(t, conf.Username, got.HttpAuth.Username)
	require.Equal(t, conf.Password, got.HttpAuth.Password)
}
