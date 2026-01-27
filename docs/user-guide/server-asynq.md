## 配置

```yaml
asjard:
  servers:
    asynq:
      enabled: false
      redis: default
      optios:
        ## 为0或者负数则为CPU核心数量
        concurrency: 4
        queue: {}
        strictPriority: false
        shutdownDuration: 8s
        healthCheckInterval: 15s
        delayedTaskCheckInterval: 5s
        groupGracePeriod: 1s
        groupMaxDelay: 0
        groupMaxSize: 0

```
