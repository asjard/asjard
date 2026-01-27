‌# Summary​

## 规范

- [protobuf规范](user-guide/standard-protobuf.md)
- [目录规范](user-guide/standard-service.md)
- [错误规范](user-guide/standard-error.md)

## 组件

- [启动引导](user-guide/bootstrap.md)
- [日志](user-guide/logger.md)
- [拦截器]
  - [客户端拦截器](user-guide/interceptor-client.md)
    - [熔断降级](user-guide/interceptor-client-circuit-breaker.md)
    - [循环调用检测](user-guide/interceptor-client-cycle-chain.md)
    - [请求错误日志](user-guide/interceptor-client-errlog.md)
    - [HTTP请求头转GRPC上下文](user-guide/inteceptor-client-rest2grpc.md)
    - [慢日志](user-guide/inteceptor-client-slowlog.md)
    - [请求参数校验](user-guide/inteceptor-client-validate.md)
    - [panic日志](user-guide/inteceptor-client-panic.md)

  - [服务端拦截器](user-guide/interceptor-server.md)
    - [accessLog](user-guide/interceptor-server-accessLog.md)
    - [i18n](user-guide/interceptor-server-i18n.md)
    - [监控](user-guide/interceptor-server-metrics.md)
    - [panic日志](user-guide/inteceptor-server-panic.md)
    - [限速](user-guide/inteceptor-server-ratelimit.md)
    - [请求参数解析](user-guide/inteceptor-server-restReadEntity.md)
    - [链路追踪](user-guide/inteceptor-server-trace.md)
    - [参数校验](user-guide/inteceptor-server-validate.md)

- [服务发现&注册](user-guide/registry.md)
  - [consul](user-guide/registry-consule.md)
  - [etcd](user-guide/registry-etcd.md)
  - [local](user-guide/registry-local.md)
- [客户端负载均衡](user-guide/balance.md)
  - [本地优先负载均衡](user-guide/balance-locality.md)
  - [轮询](user-guide/balance-roundrobin.md)
- [配置](user-guide/config.md)
  - [consul](user-guide/config-consul.md)
  - [env](user-guide/config-env.md)
  - [etcd](user-guide/config-etcd.md)
  - [file](user-guide/config-file.md)
- [存储](user-guide/stores.md)
  - [asynq](user-guide/stores-asynq.md)
  - [consul](user-guide/stores-consul.md)
  - [etcd](user-guide/stores-etcd.md)
  - [gorm](user-guide/stores-gorm.md)
  - [rabbitmq](user-guide/stores-rabbitmq.md)
  - [redis](user-guide/stores-redis.md)
- [缓存](user-guide/cache.md)
  - [redis](user-guide/cache-redis.md)
  - [local](user-guide/cache-loacal.md)
- [服务/协议](user-guide/server.md)
  - [grpc](user-guide/server-grpc.md)
  - [http](user-guide/server-rest.md)
  - [asynq](user-guide/server-asynq.md)
  - [rabbitmq](user-guide/server-rabbitmq.md)
- [客户端](user-guide/client.md)
  - [grpc](user-guide/cient-grpc.md)
- [其他](user-guide/other.md)
  - [分布式锁](user-guide/other-mutex.md)
    - [redis](user-guide/other-mutex-redis.md)
    - [mysql](user-guide/other-mutex-mysql.md)
  - [安全](user-guide/other-security.md)
  - [监控指标](user-guide/other-metrics.md)

## 性能

## 运维

- [Makefile]
- [代码生成]
