# Logger 日志包

一个灵活的日志包，支持 Zap 和 Slog 后端，具有自动调用者信息捕获功能（文件、行号、函数名），并支持动态配置管理。

## 主要特性

- **双后端支持**: 可选择使用 `go.uber.org/zap` 或 `log/slog`
- **动态配置**: 支持运行时调整日志配置，无需重启应用
- **自动类型标识**: 日志自动包含 `type` 字段标识后端类型（`"zap"` 或 `"slog"`）
- **调用者信息**: 自动捕获文件路径、行号和函数名
- **灵活配置**: 为两种日志器提供全面的配置选项
- **结构化日志**: 支持键值对和格式化消息
- **日志轮转**: 内置 lumberjack 支持
- **上下文支持**: 支持上下文感知的日志记录和请求追踪
- **过滤器支持**: 可配置的日志过滤功能

## 快速开始

### 基础用法

```go
package main

import "github.com/costa92/go-protoc/v2/pkg/logger"

func main() {
    // 使用默认日志器（基于 Zap）
    logger.Info("Hello, World!")
    logger.Infof("用户 %s 已登录", "john")
    logger.Infow("用户登录", "user", "john", "ip", "192.168.1.1")
}
```

### 创建自定义日志器

#### 静态 Zap 日志器

```go
opts := &logger.LogsOptions{
    Type:              logger.LoggerTypeZap,
    Dynamic:           false, // 禁用动态配置（默认）
    Level:             "info",
    Format:            "json",
    DisableCaller:     false,
    DisableStacktrace: false,
    EnableColor:       true,
    OutputPaths:       []string{"stdout", "app.log"},
    Development:       true,
    CallerSkip:        1,
}

zapLogger, err := logger.NewLogger(opts)
if err != nil {
    panic(err)
}
```

#### 动态 Zap 日志器

```go
opts := &logger.LogsOptions{
    Type:              logger.LoggerTypeZap,
    Dynamic:           true, // 启用动态配置
    Level:             "info",
    Format:            "json",
    OutputPaths:       []string{"stdout", "app.log"},
    Development:       true,
}

// 返回 DynamicLogger，包装了 ZapLogger
dynamicLogger, err := logger.NewLogger(opts)
if err != nil {
    panic(err)
}

// 正常使用
dynamicLogger.Info("应用启动")

// 运行时调整日志级别（无需重启）
dynamicLogger.(*logger.DynamicLogger).UpdateLevel("debug")
dynamicLogger.Debug("现在可以看到调试信息")

// 运行时完全切换配置
newOpts := &logger.LogsOptions{
    Type:   logger.LoggerTypeSlog, // 切换到 Slog
    Level:  "warn",
    Format: "text",
    OutputPaths: []string{"stdout"},
}
dynamicLogger.(*logger.DynamicLogger).UpdateConfig(newOpts)
```

#### Slog 日志器

```go
opts := &logger.LogsOptions{
    Type:          logger.LoggerTypeSlog,
    Dynamic:       false, // 静态配置
    Level:         "debug",
    Format:        "text",
    DisableCaller: false,
    OutputPaths:   []string{"stdout"},
    CallerSkip:    1,
}

slogLogger, err := logger.NewLogger(opts)
if err != nil {
    panic(err)
}
```

## 动态配置功能

### 核心优势

- **零停机配置**: 运行时调整，无需重启应用
- **实时调试**: 临时开启详细日志，问题解决后立即关闭
- **性能优化**: 在不同logger实现间动态切换
- **线程安全**: 支持并发读写操作

### 使用场景

#### 1. 故障排查

```go
// 创建动态日志器
opts := &logger.LogsOptions{
    Type:    logger.LoggerTypeZap,
    Dynamic: true,
    Level:   "info",
}
dynamicLogger, _ := logger.NewLogger(opts)

// 平时使用 INFO 级别
dynamicLogger.Info("正常运行")

// 发现问题时，动态开启 DEBUG（无需重启服务）
dynamicLogger.(*logger.DynamicLogger).UpdateLevel("debug")
dynamicLogger.Debug("详细故障排查信息")

// 问题解决后，调回 INFO 级别
dynamicLogger.(*logger.DynamicLogger).UpdateLevel("info")
```

#### 2. 性能调优

```go
// 开发阶段使用 Slog（标准库，调试友好）
dynamicLogger.(*logger.DynamicLogger).UpdateConfig(&logger.LogsOptions{
    Type:   logger.LoggerTypeSlog,
    Level:  "debug",
    Format: "text",
})

// 生产环境动态切换到 Zap（高性能）
dynamicLogger.(*logger.DynamicLogger).UpdateConfig(&logger.LogsOptions{
    Type:          logger.LoggerTypeZap, 
    Level:         "info",
    Format:        "json",
    DisableCaller: true, // 进一步优化性能
})
```

#### 3. 输出重定向

