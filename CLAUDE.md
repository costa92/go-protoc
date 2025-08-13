# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码库中工作提供指导。

## 项目概览

基于 Kratos v2 构建的生产就绪 Go 微服务框架，以 Protocol Buffers 为统一数据源。采用清洁架构与 Wire 依赖注入，支持 HTTP/gRPC API，具备完整的错误处理、国际化、认证和可观测性功能。

## 快速开发命令

### 日常必备命令
- `make help` - 查看所有可用命令
- `make run-api` - 启动开发服务器（支持热重载）
- `make build` - 构建优化二进制文件到 ./bin/apiserver
- `make test` - 运行所有测试（详细输出）
- `make fmt` - 格式化代码和导入排序
- `make tidy` - 清理 go.mod 依赖

### 开发环境
- `make dev-setup` - 完整开发环境设置（工具 + 服务 + 代码生成）
- `make dev-clean` - 清理开发环境并停止服务
- `make dev-quick` - 快速开发启动（跳过工具安装）
- `make dev-watch` - 文件监控与自动重构建
- `make dev-test` - 完整测试流水线（含覆盖率）
- `make dev-bench` - 运行性能基准测试

### 代码生成
- `make generate` - 使用 buf 生成 protobuf/gRPC/HTTP 代码
- `make wire` - 重新生成依赖注入代码（结构变更后运行）
- `buf generate` - 直接 protobuf 生成（buf.yaml 变更时使用）

### 工具安装
- `make install-tools` - 仅安装 CI 相关工具
- `make install-tools A=1` - 安装所有开发工具
- `make tools.install.<tool>` - 安装特定工具（wire, golangci-lint, buf 等）

### 统一服务管理
项目现已实现统一的服务管理机制，支持通过 docker-compose 和安装脚本管理所有第三方服务。

#### 数据库服务
- `make run-redis` / `make stop-redis` - Redis 缓存服务
- `make run-mariadb` / `make stop-mariadb` - MariaDB 数据库服务  
- `make run-mongodb` / `make stop-mongodb` - MongoDB 文档数据库

#### 消息队列服务
- `make run-kafka` / `make stop-kafka` - Kafka 消息服务

#### 分布式服务
- `make run-etcd` / `make stop-etcd` - etcd 分布式键值存储

#### 可观测性服务
- `make run-jaeger` / `make stop-jaeger` - Jaeger 链路追踪
- `make run-prometheus` / `make stop-prometheus` - Prometheus 监控
- `make run-grafana` / `make stop-grafana` - Grafana 仪表板
- `make run-alertmanager` / `make stop-alertmanager` - AlertManager 告警
- `make run-otelcol` / `make stop-otelcol` - OpenTelemetry 收集器
- `make run-victorialogs` / `make stop-victorialogs` - VictoriaLogs 日志管理

#### 服务组管理
- `make start-all` / `make stop-all` - 所有服务
- `make restart-all` - 重启所有服务
- `make start-database` / `make stop-database` - 数据库服务组
- `make start-observability` / `make stop-observability` - 可观测性服务组
- `make status-all` - 检查所有服务状态
- `make logs-all` - 查看所有服务日志

#### 直接脚本调用
```bash
# 服务管理脚本支持更多操作
./scripts/installation/service.sh <action> <service|group>

# 可用操作: start, stop, restart, status, logs
# 可用服务: redis, mariadb, mongodb, kafka, etcd, jaeger, prometheus, grafana, alertmanager, otelcol, victorialogs
# 可用服务组: all, database, observability
```

### 版本管理
所有第三方组件版本统一管理在 `scripts/installation/versions.sh` 中：

```bash
# 查看所有组件版本
./scripts/installation/versions.sh show

# 验证版本格式
./scripts/installation/versions.sh validate
```

当前管理的组件版本：
- **数据库**: Redis 7.2.4, MariaDB 11.2.2, MongoDB 7.0.5
- **分布式**: etcd v3.5.12, Kafka 3.6.1  
- **可观测性**: Jaeger 1.52.0, Prometheus 2.48.1, Grafana 10.2.4, AlertManager 0.26.0, OpenTelemetry Collector 0.91.0
- **日志**: VictoriaLogs v0.5.2-victorialogs

### 文档生成
- `make docs.dev` - 生成开发文档到 docs/DEVELOPMENT.md
- `make docs.api` - 从 protobuf 生成 API 文档
- `make docs.serve` - 本地提供文档服务

## 架构概览

### 清洁架构分层
```
cmd/apiserver/          → 入口点和编排
 internal/apiserver/     → 核心业务逻辑
 ├── handler/           → HTTP/gRPC 处理器（交付层）
 ├── biz/              → 用例（应用层）
 ├── store/            → 数据访问（基础设施层）
 └── config/           → 内部配置

pkg/                  → 可重用包（对外导出）
 ├── api/              → Protobuf 定义和生成代码
 ├── errorsx/          → 上下文感知错误系统（支持 i18n）
 ├── authn/           → JWT 身份认证工具
 ├── db/              → 数据库抽象
 ├── server/          → HTTP/gRPC 服务器配置
 └── options/         → 组件配置架构
```

### 依赖注入 (Wire)
- **中心化在** `internal/apiserver/wire.go`
- **生成的工厂** `internal/apiserver/wire_gen.go`
- **自动发现** - 添加新依赖后运行 `make wire`

