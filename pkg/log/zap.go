package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger 是一个使用 zap 来记录日志的记录器。
type zapLogger struct {
	z     *zap.Logger
	level zapcore.Level
}

// NewZapLogger 根据给定的选项创建一个新的 zapLogger。
func NewZapLogger(opts *Options) (Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(opts.Level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	if opts.Format == "console" {
		encoderConfig = zap.NewDevelopmentEncoderConfig()
	}
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	cfg := &zap.Config{
		Level:             zap.NewAtomicLevelAt(zapLevel),
		Development:       opts.Format == "console",
		DisableCaller:     !opts.EnableCaller,
		DisableStacktrace: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         opts.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      opts.OutputPaths,
		ErrorOutputPaths: opts.ErrorOutputPaths,
	}

	z, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	logger := &zapLogger{
		z:     z.Named(opts.Name),
		level: zapLevel,
	}

	return logger, nil
}

func (l *zapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Debugw(msg, keysAndValues...)
}

func (l *zapLogger) Infow(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Infow(msg, keysAndValues...)
}

func (l *zapLogger) Debugf(format string, args ...interface{}) {
	l.z.Sugar().Debugf(format, args...)
}

func (l *zapLogger) Infof(format string, args ...interface{}) {
	l.z.Sugar().Infof(format, args...)
}

func (l *zapLogger) Warnf(format string, args ...interface{}) {
	l.z.Sugar().Warnf(format, args...)
}

func (l *zapLogger) Errorf(format string, args ...interface{}) {
	l.z.Sugar().Errorf(format, args...)
}

func (l *zapLogger) Panicf(format string, args ...interface{}) {
	l.z.Sugar().Panicf(format, args...)
}

func (l *zapLogger) Fatalf(format string, args ...interface{}) {
	l.z.Sugar().Fatalf(format, args...)
}

func (l *zapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Errorw(msg, keysAndValues...)
}

func (l *zapLogger) Panicw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Panicw(msg, keysAndValues...)
}

func (l *zapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Fatalw(msg, keysAndValues...)
}

func (l *zapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	l.z.Sugar().Warnw(msg, keysAndValues...)
}

func (l *zapLogger) WithValues(keysAndValues ...interface{}) Logger {
	newLogger := l.z.With(handleFields(keysAndValues)...)
	return &zapLogger{z: newLogger}
}

func (l *zapLogger) Sync() {
	_ = l.z.Sync()
}

// handleFields 将一个 interface{} 切片转换为一个 zap.Field 切片。
func handleFields(args []interface{}) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	fields := make([]zap.Field, 0, len(args)/2)
	for i := 0; i < len(args); {
		// 确保我们不会因为奇数个参数而 panic
		if i == len(args)-1 {
			fields = append(fields, zap.Any("ignored", args[i]))
			break
		}

		key, val := args[i], args[i+1]
		keyStr, isString := key.(string)
		if !isString {
			// 确保 key 是字符串
			fields = append(fields, zap.Any("ignored", key))
		} else {
			fields = append(fields, zap.Any(keyStr, val))
		}
		i += 2
	}
	return fields
}

// SugaredLogger 是 zap.SugaredLogger 的一个包装器，它实现了 Logger 接口。
type SugaredLogger struct {
	*zap.SugaredLogger
}

var _ Logger = &SugaredLogger{}

// WithValues 返回一个新的 SugaredLogger，其中包含额外的上下文。
func (l *SugaredLogger) WithValues(keysAndValues ...interface{}) Logger {
	return &SugaredLogger{l.SugaredLogger.With(keysAndValues...)}
}

// Sync 同步底层 zap 日志记录器。
func (l *SugaredLogger) Sync() {
	_ = l.SugaredLogger.Sync()
}
