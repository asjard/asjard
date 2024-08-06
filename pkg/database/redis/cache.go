package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	"github.com/asjard/asjard/pkg/database"
	"github.com/redis/go-redis/v9"
)

// CacheType 缓存类型
type CacheType uint

const (
	// CacheTypeKeyValue key-value缓存
	CacheTypeKeyValue CacheType = iota
	// CacheTypeHash hash缓存
	CacheTypeHash
	// CacheTypeSet 集合缓存
	CacheTypeSet
)

var cacheTypeNames = []string{
	CacheTypeKeyValue: "KV",
	CacheTypeHash:     "Hash",
	CacheTypeSet:      "Set",
}

func (c CacheType) String() string {
	if uint(c) < uint(len(cacheTypeNames)) {
		return cacheTypeNames[uint(c)]
	}
	return "Type:" + strconv.Itoa(int(c))
}

// RedisCache redis缓存
type Cache struct {
	*database.Cache

	// 缓存key
	key string
	// hash中的field， set中的member
	field string
	// 缓存类型
	tp     CacheType
	groups []string

	client  *redis.Client
	options *CacheOptions
}

type CacheOptions struct {
	localCache database.Cacher
}

type CacheConfig struct {
	database.CacheConfig
	Client string `json:"client"`
}

type CacheOption func(options *CacheOptions)

var (
	_                  database.Cacher = &Cache{}
	defaultCacheConfig                 = CacheConfig{
		CacheConfig: database.DefaultCacheConfig,
		Client:      DefaultClientName,
	}
)

// NewKeyValueCache key/value缓存初始化
func NewKeyValueCache(model database.Modeler, options ...CacheOption) (*Cache, error) {
	newCache, err := NewCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheTypeKeyValue), nil
}

// NewHashCache hash缓存
func NewHashCache(model database.Modeler, options ...CacheOption) (*Cache, error) {
	newCache, err := NewCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheTypeHash), nil
}

// NewSetCache set缓存
func NewSetCache(model database.Modeler, options ...CacheOption) (*Cache, error) {
	newCache, err := NewCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheTypeSet), nil
}

// WithLocalCache 设置本地缓存
func WithLocalCache(cache database.Cacher) CacheOption {
	return func(options *CacheOptions) {
		options.localCache = cache
	}
}

// NewCache 缓存初始化
// TODO 配置监听
func NewCache(model database.Modeler, options ...CacheOption) (*Cache, error) {
	conf := defaultCacheConfig
	if err := config.GetWithUnmarshal("asjard.cache",
		&conf,
		config.WithChain([]string{
			fmt.Sprintf("asjard.cache.models.%s", model.ModelName()),
			"asjard.cache.redis",
			fmt.Sprintf("asjard.cache.redis.models.%s", model.ModelName()),
		})); err != nil {
		return nil, err
	}
	client, err := Client(WithClientName(conf.Client))
	if err != nil {
		return nil, err
	}
	cacheOptions := &CacheOptions{}
	for _, opt := range options {
		opt(cacheOptions)
	}
	return &Cache{
		Cache:   database.NewCache(model).WithConf(&conf.CacheConfig),
		client:  client,
		options: cacheOptions,
	}, nil
}

// WithGroup 分组
func (c *Cache) WithGroup(group string) *Cache {
	return &Cache{
		Cache:   c.Cache,
		key:     c.key,
		field:   c.field,
		tp:      c.tp,
		groups:  append(c.groups, c.Group(group)),
		client:  c.client,
		options: c.options,
	}
}

func (c *Cache) WithKey(key string) *Cache {
	return &Cache{
		Cache:   c.Cache,
		key:     c.NewKey(key),
		field:   c.field,
		tp:      c.tp,
		groups:  c.groups,
		client:  c.client,
		options: c.options,
	}
}

func (c *Cache) WithField(field string) *Cache {
	return &Cache{
		Cache:   c.Cache,
		key:     c.key,
		field:   field,
		tp:      c.tp,
		groups:  c.groups,
		client:  c.client,
		options: c.options,
	}
}

func (c *Cache) WithType(tp CacheType) *Cache {
	return &Cache{
		Cache:   c.Cache,
		key:     c.key,
		field:   c.field,
		tp:      tp,
		groups:  c.groups,
		client:  c.client,
		options: c.options,
	}
}

func (c Cache) Get(ctx context.Context, key string, out any) error {
	if key == "" {
		return nil
	}
	switch c.tp {
	case CacheTypeKeyValue:
		// 先从本地缓存获取，如果获取到则直接返回
		if c.options.localCache != nil {
			if err := c.options.localCache.Get(ctx, key, out); err == nil {
				return nil
			}
		}
		result := c.client.Get(ctx, key)
		if result.Err() != nil {
			return result.Err()
		}
		return json.Unmarshal([]byte(result.Val()), &out)
	case CacheTypeHash:
		result := c.client.HGet(ctx, key, c.field)
		if result.Err() != nil {
			return result.Err()
		}
		return json.Unmarshal([]byte(result.Val()), out)
	case CacheTypeSet:
		result := c.client.SIsMember(ctx, key, c.field)
		if result.Err() == nil {
			return result.Err()
		}
		return json.Unmarshal([]byte(result.String()), out)
	default:
		return fmt.Errorf("unimplement cache type %d", c.tp)
	}
}

func (c Cache) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	switch c.tp {
	case CacheTypeKeyValue:
		if c.options.localCache != nil {
			if err := c.options.localCache.Del(ctx, keys...); err != nil {
				return err
			}
		}
		if err := c.client.Del(ctx, keys...).Err(); err != nil {
			return err
		}
	case CacheTypeHash:
		if c.field != "" {
			for _, key := range keys {
				if err := c.client.HDel(ctx, key, c.field).Err(); err != nil {
					return err
				}
			}
		}
	case CacheTypeSet:
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

func (c Cache) Set(ctx context.Context, key string, in any) error {
	if key == "" {
		return nil
	}
	switch c.tp {
	case CacheTypeKeyValue:
		if c.options.localCache != nil {
			c.options.localCache.Set(ctx, key, in)
		}
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		if err := c.client.Set(ctx, key, string(b), c.ExpiresIn()).Err(); err != nil {
			return err
		}
	case CacheTypeHash:
		if c.field == "" {
			break
		}
		if err := c.client.HSet(ctx, key, map[string]any{
			c.field: in,
		}).Err(); err != nil {
			return err
		}
	case CacheTypeSet:
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

func (c Cache) Refresh(ctx context.Context, key string) (err error) {
	if key == "" {
		return nil
	}
	switch c.tp {
	case CacheTypeKeyValue:
		err = c.client.Expire(ctx, key, c.ExpiresIn()).Err()
	case CacheTypeHash:
		if c.field != "" {
			err = c.client.HExpire(ctx, key, c.ExpiresIn(), c.field).Err()
		}
	case CacheTypeSet:
	default:
		err = fmt.Errorf("unimplement cache type %d", c.tp)
	}
	return
}

func (c Cache) Key() string {
	return c.key
}

// 分组
func (c Cache) Group(group string) string {
	return c.Prefix() + ":groups:" + group
}

// 添加分组
func (c Cache) addGroup(ctx context.Context, key string) error {
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
func (c Cache) delGroup(ctx context.Context) error {
	// 删除分组内的所有缓存
	if len(c.groups) != 0 {
		for _, group := range c.groups {
			result := c.client.HGetAll(ctx, c.Group(group))
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
			if c.options.localCache != nil {
				if err := c.options.localCache.Del(ctx, keys...); err != nil {
					return err
				}
			}
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
	}
	return nil
}
