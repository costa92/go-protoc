package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigValidator 配置验证器
type ConfigValidator struct {
	errors   []string
	warnings []string
}

// NewConfigValidator 创建配置验证器
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors:   make([]string, 0),
		warnings: make([]string, 0),
	}
}

// Validate 验证日志配置
func (v *ConfigValidator) Validate(opts *LogsOptions) error {
	v.errors = v.errors[:0]
	v.warnings = v.warnings[:0]

	if opts == nil {
		v.errors = append(v.errors, "configuration is nil")
		return v.buildError()
	}

	// 验证日志类型
	v.validateType(opts.Type)

	// 验证日志级别
	v.validateLevel(opts.Level)

	// 验证输出格式
	v.validateFormat(opts.Format, opts.Encoding)

	// 验证输出路径
	v.validateOutputPaths(opts.OutputPaths, false)
	v.validateOutputPaths(opts.ErrorOutputPaths, true)

	// 验证日志轮转配置
	v.validateRotation(opts)

	// 验证采样配置
	v.validateSampling(opts.Sampling)

	// 验证编码器配置
	v.validateEncoderConfig(opts.EncoderConfig)

	// 验证初始字段
	v.validateInitialFields(opts.InitialFields)

	return v.buildError()
}

func (v *ConfigValidator) validateType(logType LoggerType) {
	switch logType {
	case LoggerTypeZap, LoggerTypeSlog:
		// valid
	case "":
		v.warnings = append(v.warnings, "logger type is empty, will use default (zap)")
	default:
		v.errors = append(v.errors, fmt.Sprintf("invalid logger type: %s", logType))
	}
}

func (v *ConfigValidator) validateLevel(level string) {
	validLevels := []string{"debug", "info", "warn", "error", "fatal", ""}
	isValid := false
	for _, validLevel := range validLevels {
		if strings.ToLower(level) == validLevel {
			isValid = true
			break
		}
	}
	if !isValid {
		v.errors = append(v.errors, fmt.Sprintf("invalid log level: %s", level))
	}
	if level == "" {
		v.warnings = append(v.warnings, "log level is empty, will use default (info)")
	}
}

func (v *ConfigValidator) validateFormat(format, encoding string) {
	// 如果都为空，给出警告
	if format == "" && encoding == "" {
		v.warnings = append(v.warnings, "both format and encoding are empty, will use default (json)")
		return
	}

	// 验证格式
	validFormats := []string{"json", "text", "console", ""}
	formatValid := false
	for _, valid := range validFormats {
		if format == valid {
			formatValid = true
			break
		}
	}
	if !formatValid {
		v.errors = append(v.errors, fmt.Sprintf("invalid format: %s", format))
	}

	// 验证编码
	if encoding != "" {
		encodingValid := false
		for _, valid := range validFormats {
			if encoding == valid {
				encodingValid = true
				break
			}
		}
		if !encodingValid {
			v.errors = append(v.errors, fmt.Sprintf("invalid encoding: %s", encoding))
		}
	}
}

func (v *ConfigValidator) validateOutputPaths(paths []string, isError bool) {
	if len(paths) == 0 {
		msg := "no output paths specified, will use default (stdout)"
		if isError {
			msg = "no error output paths specified, will use default (stderr)"
		}
		v.warnings = append(v.warnings, msg)
		return
	}

	for _, path := range paths {
		if path == "stdout" || path == "stderr" {
			continue
		}

		// 检查父目录是否存在
		dir := filepath.Dir(path)
		if dir != "." && dir != "/" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				v.warnings = append(v.warnings, fmt.Sprintf("output path directory does not exist: %s", dir))
			}
		}

		// 检查文件是否可写
		if _, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
			// 如果不是权限问题，给出警告
			if !os.IsPermission(err) {
				v.warnings = append(v.warnings, fmt.Sprintf("output path may not be writable: %s", path))
			}
		}
	}
}

