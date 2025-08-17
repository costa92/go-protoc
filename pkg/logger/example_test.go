package logger

import (
	"context"
	"fmt"
)

func ExampleNewLogger_zap() {
	opts := &LogsOptions{
		Type:              LoggerTypeZap,
		Level:             "info",
		Format:            "json",
		DisableCaller:     false,
		DisableStacktrace: false,
		EnableColor:       true,
		OutputPaths:       []string{"stdout"},
		Development:       true,
		CallerSkip:        1,
	}

	logger, err := NewLogger(opts)
	if err != nil {
		panic(err)
	}

	logger.Info("This is an info message")
	logger.Infof("This is an info message with format: %s", "value")
	logger.Infow("This is a structured info message", "key", "value", "number", 42)

	contextLogger := logger.With("service", "example", "version", "1.0.0")
	contextLogger.Info("This message includes context fields")

	ctx := context.Background()
	ctxLogger := logger.WithCtx(ctx, "request_id", "12345")
	ctxLogger.Info("This message includes context and request ID")
}

func ExampleNewLogger_slog() {
	opts := &LogsOptions{
		Type:              LoggerTypeSlog,
		Level:             "debug",
		Format:            "text",
		DisableCaller:     false,
		DisableStacktrace: false,
		EnableColor:       true,
		OutputPaths:       []string{"stdout"},
		CallerSkip:        1,
	}

	logger, err := NewLogger(opts)
	if err != nil {
		panic(err)
	}

	logger.Debug("This is a debug message")
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
	logger.Error("This is an error message")

	logger.Debugf("Debug message with format: %d", 123)
	logger.Debugw("Structured debug message", "user_id", 456, "action", "login")

	serviceLogger := logger.With("component", "auth", "module", "jwt")
	serviceLogger.Info("Authentication successful")
}

func ExampleSetDefaultLogger() {
	logger, _ := NewLogger(&LogsOptions{
		Type:        LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		Development: true,
	})
	SetDefaultLogger(logger)

	// Using global functions that use the default logger
	GetDefaultLogger().Info("Using default logger")
	GetDefaultLogger().Infof("Default logger with format: %s", "example")
	GetDefaultLogger().Infow("Default structured logging", "key", "value")

	GetDefaultLogger().Debug("This won't show because level is info")
}

func ExampleLoggerImpl() {
	opts := DefaultOptions()
	opts.Type = LoggerTypeZap
	opts.Development = true

	impl, err := NewLoggerImpl(opts)
	if err != nil {
		panic(err)
	}

	impl.Info("Using LoggerImpl wrapper")
	impl.Infow("Structured message", "wrapper", "LoggerImpl", "underlying", "zap")

	withContext := impl.WithCtx(context.Background(), "trace_id", "abc123")
	withContext.Info("Message with context")
}

func ExampleParseLevel() {
	for _, level := range []string{"debug", "info", "warn", "error", "fatal"} {
		parsedLevel := ParseLevel(level)
		fmt.Printf("Level: %s -> %s (zap: %v, slog: %v)\n",
			level,
			parsedLevel.String(),
			parsedLevel.ToZapLevel(),
			parsedLevel.ToSlogLevel())
	}
}
