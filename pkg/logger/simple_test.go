package logger_test

import (
	"testing"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSimpleLoggerUsage 测试基本的 logger 使用
func TestSimpleLoggerUsage(t *testing.T) {
	// 测试 Zap logger
	t.Run("ZapLogger", func(t *testing.T) {
		opts := &logger.LogsOptions{
			Type:        logger.LoggerTypeZap,
			Level:       "info",
			Format:      "json",
			OutputPaths: []string{"stdout"},
		}

		l, err := logger.NewLogger(opts)
		require.NoError(t, err)
		require.NotNil(t, l)

		// 测试基本日志功能
		l.Info("test info message")
		l.Infof("formatted %s", "message")
		l.Infow("structured", "key", "value", "number", 123)
	})

	// 测试 Slog logger
	t.Run("SlogLogger", func(t *testing.T) {
		opts := &logger.LogsOptions{
			Type:        logger.LoggerTypeSlog,
			Level:       "debug",
			Format:      "json",
			OutputPaths: []string{"stdout"},
		}

		l, err := logger.NewLogger(opts)
		require.NoError(t, err)
		require.NotNil(t, l)

		// 测试基本日志功能
		l.Debug("test debug message")
		l.Debugf("formatted %s", "debug")
		l.Debugw("structured", "key", "value", "bool", true)
	})
}

// TestLoggerTypeField 测试自动添加的 type 字段
func TestLoggerTypeField(t *testing.T) {
	// 这个测试主要是验证 logger 创建时会自动添加 type 字段
	// 实际验证需要检查输出，这里只验证创建成功

	t.Run("Zap has type field", func(t *testing.T) {
		opts := &logger.LogsOptions{
			Type:        logger.LoggerTypeZap,
			Level:       "info",
			OutputPaths: []string{"stdout"},
		}

		l, err := logger.NewLogger(opts)
		require.NoError(t, err)

		// 使用 logger，type 字段会自动包含在输出中
		l.Info("message with auto type field")
	})

	t.Run("Slog has type field", func(t *testing.T) {
		opts := &logger.LogsOptions{
			Type:        logger.LoggerTypeSlog,
			Level:       "info",
			OutputPaths: []string{"stdout"},
		}

		l, err := logger.NewLogger(opts)
		require.NoError(t, err)

		// 使用 logger，type 字段会自动包含在输出中
		l.Info("message with auto type field")
	})
}

// TestDynamicLoggerBasic 测试动态 logger 基础功能
func TestDynamicLoggerBasic(t *testing.T) {
	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}

	dl, err := logger.NewDynamicLogger(opts)
	require.NoError(t, err)
	require.NotNil(t, dl)

	// 测试基本功能
	dl.Info("dynamic logger info")

	// 测试级别更新
	err = dl.UpdateLevel("debug")
	assert.NoError(t, err)

	dl.Debug("debug after level update")

	// 测试获取配置
	currentOpts := dl.GetOptions()
	assert.Equal(t, "debug", currentOpts.Level)
}

// TestConfigValidation 测试配置验证基础功能
func TestConfigValidation(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		opts := &logger.LogsOptions{
			Type:             logger.LoggerTypeZap,
			Level:            "info",
			Format:           "json",
			OutputPaths:      []string{"stdout"},
			ErrorOutputPaths: []string{"stderr"},
		}

		warnings, err := logger.ValidateConfig(opts)
		assert.NoError(t, err)
		assert.Empty(t, warnings)
	})

	t.Run("invalid logger type", func(t *testing.T) {
		opts := &logger.LogsOptions{
			Type: "invalid-type",
		}

		warnings, err := logger.ValidateConfig(opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid logger type")
		_ = warnings
	})

	t.Run("warnings without error", func(t *testing.T) {
		opts := &logger.LogsOptions{
			Type:        "", // 空类型会使用默认值，产生警告
			Level:       "info",
			OutputPaths: []string{"stdout"},
		}

		warnings, err := logger.ValidateConfig(opts)
		assert.NoError(t, err)
		assert.NotEmpty(t, warnings)
		assert.Contains(t, warnings[0], "logger type is empty")
	})
}
