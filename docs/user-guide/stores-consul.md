## 配置

```yaml
asjard:
  ## 数据相关配置
  stores:
    consul:
      clients:
        default:
          ## address,username,password受cipherName保护
          # address: 127.0.0.1:8500
          # schema: ""
          # pathPrefix: ""
          # datacenter: ""
          # username: ""
          # password: ""
          # waitTime: 0s
          # token: ""
          # namespace: ""
          # partition: ""
          ## 解密组件名称
          cipherName: ""
          cipherParams: {}
```

## 使用

```go
import "github.com/asjard/asjard/pkg/stores/consul"

client, err := consul.Client()
// OR
client, err := consul.Client(consul.WithClientName("config"))
```
