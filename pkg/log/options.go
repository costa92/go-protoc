package log

import (
	"github.com/spf13/pflag"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	// DisableCaller specifies whether to include caller information in the log.
	DisableCaller bool `json:"disable-caller,omitempty" mapstructure:"disable-caller"`
	// DisableStacktrace specifies whether to record a stack trace for all messages at or above panic level.
	DisableStacktrace bool `json:"disable-stacktrace,omitempty" mapstructure:"disable-stacktrace"`
	// EnableColor specifies whether to output colored logs.
	EnableColor bool `json:"enable-color"       mapstructure:"enable-color"`
	// Level specifies the minimum log level. Valid values are: debug, info, warn, error, dpanic, panic, and fatal.
	Level string `json:"level,omitempty" mapstructure:"level"`
	// Format specifies the log output format. Valid values are: console and json.
	Format string `json:"format,omitempty" mapstructure:"format"`
	// OutputPaths specifies the output paths for the logs.
	OutputPaths []string `json:"output-paths,omitempty" mapstructure:"output-paths"`
	// VictoriaLogs 配置
	VictoriaLogs *VictoriaLogsConfig `json:"victoria-logs,omitempty" mapstructure:"victoria-logs"`
}

// VictoriaLogsConfig VictoriaLogs 集成配置
type VictoriaLogsConfig struct {
	// Enabled 是否启用VictoriaLogs集成
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
	// Endpoint VictoriaLogs API端点
	Endpoint string `json:"endpoint,omitempty" mapstructure:"endpoint"`
	// Service 服务名称
	Service string `json:"service,omitempty" mapstructure:"service"`
	// Version 服务版本
	Version string `json:"version,omitempty" mapstructure:"version"`
	// Environment 环境标识
	Environment string `json:"environment,omitempty" mapstructure:"environment"`
	// BufferSize 缓冲区大小
	BufferSize int `json:"buffer-size,omitempty" mapstructure:"buffer-size"`
	// BatchSize 批量发送大小
	BatchSize int `json:"batch-size,omitempty" mapstructure:"batch-size"`
	// FlushInterval 强制刷新间隔（秒）
	FlushInterval int `json:"flush-interval,omitempty" mapstructure:"flush-interval"`
	// Timeout HTTP请求超时（秒）
	Timeout int `json:"timeout,omitempty" mapstructure:"timeout"`
}

// NewOptions creates a new Options object with default values.
func NewOptions() *Options {
	return &Options{
		Level:       zapcore.InfoLevel.String(),
		Format:      "console",
		OutputPaths: []string{"stdout"},
		VictoriaLogs: &VictoriaLogsConfig{
			Enabled:       false, // 恢复原始默认值，通过配置文件启用
			Endpoint:      "http://127.0.0.1:9428",
			Service:       "apiserver",
			Version:       "v2.0.0",
			Environment:   "development",
			BufferSize:    1000,
			BatchSize:     100,
			FlushInterval: 5,
			Timeout:       10,
		},
	}
}

// Validate verifies flags passed to LogsOptions.
func (o *Options) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds command line flags for the configuration.
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Level, "log.level", o.Level, "Minimum log output `LEVEL`.")
	fs.BoolVar(&o.DisableCaller, "log.disable-caller", o.DisableCaller, "Disable output of caller information in the log.")
	fs.BoolVar(&o.DisableStacktrace, "log.disable-stacktrace", o.DisableStacktrace, ""+
		"Disable the log to record a stack trace for all messages at or above panic level.")
	fs.BoolVar(&o.EnableColor, "log.enable-color", o.EnableColor, "Enable output ansi colors in plain format logs.")
	fs.StringVar(&o.Format, "log.format", o.Format, "Log output `FORMAT`, support plain or json format.")
	fs.StringSliceVar(&o.OutputPaths, "log.output-paths", o.OutputPaths, "Output paths of log.")

	// VictoriaLogs 相关配置
	if o.VictoriaLogs != nil {
		fs.BoolVar(&o.VictoriaLogs.Enabled, "log.victoria-logs.enabled", o.VictoriaLogs.Enabled, "Enable VictoriaLogs integration.")
		fs.StringVar(&o.VictoriaLogs.Endpoint, "log.victoria-logs.endpoint", o.VictoriaLogs.Endpoint, "VictoriaLogs API endpoint.")
		fs.StringVar(&o.VictoriaLogs.Service, "log.victoria-logs.service", o.VictoriaLogs.Service, "Service name for VictoriaLogs.")
		fs.StringVar(&o.VictoriaLogs.Version, "log.victoria-logs.version", o.VictoriaLogs.Version, "Service version for VictoriaLogs.")
		fs.StringVar(&o.VictoriaLogs.Environment, "log.victoria-logs.environment", o.VictoriaLogs.Environment, "Environment for VictoriaLogs.")
		fs.IntVar(&o.VictoriaLogs.BufferSize, "log.victoria-logs.buffer-size", o.VictoriaLogs.BufferSize, "VictoriaLogs buffer size.")
		fs.IntVar(&o.VictoriaLogs.BatchSize, "log.victoria-logs.batch-size", o.VictoriaLogs.BatchSize, "VictoriaLogs batch size.")
		fs.IntVar(&o.VictoriaLogs.FlushInterval, "log.victoria-logs.flush-interval", o.VictoriaLogs.FlushInterval, "VictoriaLogs flush interval in seconds.")
		fs.IntVar(&o.VictoriaLogs.Timeout, "log.victoria-logs.timeout", o.VictoriaLogs.Timeout, "VictoriaLogs HTTP timeout in seconds.")
	}
}
