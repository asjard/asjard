package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

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
		assert.Nil(t, localCache.Set(context.Background(), key, value, 5*time.Minute))
		var result string
		_, err := localCache.Get(context.Background(), key, &result)
		assert.Nil(t, err)
		assert.Equal(t, value, result)
		assert.Nil(t, localCache.Del(context.Background(), key))
		result = ""
		_, err = localCache.Get(context.Background(), key, &result)
		assert.NotNil(t, err)
	}
}
