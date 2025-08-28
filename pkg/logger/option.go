package logger

import (
	"time"

	"github.com/spf13/pflag"
)

type LoggerType string

const (
	LoggerTypeZap  LoggerType = "zap"
	LoggerTypeSlog LoggerType = "slog"
)

// LogsOptions 定义日志系统的配置选项
type LogsOptions struct {
	// Type 指定日志记录器类型 (zap 或 slog)
	Type LoggerType `json:"type" mapstructure:"type"`
	// Dynamic 启用动态配置功能，支持运行时调整日志配置
	Dynamic bool `json:"dynamic" mapstructure:"dynamic"`
	// Level 设置日志级别 (debug, info, warn, error, fatal)
	Level string `json:"level" mapstructure:"level"`
	// Format 指定日志输出格式 (json, console)
	Format string `json:"format" mapstructure:"format"`
	// DisableCaller 禁用调用者信息输出
	DisableCaller bool `json:"disable-caller" mapstructure:"disable-caller"`
	// DisableStacktrace 禁用堆栈跟踪输出
	DisableStacktrace bool `json:"disable-stacktrace" mapstructure:"disable-stacktrace"`
	// EnableColor 启用控制台彩色输出
	EnableColor bool `json:"enable-color" mapstructure:"enable-color"`
	// OutputPaths 指定日志输出路径列表
	OutputPaths []string `json:"output-paths" mapstructure:"output-paths"`
	// ErrorOutputPaths 指定错误日志输出路径列表
	ErrorOutputPaths []string `json:"error-output-paths" mapstructure:"error-output-paths"`
	// Development 启用开发模式 (更详细的日志输出)
	Development bool `json:"development" mapstructure:"development"`
	// Encoding 指定编码器类型 (json, console)
	Encoding string `json:"encoding" mapstructure:"encoding"`
	// EncoderConfig 编码器配置选项
	EncoderConfig *EncoderConfig `json:"encoder-config" mapstructure:"encoder-config"`
	// InitialFields 设置初始字段键值对
	InitialFields map[string]interface{} `json:"initial-fields" mapstructure:"initial-fields"`
	// MaxSize 日志文件最大大小 (MB)
	MaxSize int `json:"max-size" mapstructure:"max-size"`
	// MaxAge 日志文件最大保留天数
	MaxAge int `json:"max-age" mapstructure:"max-age"`
	// MaxBackups 最大备份文件数量
	MaxBackups int `json:"max-backups" mapstructure:"max-backups"`
	// Compress 是否压缩旧日志文件
	Compress bool `json:"compress" mapstructure:"compress"`
	// Sampling 日志采样配置
	Sampling *SamplingConfig `json:"sampling" mapstructure:"sampling"`
	// CallerSkip 调用者跳过的堆栈帧数
	CallerSkip int `json:"caller-skip" mapstructure:"caller-skip"`
	// DisableFunctionAtInfo 在info级别日志中禁用函数名显示
	DisableFunctionAtInfo bool `json:"disable-function-at-info" mapstructure:"disable-function-at-info"`
	// OTLP OTLP日志导出配置
	OTLP *OTLPConfig `json:"otlp" mapstructure:"otlp"`
}

// EncoderConfig 定义日志编码器的配置选项
type EncoderConfig struct {
	// TimeKey 时间字段的键名
	TimeKey string `json:"time-key" mapstructure:"time-key"`
	// LevelKey 日志级别字段的键名
	LevelKey string `json:"level-key" mapstructure:"level-key"`
	// MessageKey 消息字段的键名
	MessageKey string `json:"message-key" mapstructure:"message-key"`
	// CallerKey 调用者信息字段的键名
	CallerKey string `json:"caller-key" mapstructure:"caller-key"`
	// StacktraceKey 堆栈跟踪字段的键名
	StacktraceKey string `json:"stacktrace-key" mapstructure:"stacktrace-key"`
	// FunctionKey 函数名字段的键名
	FunctionKey string `json:"function-key" mapstructure:"function-key"`
	// TimeEncoder 时间编码格式 (iso8601, rfc3339, epoch, millis, nanos)
	TimeEncoder string `json:"time-encoder" mapstructure:"time-encoder"`
	// LevelEncoder 级别编码格式 (lowercase, uppercase, capital)
	LevelEncoder string `json:"level-encoder" mapstructure:"level-encoder"`
	// CallerEncoder 调用者编码格式 (short, full)
	CallerEncoder string `json:"caller-encoder" mapstructure:"caller-encoder"`
	// DurationUnit 持续时间单位
	DurationUnit time.Duration `json:"duration-unit" mapstructure:"duration-unit"`
}

