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
使用脚本管理数据库服务：
```bash
./scripts/installation/service.sh start redis
./scripts/installation/service.sh start mariadb  
./scripts/installation/service.sh start mongodb
```

#### MySQL数据库管理（新增）
项目新增了完整的MySQL数据库管理功能，支持自动化的数据库设置、迁移和维护：

**快速开始**：
- `make db-setup` - 一键完整数据库设置（启动 + 迁移）
- `make db-start` - 启动MySQL容器
- `make db-migrate` - 执行数据库迁移
- `make db-help` - 查看所有数据库命令

**数据库操作**：
- `make db-connect` - 通过MySQL CLI连接数据库
- `make db-shell` - 在Docker容器中打开MySQL shell
- `make db-status` - 检查容器状态
- `make db-logs` - 查看数据库日志

**迁移管理**：
- `make db-migrate-dry` - 查看迁移计划（干运行）
- `make db-create-migration NAME=migration_name` - 创建新迁移文件

**备份与恢复**：
- `make db-backup` - 创建数据库备份
- `make db-restore BACKUP_FILE=backup.sql` - 从备份恢复

**维护命令**：
- `make db-reset` - 重置数据库（删除所有数据）
- `make db-clean` - 删除容器和卷（永久删除数据）

**数据库配置**：
- 数据库: `onex`
- 主机: `127.0.0.1:3306`
- 用户名: `root`
- 密码: `proj(#)666`

**默认用户表结构**：
- `users` - 主用户表（用户名、邮箱、密码等基本信息）
- `user_profiles` - 用户配置表（个人简介、偏好设置等）
- `user_roles` - 用户角色表（权限管理）

#### 其他服务
使用脚本管理各类服务：
```bash
# 消息队列服务
./scripts/installation/service.sh start kafka

# 分布式服务  
./scripts/installation/service.sh start etcd

# 可观测性服务
./scripts/installation/service.sh start jaeger
./scripts/installation/service.sh start prometheus
./scripts/installation/service.sh start grafana
./scripts/installation/service.sh start alertmanager
./scripts/installation/service.sh start otelcol
./scripts/installation/service.sh start victorialogs
```

#### 服务组管理
```bash
# 服务组操作
./scripts/installation/service.sh start all
./scripts/installation/service.sh stop all
./scripts/installation/service.sh restart all
./scripts/installation/service.sh start database
./scripts/installation/service.sh start observability
./scripts/installation/service.sh status all
./scripts/installation/service.sh logs all
```

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
3. 启动依赖服务：`./scripts/installation/service.sh start redis`（或 `docker-compose -f deployments/redis/docker-compose.yml up`）
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
- **Redis**：`./scripts/installation/service.sh start redis` 或 `docker-compose -f deployments/redis/docker-compose.yml up`
- **Jaeger**：`./scripts/installation/service.sh start jaeger` 或 `docker-compose -f deployments/jaeger/docker-compose.yml up`
- **Kafka**：`./scripts/installation/service.sh start kafka` 或 `docker-compose -f deployments/kafka/docker-compose.yml up`
- **所有服务**：`./scripts/installation/service.sh start all`（同时启动 Redis, Jaeger, Kafka）

### Observability Stack (统一收集与分析)

项目集成了完整的可观测性stack，提供统一的数据收集和分析能力：

| 功能 | 工具 | 用途 | 接入方式 |
|------|------|------|----------|
| **全栈监控** | Grafana+Prometheus | 系统/应用指标可视化 | `./scripts/installation/service.sh start grafana` |
| **链路追踪** | Jaeger + OpenTelemetry | 分布式请求追踪 | 已内置集成 |
| **告警管理** | Alertmanager | 统一告警路由 | 与Prometheus集成 |
| **日志采集** | OpenTelemetry Collector | 统一日志/指标/追踪收集 | `./scripts/installation/service.sh start otelcol` |
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
- **指标**: `GET /metrics` (Prometheus) - 包含连接池监控指标
- **链路追踪**: `GET /jaeger` (Jaeger UI) - 分布式链路追踪
- **Grafana**: `http://localhost:3000` - 统一监控面板
- **Alertmanager**: `http://localhost:9093` - 告警管理界面
- **调试**: `GET /debug/pprof` (启用时)

### 连接池监控功能

项目实现了**依赖注入式连接池监控**，完美平衡了单一职责和监控需求：

#### 设计特点
- ✅ **完全解耦**: `pkg/db` 专注数据库连接，`pkg/metrics` 专门负责监控
- ✅ **可选启用**: 通过 Wire 依赖注入，可选择性开启监控功能
- ✅ **零侵入**: 不启用监控时，数据库包没有任何监控开销
- ✅ **向后兼容**: 原有数据库使用方式完全不变

#### 使用方式

