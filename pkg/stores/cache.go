package stores

import (
	"context"
	"math/rand"
	"time"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
)

// Cacher 缓存需要实现的方法
type Cacher interface {
	// 从缓存获取数据
	// fromCurrent 获取到的数据是从当前缓存中获取到的
	Get(ctx context.Context, key string, out any) (fromCurrent bool, err error)
	// 从缓存删除数据
	Del(ctx context.Context, keys ...string) error
	// 设置缓存数据
	Set(ctx context.Context, key string, in any, expiresIn time.Duration) error
	// 刷新缓存过期时间
	Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) error
	// 返回缓存Key名称
	Key() string

	// 是否开启了缓存
	Enabled() bool
	// 是否自动刷新缓存
	AutoRefresh() bool
	// 过期时间
	ExpiresIn() time.Duration
	// 空值过期时间
	EmptyExpiresIn() time.Duration
}

// 缓存配置
type CacheConfig struct {
	// 是否开启缓存
	Enabled bool `json:"enabled"`
	// 是否自动刷新
	AutoRefresh bool `json:"autoRefresh"`
	// 缓存过期时间
	ExpiresIn utils.JSONDuration `json:"expiresIn"`
	// 空值缓存过期时间
	EmptyExpiresIn utils.JSONDuration `json:"emptyExpiresIn"`
}

// Cache 缓存相关
type Cache struct {
	// 缓存配置
	conf *CacheConfig
	// 缓存表
	model Modeler
	app   runtime.APP
}

var (
	// DefaultCacheConfig 默认配置
	DefaultCacheConfig = CacheConfig{
		ExpiresIn: utils.JSONDuration{Duration: 10 * time.Minute},
	}
)

// NewCache 创建新缓存
func NewCache(model Modeler) *Cache {
	return &Cache{
		model: model,
		app:   runtime.GetAPP(),
		conf:  &DefaultCacheConfig,
	}
}

// WithConf 设置配置文件
func (c *Cache) WithConf(conf *CacheConfig) *Cache {
	c.conf = conf
	return c
}

// Enabled 否开启缓存
func (c *Cache) Enabled() bool {
	return c.conf.Enabled
}

// AutoRefresh 是否自动刷新
func (c *Cache) AutoRefresh() bool {
	return c.conf.AutoRefresh
}

// NewKey 缓存key
func (c *Cache) NewKey(key string) string {
	return c.app.ResourceKey("caches", c.ModelKey(key), runtime.WithDelimiter(":"))
}

// App 返回app信息
func (c *Cache) App() runtime.APP {
	return c.app
}

// ModelKey key组合
func (c *Cache) ModelKey(key string) string {
	return c.model.ModelName() + ":" + key
}

// ExpiresIn 缓存过期时间
// 添加随机事件
func (c *Cache) ExpiresIn() time.Duration {
	return c.conf.ExpiresIn.Duration + time.Duration(rand.Int63n(int64(c.conf.ExpiresIn.Duration)))
}

// EmptyExpiresIn 空值缓存过期时间
// 添加随机时间
func (c *Cache) EmptyExpiresIn() time.Duration {
	expiresIn := c.conf.EmptyExpiresIn.Duration
	if expiresIn == 0 {
		expiresIn = c.conf.ExpiresIn.Duration / 2
	}
	return expiresIn + time.Duration(rand.Int63n(int64(expiresIn)))
}
