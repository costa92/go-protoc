# Project Examples

本目录包含项目各个功能模块的完整示例代码，展示生产级别的最佳实践和集成方案。

## 📁 目录结构

```
examples/
├── README.md              # 本文档
├── otel/                  # OpenTelemetry 可观测性示例
│   ├── README.md          # OTEL 专项文档
│   ├── agent.go           # OTEL Agent 层实现
│   ├── gateway.go         # OTEL Gateway 管理
│   ├── saas.go           # SaaS 平台集成
│   └── complete.go       # 端到端完整演示
├── errorsx/              # 错误处理系统示例
│   ├── README.md         # ErrorsX 专项文档
│   └── examples.go       # 错误处理示例
└── database/             # 数据库集成示例
    ├── README.md         # 数据库专项文档
    └── tracing.go        # 数据库链路追踪
```

## 🚀 快速导航

### [🔍 OpenTelemetry 可观测性](./otel/)
完整的 **Agent → Gateway → SaaS** 架构演示

- **核心特性**: 分布式追踪、指标监控、结构化日志
- **技术栈**: OTEL Collector, VictoriaLogs, Jaeger, Prometheus
- **快速开始**: `cd otel && go run complete.go`

### [❌ 错误处理系统](./errorsx/)  
统一的多语言错误处理和 gRPC 集成

- **核心特性**: 国际化错误、结构化错误码、链式操作
- **技术栈**: ErrorsX, i18n, gRPC Status
- **快速开始**: `cd errorsx && go run examples.go`

### [🗄️ 数据库集成](./database/)
数据库操作的链路追踪和性能监控

- **核心特性**: GORM 追踪、Redis 监控、连接池管理
- **技术栈**: MySQL, Redis, OTEL, GORM
- **快速开始**: `cd database && go run tracing.go`

## 🔧 环境准备

运行示例前需要准备相应的基础设施：

### 通用准备

```bash  
# 1. 安装开发工具
make install-tools A=1

# 2. 启动基础服务
./scripts/installation/service.sh start database  # 数据库服务
./scripts/installation/service.sh start redis     # Redis 缓存
```

### OTEL 示例专用

```bash
# 部署完整 OTEL 基础设施
make deploy.install.docker.otel

# 验证服务状态
docker ps | grep proj-
# 应显示: proj-otelcol, proj-victorialogs, proj-jaeger 等
```

### 服务健康检查

```bash
# Gateway 健康检查
curl http://127.0.0.1:13133

# VictoriaLogs 检查  
curl http://127.0.0.1:9428

# Jaeger 检查
curl http://127.0.0.1:14269

# Prometheus 指标
curl http://127.0.0.1:8889/metrics
```

## 🎯 推荐学习路径

### 1. 新手入门 - 错误处理系统
```bash
cd examples/errorsx
go run examples.go
```
- 理解项目错误处理规范
- 学习国际化和错误码设计
- 掌握 gRPC 错误转换

### 2. 进阶应用 - 数据库集成  
```bash
cd examples/database
go run tracing.go
```
- 学习数据库操作最佳实践
- 理解链路追踪集成
- 掌握性能监控方法

### 3. 高级架构 - 可观测性系统
```bash  
cd examples/otel
go run complete.go
```
- 理解微服务可观测性架构
- 学习 OTEL 生产实践
- 掌握监控数据分析

## 📊 验证和监控

### 示例数据验证方式

每个示例模块都提供了相应的数据验证方法：

#### OTEL 示例验证

```bash
# 1. Gateway 处理状态
docker logs proj-otelcol
curl http://127.0.0.1:8888/metrics | grep otelcol

# 2. 查看链路追踪 (Jaeger)
http://127.0.0.1:16686
# 搜索服务: complete-example

# 3. 查看日志 (VictoriaLogs)  
http://127.0.0.1:9428/select/vmui/
# 查询: service.name:"complete-example"

# 4. 查看指标 (Prometheus)
http://127.0.0.1:9090
# 查询: {service="complete-example"}
```

#### 数据库示例验证

```bash
# 查看数据库链路追踪
http://127.0.0.1:16686
# 搜索服务: database-example

# 检查数据库连接状态
./scripts/installation/service.sh status mariadb
./scripts/installation/service.sh status redis
```

#### 错误处理验证

```bash
# 查看错误处理日志输出
# 示例会展示不同语言的错误信息
# 以及 gRPC 错误码转换结果
```

## 🔧 配置说明

### 端口和地址配置

