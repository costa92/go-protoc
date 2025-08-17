package logger

import (
	"context"
	"testing"
	"time"

	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gormlogger "gorm.io/gorm/logger"
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

func TestNewGormLogger_Success(t *testing.T) {
	opts := &LogsOptions{
		Type:        LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		Development: true,
	}
	
	gormLogger := NewGormLogger(opts, gormlogger.Info)
	
	assert.NotNil(t, gormLogger)
	assert.Implements(t, (*gormlogger.Interface)(nil), gormLogger)
}

func TestNewGormLogger_WithNilOptions(t *testing.T) {
	gormLogger := NewGormLogger(nil, gormlogger.Error)
	
	assert.NotNil(t, gormLogger)
	assert.Implements(t, (*gormlogger.Interface)(nil), gormLogger)
}

func TestNewGormLogger_WithInvalidOptions(t *testing.T) {
	opts := &LogsOptions{
		Type: "invalid-type", // This will cause NewLogger to fail
	}
	
	gormLogger := NewGormLogger(opts, gormlogger.Warn)
	
	// Should fallback to default logger and still work
	assert.NotNil(t, gormLogger)
	assert.Implements(t, (*gormlogger.Interface)(nil), gormLogger)
}

func TestNewStoreLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	
	storeLogger := NewStoreLogger(mockLogger)
	
	assert.NotNil(t, storeLogger)
	assert.Implements(t, (*gormlogger.Interface)(nil), storeLogger)
	
	// Test that it works by calling a method
	ctx := context.Background()
	mockCtxLogger := &MockLogger{}
	mockLogger.On("WithCtx", ctx, "test", "value").Return(mockCtxLogger).Once()
	mockCtxLogger.On("Infof", "test message", []interface{}{"test", "value"}).Once()
	
	storeLogger.Info(ctx, "test message", "test", "value")
	
	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestNewKratosLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	mockWithLogger := &MockLogger{}
	
	id := "test-service-id"
	name := "test-service"
	version := "v1.0.0"
	
	expectedWith := []interface{}{
		"service.id", id,
		"service.name", name,
		"service.version", version,
	}
	
	mockLogger.On("With", expectedWith).Return(mockWithLogger).Once()
	
	kratosLogger := NewKratosLogger(mockLogger, id, name, version)
	
	assert.NotNil(t, kratosLogger)
	assert.Implements(t, (*krtlog.Logger)(nil), kratosLogger)
	mockLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_LogMode(t *testing.T) {
	mockLogger := &MockLogger{}
	storeLogger := NewStoreLogger(mockLogger)
	
	tests := []struct {
		name        string
		level       gormlogger.LogLevel
		expectLevel Level
	}{
		{"Silent level", gormlogger.Silent, FatalLevel},
		{"Error level", gormlogger.Error, ErrorLevel},
		{"Warn level", gormlogger.Warn, WarnLevel},
		{"Info level", gormlogger.Info, InfoLevel},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockCopiedLogger := &MockLogger{}
			mockLogger.On("WithCallerSkip", 0).Return(mockCopiedLogger).Once()
			mockCopiedLogger.On("SetLevel", tt.expectLevel).Once()
			
			newLogger := storeLogger.LogMode(tt.level)
			
			assert.NotNil(t, newLogger)
			assert.Implements(t, (*gormlogger.Interface)(nil), newLogger)
			mockLogger.AssertExpectations(t)
			mockCopiedLogger.AssertExpectations(t)
		})
	}
}

