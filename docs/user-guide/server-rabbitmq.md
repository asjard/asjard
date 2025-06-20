## 配置

```yaml
asjard:
  servers:
    ## rabbitmq server相关配置
    rabbitmq:
      ## asjard.stores.rabbitmq.client.{name}
      clientName: default
      ## channel qos配置
      ## 一次能接受的最大消息数量
      prefetchCount: 1
      ## 服务器传递最大容量
      prefetchSize: 0
      global: false

```
