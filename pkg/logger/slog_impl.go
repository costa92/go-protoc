package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type SlogLogger struct {
	logger     *slog.Logger
	opts       *LogsOptions
	level      Level
	filter     Filter
	callerSkip int
}

var _ Logger = (*SlogLogger)(nil)

func NewSlogLogger(opts *LogsOptions) (*SlogLogger, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	var writers []io.Writer
	for _, outputPath := range opts.OutputPaths {
		var writer io.Writer

		if outputPath == "stdout" {
			writer = os.Stdout
		} else if outputPath == "stderr" {
			writer = os.Stderr
		} else {
			lumberJackLogger := &lumberjack.Logger{
				Filename:   outputPath,
				MaxSize:    opts.MaxSize,
				MaxAge:     opts.MaxAge,
				MaxBackups: opts.MaxBackups,
				Compress:   opts.Compress,
			}
			writer = lumberJackLogger
		}
		writers = append(writers, writer)
	}

	var output io.Writer
	if len(writers) == 1 {
		output = writers[0]
	} else {
		output = io.MultiWriter(writers...)
	}

	handlerOpts := &slog.HandlerOptions{
		Level:     ParseLevel(opts.Level).ToSlogLevel(),
		AddSource: !opts.DisableCaller,
	}

	var handler slog.Handler
	if opts.Format == "json" || opts.Encoding == "json" {
		handler = slog.NewJSONHandler(output, handlerOpts)
	} else {
		textOpts := &slog.HandlerOptions{
			Level:     handlerOpts.Level,
			AddSource: handlerOpts.AddSource,
		}
		handler = slog.NewTextHandler(output, textOpts)
	}

	logger := slog.New(&callerHandler{
		Handler:    handler,
		callerSkip: opts.CallerSkip,
	})

	// 自动添加 logger 类型标识
	logger = logger.With("type", "slog")

	// 如果有初始字段配置，也添加进去
	if opts.InitialFields != nil {
		var args []interface{}
		for k, v := range opts.InitialFields {
			args = append(args, k, v)
		}
		if len(args) > 0 {
			logger = logger.With(args...)
		}
	}

	return &SlogLogger{
		logger: logger,
		opts:   opts,
		level:  ParseLevel(opts.Level),
	}, nil
}

type callerHandler struct {
	slog.Handler
	callerSkip int
}

func (h *callerHandler) Handle(ctx context.Context, r slog.Record) error {
	if h.callerSkip > 0 {
		pc, file, line, ok := runtime.Caller(h.callerSkip + 3)
		if ok {
			if idx := strings.LastIndex(file, "/"); idx >= 0 {
				file = file[idx+1:]
			}

			fn := runtime.FuncForPC(pc)
			function := "unknown"
			if fn != nil {
				function = fn.Name()
				if idx := strings.LastIndex(function, "."); idx >= 0 {
					function = function[idx+1:]
				}
			}

			r.AddAttrs(
				slog.String("file", file),
				slog.Int("line", line),
				slog.String("func", function),
			)
		}
	}
	return h.Handler.Handle(ctx, r)
}

func (s *SlogLogger) getCaller(skip int) (file string, line int, function string) {
	pc, file, line, ok := runtime.Caller(skip + s.opts.CallerSkip + s.callerSkip + 1)
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

func (s *SlogLogger) shouldLog(level Level, ctx context.Context, msg string) bool {
	if level < s.level {
		return false
	}

	if s.filter != nil && !s.filter.IsEmpty() {
		file, line, function := s.getCaller(3)
		meta := &Metadata{
			File: file,
			Line: line,
			Func: function,
			Msg:  msg,
		}
		return s.filter.Enabled(ctx, meta)
	}

	return true
}

func (s *SlogLogger) Debug(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if s.shouldLog(DebugLevel, context.Background(), msg) {
		s.logger.Debug(msg)
	}
}

func (s *SlogLogger) Info(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if s.shouldLog(InfoLevel, context.Background(), msg) {
		s.logger.Info(msg)
	}
}

func (s *SlogLogger) Warn(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if s.shouldLog(WarnLevel, context.Background(), msg) {
		s.logger.Warn(msg)
	}
}

func (s *SlogLogger) Error(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if s.shouldLog(ErrorLevel, context.Background(), msg) {
		s.logger.Error(msg)
	}
}

func (s *SlogLogger) Fatal(args ...interface{}) {
	msg := fmt.Sprint(args...)
	if s.shouldLog(FatalLevel, context.Background(), msg) {
		s.logger.Log(context.Background(), ParseLevel("fatal").ToSlogLevel(), msg)
		os.Exit(1)
	}
}

func (s *SlogLogger) Debugf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if s.shouldLog(DebugLevel, context.Background(), msg) {
		s.logger.Debug(msg)
	}
}