### 请求流程
```
HTTP 请求 → gRPC-Gateway → 处理器 → 业务层 → 存储层 → 数据库
     ↓                                          ↓
  OpenAPI 文档（自动生成）        GORM + 上下文事务
```

## 配置和环境

### 本地设置
1. 复制并编辑：`cp configs/apiserver.yaml configs/apiserver_local.yaml`
2. 在本地配置中配置数据库
3. 启动依赖服务：`make run-redis`（或 `docker-compose -f deployments/redis/docker-compose.yml up`）
4. 运行：`go run cmd/apiserver/main.go -c configs/apiserver_local.yaml`

### 快速设置替代方案
```bash
make dev-setup      # 安装工具 + 启动服务 + 生成代码
make run-api        # 启动 API 服务器
```

### 开发依赖
- **数据库**：MySQL 8.0+ 或 PostgreSQL 12+
- **缓存**：Redis 6.2+（提供 docker-compose）
- **可观测性**：Jaeger, Prometheus（提供 docker-compose）

### 配置文件
- `configs/apiserver.yaml` - 默认配置
- `configs/apiserver_v1.yaml` - 替代配置模板
- 环境特定配置使用格式：`apiserver_<env>.yaml`

## 测试命令

### 标准测试
```bash
go test ./...                    # 运行所有测试
go test -v ./pkg/errorsx/       # 详细运行包测试
go test -run TestSpecific       # 运行特定测试
go test -bench=. ./...          # 运行性能测试
```

### 集成测试
```bash
docker-compose -f deployments/redis/docker-compose.yml up -d
go test -tags=integration ./...   # 运行集成测试
```

## API 开发

### 添加新端点
1. **API 定义**：编辑 `pkg/api/apiserver/v1/apiserver.proto`
2. **生成代码**：`make generate`
3. **实现处理器**：在 `internal/apiserver/handler/` 中创建
4. **Wire 依赖**：运行 `make wire`
5. **文档**：自动生成在 `api/openapi/apiserver/v1/`

### 关键开发约定
- **错误代码**：在 protobuf 中定义，使用 `protoc-gen-go-errors-code` 自动生成
- **数据验证**：使用 protoc-gen-validate 注解
- **国际化**：使用 `pkg/i18n/` 和上下文语言检测
- **日志记录**：通过上下文中间件的结构化日志
- **测试**：测试名称遵循 `Test<Level><Description>` 模式

## 包导航指南

### 起始点
- **服务器启动**：`cmd/apiserver/app/server.go:Start()`
- **处理器示例**：`internal/apiserver/handler/user.go`
- **Wire 设置**：`internal/apiserver/wire_gen.go:InitializeWebServer()`
- **错误处理**：`pkg/errorsx/builder.go:NewCode()`
- **数据库设置**：`pkg/db/mysql.go:NewMySQL()`

### 配置模式
- **特性开关**：`pkg/feature/` - 按环境切换特性
- **选项模式**：`pkg/options/` - 一致的组件配置
- **环境变量覆盖**：Viper 自动处理环境变量到结构体映射

## 基础设施模板

### 可用的 Docker 服务
- **Redis**：`make run-redis` 或 `docker-compose -f deployments/redis/docker-compose.yml up`
- **Jaeger**：`make run-jaeger` 或 `docker-compose -f deployments/jaeger/docker-compose.yml up`
- **Kafka**：`make run-kafka` 或 `docker-compose -f deployments/kafka/docker-compose.yml up`
- **所有服务**：`make start-all`（同时启动 Redis, Jaeger, Kafka）

### Observability Stack (统一收集与分析)

项目集成了完整的可观测性stack，提供统一的数据收集和分析能力：

| 功能 | 工具 | 用途 | 接入方式 |
|------|------|------|----------|
| **全栈监控** | Grafana+Prometheus | 系统/应用指标可视化 | `make run-grafana` |
| **链路追踪** | Jaeger + OpenTelemetry | 分布式请求追踪 | 已内置集成 |
| **告警管理** | Alertmanager | 统一告警路由 | 与Prometheus集成 |
| **日志采集** | OpenTelemetry Collector | 统一日志/指标/追踪收集 | `make run-otelcol` |
| **仪表板** | Grafana Dashboards | 预置监控面板 | 启动后访问 `localhost:3000` |

### 架构关系
```
应用程序 → [OpenTelemetry Collector] → [Prometheus] → [Grafana]
     ↓                                ↓                ↓
  生成追踪/日志/指标             存储时序数据      可视化分析
                                         ↓
                                   [Alertmanager] → 告警通知
```

### 监控端点
- **健康检查**: `GET /health`
- **指标**: `GET /metrics` (Prometheus)
- **链路追踪**: `GET /jaeger` (Jaeger UI) - 分布式链路追踪
- **Grafana**: `http://localhost:3000` - 统一监控面板
- **Alertmanager**: `http://localhost:9093` - 告警管理界面
- **调试**: `GET /debug/pprof` (启用时)

## 模块信息
- **Go 版本**: 需要 1.24.0+
- **模块路径**: `github.com/costa92/go-protoc/v2`
- **分支**: 当前在 `v2` 分支
- **Protobuf**: 使用 buf.build 进行依赖管理

## 项目专用工具
- **项目重命名**: 使用 `make rename-project OLD_PATH=X NEW_PATH=Y` 更改模块路径
- **Git 钩子**: 自动安装的 Git 钩子: githooks/{pre-commit,commit-msg,pre-push}

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.