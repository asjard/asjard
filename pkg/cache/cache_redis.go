package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/core/runtime"
	"github.com/asjard/asjard/pkg/stores"
	"github.com/asjard/asjard/pkg/stores/xredis"
	"github.com/redis/go-redis/v9"
)

// CacheType 缓存类型
type CacheRedisType uint

const (
	// CacheRedisTypeKeyValue key-value缓存
	CacheRedisTypeKeyValue CacheRedisType = iota
	// CacheRedisTypeHash hash缓存
	CacheRedisTypeHash
	// CacheRedisTypeSet 集合缓存
	CacheRedisTypeSet
)

var cacheTypeNames = []string{
	CacheRedisTypeKeyValue: "KV",
	CacheRedisTypeHash:     "Hash",
	CacheRedisTypeSet:      "Set",
}

func (c CacheRedisType) String() string {
	if uint(c) < uint(len(cacheTypeNames)) {
		return cacheTypeNames[uint(c)]
	}
	return "Type:" + strconv.Itoa(int(c))
}

// CacheRedis redis缓存
type CacheRedis struct {
	*stores.Cache

	// 缓存key
	key string
	// 延迟缓存key,优先使用，如果为nil则使用key
	keyFunc func() string
	// hash中的field， set中的member
	field string
	// 缓存类型
	tp     CacheRedisType
	groups []string

	modelName string
	client    *redis.Client
	options   *CacheRedisOptions
}

// CacheRedisOptions 初始化redis缓存的一些参数
type CacheRedisOptions struct {
	localCache stores.Cacher
}

// CacheRedisConfig 缓存配置
type CacheRedisConfig struct {
	stores.CacheConfig
	Client string `json:"client"`
}

type CacheRedisOption func(options *CacheRedisOptions)

var (
	_ stores.Cacher = &CacheRedis{}
	// 默认缓存配置
	defaultCacheRedisConfig = CacheRedisConfig{
		CacheConfig: stores.DefaultCacheConfig,
		Client:      xredis.DefaultClientName,
	}
)

// NewKeyValueCache key/value缓存初始化
func NewRedisKeyValueCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeKeyValue), nil
}

// NewHashCache hash缓存
func NewRedisHashCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeHash), nil
}

// NewSetCache set缓存
func NewRedisSetCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeSet), nil
}

// WithLocalCache 设置本地缓存
func WithLocalCache(cache stores.Cacher) CacheRedisOption {
	return func(options *CacheRedisOptions) {
		options.localCache = cache
	}
}

// NewCache 缓存初始化
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

// WithGroup 分组
func (c *CacheRedis) WithGroup(group string) *CacheRedis {
	return &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: c.keyFunc,
		field:   c.field,
		tp:      c.tp,
		groups:  append(c.groups, c.Group(group)),
		client:  c.client,
		options: c.options,
	}
}

// WithKey 设置缓存key
func (c *CacheRedis) WithKey(key string) *CacheRedis {
	return &CacheRedis{
		Cache:   c.Cache,
		key:     c.NewKey(key),
		keyFunc: c.keyFunc,
		field:   c.field,
		tp:      c.tp,
		groups:  c.groups,
		client:  c.client,
		options: c.options,
	}
}

// WithKeyFunc 延迟设置缓存key，在部分场景下缓存key有可能需要创建/更新完数据库后才能拿得到完整key
func (c *CacheRedis) WithKeyFunc(keyFunc func() string) *CacheRedis {
	return &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: keyFunc,
		field:   c.field,
		tp:      c.tp,
		groups:  c.groups,
		client:  c.client,
		options: c.options,
	}
}

// WithField hash, set设置field
func (c *CacheRedis) WithField(field string) *CacheRedis {
	return &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: c.keyFunc,
		field:   field,
		tp:      c.tp,
		groups:  c.groups,
		client:  c.client,
		options: c.options,
	}
}

// WithType 设置缓存类型
func (c *CacheRedis) WithType(tp CacheRedisType) *CacheRedis {
	return &CacheRedis{
		Cache:   c.Cache,
		key:     c.key,
		keyFunc: c.keyFunc,
		field:   c.field,
		tp:      tp,
		groups:  c.groups,
		client:  c.client,
		options: c.options,
	}
}

// Get 从缓存获取数据
// 如果设置了本地缓存先从本地缓存获取数据
// 获取不到再去redis获取数据
func (c CacheRedis) Get(ctx context.Context, key string, out any) (bool, error) {
	if key == "" {
		return true, nil
	}
	switch c.tp {
	case CacheRedisTypeKeyValue:
		// 先从本地缓存获取，如果获取到则直接返回
		if c.options.localCache != nil && c.options.localCache.Enabled() {
			if _, err := c.options.localCache.Get(ctx, key, out); err == nil {
				return false, nil
			} else {
				logger.Debug("redis cache read data from local fail", "key", key, "err", err)
			}
		}
		result := c.client.Get(ctx, key)
		if result.Err() != nil {
			return true, result.Err()
		}
		return true, json.Unmarshal([]byte(result.Val()), &out)
	case CacheRedisTypeHash:
		result := c.client.HGet(ctx, key, c.field)
		if result.Err() != nil {
			return true, result.Err()
		}
		return true, json.Unmarshal([]byte(result.Val()), out)
	case CacheRedisTypeSet:
		result := c.client.SIsMember(ctx, key, c.field)
		if result.Err() == nil {
			return true, result.Err()
		}
		return true, json.Unmarshal([]byte(result.String()), out)
	default:
		return true, fmt.Errorf("unimplement cache type %d", c.tp)
	}
}

