## 服务描述

详细示例参考[这里](https://github.com/asjard/asjard/blob/develop/_examples/svc-example/apis/api/conf/service.yaml)

```yaml
asjard:
  service:
    ## 项目名称
    ## 一个项目（Project）下可以包含多个不同的微服务
    # app: asjard

    ## 当前部署环境
    ## 例如：dev (开发), sit (集成测试), uat (验收测试), rc (预发布), pro (生产)
    ## 如果连接了注册中心，该环境标识通常用于隔离不同阶段的服务实例
    # environment: "dev"

    ## 部署地域 (Region)
    ## 例如：cn-shanghai (华东), cn-beijing (华北)
    ## 代表地理位置上的隔离，不同地域间通常内网不互通，需走公网连接
    # region: "default"

    ## 可用区 (Availability Zone)
    ## 例如：az-1, az-2
    ## 指同一地域下的不同机房，设备间内网互通，用于实现同城容灾
    # avaliablezone: "default"

    # website: "https://github.com/${asjard.service.app}/${asjard.service.instance.name}"
    # favicon: "favicon.ico"

    ## 服务描述
    ## 支持 Markdown 格式，用于在管理后台展示服务详情
    desc: |
      这里是服务描述，支持 Markdown 格式

    instance:
      ## 系统唯一识别码
      ## 数字格式 (100-999)，常用于生成内部全局错误码
      # systemCode: 100

      ## 是否可跨组织共享
      ## 设置为 true 时，表示该服务可以被不同组织单元或跨 AZ 的服务发现并调用
      # shareable: false

      ## 实例名称 (Name)
      ## 指具体的部署入口名称（如 "svc-example-api", "svc-example-openapi"）
      ## 用于实现精细化的流量调度和服务发现
      name: svc-example-api

      ## 服务组/逻辑服务 (Group)
      ## 标识逻辑上的服务主体（如 "svc-example"）
      ## 多个不同的入口 (Name) 可以共享同一个 Group，表示它们属于同一套代码逻辑和业务范畴
      group: svc-example

      ## 服务版本
      ## 建议遵循语义化版本规范 (例如 "1.2.3")
      # version: 1.0.0

      ## 自定义元数据
      ## 用于存储额外的键值对信息，支持自定义的服务发现过滤逻辑
      # metadata:
```
