> protobuf驱动插件式的微服务框架,通过简单的配置即可实现相应功能或者变更程序逻辑,插件式按需加载自己想要的功能或者定制自己的插件以满足业务需求

## 开发日志

- [x] [系统启动](./core/bootstrap/README.md)
- [x] [客户端管理](./core/client/README.md)
- [x] [配置管理](./core/config/README.md)
- [x] [日志管理](./core/logger/README.md)
- [x] [注册发现管理](./core/registry/REAME.md)
- [x] [运行时](./core/runtime/README.md)
- [x] [安全管理](./core/security/README.md)
- [x] [服务管理](./core/server/README.md)
- [x] 添加循环调用拦截器
- [x] 熔断拦截器
- [x] rest请求头注入到rpc上下文
- [x] accesslog拦截器按错误级别输出日志
- [x] rest添加metrics接口
- [x] rest添加swagger
- [x] 所有协议添加health接口
- [x] server new方法使用options方式传参
- [x] protoc-gen-rest-go支持自定义api类型和version(api: "api", version:"v1")
- [x] 修复文件配置源更新事件问题
- [x] 所有配置添加默认配置，在不配置的情况下也能正常运行
- [x] 监控
- [x] rest gateway
- [x] 添加grafana看板
- [x] api添加i18n拦截器
- [x] client支持指定实例直连
- [x] 支持mysql连接
- [x] 支持etcd连接
- [x] 添加etcd服务发现注册中心
- [x] 支持redis连接
- [ ] 限速
- [ ] 链路追踪
- [ ] 拦截器配置自动更新，无需重启
- [ ] stream支持
- [ ] openapi更新default response
- [ ] 添加rest服务返回自定义拦截器
- [ ] 添加测试用例，文档，cli工具
- [ ] 添加远程配置中心
- [ ] 配置监听添加方法监听
- [ ] 修复文件配置源同一个配置在不同配置文件中优先级问题
- [ ] protoc-gen-ts实现
- [ ] access_log支持和主日志分不同文件存放
- [ ] 支持mongo连接
- [ ] 不同服务，方法，支持指定负载均衡策略，从指定服务发现中心发现服务

## 背景

## 特性

- []

## 快速开始

> 更多示例请参考[examples](./examples)

### 编写proto协议文件

```proto
syntax = "proto3";

package api.v1.hello;

option go_package = "github.com/asjard/asjard/examples/protobuf/hello;hello";

import "github.com/asjard/protobuf/http.proto";

// 服务注释
service Hello {
    // 注释
    // 后续swagger中会被使用到
    rpc Say(SayReq) returns (SayReq) {
        // rest路由配置
        // 可以有多个
        option (asjard.api.http) = {
            get : "/v1"
        };
        option (asjard.api.http) = {
            post : "/v1/region/{region_id}/project/{project_id}/user/{user_id}"
        };
    };
}
message SayReq {
    string          region_id  = 1;
    string          project_id = 2;
    int64           user_id    = 3;
    repeated string str_list   = 4;
    repeated int64  int_list   = 5;
    SayObj          obj        = 6;
    repeated SayObj objs       = 7;
}

message SayObj {
    int32  field_int = 1;
    string field_str = 2;
}
```

### 按需生成

```sh
protoc --go_out=${GOPATH}/src -I${GOPATH}/src -I. ./*.proto

# 生成grpc需要的文件
protoc --go-grpc_out=${GOPATH}/src -I${GOPATH}/src -I. ./*.proto

# 生成rest需要的文件, rest依赖grpc生成的文件
protoc --go-rest_out=${GOPATH}/src -I${GOPATH}/src -I. ./*.proto
```

### 编写服务

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/asjard/asjard"
	"github.com/asjard/asjard/core/client"
	"github.com/asjard/asjard/core/config"
	"github.com/asjard/asjard/core/logger"
	pb "github.com/asjard/asjard/examples/protobuf/hello"
	_ "github.com/asjard/asjard/pkg/client/grpc"
	mgrpc "github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// Hello 同一个方法既可以当做GRPC的handler，也可以当做http的handler
type Hello struct {
	pb.UnimplementedHelloServer
	conn pb.HelloClient
}

var _ pb.HelloServer = &Hello{}

// Bootstrap grpc客户端初始化
func (c *Hello) Bootstrap() error {
	conn, err := client.NewClient(mgrpc.Protocol, "helloGrpc").Conn()
	if err != nil {
		return err
	}
	c.conn = pb.NewHelloClient(conn)
	return nil
}

// Shutdown
func (c *Hello) Shutdown() {}

// Say .
func (c *Hello) Say(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	resp, err := c.conn.Call(ctx, in)
	if err != nil {
		return nil, err
	}
	return resp, err
}

// Call .
func (c *Hello) Call(ctx context.Context, in *pb.SayReq) (*pb.SayReq, error) {
	in.RegionId = "timeout: " + config.GetString("sleep", "")
	return in, nil
}

// RestServiceDesc rest服务描述, 如果提供rest服务，则必须提供此方法
func (Hello) RestServiceDesc() *rest.ServiceDesc {
	return &pb.HelloRestServiceDesc
}

// GrpcServiceDesc grpc服务描述, 如果提供grpc服务, 则必须提供此方法
func (Hello) GrpcServiceDesc() *grpc.ServiceDesc {
	return &pb.Hello_ServiceDesc
}

func main() {
	server := asjard.New()
	// 添加rest和grpc服务
	server.AddHandler(rest.Protocol, &Hello{}, rest.Protocol, mgrpc.Protocol)
	if err := server.Start(); err != nil {
		panic(err)
	}
}

```

## [变更日志](CHANGELOG.md)
