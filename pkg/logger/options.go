package logger

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

// Options 包含日志记录器的配置选项
type LogOptions struct {
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
	// EnableColor 确定是否在控制台输出中启用颜色。
	EnableColor bool `json:"enable-color"       mapstructure:"enable-color"`
}

// NewOptions 创建一个带有默认值的新 Options 对象。
func NewLogOptions() *LogOptions {
	return &LogOptions{
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		Level:            zapcore.InfoLevel.String(),
		Format:           "console",
		EnableCaller:     true,
		EnableColor:      true,
		// Name 是日志记录器的名称，通常用于标识日志来源。
		Name: "go-protoc",
	}
}

// Validate 校验日志选项。
func (o *LogOptions) Validate() []error {
	var errs []error
	// 你可以在此添加更复杂的校验逻辑
	return errs
}

// AddFlags 向指定的 FlagSet 添加日志相关的标志。
func (o *LogOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, "log.level", o.Level, "Log level (debug, info, warn, error, fatal)")
	fs.StringVar(&o.Format, "log.format", o.Format, "Log format (console, json)")
	fs.BoolVar(&o.EnableColor, "log.enable-color", o.EnableColor, "Enable color in log output")
	fs.BoolVar(&o.EnableCaller, "log.enable-caller", o.EnableCaller, "Enable caller in log output")
	fs.StringSliceVar(&o.OutputPaths, "log.output-paths", o.OutputPaths, "Log output paths")
	fs.StringSliceVar(&o.ErrorOutputPaths, "log.error-output-paths", o.ErrorOutputPaths, "Log error output paths")
}
