> protobuf编写及约定

详细示例参考[这里](https://github.com/asjard/examples/blob/develop/_examples/protos-repo)

### 规范和约定

- protbuf规范参考[这里](https://protobuf.dev/programming-guides/proto3/)
- protobuf中使用添加openapi文档参考[这里](https://github.com/google/gnostic/tree/main/openapiv3), `protoc-gen-rest`命令中已集成

```proto
syntax = "proto3";

// 约定此处格式为: {接口类型}.{接口版本}.{服务名称}[.服务分类]
// 其中{接口类型},{接口版本},[服务分类]会在rest服务中用来生成路由前缀
// 可在ajard.api.http中通过api,和version字段修改
// 例如如下package名称生成的路由前缀为/api/v1/docs
//
// {服务名称}会在生成gw代码时用以确定客户端名称, 例如:
// conn, err := client.NewClient(grpc.Protocol, config.GetString("asjard.topology.services.{hello}.name", "{hello}")).Conn()
// 如果修改了asjard.service.instance.name 则可以通过asjard.topology.service.hello.name修改
package api.v1.hello;

option go_package = "github.com/asjard/asjard/examples/example/hellopb";

import "github.com/asjard/protobuf/http.proto";
import "github.com/asjard/protobuf/validate.proto";
import "github.com/asjard/protobuf/mq.proto";

// 需要实现的功能
// 一个proto文件中只写一个service
// 如果在rest中没有配置group字段，则service名称为group字段的值
service Hello {
  // 可以对整个服务定义路由信息
  // option(asjard.api.serviceHttp) = {api: "api", versin: "v1", group:"hello", writer_name: "custome_writer"}
  // 功能描述,
  // 支持markdown
  // 可渲染在openapi文档的接口描述中
  // 渲染在rest服务的路由描述中
  rpc Say(SayReq) returns (SayResp) {
    // 如果是要对外暴露rest服务则写如下路由信息
    // 可以有多条路由信息
    option (asjard.api.http) = {
      // key为请求方式, 支持 get, put, post, delete, patch, header
      // value为请求路径, 相对路径
      // 完整路径为 接口类型 + / + 接口版本 + / + 接口分组 + / + 请求路径
      // GET /api/v1/hello
      get : "/"
      // 如果不为空则使用此处的接口分类
      api : ""
      // 如果不为空则使用此处的接口版本
      version : ""
      // 如果不为空则使用此处的接口分组, 都为空则为service复数名称
      group : ""
      // 当前接口自定义writer
      writer_name: ""
    };
    option (asjard.api.http) = {
      // POST /api/v1/hello
      post : "/"
    };
    option (asjard.api.http) = {
      // DELETE /api/v1/hello/{name}
      delete : "/{name}"
    };
    option (asjard.api.http) = {
      // PUT /api/v1/hello/{name}
      put: "/{name}"
    };
    option (asjard.api.http) = {
      // GET /openapi/v2/groupName/hello/{name}
      get: "/hello/{name}"
      group: "groupName"
      api: "openapi"
      version: "v2"
    };
    // 提供MQ服务
    option(asjard.api.mq) = {};
  };
}

// 请求参数
message SayReq {
  // 字段描述
  // 会渲染到openapi的字段描述中
  string name = 1 [(asjard.api.validate).rules="required,max=20"];
}

// 请求返回
message SayResp { string message = 2; }
```
