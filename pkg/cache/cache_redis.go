package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/stores"
	"github.com/asjard/asjard/pkg/stores/xredis"
	"github.com/asjard/asjard/pkg/tools"
	"github.com/redis/go-redis/v9"
)

// CacheRedisType specifies which Redis data structure to use for storage.
type CacheRedisType uint

const (
	// CacheRedisTypeKeyValue standard Redis string (SET/GET).
	CacheRedisTypeKeyValue CacheRedisType = iota
	// CacheRedisTypeHash Redis Hash structure (HSET/HGET).
	CacheRedisTypeHash
	// CacheRedisTypeSet Redis Set structure (SADD/SISMEMBER).
	CacheRedisTypeSet
)

var cacheTypeNames = []string{
	CacheRedisTypeKeyValue: "KV",
	CacheRedisTypeHash:     "Hash",
	CacheRedisTypeSet:      "Set",
}

// String returns the readable name of the cache type.
func (c CacheRedisType) String() string {
	if uint(c) < uint(len(cacheTypeNames)) {
		return cacheTypeNames[uint(c)]
	}
	return "Type:" + strconv.Itoa(int(c))
}

// CacheRedis is the primary Redis implementation of the Cacher interface.
type CacheRedis struct {
	*stores.Cache

	key     string
	keyFunc func() string
	// field acts as the Hash Key or Set Member.
	field  string
	tp     CacheRedisType
	groups []string // List of groups this cache belongs to for mass invalidation.

	modelName string
	// client uses atomic.Pointer to support thread-safe hot-swapping
	// of Redis connections during runtime configuration updates.
	client  atomic.Pointer[redis.Client]
	options *CacheRedisOptions
}

// CacheRedisOptions defines extra behaviors like L1 local caching.
type CacheRedisOptions struct {
	localCache stores.Cacher
}

// CacheRedisConfig holds the configuration for the Redis provider.
type CacheRedisConfig struct {
	stores.CacheConfig
	Client string `json:"client"` // Name of the redis client instance.
}

type CacheRedisOption func(options *CacheRedisOptions)

var (
	// Compile-time interface verification.
	_                       stores.Cacher = &CacheRedis{}
	defaultCacheRedisConfig               = CacheRedisConfig{
		CacheConfig: stores.DefaultCacheConfig,
		Client:      xredis.DefaultClientName,
	}
)

// NewRedisKeyValueCache helper to create a String-based Redis cache.
func NewRedisKeyValueCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeKeyValue), nil
}

// NewRedisHashCache helper to create a Hash-based Redis cache.
func NewRedisHashCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeHash), nil
}

// NewRedisSetCache helper to create a Set-based Redis cache.
func NewRedisSetCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeSet), nil
}

// WithLocalCache enables L1 (memory) caching before hitting L2 (Redis).
func WithLocalCache(cache stores.Cacher) CacheRedisOption {
	return func(options *CacheRedisOptions) {
		options.localCache = cache
	}
}

// NewRedisCache core constructor with functional options and dynamic config watching.
func NewRedisCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	cacheOptions := &CacheRedisOptions{}
	for _, opt := range options {
		opt(cacheOptions)
	}
	cache := &CacheRedis{
		Cache:     stores.NewCache(model),
		modelName: model.ModelName(),
		options:   cacheOptions,
	}
	return cache.loadAndWatch()
}

// WithGroup adds the current cache to a logical group.
// When the group is deleted, all member keys are also purged.
func (c *CacheRedis) WithGroup(group string) *CacheRedis {
	cr := &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: c.keyFunc,
		field:   c.field,
		tp:      c.tp,
		groups:  append(c.groups, c.Group(group)),
		options: c.options,
	}
	cr.client.Store(c.client.Load())
	return cr
}

// WithKey clones the cache handler with a specific key.
func (c *CacheRedis) WithKey(key string) *CacheRedis {
	cr := &CacheRedis{
		Cache:   c.Cache,
		key:     c.NewKey(key),
		keyFunc: c.keyFunc,
		field:   c.field,
		tp:      c.tp,
		groups:  c.groups,
		options: c.options,
	}
	cr.client.Store(c.client.Load())
	return cr
}

// WithKeyFunc clones the cache handler with a dynamic key generator.
func (c *CacheRedis) WithKeyFunc(keyFunc func() string) *CacheRedis {
	cr := &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: keyFunc,
		field:   c.field,
		tp:      c.tp,
		groups:  c.groups,
		options: c.options,
	}
	cr.client.Store(c.client.Load())
	return cr
}

