package logger_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// TestConfigValidator 测试配置验证器
func TestConfigValidator(t *testing.T) {
	tests := []struct {
		name         string
		opts         *logger.LogsOptions
		wantErr      bool
		wantWarnings int
		checkError   func(t *testing.T, err error)
		checkWarning func(t *testing.T, warnings []string)
	}{
		{
			name: "valid configuration",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
			},
			wantErr:      false,
			wantWarnings: 0,
		},
		{
			name:    "nil configuration",
			opts:    nil,
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "configuration is nil")
			},
		},
		{
			name: "invalid logger type",
			opts: &logger.LogsOptions{
				Type:             "invalid-type",
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid logger type")
			},
		},
		{
			name: "empty logger type with warning",
			opts: &logger.LogsOptions{
				Type:             "",
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
			},
			wantErr:      false,
			wantWarnings: 1,
			checkWarning: func(t *testing.T, warnings []string) {
				assert.Contains(t, warnings[0], "logger type is empty")
			},
		},
		{
			name: "invalid log level",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "invalid-level",
				Format:           "json",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid log level")
			},
		},
		{
			name: "empty log level with warning",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
			},
			wantErr:      false,
			wantWarnings: 2, // empty level + empty format
			checkWarning: func(t *testing.T, warnings []string) {
				hasLevelWarning := false
				for _, w := range warnings {
					if strings.Contains(w, "log level is empty") {
						hasLevelWarning = true
						break
					}
				}
				assert.True(t, hasLevelWarning)
			},
		},
		{
			name: "invalid format",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "invalid-format",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid format")
			},
		},
		{
			name: "no output paths with warning",
			opts: &logger.LogsOptions{
				Type:        logger.LoggerTypeZap,
				Level:       "info",
				Format:      "json",
				OutputPaths: []string{},
			},
			wantErr:      false,
			wantWarnings: 2, // no output paths + no error output paths
			checkWarning: func(t *testing.T, warnings []string) {
				hasOutputWarning := false
				for _, w := range warnings {
					if strings.Contains(w, "no output paths specified") {
						hasOutputWarning = true
						break
					}
				}
				assert.True(t, hasOutputWarning)
			},
		},
		{
			name: "output path directory does not exist",
			opts: &logger.LogsOptions{
				Type:        logger.LoggerTypeZap,
				Level:       "info",
				Format:      "json",
				OutputPaths: []string{"/nonexistent/directory/test.log"},
			},
			wantErr:      false,
			wantWarnings: 3, // directory not exist + path may not be writable + no error output paths
			checkWarning: func(t *testing.T, warnings []string) {
				hasDirWarning := false
				for _, w := range warnings {
					if strings.Contains(w, "output path directory does not exist") {
						hasDirWarning = true
						break
					}
				}
				assert.True(t, hasDirWarning)
			},
		},
		{
			name: "invalid MaxSize",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"test.log"},
				ErrorOutputPaths: []string{"stderr"},
				MaxSize:          0, // 0 不会触发错误
			},
			wantErr:      false,
			wantWarnings: 0,
		},
		{
			name: "very large MaxSize with warning",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"test.log"},
				ErrorOutputPaths: []string{"stderr"},
				MaxSize:          2000,
			},
			wantErr:      false,
			wantWarnings: 1, // MaxSize is very large
			checkWarning: func(t *testing.T, warnings []string) {
				hasMaxSizeWarning := false
				for _, w := range warnings {
					if strings.Contains(w, "MaxSize is very large") {
						hasMaxSizeWarning = true
						break
					}
				}
				assert.True(t, hasMaxSizeWarning)
			},
		},
		{
			name: "negative MaxAge",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"test.log"},
				ErrorOutputPaths: []string{"stderr"},
				MaxAge:           -1,
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "MaxAge cannot be negative")
			},
		},
		{
			name: "invalid sampling config",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
				Sampling: &logger.SamplingConfig{
					Initial:    -1,
					Thereafter: -1,
				},
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "Sampling.Initial cannot be negative")
			},
		},
		{
			name: "invalid encoder config",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
				EncoderConfig: &logger.EncoderConfig{
					TimeEncoder:   "invalid-encoder",
					LevelEncoder:  "invalid-encoder",
					CallerEncoder: "invalid-encoder",
				},
			},
			wantErr: true,
			checkError: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "invalid TimeEncoder")
			},
		},
		{
			name: "reserved fields in InitialFields",
			opts: &logger.LogsOptions{
				Type:             logger.LoggerTypeZap,
				Level:            "info",
				Format:           "json",
				OutputPaths:      []string{"stdout"},
				ErrorOutputPaths: []string{"stderr"},
				InitialFields: map[string]interface{}{
					"level": "custom-level",
					"msg":   "custom-msg",
					"type":  "custom-type",
				},
			},
			wantErr:      false,
			wantWarnings: 3,
			checkWarning: func(t *testing.T, warnings []string) {
				reservedFieldCount := 0
				for _, w := range warnings {
					if strings.Contains(w, "reserved field") {
						reservedFieldCount++
					}
				}
				assert.Equal(t, 3, reservedFieldCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings, err := logger.ValidateConfig(tt.opts)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.checkError != nil && err != nil {
					tt.checkError(t, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.Len(t, warnings, tt.wantWarnings)
			if tt.checkWarning != nil && len(warnings) > 0 {
				tt.checkWarning(t, warnings)
			}
		})
	}
}

// TestConfigValidatorComprehensive 综合测试配置验证
func TestConfigValidatorComprehensive(t *testing.T) {
	tmpDir := t.TempDir()
	validLogFile := filepath.Join(tmpDir, "valid.log")
	
	// 创建一个有效的日志文件路径
	os.WriteFile(validLogFile, []byte(""), 0644)

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout", validLogFile},
		ErrorOutputPaths: []string{"stderr"},
		MaxSize:     100,
		MaxAge:      30,
		MaxBackups:  5,
		Compress:    true,
		Sampling: &logger.SamplingConfig{
			Initial:    100,
			Thereafter: 10,
		},
		EncoderConfig: &logger.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			MessageKey:    "msg",
			CallerKey:     "caller",
			TimeEncoder:   "iso8601",
			LevelEncoder:  "lowercase",
			CallerEncoder: "short",
			DurationUnit:  time.Millisecond,
		},
		InitialFields: map[string]interface{}{
			"service":     "test-service",
			"environment": "testing",
		},
	}

	warnings, err := logger.ValidateConfig(opts)
	assert.NoError(t, err)
	assert.Empty(t, warnings)
}

