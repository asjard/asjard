package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/stores"
	"github.com/asjard/asjard/pkg/stores/xredis"
	"github.com/asjard/asjard/utils"
	"github.com/coocood/freecache"
	"github.com/redis/go-redis/v9"
)

// CacheLocal implements a thread-safe local in-memory cache.
type CacheLocal struct {
	*stores.Cache

	key        string
	keyFunc    func() string
	instanceId string // Unique ID of this application instance

	modelName string
	redis     *redis.Client
	// pubsub handles cross-instance cache invalidation via Redis.
	pubsub  *redis.PubSub
	cache   *freecache.Cache // Underlying high-performance memory cache
	maxSize int
}

// CacheLocalConfig defines settings for the local cache.
type CacheLocalConfig struct {
	stores.CacheConfig
	// RedisClient name used for synchronization.
	RedisClient string `json:"redisClient"`
	// MaxSize is the maximum memory allocated for the local cache.
	MaxSize int `json:"maxSize"`
}

var (
	// Compile-time check to ensure CacheLocal implements the Cacher interface.
	_                       stores.Cacher = &CacheLocal{}
	defaultCacheLocalConfig               = CacheLocalConfig{
		CacheConfig: stores.DefaultCacheConfig,
		MaxSize:     100 * 1024 * 1024, // Default 100MB
	}
)

// NewLocalCache initializes the local cache for a specific data model.
func NewLocalCache(model stores.Modeler) (*CacheLocal, error) {
	cache := &CacheLocal{
		Cache:      stores.NewCache(model),
		modelName:  model.ModelName(),
		instanceId: runtime.GetAPP().Instance.ID,
	}
	return cache.loadAndWatch()
}

// WithKey creates a copy of the cache handler for a specific key.
func (c *CacheLocal) WithKey(key string) *CacheLocal {
	return &CacheLocal{
		Cache:      c.Cache,
		key:        c.NewKey(key),
		keyFunc:    c.keyFunc,
		redis:      c.redis,
		pubsub:     c.pubsub,
		cache:      c.cache,
		instanceId: c.instanceId,
	}
}

// WithKeyFunc creates a copy using a dynamic key generator function.
func (c *CacheLocal) WithKeyFunc(keyFunc func() string) *CacheLocal {
	return &CacheLocal{
		Cache:      c.Cache,
		key:        c.key,
		keyFunc:    keyFunc,
		redis:      c.redis,
		pubsub:     c.pubsub,
		cache:      c.cache,
		instanceId: c.instanceId,
	}
}

// Get retrieves an item from the local memory.
func (c *CacheLocal) Get(ctx context.Context, key string, out any) (bool, error) {
	if key == "" {
		return true, nil
	}
	// Fetch bytes from freecache.
	value, err := c.cache.Get(utils.UnsafeString2Byte(key))
	if err != nil {
		return true, err // Key not found (or actual error)
	}
	logger.Debug("get value from local cache", "key", key, "value", value)
	return true, json.Unmarshal(value, out)
}

// Del removes keys locally AND informs other instances to do the same.
func (c *CacheLocal) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	// 1. Delete locally first.
	if err := c.del(keys...); err != nil {
		return err
	}
	// 2. Broadcast deletion to other instances.
	return c.delPublish(ctx, keys...)
}

func (c *CacheLocal) del(keys ...string) error {
	for _, key := range keys {
		c.cache.Del(utils.UnsafeString2Byte(key))
	}
	return nil
}

// Set stores data in the local memory with an expiration time.
func (c *CacheLocal) Set(ctx context.Context, key string, in any, expiresIn time.Duration) error {
	if key == "" {
		return nil
	}
	value, err := json.Marshal(in)
	if err != nil {
		return err
	}
	logger.Debug("set local", "key", key)
	return c.cache.Set(utils.UnsafeString2Byte(key), value, int(expiresIn.Seconds()))
}

// Refresh updates the TTL (Time to Live) for a local cache entry.
func (c *CacheLocal) Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) error {
	return c.cache.Touch(utils.UnsafeString2Byte(key), int(expiresIn.Seconds()))
}

