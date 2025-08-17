# Logger Package

A flexible logging package that supports both Zap and Slog backends with automatic caller information (file, line, function).

## Features

- **Dual Backend Support**: Choose between `go.uber.org/zap` and `log/slog`
- **Automatic Type Identification**: Logs automatically include `type` field to identify the backend (`"zap"` or `"slog"`)
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
| `InitialFields` | `map[string]interface{}` | Initial fields added to all log entries |
| `CallerSkip` | `int` | Number of caller frames to skip |

**Note**: The `type` field is automatically added to all log entries to identify the logger backend (`"zap"` or `"slog"`). This field cannot be overridden by `InitialFields`.

### Initial Fields Configuration

You can add custom fields that will be included in all log entries:

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

logger.Info("Service started")
// Output: {"level":"info","ts":"...","caller":"main.go:15","msg":"Service started","type":"zap","service":"auth-service","version":"v1.2.3","environment":"production"}
```

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
    WithCallerSkip(skip int) Logger
    SetLevel(level Level)
}
```

## Caller Information

Both logger implementations automatically capture:
- **File**: Source file name (e.g., "main.go")
- **Line**: Line number where log was called
- **Function**: Function name where log was called

This information is included in log output and can be configured via `CallerSkip`.

## Logger Type Identification

All loggers automatically include a `type` field in their output to identify which backend is being used:

### Zap Logger Output
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

### Slog Logger Output
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

### Benefits
- **Easy Debugging**: Quickly identify which logger backend generated each log entry
- **Mixed Environments**: Useful when different services use different logger types
- **Monitoring**: Enable filtering and alerting based on logger type
- **Automatic**: No configuration required - the `type` field is added automatically

## Context and Structured Logging

### Adding Context Fields
```go
logger := logger.With("service", "auth", "version", "1.0.0")
logger.Info("Service started")
// Output: {"level":"info","ts":"...","caller":"main.go:10","msg":"Service started","type":"zap","service":"auth","version":"1.0.0"}
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

## Global Logger Functions

The package provides convenient global functions that use the default logger:

### Basic Logging Functions
```go
import "github.com/costa92/go-protoc/v2/pkg/logger"

// Simple logging
logger.Info("Application started")
logger.Error("Connection failed")

// Formatted logging
logger.Infof("User %s logged in at %v", username, time.Now())
logger.Errorf("Failed to process %d items", count)

// Structured logging
logger.Infow("User action", "user_id", 123, "action", "login", "success", true)
logger.Errorw("Database error", "error", err, "query", "SELECT * FROM users")
```

### Context and Enhanced Logging
```go
import (
    "context"
    "github.com/costa92/go-protoc/v2/pkg/logger"
)

// Create logger with additional context
contextLogger := logger.With("service", "auth", "version", "v1.0.0")
contextLogger.Info("Service initialized")

// Context-aware logging
ctx := context.Background()
requestLogger := logger.WithCtx(ctx, "request_id", "req-123", "user_id", 456)
requestLogger.Info("Processing request")
```

All global functions automatically include the `type` field and proper caller information.

## Usage Examples

See `example_test.go` for comprehensive usage examples including:
- Basic logging with both backends
- Structured logging
- Context-aware logging
- Default logger usage
- LoggerImpl wrapper usage
- Level parsing and conversion
- Global logger functions
