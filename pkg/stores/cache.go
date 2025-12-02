package stores

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
)

// Cacher defines the functions cache need to perform
type Cacher interface {
	// get data from cache.
	// if current cache dependence on other cache
	// need return if not data from current cache flag
	// if get data from current cache it will excute refresh cache logic.
	Get(ctx context.Context, key string, out any) (fromCurrent bool, err error)
	// remove data from cache.
	Del(ctx context.Context, keys ...string) error
	// set data in cache.
	Set(ctx context.Context, key string, in any, expiresIn time.Duration) error
	// refresh cache expire time.
	Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) error
	// return cache key.
	Key() string

	// this functions at blow will be implement in default cache.
	// is cache enabled.
	Enabled() bool
	// whether refresh cache expire time.
	AutoRefresh() bool
	// return cache expire time.
	ExpiresIn() time.Duration
	// return empty cache expire time.
	// if get from database is empty, it will set an empty data in cache.
	EmptyExpiresIn() time.Duration
}

// CacheConfig defines the cache config.
type CacheConfig struct {
	// enable or disable cache
	Enabled bool `json:"enabled"`
	// whether refresh cache
	AutoRefresh bool `json:"autoRefresh"`
	// ignore version different in cache key.
	// default will take version tag in cache key.
	// in some scence, cache no need to change in different service version.
	IgnoreVersionDiff bool `json:"ignoreVersionDiff"`
	// ignore service different in cache key.
	// default will take app tag in cache key.
	// if set true, you can share cache in different product.
	IgnoreAppDiff bool `json:"ignoreAppDiff"`
	// ignore environment different in cache key.
	// default will take env tag in cache key.
	// if set true, you can share cache in different environment.
	IgnoreEnvDiff bool `json:"ignoreEnvDiff"`
	// ignore service different in cache key.
	// default will take service tag in cache key.
	// if set true, you can share cache in different service.
	IgnoreServiceDiff bool `json:"ignoreServiceDiff"`
	// ignore region different in cache key.
	IgnoreRegionDiff bool `json:"ignoreRegionDiff"`
	// ignore az different in cache key
	IgnoreAzDiff bool `json:"ignoreAzDiff"`
	// cache expire time
	ExpiresIn utils.JSONDuration `json:"expiresIn"`
	// empty cache expire time
	// if empty it will set to half of ExpiresIn
	EmptyExpiresIn utils.JSONDuration `json:"emptyExpiresIn"`
}

// Cache defines some common functions on this struct, embed on other cache.
type Cache struct {
	// config read write lock, protect conf field
	cm sync.RWMutex
	// cache config
	conf  *CacheConfig
	model Modeler
	app   runtime.APP
}

var (
	// DefaultCacheConfig default cache config
	DefaultCacheConfig = CacheConfig{
		ExpiresIn:         utils.JSONDuration{Duration: 10 * time.Minute},
		IgnoreVersionDiff: true,
	}
)

// NewCache create a new cache.
func NewCache(model Modeler) *Cache {
	return &Cache{
		model: model,
		app:   runtime.GetAPP(),
		conf:  &DefaultCacheConfig,
	}
}

// WithConf set cache config.
func (c *Cache) WithConf(conf *CacheConfig) *Cache {
	c.cm.Lock()
	defer c.cm.Unlock()
	c.conf = conf
	return c
}

// Enabled return is cache enabled.
func (c *Cache) Enabled() bool {
	c.cm.RLock()
	defer c.cm.RUnlock()
	return c.conf.Enabled
}

// AutoRefresh return whether refresh cache.
func (c *Cache) AutoRefresh() bool {
	c.cm.RLock()
	defer c.cm.RUnlock()
	return c.conf.AutoRefresh
}

// NewKey return cache key.
func (c *Cache) NewKey(key string) string {
	c.cm.RLock()
	defer c.cm.RUnlock()
	return c.app.ResourceKey("caches", c.ModelKey(key),
		runtime.WithDelimiter(":"),
		runtime.WithoutVersion(c.conf.IgnoreVersionDiff),
		runtime.WithoutApp(c.conf.IgnoreAppDiff),
		runtime.WithoutEnv(c.conf.IgnoreEnvDiff),
		runtime.WithoutService(c.conf.IgnoreServiceDiff),
		runtime.WithoutRegion(c.conf.IgnoreRegionDiff),
		runtime.WithoutApp(c.conf.IgnoreAzDiff))
}

// App return service info.
func (c *Cache) App() runtime.APP {
	return c.app
}

// ModelKey return model cache key.
func (c *Cache) ModelKey(key string) string {
	return c.model.ModelName() + ":" + key
}

// ExpiresIn return cache expire time
// it will add any random time in confined expire time.
func (c *Cache) ExpiresIn() time.Duration {
	c.cm.RLock()
	defer c.cm.RUnlock()
	return c.conf.ExpiresIn.Duration + time.Duration(rand.Int63n(int64(c.conf.ExpiresIn.Duration)))
}

// EmptyExpiresIn return empty cache expire time
// if not config it will be setted to half of ExpiresIn
func (c *Cache) EmptyExpiresIn() time.Duration {
	c.cm.RLock()
	defer c.cm.RUnlock()
	expiresIn := c.conf.EmptyExpiresIn.Duration
	if expiresIn == 0 {
		expiresIn = c.conf.ExpiresIn.Duration / 2
	}
	return expiresIn + time.Duration(rand.Int63n(int64(expiresIn)))
}
