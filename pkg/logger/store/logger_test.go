package store

import (
	"context"
	"testing"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gormlogger "gorm.io/gorm/logger"
)

// MockLogger implements the logger.Logger interface for testing
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
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Infow(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Warnw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Errorw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	m.Called(msg, keysAndValues)
}

func (m *MockLogger) With(keyValues ...interface{}) logger.Logger {
	args := m.Called(keyValues)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithCtx(ctx context.Context, keyValues ...interface{}) logger.Logger {
	args := m.Called(append([]interface{}{ctx}, keyValues...)...)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithCallerSkip(skip int) logger.Logger {
	args := m.Called(skip)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) SetLevel(level logger.Level) {
	m.Called(level)
}

func TestNewLogger(t *testing.T) {
	mockLogger := &MockLogger{}
	storeLogger := NewLogger(mockLogger)

	assert.NotNil(t, storeLogger)
	assert.Implements(t, (*gormlogger.Interface)(nil), storeLogger)
}

func TestStoreLoggerAdapter_LogMode(t *testing.T) {
	tests := []struct {
		name        string
		level       gormlogger.LogLevel
		expectLevel logger.Level
	}{
		{"Silent level", gormlogger.Silent, logger.FatalLevel},
		{"Error level", gormlogger.Error, logger.ErrorLevel},
		{"Warn level", gormlogger.Warn, logger.WarnLevel},
		{"Info level", gormlogger.Info, logger.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			mockCopiedLogger := &MockLogger{}
			storeLogger := NewLogger(mockLogger).(*storeLoggerAdapter)

			mockLogger.On("WithCallerSkip", 0).Return(mockCopiedLogger).Once()
			mockCopiedLogger.On("SetLevel", tt.expectLevel).Once()

			newLogger := storeLogger.LogMode(tt.level)

			assert.NotNil(t, newLogger)
			assert.IsType(t, &storeLoggerAdapter{}, newLogger)
			mockLogger.AssertExpectations(t)
			mockCopiedLogger.AssertExpectations(t)
		})
	}
}

func TestStoreLoggerAdapter_Info(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewLogger(mockLogger).(*storeLoggerAdapter)

	ctx := context.Background()
	msg := "test info message"
	data := []interface{}{"key1", "value1", "key2", "value2"}

	// WithCtx receives variadic args, so we need to match them individually
	mockLogger.On("WithCtx", ctx, "key1", "value1", "key2", "value2").Return(mockCtxLogger).Once()
	mockCtxLogger.On("Infof", msg, data).Once()

	storeLogger.Info(ctx, msg, data...)

	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Warn(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewLogger(mockLogger).(*storeLoggerAdapter)

	ctx := context.Background()
	msg := "test warn message"
	data := []interface{}{"key", "value"}

	mockLogger.On("WithCtx", ctx, "key", "value").Return(mockCtxLogger).Once()
	mockCtxLogger.On("Warnf", msg, data).Once()

	storeLogger.Warn(ctx, msg, data...)

	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Error(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewLogger(mockLogger).(*storeLoggerAdapter)

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
	storeLogger := NewLogger(mockLogger).(*storeLoggerAdapter)

	ctx := context.Background()
	begin := time.Now()
	sql := "SELECT * FROM users WHERE id = ?"
	rows := int64(5)

	fc := func() (string, int64) {
		return sql, rows
	}

	mockLogger.On("WithCtx", ctx, "sql", sql, "rows", rows, "elapsed", mock.AnythingOfType("time.Duration")).Return(mockCtxLogger).Once()

	mockCtxLogger.On("Debugw", "SQL query executed", mock.Anything).Once()

	storeLogger.Trace(ctx, begin, fc, nil)

	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Trace_WithError(t *testing.T) {
	mockLogger := &MockLogger{}
	mockCtxLogger := &MockLogger{}
	storeLogger := NewLogger(mockLogger).(*storeLoggerAdapter)

	ctx := context.Background()
	begin := time.Now()
	sql := "SELECT * FROM users WHERE id = ?"
	rows := int64(0)
	testErr := assert.AnError

	fc := func() (string, int64) {
		return sql, rows
	}

	mockLogger.On("WithCtx", ctx, "sql", sql, "rows", rows, "elapsed", mock.AnythingOfType("time.Duration"), "error", testErr).Return(mockCtxLogger).Once()

	mockCtxLogger.On("Errorw", "SQL query failed", mock.Anything).Once()

	storeLogger.Trace(ctx, begin, fc, testErr)

	mockLogger.AssertExpectations(t)
	mockCtxLogger.AssertExpectations(t)
}

func TestStoreLoggerAdapter_Integration(t *testing.T) {
	// Integration test using real logger
	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "debug",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		Development: true,
	}

	realLogger, err := logger.NewLogger(opts)
	assert.NoError(t, err)

	storeLogger := NewLogger(realLogger)
	assert.NotNil(t, storeLogger)

	ctx := context.Background()

	// Test different log levels
	storeLogger.Info(ctx, "Integration test info", "test", true)
	storeLogger.Warn(ctx, "Integration test warn", "level", "warn")
	storeLogger.Error(ctx, "Integration test error", "error", "test error")

	// Test trace logging
	begin := time.Now()
	fc := func() (string, int64) {
		return "SELECT COUNT(*) FROM test_table", 10
	}

	storeLogger.Trace(ctx, begin, fc, nil)
	storeLogger.Trace(ctx, begin, fc, assert.AnError)

	// Test log mode changes
	newLogger := storeLogger.LogMode(gormlogger.Error)
	assert.NotNil(t, newLogger)
	assert.NotEqual(t, storeLogger, newLogger) // Should return a new instance
}
