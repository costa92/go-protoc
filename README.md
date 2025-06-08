# Go-Protoc 服务框架

一个基于gRPC和HTTP Gateway的Go微服务框架，支持多版本API、配置管理和可观测性。

## 主要特性

- **多版本API支持**：同时支持多个API版本，便于API演进
- **双协议支持**：通过gRPC-Gateway支持gRPC和HTTP两种协议
- **配置管理**：使用Viper支持从文件和环境变量加载配置
- **可观测性**：
  - 集成OpenTelemetry实现分布式追踪
  - 支持Prometheus指标监控
  - 结构化日志记录
- **开发工具**：
  - 完整的Makefile命令
  - 代码质量检查工具
  - Swagger API文档

## 目录结构

```
├── cmd/            # 各个主程序（可执行文件）入口
├── pkg/            # 可被外部项目引用的库代码
├── internal/       # 仅限本项目内部使用的代码
├── api/            # API 定义（如 Protobuf、OpenAPI 等）
├── configs/        # 配置文件
├── scripts/        # 各类运维脚本
├── build/          # 打包与持续集成相关文件
├── deployments/    # 部署相关文件（如 Docker、K8s）
├── test/           # 额外的外部测试代码
├── go.mod
├── go.sum
└── README.md
```

## 安装与配置

### 依赖工具

- Go 1.18+
- Protocol Buffers 编译器 (protoc)
- 相关插件：
  - `protoc-gen-go`
  - `protoc-gen-go-grpc`
  - `protoc-gen-grpc-gateway`

### 安装工具

```bash
# 安装开发所需工具
make install-tools
```

### 配置管理

配置文件位于 `configs/config.yaml`，也可通过环境变量覆盖配置：

```bash
# 通过环境变量配置HTTP端口
export GO_PROTOC_SERVER_HTTP_ADDR=:8080

# 指定配置文件路径
export CONFIG_PATH=/path/to/config.yaml
```

## 构建与运行

### 构建服务

```bash
# 生成Protocol Buffers代码
make proto

# 构建服务
go build -o bin/apiserver cmd/apiserver/main.go
```

### 运行服务

```bash
# 直接运行
./bin/apiserver

# 或者使用go run
go run cmd/apiserver/main.go
```

### API访问

- gRPC服务默认运行在`:8091`
- HTTP服务默认运行在`:8090`
- API文档访问：`http://localhost:8090/swagger/index.html`
- Prometheus指标：`http://localhost:8090/metrics`

## 开发指南

### 添加新API

1. 在`api/`目录下创建Protobuf定义文件
2. 运行`make proto`生成代码
3. 在`internal/`中实现服务逻辑
4. 在`internal/{service}/installer.go`中注册服务

### 单元测试与代码质量

```bash
# 运行测试
make test

# 检查代码覆盖率
make test-coverage

# 代码格式化
make fmt

# 代码静态检查
make vet

# 代码质量分析
make lint
```

## 可观测性

### 链路追踪

服务默认集成了OpenTelemetry，支持以下导出器：

- `stdout`: 输出到标准输出（默认）
- `jaeger`: 输出到Jaeger
- `otlp`: 输出到OpenTelemetry Collector

配置示例：

```yaml
observability:
  tracing:
    service_name: "my-service"
    enabled: true
    exporter: "jaeger"
```

### 指标监控

服务暴露Prometheus格式的指标，包括：

- HTTP请求计数和耗时
- gRPC请求计数和耗时

访问`/metrics`端点获取指标数据。

## 许可证

本项目采用MIT许可证。
