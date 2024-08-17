## ETCD配置源

### 配置优先级

> 从上向下优先级依次递增,多个字段之间以英文`/`分隔,不以`/`结尾

- `/{app}/configs/global/`: 项目相关全局配置
- `/{app}/configs/global/{env}/`: 项目相关全局配置
- `/{app}/configs/service/{service}/`: 服务相关配置
- `/{app}/configs/service/{service}/{region}/`: 服务region相关配置
- `/{app}/configs/service/{service}/{region}/{az}/`: 服务region，az配置
- `/{app}/configs/service/{env}/{service}/`: 服务相关配置
- `/{app}/configs/service/{env}/{service}/{region}/`: 服务region相关配置
- `/{app}/configs/service/{env}/{service}/{region}/{az}/`: 服务region，az配置
- `/{app}/configs/runtime/{instance.ID}/`: 实例配置

### 使用

```go

import (
	// 导入etcd配置源
	_ "github.com/asjard/asjard/pkg/config/etcd"
)

// 例如全局配置
// /app/configs/global/examples/timeout => 5ms
config.GetDuration("examples.timeout", time.Second)
// Output: 5ms

// 服务配置
// /app/configs/service/exampleService/examples/timeout => 6ms
config.GetDuration("examples.timeout", time.Second)
// Output: 6ms

// 其他同上
```
