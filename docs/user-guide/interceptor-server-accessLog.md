## 拦截器名称

ccessLog

## 支持协议

- 所有

## 功能

- 请求日志打印

## 配置

```yaml
asjard:
  logger:
    accessLog:
      ## 是否开启accessLog
      enabled: true
      ## 配置格式: [protocol://]{fullMethod}
      ## 例如grpc协议的某个方法: grpc:///api.v1.hello.Hello/Call
      ## 或者协议无关的某个方法: /api.v1.hello.Hello/Call
      ## 拦截协议的所有方法: grpc
      skipMethods:
        - grpc
```
