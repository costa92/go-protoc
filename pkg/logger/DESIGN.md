# Logger Package 设计文档

本文档详细描述 Logger 包的架构设计、核心组件、关键决策和实现细节。

## 设计目标

### 核心目标

- **简化配置**: 提供直观的配置方式，减少用户配置复杂度
- **智能化**: 自动检测和优化配置，减少人为错误
- **可观测性优先**: 面向现代可观测性架构设计
- **高性能**: 支持高吞吐量日志处理
- **灵活性**: 支持多种输出方式和部署环境

### 设计原则

1. **渐进式增强**: 默认简单，需要时可以详细配置
2. **约定优于配置**: 智能默认值，减少必要配置
3. **向后兼容**: 支持旧配置方式平滑迁移
4. **单一职责**: 每个组件专注于特定功能
5. **可扩展性**: 易于添加新的输出后端和功能

## 总体架构

### 分层架构

```
┌─────────────────────────────────────────┐
│              Configuration Layer         │ ← 配置管理层
├─────────────────────────────────────────┤
│                Factory Layer            │ ← 工厂创建层
├─────────────────────────────────────────┤
│               Interface Layer           │ ← 统一接口层
├─────────────────────────────────────────┤
│            Implementation Layer         │ ← 具体实现层
│  ┌─────────────┬─────────────────────────┤
│  │   ZapImpl   │        SlogImpl        │ │
│  └─────────────┴─────────────────────────┤
├─────────────────────────────────────────┤
│               Output Layer              │ ← 输出处理层
│  ┌─────────┬─────────┬─────────────────┐ │
│  │  Stdout │  File   │      OTLP       │ │
│  └─────────┴─────────┴─────────────────┘ │
└─────────────────────────────────────────┘
```

### 模块组成

```
pkg/logger/
├── logger.go           # 核心接口定义
├── factory.go          # 日志器工厂
├── presets.go          # 预设配置系统
├── zap_impl.go         # Zap 实现
├── slog_impl.go        # Slog 实现
├── otlp.go             # OTLP 支持
├── options.go          # 配置选项
├── levels.go           # 日志级别处理
├── global.go           # 全局日志器
└── utils.go            # 工具函数
```

## 核心设计

### 1. 双层配置系统

#### QuickConfig（简化配置）

```go
type QuickConfig struct {
    Preset         LogPreset `json:"preset"`           // 预设模式
    Type           string    `json:"type"`             // 日志器类型
    Level          string    `json:"level"`            // 日志级别
    LogDir         string    `json:"log-dir"`          // 日志目录
    OTLPEndpoint   string    `json:"otlp-endpoint"`    // OTLP端点
}
```

**设计理念**：

- **最小化必要配置**: 大多数场景只需要1-2个字段
- **智能默认值**: 基于预设自动填充详细配置
- **自动检测**: 通过 `otlp-endpoint` 自动启用OTLP

#### LogsOptions（详细配置）

```go
type LogsOptions struct {
    // 核心配置
    Type                  LoggerType
    Level                 string
    Format                string

    // 输出配置
    OutputPaths           []string
    ErrorOutputPaths      []string

    // 功能配置
    DisableCaller         bool
    DisableStacktrace     bool
    EnableColor           bool
    Development           bool

    // OTLP配置
    OTLP                  *OTLPConfig

    // 高级配置
    EncoderConfig         *EncoderConfig
    InitialFields         map[string]interface{}
    CallerSkip            int

    // 轮转配置
    MaxSize               int
    MaxAge                int
    MaxBackups            int
    Compress              bool
}
```

**设计理念**：

- **完全控制**: 暴露所有可配置选项
- **向后兼容**: 支持现有详细配置
- **灵活性**: 支持复杂的定制需求

### 2. 预设系统设计

#### 预设枚举

```go
const (
    PresetDevelopment   LogPreset = "development"    // 开发环境
    PresetProduction    LogPreset = "production"     // 生产环境
    PresetTesting       LogPreset = "testing"        // 测试环境
    PresetObservability LogPreset = "observability"  // 可观测性
)
```

