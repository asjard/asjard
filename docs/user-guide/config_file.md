> 文件配置源

## 配置所在目录

- 如果配置了环境变量`ASJARD_CONF_DIR`则读取该目录及子目录下的所有文件
- 否则读取环境变量`ASJARD_HOME_DIR`的值并拼接`conf`目录,读取该目录下及子目录下的所有文件
- 如果以上两个环境变量都没有设置,则读取`可执行程序`平级目录下的`conf`目录下及子目录下的所有文件

## 配置优先级

配置目录下所有文件配置值优先级都一样,

## 支持文件格式

- [x] yaml,yml
- [x] json
- [x] toml
- [x] props,properties

## 使用

例如配置目录下可以添加yaml文件包含如下内容

```yaml
asjard:
  service:
    ## 项目名称
    ## 一个项目下可能会有多个服务
    ## 不实时生效，修改后需重新启动服务
    app: projectAsjardExample
    ## 当前部署环境，例如: dev, sit, uat,rc,pro等
    environment: "dev"
```

程序中可以这样使用

```go
import "github.com/asjard/asjard/core/config"

config.GetString("asjard.service.app", "asjard")
```
