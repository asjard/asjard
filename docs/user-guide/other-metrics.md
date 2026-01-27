> 监控

### 配置

```yaml
asjard:
  ## 监控相关配置
  metrics:
    ## 是否开启监控
    # enabled: false
    ## 需要收集的指标
    # collectors: ""
    ## 内建配置的收集指标
    # builtInCollectors:
    #   - go_collector
    #   - process_collector
    #   - db_default
    #   - api_requests_total
    #   - api_requests_latency_seconds
    #   - api_requests_size_bytes
    #   - api_response_size_bytes
    ## 推送到pushgateway中
    pushGateway:
      ## gateway地址
      # endpoint: http://127.0.0.1:9091
      ## 推送间隔
      # interval: 5s
```

### 自定义指标

```go
import "github.com/asjard/asjard/core/metrics"

func main() {
	// 注册一个计数器指标
	// 如果注册成功，counter返回非nil值，如果注册失败则返回nil值
	// 可反复注册
	// 如果配置了pushgateway则不能携带app,env,service,service_version,instance这些label
	counter := metrics.RegisterCounter("name_of_counter_metrics", "This is a counter help", []string{"label_1", "label_2"})
	if counter !=nil{
		counter.With(map[string][string]{"label_1": "label_1_value", "label_2": "label_2_value"}).Inc()
	}

}
```

## 看板

> 请严格按照[错误码约定](error.md#错误码约定)返回错误,否则无法识别的错误将一律标记为`系统内部错误`

参考[grafana](../media/grafana_asjard.json),效果如下(持续完善中):

![grafana](../media/grafana_dashboard.png)
