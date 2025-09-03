# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Production-ready Go microservice framework built on Kratos v2 with Protocol Buffers as unified data source. Uses Clean Architecture, Wire dependency injection, supports HTTP/gRPC APIs with complete error handling, i18n, auth, and observability.

**Module Path**: `github.com/costa92/go-protoc/v2`
**Go Version**: 1.24.0+
**Current Branch**: `v2`

## Critical Constraints

- **Environment Variables**: MUST source from `manifests/env/` directory - do NOT create additional .env files
- **Environment File Hierarchy**: Environment configs inherit from `manifests/env/env.base` (fixed values first, logical groupings)
- **Version Management**: ⚠️ **ALL versions MUST come from `manifests/env/env.base`** - scripts must source from manifests/env, NOT the reverse
- **Multi-Environment Support**: Use `PROJ_ENVIRONMENT={dev,test,prod}` for environment-specific configurations
- **Service Management**: ⚠️ **PREFER Docker Template System** (`make docker.<service>.start`) over legacy service scripts
- **Infrastructure Setup**: All service installations use template system in `scripts/installation/templates/`
- **Protobuf Files**: Must be placed under `pkg/api/<service>/<version>/` structure (per `.cursor/rules/buf.mdc`)
- **Generated/System Files**: ⚠️ **NEVER modify files in `_*` directories** (`_output/`, `_thirdparty/`, `_generated/`) - these contain auto-generated files, build artifacts, and third-party service data

## Essential Development Commands

### Daily Commands

- `make help` - Show all available commands
- `make run-api` - Start development server (with hot reload)
- `make build` - Build optimized binary to ./bin/apiserver
- `make test` - Run all tests with verbose output
- `make fmt` - Format code and sort imports
- `make tidy` - Clean go.mod dependencies

### Code Generation

- `make generate` - Generate protobuf/gRPC/HTTP code using buf
- `make wire` - Regenerate dependency injection (run after structural changes)

### Environment Setup

- `make dev-setup` - Complete development environment (tools + services + codegen)
- `make dev-clean` - Clean development environment and stop services
- `make dev-quick` - Quick development start (skip tool installation)
- `make dev-watch` - Development mode with file watching and auto-rebuild
- `make dev-test` - Run full test pipeline with coverage
- `make dev-bench` - Run benchmarks in development environment

### Infrastructure & Service Management

**⚠️ IMPORTANT: Use Docker Template System (Preferred)**

```bash
# Individual service management via Make templates
make docker.redis.start         # Start Redis using template system
make docker.mariadb.start       # Start MariaDB via template
make docker.prometheus.start    # Start Prometheus with auto-detection
make docker.otelcol.start       # Start OTEL Collector
make docker.victorialogs.start  # Start VictoriaLogs

# Service operations
make docker.redis.stop          # Stop service
make docker.redis.status        # Check status
make docker.redis.restart       # Restart service
make docker.redis.cleanup       # Clean up generated scripts
```

**Available Template Services:** redis, mariadb, mongodb, kafka, etcd, jaeger, prometheus, grafana, victorialogs, otelcol, pyroscope

**Batch Operations:**

```bash
make docker.test-services.start # Start all test services
make docker.test-services.stop  # Stop all test services
```

**Legacy Service Script (NOT RECOMMENDED):**

```bash
# These commands exist but are NOT the preferred method
# Use docker templates above instead
./scripts/installation/service.sh start redis   # Not recommended
./scripts/installation/service.sh start mariadb # Not recommended
```

### Database Management

- `make db-setup` - Complete MySQL setup (start + migrate)
- `make db-connect` - Connect to MySQL via CLI
- `make db-migrate` - Execute database migrations
- Database config: `onex` database, `127.0.0.1:3306`, user: `root`, password: `proj(#)666`

### Build and Version Management

Version-aware build system with automatic Git version injection:

- `make build` - Build with version info (default: apiserver)
- `SERVICE_NAME=myservice make build` - Build with custom service name
- `./bin/apiserver --version` - Show version info
- Auto-injected: Git version, branch, commit, build time, clean/dirty status

## Architecture Overview

### Clean Architecture Layers

```
cmd/                    → Entry points (apiserver, ai, pump)
internal/apiserver/     → Core business logic
 ├── handler/           → HTTP/gRPC handlers (delivery layer)
 ├── biz/              → Use cases (application layer)
 ├── store/            → Data access (infrastructure layer)
 └── config/           → Internal configuration

pkg/                   → Reusable packages
 ├── api/              → Protobuf definitions and generated code
 ├── errorsx/          → Context-aware error system (i18n support)
 ├── authn/           → JWT authentication utilities
 ├── db/              → Database abstractions
 ├── logger/          → Unified logging interface with OTLP support
 ├── server/          → HTTP/gRPC server configuration
 ├── version/         → Version info management
 ├── metrics/         → Connection pool monitoring system
 └── options/         → Component configuration architecture
```