// WithField clones the cache handler for a specific Hash field or Set member.
func (c *CacheRedis) WithField(field string) *CacheRedis {
	cr := &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: c.keyFunc,
		field:   field,
		tp:      c.tp,
		groups:  c.groups,
		options: c.options,
	}
	cr.client.Store(c.client.Load())
	return cr
}

// WithType clones the cache handler with a different storage strategy.
func (c *CacheRedis) WithType(tp CacheRedisType) *CacheRedis {
	cr := &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: c.keyFunc,
		field:   c.field,
		tp:      tp,
		groups:  c.groups,
		options: c.options,
	}
	cr.client.Store(c.client.Load())
	return cr
}

// Get attempts to find data in L1 local cache first, then fails over to Redis.
func (c *CacheRedis) Get(ctx context.Context, key string, out any) (bool, error) {
	if key == "" {
		return true, nil
	}
	client := c.client.Load()
	switch c.tp {
	case CacheRedisTypeKeyValue:
		// Attempt L1 read.
		if c.options.localCache != nil && c.options.localCache.Enabled() {
			if _, err := c.options.localCache.Get(ctx, key, out); err == nil {
				return false, nil
			} else {
				logger.L(ctx).Debug("redis cache read data from local fail", "key", key, "err", err)
			}
		}
		// Fallback to L2 (Redis).
		result := client.Get(ctx, key)
		if result.Err() != nil {
			return true, result.Err()
		}
		return true, json.Unmarshal([]byte(result.Val()), &out)
	case CacheRedisTypeHash:
		result := client.HGet(ctx, key, c.field)
		if result.Err() != nil {
			return true, result.Err()
		}
		return true, json.Unmarshal([]byte(result.Val()), out)
	case CacheRedisTypeSet:
		result := client.SIsMember(ctx, key, c.field)
		if result.Err() == nil {
			return true, result.Err()
		}
		return true, json.Unmarshal([]byte(result.String()), out)
	default:
		return true, fmt.Errorf("unimplement cache type %d", c.tp)
	}
}

