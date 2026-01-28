## 导入模版仓库

- 通过submodule的方式导入[bifrost-template](https://github.com/asjard/bifrost-template),或者只是参考

## 服务中自定义makefile

```makefile
export BIFROST_DIR ?= ../../third_party/bifrost
export PROTO_DIR ?= ../protos-repo
export PROTOC_OPT ?= "-I../third_party"


export GEN_PROTO_GO_VALIDATE ?= true
export GEN_PROTO_GO_REST ?= true
export GEN_PROTO_GO_REST_GW ?= true
export GEN_PROTO_GO_AMQP ?= true
export GEN_PROTO_GO_VALIDATE_OPT ?= validate_enum=true
# export DEBUG ?= true

-include $(BIFROST_DIR)/Makefile_build


run_dev: ## 本地运行服务
	ASJARD_CONF_DIR="$(PWD)/conf $(PWD)/apis/api/conf"  go run -race ./apis/api

```

## 查看帮助

```bash
make help
```

你将看到类似如下结果

```bash
Commands:
  run_dev      本地运行服务
  build        编译
  gen_proto    生成协议文件
  clean_proto  清理proto协议生成的文件
  archive      打包
  install      安装
  restart      重启
  start        启动
  stop         停止
  uninstall    卸载
  run          本地运行,依赖docker和docker-compose
  down         本地卸载,依赖docker-compose
  test         运行测试用例
  configure    配置
  help         使用帮助
Envs:
  BIFROST_DIR                bifrost项目所在目录,结尾不要有/                默认: .                      当前: ../../third_party/bifrost
  PROTO_DIR                  proto协议文件目录,结尾不要有/                  默认: .                      当前: ../protos-repo
  PROTOC_OPT                 protoc命令参数                                 默认:                        当前: -I../third_party
  GEN_PROTO_GO               是否根据protobuf文件生成*.pb.go文件            默认: true                   当前: true
  GEN_PROTO_GO_OUT           生成的*.pb.go文件输出目录                      默认: .                      当前: .
  GEN_PROTO_GO_OPT           生成*.pb.go需要的参数                          默认:                        当前:
  GEN_PROTO_GO_GRPC          是否根据protobuf文件生成*_grpc.pb.go文件       默认: true                   当前: true
  GEN_PROTO_GO_GRPC_OPT      生成*_grpc.pb.go所需要的参数                   默认:                        当前:
  GEN_PROTO_GO_REST          是否根据protobuf文件生成*_rest.pb.go文件       默认: true                   当前: true
  GEN_PROTO_GO_REST_OPT      生成*_rest.pb.go所需要的参数                   默认:                        当前:
  GEN_PROTO_GO_REST_GW       是否根据protobuf文件生成*_rest_gw.pb.go文件    默认: true                   当前: true
  GEN_PROTO_GO_REST_GW_OPT   生成*_rest_gw.pb.go所需要的参数                默认:                        当前:
  GEN_PROTO_TS               是否根据protobuf文件生成*.d.tsx文件            默认: false                  当前: false
  GEN_PROTO_TS_OPT           生成*.d.tsx所需要的参数                        默认:                        当前:
  GEN_PROTO_TS_ENUM          是否根据protobuf文件生成*.enum.tsx文件         默认: false                  当前: false
  GEN_PROTO_TS_ENUM_OPT      生成*.enum.tsx所需要的参数                     默认:                        当前:
  GEN_PROTO_TS_UMI           是否根据protobuf生成*.umi.tsx文件              默认: fasle                  当前: fasle
  GEN_PROTO_TS_UMI_OPT       生成*.umi.tsx所需要的参数                      默认:                        当前:
  GEN_PROTO_TS_OUT           生成的*.pb.ts文件输出目录                      默认: ./..                   当前: ./..
  GEN_PROTO_GO_ASYNQ         是否根据protobuf文件生成*_asynq.pb.go文件      默认: false                  当前: false
  GEN_PROTO_GO_ASYNQ_OPT     生成*_asynq.pb.go所需要的参数                  默认:                        当前:
  GEN_PROTO_GO_AMQP          是否根据protobuf文件生成*_amqp.pb.go文件       默认: false                  当前: true
  GEN_PROTO_GO_AMQP_OPT      生成*_amqp.pb.go所需要的参数                   默认:                        当前:
  GEN_PROTO_GO_VALIDATE      是否根据protobuf文件生成*_validate.pb.go文件   默认: true                   当前: true
  GEN_PROTO_GO_VALIDATE_OPT  生成*_validate.pb.go文件的参数                 默认:                        当前: validate_enum=true
  GOOS                       运行环境,可选:linux,darwin,windows             默认: linux                  当前: linux
  CGO_ENABLED                是否开启CGO,可选:0,1                           默认: 0                      当前: 0
  BUILD_DIR                  编译目录                                       默认: .                      当前: .
  DEBUG                      是否开启DEBUG                                  默认: false                  当前: false
  PROJECT_NAME               项目名称                                       默认: bifrost                当前: bifrost
  DEPLOY_ENV                 部署环境,可选:dev,sit,uat,pro                  默认: dev                    当前: dev
  SERVICE_NAME               服务名称                                       默认: bifrost                当前: bifrost
  VERSION                    版本号,默认当前目录下version文件的内容         默认: v0.0.1   当前: v0.0.1
```