func (s *SlogLogger) Infof(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if s.shouldLog(InfoLevel, context.Background(), msg) {
		s.logger.Info(msg)
	}
}

func (s *SlogLogger) Warnf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if s.shouldLog(WarnLevel, context.Background(), msg) {
		s.logger.Warn(msg)
	}
}

func (s *SlogLogger) Errorf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if s.shouldLog(ErrorLevel, context.Background(), msg) {
		s.logger.Error(msg)
	}
}

func (s *SlogLogger) Fatalf(template string, args ...interface{}) {
	msg := fmt.Sprintf(template, args...)
	if s.shouldLog(FatalLevel, context.Background(), msg) {
		s.logger.Log(context.Background(), ParseLevel("fatal").ToSlogLevel(), msg)
		os.Exit(1)
	}
}

func (s *SlogLogger) Debugw(msg string, keysAndValues ...interface{}) {
	if s.shouldLog(DebugLevel, context.Background(), msg) {
		attrs := s.keysAndValuesToAttrs(keysAndValues...)
		s.logger.LogAttrs(context.Background(), slog.LevelDebug, msg, attrs...)
	}
}

func (s *SlogLogger) Infow(msg string, keysAndValues ...interface{}) {
	if s.shouldLog(InfoLevel, context.Background(), msg) {
		attrs := s.keysAndValuesToAttrs(keysAndValues...)
		s.logger.LogAttrs(context.Background(), slog.LevelInfo, msg, attrs...)
	}
}

func (s *SlogLogger) Warnw(msg string, keysAndValues ...interface{}) {
	if s.shouldLog(WarnLevel, context.Background(), msg) {
		attrs := s.keysAndValuesToAttrs(keysAndValues...)
		s.logger.LogAttrs(context.Background(), slog.LevelWarn, msg, attrs...)
	}
}

func (s *SlogLogger) Errorw(msg string, keysAndValues ...interface{}) {
	if s.shouldLog(ErrorLevel, context.Background(), msg) {
		attrs := s.keysAndValuesToAttrs(keysAndValues...)
		s.logger.LogAttrs(context.Background(), slog.LevelError, msg, attrs...)
	}
}

func (s *SlogLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	if s.shouldLog(FatalLevel, context.Background(), msg) {
		attrs := s.keysAndValuesToAttrs(keysAndValues...)
		s.logger.LogAttrs(context.Background(), ParseLevel("fatal").ToSlogLevel(), msg, attrs...)
		os.Exit(1)
	}
}

func (s *SlogLogger) keysAndValuesToAttrs(keysAndValues ...interface{}) []slog.Attr {
	var attrs []slog.Attr
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := fmt.Sprint(keysAndValues[i])
			value := keysAndValues[i+1]
			attrs = append(attrs, slog.Any(key, value))
		}
	}
	return attrs
}

func (s *SlogLogger) With(keyValues ...interface{}) Logger {
	attrs := s.keysAndValuesToAttrs(keyValues...)
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	newLogger := &SlogLogger{
		logger:     s.logger.With(anyAttrs...),
		opts:       s.opts,
		level:      s.level,
		filter:     s.filter,
		callerSkip: s.callerSkip,
	}
	return newLogger
}

func (s *SlogLogger) WithCtx(ctx context.Context, keyValues ...interface{}) Logger {
	return s.With(keyValues...)
}

func (s *SlogLogger) WithCallerSkip(skip int) Logger {
	newLogger := &SlogLogger{
		logger:     s.logger,
		opts:       s.opts,
		level:      s.level,
		filter:     s.filter,
		callerSkip: s.callerSkip + skip,
	}
	return newLogger
}

func (s *SlogLogger) SetLevel(level Level) {
	s.level = level
}
