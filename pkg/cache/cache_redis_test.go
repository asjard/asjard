package cache

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/pkg/stores/xredis"
	"github.com/stretchr/testify/assert"
)

type testTable struct{}

func (testTable) ModelName() string {
	return "test_local_cache_model"
}

func TestMain(m *testing.M) {
	if err := config.Load(-1); err != nil {
		panic(err)
	}
	config.Set("asjard.stores.redis.clients.default.address", "127.0.0.1:6379")
	config.Set("asjard.cache.redis.enabled", true)

	if err := bootstrap.Bootstrap(); err != nil {
		panic(err)
	}
	m.Run()
}

func TestRedisCache(t *testing.T) {
	client, err := xredis.Client()
	assert.Nil(t, err)
	assert.NotNil(t, client)
	t.Run("TestKeyValue", func(t *testing.T) {
		cache, err := NewRedisKeyValueCache(&testTable{})
		assert.Nil(t, err)
		testKey := "test_redis_key"
		testValue := "test_redis_value"
		t.Run("TestKey", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("%s_%d", testKey, i)
				value := fmt.Sprintf("%s_%d", testValue, i)
				assert.Nil(t, cache.Set(context.Background(), key, value, 5*time.Minute))
				setCheck := client.Get(context.Background(), key)
				assert.Nil(t, setCheck.Err())

				assert.Equal(t, strconv.Quote(value), setCheck.Val())
				ttl := client.TTL(context.Background(), key)
				assert.Nil(t, ttl.Err())
				assert.NotZero(t, ttl.Val())

				var result string
				_, err := cache.Get(context.Background(), key, &result)
				assert.Nil(t, err)
				assert.Equal(t, value, result)

				// 删除
				assert.Nil(t, cache.Del(context.Background(), key))
				// 删除检查
				existCheck := client.Get(context.Background(), key)
				assert.NotNil(t, existCheck.Err())

				result = ""
				_, err = cache.Get(context.Background(), key, &result)
				assert.NotNil(t, err)

			}
		})
		t.Run("TestWithGroup", func(t *testing.T) {
			testGroup := "test_redis_group"
			cache = cache.WithGroup(testGroup)
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("%s_%d", testKey, i)
				value := fmt.Sprintf("%s_%d", testValue, i)
				assert.Nil(t, cache.Set(context.Background(), key, value, 5*time.Minute))
				hcheck := client.HGet(context.Background(), cache.Group(testGroup), key)
				assert.Nil(t, hcheck.Err())
				assert.NotEmpty(t, hcheck.Val())

				assert.Nil(t, cache.Del(context.Background(), key))
				groupExist := client.Get(context.Background(), cache.Group(testGroup))
				assert.NotNil(t, groupExist.Err())
				keyExist := client.Get(context.Background(), key)
				assert.NotNil(t, keyExist.Err())
			}
		})
		t.Run("TestKey", func(t *testing.T) {
			// default is ignore
			assert.NotContains(t, cache.WithKey("test_version").Key(), "1.0.0")
			config.Set("asjard.cache.redis.ignoreVersionDiff", true)
			time.Sleep(500 * time.Millisecond)
			assert.NotContains(t, cache.WithKey("test_version").Key(), "1.0.0")
			config.Set("asjard.cache.redis.ignoreVersionDiff", false)
			time.Sleep(500 * time.Millisecond)
			assert.Contains(t, cache.WithKey("test_version").Key(), "1.0.0")
		})
	})
}
