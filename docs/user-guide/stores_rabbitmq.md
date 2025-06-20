## 配置

```yaml
asjard:
  stores:
    rabbitmq:
      clients:
        default:
          url: ""
          vhost: ""
        options:
          channelMax: 0
          frameSize: 0
          heartBeat: 1s
```


## 使用

```go
import "github.com/asjard/asjard/pkg/stores/xrabbitmq"

client, err := xrabbitmq.Client()
if err != nil {
	return err
}
```
