package xredis

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/asjard/asjard/core/bootstrap"
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
	config.Set("asjard.stores.redis.clients.default.address", "127.0.0.1:6379")
	config.Set("asjard.stores.redis.clients.cipher.address", "MTI3LjAuMC4xOjYzNzk=")
	config.Set("asjard.stores.redis.clients.cipher.cipherName", "base64")

	if err := bootstrap.Bootstrap(); err != nil {
		panic(err)
	}
	m.Run()
	clientManager.Stop()
}

func TestNewClients(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		client, err := Client()
		t.Log(err)
		assert.Nil(t, err)
		assert.NotNil(t, client)
		result := client.Set(context.Background(), "test_default_redis_key", "test_default_redis_value", 5*time.Second)
		assert.Nil(t, result.Err())
		delResult := client.Del(context.Background(), "test_default_redis_key")
		assert.Nil(t, delResult.Err())
	})
	t.Run("cipher", func(t *testing.T) {
		client, err := Client(WithClientName("cipher"))
		t.Log(err)
		assert.Nil(t, err)
		assert.NotNil(t, client)
		result := client.Set(context.Background(), "test_cipher_redis_key", "test_cipher_redis_value", 5*time.Second)
		assert.Nil(t, result.Err())
		delResult := client.Del(context.Background(), "test_cipher_redis_key")
		assert.Nil(t, delResult.Err())
	})
	t.Run("another", func(t *testing.T) {
		s := miniredis.RunT(t)
		config.Set("asjard.stores.redis.clients.another.address", "127.0.0.1:6379")
		time.Sleep(time.Second)
		client, err := Client(WithClientName("another"))
		assert.Nil(t, err)
		assert.NotNil(t, client)
		result := client.Set(context.Background(), "test_another_redis_key", "test_another_redis_value", 5*time.Second)
		assert.Nil(t, result.Err())
		delResult := client.Del(context.Background(), "test_another_redis_key")
		assert.Nil(t, delResult.Err())
		s.Close()
	})
	t.Run("reuse_unchanged", func(t *testing.T) {
		c1, err := Client()
		require.NoError(t, err)
		clientManager.watch(&config.Event{Key: "asjard.stores.redis.clients.default.address"})
		c2, err := Client()
		require.NoError(t, err)
		require.Same(t, c1, c2)
	})
	t.Run("shudown", func(t *testing.T) {
		clientManager.Stop()
		client, err := Client()
		assert.NotNil(t, err)
		assert.Nil(t, client)
	})
}

func TestNewClientsReuseAndReplace(t *testing.T) {
	s := miniredis.RunT(t)
	m := &ClientManager{configs: make(map[string]*ClientConnConfig)}
	conf := &ClientConnConfig{
		Address: s.Addr(),
		Options: Options{DisableCheckStatus: true},
	}
	require.NoError(t, m.newClients(map[string]*ClientConnConfig{"default": conf}))
	first, ok := m.clients.Load("default")
	require.True(t, ok)

	require.NoError(t, m.newClients(map[string]*ClientConnConfig{"default": {
		Address: s.Addr(),
		Options: Options{DisableCheckStatus: true},
	}}))
	second, ok := m.clients.Load("default")
	require.True(t, ok)
	require.Same(t, first, second)

	s2 := miniredis.RunT(t)
	require.NoError(t, m.newClients(map[string]*ClientConnConfig{"default": {
		Address: s2.Addr(),
		Options: Options{DisableCheckStatus: true},
	}}))
	replaced, ok := m.clients.Load("default")
	require.True(t, ok)
	require.NotSame(t, first, replaced)
	m.Stop()
}

func TestClientOptionsAndRedisConfig(t *testing.T) {
	opts := defaultClientOptions()
	require.Equal(t, DefaultClientName, opts.clientName)
	WithClientName("named")(opts)
	require.Equal(t, "named", opts.clientName)
	WithClientName("")(opts)
	require.Equal(t, "named", opts.clientName)

	conf := &ClientConnConfig{Address: "127.0.0.1:6379", UserName: "user", Password: "pass", DB: 2, Options: Options{
		ClientName: "client", DialTimeout: utils.JSONDuration{Duration: time.Second}, PoolFIFO: true,
	}}
	got, err := (&ClientManager{}).newClientOptions(conf)
	require.NoError(t, err)
	require.Equal(t, conf.Address, got.Addr)
	require.Equal(t, conf.UserName, got.Username)
	require.Equal(t, conf.Password, got.Password)
	require.Equal(t, conf.DB, got.DB)
	require.True(t, got.PoolFIFO)
}

func TestRedisMissingTLSFiles(t *testing.T) {
	conf := &ClientConnConfig{Options: Options{CAFile: "missing-ca", CertFile: "missing-cert", KeyFile: "missing-key"}}
	_, err := (&ClientManager{}).newClientOptions(conf)
	require.Error(t, err)
}
