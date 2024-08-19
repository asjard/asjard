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

// CacheLocal 本地缓存
type CacheLocal struct {
	*stores.Cache

	// 缓存Key
	key string
	// 本实例ID
	instanceId string

	modelName string
	// redis客户端
	redis   *redis.Client
	pubsub  *redis.PubSub
	cache   *freecache.Cache
	maxSize int
}

// CacheLocalConfig 本地缓存配置
type CacheLocalConfig struct {
	stores.CacheConfig
	// redis客户端
	RedisClient string `json:"redisClient"`
	// 最大使用内存
	MaxSize int `json:"maxSize"`
}

var (
	_                  stores.Cacher = &CacheLocal{}
	defaultCacheConfig               = CacheLocalConfig{
		CacheConfig: stores.DefaultCacheConfig,
		MaxSize:     100 * 1024 * 1024,
	}
)

// NewLocalCache 本地缓存初始化
func NewLocalCache(model stores.Modeler) (*CacheLocal, error) {
	cache := &CacheLocal{
		Cache:      stores.NewCache(model),
		modelName:  model.ModelName(),
		instanceId: runtime.GetAPP().Instance.ID,
	}
	return cache.loadAndWatch()
}

// WithKey 设置缓存key
func (c *CacheLocal) WithKey(key string) *CacheLocal {
	return &CacheLocal{
		Cache:      c.Cache,
		key:        c.NewKey(key),
		redis:      c.redis,
		pubsub:     c.pubsub,
		cache:      c.cache,
		instanceId: c.instanceId,
	}
}

// 从缓存获取数据
func (c *CacheLocal) Get(ctx context.Context, key string, out any) (bool, error) {
	if key == "" {
		return true, nil
	}
	value, err := c.cache.Get(utils.String2Byte(key))
	if err != nil {
		return true, err
	}
	logger.Debug("get value from local cache", "key", key, "value", value)
	return true, json.Unmarshal(value, out)
}

// 从缓存删除数据
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
		c.cache.Del(utils.String2Byte(key))
	}
	return nil
}

// 设置缓存数据
func (c *CacheLocal) Set(ctx context.Context, key string, in any, expiresIn time.Duration) error {
	if key == "" {
		return nil
	}
	value, err := json.Marshal(in)
	if err != nil {
		return err
	}
	logger.Debug("set local", "key", key)
	return c.cache.Set(utils.String2Byte(key), value, int(expiresIn.Seconds()))
}

// 刷新缓存过期时间
func (c *CacheLocal) Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) error {
	return c.cache.Touch(utils.String2Byte(key), int(expiresIn.Seconds()))
}

// 返回缓存Key名称
func (c *CacheLocal) Key() string {
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

// 本地缓存广播消息
type cacheLocalDelPublishMessage struct {
	// 实例ID
	InstanceId string
	// key列表
	Keys []string
}

// 删除发布
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

// 删除订阅
func (c *CacheLocal) delSubscribe() {
	if c.redis == nil {
		return
	}
	if c.pubsub != nil {
		c.pubsub.Close()
	}
	logger.Debug("local cache del subscribe")
	c.pubsub = c.redis.Subscribe(context.Background(), c.delChannel())
	for msg := range c.pubsub.Channel() {
		logger.Debug("local cache del subscribe recive", "msg", msg.Payload)
		var delMsg cacheLocalDelPublishMessage
		if err := json.Unmarshal(utils.String2Byte(msg.Payload), &delMsg); err != nil {
			logger.Error("local cache pubsub unmarshal fail", "payload", msg.Payload, "err", err)
		}
		if delMsg.InstanceId != c.instanceId {
			if err := c.del(delMsg.Keys...); err != nil {
				logger.Error("local cache del keys fail", "keys", delMsg.Keys, "err", err)
			}
		}
	}
}

func (c *CacheLocal) delChannel() string {
	return c.Prefix() + ":channels:delete"
}

func (c *CacheLocal) load() error {
	conf := defaultCacheConfig
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