// TestValidatorWithActualLogger 测试验证器与实际 logger 创建的集成
func TestValidatorWithActualLogger(t *testing.T) {
	tests := []struct {
		name           string
		opts           *logger.LogsOptions
		shouldValidate bool
		shouldCreate   bool
	}{
		{
			name: "valid config passes both",
			opts: &logger.LogsOptions{
				Type:        logger.LoggerTypeZap,
				Level:       "info",
				Format:      "json",
				OutputPaths: []string{"stdout"},
			},
			shouldValidate: true,
			shouldCreate:   true,
		},
		{
			name: "invalid logger type fails both",
			opts: &logger.LogsOptions{
				Type: "invalid",
			},
			shouldValidate: false,
			shouldCreate:   false,
		},
		{
			name: "warnings don't prevent creation",
			opts: &logger.LogsOptions{
				Type:        logger.LoggerTypeZap,  // use explicit type
				Level:       "",  // will use default
				OutputPaths: []string{}, // will use default
			},
			shouldValidate: true,  // passes with warnings
			shouldCreate:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置
			warnings, err := logger.ValidateConfig(tt.opts)
			
			if tt.shouldValidate {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}

			// 尝试创建 logger
			l, err := logger.NewLogger(tt.opts)
			
			if tt.shouldCreate {
				assert.NoError(t, err)
				assert.NotNil(t, l)
			} else {
				assert.Error(t, err)
			}

			// 如果有警告，记录它们（实际使用中可能会显示给用户）
			if len(warnings) > 0 {
				t.Logf("Configuration warnings: %v", warnings)
			}
		})
	}
}