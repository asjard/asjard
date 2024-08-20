## 本地缓存

### 配置

> 比[全局配置](cache.md)多`redisClient` redis广播客户端配置,
> 当有多个实例时可通过redis广播通知其他实例删除缓存
> 如果不配置[全局配置](cache.md)中的字段则继承全局配置

```yaml
asjard:
  cache:
    ## 本地缓存相关配置
    local:
      ## 是否开启本地缓存
      ## 如果不配置则继承asjard.cache.enabled
      ## 除redisClient字段外其他同理
      enabled: true
      ## 过期时间
      expiresIn: 5m
      ## 最大占用空间,单位MB
      maxSize: 102400
      ## 多个实例时删除缓存广播redis客户端
      redisClient: default
      models:
        modelName:
          enabled: true
        ## testNoCache表不开启缓存
        testNoCache:
          enabled: false
```

### 使用

您可以参考[https://github.com/asjard/examples/blob/main/mysql/model/table.go](https://github.com/asjard/examples/blob/main/mysql/model/table.go)

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
