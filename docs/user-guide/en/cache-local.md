## Local Cache

## Config

> Reference [Cache](cache.md)
> If you have many instances, you can delete other instances local cache use redis

```yaml
asjard:
  ## inherit from asjard.cache
  local:
    ## Redis client
    ## Delete other instances local cache
    # redisClient: default
    ## Max memory size use, default 100 * 1024 * 1024
    # maxSize: 104857600
```

## Use

Reference [https://github.com/asjard/examples/blob/main/mysql/model/table.go](https://github.com/asjard/examples/blob/main/mysql/model/table.go)

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

// TableName table name
func (ExampleTable) TableName() string {
	return "example_table"
}

// ModelName model name will used at cache key
func (ExampleTable) ModelName() string {
	return "example_database_example_table"
}

// Bootstrap use bootstrap to init local cache
func (model *ExampleModel) Bootstrap() (err error) {
	model.localCache, err = cache.NewLocalCache(model.ExampleTable)
	if err != nil {
		return err
	}
	return nil
}

func (model *ExampleModel) Search(ctx context.Context, in *pb.SearchReq) (*pb.ExampleList, error) {
	var result pb.ExampleList
	if err := model.GetData(ctx, &result,
		model.localCache.WithKey(model.searchCacheKey(in)),
		func() (any, error) {
			return model.ExampleTable.Search(ctx, in)
		}); err != nil {
		return nil, err
	}
	return &result, nil
}
```
