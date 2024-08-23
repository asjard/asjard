## 配置

```yaml
asjard:
  ## 数据相关配置
  stores:
    redis:
      clients:
        default:
          host: 127.0.0.1
          port: 6379
          db: 0
          auth: xxx
      options:
        ## 是否检查状态, 如果为ture则在连接到redis后不执行Ping
        disableCheckStatus: false
        ## 是否开启链路追踪
        traceable: false
```

## 使用

```go
import "github.com/asjard/asjard/pkg/stores/xredis"

// 使用默认客户端
client, err := xredis.Client()
if err != nil {
	return err
}

// 自定义客户端
// 前提是需要配置asjard.stores.redis.clients.xxx
client, err := xredis.Client(xetcd.WithClientName("xxx"))
```
