package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ZapLogger struct {
	logger        *zap.Logger
	sugaredLogger *zap.SugaredLogger
	opts          *LogsOptions
	level         Level
	filter        Filter
	callerSkip    int
}

var _ Logger = (*ZapLogger)(nil)

func NewZapLogger(opts *LogsOptions) (*ZapLogger, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	config := zap.NewProductionConfig()
	if opts.Development {
		config = zap.NewDevelopmentConfig()
	}

	config.Level = zap.NewAtomicLevelAt(ParseLevel(opts.Level).ToZapLevel())
	config.DisableCaller = opts.DisableCaller
	config.DisableStacktrace = opts.DisableStacktrace
	config.Development = opts.Development

	if opts.Encoding != "" {
		config.Encoding = opts.Encoding
	}

	if opts.EncoderConfig != nil {
		config.EncoderConfig = buildZapEncoderConfig(opts.EncoderConfig)
	}

	if len(opts.OutputPaths) > 0 {
		config.OutputPaths = opts.OutputPaths
	}

	if len(opts.ErrorOutputPaths) > 0 {
		config.ErrorOutputPaths = opts.ErrorOutputPaths
	}

	if opts.InitialFields != nil {
		config.InitialFields = opts.InitialFields
	}

	if opts.Sampling != nil {
		config.Sampling = &zap.SamplingConfig{
			Initial:    opts.Sampling.Initial,
			Thereafter: opts.Sampling.Thereafter,
		}
	}

	var cores []zapcore.Core

	for _, outputPath := range opts.OutputPaths {
		var writer zapcore.WriteSyncer
		
		if outputPath == "stdout" {
			writer = zapcore.AddSync(os.Stdout)
		} else if outputPath == "stderr" {
			writer = zapcore.AddSync(os.Stderr)
		} else {
			lumberJackLogger := &lumberjack.Logger{
				Filename:   outputPath,
				MaxSize:    opts.MaxSize,
				MaxAge:     opts.MaxAge,
				MaxBackups: opts.MaxBackups,
				Compress:   opts.Compress,
			}
			writer = zapcore.AddSync(lumberJackLogger)
		}

		var encoder zapcore.Encoder
		if opts.Encoding == "console" {
			encoder = zapcore.NewConsoleEncoder(config.EncoderConfig)
		} else {
			encoder = zapcore.NewJSONEncoder(config.EncoderConfig)
		}

		core := zapcore.NewCore(encoder, writer, config.Level)
		cores = append(cores, core)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(opts.CallerSkip))
	
	if opts.Development {
		logger = logger.WithOptions(zap.Development())
	}

	return &ZapLogger{
		logger:        logger,
		sugaredLogger: logger.Sugar(),
		opts:          opts,
		level:         ParseLevel(opts.Level),
	}, nil
}

func buildZapEncoderConfig(cfg *EncoderConfig) zapcore.EncoderConfig {
	encoderConfig := zap.NewProductionEncoderConfig()
	
	if cfg.TimeKey != "" {
		encoderConfig.TimeKey = cfg.TimeKey
	}
	if cfg.LevelKey != "" {
		encoderConfig.LevelKey = cfg.LevelKey
	}
	if cfg.MessageKey != "" {
		encoderConfig.MessageKey = cfg.MessageKey
	}
	if cfg.CallerKey != "" {
		encoderConfig.CallerKey = cfg.CallerKey
	}
	if cfg.StacktraceKey != "" {
		encoderConfig.StacktraceKey = cfg.StacktraceKey
	}
	if cfg.FunctionKey != "" {
		encoderConfig.FunctionKey = cfg.FunctionKey
	}

	switch cfg.TimeEncoder {
	case "iso8601":
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	case "rfc3339":
		encoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	case "epoch":
		encoderConfig.EncodeTime = zapcore.EpochTimeEncoder
	case "millis":
		encoderConfig.EncodeTime = zapcore.EpochMillisTimeEncoder
	case "nanos":
		encoderConfig.EncodeTime = zapcore.EpochNanosTimeEncoder
	default:
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	switch cfg.LevelEncoder {
	case "lowercase":
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	case "capital":
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	case "color":
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}

	switch cfg.CallerEncoder {
	case "short":
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	case "full":
		encoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	default:
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	}

	if cfg.DurationUnit > 0 {
		encoderConfig.EncodeDuration = func(d time.Duration, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendFloat64(float64(d) / float64(cfg.DurationUnit))
		}
	}

	return encoderConfig
}

func (z *ZapLogger) getCaller(skip int) (file string, line int, function string) {
	pc, file, line, ok := runtime.Caller(skip + z.opts.CallerSkip + z.callerSkip + 1)
	if !ok {
		return "unknown", 0, "unknown"
	}

	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	fn := runtime.FuncForPC(pc)
	function = "unknown"
	if fn != nil {
		function = fn.Name()
		if idx := strings.LastIndex(function, "."); idx >= 0 {
			function = function[idx+1:]
		}
	}

	return file, line, function
}

func (z *ZapLogger) shouldLog(level Level, ctx context.Context, msg string) bool {
	if level < z.level {
		return false
	}

	if z.filter != nil && !z.filter.IsEmpty() {
		file, line, function := z.getCaller(3)
		meta := &Metadata{
			File: file,
			Line: line,
			Func: function,
			Msg:  msg,
		}
		return z.filter.Enabled(ctx, meta)
	}

	return true
}

func (z *ZapLogger) Debug(args ...interface{}) {
	if z.shouldLog(DebugLevel, context.Background(), fmt.Sprint(args...)) {
		z.sugaredLogger.Debug(args...)
	}
}

func (z *ZapLogger) Info(args ...interface{}) {
	if z.shouldLog(InfoLevel, context.Background(), fmt.Sprint(args...)) {
		z.sugaredLogger.Info(args...)
	}
}

func (z *ZapLogger) Warn(args ...interface{}) {
	if z.shouldLog(WarnLevel, context.Background(), fmt.Sprint(args...)) {
		z.sugaredLogger.Warn(args...)
	}
}

func (z *ZapLogger) Error(args ...interface{}) {
	if z.shouldLog(ErrorLevel, context.Background(), fmt.Sprint(args...)) {
		z.sugaredLogger.Error(args...)
	}
}

func (z *ZapLogger) Fatal(args ...interface{}) {
	if z.shouldLog(FatalLevel, context.Background(), fmt.Sprint(args...)) {
		z.sugaredLogger.Fatal(args...)
	}
}

func (z *ZapLogger) Debugf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if z.shouldLog(DebugLevel, context.Background(), msg) {
		z.sugaredLogger.Debugf(template, args...)
	}
}