**启用监控 (推荐)**:
```bash
# 监控功能已通过 Wire 自动装配到项目中
make run-api  # 启动时会显示: "Connection pool metrics collection is enabled through dependency injection"
```

**监控指标**:
- `database_pool_connections` - MySQL/PostgreSQL连接池状态
- `redis_pool_connections` - Redis连接池状态  
- `database_queries_total` - 数据库查询统计
- `redis_commands_total` - Redis命令统计

**查看指标**:
```bash
curl http://localhost:8080/metrics | grep -E "(database_|redis_)"
```

#### 架构关系
```
Wire依赖注入 → PoolMonitor → pkg/db (可选监控) → pkg/metrics (指标收集)
```

## 模块信息
- **Go 版本**: 需要 1.24.0+
- **模块路径**: `github.com/costa92/go-protoc/v2`
- **分支**: 当前在 `v2` 分支
- **Protobuf**: 使用 buf.build 进行依赖管理

## 项目专用工具
- **项目重命名**: 使用 `make rename-project OLD_PATH=X NEW_PATH=Y` 更改模块路径
- **Git 钩子**: 自动安装的 Git 钩子: githooks/{pre-commit,commit-msg,pre-push}

## VictoriaLogs 日志集成

### 🔥 已完成集成
项目已完全集成 **VictoriaLogs** 作为统一日志管理解决方案，支持结构化日志存储、查询和可视化。

#### 📊 核心特性
- ✅ **完全集成**: 基于 Zap 的高性能日志系统
- ✅ **结构化日志**: JSON 格式，支持字段检索和过滤
- ✅ **批量发送**: 异步处理，1000 条缓冲区，100 条批量发送
- ✅ **多路输出**: 同时输出到控制台、文件和 VictoriaLogs
- ✅ **UI 界面**: Web 界面支持实时日志查询和可视化

#### 🌐 访问方式
```bash
# Web UI 界面（推荐）
http://127.0.0.1:9428/select/vmui/

# API 查询接口
curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=*'
```

#### 🚀 管理命令
```bash
# VictoriaLogs 服务管理
make deploy.install.docker.victorialogs    # 安装服务
make deploy.status.victorialogs             # 检查状态
make deploy.uninstall.docker.victorialogs  # 卸载服务

# Victoria 完整套件管理
make deploy.install.all.victoria            # 安装所有组件
make deploy.uninstall.all.victoria          # 卸载所有组件
make deploy.status.victoria                 # 检查套件状态

# 统一脚本管理
./scripts/installation/victoria.sh install.all      # 安装所有组件
./scripts/installation/victoria.sh uninstall.all    # 卸载所有组件
./scripts/installation/victoria.sh status           # 检查状态

# 直接服务管理
./scripts/installation/service.sh start victorialogs
./scripts/installation/service.sh stop victorialogs
```

#### 🔍 日志查询工具
项目提供了专用的日志查询脚本：

```bash
# 使用查询脚本
./scripts/victoria-logs-query.sh stats             # 日志统计
./scripts/victoria-logs-query.sh recent            # 最近日志
./scripts/victoria-logs-query.sh info              # Info 级别日志
./scripts/victoria-logs-query.sh "_msg:*MySQL*"    # 自定义查询

# 使用测试脚本
./scripts/test-victorialogs-ui.sh                  # 完整集成测试
```

#### 📝 查询语法示例
```bash
# 基础查询
*                           # 所有日志
level:info                  # 按级别过滤
_msg:*server*              # 消息内容包含 "server"
_time:>2025-08-17          # 时间范围过滤

# 复合查询
level:info AND _msg:*MySQL*     # Info 级别且包含 MySQL
level:(error OR warn)           # 错误或警告级别
service:apiserver               # 特定服务日志
```

#### ⚙️ 配置说明
日志配置位于 `configs/apiserver.yaml`:

```yaml
log:
  level: "info"
  format: "json"                    # VictoriaLogs 推荐格式
  output-paths: ["stdout", "logs/app.log"]
  victoria-logs:
    enabled: true                   # 启用 VictoriaLogs
    endpoint: "http://127.0.0.1:9428"
    service: "apiserver"
    version: "v2.0.0"
    environment: "development"
    buffer-size: 1000              # 缓冲区大小
    batch-size: 100                # 批量发送大小
    flush-interval: 5              # 刷新间隔（秒）
    timeout: 10                    # 请求超时（秒）
```

#### 🏗️ 架构集成
```
应用程序 → Zap Logger → Tee Core → 多路输出:
                                  ├─ 控制台 (开发调试)
                                  ├─ 文件 (本地持久化)
                                  └─ VictoriaLogs (集中管理)
```

