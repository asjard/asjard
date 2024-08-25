## 配置

```yaml
## 配置中心相关
asjard:
  config:
    ## consul配置中心相关配置
    consul:
      ## 配置中心名称
      ## 依赖asjard.stores.consul.clients.{cllient}的配置
      # client: default
      ## 分隔符
      ## consul中多个key之间分隔符
      # delimiter: "/"
```

### 使用

```go
import (
	// 导入consul配置源
	_ "github.com/asjard/asjard/pkg/config/consul"
)

// 其他使用方法同ETCD
```
