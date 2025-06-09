# Go-Protoc 项目

## 项目结构

```
.
├── api/            # API 定义和生成的代码
├── cmd/            # 应用程序入口
├── configs/        # 配置文件
├── docs/          # 文档
│   ├── errors/    # 错误处理文档
│   └── middleware/ # 中间件文档
├── pkg/           # 项目包
│   ├── api/       # API 实现
│   ├── errors/    # 错误处理
│   └── middleware/ # 中间件
└── scripts/       # 脚本文件
```

## 功能特性

- gRPC 和 HTTP 双协议支持
- Swagger/OpenAPI 文档
- 错误处理系统
- 中间件支持
  - 日志记录
  - 请求追踪
  - 错误恢复
  - 超时控制
  - 可观察性（白名单）
- Prometheus 指标
- 健康检查

## 快速开始

### 安装依赖

```bash
make install-tools
```

### 生成 API 代码

```bash
make proto
```

### 运行服务

```bash
go run cmd/apiserver/main.go
```

## API 文档

- Swagger UI: http://localhost:8090/swagger/index.html
- OpenAPI 规范: http://localhost:8090/swagger/doc.json

## 监控和调试

- Prometheus 指标: http://localhost:8090/metrics
- pprof 调试: http://localhost:8090/debug/pprof/
- 健康检查: http://localhost:8090/healthz

## 中间件

### 可观察性中间件

项目包含了一套完整的可观察性中间件系统，支持请求日志记录、监控指标收集和错误追踪。为了避免系统路径产生不必要的日志和指标，实现了白名单机制。

#### 默认白名单路径

以下路径默认不会被记录：

- `/metrics` - Prometheus 指标
- `/debug/` - Debug 端点
- `/swagger/` - Swagger UI
- `/healthz` - 健康检查
- `/favicon.ico` - 浏览器图标请求

详细文档请参考：[可观察性中间件配置指南](docs/middleware/observability.md)

### 错误处理

项目实现了统一的错误处理系统，包括：

- 标准错误码
- 错误响应格式
- 错误类型判断
- 错误详情支持

详细文档请参考：[错误处理文档](docs/errors/README.md)

## 配置

配置文件位于 `configs/` 目录下，支持：

- YAML 配置文件
- 环境变量覆盖
- 命令行参数

## 开发指南

### 添加新的 API

1. 在 `pkg/api/` 目录下创建新的 proto 文件
2. 运行 `make proto` 生成代码
3. 实现服务接口
4. 在 `cmd/apiserver/main.go` 中注册服务

### 自定义中间件配置

可以通过创建自定义配置来控制中间件行为：

```go
customConfig := &config.ObservabilityConfig{
    SkipPaths: []string{
        "/metrics",
        "/custom/path",
        // 添加更多路径
    },
}

router.Use(http.LoggingMiddlewareWithConfig(logger, customConfig))
```

## 贡献指南

1. Fork 项目
2. 创建特性分支
3. 提交变更
4. 推送到分支
5. 创建 Pull Request

## 许可证

[MIT License](LICENSE)