// SamplingConfig 定义日志采样的配置选项
type SamplingConfig struct {
	// Initial 初始采样数量 (每秒允许的日志条数)
	Initial int `json:"initial" mapstructure:"initial"`
	// Thereafter 后续采样频率 (达到初始数量后，每N条日志记录1条)
	Thereafter int `json:"thereafter" mapstructure:"thereafter"`
}

func (l *LogsOptions) clone() *LogsOptions {
	if l == nil {
		return &LogsOptions{}
	}

	clone := &LogsOptions{
		Type:                  l.Type,
		Dynamic:               l.Dynamic,
		Level:                 l.Level,
		Format:                l.Format,
		DisableCaller:         l.DisableCaller,
		DisableStacktrace:     l.DisableStacktrace,
		EnableColor:           l.EnableColor,
		Development:           l.Development,
		Encoding:              l.Encoding,
		MaxSize:               l.MaxSize,
		MaxAge:                l.MaxAge,
		MaxBackups:            l.MaxBackups,
		Compress:              l.Compress,
		CallerSkip:            l.CallerSkip,
		DisableFunctionAtInfo: l.DisableFunctionAtInfo,
	}

	if len(l.OutputPaths) > 0 {
		clone.OutputPaths = make([]string, len(l.OutputPaths))
		copy(clone.OutputPaths, l.OutputPaths)
	}

	if len(l.ErrorOutputPaths) > 0 {
		clone.ErrorOutputPaths = make([]string, len(l.ErrorOutputPaths))
		copy(clone.ErrorOutputPaths, l.ErrorOutputPaths)
	}

	if l.EncoderConfig != nil {
		clone.EncoderConfig = &EncoderConfig{
			TimeKey:       l.EncoderConfig.TimeKey,
			LevelKey:      l.EncoderConfig.LevelKey,
			MessageKey:    l.EncoderConfig.MessageKey,
			CallerKey:     l.EncoderConfig.CallerKey,
			StacktraceKey: l.EncoderConfig.StacktraceKey,
			FunctionKey:   l.EncoderConfig.FunctionKey,
			TimeEncoder:   l.EncoderConfig.TimeEncoder,
			LevelEncoder:  l.EncoderConfig.LevelEncoder,
			CallerEncoder: l.EncoderConfig.CallerEncoder,
			DurationUnit:  l.EncoderConfig.DurationUnit,
		}
	}

	if l.InitialFields != nil {
		clone.InitialFields = make(map[string]interface{})
		for k, v := range l.InitialFields {
			clone.InitialFields[k] = v
		}
	}

	if l.Sampling != nil {
		clone.Sampling = &SamplingConfig{
			Initial:    l.Sampling.Initial,
			Thereafter: l.Sampling.Thereafter,
		}
	}

	return clone
}

