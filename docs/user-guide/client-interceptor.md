> 客户端拦截器

## 配置

### 熔断

```yaml
asjard:
  interceptors:
    client:
      circuitBreaker:
        ## 超时时间,单位md
        timeout: 1000
        ## 最大并发请求
        max_concurrent_request: 10
        ## 一个窗口内请求数量的阈值，判断熔断的条件之一
        request_volume_threshold: 20
        ## 单位ms
        sleep_window: 5000
        ## 请求错误百分比
        error_percent_threshold: 50
```
