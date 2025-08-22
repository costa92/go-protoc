package logger

import (
	"context"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

// NewGormLogger creates a GORM-compatible logger using the provided logger options.
// This is a convenience function for integration with database configurations.
func NewGormLogger(opts *LogsOptions, logLevel gormlogger.LogLevel) gormlogger.Interface {
	l, err := NewLogger(opts)
	if err != nil {
		// Fallback to default logger if creation fails
		l = GetDefaultLogger()
	}
	storeLogger := NewStoreLogger(l)
	return storeLogger.LogMode(logLevel)
}

// NewStoreLogger creates a GORM-compatible logger using the provided logger.
func NewStoreLogger(l Logger) gormlogger.Interface {
	return &storeLoggerAdapter{logger: l}
}

// storeLoggerAdapter implements GORM's logger interface using the generic logger.
type storeLoggerAdapter struct {
	logger Logger
}

// LogMode implements gorm logger interface.
func (l *storeLoggerAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// Create a copy of the logger to avoid affecting the original
	newLoggerInstance := l.logger.WithCallerSkip(0) // Create a copy
	newLogger := &storeLoggerAdapter{logger: newLoggerInstance}
	switch level {
	case gormlogger.Silent:
		newLogger.logger.SetLevel(FatalLevel)
	case gormlogger.Error:
		newLogger.logger.SetLevel(ErrorLevel)
	case gormlogger.Warn:
		newLogger.logger.SetLevel(WarnLevel)
	case gormlogger.Info:
		newLogger.logger.SetLevel(InfoLevel)
	}
	return newLogger
}

// Info implements gorm logger interface.
func (l *storeLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WithCtx(ctx, data...).Infof(msg, data...)
}

// Warn implements gorm logger interface.
func (l *storeLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WithCtx(ctx, data...).Warnf(msg, data...)
}

// Error implements gorm logger interface.
func (l *storeLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WithCtx(ctx, data...).Errorf(msg, data...)
}

// Trace implements gorm logger interface.
func (l *storeLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		l.logger.WithCtx(ctx, "sql", sql, "rows", rows, "elapsed", elapsed, "error", err).Errorw("SQL query failed",
			"sql", sql,
			"rows", rows,
			"elapsed", elapsed,
			"error", err,
		)
	} else {
		l.logger.WithCtx(ctx, "sql", sql, "rows", rows, "elapsed", elapsed).Debugw("SQL query executed",
			"sql", sql,
			"rows", rows,
			"elapsed", elapsed,
		)
	}
}