// Key resolves the current cache key.
func (c *CacheLocal) Key() string {
	if c.keyFunc != nil {
		return c.NewKey(c.keyFunc())
	}
	return c.key
}

// loadAndWatch initializes configuration and listens for dynamic config updates.
func (c *CacheLocal) loadAndWatch() (*CacheLocal, error) {
	if err := c.load(); err != nil {
		logger.Error("local cache load config fail", "err", err)
		return nil, err
	}
	// Watch for runtime configuration changes (e.g., resizing cache).
	config.AddPatternListener("asjard.cache.local.*", c.watch)
	return c, nil
}

// cacheLocalDelPublishMessage is the DTO for the Redis invalidation channel.
type cacheLocalDelPublishMessage struct {
	InstanceId string   // Who triggered the delete
	Keys       []string // Which keys to invalidate
}

// delPublish sends the invalidation message to the Redis channel.
func (c *CacheLocal) delPublish(ctx context.Context, keys ...string) error {
	if c.redis == nil {
		return nil
	}
	msg := &cacheLocalDelPublishMessage{
		InstanceId: c.instanceId,
		Keys:       keys,
	}
	v, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	logger.Debug("local cache del publish", "msg", string(v))
	return c.redis.Publish(ctx, c.delChannel(), string(v)).Err()
}

// delSubscribe listens for invalidation messages from other instances.
func (c *CacheLocal) delSubscribe() {
	if c.redis == nil {
		return
	}
	if c.pubsub != nil {
		c.pubsub.Close()
	}

	logger.Debug("local cache del subscribe")
	c.pubsub = c.redis.Subscribe(context.Background(), c.delChannel())

	for {
		select {
		case <-runtime.Exit: // Framework graceful shutdown
			logger.Debug("local cache del subscribe exit")
			c.pubsub.Close()
			return
		case msg := <-c.pubsub.Channel():
			var delMsg cacheLocalDelPublishMessage
			if err := json.Unmarshal(utils.UnsafeString2Byte(msg.Payload), &delMsg); err != nil {
				logger.Error("local cache pubsub unmarshal fail", "payload", msg.Payload, "err", err)
				continue
			}
			// Only delete if the message came from a DIFFERENT instance.
			if delMsg.InstanceId != c.instanceId {
				if err := c.del(delMsg.Keys...); err != nil {
					logger.Error("local cache del keys fail", "keys", delMsg.Keys, "err", err)
				}
			}
		}
	}
}

// delChannel generates a unique Redis channel name for invalidation events.
func (c *CacheLocal) delChannel() string {
	return c.App().ResourceKey("caches_local_channel",
		"delete",
		runtime.WithDelimiter(":"))
}

// load merges multiple configuration levels (Global -> Local -> Model-specific).
func (c *CacheLocal) load() error {
	conf := defaultCacheLocalConfig
	if err := config.GetWithUnmarshal("asjard.cache",
		&conf,
		config.WithChain([]string{
			fmt.Sprintf("asjard.cache.models.%s", c.modelName),
			"asjard.cache.local",
			fmt.Sprintf("asjard.cache.local.models.%s", c.modelName),
		})); err != nil {
		return err
	}

	c.Cache.WithConf(&conf.CacheConfig)

	// Initialize Redis client if a name is provided.
	if conf.RedisClient != "" {
		client, err := xredis.Client(xredis.WithClientName(conf.RedisClient))
		if err != nil {
			return err
		}
		c.redis = client
		go c.delSubscribe() // Start listening for remote invalidations.
	}

	// Handle dynamic cache resizing.
	if c.maxSize != conf.MaxSize {
		c.cache = freecache.NewCache(conf.MaxSize)
		c.maxSize = conf.MaxSize
	}
	return nil
}

func (c *CacheLocal) watch(event *config.Event) {
	if err := c.load(); err != nil {
		logger.Error("local cache watch config fail", "err", err)
	}
}
