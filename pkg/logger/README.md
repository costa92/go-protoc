# Logger Package

统一日志记录包，提供简化和详细两种配置方式，支持文件输出、OTLP 协议发送、结构化日志和可观测性集成。同时支持 Zap 和 Slog 后端，具有动态配置管理和智能OTLP自动检测功能。

## 功能特性

- **多种日志器支持**: 支持 Zap 和 Slog（仅 1.24.0+）两种后端
- **灵活配置方式**: 
  - **简化配置**（推荐）: 使用预设模式快速配置
  - **详细配置**: 完全控制所有日志选项
- **预设配置模式**:
  - `development`: 开发环境（控制台输出，彩色，详细信息）
  - `production`: 生产环境（JSON格式，文件输出，结构化）
  - `testing`: 测试环境（简化输出，仅错误级别）
  - **`observability`**: 可观测性（针对 OTLP 发送和日志收集系统优化）
- **智能 OTLP 支持**: 
  - **自动检测**: 配置 `otlp-endpoint` 即自动启用 OTLP
  - **双路输出**: 支持同时发送到 OTEL Collector 和本地文件
  - **智能输出控制**: OTLP 启用时自动禁用文件输出（除非明确配置）
- **条件性文件输出**: 
  - 当配置 `otlp-endpoint` 时，默认不写入文件
  - 当 `log-dir` 为空时，不写入文件
  - 测试模式下，不写入文件
- **结构化日志**: JSON 格式输出，便于 OTEL Collector 解析和分析
- **版本信息集成**: 自动添加服务名、版本、分支等信息到日志
- **热重载配置**: 支持配置文件变更时动态重新加载
- **错误恢复**: 配置错误时自动降级到默认配置

## 架构设计

### OTLP 日志收集架构

```
应用程序 → OTLP Exporter → OTEL Agent (4327) → OTEL Collector (4317) → VictoriaLogs (9428)
     ↘
       本地文件 (可选，根据配置)
```

### 智能输出策略

- **OTLP 优先**: 当配置 OTLP 端点时，优先使用 OTLP 发送，禁用文件输出
- **文件备用**: 当未配置 OTLP 或需要双路输出时，写入本地文件
- **条件控制**: 通过 `log-dir` 和 `otlp-endpoint` 智能控制输出路径

## 快速开始

### 1. 简化配置方式（推荐）

#### 基础 OTLP 配置

```yaml
# configs/apiserver.yaml
log:
  preset: "observability"        # 可观测性预设，适合 OTLP 发送
  type: "zap"                   # 日志器类型
  level: "info"                 # 日志级别
  
  # OTLP 配置（简化方式）- 有端点则自动启用
  otlp-endpoint: "127.0.0.1:4327" # OTEL Agent gRPC 端点（设置则自动启用OTLP）
```

#### 本地文件配置

```yaml
# 仅本地文件输出
log:
  preset: "development"         # 开发环境预设
  type: "zap"
  level: "debug"
  log-dir: "logs"              # 设置日志目录，启用文件输出
```

#### 代码使用

```go
package main

import (
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

func main() {
    // 使用简化配置创建日志器
    config := &logger.QuickConfig{
        Preset:       logger.PresetObservability,
        Type:         "zap",
        Level:        "info",
        OTLPEndpoint: "127.0.0.1:4327", // 自动启用OTLP
    }
    
    // 转换为完整配置
    fullOptions := config.ToFullOptions()
    
    // 创建日志器
    log, err := logger.NewLogger(fullOptions)
    if err != nil {
        panic(err)
    }
    
    // 使用日志器
    log.Info("服务启动")
    log.Infow("用户操作", "user_id", 12345, "action", "login")
}
```

### 2. 详细配置方式

当需要精细控制时，可使用详细配置：

```yaml
log:
  type: "zap"
  level: "info"
  format: "json"
  disable-caller: false
  output-paths: ["stdout", "logs/apiserver/app.log"]
  error-output-paths: ["stderr"]
  development: false
  encoding: "json"

  # 详细的 OTLP 配置
  otlp:
    enabled: true
    endpoint: "127.0.0.1:4327"
    protocol: "grpc"
    timeout: "10s"
    batch_timeout: "2s"
    batch_size: 100
    insecure: true
    service_name: "apiserver"
    service_version: "v1.0.0"
    environment: "development"
```