```go
// 正常输出到文件
dynamicLogger.(*logger.DynamicLogger).UpdateConfig(&logger.LogsOptions{
    Type:        logger.LoggerTypeZap,
    OutputPaths: []string{"app.log"},
})

// 紧急情况输出到控制台便于观察
dynamicLogger.(*logger.DynamicLogger).UpdateConfig(&logger.LogsOptions{
    Type:        logger.LoggerTypeZap,
    OutputPaths: []string{"stdout"},
    Format:      "console", // 更易读的格式
})
```

### HTTP 接口动态控制

```go
// 提供 HTTP 接口动态调整日志级别
http.HandleFunc("/admin/log-level", func(w http.ResponseWriter, r *http.Request) {
    level := r.URL.Query().Get("level")
    if dynamicLogger, ok := logger.GetDefaultLogger().(*logger.DynamicLogger); ok {
        err := dynamicLogger.UpdateLevel(level)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }
        fmt.Fprintf(w, "日志级别已更新为: %s", level)
    } else {
        http.Error(w, "当前logger不支持动态配置", http.StatusBadRequest)
    }
})
```

## 配置选项

### LogsOptions

| 字段 | 类型 | 说明 |
|------|------|------|
| `Type` | `LoggerType` | 日志器类型: `LoggerTypeZap` 或 `LoggerTypeSlog` |
| `Dynamic` | `bool` | 启用动态配置功能 |
| `Level` | `string` | 日志级别: "debug", "info", "warn", "error", "fatal" |
| `Format` | `string` | 输出格式: "json" 或 "text" |
| `DisableCaller` | `bool` | 禁用调用者信息 |
| `DisableStacktrace` | `bool` | 禁用堆栈跟踪 |
| `EnableColor` | `bool` | 启用彩色输出 |
| `OutputPaths` | `[]string` | 输出目标 |
| `ErrorOutputPaths` | `[]string` | 错误输出目标 |
| `Development` | `bool` | 开发模式 |
| `Encoding` | `string` | 编码格式 |
| `InitialFields` | `map[string]interface{}` | 添加到所有日志条目的初始字段 |
| `CallerSkip` | `int` | 跳过的调用者帧数 |

**注意**: `type` 字段会自动添加到所有日志条目中，用于标识日志器后端（`"zap"` 或 `"slog"`）。此字段不能被 `InitialFields` 覆盖。

### 初始字段配置

您可以添加自定义字段，这些字段将包含在所有日志条目中：

```go
opts := &logger.LogsOptions{
    Type:        logger.LoggerTypeZap,
    Level:       "info",
    Format:      "json",
    OutputPaths: []string{"stdout"},
    InitialFields: map[string]interface{}{
        "service":     "auth-service",
        "version":     "v1.2.3",
        "environment": "production",
    },
}

logger, err := logger.NewLogger(opts)
if err != nil {
    panic(err)
}

logger.Info("服务启动")
// 输出: {"level":"info","ts":"...","caller":"main.go:15","msg":"服务启动","type":"zap","service":"auth-service","version":"v1.2.3","environment":"production"}
```

### 日志轮转（通过 lumberjack）

```go
opts := &logger.LogsOptions{
    Type:        logger.LoggerTypeZap,
    OutputPaths: []string{"app.log"},
    MaxSize:     100,    // MB
    MaxAge:      7,      // 天
    MaxBackups:  3,      // 文件数
    Compress:    true,   // 压缩轮转文件
}
```

## 日志器接口

所有日志器都实现 `Logger` 接口：

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

## 调用者信息

两种日志器实现都会自动捕获：

- **文件**: 源文件名（如 "main.go"）
- **行号**: 调用日志的行号
- **函数**: 调用日志的函数名

这些信息包含在日志输出中，可通过 `CallerSkip` 配置。

## 日志器类型标识

所有日志器在输出中自动包含 `type` 字段，用于标识使用的后端：

### Zap 日志器输出

```json
{
  "level": "info",
  "ts": "2025-08-17T23:50:30.449+0800",
  "caller": "main.go:10",
  "func": "main.main",
  "msg": "Hello, World!",
  "type": "zap"
}
```

### Slog 日志器输出

```json
{
  "time": "2025-08-17T23:50:20.203563+08:00",
  "level": "INFO",
  "source": {
    "function": "main.main",
    "file": "/path/to/main.go",
    "line": 10
  },
  "msg": "Hello, World!",
  "type": "slog"
}
```

### 优势

- **便于调试**: 快速识别每个日志条目由哪个日志器后端生成
- **混合环境**: 当不同服务使用不同日志器类型时很有用
- **监控**: 基于日志器类型进行过滤和告警
- **自动化**: 无需配置 - `type` 字段自动添加

## 上下文和结构化日志

### 添加上下文字段

```go
logger := logger.With("service", "auth", "version", "1.0.0")
logger.Info("服务启动")
// 输出: {"level":"info","ts":"...","caller":"main.go:10","msg":"服务启动","type":"zap","service":"auth","version":"1.0.0"}
```

### 上下文感知日志