| 服务 | 地址 | 用途 |
|------|------|------|
| OTEL Gateway HTTP | `127.0.0.1:4328` | OTLP HTTP 接收 |  
| OTEL Gateway gRPC | `127.0.0.1:4327` | OTLP gRPC 接收 |
| VictoriaLogs | `127.0.0.1:9428` | 日志存储查询 |
| Jaeger UI | `127.0.0.1:16686` | 链路追踪界面 |
| Prometheus | `127.0.0.1:9090` | 指标存储告警 |
| MySQL | `127.0.0.1:3306` | 数据库服务 |
| Redis | `127.0.0.1:6379` | 缓存服务 |

### 数据处理流程

```
应用程序 → [OTEL Agent] → [OTEL Gateway] → [SaaS 平台]
    ↓              ↓              ↓           ↓
 生成遥测       SDK 收集      统一处理      存储分析
```

## 🛠️ 自定义和扩展

### 示例代码自定义

各个示例都支持自定义配置和扩展：

```go
// OTEL 自定义配置
config := OTELAgentConfig{
    ServiceName:         "my-custom-service",
    ServiceVersion:      "v2.0.0",
    Environment:         "staging",
    GatewayHTTPEndpoint: "http://127.0.0.1:4328",
    EnableLogs:          true,
    EnableMetrics:       true, 
    EnableTraces:        true,
}

// 错误处理自定义
err := errorsx.New("CUSTOM_ERROR", "自定义错误信息")
err = errorsx.WithDetails(err, "custom_field", "value")

// 数据库操作自定义追踪
ctx, span := tracer.Start(ctx, "custom-db-operation")
span.SetAttributes(attribute.String("table", "custom_table"))
defer span.End()
```

## 🔍 故障排除

### 常见问题解决

#### 1. 服务连接失败

```bash
# 检查服务状态
./scripts/installation/service.sh status all

# 重启特定服务
docker restart proj-otelcol  # OTEL Gateway
docker restart proj-mariadb  # MySQL
docker restart proj-redis    # Redis
```

#### 2. 数据未显示

```bash
# OTEL 数据流检查
curl http://127.0.0.1:8888/metrics | grep -E "(receiver|exporter)"
docker logs proj-otelcol

# 数据库连接检查  
./scripts/installation/service.sh status mariadb
mysql -h127.0.0.1 -P3306 -uroot -p'proj(#)666' -e "SELECT 1"
```

#### 3. 性能问题

```bash
# 检查资源使用
docker stats proj-otelcol proj-mariadb proj-redis

# 查看容器日志
docker logs proj-otelcol --tail 50
```

### 调试模式

```bash  
# 启用详细日志
export OTEL_LOG_LEVEL=debug
export DB_LOG_LEVEL=debug
go run <example>.go
```

## 📚 相关资源

### 项目内部文档

- **OTEL 架构**: [架构设计文档](../docs/otel-architecture-design.md)  
- **安装脚本**: [服务管理文档](../scripts/installation/README.md)
- **错误处理**: [ErrorsX 包文档](../pkg/errorsx/README.md)
- **数据库包**: [数据库抽象文档](../pkg/db/README.md)

### 外部参考资源

- [OpenTelemetry Go 官方文档](https://opentelemetry.io/docs/go/)
- [GORM 官方文档](https://gorm.io/docs/)
- [VictoriaLogs 查询语法](https://docs.victoriametrics.com/VictoriaLogs/LogsQL.html)
- [Jaeger 追踪分析](https://www.jaegertracing.io/docs/)
- [Prometheus 查询语言](https://prometheus.io/docs/prometheus/latest/querying/)

## 🚀 性能建议

### 生产环境优化

```yaml
# 高负载生产配置建议
processors:
  batch:
    send_batch_size: 500    # 批处理大小
    timeout: 1s             # 批处理超时
  probabilistic_sampler:
    sampling_percentage: 1   # 1% 采样率
  memory_limiter:
    limit_mib: 512          # 内存限制
```

### 监控最佳实践

1. **合理采样**: 高流量环境使用概率采样
2. **批量处理**: 根据流量配置批处理大小  
3. **资源限制**: 设置内存和CPU限制
4. **网络优化**: 高吞吐场景使用 gRPC

## 🤝 贡献指南

添加新示例时请遵循：

1. **代码规范**: 遵循项目现有代码结构和模式
2. **文档完善**: 包含详细的文档和注释
3. **验证步骤**: 添加功能验证和测试步骤
4. **更新文档**: 更新相关 README 文档
5. **完整测试**: 确保与现有基础设施兼容

## 📄 许可证

本示例代码遵循项目主仓库的许可证条款。