## 配置预设说明

### PresetObservability (可观测性)

专为OTLP发送和日志收集系统优化：

```go
&LogsOptions{
    Type:     LoggerTypeZap,
    Level:    "info",
    Format:   "json",           // 结构化JSON便于解析
    Encoding: "json",
    EncoderConfig: &EncoderConfig{
        TimeKey:      "ts",      // 与OTEL配置兼容
        LevelKey:     "level",
        MessageKey:   "msg",     // 会被重映射为_msg
        CallerKey:    "caller",
        TimeEncoder:  "iso8601",
        LevelEncoder: "lowercase",
    },
    InitialFields: map[string]interface{}{
        "service.name":    serviceName,
        "service.version": "v2.0.0",
    },
}
```

### PresetDevelopment (开发环境)

```go
&LogsOptions{
    Type:          LoggerTypeZap,
    Level:         "debug",
    Format:        "console",    // 便于阅读
    EnableColor:   true,
    Development:   true,
    OutputPaths:   []string{"stdout"},
}
```

### PresetProduction (生产环境)

```go
&LogsOptions{
    Type:                  LoggerTypeZap,
    Level:                 "info",
    Format:                "json",
    EnableColor:           false,
    Development:           false,
    DisableFunctionAtInfo: true,  // 优化性能
    MaxSize:               100,
    MaxAge:                7,
    Compress:              true,
}
```

## 智能输出控制

### 输出决策逻辑

```go
// 文件输出条件
enableFileOutput := c.Preset != PresetTesting &&    // 非测试模式
                   !enableOTLP &&                   // OTLP未启用
                   c.LogDir != ""                   // log-dir已配置

if enableFileOutput {
    // 添加文件输出路径
    logFile := filepath.Join(c.LogDir, serviceName, "app.log")
    opts.OutputPaths = append(opts.OutputPaths, logFile)
}
```

### 配置组合效果

| otlp-endpoint | log-dir | 输出路径 | 说明 |
|---------------|---------|----------|------|
| 已配置 | 任意 | stdout | OTLP优先，禁用文件 |
| 未配置 | 已配置 | stdout + file | 双路输出 |
| 未配置 | 未配置 | stdout | 仅控制台输出 |

## OTLP 配置详解

### 简化配置（推荐）

```yaml
log:
  otlp-endpoint: "127.0.0.1:4327"  # 仅此一行即可启用OTLP
```

### 完整配置

```yaml
log:
  otlp:
    enabled: true
    endpoint: "127.0.0.1:4327"
    protocol: "grpc"
    timeout: "10s"
    batch_timeout: "2s"
    batch_size: 100
    insecure: true
    headers:
      x-api-key: "your-api-key"
    service_name: "apiserver"
    service_version: "v1.0.0"
    environment: "development"
```

### 环境特定端点

```yaml
# 本地开发
otlp-endpoint: "127.0.0.1:4327"

# Docker环境
otlp-endpoint: "otelcol:4317"

# Kubernetes
otlp-endpoint: "otelcol.logging.svc.cluster.local:4317"
```

## API 参考

### QuickConfig

简化配置结构：

```go
type QuickConfig struct {
    Preset         LogPreset `json:"preset"`           // 预设模式
    Type           string    `json:"type"`             // 日志器类型: zap, slog
    Level          string    `json:"level"`            // 覆盖预设的日志级别
    LogDir         string    `json:"log-dir"`          // 日志目录
    OTLPEndpoint   string    `json:"otlp-endpoint"`    // OTLP端点（设置则自动启用OTLP）
}
```

### 转换方法

```go
// 转换为完整配置
func (c *QuickConfig) ToFullOptions() *LogsOptions

// 从简化配置创建日志器
func NewLoggerFromQuickConfig(config *QuickConfig) (Logger, error)

// 默认简化配置
func DefaultQuickConfig() *QuickConfig
```

### Logger 接口

