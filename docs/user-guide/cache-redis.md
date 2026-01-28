## Redis缓存

### 配置

> 比[全局配置](cache.md)多`client`redis客户端配置

```yaml
## 缓存相关配置
asjard:
  cache:
    ## redis缓存相关配置
    ## 如果不配置则继承asjard.cache
    redis:
      ## redis客户端
      ## asjard.stores.redis.clients.{这里的名称}
      # client: default
```

### 使用

您可以参考[https://github.com/asjard/examples/blob/main/mysql/model/table.go](https://github.com/asjard/examples/blob/main/mysql/model/table.go)

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
