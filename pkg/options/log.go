package options

import (
	"github.com/spf13/pflag"
)

// LogOptions 定义了与日志相关的配置。
type LogOptions struct {
	Level            string   `json:"level"              mapstructure:"level"`
	Format           string   `json:"format"             mapstructure:"format"`
	EnableColor      bool     `json:"enable-color"       mapstructure:"enable-color"`
	EnableCaller     bool     `json:"enable-caller"      mapstructure:"enable-caller"`
	OutputPaths      []string `json:"output-paths"       mapstructure:"output-paths"`
	ErrorOutputPaths []string `json:"error-output-paths" mapstructure:"error-output-paths"`
}

// NewLogOptions 创建一个带有默认值的 LogOptions。
func NewLogOptions() *LogOptions {
	return &LogOptions{
		Level:            "info",
		Format:           "console",
		EnableColor:      true,
		EnableCaller:     true,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
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
