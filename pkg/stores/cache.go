package stores

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/utils"
)

// Cacher defines the standard interface that all cache implementations (Redis, Memcached, Local, etc.) must satisfy.
// It separates the logic of data retrieval from the logic of cache management.
type Cacher interface {
	// Get retrieves data from the cache.
	// out: the pointer where the result will be unmarshaled.
	// fromCurrent: indicates if the data was found in the primary cache layer.
	Get(ctx context.Context, key string, out any) (fromCurrent bool, err error)

	// Del removes one or more keys from the cache.
	Del(ctx context.Context, keys ...string) error

	// Set stores data in the cache with a specific expiration duration.
	Set(ctx context.Context, key string, in any, expiresIn time.Duration) error

	// Refresh extends the expiration time of an existing cache entry.
	Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) error

	// Key returns the identifier for this cache instance.
	Key() string

	// Standard metadata functions for the cache instance.
	Enabled() bool
	AutoRefresh() bool
	ExpiresIn() time.Duration
	EmptyExpiresIn() time.Duration
}

// CacheConfig defines the behavioral settings for a specific cache instance.
// It includes granular controls for how keys are namespaced across different environments and versions.
type CacheConfig struct {
	// Enabled toggles the cache layer on or off.
	Enabled bool `json:"enabled"`

	// AutoRefresh determines if the system should automatically extend TTL on access.
	AutoRefresh bool `json:"autoRefresh"`

	// Key Namespace Overrides:
	// These flags allow sharing cache data across different deployment dimensions.
	IgnoreVersionDiff bool `json:"ignoreVersionDiff"` // Share cache across different app versions
	IgnoreAppDiff     bool `json:"ignoreAppDiff"`     // Share cache across different applications
	IgnoreEnvDiff     bool `json:"ignoreEnvDiff"`     // Share cache across Dev/Staging/Prod
	IgnoreServiceDiff bool `json:"ignoreServiceDiff"` // Share cache across different microservices
	IgnoreRegionDiff  bool `json:"ignoreRegionDiff"`  // Share cache across geographic regions
	IgnoreAzDiff      bool `json:"ignoreAzDiff"`      // Share cache across Availability Zones

	// ExpiresIn is the standard TTL for successful data lookups.
	ExpiresIn utils.JSONDuration `json:"expiresIn"`

	// EmptyExpiresIn is the TTL for "Negative Caching" (caching the absence of data).
	// Prevents "Cache Penetration" by caching null results from the DB.
	EmptyExpiresIn utils.JSONDuration `json:"emptyExpiresIn"`
}

// Cache is the base implementation struct intended to be embedded in specific cache providers.
// It provides helper methods for key generation and configuration management.
type Cache struct {
	// cm protects access to the conf field for thread-safe runtime configuration updates.
	cm   sync.RWMutex
	conf *CacheConfig

	// model provides metadata about the data being cached (e.g., Table Name).
	model Modeler
	// app provides runtime context about the current service instance.
	app runtime.APP
}

var (
	// DefaultCacheConfig provides a safe baseline: 10-minute TTL and version independence.
	DefaultCacheConfig = CacheConfig{
		ExpiresIn:         utils.JSONDuration{Duration: 10 * time.Minute},
		IgnoreVersionDiff: true,
	}
)

// NewCache initializes a basic cache structure with a reference to the data model.
func NewCache(model Modeler) *Cache {
	return &Cache{
		model: model,
		app:   runtime.GetAPP(),
		conf:  &DefaultCacheConfig,
	}
}

// WithConf allows for fluent-style configuration of the cache instance.
func (c *Cache) WithConf(conf *CacheConfig) *Cache {
	c.cm.Lock()
	defer c.cm.Unlock()
	c.conf = conf
	return c
}

// Enabled checks if caching is currently active.
func (c *Cache) Enabled() bool {
	c.cm.RLock()
	defer c.cm.RUnlock()
	return c.conf.Enabled
}

// AutoRefresh checks if the TTL should be renewed on every Get.
func (c *Cache) AutoRefresh() bool {
	c.cm.RLock()
	defer c.cm.RUnlock()
	return c.conf.AutoRefresh
}

// NewKey generates a fully qualified, namespaced cache key.
// It combines the application identity (App, Env, Region) with the model key.
func (c *Cache) NewKey(key string) string {
	c.cm.RLock()
	defer c.cm.RUnlock()
	// Uses the runtime ResourceKey builder to ensure consistent naming conventions.
	return c.app.ResourceKey("caches", c.ModelKey(key),
		runtime.WithDelimiter(":"),
		runtime.WithoutVersion(c.conf.IgnoreVersionDiff),
		runtime.WithoutApp(c.conf.IgnoreAppDiff),
		runtime.WithoutEnv(c.conf.IgnoreEnvDiff),
		runtime.WithoutService(c.conf.IgnoreServiceDiff),
		runtime.WithoutRegion(c.conf.IgnoreRegionDiff),
		runtime.WithoutApp(c.conf.IgnoreAzDiff))
}

// App returns the runtime application information.
func (c *Cache) App() runtime.APP {
	return c.app
}

// ModelKey creates a model-specific suffix for the cache key (e.g., "users:123").
func (c *Cache) ModelKey(key string) string {
	return c.model.ModelName() + ":" + key
}

// ExpiresIn calculates the TTL for a cache entry.
// JITTER: It adds a random duration up to 100% of the config value to prevent "Cache Avalanche"
// (where many keys expire at the exact same time, overwhelming the database).
func (c *Cache) ExpiresIn() time.Duration {
	c.cm.RLock()
	defer c.cm.RUnlock()
	// Total TTL = Configured TTL + Random(0, Configured TTL)
	return c.conf.ExpiresIn.Duration + time.Duration(rand.Int63n(int64(c.conf.ExpiresIn.Duration)))
}

// EmptyExpiresIn calculates the TTL for null/empty results.
// If not explicitly set, it defaults to 50% of the standard ExpiresIn.
func (c *Cache) EmptyExpiresIn() time.Duration {
	c.cm.RLock()
	defer c.cm.RUnlock()
	expiresIn := c.conf.EmptyExpiresIn.Duration
	if expiresIn == 0 {
		expiresIn = c.conf.ExpiresIn.Duration / 2
	}
	// Also includes jitter for the same reasons as ExpiresIn.
	return expiresIn + time.Duration(rand.Int63n(int64(expiresIn)))
}
