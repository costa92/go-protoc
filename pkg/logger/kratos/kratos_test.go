package kratos

import (
	"context"
	"testing"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	krtlog "github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockLogger) With(keyValues ...interface{}) logger.Logger {
	args := m.Called(keyValues)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) WithCtx(ctx context.Context, keyValues ...interface{}) logger.Logger {
	args := m.Called(ctx, keyValues)
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

	kratosLogger := NewLogger(mockLogger, id, name, version)

	assert.NotNil(t, kratosLogger)
	mockLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Log_Debug(t *testing.T) {
	mockLogger := &MockLogger{}
	adapter := &kratosLoggerAdapter{logger: mockLogger}

	keyvals := []interface{}{"key1", "value1", "key2", "value2"}
	mockLogger.On("Debugw", "", "key1", "value1", "key2", "value2").Once()

	err := adapter.Log(krtlog.LevelDebug, keyvals...)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Log_Info(t *testing.T) {
	mockLogger := &MockLogger{}
	adapter := &kratosLoggerAdapter{logger: mockLogger}

	keyvals := []interface{}{"message", "test info"}
	mockLogger.On("Infow", "", "message", "test info").Once()

	err := adapter.Log(krtlog.LevelInfo, keyvals...)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Log_Warn(t *testing.T) {
	mockLogger := &MockLogger{}
	adapter := &kratosLoggerAdapter{logger: mockLogger}

	keyvals := []interface{}{"warning", "test warning"}
	mockLogger.On("Warnw", "", "warning", "test warning").Once()

	err := adapter.Log(krtlog.LevelWarn, keyvals...)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Log_Error(t *testing.T) {
	mockLogger := &MockLogger{}
	adapter := &kratosLoggerAdapter{logger: mockLogger}

	keyvals := []interface{}{"error", "test error"}
	mockLogger.On("Errorw", "", "error", "test error").Once()

	err := adapter.Log(krtlog.LevelError, keyvals...)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Log_Fatal(t *testing.T) {
	mockLogger := &MockLogger{}
	adapter := &kratosLoggerAdapter{logger: mockLogger}

	keyvals := []interface{}{"fatal", "test fatal"}
	mockLogger.On("Fatalw", "", "fatal", "test fatal").Once()

	err := adapter.Log(krtlog.LevelFatal, keyvals...)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Log_UnknownLevel(t *testing.T) {
	mockLogger := &MockLogger{}
	adapter := &kratosLoggerAdapter{logger: mockLogger}

	keyvals := []interface{}{"unknown", "test"}
	mockLogger.On("Infow", "", "unknown", "test").Once() // Should default to Info

	err := adapter.Log(krtlog.Level(99), keyvals...) // Unknown level (use a level that doesn't exist)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestKratosLoggerAdapter_Integration(t *testing.T) {
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

	kratosLogger := NewLogger(realLogger, "test-app", "test-service", "v1.0.0")
	assert.NotNil(t, kratosLogger)

	// Test logging at different levels
	err = kratosLogger.Log(krtlog.LevelDebug, "msg", "debug message", "component", "test")
	assert.NoError(t, err)

	err = kratosLogger.Log(krtlog.LevelInfo, "msg", "info message", "service", "test")
	assert.NoError(t, err)

	err = kratosLogger.Log(krtlog.LevelWarn, "msg", "warn message", "warning", true)
	assert.NoError(t, err)

	err = kratosLogger.Log(krtlog.LevelError, "msg", "error message", "error", "test error")
	assert.NoError(t, err)
}

// Helper function to extract the adapter from the kratos.With wrapper
func extractAdapter(t *testing.T, kratosLogger krtlog.Logger) *kratosLoggerAdapter {
	// Since kratos.With wraps our adapter, we need to test through the wrapper
	// or access the adapter directly for unit testing purposes

	// For testing, we'll create a simple adapter directly
	mockLogger := &MockLogger{}
	mockWithLogger := &MockLogger{}
	mockLogger.On("With", mock.Anything).Return(mockWithLogger)

	// Create adapter directly for testing
	adapter := &kratosLoggerAdapter{
		logger: mockWithLogger,
	}

	return adapter
}

func TestKratosLoggerAdapter_DirectAccess(t *testing.T) {
	// Test the adapter directly without kratos.With wrapper
	mockLogger := &MockLogger{}

	adapter := &kratosLoggerAdapter{
		logger: mockLogger,
	}

	tests := []struct {
		name     string
		level    krtlog.Level
		keyvals  []interface{}
		expected string // expected method call
	}{
		{"Debug level", krtlog.LevelDebug, []interface{}{"key", "value"}, "Debugw"},
		{"Info level", krtlog.LevelInfo, []interface{}{"msg", "info"}, "Infow"},
		{"Warn level", krtlog.LevelWarn, []interface{}{"warning", true}, "Warnw"},
		{"Error level", krtlog.LevelError, []interface{}{"error", "failed"}, "Errorw"},
		{"Fatal level", krtlog.LevelFatal, []interface{}{"fatal", "critical"}, "Fatalw"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]interface{}{""}, tt.keyvals...)
			mockLogger.On(tt.expected, args...).Once()

			err := adapter.Log(tt.level, tt.keyvals...)

			assert.NoError(t, err)
			mockLogger.AssertExpectations(t)
		})
	}
}
