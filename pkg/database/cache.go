package database

import (
	"time"

	"github.com/asjard/asjard/pkg/database/redis"
)

// cacheType 缓存类型
type cacheType int

const (
	// CacheTypeKeyValue key-value缓存
	cacheTypeKeyValue cacheType = iota
	// CacheTypeHash hash缓存
	cacheTypeHash
	// CacheTypeSet 集合缓存
	cacheTypeSet
)

// Cacher 缓存需要实现的方法
type Cacher interface {
	// 从缓存获取数据
	get(out interface{}) error
	// 从缓存删除数据
	del() error
	// 设置缓存数据
	set(in interface{}) error
	// 刷新缓存过期时间
	refreshExpire() error
	// 设置空值到缓存
	setEmpty(in interface{}) error
}

// Cache 缓存相关
type Cache struct {
	// 缓存key
	key string
	// hash中的field， set中的member
	field string
	// 缓存类型
	tp cacheType

	// 缓存过期时间
	timeout time.Duration
	// 是否自动刷新缓存
	autoRefresh bool

	// redis
	rds *redis.Client

	// redis缓存
	rc *RedisCache
	// 本地缓存
	lc *LocalCache
}

// NewKeyValueCache key-value缓存
func NewKeyValueCache(key string) *Cache {
	return newCache(cacheTypeKeyValue, key)
}

// NewHashCache hash缓存
func NewHashCache(key, field string) *Cache {
	return newCache(cacheTypeHash, key).
		withField(field)
}

// NewSetCache set缓存
func NewSetCache(key, member string) *Cache {
	return newCache(cacheTypeSet, key).
		withField(member)
}

// newCache 创建新缓存
func newCache(tp cacheType, key string) *Cache {
	return &Cache{
		key: key,
		tp:  tp,
	}
}

// WithExpiresIn 缓存过期时间
func (c *Cache) WithExpiresIn(timeout time.Duration) *Cache {
	c.timeout = timeout
	return c
}

// WithAutoRefresh 自动刷新
func (c *Cache) WithAutoRefresh(autoRefresh bool) *Cache {
	c.autoRefresh = autoRefresh
	return c
}

// WithRedis 设置redis
func (c *Cache) WithRedis(rds *redis.Client) *Cache {
	c.rds = rds
	return c
}

func (c *Cache) withField(field string) *Cache {
	c.field = field
	return c
}

// 从缓存获取数据
func (c Cache) get(out interface{}) error {
	return nil
}

// 删除缓存数据
func (c Cache) del() error {
	return nil
}

// 设置缓存数据
func (c Cache) set(in interface{}) error {
	return nil
}

// 刷新缓存数据
func (c Cache) refreshExpire() error {
	return nil
}

// 设置空数据
func (c Cache) setEmpty(in interface{}) error {
	return nil
}
