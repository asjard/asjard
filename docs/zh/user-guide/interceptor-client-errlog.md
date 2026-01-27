## 拦截器名称

errLog

## 支持协议

- 所有

## 功能

- 统一请求错误日志打印

### 配置

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