#### 预设配置策略

```go
func (c *QuickConfig) createPresetOptions(serviceName string) *LogsOptions {
    switch c.Preset {
    case PresetDevelopment:
        return c.developmentPreset(serviceName)
    case PresetProduction:
        return c.productionPreset(serviceName)
    case PresetTesting:
        return c.testingPreset(serviceName)
    case PresetObservability:
        return c.observabilityPreset(serviceName)
    default:
        return c.developmentPreset(serviceName)
    }
}
```

**每个预设的设计考量**：

- **Development**: 可读性优先，便于本地调试
- **Production**: 性能和结构化优先
- **Testing**: 最小输出，避免测试噪音
- **Observability**: 专为OTLP和日志收集优化

### 3. 智能输出控制

#### 输出决策算法

```go
func (c *QuickConfig) ToFullOptions() *LogsOptions {
    // 检测是否启用OTLP（有端点配置则启用）
    enableOTLP := c.OTLPEndpoint != ""

    // 设置日志文件路径 - 当启用OTLP时，或log-dir为空时，不输出到文件
    if c.Preset != PresetTesting && !enableOTLP && c.LogDir != "" {
        logFile := filepath.Join(c.LogDir, serviceName, "app.log")
        if err := ensureLogDir(filepath.Dir(logFile)); err == nil {
            opts.OutputPaths = append(opts.OutputPaths, logFile)
        }
    }

    return opts
}
```

#### 决策表

| 条件 | OTLP Endpoint | Log Dir | Preset | 文件输出 | 说明 |
|------|---------------|---------|--------|----------|------|
| 1 | ✅ 已配置 | 任意 | 任意 | ❌ | OTLP优先策略 |
| 2 | ❌ 未配置 | ✅ 已配置 | 非Testing | ✅ | 标准文件输出 |
| 3 | ❌ 未配置 | ❌ 未配置 | 任意 | ❌ | 仅控制台输出 |
| 4 | 任意 | 任意 | Testing | ❌ | 测试环境策略 |

**设计原理**：

1. **OTLP优先**: 避免重复存储，减少I/O开销
2. **明确意图**: 用户配置log-dir表明需要文件输出
3. **测试友好**: 测试环境避免文件污染
4. **容错设计**: 目录创建失败不影响启动

### 4. OTLP 集成设计

#### 自动检测机制

```go
// 简化配置中的自动检测
enableOTLP := c.OTLPEndpoint != ""

// 自动创建OTLP配置
if enableOTLP {
    opts.OTLP = &OTLPConfig{
        Enabled:        true,
        Endpoint:       c.OTLPEndpoint,
        Protocol:       "grpc",           // 默认gRPC协议
        Timeout:        5 * time.Second,  // 保守的超时设置
        BatchTimeout:   1 * time.Second,  // 快速批处理
        BatchSize:      100,              // 平衡内存和网络
        Insecure:       true,             // 开发友好
        Headers:        make(map[string]string),
        ServiceName:    serviceName,
        ServiceVersion: versionInfo.GitVersion,
        Environment:    "development",
    }
}
```

#### 配置优化策略

```go
// 根据预设调整 OTLP 配置
if c.Preset == PresetProduction {
    opts.OTLP.Environment = "production"
    opts.OTLP.BatchSize = 500              // 生产环境更大批次
    opts.OTLP.BatchTimeout = 2 * time.Second  // 更长缓冲时间
}
```

**设计考量**：

- **零配置**: 仅需端点即可工作
- **环境感知**: 根据预设调整性能参数
- **版本集成**: 自动注入服务版本信息
- **协议选择**: 默认gRPC，性能更优

## 关键实现细节

### 1. 工厂模式实现

#### 统一工厂接口

```go
func NewLogger(opts *LogsOptions) (Logger, error) {
    // 参数验证
    if opts == nil {
        opts = DefaultOptions()
    }

    // 类型路由
    switch opts.Type {
    case LoggerTypeZap:
        return NewZapLogger(opts)
    case LoggerTypeSlog:
        return NewSlogLogger(opts)
    default:
        return NewZapLogger(opts)  // 默认使用Zap
    }
}
```

