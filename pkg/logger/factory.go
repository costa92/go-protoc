package logger

import (
	"fmt"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[Logger]

func NewLogger(opts *LogsOptions) (Logger, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	switch opts.Type {
	case LoggerTypeZap:
		return NewZapLogger(opts)
	case LoggerTypeSlog:
		return NewSlogLogger(opts)
	default:
		return nil, fmt.Errorf("unsupported logger type: %s", opts.Type)
	}
}

func SetDefaultLogger(logger Logger) {
	defaultLogger.Store(&logger)
}

func GetDefaultLogger() Logger {
	if logger := defaultLogger.Load(); logger != nil {
		return *logger
	}
	
	logger, err := NewLogger(DefaultOptions())
	if err != nil {
		panic(fmt.Sprintf("failed to create default logger: %v", err))
	}
	
	SetDefaultLogger(logger)
	return logger
}

// getGlobalLogger 获取用于全局函数的 logger，额外跳过一层调用栈
func getGlobalLogger() Logger {
	logger := GetDefaultLogger()
	// 为全局函数增加额外的 caller skip
	return logger.WithCallerSkip(1)
}

func Debug(args ...interface{}) {
	getGlobalLogger().Debug(args...)
}

func Info(args ...interface{}) {
	getGlobalLogger().Info(args...)
}

func Warn(args ...interface{}) {
	getGlobalLogger().Warn(args...)
}

func Error(args ...interface{}) {
	getGlobalLogger().Error(args...)
}

func Fatal(args ...interface{}) {
	getGlobalLogger().Fatal(args...)
}

func Debugf(template string, args ...interface{}) {
	getGlobalLogger().Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	getGlobalLogger().Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	getGlobalLogger().Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	getGlobalLogger().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	getGlobalLogger().Fatalf(template, args...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	getGlobalLogger().Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	getGlobalLogger().Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	getGlobalLogger().Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	getGlobalLogger().Errorw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	getGlobalLogger().Fatalw(msg, keysAndValues...)
}