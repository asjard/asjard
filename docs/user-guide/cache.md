> 缓存使用

详细实例参考`examples/mysql`

## 全局配置

```yaml
asjard:
  cache:
    ## 全局是否开启缓存
    enabled: true
    ## 全局是否自动刷新
    autoRefresh: true
    ## 全局过期时间
    expiresIn: 5m
    ## 全局表缓存配置
    models:
      ## 表名
      modelName:
        ## 表级缓存配置
        enabled: true
        autoRefresh: true
        expiresIn: 5m
```

## 本地缓存

### 配置

> 比全局配置多`publishClient`redis广播客户端配置,移除`autoRefresh`, 本地缓存不自动刷新过期时间，过期强制删除

```yaml
asjard:
  cache:
    ## 本地缓存相关配置
    local:
      ## 是否开启本地缓存
      enabled: true
      ## 过期时间
      expiresIn: 5m
      ## 最大占用空间,单位MB
      maxSize: 102400
      ## 多个实例时删除缓存广播redis客户端
      publishClient: default
      models:
        modelName:
          enabled: true
        ## testNoCache表不开启缓存
        testNoCache:
          enabled: false
```

### 使用

```go
import "github.com/asjard/asjard/pkg/database/cache"

type ExampleTable struct {
	Id        int64  `gorm:"column:id;type:INT(20);primaryKey;autoIncrement"`
	Name      string `gorm:"column:name;type:VARCHAR(20);uniqueIndex"`
	Age       uint32 `gorm:"column:age;type:INT"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ExampleModel struct {
	database.Model
	*ExampleTable
	localCache *cache.CacheLocal
}

// TableName 数据库表名
func (ExampleTable) TableName() string {
	return "example_table"
}

// ModelName 全局唯一的表明
func (ExampleTable) ModelName() string {
	return "example_database_example_table"
}

// Bootstrap 缓存初始化
func (model *ExampleModel) Bootstrap() (err error) {
	// 本地缓存初始化
	model.localCache, err = cache.NewLocalCache(model.ExampleTable)
	if err != nil {
		return err
	}
	return nil
}

func (model *ExampleModel) Search(ctx context.Context, in *pb.SearchReq) (*pb.ExampleList, error) {
	var result pb.ExampleList
	if err := model.GetData(ctx, &result,
		// 通过本地缓存获取数据
		model.localCache.WithKey(model.searchCacheKey(in)),
		func() (any, error) {
			return model.ExampleTable.Search(ctx, in)
		}); err != nil {
		return nil, err
	}
	return &result, nil
}
```

## Redis缓存

### 配置

> 比全局配置多`client`redis客户端配置

```yaml
asjard:
  cache:
    ## redis缓存相关配置
    redis:
      ## redis客户端
      client: default
      # 是否自动刷新
      autoRefresh: true
      # 过期时间
      expiresIn: 5m
      # 是否开启某个表的缓存
      models:
        modelName:
          enabled: true
```

### 使用

```go

// Bootstrap 缓存初始化
func (model *ExampleModel) Bootstrap() (err error) {
	localCache, err := cache.NewLocalCache(model.ExampleTable)
	if err != nil {
		return err
	}
	// redis缓存初始化, 本地缓存作为redis缓存的数据源
	// 现获取本地缓存，如果不存在再获取redis缓存，都获取不到获取再去数据库获取
	model.kvCache, err = redis.NewKeyValueCache(model.ExampleTable,
		redis.WithLocalCache(localCache))
	if err != nil {
		return err
	}
	return nil
}

func (model *ExampleModel) Update(ctx context.Context, in *pb.CreateOrUpdateReq) (*pb.ExampleInfo, error) {
	if err := model.SetData(ctx,
		// 删除缓存key，及分组下的所有缓存key
		model.kvCache.WithGroup(model.searchGroup()).WithKey(model.getCacheKey(in.Name)),
		func() error {
			if _, err := model.ExampleTable.Update(ctx, in); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return model.Get(ctx, &pb.ReqWithName{Name: in.Name})
}

func (model *ExampleModel) Search(ctx context.Context, in *pb.SearchReq) (*pb.ExampleList, error) {
	var result pb.ExampleList
	if err := model.GetData(ctx, &result,
		// 通过redis缓存获取
		// 并将缓存key添加到一个分组中
		model.kvCache.WithKey(model.searchCacheKey(in)).WithGroup(model.searchGroup()),
		func() (any, error) {
			return model.ExampleTable.Search(ctx, in)
		}); err != nil {
		return nil, err
	}
	return &result, nil
}
```
