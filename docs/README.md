---
title: 快速开始
---

## 安装protoc插件

> 按需安装, 框架自动生成代码的命令可在`cmd/`目录下查看

```bash
## 生成.pb.go文件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
## 生成_grpc.pb.go文件，grpc服务端，客户端
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
## 生成_amqp.pb.go文件，rabbitmq服务端，客户端
go install github.com/asjard/asjard/cmd/protoc-gen-go-amqp@latest
## 生成_asynq.pb.go文件, asynq服务端，客户端
go install github.com/asjard/asjard/cmd/protoc-gen-go-asynq@latest
## 生成_rest.pb.go文件, http服务端
go install github.com/asjard/asjard/cmd/protoc-gen-go-rest@latest
## 生成_rest_gw.pb.go文件, http协议转grpc协议
go install github.com/asjard/asjard/cmd/protoc-gen-go-rest2grpc-gw@latest
## 生成_validate.pb.go文件,参数校验
go install github.com/asjard/asjard/cmd/protoc-gen-go-validate@latest
## 生成enum.pb.ts文件，typescript枚举生成
go install github.com/asjard/asjard/cmd/protoc-gen-ts-enum@latest
## 生成umi.pb.ts文件, umi request请求生成
go install github.com/asjard/asjard/cmd/protoc-gen-ts-umi@latest
## 生成pb.ts文件, typescript类型定义
go install github.com/asjard/asjard/cmd/protoc-gen-ts@latest
```

## 定义protobuf文件

```proto
syntax = "proto3";

package api.v1.example.docs;

option go_package = "protos-repo/example/api/v1/sample";

import "github.com/asjard/protobuf/http.proto";
import "github.com/asjard/protobuf/validate.proto";

service Sample {
  rpc SayHello(HelloRequest) returns (HelloReply) {
    option (asjard.api.http) = {
      get : "/helloworld/{name}"
    };
    option (asjard.api.http) = {
      get : '/hello'
    };
  }
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1 [ (asjard.api.validate).rules = "required,max=20" ];
}

// The response message containing the greetings
message HelloReply { string message = 1; }

```

通过协议文件生成golang代码

```bash
protoc --go_out=. --go-grpc_out=. --go-rest_out=. -I${GOPATH}/src -I. sample.proto
```

## 新增服务

```go
package main

import (
	"context"

	pb "github.com/asjard/asjard/_examples/protos-repo/example/api/v1/sample"
	"github.com/asjard/asjard/pkg/server/grpc"
	"github.com/asjard/asjard/pkg/server/rest"
	"github.com/asjard/asjard"
)

type SampleAPI struct {
	pb.UnimplementedSampleServer
}

func NewSampleAPI() *SampleAPI {
	return &SampleAPI{}
}

func (api *SampleAPI) Start() error                       { return nil }
func (api *SampleAPI) Stop()                              {}

// GRPC服务
func (api *SampleAPI) GrpcServiceDesc() *grpc.ServiceDesc { return &pb.Sample_ServiceDesc }

// HTTP服务
func (api *SampleAPI) RestServiceDesc() *rest.ServiceDesc { return &pb.SampleRestServiceDesc }

func (api *SampleAPI) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: "hello " + in.Name,
	}, nil
}

func main() {
	server := asjard.New()
	if err := server.AddHandler(NewSampleAPI(), grpc.Protocol, rest.Protocol); err != nil {
		log.Fatal(err)
	}
	log.Fatal(server.Start())
}
```

## 添加配置文件

```yaml
asjard.service.instance.systemCode: 100
asjard.service.instance.name: svc-example-api
asjard:
  servers:
    grpc:
      enabled: true
      addresses:
        listen: 0.0.0.0:9${asjard.service.instance.systemCode}
    rest:
      enabled: true
      addresses:
        listen: 0.0.0.0:8${asjard.service.instance.systemCode}
```

## 启动服务

```bash
ASJARD_CONF_DIR="$(PWD)/conf"  go run ./main.go
```

启动成功后将出现类似如下信息:

```bash

                                       App:      example
                                       Env:      local
    _   ___    _  _   ___ ___          Region:   default
   /_\ / __|_ | |/_\ | _ \   \          Az:       default
  / _ \\__ \ || / _ \|   / |) |
 /_/ \_\___/\__/_/ \_\_|_\___/ 0.8.8
                                       ID:       91183721-3157-4798-be2b-ce087b958f41
                                       Name:     svc-example-api
                                       Version:  1.0.0
                                       Servers:  grpc://0.0.0.0:9100;rest://0.0.0.0:8100
                                       ConfDir:  ./conf
```

## 测试

### 成功

```bash
curl -s  -X GET 'http://localhost:8100/api/v1/docs/samples/helloworld/john'|python3 -m json.tool
## 或者
curl -s  -X GET 'http://localhost:8100/api/v1/docs/samples/hello?name=john'|python3 -m json.tool
```

你将得到类似如下的结果

```json
{
  "code": 0,
  "err_code": 0,
  "status": 200,
  "system": 0,
  "success": true,
  "message": "",
  "prompt": "",
  "doc": "",
  "request_id": "1d2e0f7e7668f86d9def719777ebb78f",
  "request_method": "/api.v1.example.docs.Sample/SayHello",
  "data": {
    "@type": "type.googleapis.com/api.v1.example.docs.HelloReply",
    "message": "hello john"
  }
}
```

### 失败

```bash
curl -s  -X GET 'http://localhost:8100/api/v1/docs/samples/helloworld'|python3 -m json.tool
```

你将得到类似如下的结果

```json
{
  "code": 1004045,
  "err_code": 5,
  "status": 404,
  "system": 100,
  "success": false,
  "message": "page not found",
  "prompt": "",
  "doc": "",
  "request_id": "9e4e8910729daa5f54225388dd2bfd47",
  "request_method": "/asjard.api.ErrorHandler/NotFound",
  "data": null
}
```

如上示例可在`github.com/asjard/asjard/_examples/svc-example/apis/api/v1/sample.go`和`github.com/asjard/asjard/_examples/protos-repo/examples/api/v1/sample.proto`中查看

通过以上示例你启动了一个最简单的包含HTTP和GRPC的服务，后续我将通过配置为你开启各种功能，比如`cors`,`openAPI`,`参数校验`,`trace`等功能
