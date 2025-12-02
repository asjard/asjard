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

// CacheLocal implement local cache
type CacheLocal struct {
	*stores.Cache

	key        string
	keyFunc    func() string
	instanceId string

	modelName string
	redis     *redis.Client
	// publish and subcribe other instance delete cache event
	pubsub  *redis.PubSub
	cache   *freecache.Cache
	maxSize int
}

// CacheLocalConfig define local cache config
type CacheLocalConfig struct {
	stores.CacheConfig
	// redis client name
	RedisClient string `json:"redisClient"`
	// max memory used by local cache
	MaxSize int `json:"maxSize"`
}

var (
	_                       stores.Cacher = &CacheLocal{}
	defaultCacheLocalConfig               = CacheLocalConfig{
		CacheConfig: stores.DefaultCacheConfig,
		MaxSize:     100 * 1024 * 1024,
	}
)

// NewLocalCache create local cache
func NewLocalCache(model stores.Modeler) (*CacheLocal, error) {
	cache := &CacheLocal{
		Cache:      stores.NewCache(model),
		modelName:  model.ModelName(),
		instanceId: runtime.GetAPP().Instance.ID,
	}
	return cache.loadAndWatch()
}

// WithKey set cache key.
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

// WithKeyFunc set cache key use function.
// if keyFunc was settled, it will be first to use.
// it is only called when used.
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

// Get cache from local and set into out params.
func (c *CacheLocal) Get(ctx context.Context, key string, out any) (bool, error) {
	if key == "" {
		return true, nil
	}
	value, err := c.cache.Get(utils.UnsafeString2Byte(key))
	if err != nil {
		return true, err
	}
	logger.Debug("get value from local cache", "key", key, "value", value)
	return true, json.Unmarshal(value, out)
}

// Del delete cache from local and publish delete event to other instance if redis.client was setted.
func (c *CacheLocal) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := c.del(keys...); err != nil {
		return err
	}
	return c.delPublish(ctx, keys...)
}

func (c *CacheLocal) del(keys ...string) error {
	for _, key := range keys {
		c.cache.Del(utils.UnsafeString2Byte(key))
	}
	return nil
}

// Set data into local cache.
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

// Refresh local cache expire time.
func (c *CacheLocal) Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) error {
	return c.cache.Touch(utils.UnsafeString2Byte(key), int(expiresIn.Seconds()))
}

// Key return local cache key.
func (c *CacheLocal) Key() string {
	if c.keyFunc != nil {
		return c.NewKey(c.keyFunc())
	}
	return c.key
}

func (c *CacheLocal) loadAndWatch() (*CacheLocal, error) {
	if err := c.load(); err != nil {
		logger.Error("local cache load config fail", "err", err)
		return nil, err
	}
	config.AddPatternListener("asjard.cache.local.*", c.watch)
	return c, nil
}

type cacheLocalDelPublishMessage struct {
	// 实例ID
	InstanceId string
	// key列表
	Keys []string
}

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
		case <-runtime.Exit:
			logger.Debug("local cache del subscribe exit")
			c.pubsub.Close()
			return
		case msg := <-c.pubsub.Channel():
			logger.Debug("local cache del subscribe recive", "msg", msg.Payload)
			var delMsg cacheLocalDelPublishMessage
			if err := json.Unmarshal(utils.UnsafeString2Byte(msg.Payload), &delMsg); err != nil {
				logger.Error("local cache pubsub unmarshal fail", "payload", msg.Payload, "err", err)
			}
			if delMsg.InstanceId != c.instanceId {
				if err := c.del(delMsg.Keys...); err != nil {
					logger.Error("local cache del keys fail", "keys", delMsg.Keys, "err", err)
				}
			}
		}
	}
}

func (c *CacheLocal) delChannel() string {
	return c.App().ResourceKey("caches_local_channel",
		"delete",
		runtime.WithDelimiter(":"))
}

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
	logger.Debug("load local cache", "conf", conf)
	c.Cache.WithConf(&conf.CacheConfig)
	if conf.RedisClient != "" {
		client, err := xredis.Client(xredis.WithClientName(conf.RedisClient))
		if err != nil {
			return err
		}
		c.redis = client
		go c.delSubscribe()
	}
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
