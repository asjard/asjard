## 本地缓存

详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/svc-example/services/user.go)

### 配置

> 比[全局配置](cache.md)多`redisClient` redis广播客户端配置,
> 当有多个实例时可通过redis广播通知其他实例删除缓存
> 如果不配置[全局配置](cache.md)中的字段则继承全局配置

```yaml
## 缓存相关配置
asjard:
  ## 本地缓存相关配置
  ## 除redisClient字段其他字段如果不配置则继承asjard.cache
  local:
    ## redis客户端
    ## 多实例情况下需要通过redis删除其他节点的缓存
    # redisClient: default
    ## 最大内存使用, 默认100 * 1024 * 1024
    # maxSize: 104857600
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
