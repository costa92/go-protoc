# Logger Package

A flexible logging package that supports both Zap and Slog backends with automatic caller information (file, line, function).

## Features

- **Dual Backend Support**: Choose between `go.uber.org/zap` and `log/slog`
- **Caller Information**: Automatically captures file path, line number, and function name
- **Flexible Configuration**: Comprehensive options for both loggers
- **Structured Logging**: Support for key-value pairs and formatted messages
- **Log Rotation**: Built-in support with lumberjack
- **Context Support**: Context-aware logging with request tracing
- **Filter Support**: Configurable log filtering capabilities

## Quick Start

### Basic Usage

```go
package main

import "github.com/costa92/go-protoc/v2/pkg/logger"

func main() {
    // Use default logger (Zap-based)
    logger.Info("Hello, World!")
    logger.Infof("User %s logged in", "john")
    logger.Infow("User login", "user", "john", "ip", "192.168.1.1")
}
```

### Creating Custom Loggers

#### Zap Logger
```go
opts := &logger.LogsOptions{
    Type:              logger.LoggerTypeZap,
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

#### Slog Logger
```go
opts := &logger.LogsOptions{
    Type:          logger.LoggerTypeSlog,
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

## Configuration Options

### LogsOptions

| Field | Type | Description |
|-------|------|-------------|
| `Type` | `LoggerType` | Logger type: `LoggerTypeZap` or `LoggerTypeSlog` |
| `Level` | `string` | Log level: "debug", "info", "warn", "error", "fatal" |
| `Format` | `string` | Output format: "json" or "text" |
| `DisableCaller` | `bool` | Disable caller information |
| `DisableStacktrace` | `bool` | Disable stack traces |
| `EnableColor` | `bool` | Enable colored output |
| `OutputPaths` | `[]string` | Output destinations |
| `ErrorOutputPaths` | `[]string` | Error output destinations |
| `Development` | `bool` | Development mode |
| `Encoding` | `string` | Encoding format |
| `CallerSkip` | `int` | Number of caller frames to skip |

### Log Rotation (via lumberjack)

```go
opts := &logger.LogsOptions{
    Type:        logger.LoggerTypeZap,
    OutputPaths: []string{"app.log"},
    MaxSize:     100,    // MB
    MaxAge:      7,      // days
    MaxBackups:  3,      // files
    Compress:    true,   // compress rotated files
}
```

## Logger Interface

All loggers implement the `Logger` interface:

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
    SetLevel(level Level)
}
```

## Caller Information

Both logger implementations automatically capture:
- **File**: Source file name (e.g., "main.go")
- **Line**: Line number where log was called
- **Function**: Function name where log was called

This information is included in log output and can be configured via `CallerSkip`.

## Context and Structured Logging

### Adding Context Fields
```go
logger := logger.With("service", "auth", "version", "1.0.0")
logger.Info("Service started")
// Output: {"level":"info","ts":"...","caller":"main.go:10","msg":"Service started","service":"auth","version":"1.0.0"}
```

### Context-Aware Logging
```go
ctx := context.Background()
logger := logger.WithCtx(ctx, "request_id", "abc123")
logger.Info("Processing request")
```

### Structured Logging
```go
logger.Infow("User action",
    "user_id", 12345,
    "action", "login",
    "ip", "192.168.1.1",
    "success", true,
)
```

## Filter Support

The logger supports filtering based on log metadata:

```go
type Metadata struct {
    File string `expr:"file"`
    Line int    `expr:"line"`
    Func string `expr:"func"`
    Uid  uint32 `expr:"uid"`
    Msg  string `expr:"msg"`
}
```

## Log Levels

Supported log levels with conversions:

| Level | String | Zap Level | Slog Level |
|-------|--------|-----------|------------|
| Debug | "debug" | DebugLevel | LevelDebug |
| Info | "info" | InfoLevel | LevelInfo |
| Warn | "warn" | WarnLevel | LevelWarn |
| Error | "error" | ErrorLevel | LevelError |
| Fatal | "fatal" | FatalLevel | LevelError+4 |

## Default Configuration

```go
func DefaultOptions() *LogsOptions {
    return &LogsOptions{
        Type:              LoggerTypeZap,
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

## Dependencies

Required dependencies:
- `go.uber.org/zap` - For Zap logger backend
- `gopkg.in/natefinch/lumberjack.v2` - For log rotation
- `log/slog` - For Slog logger backend (standard library)

Add to your `go.mod`:
```bash
go get go.uber.org/zap
go get gopkg.in/natefinch/lumberjack.v2
```

## Usage Examples

See `example_test.go` for comprehensive usage examples including:
- Basic logging with both backends
- Structured logging
- Context-aware logging
- Default logger usage
- LoggerImpl wrapper usage
- Level parsing and conversion