### Infrastructure Template System

The project uses a sophisticated template-based infrastructure management system:

**Template Structure:**

```
scripts/installation/templates/docker/
 ├── prometheus/        → Prometheus with auto-discovery
 │   ├── docker-run.sh.tpl              → Main startup template
 │   ├── exporters/exporter-manager.sh.tpl  → Unified exporter manager
 │   └── rule_files/*.yml.tpl           → Alert rule templates
 ├── redis/            → Redis configuration templates
 ├── mariadb/          → MariaDB templates
 └── kafka/            → Kafka stack templates
```

**System Directories (DO NOT MODIFY):**

```
_output/               → Build artifacts, binaries, compiled assets
_thirdparty/          → Third-party service data and configurations
 ├── redis/data/       → Redis persistence files
 ├── mariadb/mysql/    → MariaDB database files
 ├── nacos/config/     → Nacos service configurations
 └── [service]/logs/   → Service log files
_generated/           → Auto-generated Docker scripts and configurations
 └── docker-scripts/   → Generated from templates, regenerated on each run
```

**Key Features:**

- **Environment Variable Injection**: All templates use `envsubst` for dynamic configuration
- **Service Auto-Discovery**: Prometheus automatically detects running services and starts exporters
- **Modular Design**: Reusable components across different services
- **Security**: Database credentials from environment variables, never hardcoded

### Dependency Injection (Wire)

- **Centralized**: `internal/apiserver/wire.go`
- **Generated**: `internal/apiserver/wire_gen.go`
- **Auto-discovery**: Run `make wire` after adding new dependencies
- **Connection Pool Monitoring**: Integrated via Wire for optional metrics collection

### Request Flow

```
HTTP Request → gRPC-Gateway → Handler → Business → Store → Database
     ↓                                             ↓
OpenAPI Docs (auto-generated)            GORM + Context Transactions + Pool Monitoring
```

## Development Workflow

### Adding New API Endpoints

1. **Define API**: Edit `pkg/api/apiserver/v1/apiserver.proto`
2. **Generate Code**: `make generate`
3. **Implement Handler**: Create in `internal/apiserver/handler/`
4. **Wire Dependencies**: Run `make wire`
5. **Documentation**: Auto-generated in `api/openapi/apiserver/v1/`

### Testing

```bash
go test ./...                    # Run all tests
go test -v ./pkg/errorsx/       # Verbose package test
go test -run TestSpecific       # Run specific test
go test -bench=. ./...          # Run benchmarks
```

### Local Development Setup

```bash
cp configs/apiserver.yaml configs/apiserver_local.yaml  # Copy config
# Edit local config for database settings
./scripts/installation/service.sh start redis           # Start dependencies
go run cmd/apiserver/main.go -c configs/apiserver_local.yaml  # Run server
```

### Key Development Conventions

- **Error Codes**: Defined in protobuf, auto-generated using `protoc-gen-go-errors-code`
- **Validation**: Use protoc-gen-validate annotations
- **i18n**: Use `pkg/i18n/` with context language detection
- **Logging**: Structured logging via context middleware
- **Testing**: Follow `Test<Level><Description>` pattern

### Protocol Buffers Development Rules

⚠️ **CRITICAL**: Follow these protobuf conventions (from `.cursor/rules/buf.mdc`):

