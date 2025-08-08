> 客户端拦截器

## 配置

### 熔断

```yaml
asjard:
  ## 拦截器相关配置
  interceptors:
    ## 客户端拦截器
    client:
      ## 断路器相关配置
      ## 服务
      ## https://github.com/afex/hystrix-go/blob/master/hystrix/settings.go#CommandConfig
      ## 优先级 methods -> service -> default
      circuitBreaker:
        ## 默认配置
        ## 超时时间,单位毫秒
        # timeout: 1000
        # max_concurrent_requests: 1000
        # request_volume_threshold: 20
        # sleep_window: 5000
        # error_percent_threshold: 50
        ## 方法优先级
        ## protocol://service/method
        ## protocol://service
        ## protocol:///method
        ## protocol
        ## //service/method
        ## ///method
        ## //service
        methods:
          - name: grpc://servicesName/method
            timeout: 1000
          - name: //serviceName
```

### rest转grpc协议

```yaml
asjard:
  ## 拦截器相关配置
  interceptors:
    ## 客户端拦截器
    client:
      ## rest请求头注入到rpc的context上下文中
      rest2RpcContext:
        ## 允许注入的请求头
        # allowHeaders: ""
        ## 内建允许注入的请求头
        # builtInAllowHeaders:
        #   - x-request-region
        #   - x-request-az
        #   - x-request-id
        #   - x-request-instance
        #   - x-forward-for
        #   - traceparent
```

### 慢日志

```yaml
asjard:
  interceptors:
    client:
      slowLog:
        ## 慢阈值
        # slowThreshold: 0
        ## 需要忽略的方法
        # skipMethods:
```

### 错误日志

```yaml
asjard:
  interceptors:
    client:
      errLog:
        ## 是否开启错误日志
        # enabled: true
        ## 需要忽略的方法
        # skipMethods: ""
```
