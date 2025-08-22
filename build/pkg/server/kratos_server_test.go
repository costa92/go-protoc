package server

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

func TestNewKratosLogger(t *testing.T) {
	id := "test-service-id"
	name := "test-service"
	version := "v1.0.0"
	
	kratosLogger := NewKratosLogger(id, name, version)
	
	assert.NotNil(t, kratosLogger)
	assert.Implements(t, (*krtlog.Logger)(nil), kratosLogger)
	
	// Test that logging works without errors
	err := kratosLogger.Log(krtlog.LevelInfo, "msg", "test message from original logger")
	assert.NoError(t, err)
}

func TestNewKratosLoggerWithGeneric(t *testing.T) {
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
	
	kratosLogger := NewKratosLoggerWithGeneric(mockLogger, id, name, version)
	
	assert.NotNil(t, kratosLogger)
	assert.Implements(t, (*krtlog.Logger)(nil), kratosLogger)
	mockLogger.AssertExpectations(t)
}

func TestNewKratosLoggerFromOptions_Success(t *testing.T) {
	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		Development: true,
	}
	
	id := "test-app"
	name := "test-service"
	version := "v1.0.0"
	
	kratosLogger := NewKratosLoggerFromOptions(opts, id, name, version)
	
	assert.NotNil(t, kratosLogger)
	assert.Implements(t, (*krtlog.Logger)(nil), kratosLogger)
	
	// Test that logging works
	err := kratosLogger.Log(krtlog.LevelInfo, "msg", "test message from options logger")
	assert.NoError(t, err)
}

func TestNewKratosLoggerFromOptions_WithNilOptions(t *testing.T) {
	id := "test-app"
	name := "test-service"
	version := "v1.0.0"
	
	kratosLogger := NewKratosLoggerFromOptions(nil, id, name, version)
	
	// Should fall back to default logger and still work
	assert.NotNil(t, kratosLogger)
	assert.Implements(t, (*krtlog.Logger)(nil), kratosLogger)
	
	// Test that logging works with fallback
	err := kratosLogger.Log(krtlog.LevelInfo, "msg", "test message with nil options")
	assert.NoError(t, err)
}

func TestNewKratosLoggerFromOptions_WithInvalidOptions(t *testing.T) {
	opts := &logger.LogsOptions{
		Type: "invalid-logger-type", // This will cause NewLogger to fail
	}
	
	id := "test-app"
	name := "test-service"
	version := "v1.0.0"
	
	kratosLogger := NewKratosLoggerFromOptions(opts, id, name, version)
	
	// Should fall back to default logger and still work
	assert.NotNil(t, kratosLogger)
	assert.Implements(t, (*krtlog.Logger)(nil), kratosLogger)
	
	// Test that logging works with fallback
	err := kratosLogger.Log(krtlog.LevelInfo, "msg", "test message with invalid options")
	assert.NoError(t, err)
}

func TestKratosLoggerIntegration_DifferentLoggers(t *testing.T) {
	id := "integration-app"
	name := "integration-service"
	version := "v2.0.0"
	
	// Test with different logger types
	tests := []struct {
		name string
		opts *logger.LogsOptions
	}{
		{
			name: "Zap logger",
			opts: &logger.LogsOptions{
				Type:        logger.LoggerTypeZap,
				Level:       "debug",
				Format:      "json",
				OutputPaths: []string{"stdout"},
				Development: true,
			},
		},
		{
			name: "Slog logger",
			opts: &logger.LogsOptions{
				Type:        logger.LoggerTypeSlog,
				Level:       "info",
				Format:      "text",
				OutputPaths: []string{"stdout"},
				Development: true,
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kratosLogger := NewKratosLoggerFromOptions(tt.opts, id, name, version)
			assert.NotNil(t, kratosLogger)
			
			// Test different log levels
			testCases := []struct {
				level   krtlog.Level
				message string
			}{
				{krtlog.LevelDebug, "debug message"},
				{krtlog.LevelInfo, "info message"},
				{krtlog.LevelWarn, "warn message"},
				{krtlog.LevelError, "error message"},
			}
			
			for _, tc := range testCases {
				err := kratosLogger.Log(tc.level, "msg", tc.message, "test", tt.name)
				assert.NoError(t, err)
			}
		})
	}
}

func TestKratosLoggerMetadata(t *testing.T) {
	// Test that service metadata is properly injected
	mockLogger := &MockLogger{}
	mockWithLogger := &MockLogger{}
	
	id := "metadata-test-id"
	name := "metadata-test-service"
	version := "v3.0.0"
	
	expectedWith := []interface{}{
		"service.id", id,
		"service.name", name,
		"service.version", version,
	}
	
	mockLogger.On("With", expectedWith).Return(mockWithLogger).Once()
	
	kratosLogger := NewKratosLoggerWithGeneric(mockLogger, id, name, version)
	
	assert.NotNil(t, kratosLogger)
	mockLogger.AssertExpectations(t)
	
	// The actual metadata injection is handled by the kratos.With wrapper
	// and our adapter, so we can test that the logger is properly created
	assert.Implements(t, (*krtlog.Logger)(nil), kratosLogger)
}

func TestKratosLoggerComparison(t *testing.T) {
	// Compare the original logger with the new generic loggers
	id := "comparison-test"
	name := "comparison-service"
	version := "v1.0.0"
	
	// Original logger (using pkg/log)
	originalLogger := NewKratosLogger(id, name, version)
	assert.NotNil(t, originalLogger)
	
	// Generic logger with Zap
	zapOpts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		Development: true,
	}
	genericZapLogger := NewKratosLoggerFromOptions(zapOpts, id, name, version)
	assert.NotNil(t, genericZapLogger)
	
	// Generic logger with Slog
	slogOpts := &logger.LogsOptions{
		Type:        logger.LoggerTypeSlog,
		Level:       "info",
		Format:      "text",
		OutputPaths: []string{"stdout"},
		Development: true,
	}
	genericSlogLogger := NewKratosLoggerFromOptions(slogOpts, id, name, version)
	assert.NotNil(t, genericSlogLogger)
	
	// Test that all loggers implement the same interface
	loggers := []krtlog.Logger{originalLogger, genericZapLogger, genericSlogLogger}
	
	for i, logger := range loggers {
		t.Run(func() string {
			switch i {
			case 0:
				return "original"
			case 1:
				return "generic-zap"
			case 2:
				return "generic-slog"
			default:
				return "unknown"
			}
		}(), func(t *testing.T) {
			assert.Implements(t, (*krtlog.Logger)(nil), logger)
			
			// Test that logging works for all
			err := logger.Log(krtlog.LevelInfo, "msg", "comparison test", "logger_index", i)
			assert.NoError(t, err)
		})
	}
}