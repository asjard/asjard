## accessLog

支持协议: all

```yaml
asjard:
  logger:
    accessLog:
      ## 是否开启accessLog
      enabled: true
      ## 配置格式: [protocol://]{fullMethod}
      ## 例如grpc协议的某个方法: grpc:///api.v1.hello.Hello/Call
      ## 或者协议无关的某个方法: /api.v1.hello.Hello/Call
      ## 拦截协议的所有方法: grpc
      skipMethods:
        - grpc
```

## i18n

支持协议: rest

国际化配置存放路径, 优先读取`ASJARD_I18N_DIR`环境变量
如果为空读取`ASJARD_HOME_DIR`
如果为空则为`可执行程序`所在路径的`locals`目录

文件格式为`{lang}.json`， 其中`lang`为语言

文件内容格式为:

```json
{
  "123": {
    "prompt": "错误提示"
    "doc": "可以自行处理这个错误的文档地址"
  }
}
```

其中`123`为错误码
