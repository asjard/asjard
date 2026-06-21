package cache

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/asjard/asjard/core/bootstrap"
	"github.com/asjard/asjard/core/config"
	_ "github.com/asjard/asjard/pkg/config/mem"
	"github.com/asjard/asjard/pkg/stores/xredis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTable struct{}

func (testTable) ModelName() string {
	return "test_local_cache_model"
}

func setupRedisIntegration(t *testing.T) {
	t.Helper()
	config.Set("asjard.stores.redis.clients.default.address", "127.0.0.1:6379")
	config.Set("asjard.cache.redis.enabled", true)

	if err := bootstrap.Bootstrap(); err != nil {
		t.Fatal(err)
	}
}

func TestRedisCache(t *testing.T) {
	setupRedisIntegration(t)
	client, err := xredis.Client()
	assert.Nil(t, err)
	assert.NotNil(t, client)
	t.Run("TestKeyValue", func(t *testing.T) {
		cache, err := NewRedisKeyValueCache(&testTable{})
		assert.Nil(t, err)
		testKey := "test_redis_key"
		testValue := "test_redis_value"
		t.Run("TestKey", func(t *testing.T) {
			for i := range 100 {
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
			for i := range 100 {
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
		t.Run("TestKeyOption", func(t *testing.T) {
			// default is ignore
			assert.NotContains(t, cache.WithKey("test_without_version").Key(), "1.0.0")
			require.NoError(t, config.Set("asjard.cache.redis.careVersionDiff", true))
			require.Eventually(t, func() bool {
				return strings.Contains(cache.WithKey("test_with_version").Key(), "1.0.0")
			}, 3*time.Second, 20*time.Millisecond)
		})
	})
}
