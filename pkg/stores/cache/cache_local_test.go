package cache

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testModel struct {
}

func (testModel) ModelName() string {
	return "test_local_cache_model"
}

func TestCache(t *testing.T) {
	localCache, err := NewLocalCache(&testModel{})
	assert.Nil(t, err)
	testKey := "test_key"
	testValue := "test_value"
	for i := 0; i < 5000; i++ {
		key := fmt.Sprintf("%s_%d", testKey, i)
		value := fmt.Sprintf("%s_%d", testValue, i)
		assert.Nil(t, localCache.Set(context.Background(), key, value))
		var result string
		assert.Nil(t, localCache.Get(context.Background(), key, &result))
		assert.Equal(t, value, result)
		assert.Nil(t, localCache.Del(context.Background(), key))
		result = ""
		assert.NotNil(t, localCache.Get(context.Background(), key, &result))
	}
}