// Del invalidates keys in Redis and Local cache, then cleans up group indexes.
func (c *CacheRedis) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	client := c.client.Load()
	switch c.tp {
	case CacheRedisTypeKeyValue:
		if err := c.delKeys(ctx, client, keys...); err != nil {
			return err
		}
	case CacheRedisTypeHash:
		if c.field != "" {
			for _, key := range keys {
				if err := client.HDel(ctx, key, c.field).Err(); err != nil {
					return err
				}
			}
		}
	case CacheRedisTypeSet:
		if c.field != "" {
			for _, key := range keys {
				if err := client.SRem(ctx, key, c.field).Err(); err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("unimplement cache type %d", c.tp)
	}
	return c.delGroup(ctx)
}

// Set stores data in Redis and Local cache.
// Note: Local cache TTL is typically halved (expiresIn/2) to minimize consistency issues.
func (c *CacheRedis) Set(ctx context.Context, key string, in any, expiresIn time.Duration) error {
	if key == "" {
		return nil
	}
	client := c.client.Load()
	switch c.tp {
	case CacheRedisTypeKeyValue:
		if c.options.localCache != nil && c.options.localCache.Enabled() {
			if err := c.options.localCache.Set(ctx, key, in, expiresIn/2); err != nil {
				logger.L(ctx).Error("redis cache set local cache fail", "key", key, "err", err)
			}
		}
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		if err := client.Set(ctx, key, string(b), expiresIn).Err(); err != nil {
			return err
		}
	case CacheRedisTypeHash:
		if c.field == "" {
			break
		}
		if err := client.HSet(ctx, key, map[string]any{
			c.field: in,
		}).Err(); err != nil {
			return err
		}
	case CacheRedisTypeSet:
		if c.field == "" {
			break
		}
		if err := client.SAdd(ctx, key, c.field).Err(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unimplement cache type %d", c.tp)
	}
	return c.addGroup(ctx, key)
}

// Refresh extends the TTL for both cache layers.
func (c *CacheRedis) Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) (err error) {
	if key == "" {
		return nil
	}
	client := c.client.Load()
	switch c.tp {
	case CacheRedisTypeKeyValue:
		if c.options.localCache != nil && c.options.localCache.Enabled() {
			if err := c.options.localCache.Set(ctx, key, in, expiresIn); err != nil {
				logger.L(ctx).Error("redis cache refresh local cache fail", "err", err)
			}
		}
		err = client.Expire(ctx, key, expiresIn).Err()
	case CacheRedisTypeHash:
		if c.field != "" {
			err = client.HExpire(ctx, key, expiresIn, c.field).Err()
		}
	case CacheRedisTypeSet:
	default:
		err = fmt.Errorf("unimplement cache type %d", c.tp)
	}
	return
}

// Key calculates the final formatted cache key.
func (c *CacheRedis) Key() string {
	if c.keyFunc != nil {
		return c.NewKey(c.keyFunc())
	}
	return c.key
}

// Group generates a globally unique key for the group index.
func (c *CacheRedis) Group(group string) string {
	return c.App().ResourceKey("caches_group",
		c.ModelKey(group),
		runtime.WithDelimiter(":"))
}

// Close gracefully shuts down the Redis connection.
func (c *CacheRedis) Close() {
	if client := c.client.Load(); client != nil {
		client.Close()
	}
}

// Enabed checks if caching is currently active and redis was connect.
func (c *CacheRedis) Enabled() bool {
	return c.Cache.Enabled() && c.client.Load() != nil
}

// addGroup links a specific key to one or more groups for bulk management.
func (c *CacheRedis) addGroup(ctx context.Context, key string) error {
	client := c.client.Load()
	if len(c.groups) != 0 {
		for _, group := range c.groups {
			logger.L(ctx).Debug("add group", "group", group, "key", key)
			// Store key-type mapping in the group index (a Redis Hash).
			if err := client.HSet(ctx, group, key, c.tp.String()).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

// delGroup finds all keys associated with a group and purges them from all layers.
func (c *CacheRedis) delGroup(ctx context.Context) error {
	client := c.client.Load()
	if len(c.groups) != 0 {
		for _, group := range c.groups {
			result := client.HGetAll(ctx, group)
			if err := result.Err(); err != nil {
				return err
			}
			if len(result.Val()) == 0 {
				continue
			}
			keys := make([]string, 0, len(result.Val()))
			for key := range result.Val() {
				keys = append(keys, key)
			}

			logger.L(ctx).Debug("delete group", "group", group, "keys", keys)
			if err := c.delKeys(ctx, client, keys...); err != nil {
				return err
			}
			if err := tools.DefaultTW.AddTask(time.Second, func() {
				if err := c.delKeys(ctx, client, keys...); err != nil {
					logger.L(ctx).Error("delay delete group keys fail", "group", group)
				}
			}); err != nil {
				return err
			}
		}
		// Clear the group index itself.
		if err := client.Del(ctx, c.groups...).Err(); err != nil {
			return err
		}
	}
	return nil
}

func (c *CacheRedis) delKeys(ctx context.Context, client *redis.Client, keys ...string) error {
	// Purge from both Local and Redis layers.
	if c.options.localCache != nil && c.options.localCache.Enabled() {
		if err := c.options.localCache.Del(ctx, keys...); err != nil {
			return err
		}
	}
	logger.L(ctx).Debug("delete cache from redis", "keys", keys)
	if err := client.Del(ctx, keys...).Err(); err != nil {
		return err
	}
	return nil
}

// loadAndWatch handles initial setup and dynamic configuration reloads.
func (c *CacheRedis) loadAndWatch() (*CacheRedis, error) {
	if err := c.load(); err != nil {
		return nil, err
	}
	config.AddPatternListener("asjard.cache.redis.*", c.watch)
	return c, nil
}

// load parses configuration with a hierarchy: Model-specific -> Redis-specific -> General.
func (c *CacheRedis) load() error {
	conf := defaultCacheRedisConfig
	if err := config.GetWithUnmarshal("asjard.cache",
		&conf,
		config.WithChain([]string{
			fmt.Sprintf("asjard.cache.models.%s", c.modelName),
			"asjard.cache.redis",
			fmt.Sprintf("asjard.cache.redis.models.%s", c.modelName),
		})); err != nil {
		logger.Error("redis cache load config fail", "err", err)
		return err
	}
	logger.Debug("load redis cache", "conf", conf)
	c.Cache.WithConf(&conf.CacheConfig)
	if conf.Enabled {
		client, err := xredis.NewClient(xredis.WithClientName(conf.Client))
		if err != nil {
			logger.Error("redis cache get redis client fail", "err", err)
			return err
		}
		// Safely update the connection pointer.
		c.client.Store(client)
	}
	return nil
}

func (c *CacheRedis) watch(event *config.Event) {
	if err := c.load(); err != nil {
		logger.Error("redis cache watch config fail", "err", err)
	}
}
