> etcd服务注册发现

## 配置

```yaml
asjard:
  ## 服务发现，注册相关配置
  registry:
    ## etcd注册发现中心相关配置
    etcd:
      ## 配置asjard.stores.etcd.clients中的数据库名称(key)
      # client: default
      ## 超时时间
      # timeout: 5s
      ## 注册租约 TTL，过短会导致 keepalive 通道被关闭后反复重注册
      # ttl: 30s
```

## 使用

```go
import _"github.com/asjard/asjard/pkg/registry/etcd"
```
