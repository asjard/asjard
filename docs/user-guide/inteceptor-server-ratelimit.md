## 拦截器名称

rateLimiter

## 支持协议

- 所有

## 功能

- 服务端请求限速,快速响应失败
- 非分布式限速器，目的只是为了保护当前实例

## 配置

```yaml
asjard:
  ## interceptor configurations.
  interceptors:
    ## server interceptor
    server:
      ## ratelimiter configuration.
      rateLimiter:
        # enabled: false
        ## Maximum number of requests per second
        ## <0 means no limit
        # limit: -1
        ## Bucket capacity, if less than 0, it is the limit value.
        # burst: -1
        ## specify method ratelimiter configuration.
        methods:
          ## method name
          ## [{protocol}://]{method}
          ## All protocol health checks are limited to 10 requests per second
          # - name: /asjard.api.health.Health/Check
          #   limit: 10
          #   burst: 10
          ## The grpc protocol /api.v1.server.Server/Hello has a rate limit of 20 requests per second
          # - name: grpc:///api.v1.server.Server/Hello
          #   limit: 20
          #   burst: 20
```
