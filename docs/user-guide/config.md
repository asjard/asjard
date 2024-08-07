> 配置相关

## 支持配置源

- [x] 文件, 优先级: 2
- [x] 内存, 优先级: 99
- [x] 环境变量, 优先级: 0
- [ ] cli, 优先级: 1
- [x] etcd, 优先级: 10

## 配置优先级

数字越大的优先级越高, 相同key的配置,优先级高的覆盖优先级低的

## 文件配置

### 配置所在目录

- 如果配置了环境变量`ASJARD_CONF_DIR`则读取该目录及子目录下的所有文件
- 否则读取环境变量`ASJARD_HOME_DIR`的值并拼接`conf`目录,读取该目录下及子目录下的所有文件
- 如果以上两个环境变量都没有设置,则读取`可执行程序`平级目录下的`conf`目录下及子目录下的所有文件

### 支持文件格式

- [x] yaml,yml
- [ ] json
- [ ] ini
- [ ] prop,properties

## 环境变量配置

- 框架配置都会以`asjard`为前缀
- 不同层级的配置中间以`_`分隔, 例如`asjard_app`, 程序使用`asjard.app`读取
- 大小写敏感, 例如`asjard_app`和`asjard_APP`为两个不同的配置i

```go
// 在环境变量中配置如下配置
// export asjard_app=asjard
// 程序中可以这样读
config.GetString("asjard.app", "")
// Output: asjard
```

## ETCD配置

### 配置优先级

> 从上向下优先级依次递增,多个字段之间已英文`/`分隔,不以`/`结尾

- `/{app}/configs/global/`: 项目相关全局配置
- `/{app}/configs/service/{service}/`: 服务相关配置
- `/{app}/configs/service/{service}/{region}/`: 服务region相关配置
- `/{app}/configs/service/{service}/{region}/{az}/`: 服务region，az配置
- `/{app}/configs/runtime/{instance.ID}/`: 实例配置

### 使用

```go
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
