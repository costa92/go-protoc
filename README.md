# Go-Protoc 项目

基于 gRPC 和 HTTP Gateway 的 Go 微服务框架，支持多版本 API、统一错误处理、中间件系统和完整的可观测性功能。

## 项目结构

```bash
.
├── api/                    # API 定义和生成的代码
│   ├── openapi/           # OpenAPI/Swagger 文档
│   └── proto/             # Protocol Buffers 定义
├── cmd/                    # 应用程序入口
│   └── apiserver/         # API 服务器
├── configs/               # 配置文件
│   └── config.yaml        # 主配置文件
├── docs/                  # 文档
│   ├── errors/           # 错误处理文档
│   └── middleware/       # 中间件文档
├── pkg/                   # 项目包
│   ├── api/              # API 实现
│   │   └── helloworld/   # Hello World 服务示例
│   ├── errors/           # 错误处理
│   └── middleware/       # 中间件
│       ├── config/       # 中间件配置
│       └── http/         # HTTP 中间件
└── scripts/              # 脚本文件
    └── make-rules/       # Makefile 规则
```

## 功能特性

- **API 支持**

  - gRPC 和 HTTP 双协议支持
  - API 版本管理
  - Swagger/OpenAPI 文档自动生成
  - 请求参数验证

- **错误处理**

  - 统一错误码系统
  - 结构化错误响应
  - 错误类型判断
  - 详细错误信息

- **中间件系统**

  - 日志记录（支持结构化日志）
  - 分布式追踪（集成 OpenTelemetry）
  - 错误恢复（panic 处理）
  - 超时控制
  - 可观察性（支持白名单）

- **监控和指标**
  - Prometheus 指标收集
  - 健康检查
  - pprof 调试支持

## 环境要求

- Go 1.18+
- Protocol Buffers v3
- Make

## 快速开始

### 1. 安装依赖工具

```bash
# 安装所有必需的工具
make install-tools
```

这将安装：

- protoc-gen-go
- protoc-gen-go-grpc
- protoc-gen-grpc-gateway
- protoc-gen-openapiv2
- protoc-gen-validate
- 其他开发工具

### 2. 生成 API 代码

```bash
# 生成 Protocol Buffers 代码和 OpenAPI 文档
make proto
```

### 3. 运行服务

```bash
# 直接运行
go run cmd/apiserver/main.go

# 或者构建后运行
make build
./bin/apiserver
```

## 服务端点

### API 服务

- gRPC: `:8091`
- HTTP: `:8090`

### API 文档

- Swagger UI: <http://localhost:8090/swagger/index.html>
- OpenAPI 规范: <http://localhost:8090/swagger/doc.json>

### 监控和调试

- Prometheus 指标: <http://localhost:8090/metrics>
- pprof 调试: <http://localhost:8090/debug/pprof/>
- 健康检查: <http://localhost:8090/healthz>

## 配置说明

### 配置文件

配置文件位于 `configs/config.yaml`，支持：

- YAML 配置文件
- 环境变量覆盖
- 命令行参数

### 主要配置项

```yaml
server:
  http:
    addr: :8090
  grpc:
    addr: :8091

log:
  level: info
  format: json

middleware:
  timeout: 30s
  observability:
    skip_paths:
      - /metrics
      - /debug/
      - /swagger/
      - /healthz
```

### 环境变量

可以使用环境变量覆盖配置：

```bash
export SERVER_HTTP_ADDR=:8080
export SERVER_GRPC_ADDR=:8081
export LOG_LEVEL=debug
```

## 开发指南

### 添加新的 API

1. 在 `pkg/api/` 目录下创建新的 proto 文件：

```protobuf
syntax = "proto3";
package myservice.v1;
option go_package = "github.com/costa92/go-protoc/pkg/api/myservice/v1;myservicev1";
```

2. 运行代码生成：

```bash
make proto
```

3. 实现服务接口：

```go
type MyServiceServer struct {
    myservicev1.UnimplementedMyServiceServer
}
```

4. 在 `cmd/apiserver/main.go` 中注册服务

### 自定义中间件配置

```go
// 创建自定义配置
customConfig := &config.ObservabilityConfig{
    SkipPaths: []string{
        "/metrics",
        "/custom/path",
    },
}

// 使用自定义配置
router.Use(http.LoggingMiddlewareWithConfig(logger, customConfig))
```

### 错误处理

```go
// 使用预定义错误
if user == nil {
    return errors.ErrUserNotFound
}

// 创建自定义错误
err := errors.NewError(10100, "自定义错误消息")
    .WithDetails(map[string]string{
        "field": "username",
        "error": "用户名已存在",
    })
```

## 可用的 Make 命令

- `make install-tools`: 安装开发工具
- `make proto`: 生成 Protocol Buffers 代码
- `make build`: 构建项目
- `make test`: 运行测试
- `make lint`: 运行代码检查
- `make clean`: 清理生成的文件

## 文档

- [错误处理文档](docs/errors/README.md)
- [可观察性中间件配置指南](docs/middleware/observability.md)

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交变更 (`git commit -m 'feat: add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 许可证

[MIT License](LICENSE)
