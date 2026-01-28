> 配置相关

## 配置

```yaml
## 配置中心相关
asjard:
  config:
    ## 添加配置默认配置源
    ## Set方法默认配置源, 如果不配置或者为空，则发送给所有配置源, 默认mem
    ## 具体是否能够添加配置到配置源中要看具体配置源是否实现Set功能
    # setDefaultSource: mem
```

## 配置获取

- `Get(key string, options *Options) any`: 根据`key`获取配置

```go
import "github.com/asjard/asjard/core/config"

val := config.Get("key", &Options{})
if val == nil {
 	// 配置不存在
 	// 或者使用config.Exist("key")判断
}
```

- `GetWithPrefix(prefix string, opts ...Option) map[string]any`: 根据前缀获取所有配置,返回的key是props格式的

```go
import "github.com/asjard/asjard/core/config"

valMap := config.GetWithPrefix("asjard.service")
/* Output:
{
 	"app": "asjard",
 	"environment": "dev",
  "instance.name": "example"
}
*/
```

- `GetString(key string, defaultValue string, opts ...Option) string`: 获取配置并返回string类型,如果配置不存在则返回默认值

```go
import "github.com/asjard/asjard/core/config"

valStr := config.GetString("key", "default_value")
```

- `GetStrings(key string, defaultValue []string, opts ...Option) []string`: 获取配置并返回[]string类型, 如果配置不存在则返回默认值

````go
import "github.com/asjard/asjard/core/config"

// 可通过字符串或者列表方式配置
// 例如yaml文件中:

// ```yaml
// key: a,b,c
// ```
// 字符串方式分隔符可通过config.WithDelimiter指定，默认为','

// 或者

// ```yaml
// key:
// - a
// - b
// - c
// ```
valStrs := config.GetStrings("key", []string{})

````

- `GetByte(key string, defaultValue []byte, opts ...Option) []byte`: 获取配置并返回[]byte类型, 如果配置不存在则返回默认值

```go
import "github.com/asjard/asjard/core/config"

valByte := config.GetByte("key", []byte{})
```

- `GetBool(key string, defaultValue bool, opts ...Option) bool`: 获取配置并返回bool类型

```go
import "github.com/asjard/asjard/core/config"

// 字符串则除0, f, false, n, no, off为false为其他均为true
// 数字不等于0则为true
valBool := config.GetBool("key", false)
```

- `GetBools(key string, defaultValue []bool, opts ...Option) []bool`
- `GetInt(key string, defaultValue int, opts ...Option) int `
- `GetInts(key string, defaultValue []int, opts ...Option) []int `
- `GetInt64(key string, defaultValue int64, opts ...Option) int64 `
- `GetInt64s(key string, defaultValue []int64, opts ...Option) []int64 `
- `GetInt32(key string, defaultValue int32, opts ...Option) int32 `
- `GetInt32s(key string, defaultValue []int32, opts ...Option) []int32 `
- `GetFloat64(key string, defaultValue float64, opts ...Option) float64`
- `GetFloat64s(key string, defaultValue []float64, opts ...Option) []float64 `
- `GetFloat32(key string, defaultValue float32, opts ...Option) float32 `
- `GetFloat32s(key string, defaultValue []float32, opts ...Option) []float32 `
- `GetDuration(key string, defaultValue time.Duration, opts ...Option) time.Duration`

字符串配置可参考`time.ParseDuration`, 支持单位"ns", "us" (or "µs"), "ms", "s", "m", "h".

```go
import "github.com/asjard/asjard/core/config"

conf := `
examples:
	timeoutInt: 1 # 1ns
	timeoutStr: 1s  # 1s
	timeoutStr1: "1" # 1ns
`
valDura := config.GetDuration("key", time.Second)
```

- `GetTime(key string, defaultValue time.Time, opts ...Option) time.Time `
- `Exist(key string) bool `
- `GetAndUnmarshal(key string, outPtr any, opts ...Option) error `: 获取配置并将对值反序列化

```go
import "github.com/asjard/asjard/core/config"

conf := `
examples:
	mysql: |
		db: database
		user: user
		password: pwd
`

type Config struct{
	DB string `json:"db"`
	User string `json:"user"`
	Password string `json:"pw"`
}

var dbConf Config
// 默认使用json反序列化
if err := config.GetAndUnmarshal("examples.mysql", &dbConf); err != nil {
	// TODO err
}

// 自定义反序列化
// Unmarshaler 反序列化需要实现的方法
type Unmarshaler interface {
	Unmarshal(data []byte, v any) error
}

type CustomeUnmarshal struct{}

func (CustomeUnmarshal) Unmarshal(data []byte, v any) error {
	return nil
}

if err := config.GetWithUnmarshal("examples.mysql", &dbConf, config.WithUnmarshaler(&CustomeUnmarshal{})); err != nil {
	// TODO err
}
```

