> 本地服务发现

## 配置

```yaml
asjard:
  ## 服务发现，注册相关配置
  registry:
    ## 本地服务发现配置
    ## 实时生效，无需重启服务
    localDiscover:
      ## 服务名称
      ## 配置格式{protocol}://{ip}:{port}
      # helloGrpc:
      # - grpc://127.0.0.1:7010
      # - grpc://127.0.0.1:7011
```