// Del 删除缓存
// 如果设置了本地缓存先删除本地缓存再删除redis缓存
func (c CacheRedis) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	switch c.tp {
	case CacheRedisTypeKeyValue:
		if c.options.localCache != nil && c.options.localCache.Enabled() {
			if err := c.options.localCache.Del(ctx, keys...); err != nil {
				return err
			}
		}
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	case CacheRedisTypeHash:
		if c.field != "" {
			for _, key := range keys {
				if err := c.client.HDel(ctx, key, c.field).Err(); err != nil {
					return err
				}
			}
		}
	case CacheRedisTypeSet:
		if c.field != "" {
			for _, key := range keys {
				if err := c.client.SRem(ctx, key, c.field).Err(); err != nil {
					return err
				}
			}
		}
	default:
		return fmt.Errorf("unimplement cache type %d", c.tp)
	}
	return c.delGroup(ctx)
}

// Set 设置缓存
// 如果设置了本地缓存则同时设置本地缓存
func (c CacheRedis) Set(ctx context.Context, key string, in any, expiresIn time.Duration) error {
	if key == "" {
		return nil
	}
	switch c.tp {
	case CacheRedisTypeKeyValue:
		if c.options.localCache != nil && c.options.localCache.Enabled() {
			if err := c.options.localCache.Set(ctx, key, in, expiresIn/2); err != nil {
				logger.Error("redis cache set local cache fail", "key", key, "err", err)
			}
		}
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		if err := c.client.Set(ctx, key, string(b), expiresIn).Err(); err != nil {
			return err
		}
	case CacheRedisTypeHash:
		if c.field == "" {
			break
		}
		if err := c.client.HSet(ctx, key, map[string]any{
			c.field: in,
		}).Err(); err != nil {
			return err
		}
	case CacheRedisTypeSet:
		if c.field == "" {
			break
		}
		if err := c.client.SAdd(ctx, key, c.field).Err(); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unimplement cache type %d", c.tp)
	}
	return c.addGroup(ctx, key)
}

// Refresh 更新缓存时间
// 如果设置了本地缓存则更新本地缓存数据再刷新redis缓存过期时间
func (c CacheRedis) Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) (err error) {
	if key == "" {
		return nil
	}
	switch c.tp {
	case CacheRedisTypeKeyValue:
		if c.options.localCache != nil && c.options.localCache.Enabled() {
			if err := c.options.localCache.Set(ctx, key, in, expiresIn); err != nil {
				logger.Error("redis cache refresh local cache fail", "err", err)
			}
		}
		err = c.client.Expire(ctx, key, expiresIn).Err()
	case CacheRedisTypeHash:
		if c.field != "" {
			err = c.client.HExpire(ctx, key, expiresIn, c.field).Err()
		}
	case CacheRedisTypeSet:
	default:
		err = fmt.Errorf("unimplement cache type %d", c.tp)
	}
	return
}

func (c CacheRedis) Key() string {
	if c.keyFunc != nil {
		return c.NewKey(c.keyFunc())
	}
	return c.key
}

// Group 缓存分组
func (c CacheRedis) Group(group string) string {
	return c.App().ResourceKey("caches_group",
		c.ModelKey(group),
		runtime.WithDelimiter(":"))
}

func (c CacheRedis) Close() {
	c.client.Close()
}

// 添加分组
func (c CacheRedis) addGroup(ctx context.Context, key string) error {
	// 添加分组
	if len(c.groups) != 0 {
		for _, group := range c.groups {
			logger.Debug("add group", "group", group, "key", key)
			if err := c.client.HSet(ctx, group, key, c.tp.String()).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}

// 删除分组
func (c CacheRedis) delGroup(ctx context.Context) error {
	// 删除分组内的所有缓存
	if len(c.groups) != 0 {
		for _, group := range c.groups {
			result := c.client.HGetAll(ctx, group)
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
			logger.Debug("delete group", "group", group, "keys", keys)
			if c.options.localCache != nil && c.options.localCache.Enabled() {
				if err := c.options.localCache.Del(ctx, keys...); err != nil {
					return err
				}
			}
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if err := c.client.Del(ctx, c.groups...).Err(); err != nil {
			return err
		}
	}
	return nil
}

// 加载并监听配置变化
func (c *CacheRedis) loadAndWatch() (*CacheRedis, error) {
	if err := c.load(); err != nil {
		logger.Error("redis cache load config fail", "err", err)
		return nil, err
	}
	config.AddPatternListener("asjard.cache.redis.*", c.watch)
	return c, nil
}

func (c *CacheRedis) load() error {
	conf := defaultCacheRedisConfig
	if err := config.GetWithUnmarshal("asjard.cache",
		&conf,
		config.WithChain([]string{
			fmt.Sprintf("asjard.cache.models.%s", c.modelName),
			"asjard.cache.redis",
			fmt.Sprintf("asjard.cache.redis.models.%s", c.modelName),
		})); err != nil {
		return err
	}
	logger.Debug("load redis cache", "conf", conf)
	c.Cache.WithConf(&conf.CacheConfig)
	client, err := xredis.NewClient(xredis.WithClientName(conf.Client))
	if err != nil {
		return err
	}
	c.client = client
	return nil
}

func (c *CacheRedis) watch(event *config.Event) {
	if err := c.load(); err != nil {
		logger.Error("redis cache watch config fail", "err", err)
	}
}
