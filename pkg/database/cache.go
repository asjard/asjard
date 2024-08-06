package database

import (
	"context"
	"time"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
)

// Cacher 缓存需要实现的方法
type Cacher interface {
	// 从缓存获取数据
	Get(ctx context.Context, key string, out any) error
	// 从缓存删除数据
	Del(ctx context.Context, keys ...string) error
	// 设置缓存数据
	Set(ctx context.Context, key string, in any) error
	// 刷新缓存过期时间
	Refresh(ctx context.Context, key string) error

	// 返回缓存Key名称
	Key() string
	// 是否开启了缓存
	Enabled() bool
	// 是否自动刷新缓存
	AutoRefresh() bool
}

// 缓存配置
type CacheConfig struct {
	// 是否开启缓存
	Enabled bool `json:"enabled"`
	// 是否自动刷新
	AutoRefresh bool `json:"autoRefresh"`
	// 缓存过期时间
	ExpiresIn utils.JSONDuration `json:"expiresIn"`
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
		Enabled:     true,
		AutoRefresh: true,
		ExpiresIn:   utils.JSONDuration{Duration: 5 * time.Minute},
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

// Prefix 缓存前缀
// {app}:caches:service:{service}:{region}:{az}:{model_name}
func (c *Cache) Prefix() string {
	return c.app.App + ":caches:service:" + c.app.Instance.Name + ":" + c.app.Region + ":" + c.app.AZ + ":" + c.model.ModelName()
}

// NewKey 缓存key
func (c *Cache) NewKey(key string) string {
	return c.Prefix() + ":" + key
}

// ExpiresIn 缓存过期时间
func (c *Cache) ExpiresIn() time.Duration {
	return c.conf.ExpiresIn.Duration
}
