## 已实现的protoc插件

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