```go
ctx := context.Background()
logger := logger.WithCtx(ctx, "request_id", "abc123")
logger.Info("处理请求")
```

### 结构化日志

```go
logger.Infow("用户操作",
    "user_id", 12345,
    "action", "login",
    "ip", "192.168.1.1",
    "success", true,
)
```

## 过滤器支持

日志器支持基于日志元数据的过滤：

```go
type Metadata struct {
    File string `expr:"file"`
    Line int    `expr:"line"`
    Func string `expr:"func"`
    Uid  uint32 `expr:"uid"`
    Msg  string `expr:"msg"`
}
```

## 日志级别

支持的日志级别及转换：

| 级别 | 字符串 | Zap 级别 | Slog 级别 |
|------|--------|----------|-----------|
| Debug | "debug" | DebugLevel | LevelDebug |
| Info | "info" | InfoLevel | LevelInfo |
| Warn | "warn" | WarnLevel | LevelWarn |
| Error | "error" | ErrorLevel | LevelError |
| Fatal | "fatal" | FatalLevel | LevelError+4 |

## 默认配置

```go
func DefaultOptions() *LogsOptions {
    return &LogsOptions{
        Type:              LoggerTypeZap,
        Dynamic:           false, // 默认禁用动态配置以保持性能
        Level:             "info",
        Format:            "json",
        DisableCaller:     false,
        DisableStacktrace: false,
        EnableColor:       true,
        OutputPaths:       []string{"stdout"},
        ErrorOutputPaths:  []string{"stderr"},
        Development:       false,
        Encoding:          "json",
        CallerSkip:        1,
        EncoderConfig: &EncoderConfig{
            TimeKey:       "ts",
            LevelKey:      "level",
            MessageKey:    "msg",
            CallerKey:     "caller",
            StacktraceKey: "stacktrace",
            FunctionKey:   "func",
            TimeEncoder:   "iso8601",
            LevelEncoder:  "lowercase",
            CallerEncoder: "short",
        },
    }
}
```

## 依赖项

必需的依赖项：

- `go.uber.org/zap` - Zap 日志器后端
- `gopkg.in/natefinch/lumberjack.v2` - 日志轮转
- `log/slog` - Slog 日志器后端（标准库）

添加到您的 `go.mod`：

```bash
go get go.uber.org/zap
go get gopkg.in/natefinch/lumberjack.v2
```

## 全局日志器函数

该包提供便利的全局函数，使用默认日志器：

### 基础日志函数

```go
import "github.com/costa92/go-protoc/v2/pkg/logger"

// 简单日志
logger.Info("应用启动")
logger.Error("连接失败")

// 格式化日志
logger.Infof("用户 %s 在 %v 登录", username, time.Now())
logger.Errorf("处理 %d 个项目失败", count)

// 结构化日志
logger.Infow("用户操作", "user_id", 123, "action", "login", "success", true)
logger.Errorw("数据库错误", "error", err, "query", "SELECT * FROM users")
```

### 上下文和增强日志

```go
import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

// 创建带有额外上下文的日志器
contextLogger := logger.With("service", "auth", "version", "v1.0.0")
contextLogger.Info("服务初始化")

// 上下文感知日志
ctx := context.Background()
requestLogger := logger.WithCtx(ctx, "request_id", "req-123", "user_id", 456)
requestLogger.Info("处理请求")
```

所有全局函数都自动包含 `type` 字段和正确的调用者信息。

## 性能考虑

### 静态 vs 动态日志器

| 类型 | 性能开销 | 动态配置 | 适用场景 |
|------|----------|----------|----------|
| **静态日志器** | 最低 | ❌ | 高性能生产环境，配置固定 |
| **动态日志器** | 轻微 | ✅ | 需要运行时配置调整 |

### 推荐使用策略

- **生产环境**: 默认使用静态日志器（`Dynamic: false`）获得最佳性能
- **开发/测试**: 使用动态日志器（`Dynamic: true`）便于调试
- **关键服务**: 提供管理接口动态开启详细日志进行故障排查

## 使用示例

详细的使用示例请参见 `example_test.go`，包括：

- 两种后端的基础日志
- 结构化日志
- 上下文感知日志
- 默认日志器使用
- LoggerImpl 包装器使用
- 级别解析和转换
- 全局日志器函数
- 动态配置管理

## YAML 配置示例

```yaml
log:
  type: "zap"              # 日志器类型: zap/slog
  dynamic: true            # 启用动态配置
  level: "info"            # 日志级别
  format: "json"           # 输出格式
  disable-caller: false    # 启用调用者信息
  output-paths:            # 输出路径
    - "stdout"
    - "logs/app.log"
  error-output-paths:      # 错误输出路径
    - "stderr"
  development: false       # 生产模式
  max-size: 100           # 日志文件最大大小(MB)
  max-age: 7              # 保留天数
  max-backups: 5          # 最大备份数
  compress: true          # 压缩旧文件
  initial-fields:         # 初始字段
    service: "auth-service"
    version: "v1.0.0"
```