```go
type Logger interface {
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})

    Debugf(template string, args ...interface{})
    Infof(template string, args ...interface{})
    Warnf(template string, args ...interface{})
    Errorf(template string, args ...interface{})
    Fatalf(template string, args ...interface{})

    Debugw(msg string, keysAndValues ...interface{})
    Infow(msg string, keysAndValues ...interface{})
    Warnw(msg string, keysAndValues ...interface{})
    Errorw(msg string, keysAndValues ...interface{})
    Fatalw(msg string, keysAndValues ...interface{})

    With(keyValues ...interface{}) Logger
    WithCtx(ctx context.Context, keyValues ...interface{}) Logger
    WithCallerSkip(skip int) Logger
    SetLevel(level Level)
}
```

## 使用场景

### 1. 微服务开发环境

```yaml
log:
  preset: "development"
  type: "zap"
  level: "debug"
  log-dir: "logs"  # 本地文件便于调试
```

### 2. 生产环境 - OTLP 收集

```yaml
log:
  preset: "observability"
  type: "zap" 
  level: "info"
  otlp-endpoint: "otelcol.logging.svc.cluster.local:4317"
```

### 3. 混合架构 - 双路输出

通过详细配置实现同时输出到文件和OTLP：

```yaml
log:
  type: "zap"
  level: "info"
  format: "json"
  output-paths: ["stdout", "logs/app.log"]  # 明确指定文件输出
  
  otlp:
    enabled: true
    endpoint: "127.0.0.1:4327"
```

### 4. 测试环境

```yaml
log:
  preset: "testing"  # 自动仅错误级别，无文件输出
  type: "zap"
```

## 全局日志器函数

便利的全局函数，使用默认日志器：

```go
import "github.com/costa92/go-protoc/v2/pkg/logger"

// 基础日志
logger.Info("应用启动")
logger.Error("连接失败")

// 格式化日志
logger.Infof("用户 %s 登录", username)

// 结构化日志
logger.Infow("用户操作", "user_id", 123, "action", "login")

// 带上下文的日志器
contextLogger := logger.With("service", "auth", "version", "v1.0.0")
contextLogger.Info("服务初始化")
```

## 性能考虑

### OTLP vs 文件输出

| 输出方式 | 延迟 | 吞吐量 | 可靠性 | 适用场景 |
|----------|------|--------|--------|----------|
| OTLP gRPC | 低 | 高 | 高 | 生产环境，集中收集 |
| 本地文件 | 极低 | 极高 | 中 | 开发调试，离线分析 |
| 双路输出 | 中 | 中 | 高 | 关键服务，容错需求 |

### 批处理优化

OTLP 支持批处理发送以提高性能：

```yaml
otlp:
  batch_size: 500        # 批量大小
  batch_timeout: "2s"    # 批处理超时
```

## 故障排除

### 常见问题

1. **OTLP连接失败**
   ```
   ✅ Logger created with OTLP support
   ❌ Failed to send logs via OTLP: connection refused
   ```
   - 检查 OTEL Agent/Collector 是否运行
   - 验证端点地址和端口

2. **文件权限错误**
   ```
   ⚠️ Warning: failed to create log directories
   ```
   - 检查日志目录权限
   - 确保应用有写入权限

3. **配置解析错误**
   ```
   ❌ Failed to create logger: invalid level
   ```
   - 检查日志级别拼写
   - 验证YAML格式

### 调试模式

启用详细日志查看配置过程：

```yaml
log:
  level: "debug"  # 显示详细配置信息
```

## 迁移指南

### 从旧版本迁移

**旧配置**（废弃）：
```yaml
log:
  enable-otlp: true
  otlp-endpoint: "127.0.0.1:4327"
```

**新配置**（推荐）：
```yaml
log:
  otlp-endpoint: "127.0.0.1:4327"  # 自动启用OTLP
```

## 相关文档

- [设计文档](DESIGN.md) - 架构设计和实现细节
- [示例文档](EXAMPLES.md) - 完整使用示例
- [OTLP 集成指南](../../configs/apiserver_otlp.yaml) - OTLP 配置示例
- [可观测性文档](../../docs/logging-architecture.md) - 完整日志收集架构

## 依赖项

- `go.uber.org/zap` - Zap 日志器后端
- `gopkg.in/natefinch/lumberjack.v2` - 日志轮转
- `go.opentelemetry.io/otel` - OpenTelemetry 支持
- `log/slog` - Slog 日志器后端（标准库）