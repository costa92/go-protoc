package logger

import (
	krtlog "github.com/go-kratos/kratos/v2/log"
)

// NewKratosLogger creates a Kratos-compatible logger using the provided logger.
func NewKratosLogger(l Logger, id, name, version string) krtlog.Logger {
	kratosLogger := &kratosLoggerAdapter{
		logger: l.With(
			"service.id", id,
			"service.name", name,
			"service.version", version,
		),
	}
	
	return krtlog.With(kratosLogger,
		"ts", krtlog.DefaultTimestamp,
		"caller", krtlog.DefaultCaller,
	)
}

// kratosLoggerAdapter adapts the generic logger.Logger to implement Kratos's log.Logger interface.
type kratosLoggerAdapter struct {
	logger Logger
}

// Log implements the Kratos Logger interface.
func (l *kratosLoggerAdapter) Log(level krtlog.Level, keyvals ...interface{}) error {
	switch level {
	case krtlog.LevelDebug:
		l.logger.Debugw("", keyvals...)
	case krtlog.LevelInfo:
		l.logger.Infow("", keyvals...)
	case krtlog.LevelWarn:
		l.logger.Warnw("", keyvals...)
	case krtlog.LevelError:
		l.logger.Errorw("", keyvals...)
	case krtlog.LevelFatal:
		l.logger.Fatalw("", keyvals...)
	default:
		l.logger.Infow("", keyvals...)
	}
	return nil
}