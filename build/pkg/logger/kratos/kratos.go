package kratos

import (
	krtlog "github.com/go-kratos/kratos/v2/log"

	"github.com/costa92/go-protoc/v2/pkg/logger"
)

// kratosLoggerAdapter adapts the generic logger.Logger to implement Kratos's log.Logger interface.
type kratosLoggerAdapter struct {
	logger logger.Logger
}

// NewLogger creates a Kratos-compatible logger with service metadata.
func NewLogger(l logger.Logger, id, name, version string) krtlog.Logger {
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