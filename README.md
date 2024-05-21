### 配置中心

本地配置:
    - 环境变量
    - Cli参数
    - 文件或目录

远程配置:
    - etcd
    - configcenter
    - apollo

需先加载本地配置， 然后根据本地配置的配置去加载远程配置
远程配置可覆盖本地配置

### 服务启动加载顺序

core: 框架核心并携带一些基础实现
pkg: 对core内容的扩展实现和一些业务功能实现

HanldeChain: 处理链
    registry: 从已注册的服务列表中选取一个服务交给loadbalance
    loadbalance: 从服务列表中选取一个服务


```sh
core
    client ## 连接客户端
    server ## 服务端实现
    registry ## 服务注册发现中心
```