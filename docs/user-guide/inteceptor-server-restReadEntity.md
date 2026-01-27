## 拦截器名称

restReadEntity

## 支持协议

- HTTP

## 功能

- 将http请求参数解析到protobuf定义的message中
- 解析优先级(高优先级的值会覆盖低优先级的值),从左到右优先级依次增高
  - query < header < body < path
- GET,DELETE,CONNECT,OPTIONS,HEAD,TRACE请求不解析body体

## 配置

无配置
