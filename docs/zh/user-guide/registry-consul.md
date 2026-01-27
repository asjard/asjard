> consul 服务注册发现

## 配置

```yaml
asjard:
  ## 服务发现，注册相关配置
  registry:
    ## consul注册发现中心相关配置
    consul:
      ## 配置asjard.stores.consul.clients中的数据库名称
      # client: default
      ## 超时时间
      # timeout: 5s
```

## 使用

```go
import _ "github.com/asjard/asjard/pkg/registry/consul"
```
