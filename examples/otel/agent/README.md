# OTEL Agent Example

演示 OpenTelemetry Agent 层的完整实现，展示应用程序端的遥测数据收集和 SDK 集成。

## 🎯 功能特性

- **完整 SDK 初始化**: Traces, Metrics, Logs 三种遥测数据类型
- **Gateway 连接**: 通过 OTLP 协议连接到 OTEL Gateway
- **结构化数据生成**: 模拟真实业务场景的遥测数据
- **优雅关闭**: 正确的资源清理和数据刷新

## 🚀 运行示例

```bash
cd examples/otel/agent
go run main.go
```

## 📊 示例内容

### 1. ExampleOTELAgentUsage()
基础的 OTEL Agent 使用演示：
- 创建 Agent 配置
- 初始化 SDK 组件
- 生成示例遥测数据
- 优雅关闭

### 2. ApplicationExample()
完整应用程序集成演示：
- 微服务架构场景
- 多请求并发处理
- 数据库、缓存、外部 API 调用模拟
- 完整的业务工作流

## 🔧 配置说明

```go
config := OTELAgentConfig{
    ServiceName:         "my-microservice",
    ServiceVersion:      "v2.1.0", 
    Environment:         "production",
    GatewayHTTPEndpoint: "http://127.0.0.1:4328",  // HTTP 端点
    GatewayGRPCEndpoint: "http://127.0.0.1:4327",  // gRPC 端点
    EnableLogs:          true,
    EnableMetrics:       true,
    EnableTraces:        true,
}
```

## 📈 验证数据

运行后可在以下位置查看数据：

- **Jaeger 追踪**: http://127.0.0.1:16686
- **VictoriaLogs**: http://127.0.0.1:9428/select/vmui/
- **Gateway 指标**: `curl http://127.0.0.1:8888/metrics`

## 🔍 故障排除

如果遇到连接问题：

1. 检查 OTEL Gateway 状态：`docker ps | grep proj-otel-collector`
2. 验证 Gateway 健康：`curl http://127.0.0.1:13133`
3. 查看 Gateway 日志：`docker logs proj-otel-collector`

## 📚 相关文档

- [OTEL 主文档](../README.md)
- [Gateway 示例](../gateway/)
- [SaaS 集成示例](../saas/)