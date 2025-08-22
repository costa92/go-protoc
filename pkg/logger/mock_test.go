package logger

import (
	"context"

	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/mock"
)

// MockLogger implements the Logger interface for testing
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Info(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Warn(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Error(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Fatal(args ...interface{}) {
	m.Called(args...)
}

func (m *MockLogger) Debugf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Infof(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Warnf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Errorf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Fatalf(template string, args ...interface{}) {
	m.Called(template, args)
}

func (m *MockLogger) Debugw(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Warnw(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	args := append([]interface{}{msg}, keysAndValues...)
	m.Called(args...)
}

func (m *MockLogger) With(keyValues ...interface{}) Logger {
	args := m.Called(keyValues)
	return args.Get(0).(Logger)
}

func (m *MockLogger) WithCtx(ctx context.Context, keyValues ...interface{}) Logger {
	args := m.Called(append([]interface{}{ctx}, keyValues...)...)
	return args.Get(0).(Logger)
}

func (m *MockLogger) WithCallerSkip(skip int) Logger {
	args := m.Called(skip)
	return args.Get(0).(Logger)
}

func (m *MockLogger) SetLevel(level Level) {
	m.Called(level)
}

// testKratosAdapter is a test helper to access the internal kratos adapter functionality
type testKratosAdapter struct {
	logger Logger
}

func (l *testKratosAdapter) Log(level krtlog.Level, keyvals ...interface{}) error {
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
