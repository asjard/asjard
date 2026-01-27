## 拦截器名称

circuitBreaker

## 支持协议

- 所有

## 配置

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
