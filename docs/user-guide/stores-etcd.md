## 配置

```yaml
asjard:
  etcd:
    clients:
      default:
        endpoints:
          - 127.0.0.1:2379
      config:
        endpoints: 127.0.0.1:2379,127.0.0.1:2380
      registry:
        endpoints: 127.0.0.1:3379,127.0.0.1:3380
        options:
          dialTimeout: 5s
    options:
      dialTimeout: 5s
```

## 使用

```go
import "github.com/asjard/asjard/pkg/stores/xetcd"

// 使用默认客户端
client, err := xetcd.Client()
if err != nil {
	return err
}

// 自定义客户端
// 前提是需要配置asjard.stores.etcd.clients.xxx
client, err := xetcd.Client(xetcd.WithClientName("xxx"))
```
