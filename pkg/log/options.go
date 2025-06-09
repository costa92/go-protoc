package log

import (
	"go.uber.org/zap/zapcore"
)

// Options 包含日志记录器的配置选项
type Options struct {
	// OutputPaths 是一个文件路径列表，用于写入日志。
	// 使用 "stdout" 或 "stderr" 来登录到控制台。
	OutputPaths []string `json:"output-paths" mapstructure:"output-paths"`
	// ErrorOutputPaths 是一个文件路径列表，用于写入错误日志。
	ErrorOutputPaths []string `json:"error-output-paths" mapstructure:"error-output-paths"`
	// Level 是将要记录的最低日志级别。
	// 可用级别: "debug", "info", "warn", "error", "dpanic", "panic", "fatal"
	Level string `json:"level" mapstructure:"level"`
	// Format 是日志格式。可以是 "json" 或 "console"。
	Format string `json:"format" mapstructure:"format"`
	// EnableCaller 确定是否在日志中包含调用者信息。
	EnableCaller bool `json:"enable-caller" mapstructure:"enable-caller"`
	// Name 是日志记录器的名称。
	Name string `json:"name" mapstructure:"name"`
}

// NewOptions 创建一个带有默认值的新 Options 对象。
func NewOptions() *Options {
	return &Options{
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Level:            zapcore.InfoLevel.String(),
		Format:           "console",
		EnableCaller:     true,
		Name:             "go-protoc",
	}
}
