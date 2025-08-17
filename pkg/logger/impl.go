package logger

import (
	"context"
	"fmt"
	"log/slog"
)

type LoggerImpl struct {
	ctx           context.Context
	Option        *LogsOptions
	handler       slog.Handler
	AddCallerSkip int
	closeFuncs    []func(ctx context.Context) error
	filter        Filter
	underlying    Logger
}

var _ Logger = (*LoggerImpl)(nil)

func NewLoggerImpl(opts *LogsOptions) (*LoggerImpl, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	underlying, err := NewLogger(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create underlying logger: %w", err)
	}

	return &LoggerImpl{
		ctx:           context.Background(),
		Option:        opts,
		AddCallerSkip: opts.CallerSkip,
		underlying:    underlying,
	}, nil
}

func (l *LoggerImpl) clone() *LoggerImpl {
	c := *l
	c.Option = l.Option.clone()
	return &c
}

func (l *LoggerImpl) Debug(args ...interface{}) {
	l.underlying.Debug(args...)
}

func (l *LoggerImpl) Info(args ...interface{}) {
	l.underlying.Info(args...)
}

func (l *LoggerImpl) Warn(args ...interface{}) {
	l.underlying.Warn(args...)
}

func (l *LoggerImpl) Error(args ...interface{}) {
	l.underlying.Error(args...)
}

func (l *LoggerImpl) Fatal(args ...interface{}) {
	l.underlying.Fatal(args...)
}

func (l *LoggerImpl) Debugf(template string, args ...interface{}) {
	l.underlying.Debugf(template, args...)
}

func (l *LoggerImpl) Infof(template string, args ...interface{}) {
	l.underlying.Infof(template, args...)
}

func (l *LoggerImpl) Warnf(template string, args ...interface{}) {
	l.underlying.Warnf(template, args...)
}

func (l *LoggerImpl) Errorf(template string, args ...interface{}) {
	l.underlying.Errorf(template, args...)
}

func (l *LoggerImpl) Fatalf(template string, args ...interface{}) {
	l.underlying.Fatalf(template, args...)
}

func (l *LoggerImpl) Debugw(msg string, keysAndValues ...interface{}) {
	l.underlying.Debugw(msg, keysAndValues...)
}

func (l *LoggerImpl) Infow(msg string, keysAndValues ...interface{}) {
	l.underlying.Infow(msg, keysAndValues...)
}

func (l *LoggerImpl) Warnw(msg string, keysAndValues ...interface{}) {
	l.underlying.Warnw(msg, keysAndValues...)
}

func (l *LoggerImpl) Errorw(msg string, keysAndValues ...interface{}) {
	l.underlying.Errorw(msg, keysAndValues...)
}

func (l *LoggerImpl) Fatalw(msg string, keysAndValues ...interface{}) {
	l.underlying.Fatalw(msg, keysAndValues...)
}

func (l *LoggerImpl) With(keyValues ...interface{}) Logger {
	newImpl := l.clone()
	newImpl.underlying = l.underlying.With(keyValues...)
	return newImpl
}

func (l *LoggerImpl) WithCtx(ctx context.Context, keyValues ...interface{}) Logger {
	newImpl := l.clone()
	newImpl.ctx = ctx
	newImpl.underlying = l.underlying.WithCtx(ctx, keyValues...)
	return newImpl
}

func (l *LoggerImpl) WithCallerSkip(skip int) Logger {
	newImpl := l.clone()
	newImpl.underlying = l.underlying.WithCallerSkip(skip)
	return newImpl
}

func (l *LoggerImpl) SetLevel(level Level) {
	l.underlying.SetLevel(level)
}