// AddFlags adds command line flags for the LogsOptions configuration.
func (l *LogsOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar((*string)(&l.Type), "log.type", string(l.Type), "Logger type: zap or slog")
	fs.BoolVar(&l.Dynamic, "log.dynamic", l.Dynamic, "Enable dynamic logger configuration")
	fs.StringVar(&l.Level, "log.level", l.Level, "Minimum log output level")
	fs.StringVar(&l.Format, "log.format", l.Format, "Log output format: json, text, console")
	fs.BoolVar(&l.DisableCaller, "log.disable-caller", l.DisableCaller, "Disable output of caller information in the log")
	fs.BoolVar(&l.DisableStacktrace, "log.disable-stacktrace", l.DisableStacktrace, "Disable the log to record a stack trace for all messages at or above panic level")
	fs.BoolVar(&l.EnableColor, "log.enable-color", l.EnableColor, "Enable output ansi colors in plain format logs")
	fs.StringSliceVar(&l.OutputPaths, "log.output-paths", l.OutputPaths, "Output paths of log")
	fs.StringSliceVar(&l.ErrorOutputPaths, "log.error-output-paths", l.ErrorOutputPaths, "Error output paths of log")
	fs.BoolVar(&l.Development, "log.development", l.Development, "Enable development mode")
	fs.StringVar(&l.Encoding, "log.encoding", l.Encoding, "Log encoding format")
	fs.IntVar(&l.MaxSize, "log.max-size", l.MaxSize, "Maximum size of log file in megabytes")
	fs.IntVar(&l.MaxAge, "log.max-age", l.MaxAge, "Maximum number of days to retain log files")
	fs.IntVar(&l.MaxBackups, "log.max-backups", l.MaxBackups, "Maximum number of old log files to retain")
	fs.BoolVar(&l.Compress, "log.compress", l.Compress, "Compress old log files")
	fs.IntVar(&l.CallerSkip, "log.caller-skip", l.CallerSkip, "Number of caller frames to skip")
	fs.BoolVar(&l.DisableFunctionAtInfo, "log.disable-function-at-info", l.DisableFunctionAtInfo, "Disable function name display in info level logs")

	// EncoderConfig flags
	if l.EncoderConfig != nil {
		fs.StringVar(&l.EncoderConfig.TimeKey, "log.time-key", l.EncoderConfig.TimeKey, "Time field key in log output")
		fs.StringVar(&l.EncoderConfig.LevelKey, "log.level-key", l.EncoderConfig.LevelKey, "Level field key in log output")
		fs.StringVar(&l.EncoderConfig.MessageKey, "log.message-key", l.EncoderConfig.MessageKey, "Message field key in log output")
		fs.StringVar(&l.EncoderConfig.CallerKey, "log.caller-key", l.EncoderConfig.CallerKey, "Caller field key in log output")
		fs.StringVar(&l.EncoderConfig.StacktraceKey, "log.stacktrace-key", l.EncoderConfig.StacktraceKey, "Stacktrace field key in log output")
		fs.StringVar(&l.EncoderConfig.FunctionKey, "log.function-key", l.EncoderConfig.FunctionKey, "Function field key in log output")
		fs.StringVar(&l.EncoderConfig.TimeEncoder, "log.time-encoder", l.EncoderConfig.TimeEncoder, "Time encoder format")
		fs.StringVar(&l.EncoderConfig.LevelEncoder, "log.level-encoder", l.EncoderConfig.LevelEncoder, "Level encoder format")
		fs.StringVar(&l.EncoderConfig.CallerEncoder, "log.caller-encoder", l.EncoderConfig.CallerEncoder, "Caller encoder format")
	}

	// Sampling flags
	if l.Sampling != nil {
		fs.IntVar(&l.Sampling.Initial, "log.sampling-initial", l.Sampling.Initial, "Initial sampling rate")
		fs.IntVar(&l.Sampling.Thereafter, "log.sampling-thereafter", l.Sampling.Thereafter, "Thereafter sampling rate")
	}
}

// Validate verifies the LogsOptions configuration.
func (l *LogsOptions) Validate() []error {
	warnings, err := ValidateConfig(l)
	errs := []error{}
	if err != nil {
		errs = append(errs, err)
	}
	// 将警告作为错误返回，如果需要的话
	_ = warnings // 目前忽略警告
	return errs
}

func DefaultOptions() *LogsOptions {
	return &LogsOptions{
		Type:                  LoggerTypeZap,
		Dynamic:               false, // 默认禁用动态配置以保持性能
		Level:                 "info",
		Format:                "json",
		DisableCaller:         false,
		DisableStacktrace:     false,
		EnableColor:           true,
		OutputPaths:           []string{"stdout"},
		ErrorOutputPaths:      []string{"stderr"},
		Development:           false,
		Encoding:              "json",
		CallerSkip:            1,
		DisableFunctionAtInfo: true, // 默认在info级别禁用func字段
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
			DurationUnit:  time.Millisecond,
		},
	}
}