1. **File Location**: ALL `.proto` files MUST be placed under `pkg/api` directory
2. **Directory Structure**: Use `pkg/api/<service>/<version>/*.proto` pattern
3. **Example**: `pkg/api/apiserver/v1/error.proto`
4. **Build Tool**: Use `buf` (reference: <https://buf.build/docs/cli/quickstart/>)
5. **Generated Code**: Outputs to `pkg/api` directory automatically
6. **Documentation**: Generated to `docs` directory

## Observability & Monitoring

### Monitoring Endpoints

- **Health Check**: `GET /health`
- **Metrics**: `GET /metrics` (Prometheus format with version labels and connection pool stats)
- **Jaeger UI**: `http://localhost:16686` - Distributed tracing
- **Grafana**: `http://localhost:3000` - Unified monitoring dashboard
- **VictoriaLogs UI**: `http://127.0.0.1:9428/select/vmui/` - Log analysis
- **Prometheus**: `http://localhost:9090` - Metrics collection

### Connection Pool Monitoring

The project features **dependency-injected connection pool monitoring**:

**Key Metrics:**

- `database_pool_connections{database,state}` - MySQL/PostgreSQL pool status
- `redis_pool_connections{instance,state}` - Redis connection pool status
- `database_queries_total{database,operation,status}` - Query statistics
- `redis_commands_total{instance,command,status}` - Redis command statistics

**Architecture Benefits:**

- ✅ **Zero Intrusion**: Database packages remain clean, monitoring is optional
- ✅ **Wire Integration**: Automatically injected through dependency injection
- ✅ **Version Aware**: All metrics include service version labels

### Unified Logging System

Complete log collection via **OpenTelemetry Collector + VictoriaLogs**:

**Deployment Commands:**

```bash
./scripts/deploy-logging.sh local    # Local development
./scripts/deploy-logging.sh docker   # Docker environment with Grafana
./scripts/deploy-logging.sh k8s      # Kubernetes deployment
```

**Log Querying:**

```bash
# Basic queries
curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=level:error'
curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=service.name:apiserver'
curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=_time:>now-1h'

# Advanced queries
curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=service.name:apiserver AND level:error AND _time:>now-30m'
```

## Key Entry Points for Code Navigation

**Server Startup**: `cmd/apiserver/app/server.go:Start()`
**Version Info**: `pkg/version/version.go:Get()`
**HTTP Server**: `pkg/server/http_server.go:RunOrDie()`
**gRPC Server**: `pkg/server/grpc_server.go:RunOrDie()`
**Wire Setup**: `internal/apiserver/wire_gen.go:InitializeWebServer()`
**Handler Example**: `internal/apiserver/handler/user.go`
**Error Handling**: `pkg/errorsx/builder.go:NewCode()`
**Database Setup**: `pkg/db/mysql.go:NewMySQL()`

## Configuration

**Config Files**: `configs/apiserver.yaml` (default), `configs/apiserver_local.yaml` (local dev)
**Environment-specific**: Use format `apiserver_<env>.yaml`

## Important Notes

- **Wire Dependency Injection**: Always run `make wire` after adding new dependencies
- **Protobuf Changes**: Run `make generate` after editing .proto files
- **Version Info**: Available at runtime via `pkg/version/version.go:Get()`
- **Error Handling**: Use `pkg/errorsx` for context-aware, i18n-enabled errors
- **Testing**: Include integration tests with Docker services when needed
- **System Directories**: Never modify files in `_output/`, `_thirdparty/`, `_generated/` directories - these are automatically managed

## Infrastructure Components

### Supported Services with Versions

| Component | Version | Purpose | **Preferred Management** |
|-----------|---------|---------|--------------------------|
| Redis | 7.2.4 | Caching, sessions | `make docker.redis.start` |
| MariaDB | 11.2.2 | Primary database | `make docker.mariadb.start` |
| MongoDB | 7.0.5 | Document storage | `make docker.mongodb.start` |
| Kafka | 6.2.0 | Message streaming | `make docker.kafka.start` |
| Jaeger | 1.52.0 | Distributed tracing | `make docker.jaeger.start` |
| Prometheus | 2.48.1 | Metrics collection | `make docker.prometheus.start` |
| Grafana | 10.2.4 | Monitoring dashboards | `make docker.grafana.start` |
| VictoriaLogs | 1.28.0 | Log aggregation | `make docker.victorialogs.start` |
| OTEL Collector | 0.132.0 | Telemetry collection | `make docker.otelcol.start` |

### Environment Configuration

**Environment File Hierarchy**:
- `manifests/env/env.base` - Master configuration with all versions and base settings
- `manifests/env/env.dev` - Development environment (sources from env.base)
- `manifests/env/env.test` - Test environment (sources from env.base)  
- `manifests/env/env.prod` - Production environment (sources from env.base)
- `manifests/env/env.local` - Local overrides (sources from env.base)

**Template Generation**: `scripts/installation/lib/docker_script_manager.sh` - Dynamic script generation

**⚠️ Critical Environment File Constraints:**
- `manifests/env/` files MUST be self-contained and NOT reference `scripts/` directory  
- `scripts/` directory MUST source ALL environment variables from `manifests/env/` files
- ALL version definitions MUST come from `manifests/env/env.base`, never from `scripts/installation/versions.sh`
- Environment-specific files inherit from `env.base` and override only necessary values
- Use `PROJ_ENVIRONMENT={development,test,production}` to load appropriate environment config
- This ensures portability and avoids circular dependencies between manifests and scripts

**Environment Loading Pattern**:
```bash
# Correct way - scripts source from manifests
source manifests/env/env.dev  # This automatically loads env.base first

# NEVER do this - creates circular dependency  
source scripts/installation/versions.sh
```

**Multi-Environment Port Management**:
The project uses sophisticated port prefixing for multi-environment support:
- **Development**: ports prefixed with `1` (e.g., Redis: 16379, MySQL: 13306)
- **Test**: ports prefixed with `2` (e.g., Redis: 26379, MySQL: 23306)  
- **Production**: standard ports (e.g., Redis: 6379, MySQL: 3306)
- **Custom**: ports prefixed with `9` (e.g., Redis: 96379, MySQL: 93306)

This allows multiple environments to run simultaneously without port conflicts.

### Database Setup Requirements

When working with database monitoring (especially Prometheus MySQL exporter):

**MySQL/MariaDB Monitoring User Setup:**

```bash
# Required for Prometheus MySQL exporter - use template system
make docker.mariadb.start

# Create monitoring user with minimal privileges:
docker exec proj-mariadb mariadb -u root -p'proj(#)666' -e "
CREATE USER IF NOT EXISTS 'exporter'@'%' IDENTIFIED BY 'exporter123';
GRANT PROCESS ON *.* TO 'exporter'@'%';
GRANT REPLICATION CLIENT ON *.* TO 'exporter'@'%';
GRANT SELECT ON performance_schema.* TO 'exporter'@'%';
FLUSH PRIVILEGES;"
```

**Security Requirements:**

- Database credentials MUST use environment variables from `manifests/env/env.dev`
- Never hardcode passwords in templates or scripts
- Use dedicated monitoring users with minimal required privileges

## Advanced Log Collection System

Complete log collection via **OpenTelemetry Collector + VictoriaLogs** with multi-environment support.

### Core Features

- **Dual Collection**: OTLP protocol + file monitoring
- **Structured JSON**: Automatic parsing and field extraction
- **Real-time**: Millisecond-level log collection latency
- **Multi-environment**: Local, Docker, and Kubernetes deployment
- **Unified Config**: Centralized parameter management
- **High Availability**: Fault tolerance and retry mechanisms

### Quick Deployment

```bash
./scripts/deploy-logging.sh local    # Local development
./scripts/deploy-logging.sh docker   # Docker with Grafana
./scripts/deploy-logging.sh k8s      # Kubernetes deployment
./scripts/deploy-logging.sh test local  # Test log collection
```

### Access Points

- **VictoriaLogs UI**: `http://127.0.0.1:9428/select/vmui/` (recommended)
- **API Queries**: `curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=*'`
- **Grafana Dashboard**: `http://127.0.0.1:3000` (admin/admin for Docker env)

### Log Query Syntax Examples

```bash
# Basic queries
level:error                          # Error logs
service.name:apiserver               # Specific service
_time:>now-1h                        # Last hour

# Complex queries
service.name:apiserver AND level:error AND _time:>now-30m
level:(error OR warn) AND k8s.namespace.name:app

# Regex patterns
msg:~"user.*login"                   # User login related
error:~"database.*connection"        # Database connection errors
```

### OTLP Configuration

Application logging config in `configs/apiserver.yaml`:

```yaml
log:
  type: "zap"
  level: "info"
  format: "json"
  output-paths: ["stdout", "logs/apiserver/app.log"]
  otlp:
    enabled: true
    endpoint: "127.0.0.1:4327"       # Local/Docker
    insecure: true
    batch_size: 100
    resource_attributes:
      service.name: "apiserver"
      service.version: "v2.0.0"
      deployment.environment: "development"
```

### Troubleshooting

**VictoriaLogs _msg Field Issue**: If getting `"_msg":"missing _msg field"` error:

1. Check OTEL Collector config mapping: `attributes.msg` → `attributes._msg`
2. Restart collector: `docker restart proj-otelcol`
3. Verify: `curl -s "http://127.0.0.1:9428/select/logsql/query" -d 'query=_msg:*'`

## AI Agent Module (Planned)

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
NEVER modify or repair files within `_*` directories (`_output/`, `_thirdparty/`, `_generated/`) - these contain auto-generated content, build artifacts, and third-party service data that should not be manually edited.
NEVER reference `scripts/` files from `manifests/env/` files - environment configurations must be self-contained to avoid circular dependencies and ensure portability.

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.