## ETCD配置源

### 配置优先级

> 从上向下优先级依次递增,多个字段之间以英文`/`分隔,不以`/`结尾

- `/{app}/configs/`: 项目相关全局配置
- `/{app}/configs/{env}/`: 项目相关全局配置
- `/{app}/configs/service/{service}/`: 服务相关配置
- `/{app}/configs/service/{service}/{region}/`: 服务region相关配置
- `/{app}/configs/service/{service}/{region}/{az}/`: 服务region，az配置
- `/{app}/configs/{env}/service/{service}/`: 服务相关配置
- `/{app}/configs/{env}/service/{service}/{region}/`: 服务region相关配置
- `/{app}/configs/{env}/service/{service}/{region}/{az}/`: 服务region，az配置
- `/{app}/configs/runtime/{instance.ID}/`: 实例配置

如果同一前缀下存在文件，则文件中所有配置优先级均为该前缀的优先级

建议不要key/value方式和文件同时使用, 因为同一个前缀下优先级一样,启动时配置覆盖和运行时配置覆盖逻辑产生分歧

例如:

```bash
## key/value的方式配置examples.timeout为5ms
/examples/configs/examples/timeout
5ms

## 文件的方式配置examples.timeout为6ms
/examples/configs/global.yaml
examples:
  timeout: 6ms

# 启动时按照assic排序先加载/examples/configs/examples/timeout
# 后加载/examples/configs/global.yaml 会覆盖前面加载的
# 所以你获取到的examples.timeout为6ms
# 但是
# 当你修改时
# 由于他们的值属于同一优先级,就会出现，修改那个获取到的值就是那个
# 如果那天你修改了/examples/configs/examples/timeout后,不重启获取时正常的
# 但是重启服务后发现他没有生效,是因为同一个配置出现在了同一个优先级不同文件中
# 导致运行时的逻辑和启动时的逻辑产生了分歧
```

### 使用

添加配置,文件配置可查看[这里](config_file.md)

```bash
# key/value方式添加配置
etcdctl put /examples/configs/examples/timeout 5ms

# 按照文件方式添加配置key的结尾以json,yaml,yml,toml,props,properties结尾
# 否则以文本方式处理
etcdctl put /examples/configs/examples.yaml 'examples:
  timeout: 6ms'
```

```go

import (
	// 导入etcd配置源
	_ "github.com/asjard/asjard/pkg/config/etcd"
)

config.GetDuration("examples.timeout", time.Second)
```