#### QuickConfig工厂

```go
func NewLoggerFromQuickConfig(config *QuickConfig) (Logger, error) {
    fullOptions := config.ToFullOptions()
    return NewLogger(fullOptions)
}
```

**设计优势**：

- **统一入口**: 所有创建都通过工厂
- **类型安全**: 编译时检查类型匹配
- **易于扩展**: 添加新实现只需修改工厂

### 2. 版本信息集成

#### 自动版本注入

```go
func (c *QuickConfig) ToFullOptions() *LogsOptions {
    // 从版本系统获取服务名
    versionInfo := version.Get()
    serviceName := versionInfo.ServiceName
    if serviceName == "" {
        serviceName = "apiserver" // 默认值
    }

    // ... 配置创建 ...

    // OTLP配置中的版本信息
    if enableOTLP {
        opts.OTLP = &OTLPConfig{
            ServiceName:    serviceName,
            ServiceVersion: versionInfo.GitVersion,
            // ... 其他配置 ...
        }
    }

    return opts
}
```

#### 初始字段设置

```go
// observabilityPreset中的版本字段
InitialFields: map[string]interface{}{
    "service.name":    serviceName,
    "service.version": "v2.0.0", // 从版本包获取
},
```

**集成策略**：

- **自动获取**: 无需手动配置版本信息
- **标准字段**: 使用OpenTelemetry标准字段名
- **动态更新**: 构建时自动注入最新版本

### 3. 错误处理设计

#### 渐进式降级

```go
// 目录创建失败时的处理
if err := ensureLogDir(filepath.Dir(logFile)); err == nil {
    opts.OutputPaths = append(opts.OutputPaths, logFile)
}
// 如果目录创建失败，继续使用stdout，不阻塞启动
```

#### 用户反馈

```go
// 配置状态反馈
if len(opts.OutputPaths) == 1 && opts.OutputPaths[0] == "stdout" {
    if opts.OTLP != nil && opts.OTLP.Enabled {
        fmt.Printf("   📝 Logs will only output to stdout (no file output when OTLP endpoint configured)\n")
    } else {
        fmt.Printf("   📝 Logs will only output to stdout (log-dir not configured)\n")
    }
} else {
    fmt.Printf("   📝 Current output paths: %v\n", opts.OutputPaths)
}
```

**设计原则**：

- **非阻塞**: 配置错误不影响服务启动
- **用户友好**: 清晰的状态反馈
- **自动恢复**: 智能降级到可用配置

## 性能优化

### 1. 条件性功能启用

```go
// 仅在需要时启用复杂功能
if c.Preset != PresetTesting && !enableOTLP && c.LogDir != "" {
    // 文件输出相关处理
}

if enableOTLP {
    // OTLP相关处理
}
```

### 2. 预设性能调优

```go
// Production预设的性能优化
DisableFunctionAtInfo: true,  // 生产环境隐藏函数名以提高性能
MaxSize:               100,
MaxAge:                7,
MaxBackups:            5,
Compress:              true,
```

### 3. OTLP批处理优化

```go
// 环境相关的批处理参数
if c.Preset == PresetProduction {
    opts.OTLP.BatchSize = 500              // 更大批次
    opts.OTLP.BatchTimeout = 2 * time.Second  // 更长等待
}
```

## 扩展点设计

### 1. 新预设添加

```go
// 添加新预设只需：
const PresetNewMode LogPreset = "newmode"

func (c *QuickConfig) newModePreset(serviceName string) *LogsOptions {
    return &LogsOptions{
        // 新预设的配置...
    }
}

// 在createPresetOptions中添加case
case PresetNewMode:
    return c.newModePreset(serviceName)
```

### 2. 新输出后端支持

```go
// 添加新的日志器实现只需：
const LoggerTypeNew LoggerType = "new"

func NewNewLogger(opts *LogsOptions) (Logger, error) {
    // 新实现...
}

// 在工厂中添加case
case LoggerTypeNew:
    return NewNewLogger(opts)
```

### 3. 配置扩展

