## 本地优先轮询

- 优先当前az下实例,
- 无当前az下实例时选择其他az下的共享实例

## 配置

- 负载均衡配置

```yaml
## client configurations
clients:
  ## client loadbalance, default: localityRoundRobin
  # loadbalance: "localityRoundRobin"
  ## grpc client configuration
  grpc:
    ## grpc client loadbalance
    # loadbalance: ""
```

- 共享实例

```yaml
asjard:
  service:
    instance:
      ## Can be accessed by services in different az
      # shareable: false
```