func (z *ZapLogger) Infof(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if z.shouldLog(InfoLevel, context.Background(), msg) {
		z.sugaredLogger.Infof(template, args...)
	}
}

func (z *ZapLogger) Warnf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if z.shouldLog(WarnLevel, context.Background(), msg) {
		z.sugaredLogger.Warnf(template, args...)
	}
}

func (z *ZapLogger) Errorf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if z.shouldLog(ErrorLevel, context.Background(), msg) {
		z.sugaredLogger.Errorf(template, args...)
	}
}

func (z *ZapLogger) Fatalf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if z.shouldLog(FatalLevel, context.Background(), msg) {
		z.sugaredLogger.Fatalf(template, args...)
	}
}

func (z *ZapLogger) Debugw(msg string, keysAndValues ...interface{}) {
	if z.shouldLog(DebugLevel, context.Background(), msg) {
		z.sugaredLogger.Debugw(msg, keysAndValues...)
	}
}

func (z *ZapLogger) Infow(msg string, keysAndValues ...interface{}) {
	if z.shouldLog(InfoLevel, context.Background(), msg) {
		z.sugaredLogger.Infow(msg, keysAndValues...)
	}
}

func (z *ZapLogger) Warnw(msg string, keysAndValues ...interface{}) {
	if z.shouldLog(WarnLevel, context.Background(), msg) {
		z.sugaredLogger.Warnw(msg, keysAndValues...)
	}
}

func (z *ZapLogger) Errorw(msg string, keysAndValues ...interface{}) {
	if z.shouldLog(ErrorLevel, context.Background(), msg) {
		z.sugaredLogger.Errorw(msg, keysAndValues...)
	}
}

func (z *ZapLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	if z.shouldLog(FatalLevel, context.Background(), msg) {
		z.sugaredLogger.Fatalw(msg, keysAndValues...)
	}
}

func (z *ZapLogger) With(keyValues ...interface{}) Logger {
	newLogger := &ZapLogger{
		logger:        z.logger,
		sugaredLogger: z.sugaredLogger.With(keyValues...),
		opts:          z.opts,
		level:         z.level,
		filter:        z.filter,
		callerSkip:    z.callerSkip,
	}
	return newLogger
}

func (z *ZapLogger) WithCtx(ctx context.Context, keyValues ...interface{}) Logger {
	return z.With(keyValues...)
}

func (z *ZapLogger) WithCallerSkip(skip int) Logger {
	// 创建一个新的 zap logger 实例，调整 CallerSkip
	newZapLogger := z.logger.WithOptions(zap.AddCallerSkip(skip))
	newLogger := &ZapLogger{
		logger:        newZapLogger,
		sugaredLogger: newZapLogger.Sugar(),
		opts:          z.opts,
		level:         z.level,
		filter:        z.filter,
		callerSkip:    z.callerSkip + skip,
	}
	return newLogger
}

func (z *ZapLogger) SetLevel(level Level) {
	z.level = level
}