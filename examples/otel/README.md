# OpenTelemetry (OTEL) Examples

完整的 OpenTelemetry **Agent → Gateway → SaaS** 架构示例实现。

## 📁 目录结构

```
examples/otel/
├── README.md                   # 本文档
├── agent/                      # OTEL Agent 层实现
│   ├── README.md               # Agent 专项文档
│   └── main.go                 # Agent 示例代码
├── gateway/                    # OTEL Gateway 管理  
│   ├── README.md               # Gateway 专项文档
│   └── main.go                 # Gateway 示例代码
├── saas/                       # SaaS 平台集成
│   ├── README.md               # SaaS 专项文档
│   └── main.go                 # SaaS 示例代码
├── complete/                   # 端到端完整演示
│   ├── README.md               # Complete 专项文档
│   └── main.go                 # 完整示例代码
├── simple-demo/                # 简单 HTTP 演示
│   ├── README.md               # Simple 专项文档
│   └── main.go                 # HTTP 协议示例
├── grpc-demo/                  # gRPC 协议演示
│   ├── README.md               # gRPC 专项文档
│   └── main.go                 # gRPC 协议示例
└── agent-collector-demo/       # Agent-Collector 联合使用演示
    ├── README.md               # Agent-Collector 专项文档
    └── main.go                 # 联合使用对比示例
```

## 🏗️ 架构概览

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   OTEL Agent    │───▶│  OTEL Gateway   │───▶│  SaaS Platforms │
│   (agent/)      │    │   (gateway/)    │    │    (saas/)      │
│                 │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               ↑
                    ┌─────────────────┐
                    │ Complete Demo   │
                    │  (complete/)    │
                    └─────────────────┘
                             ↑
               ┌──────────────────────────────────────────────────┐
               │                 Quick Start                      │
               │ ┌──────────────┐ ┌─────────────┐ ┌─────────────┐ │
               │ │ simple-demo/ │ │ grpc-demo/  │ │agent-collec │ │
               │ │   (HTTP)     │ │   (gRPC)    │ │ tor-demo/   │ │
               │ └──────────────┘ └─────────────┘ │(架构对比)   │ │
               └─────────────────────────────────────┴─────────────┘
```

## 🚀 快速开始

### 1. 部署 OTEL 基础设施

```bash
# 部署完整的 OTEL 栈
make deploy.install.docker.otel

# 验证部署状态
docker ps | grep proj-
```

### 2. 运行示例

#### Agent-Collector 联合使用演示 (推荐新手)

```bash
cd examples/otel/agent-collector-demo
go run main.go
```

#### 完整端到端演示 (推荐进阶)

```bash
cd examples/otel/complete
go run main.go
```

#### 单独组件演示

```bash
# Agent 层示例
go run -c "
package main
import \"github.com/costa92/go-protoc/v2/examples/otel\"
func main() { 
    otel.ExampleOTELAgentUsage()
}
"

# Gateway 管理示例  
go run -c "
package main
import \"github.com/costa92/go-protoc/v2/examples/otel\"
func main() { 
    otel.ExampleGatewayUsage() 
}
"

# SaaS 集成示例
go run -c "
package main  
import \"github.com/costa92/go-protoc/v2/examples/otel\"
func main() { 
    otel.ExampleSaaSIntegration() 
}
"
```

## 📊 组件说明

### Agent (agent.go)
- **功能**: 应用程序端遥测 SDK 集成
- **特性**: 
  - 完整的 SDK 初始化 (traces, metrics, logs)
  - 与 Gateway 的 OTLP 连接
  - 结构化遥测数据生成
  - 优雅关闭处理

### Gateway (gateway.go) 
- **功能**: 集中式数据收集和处理中心
- **特性**:
  - 健康状态监控
  - 处理指标统计
  - 配置管理
  - 性能监控

### SaaS (saas.go)
- **功能**: 多 SaaS 平台数据集成
- **特性**:
  - VictoriaLogs 日志查询
  - Jaeger 链路追踪分析  
  - Prometheus 指标监控
  - 跨平台数据关联

### Complete (complete.go)
- **功能**: 端到端完整工作流演示
- **特性**:
  - 多业务场景模拟
  - 实时数据流监控
  - 自动化集成测试
  - 完整验证流程

## 🔧 配置说明

### OTEL Agent 配置
```go
config := OTELAgentConfig{
    ServiceName:         "your-service",
    ServiceVersion:      "v1.0.0", 
    Environment:         "production",
    GatewayHTTPEndpoint: "http://127.0.0.1:4328", // OTLP HTTP
    GatewayGRPCEndpoint: "http://127.0.0.1:4327", // OTLP gRPC
    EnableLogs:          true,
    EnableMetrics:       true,
    EnableTraces:        true,
}
```

### SaaS 平台端点
- **VictoriaLogs**: `http://127.0.0.1:9428` (日志存储与查询)
- **Jaeger**: `http://127.0.0.1:16686` (分布式链路追踪)  
- **Prometheus**: `http://127.0.0.1:9090` (指标存储与告警)

## 📈 验证数据流

### 1. Gateway 处理状态
```bash
# 查看 Gateway 日志
docker logs proj-otelcol

# 检查处理指标
curl http://127.0.0.1:8888/metrics | grep otelcol
```

### 2. SaaS 平台数据验证

#### VictoriaLogs
```bash
# Web 界面
http://127.0.0.1:9428/select/vmui/

# API 查询
curl -s "http://127.0.0.1:9428/select/logsql/query" \
  -d 'query=service.name:"your-service"'
```

#### Jaeger
```bash  
# Web 界面
http://127.0.0.1:16686
# 搜索服务: your-service
```

#### Prometheus
```bash
# Web 界面  
http://127.0.0.1:9090
# 查询: {service="your-service"}
```

## 🛠️ 自定义开发

### 添加自定义遥测

```go
// 自定义追踪
tracer := otel.Tracer("my-service")
ctx, span := tracer.Start(ctx, "custom-operation")
span.SetAttributes(attribute.String("custom.key", "value"))
defer span.End()

// 自定义指标
meter := otel.Meter("my-service")
counter, _ := meter.Int64Counter("custom_operations_total")
counter.Add(ctx, 1, attribute.String("type", "custom"))

// 自定义日志
logger.InfoContext(ctx, "Custom operation",
    "operation_id", "12345",
    "result", "success",
)
```

### 添加 SaaS 平台

```go
config := SaaSConfig{
    Platform: SaaSPlatform("new-platform"),
    Config: map[string]interface{}{
        "endpoint": "https://api.new-platform.com",
        "api_key": "your-key",
        "timeout": "30s",
    },
}

manager := NewSaaSManager()
client, err := manager.AddPlatform(config.Platform, config.Config)
```

## 🔍 故障排除

### 常见问题

1. **Gateway 连接失败**
   ```bash
   docker restart proj-otelcol
   curl http://127.0.0.1:13133  # 健康检查
   ```

2. **SaaS 平台无数据**
   ```bash
   # 检查数据流
   curl http://127.0.0.1:8888/metrics | grep -E "(receiver|exporter)"
   ```

3. **高内存占用**
   ```bash
   # 检查资源使用
   docker stats proj-otelcol
   ```

### 调试模式

```bash
export OTEL_LOG_LEVEL=debug
go run complete.go
```

## 📚 相关文档

- [主项目文档](../README.md)
- [OTEL 安装文档](../../scripts/installation/otel/README.md)
- [架构设计文档](../../docs/otel-architecture-design.md)