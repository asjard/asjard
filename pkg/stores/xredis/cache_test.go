package xredis

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testTable struct{}

func (testTable) ModelName() string {
	return "test_local_cache_model"
}

func TestCache(t *testing.T) {
	client, err := Client()
	assert.Nil(t, err)
	assert.NotNil(t, client)
	t.Run("TestKeyValue", func(t *testing.T) {
		cache, err := NewKeyValueCache(&testTable{})
		assert.Nil(t, err)
		testKey := "test_redis_key"
		testValue := "test_redis_value"
		t.Run("TestKey", func(t *testing.T) {
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("%s_%d", testKey, i)
				value := fmt.Sprintf("%s_%d", testValue, i)
				assert.Nil(t, cache.Set(context.Background(), key, value))
				setCheck := client.Get(context.Background(), key)
				assert.Nil(t, setCheck.Err())

				assert.Equal(t, strconv.Quote(value), setCheck.Val())
				ttl := client.TTL(context.Background(), key)
				assert.Nil(t, ttl.Err())
				assert.NotZero(t, ttl.Val())

				var result string
				assert.Nil(t, cache.Get(context.Background(), key, &result))
				assert.Equal(t, value, result)

				// 删除
				assert.Nil(t, cache.Del(context.Background(), key))
				// 删除检查
				existCheck := client.Get(context.Background(), key)
				assert.NotNil(t, existCheck.Err())

				result = ""
				assert.NotNil(t, cache.Get(context.Background(), key, &result))

			}
		})
		t.Run("TestWithGroup", func(t *testing.T) {
			testGroup := "test_redis_group"
			cache = cache.WithGroup(testGroup)
			for i := 0; i < 100; i++ {
				key := fmt.Sprintf("%s_%d", testKey, i)
				value := fmt.Sprintf("%s_%d", testValue, i)
				assert.Nil(t, cache.Set(context.Background(), key, value))
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
	})
}
