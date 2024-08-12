> 开启和关闭pprof

## 配置

```yaml
asjard:
  servers:
    pprof:
      ## 开启后通过http://address.listen/debug/pprof访问
      enabled: true
      addresses:
        listen: 127.0.0.1:7032
```
