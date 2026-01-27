> 客户端拦截器

## 配置

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
