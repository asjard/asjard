## 环境变量配置源

- 框架配置都会以`asjard`为前缀
- 不同层级的配置中间以`_`分隔, 例如`asjard_app`, 程序使用`asjard.app`读取
- 大小写敏感, 例如`asjard_app`和`asjard_APP`为两个不同的配置i

```go
// 在环境变量中配置如下配置
// export asjard_app=asjard
// 程序中可以这样读
config.GetString("asjard.app", "")
// Output: asjard
```