func TestStoreLoggerAdapter_Info(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewStoreLogger(mockLogger)
	
	ctx := context.Background()
	msg := "test info message"
	data := []interface{}{"key1", "value1"}
	
	mockLogger.On("WithCtx", ctx, "key1", "value1").Return(mockCtxLogger).Once()
	mockCtxLogger.On("Infof", msg, data).Once()
	
	storeLogger.Info(ctx, msg, data...)
	
	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Warn(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewStoreLogger(mockLogger)
	
	ctx := context.Background()
	msg := "test warn message"
	data := []interface{}{"warning", true}
	
	mockLogger.On("WithCtx", ctx, "warning", true).Return(mockCtxLogger).Once()
	mockCtxLogger.On("Warnf", msg, data).Once()
	
	storeLogger.Warn(ctx, msg, data...)
	
	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Error(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewStoreLogger(mockLogger)
	
	ctx := context.Background()
	msg := "test error message"
	data := []interface{}{"error", "test error"}
	
	mockLogger.On("WithCtx", ctx, "error", "test error").Return(mockCtxLogger).Once()
	mockCtxLogger.On("Errorf", msg, data).Once()
	
	storeLogger.Error(ctx, msg, data...)
	
	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Trace_Success(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewStoreLogger(mockLogger)
	
	ctx := context.Background()
	begin := time.Now()
	sql := "SELECT * FROM users"
	rows := int64(10)
	
	fc := func() (string, int64) {
		return sql, rows
	}
	
	mockLogger.On("WithCtx", ctx, "sql", sql, "rows", rows, "elapsed", mock.AnythingOfType("time.Duration")).Return(mockCtxLogger).Once()
	
	mockCtxLogger.On("Debugw", "SQL query executed", "sql", sql, "rows", rows, "elapsed", mock.AnythingOfType("time.Duration")).Once()
	
	storeLogger.Trace(ctx, begin, fc, nil)
	
	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Trace_WithError(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewStoreLogger(mockLogger)
	
	ctx := context.Background()
	begin := time.Now()
	sql := "INSERT INTO users"
	rows := int64(0)
	testErr := assert.AnError
	
	fc := func() (string, int64) {
		return sql, rows
	}
	
	mockLogger.On("WithCtx", ctx, "sql", sql, "rows", rows, "elapsed", mock.AnythingOfType("time.Duration"), "error", testErr).Return(mockCtxLogger).Once()
	
	mockCtxLogger.On("Errorw", "SQL query failed", "sql", sql, "rows", rows, "elapsed", mock.AnythingOfType("time.Duration"), "error", testErr).Once()
	
	storeLogger.Trace(ctx, begin, fc, testErr)
	
	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Log(t *testing.T) {
	mockLogger := &MockLogger{}
	// For unit testing purposes, we'll create a minimal adapter
	adapter := &testKratosAdapter{logger: mockLogger}
	
	tests := []struct {
		name     string
		level    krtlog.Level
		keyvals  []interface{}
		expected string
	}{
		{"Debug level", krtlog.LevelDebug, []interface{}{"msg", "debug"}, "Debugw"},
		{"Info level", krtlog.LevelInfo, []interface{}{"msg", "info"}, "Infow"},
		{"Warn level", krtlog.LevelWarn, []interface{}{"msg", "warn"}, "Warnw"},
		{"Error level", krtlog.LevelError, []interface{}{"msg", "error"}, "Errorw"},
		{"Fatal level", krtlog.LevelFatal, []interface{}{"msg", "fatal"}, "Fatalw"},
		{"Unknown level", krtlog.Level(99), []interface{}{"msg", "unknown"}, "Infow"}, // defaults to Info
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]interface{}{""}, tt.keyvals...)
			mockLogger.On(tt.expected, args...).Return().Once()
			
			err := adapter.Log(tt.level, tt.keyvals...)
			
			assert.NoError(t, err)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestIntegration_RealLoggers(t *testing.T) {
	// Test with real zap logger
	zapOpts := &LogsOptions{
		Type:        LoggerTypeZap,
		Level:       "debug",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		Development: true,
	}
	
	zapLogger, err := NewLogger(zapOpts)
	assert.NoError(t, err)
	
	// Test store logger with real logger
	storeLogger := NewStoreLogger(zapLogger)
	assert.NotNil(t, storeLogger)
	
	ctx := context.Background()
	storeLogger.Info(ctx, "Integration test info", "test", true)
	
	// Test GORM logger creation
	gormLogger := NewGormLogger(zapOpts, gormlogger.Info)
	assert.NotNil(t, gormLogger)
	
	// Test kratos logger with real logger
	kratosLogger := NewKratosLogger(zapLogger, "test-app", "test-service", "v1.0.0")
	assert.NotNil(t, kratosLogger)
	
	err = kratosLogger.Log(krtlog.LevelInfo, "msg", "Integration test", "framework", "kratos")
	assert.NoError(t, err)
	
	// Test with slog logger
	slogOpts := &LogsOptions{
		Type:        LoggerTypeSlog,
		Level:       "info",
		Format:      "text",
		OutputPaths: []string{"stdout"},
		Development: true,
	}
	
	slogLogger, err := NewLogger(slogOpts)
	assert.NoError(t, err)
	
	slogStoreLogger := NewStoreLogger(slogLogger)
	assert.NotNil(t, slogStoreLogger)
	
	slogKratosLogger := NewKratosLogger(slogLogger, "slog-app", "slog-service", "v2.0.0")
	assert.NotNil(t, slogKratosLogger)
}