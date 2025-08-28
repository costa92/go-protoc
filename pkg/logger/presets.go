package logger

import (
	"os"
	"path/filepath"
	
	"github.com/costa92/go-protoc/v2/pkg/version"
)

// LogPreset 预定义日志配置模式
type LogPreset string

const (
	// PresetDevelopment 开发环境：控制台输出，调试级别，彩色，详细信息
	PresetDevelopment LogPreset = "development"
	// PresetProduction 生产环境：JSON格式，文件输出，info级别，结构化
	PresetProduction LogPreset = "production"
	// PresetTesting 测试环境：简化输出，error级别
	PresetTesting LogPreset = "testing"
	// PresetObservability 可观测性：专为日志收集系统优化
	PresetObservability LogPreset = "observability"
)

// QuickConfig 简化的日志配置
type QuickConfig struct {
	Preset      LogPreset `json:"preset" mapstructure:"preset"`           // 预设模式
	Type        string    `json:"type" mapstructure:"type"`               // 日志器类型: zap, slog 
	Level       string    `json:"level" mapstructure:"level"`             // 覆盖预设的日志级别
	LogDir      string    `json:"log-dir" mapstructure:"log-dir"`         // 日志目录
	
	// 可观测性配置
	EnableOTLP     bool   `json:"enable-otlp" mapstructure:"enable-otlp"`         // 启用OTLP
	OTLPEndpoint   string `json:"otlp-endpoint" mapstructure:"otlp-endpoint"`     // OTLP端点
}

// ToFullOptions 将简化配置转换为完整配置
func (c *QuickConfig) ToFullOptions() *LogsOptions {
	// 从版本系统获取服务名
	versionInfo := version.Get()
	serviceName := versionInfo.ServiceName
	if serviceName == "" {
		serviceName = "apiserver" // 默认值
	}
	
	// 设置默认值
	if c.LogDir == "" {
		c.LogDir = "logs"
	}
	if c.Type == "" {
		c.Type = "zap"
	}

	// 根据预设创建基础配置
	opts := c.createPresetOptions(serviceName)
	
	// 应用用户覆盖
	if c.Level != "" {
		opts.Level = c.Level
	}
	if c.Type != "" {
		opts.Type = LoggerType(c.Type)
	}
	
	// 设置日志文件路径
	if c.Preset != PresetTesting {
		logFile := filepath.Join(c.LogDir, serviceName, "app.log")
		// 确保目录存在
		if err := ensureLogDir(filepath.Dir(logFile)); err == nil {
			opts.OutputPaths = append(opts.OutputPaths, logFile)
		}
		// 如果目录创建失败，继续使用stdout，不阻塞启动
	}
	
	return opts
}

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

// developmentPreset 开发环境预设
func (c *QuickConfig) developmentPreset(serviceName string) *LogsOptions {
	return &LogsOptions{
		Type:                  LoggerTypeZap,
		Level:                 "debug",
		Format:                "console", // 开发环境用console格式，便于阅读
		DisableCaller:         false,
		DisableStacktrace:     false,
		EnableColor:           true,
		OutputPaths:           []string{"stdout"},
		ErrorOutputPaths:      []string{"stderr"},
		Development:           true,
		Encoding:              "console",
		CallerSkip:            1,
		DisableFunctionAtInfo: false, // 开发环境显示函数名
		EncoderConfig: &EncoderConfig{
			TimeKey:       "time",
			LevelKey:      "level",
			MessageKey:    "msg",
			CallerKey:     "caller",
			StacktraceKey: "stacktrace",
			FunctionKey:   "func",
			TimeEncoder:   "iso8601",
			LevelEncoder:  "capital", // 开发环境用大写级别
			CallerEncoder: "short",
		},
		InitialFields: map[string]interface{}{
			"service": serviceName,
		},
	}
}

// productionPreset 生产环境预设
func (c *QuickConfig) productionPreset(serviceName string) *LogsOptions {
	return &LogsOptions{
		Type:                  LoggerTypeZap,
		Level:                 "info",
		Format:                "json", // 生产环境用JSON格式
		DisableCaller:         false,
		DisableStacktrace:     false,
		EnableColor:           false,
		OutputPaths:           []string{"stdout"},
		ErrorOutputPaths:      []string{"stderr"},
		Development:           false,
		Encoding:              "json",
		CallerSkip:            1,
		DisableFunctionAtInfo: true, // 生产环境隐藏函数名以提高性能
		MaxSize:               100,
		MaxAge:                7,
		MaxBackups:            5,
		Compress:              true,
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
		InitialFields: map[string]interface{}{
			"service": serviceName,
		},
	}
}

// testingPreset 测试环境预设
func (c *QuickConfig) testingPreset(serviceName string) *LogsOptions {
	return &LogsOptions{
		Type:                  LoggerTypeZap,
		Level:                 "error", // 测试时只显示错误
		Format:                "json",
		DisableCaller:         true, // 测试环境不需要调用者信息
		DisableStacktrace:     true,
		EnableColor:           false,
		OutputPaths:           []string{"stdout"},
		ErrorOutputPaths:      []string{"stderr"},
		Development:           false,
		Encoding:              "json",
		CallerSkip:            1,
		DisableFunctionAtInfo: true,
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
		InitialFields: map[string]interface{}{
			"service": serviceName,
		},
	}
}

// observabilityPreset 可观测性预设 - 针对OTEL Collector优化
func (c *QuickConfig) observabilityPreset(serviceName string) *LogsOptions {
	return &LogsOptions{
		Type:                  LoggerTypeZap,
		Level:                 "info",
		Format:                "json", // 结构化JSON便于解析
		DisableCaller:         false,
		DisableStacktrace:     false,
		EnableColor:           false,
		OutputPaths:           []string{"stdout"},
		ErrorOutputPaths:      []string{"stderr"},
		Development:           false,
		Encoding:              "json",
		CallerSkip:            1,
		DisableFunctionAtInfo: false, // 可观测性需要完整信息
		MaxSize:               100,
		MaxAge:                30, // 长期保留用于分析
		MaxBackups:            10,
		Compress:              true,
		// 使用统一的字段名以兼容OTEL Collector配置
		EncoderConfig: &EncoderConfig{
			TimeKey:       "ts",             // 与现有OTEL配置兼容
			LevelKey:      "level",          // 与现有OTEL配置兼容
			MessageKey:    "msg",            // 与现有OTEL配置兼容（会被重映射为_msg）
			CallerKey:     "caller",         // 与现有OTEL配置兼容
			StacktraceKey: "stacktrace",
			FunctionKey:   "func",           // 与现有OTEL配置兼容
			TimeEncoder:   "iso8601",        // 保持一致的时间格式
			LevelEncoder:  "lowercase",      // 保持一致的级别格式
			CallerEncoder: "short",          // 平衡可读性和性能
		},
		InitialFields: map[string]interface{}{
			"service.name":    serviceName,
			"service.version": "v2.0.0", // 从版本包获取
		},
	}
}

// DefaultQuickConfig 返回默认的简化配置
func DefaultQuickConfig() *QuickConfig {
	return &QuickConfig{
		Preset:         PresetObservability, // 默认使用可观测性预设
		Type:           "zap",
		LogDir:         "logs",
		EnableOTLP:     true,
		OTLPEndpoint:   "127.0.0.1:4327",
	}
}

// NewLoggerFromQuickConfig 从简化配置创建日志器
func NewLoggerFromQuickConfig(config *QuickConfig) (Logger, error) {
	fullOptions := config.ToFullOptions()
	return NewLogger(fullOptions)
}

// ensureLogDir 确保日志目录存在
func ensureLogDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}