#### 📈 监控指标
- **日志吞吐量**: 支持高并发日志写入
- **查询性能**: 毫秒级日志检索响应
- **存储效率**: 压缩存储，节省磁盘空间
- **可视化**: 时间序列图表，日志分布统计

#### 🔗 与其他组件集成
- **Jaeger 追踪**: 日志中包含 trace_id 关联
- **Prometheus 指标**: 日志错误率监控
- **数据库日志**: GORM 查询日志自动收集
- **中间件日志**: HTTP/gRPC 请求响应日志

## AI Agent 模块开发计划

### 🤖 AI Agent 架构设计

AI Agent 作为项目的核心扩展模块，旨在提供智能化的代码生成、项目分析和对话交互功能。

#### 目录结构规划
```
项目根目录/
├── cmd/ai/                          # AI服务入口（已存在）
│   ├── main.go                      # AI服务主程序
│   └── app/                         # AI应用配置
├── internal/ai/                     # AI服务内部实现
│   ├── handler/                     # AI API处理器
│   │   ├── chat.go                  # 对话处理
│   │   ├── code.go                  # 代码生成
│   │   └── analysis.go              # 项目分析
│   ├── biz/                        # AI业务逻辑层
│   │   ├── agent/                   # AI代理核心
│   │   ├── llm/                     # LLM集成
│   │   ├── knowledge/               # 知识库管理
│   │   └── workflow/                # 工作流引擎
│   ├── store/                      # AI数据存储
│   │   ├── conversation.go          # 对话存储
│   │   ├── knowledge.go             # 知识库存储
│   │   └── workflow.go              # 工作流存储
│   └── pkg/                        # AI内部工具
├── pkg/api/ai/v1/                  # AI API定义
│   ├── ai.proto                    # AI服务定义
│   └── errors.proto                # AI错误码
└── configs/ai/                     # AI配置文件
```

#### 核心架构分层
```
┌─────────────────────────────────────┐
│           Handler Layer             │ ← gRPC/HTTP API接口层
├─────────────────────────────────────┤
│          Business Layer             │ ← AI业务逻辑层
│  ┌─────────────────────────────────┐ │
│  │        Agent Core              │ │ ← AI代理核心
│  │  ┌─────────┬─────────┬────────┐ │ │
│  │  │   LLM   │Knowledge│Workflow│ │ │ ← 核心组件
│  │  └─────────┴─────────┴────────┘ │ │
│  └─────────────────────────────────┘ │
├─────────────────────────────────────┤
│          Storage Layer              │ ← 数据存储层
└─────────────────────────────────────┘
```

### API 接口规划

#### 核心服务接口
```protobuf
service AIAgent {
  // 对话聊天接口 - 支持流式响应
  rpc Chat(ChatRequest) returns (stream ChatResponse);
  
  // 代码生成接口 - 基于需求生成代码
  rpc GenerateCode(CodeGenRequest) returns (CodeGenResponse);
  
  // 项目分析接口 - 分析项目结构和质量
  rpc AnalyzeProject(AnalysisRequest) returns (AnalysisResponse);
  
  // 知识库管理 - 项目知识索引和搜索
  rpc UpdateKnowledge(KnowledgeRequest) returns (KnowledgeResponse);
  
  // 工作流执行 - 复杂任务自动化
  rpc ExecuteWorkflow(WorkflowRequest) returns (stream WorkflowResponse);
}
```

#### HTTP 路由映射
- `POST /v1/ai/chat` - 智能对话
- `POST /v1/ai/code/generate` - 代码生成
- `POST /v1/ai/analyze` - 项目分析
- `POST /v1/ai/knowledge` - 知识库管理
- `POST /v1/ai/workflow` - 工作流执行

### 核心组件设计

#### 1. LLM Provider 抽象层
- **多Provider支持**: OpenAI、Claude、本地模型
- **负载均衡**: 智能路由和容错机制
- **成本控制**: Token使用统计和限流
- **缓存策略**: 响应缓存和预热机制

#### 2. 知识库管理系统
- **项目索引**: 自动扫描和索引项目代码
- **向量搜索**: 基于语义的代码和文档检索
- **增量更新**: 监控文件变化，增量更新索引
- **上下文增强**: RAG机制提供相关上下文

#### 3. 工作流引擎
- **任务编排**: 可视化的工作流定义
- **并发执行**: 支持步骤间依赖和并行执行
- **状态管理**: 完整的执行状态跟踪和恢复
- **错误处理**: 重试机制和失败回滚

### 数据模型设计

