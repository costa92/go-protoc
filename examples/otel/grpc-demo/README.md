# OTEL gRPC Demo

使用 gRPC 协议的 OpenTelemetry 演示，展示高性能的追踪数据传输。

## 🎯 功能特性

- **gRPC 协议**: 使用 OTLP gRPC 导出器 (端口 4327)
- **高性能传输**: gRPC 提供更好的网络性能
- **安全连接**: 支持 TLS 和非安全连接配置
- **生产就绪**: 适合高吞吐量场景

## 🚀 运行示例

```bash
cd examples/otel/grpc-demo
go run main.go
```

## 📊 示例内容

### 生成的业务场景
1. **user-registration** - 用户注册流程
2. **login-flow** - 登录验证流程  
3. **data-processing** - 数据处理流程
4. **api-call** - API 调用流程
5. **cache-update** - 缓存更新流程

### 每个场景包含的操作
- **input-validation** (30ms) - 输入验证
- **business-processing** (80ms) - 业务处理
- **data-storage** (60ms) - 数据存储
- **response-generation** (20ms) - 响应生成

## 🔧 技术细节

### OTLP gRPC 配置
```go
exporter, err := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("127.0.0.1:4327"),
    otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
)
```

### gRPC 连接选项
- **端点**: `127.0.0.1:4327` (项目配置的 gRPC 端口)
- **安全**: 使用 `insecure.NewCredentials()` 用于演示
- **超时**: 自动处理连接超时和重试

### 性能优化配置
```go
tp := sdktrace.NewTracerProvider(
    sdktrace.WithBatcher(exporter,
        sdktrace.WithBatchTimeout(5*time.Second),   // 批处理超时
        sdktrace.WithMaxExportBatchSize(100),       // 批处理大小
    ),
    sdktrace.WithResource(res),
    sdktrace.WithSampler(sdktrace.AlwaysSample()),   // 采样配置
)
```

## 🚀 gRPC vs HTTP 性能对比

| 特性 | gRPC (4327) | HTTP (4328) |
|------|-------------|-------------|
| **协议** | HTTP/2 | HTTP/1.1 |
| **序列化** | Protobuf | JSON |
| **压缩** | 内置 gzip | 可选 |
| **多路复用** | ✅ 支持 | ❌ 不支持 |
| **延迟** | 更低 | 较高 |
| **吞吐量** | 更高 | 较低 |

## 📈 验证数据

### 1. 检查 Gateway gRPC 接收状态
```bash
docker logs proj-otel-collector --tail 20 | grep grpc
curl http://127.0.0.1:8888/metrics | grep grpc
```

### 2. 查看 Jaeger 中的追踪数据
- 打开: http://127.0.0.1:16686
- 服务名: `otel-grpc-demo`
- 查找操作: `user-registration`, `login-flow` 等

### 3. 性能监控
```bash
# 监控 gRPC 连接状态
curl http://127.0.0.1:8888/metrics | grep otlp_receiver

# 检查处理延迟
curl http://127.0.0.1:8888/metrics | grep duration
```

## 🔧 高级配置

### 生产环境 gRPC 配置
```go
// TLS 安全连接
creds, err := credentials.NewClientTLSFromFile("cert.pem", "")
exporter, err := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("gateway.company.com:4327"),
    otlptracegrpc.WithTLSCredentials(creds),
)

// 自定义 gRPC 选项
dialOpts := []grpc.DialOption{
    grpc.WithKeepaliveParams(keepalive.ClientParameters{
        Time:                10 * time.Second,
        Timeout:             time.Second,
        PermitWithoutStream: true,
    }),
    grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(100 * 1024 * 1024)),
}

exporter, err := otlptracegrpc.New(ctx,
    otlptracegrpc.WithEndpoint("127.0.0.1:4327"),
    otlptracegrpc.WithDialOption(dialOpts...),
)
```

### 负载均衡配置
```go
// 多个 Gateway 端点
endpoints := []string{
    "gateway-1.company.com:4327",
    "gateway-2.company.com:4327",
    "gateway-3.company.com:4327",
}

// 使用 gRPC 负载均衡
conn, err := grpc.Dial(
    "dns:///gateway.company.com:4327",
    grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
)
```

## 🔍 故障排除

### gRPC 特有问题

1. **认证握手失败**
   ```bash
   # 错误: authentication handshake failed: EOF
   # 解决: 检查 TLS 配置或使用 insecure 连接
   ```

2. **连接超时**
   ```bash
   # 错误: context deadline exceeded
   # 解决: 调整超时设置或检查网络连接
   ```

3. **端口连接问题**
   ```bash
   # 检查 gRPC 端口是否开放
   telnet 127.0.0.1 4327
   
   # 检查容器端口映射
   docker port proj-otel-collector | grep 4327
   ```

### 调试工具
```bash
# 使用 grpcurl 测试连接
grpcurl -plaintext 127.0.0.1:4327 list

# 检查 gRPC 服务状态
grpcurl -plaintext 127.0.0.1:4327 grpc.health.v1.Health/Check
```

## 📊 性能基准

### 吞吐量测试
```go
// 高负载场景测试
func BenchmarkGRPCExporter(b *testing.B) {
    // 创建大量 spans
    for i := 0; i < b.N; i++ {
        generateSpan()
    }
    // 测量导出性能
}
```

### 内存使用监控
```bash
# 监控内存使用
watch -n 5 'docker stats proj-otel-collector --no-stream'

# 检查 gRPC 连接池
curl http://127.0.0.1:8888/metrics | grep connection_pool
```

## 🔗 适用场景

### 推荐使用 gRPC 的场景
- ✅ 高吞吐量环境 (>1000 traces/sec)
- ✅ 低延迟要求 (<10ms)
- ✅ 微服务间通信
- ✅ 生产环境部署

### HTTP 更适合的场景
- ✅ 简单测试和开发
- ✅ 防火墙限制环境
- ✅ 调试和排错
- ✅ 浏览器环境

## 📚 相关文档

- [OTEL 主文档](../README.md)
- [Simple Demo](../simple-demo/) - HTTP 版本对比
- [Complete Demo](../complete/) - 完整架构示例