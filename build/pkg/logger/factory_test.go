package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFactoryFunctions(t *testing.T) {
	// Test that the global functions don't panic
	Debug("test debug")
	Info("test info")
	Warn("test warn")
	Error("test error")
	
	Debugf("test %s", "debugf")
	Infof("test %s", "infof")
	Warnf("test %s", "warnf")
	Errorf("test %s", "errorf")
	
	Debugw("test debugw", "key", "value")
	Infow("test infow", "key", "value")
	Warnw("test warnw", "key", "value")
	Errorw("test errorw", "key", "value")
}

func TestWith(t *testing.T) {
	logger := With("service", "test")
	assert.NotNil(t, logger)
	assert.Implements(t, (*Logger)(nil), logger)
}

func TestWithCtx(t *testing.T) {
	ctx := context.Background()
	logger := WithCtx(ctx, "request_id", "123")
	assert.NotNil(t, logger)
	assert.Implements(t, (*Logger)(nil), logger)
}

func TestGetDefaultLogger(t *testing.T) {
	logger := GetDefaultLogger()
	assert.NotNil(t, logger)
	assert.Implements(t, (*Logger)(nil), logger)
}

func TestSetDefaultLogger(t *testing.T) {
	opts := &LogsOptions{
		Type:        LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		Development: true,
	}
	
	newLogger, err := NewLogger(opts)
	assert.NoError(t, err)
	
	SetDefaultLogger(newLogger)
	
	retrievedLogger := GetDefaultLogger()
	assert.NotNil(t, retrievedLogger)
}