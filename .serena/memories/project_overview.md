# 项目概述

## 项目目的
这是一个基于 Go 语言的微服务框架项目，集成了多种常用组件和工具。主要特点：
- 使用 Protocol Buffers 进行 API 定义
- 支持 gRPC 和 HTTP 服务
- 集成了多种注册中心（Consul、etcd）
- 提供了完整的监控、追踪和日志功能

## 技术栈
- 框架：Kratos v2
- API：gRPC + HTTP (grpc-gateway)
- 数据库：MySQL、PostgreSQL、MongoDB、Redis
- 消息队列：Kafka
- 注册中心：Consul、etcd
- 监控：Prometheus
- 追踪：OpenTelemetry
- 日志：Zap
- 依赖注入：Wire
- 配置管理：Viper
- CLI：Cobra

## 项目结构
- api/: API 定义和生成的代码
- cmd/: 主程序入口
- configs/: 配置文件
- internal/: 内部实现代码
- pkg/: 可重用的包
- scripts/: 构建和工具脚本
- third_party/: 第三方依赖
- tools/: 开发工具