> 建议但不强制,也可按照自己项目习惯定义自己的项目结构

```bash
_examples
├── protos-repo     ## 协议仓库
│   ├── Makefile    ## 运维入口
│   ├── README.md   ## 仓库描述
│   ├── example     ## 服务
│   │   └── api     ## 服务所对应的API类型，比如管理后台协议: api, 开放协议: openapi, 商户协议: merchantapi
│   │       └── v1  ## 协议版本
│   │           ├── example ## protoc生成的文件，无需修改
│   │           │   ├── example.pb.go
│   │           │   ├── example_amqp.pb.go
│   │           │   ├── example_grpc.pb.go
│   │           │   ├── example_rest.pb.go
│   │           │   ├── example_rest_gw.pb.go
│   │           │   └── example_validate.pb.go
│   │           └── example.proto   ## 协议描述
│   ├── third_party ## 三方协议或者无法通过当前语言包管理器管理的三方库，可通过fork或submodule管理在此目录下
│   │   └── github.com
│   └── version ## 协议版本
└── svc-example     ## example服务
    ├── Dockerfile  ## docker image创建配置
    ├── Makefile    ## 运维入口
    ├── apis        ## 所有实现的API都归类到此目录下
    │   ├── api     ## 对应API类型的实现
    │   │   ├── conf  ## 当前API类型所独有的配置
    │   │   │   ├── server.yaml
    │   │   │   └── service.yaml
    │   │   ├── main.go ## 当前API类型的程序入口
    │   │   └── v1  ## API版本
    │   │       └── example.go ## 业务逻辑
    │   └── openapi
    │       ├── main.go
    │       └── v1
    ├── conf  ## 当前服务的全局配置
    │   └── example.yaml
    ├── datas ## 数据持久化
    │   └── example.go
    ├── services ## 业务逻辑和数据持久化的连接层,比如缓存在这一层实现
    │   └── example.go
    ├── third_party ## 无法通过当前语言包管理器管理的三方库，可通过fork或submodule管理在此目录下
    └── version ## 服务版本
```