func (v *ConfigValidator) validateRotation(opts *LogsOptions) {
	// 只有当使用文件输出时才验证轮转配置
	hasFileOutput := false
	for _, path := range opts.OutputPaths {
		if path != "stdout" && path != "stderr" {
			hasFileOutput = true
			break
		}
	}

	if !hasFileOutput {
		return
	}

	if opts.MaxSize != 0 && opts.MaxSize < 1 {
		v.errors = append(v.errors, "MaxSize must be at least 1 MB")
	}
	if opts.MaxSize > 1024 {
		v.warnings = append(v.warnings, fmt.Sprintf("MaxSize is very large (%d MB), this may consume significant disk space", opts.MaxSize))
	}

	if opts.MaxAge < 0 {
		v.errors = append(v.errors, "MaxAge cannot be negative")
	}
	if opts.MaxAge > 365 {
		v.warnings = append(v.warnings, fmt.Sprintf("MaxAge is very long (%d days), logs may consume significant disk space", opts.MaxAge))
	}

	if opts.MaxBackups < 0 {
		v.errors = append(v.errors, "MaxBackups cannot be negative")
	}
	if opts.MaxBackups > 100 {
		v.warnings = append(v.warnings, fmt.Sprintf("MaxBackups is very large (%d), this may consume significant disk space", opts.MaxBackups))
	}
}

func (v *ConfigValidator) validateSampling(sampling *SamplingConfig) {
	if sampling == nil {
		return
	}

	if sampling.Initial < 0 {
		v.errors = append(v.errors, "Sampling.Initial cannot be negative")
	}
	if sampling.Thereafter < 0 {
		v.errors = append(v.errors, "Sampling.Thereafter cannot be negative")
	}

	if sampling.Initial == 0 && sampling.Thereafter > 0 {
		v.warnings = append(v.warnings, "Sampling.Initial is 0 but Thereafter is set, this will drop all logs")
	}
}

func (v *ConfigValidator) validateEncoderConfig(cfg *EncoderConfig) {
	if cfg == nil {
		return
	}

	// 验证时间编码器
	if cfg.TimeEncoder != "" {
		validEncoders := []string{"iso8601", "rfc3339", "epoch", "millis", "nanos"}
		valid := false
		for _, encoder := range validEncoders {
			if cfg.TimeEncoder == encoder {
				valid = true
				break
			}
		}
		if !valid {
			v.errors = append(v.errors, fmt.Sprintf("invalid TimeEncoder: %s", cfg.TimeEncoder))
		}
	}

	// 验证级别编码器
	if cfg.LevelEncoder != "" {
		validEncoders := []string{"lowercase", "uppercase", "capital"}
		valid := false
		for _, encoder := range validEncoders {
			if cfg.LevelEncoder == encoder {
				valid = true
				break
			}
		}
		if !valid {
			v.errors = append(v.errors, fmt.Sprintf("invalid LevelEncoder: %s", cfg.LevelEncoder))
		}
	}

	// 验证调用者编码器
	if cfg.CallerEncoder != "" {
		validEncoders := []string{"short", "full"}
		valid := false
		for _, encoder := range validEncoders {
			if cfg.CallerEncoder == encoder {
				valid = true
				break
			}
		}
		if !valid {
			v.errors = append(v.errors, fmt.Sprintf("invalid CallerEncoder: %s", cfg.CallerEncoder))
		}
	}
}

func (v *ConfigValidator) validateInitialFields(fields map[string]interface{}) {
	if fields == nil {
		return
	}

	// 检查保留字段
	reservedFields := []string{"level", "ts", "time", "msg", "message", "caller", "stacktrace", "type"}
	for key := range fields {
		for _, reserved := range reservedFields {
			if key == reserved {
				v.warnings = append(v.warnings, fmt.Sprintf("InitialFields contains reserved field '%s', this may be overwritten", key))
			}
		}
	}
}

func (v *ConfigValidator) buildError() error {
	if len(v.errors) == 0 {
		return nil
	}

	return fmt.Errorf("configuration validation failed:\n%s", strings.Join(v.errors, "\n"))
}

// GetWarnings 获取验证警告
func (v *ConfigValidator) GetWarnings() []string {
	return v.warnings
}

// GetErrors 获取验证错误
func (v *ConfigValidator) GetErrors() []string {
	return v.errors
}

// ValidateConfig 便捷函数，验证配置并返回错误和警告
func ValidateConfig(opts *LogsOptions) (warnings []string, err error) {
	validator := NewConfigValidator()
	err = validator.Validate(opts)
	warnings = validator.GetWarnings()
	return
}
