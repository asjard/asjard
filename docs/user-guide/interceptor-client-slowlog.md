## 拦截器名称

slowLog

## 支持协议

- 所有

## 功能

- 打印请求耗时超多阈值的请求日志

## 配置

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
