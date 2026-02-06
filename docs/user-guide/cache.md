> 缓存使用

详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/svc-example/services/user.go)

## 全局配置

> 其他自定义缓存如果没有配置如下字段，则继承全局配置

```yaml
## 缓存相关配置
asjard:
  cache:
    ## 全局是否开启缓存
    # enabled: false
    ## 全局是否自动刷新
    # autoRefresh: false
    ## 关注版本差异
    ## 默认缓存key不带version标签
    ## 如果设置为true升级服务版本后缓存将会集体失效
    # careVersionDiff: false
    # ignoreAppDiff: false
    # ignoreEnvDiff: false
    # ignoreServiceDiff: false
    # ignoreRegionDiff: false
    # ignoreAzDiff: false
    ## 全局过期时间
    # expiresIn: 10m
    ## 空值过期时间
    ## 如果不设置则为expiresIn的一半
    # emptyExpiresIn: 5m
    ## 全局表缓存配置
    models:
      ## 表名
      ## 配置同asjard.cache相关配置
      modelName:
        # enabled: false
        # autoRefresh: false
```

## 自定义缓存

实现如下方法

```go
// Cacher 缓存需要实现的方法
type Cacher interface {
	// 从缓存获取数据
	// fromCurrent 获取到的数据是从当前缓存中获取到的
	Get(ctx context.Context, key string, out any) (fromCurrent bool, err error)
	// 从缓存删除数据
	Del(ctx context.Context, keys ...string) error
	// 设置缓存数据
	Set(ctx context.Context, key string, in any, expiresIn time.Duration) error
	// 刷新缓存过期时间
	Refresh(ctx context.Context, key string, in any, expiresIn time.Duration) error
	// 返回缓存Key名称
	Key() string

	// 是否开启了缓存
	Enabled() bool
	// 是否自动刷新缓存
	AutoRefresh() bool
	// 过期时间
	ExpiresIn() time.Duration
	// 空值过期时间
	EmptyExpiresIn() time.Duration
}
```

具体实现可参考[https://github.com/asjard/asjard/blob/develop/pkg/cache/cache_redis.go](https://github.com/asjard/asjard/blob/develop/pkg/cache/cache_redis.go)redis缓存的实现