- `GetAndJsonUnmarshal(key string, outPtr any, opts ...Option) error`: 配置值json反序列化
- `GetAndYamlUnmarshal(key string, outPtr any, opts ...Option) error `: 配置值yaml反序列化
- `GetWithUnmarshal(prefix string, outPtr any, opts ...Option) error `

```go
import "github.com/asjard/asjard/core/config"

conf := `
examples:
	mysql:
		db: xx
		user: user
		password: pwd
`

type Config struct{
	DB string `json:"db"`
	User string `json:"user"`
	Password string `json:"pw"`
}

var dbConf Config
// 默认使用json反序列化
if err := config.GetWithUnmarshal("examples.mysql", &dbConf); err != nil {
	// TODO err
}
```

- `GetWithJsonUnmarshal(prefix string, outPtr any, opts ...Option) error `
- `GetWithYamlUnmarshal(prefix string, outPtr any, opts ...Option) error `

## 配置[加解密](security.md)

> 加密内容始终都是以密文存储在内存中的，只有在获取时才会解密

- 手动解密

```go
import "github.com/asjard/asjard/core/config"

// base64解码后返回
val := config.GetString("encrypted_key","default_value", config.WithCipher("base64"))
```

- 自动解密

```yaml
## value值以'encrypted_'为前缀,用英文':'分割值与前缀
## 值明文为encrypted_value
encrypted_key: encrypted_base64:ZW5jcnlwdGVkX3ZhbHVl
```

```go
val := config.GetString("encrypted_key", "default_value")
// output: encrypted_value
// 或者禁用自动解密
val := config.GetString("encrypted_key", "default_value", config.WithDisableAutoDecryptValue())
// output: encrypted_base64:ZW5jcnlwdGVkX3ZhbHVl
```

## 配置监听

```go
import "github.com/asjard/asjard/core/config"

var val string

val = config.GetString("key", "default_value", config.WithWatch(func(event *config.Event){
	val = cast.ToString(event.Value.Value)
})
```

或者

- `AddListener(key string, callback func(*Event)) `: 当key发生变化时通过callback通知
- `AddPatternListener(pattern string, callback func(*Event)) `: pattern正则表达式匹配key变化时通过callback通知

## 配置源

> 框架内置`环境变量`,`文件`,`内存`配置源, 无需导入

| 支持 | 配置源                     | 优先级 | 描述                           |
| :--: | :------------------------- | :----: | ------------------------------ |
|  ✅  | [环境变量](config-env.md)  |   0    |
|      | cli                        |   1    |
|  ✅  | [文件](config-file.md)     |   2    |
|  ✅  | [etcd](config-etcd.md)     |   10   | key/value, file模式配置        |
|  ✅  | [consul](config-consul.md) |   11   | ket/value模式配置,file模式配置 |
|      | nacos                      |   12   | 没有删除事件                   |
|      | apollo                     |   13   |
|      | configmap                  |   14   |
|  ✅  | [本地内存](config-mem.md)  |   99   |

## 配置源优先级

数字越大的优先级越高, 相同key的配置,优先级高的配置源覆盖优先级低的配置源

## 自定义配置源

实现如下方法

```go
// Sourcer 配置源需要实现的方法
type Sourcer interface {
	// 获取所有配置,首次初始化完毕后会去配置源获取一次所有配置,
	// 维护在config_manager的本地内存中,
	// 返回的配置应该为properties格式的，并区分大小写。
	// 返回值可以通过ConvertToProperties方法获取
	GetAll() map[string]*Value
	// 添加配置到配置源中,
	// 慎用,存在安全隐患和配置源实现复杂问题
	// 理论只应该在mem配置源中使用,非必要不要使用
	Set(key string, value any) error
	// 监听配置变化,当配置源中的配置发生变化时,
	// 通过此回调方法通知config_manager进行配置变更
	Watch(func(event *Event)) error
	// 和配置中心断开连接
	Disconnect()
	// 配置中心的优先级
	Priority() int
	// 配置源名称
	Name() string
}
```

## 多配置源同时使用

```go
import (
	// 导入etcd配置源
	_ "github.com/asjard/asjard/pkg/config/etcd"
	// 导入consul配置源
	_ "github.com/asjard/asjard/pkg/config/consul"
)
// 同一个配置项始终会获得最高优先级配置源的值
```

more examples at [here](https://github.com/asjard/asjard/tree/develop/_examples/svc-example/apis/api/v1/config.go)
