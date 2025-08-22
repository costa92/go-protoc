package logger

import (
	"log/slog"
	"sync/atomic"
)

var defaultLoggerImpl atomic.Pointer[LoggerImpl]

var (
	SlogDebugLevel = slog.LevelDebug
	SlogInfoLevel  = slog.LevelInfo
	SlogWarnLevel  = slog.LevelWarn
	SlogErrorLevel = slog.LevelError
)

func SetDefaultLoggerImpl(logger *LoggerImpl) {
	defaultLoggerImpl.Store(logger)
}

func GetDefaultLoggerImpl() *LoggerImpl {
	if logger := defaultLoggerImpl.Load(); logger != nil {
		return logger
	}
	
	logger, err := NewLoggerImpl(DefaultOptions())
	if err != nil {
		panic("failed to create default logger implementation: " + err.Error())
	}
	
	SetDefaultLoggerImpl(logger)
	return logger
}
