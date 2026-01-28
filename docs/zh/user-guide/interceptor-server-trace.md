## 拦截器名称

trace

## 支持协议

- 所有

## 功能

- 链路追踪

## 配置

```yaml
asjard:
  trace:
    enabled: false

    ## the address that support otel protocol
    ## http://127.0.0.1:4318
    ## grpc://127.0.0.1:4319
    # endpoint: http://127.0.0.1:4318

    # timeout: 1s
    ## relative ASJARD_CERT_DIR
    # certFile: ""
    # keyFile: ""
    # caFile: ""
```