#### 对话会话管理
```go
type Conversation struct {
    ID        string    `gorm:"primaryKey"`
    UserID    string    `gorm:"index"`
    Title     string
    Context   JSON      `gorm:"type:json"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Message struct {
    ID             string `gorm:"primaryKey"`
    ConversationID string `gorm:"index"`
    Role           string // user, assistant, system
    Content        string `gorm:"type:text"`
    Metadata       JSON   `gorm:"type:json"`
    CreatedAt      time.Time
}
```

#### 知识库条目
```go
type KnowledgeEntry struct {
    ID          string    `gorm:"primaryKey"`
    Type        string    // code, doc, config
    Path        string    `gorm:"index"`
    Content     string    `gorm:"type:longtext"`
    Embedding   []float32 `gorm:"type:json"`
    Hash        string    `gorm:"index"`
    UpdatedAt   time.Time
}
```

#### 工作流执行记录
```go
type WorkflowExecution struct {
    ID         string              `gorm:"primaryKey"`
    Name       string
    Status     WorkflowStatus
    Steps      []WorkflowStep      `gorm:"type:json"`
    Results    map[string]any      `gorm:"type:json"`
    CreatedAt  time.Time
    FinishedAt *time.Time
}
```

### 配置管理

#### AI 专用配置
```yaml
# configs/ai/ai.yaml
ai:
  llm:
    default_provider: "openai"
    providers:
      openai:
        api_key: "${AI_OPENAI_API_KEY}"
        model: "gpt-4"
        max_tokens: 4000
      claude:
        api_key: "${AI_CLAUDE_API_KEY}"
        model: "claude-3-sonnet"
        max_tokens: 4000
    rate_limit:
      requests_per_minute: 60
      tokens_per_day: 100000
  
  knowledge:
    vectordb:
      type: "chroma"
      host: "localhost:8000"
    embedding:
      model: "text-embedding-ada-002"
      dimensions: 1536
    index_path: "./data/knowledge"
  
  workflow:
    max_concurrent: 5
    timeout: "30m"
    retry:
      max_attempts: 3
      backoff: "exponential"
```

### 开发命令扩展

#### AI 专用命令
```bash
# AI 服务管理 (需要实现)
# TODO: 以下命令待实现
# make run-ai                  # 启动 AI 服务器
# make stop-ai                 # 停止 AI 服务器
# make restart-ai              # 重启 AI 服务器

# AI 开发工具 (需要实现)
# make ai-generate             # 生成 AI API 代码
# make ai-index                # 构建项目知识索引
# make ai-test                 # 运行 AI 模块测试

# 知识库管理 (需要实现)
# make knowledge-build         # 构建完整知识库
# make knowledge-update        # 增量更新知识库
# make knowledge-search QUERY="<query>"  # 搜索知识库

# 向量数据库 (需要实现)
# make run-vectordb           # 启动向量数据库
# make stop-vectordb          # 停止向量数据库
```

### 可观测性增强

#### AI 专用监控指标
- `ai_llm_requests_total` - LLM调用总数
- `ai_llm_request_duration_seconds` - LLM请求延迟
- `ai_token_usage_total` - Token使用量统计
- `ai_knowledge_search_duration_seconds` - 知识库搜索延迟
- `ai_workflow_execution_total` - 工作流执行统计

#### 链路追踪增强
- LLM Provider调用链路
- 知识库检索操作追踪
- 工作流步骤执行追踪
- Token使用情况追踪

### 实施路线图

#### 阶段一：基础架构搭建 (2-3周)
- [ ] AI API定义和代码生成
- [ ] 基础存储层实现
- [ ] 简单LLM Provider集成
- [ ] 基础配置和依赖注入

#### 阶段二：核心功能实现 (3-4周)
- [ ] 对话功能完整实现
- [ ] 代码生成功能开发
- [ ] 基础知识库系统
- [ ] 错误处理和中间件集成

#### 阶段三：高级特性开发 (4-5周)
- [ ] 工作流引擎实现
- [ ] 多Provider支持和负载均衡
- [ ] 向量搜索和RAG优化
- [ ] 流式响应和实时更新

#### 阶段四：生产优化 (2-3周)
- [ ] 性能优化和缓存策略
- [ ] 完整的监控和可观测性
- [ ] 压力测试和稳定性优化
- [ ] 文档完善和部署指南

### 集成策略

#### 与现有架构的无缝集成
1. **遵循清洁架构**: 使用相同的分层模式和依赖关系
2. **Wire依赖注入**: 完全集成到现有的依赖注入体系
3. **错误处理**: 使用统一的ErrorX错误处理机制
4. **中间件复用**: 复用现有的认证、日志、追踪中间件
5. **配置管理**: 扩展现有的配置系统和环境变量支持
6. **可观测性**: 集成到现有的监控和追踪体系

#### 数据库扩展
- 复用现有的GORM配置和事务机制
- 扩展数据库模型以支持AI相关数据
- 保持数据库迁移的一致性

这个AI Agent模块设计充分利用了现有项目的成熟架构，保持了设计一致性，同时为未来的AI功能扩展提供了强大的基础设施支持。

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.