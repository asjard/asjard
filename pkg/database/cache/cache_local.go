package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/database"
	mredis "github.com/asjard/asjard/pkg/database/redis"
	"github.com/coocood/freecache"
	"github.com/redis/go-redis/v9"
)

// CacheLocal 本地缓存
type CacheLocal struct {
	*database.Cache

	// 缓存Key
	key string
	// 本实例ID
	instanceId string

	modelName string
	// 广播redis客户端
	publishClient *redis.Client
	pubsub        *redis.PubSub
	cache         *freecache.Cache
	maxSize       int
}

// CacheLocalConfig 本地缓存配置
type CacheLocalConfig struct {
	database.CacheConfig
	// 广播redis客户端
	// 多个实例下，需要所有实例删除本地缓存
	PublishClient string `json:"publishClient"`
	// 最大使用内存
	MaxSize int `json:"maxSize"`
}

var (
	_                  database.Cacher = &CacheLocal{}
	defaultCacheConfig                 = CacheLocalConfig{
		CacheConfig: database.DefaultCacheConfig,
		MaxSize:     100 * 1024 * 1024,
	}
)

// NewLocalCache 本地缓存初始化
func NewLocalCache(model database.Modeler) (*CacheLocal, error) {
	cache := &CacheLocal{
		Cache:      database.NewCache(model),
		modelName:  model.ModelName(),
		instanceId: runtime.GetAPP().Instance.ID,
	}
	return cache.loadAndWatch()
}

// WithKey 设置缓存key
func (c *CacheLocal) WithKey(key string) *CacheLocal {
	return &CacheLocal{
		Cache:         c.Cache,
		key:           c.NewKey(key),
		publishClient: c.publishClient,
		pubsub:        c.pubsub,
		cache:         c.cache,
		instanceId:    c.instanceId,
	}
}

// 从缓存获取数据
func (c *CacheLocal) Get(ctx context.Context, key string, out any) error {
	value, err := c.cache.Get([]byte(key))
	if err != nil {
		return err
	}
	return json.Unmarshal(value, out)
}

// 从缓存删除数据
func (c *CacheLocal) Del(ctx context.Context, keys ...string) error {
	if err := c.del(keys...); err != nil {
		return err
	}
	return c.delPublish(ctx, keys...)
}

func (c *CacheLocal) del(keys ...string) error {
	for _, key := range keys {
		c.cache.Del([]byte(key))
	}
	return nil
}

// 设置缓存数据
func (c *CacheLocal) Set(ctx context.Context, key string, in any) error {
	value, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return c.cache.Set([]byte(key), value, int(c.ExpiresIn().Seconds()))
}

// 刷新缓存过期时间
// 不实现缓存刷新，强制自动过期
func (c *CacheLocal) Refresh(ctx context.Context, key string) error {
	return nil
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
	if c.publishClient == nil {
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
	return c.publishClient.Publish(ctx, c.delChannel(), string(v)).Err()
}

// 删除订阅
func (c *CacheLocal) delSubscribe() {
	if c.publishClient == nil {
		return
	}
	if c.pubsub != nil {
		c.pubsub.Close()
	}
	logger.Debug("local cache del subscribe")
	c.pubsub = c.publishClient.Subscribe(context.Background(), c.delChannel())
	for msg := range c.pubsub.Channel() {
		logger.Debug("local cache del subscribe recive", "msg", msg.Payload)
		var delMsg cacheLocalDelPublishMessage
		if err := json.Unmarshal([]byte(msg.Payload), &delMsg); err != nil {
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
	if conf.PublishClient != "" {
		client, err := mredis.Client(mredis.WithClientName(conf.PublishClient))
		if err != nil {
			return err
		}
		c.publishClient = client
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