```go
// QuickConfig扩展
type QuickConfig struct {
    // 现有字段...
    NewField string `json:"new-field"`
}

// 在ToFullOptions中处理
if c.NewField != "" {
    // 处理新字段...
}
```

## 测试策略

### 1. 配置转换测试

```go
func TestQuickConfigToFullOptions(t *testing.T) {
    tests := []struct {
        name     string
        config   *QuickConfig
        validate func(*LogsOptions) error
    }{
        {
            name: "OTLP enabled should disable file output",
            config: &QuickConfig{
                OTLPEndpoint: "127.0.0.1:4327",
                LogDir:       "logs",
            },
            validate: func(opts *LogsOptions) error {
                if hasFileOutput(opts.OutputPaths) {
                    return errors.New("file output should be disabled when OTLP enabled")
                }
                return nil
            },
        },
        // 更多测试用例...
    }
}
```

### 2. 预设一致性测试

```go
func TestPresetConsistency(t *testing.T) {
    presets := []LogPreset{
        PresetDevelopment,
        PresetProduction,
        PresetTesting,
        PresetObservability,
    }

    for _, preset := range presets {
        config := &QuickConfig{Preset: preset}
        opts := config.ToFullOptions()

        // 验证预设配置的一致性
        validatePreset(t, preset, opts)
    }
}
```

### 3. 集成测试

```go
func TestOTLPIntegration(t *testing.T) {
    // 启动测试OTLP接收器
    server := startTestOTLPServer(t)
    defer server.Stop()

    config := &QuickConfig{
        OTLPEndpoint: server.Address(),
    }

    logger, err := NewLoggerFromQuickConfig(config)
    require.NoError(t, err)

    logger.Info("test message")

    // 验证消息是否通过OTLP发送
    assert.True(t, server.ReceivedMessage("test message"))
}
```

## 配置迁移策略

### 1. 向后兼容性

```go
// 旧配置支持
func parseDetailedConfig() *LogsOptions {
    logOptions := DefaultOptions()

    // 继续支持旧的详细配置字段
    if viper.IsSet("log.type") {
        logOptions.Type = LoggerType(viper.GetString("log.type"))
    }
    // ... 其他旧字段 ...

    return logOptions
}
```

### 2. 配置检测优先级

```go
func reinitializeLoggerFromConfig() {
    var logOptions *LogsOptions

    // 优先使用简化配置（新方式）
    if viper.IsSet("log.preset") {
        quickConfig := parseQuickConfig()
        logOptions = quickConfig.ToFullOptions()
    } else {
        // 兼容旧的详细配置
        logOptions = parseDetailedConfig()
    }

    // 创建日志器...
}
```

### 3. 迁移辅助工具

```go
// 配置迁移建议
func SuggestMigration(oldConfig *LogsOptions) *QuickConfig {
    suggestion := &QuickConfig{}

    // 分析旧配置，建议相应的简化配置
    if oldConfig.Development {
        suggestion.Preset = PresetDevelopment
    } else if isObservabilityConfig(oldConfig) {
        suggestion.Preset = PresetObservability
    } else {
        suggestion.Preset = PresetProduction
    }

    return suggestion
}
```

## 未来演进方向

### 1. 短期目标

- **更多预设**: 添加特定场景的预设（如微服务、边缘计算等）
- **配置验证**: 增强配置验证和错误提示
- **性能监控**: 添加日志性能指标收集

### 2. 中期目标

- **动态配置**: 运行时配置热更新
- **插件系统**: 支持第三方输出插件
- **智能采样**: 基于负载的自适应采样

### 3. 长期目标

- **AI辅助**: 基于使用模式的智能配置推荐
- **多租户**: 支持多租户环境的日志隔离
- **边缘优化**: 针对边缘计算环境的专门优化

## 总结

Logger包的设计遵循"简单优先，需要时复杂"的哲学，通过双层配置系统、智能预设和自动检测机制，在保持强大功能的同时提供了极佳的用户体验。智能输出控制和OTLP集成使其特别适合现代可观测性架构，而扩展点设计确保了未来的可扩展性。
