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

// CacheRedisType redis type that used in redis cache.
type CacheRedisType uint

const (
	// CacheRedisTypeKeyValue key-value cache type
	CacheRedisTypeKeyValue CacheRedisType = iota
	// CacheRedisTypeHash hash cache type
	CacheRedisTypeHash
	// CacheRedisTypeSet set cache type
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

// CacheRedis redis cache implement.
type CacheRedis struct {
	*stores.Cache

	key     string
	keyFunc func() string
	// field in hash, member in set
	field  string
	tp     CacheRedisType
	groups []string

	modelName string
	client    *redis.Client
	options   *CacheRedisOptions
}

// CacheRedisOptions .
type CacheRedisOptions struct {
	localCache stores.Cacher
}

// CacheRedisConfig redis cache config.
type CacheRedisConfig struct {
	stores.CacheConfig
	Client string `json:"client"`
}

type CacheRedisOption func(options *CacheRedisOptions)

var (
	_                       stores.Cacher = &CacheRedis{}
	defaultCacheRedisConfig               = CacheRedisConfig{
		CacheConfig: stores.DefaultCacheConfig,
		Client:      xredis.DefaultClientName,
	}
)

// NewKeyValueCache create a key-value redis cache.
func NewRedisKeyValueCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeKeyValue), nil
}

// NewHashCache create a hash redis cache.
func NewRedisHashCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeHash), nil
}

// NewSetCache create a set redis cache.
func NewRedisSetCache(model stores.Modeler, options ...CacheRedisOption) (*CacheRedis, error) {
	newCache, err := NewRedisCache(model, options...)
	if err != nil {
		return nil, err
	}
	return newCache.WithType(CacheRedisTypeSet), nil
}

// WithLocalCache set local cache.
func WithLocalCache(cache stores.Cacher) CacheRedisOption {
	return func(options *CacheRedisOptions) {
		options.localCache = cache
	}
}

// NewCache create redis cache with options.
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

// WithGroup set cache group
// it will delete group and all keys in group when delete group.
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

// WithKey set cache key.
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

// WithKeyFunc set cache key use function.
// if keyFunc was settled, it will be first to use.
// it is only called when used.
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

// WithField hash, set field in hash or set.
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

// WithType set cache type.
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

// Get data from cache
// first get data from local cache if setted local cache,
// if can not get data from local then get data from redis.
func (c CacheRedis) Get(ctx context.Context, key string, out any) (bool, error) {
	if key == "" {
		return true, nil
	}
	switch c.tp {
	case CacheRedisTypeKeyValue:
		// get data from local cache.
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

// Del delet cache
// delete cache from redis and local if setted local cache.
// delete all keys in group and group if group setted.
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

// Set data in cache
// if local cache enabled then set data in local cache also.
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

// Refresh cache expire time
// if local cache enabled refresh local cache also.
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

// Group cache group.
func (c CacheRedis) Group(group string) string {
	return c.App().ResourceKey("caches_group",
		c.ModelKey(group),
		runtime.WithDelimiter(":"))
}

func (c CacheRedis) Close() {
	c.client.Close()
}

func (c CacheRedis) addGroup(ctx context.Context, key string) error {
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

func (c CacheRedis) delGroup(ctx context.Context) error {
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
