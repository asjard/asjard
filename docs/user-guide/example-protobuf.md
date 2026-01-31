详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/protos-repo)

## 定义protobuf协议文件

```proto
syntax = "proto3";

package api.v1.example;

option go_package = "protos-repo/example/api/v1/example";

import "github.com/asjard/protobuf/http.proto";
import "github.com/asjard/protobuf/validate.proto";

service Example {
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

## 生成代码

```bash
make gen_proto
```
