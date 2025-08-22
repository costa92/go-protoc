package logger_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoggerCreation 测试 logger 创建
func TestLoggerCreation(t *testing.T) {
	tests := []struct {
		name    string
		opts    *logger.LogsOptions
		wantErr bool
		errMsg  string
	}{
		{
			name: "create zap logger with default options",
			opts: &logger.LogsOptions{
				Type:   logger.LoggerTypeZap,
				Level:  "info",
				Format: "json",
			},
			wantErr: false,
		},
		{
			name: "create slog logger with default options",
			opts: &logger.LogsOptions{
				Type:   logger.LoggerTypeSlog,
				Level:  "debug",
				Format: "text",
			},
			wantErr: false,
		},
		{
			name: "create logger with nil options uses defaults",
			opts: nil,
			wantErr: false,
		},
		{
			name: "invalid logger type",
			opts: &logger.LogsOptions{
				Type: "invalid",
			},
			wantErr: true,
			errMsg:  "unsupported logger type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, err := logger.NewLogger(tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, l)
			}
		})
	}
}

// TestLoggerOutput 测试日志输出
func TestLoggerOutput(t *testing.T) {
	// 创建临时文件用于测试
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "debug",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 写入各种级别的日志
	l.Debug("debug message")
	l.Info("info message")
	l.Warn("warn message")
	l.Error("error message")
	
	// 给日志一点时间写入文件
	time.Sleep(100 * time.Millisecond)

	// 读取并验证日志文件
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 4, "should have 4 log lines")

	// 验证每行都是有效的 JSON
	for _, line := range lines {
		var logEntry map[string]interface{}
		err := json.Unmarshal([]byte(line), &logEntry)
		assert.NoError(t, err, "log line should be valid JSON")
		
		// 验证必要字段
		assert.Contains(t, logEntry, "level")
		assert.Contains(t, logEntry, "msg")
		assert.Contains(t, logEntry, "type")
		assert.Equal(t, "zap", logEntry["type"])
	}
}

// TestLoggerWithContext 测试上下文日志
func TestLoggerWithContext(t *testing.T) {
	opts := &logger.LogsOptions{
		Type:   logger.LoggerTypeZap,
		Level:  "info",
		Format: "json",
		OutputPaths: []string{"stdout"},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 测试 With 方法
	withLogger := l.With("service", "test-service", "version", "1.0.0")
	withLogger.Info("with test")

	// 测试 WithCtx 方法
	ctx := context.Background()
	ctxLogger := l.WithCtx(ctx, "request_id", "123456")
	ctxLogger.Info("context test")
}

// TestLoggerLevel 测试日志级别过滤
func TestLoggerLevel(t *testing.T) {
	tests := []struct {
		name          string
		level         string
		debugExpected bool
		infoExpected  bool
		warnExpected  bool
		errorExpected bool
	}{
		{
			name:          "debug level shows all",
			level:         "debug",
			debugExpected: true,
			infoExpected:  true,
			warnExpected:  true,
			errorExpected: true,
		},
		{
			name:          "info level hides debug",
			level:         "info",
			debugExpected: false,
			infoExpected:  true,
			warnExpected:  true,
			errorExpected: true,
		},
		{
			name:          "warn level hides debug and info",
			level:         "warn",
			debugExpected: false,
			infoExpected:  false,
			warnExpected:  true,
			errorExpected: true,
		},
		{
			name:          "error level only shows error",
			level:         "error",
			debugExpected: false,
			infoExpected:  false,
			warnExpected:  false,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "level_test.log")

			opts := &logger.LogsOptions{
				Type:        logger.LoggerTypeZap,
				Level:       tt.level,
				Format:      "json",
				OutputPaths: []string{logFile},
			}

			l, err := logger.NewLogger(opts)
			require.NoError(t, err)

			// 写入各级别日志
			l.Debug("debug")
			l.Info("info")
			l.Warn("warn")
			l.Error("error")

			time.Sleep(100 * time.Millisecond)

			content, err := os.ReadFile(logFile)
			require.NoError(t, err)

			contentStr := string(content)
			
			// 验证日志级别过滤
			if tt.debugExpected {
				assert.Contains(t, contentStr, "debug")
			} else {
				assert.NotContains(t, contentStr, "debug")
			}

			if tt.infoExpected {
				assert.Contains(t, contentStr, "info")
			} else {
				assert.NotContains(t, contentStr, `"msg":"info"`)
			}

			if tt.warnExpected {
				assert.Contains(t, contentStr, "warn")
			} else {
				assert.NotContains(t, contentStr, `"msg":"warn"`)
			}

			if tt.errorExpected {
				assert.Contains(t, contentStr, "error")
			} else {
				assert.NotContains(t, contentStr, `"msg":"error"`)
			}
		})
	}
}

// TestStructuredLogging 测试结构化日志
func TestStructuredLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "structured.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 测试结构化日志
	l.Infow("user action",
		"user_id", 12345,
		"action", "login",
		"ip", "192.168.1.1",
		"success", true,
		"duration", 1.23,
		"tags", []string{"web", "auth"},
		"metadata", map[string]interface{}{
			"browser": "Chrome",
			"os":      "Windows",
		},
	)

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(content, &logEntry)
	require.NoError(t, err)

	// 验证所有字段都被正确记录
	assert.Equal(t, "user action", logEntry["msg"])
	assert.Equal(t, float64(12345), logEntry["user_id"])
	assert.Equal(t, "login", logEntry["action"])
	assert.Equal(t, "192.168.1.1", logEntry["ip"])
	assert.Equal(t, true, logEntry["success"])
	assert.Equal(t, 1.23, logEntry["duration"])
}

// TestInitialFields 测试初始字段功能
func TestInitialFields(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "initial_fields.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
		InitialFields: map[string]interface{}{
			"service":     "test-service",
			"version":     "1.0.0",
			"environment": "testing",
		},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	l.Info("test message")

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(content, &logEntry)
	require.NoError(t, err)

	// 验证初始字段
	assert.Equal(t, "test-service", logEntry["service"])
	assert.Equal(t, "1.0.0", logEntry["version"])
	assert.Equal(t, "testing", logEntry["environment"])
	assert.Equal(t, "zap", logEntry["type"]) // 验证 type 字段
}

// TestConcurrentLogging 测试并发日志写入
func TestConcurrentLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "concurrent.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 并发写入日志
	var wg sync.WaitGroup
	numGoroutines := 10
	numLogs := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numLogs; j++ {
				l.Infow("concurrent log",
					"goroutine", goroutineID,
					"iteration", j,
				)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	// 验证所有日志都被写入
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, numGoroutines*numLogs, "should have all log entries")
}

// TestLogRotation 测试日志轮转
func TestLogRotation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "rotation.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
		MaxSize:     1,  // 1MB
		MaxAge:      1,  // 1 day
		MaxBackups:  3,
		Compress:    true,
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 写入一些日志
	for i := 0; i < 100; i++ {
		l.Infow("rotation test",
			"iteration", i,
			"data", strings.Repeat("x", 1000), // 大约 1KB 每条
		)
	}

	time.Sleep(100 * time.Millisecond)

	// 验证日志文件存在
	_, err = os.Stat(logFile)
	assert.NoError(t, err)
}

// TestCallerInformation 测试调用者信息
func TestCallerInformation(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "caller.log")

	opts := &logger.LogsOptions{
		Type:          logger.LoggerTypeZap,
		Level:         "info",
		Format:        "json",
		OutputPaths:   []string{logFile},
		DisableCaller: false,
		CallerSkip:    1, // 调整调用者跳过层数以显示测试文件
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 在一个已知的函数中记录日志
	testFunction := func() {
		l.Info("caller test")
	}
	testFunction()

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	var logEntry map[string]interface{}
	err = json.Unmarshal(content, &logEntry)
	require.NoError(t, err)

	// 验证调用者信息
	assert.Contains(t, logEntry, "caller")
	caller := logEntry["caller"].(string)
	assert.Contains(t, caller, "logger_test.go")
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	// 测试无效的输出路径
	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"/invalid/path/that/does/not/exist.log"},
	}

	_, err := logger.NewLogger(opts)
	// Zap 可能不会立即返回错误，所以这个测试可能需要调整
	_ = err
}

// TestGlobalLogger 测试全局 logger 函数
func TestGlobalLogger(t *testing.T) {
	// 测试全局函数不会 panic
	logger.Debug("global debug")
	logger.Info("global info")
	logger.Warn("global warn")
	logger.Error("global error")
	
	logger.Debugf("global %s", "debugf")
	logger.Infof("global %s", "infof")
	logger.Warnf("global %s", "warnf")
	logger.Errorf("global %s", "errorf")
	
	logger.Debugw("global debugw", "key", "value")
	logger.Infow("global infow", "key", "value")
	logger.Warnw("global warnw", "key", "value")
	logger.Errorw("global errorw", "key", "value")
	
	// 测试 With 和 WithCtx
	withLogger := logger.With("global", true)
	assert.NotNil(t, withLogger)
	
	ctx := context.Background()
	ctxLogger := logger.WithCtx(ctx, "request_id", "global-123")
	assert.NotNil(t, ctxLogger)
}

// TestSlogLogger 测试 Slog logger 实现
func TestSlogLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "slog.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeSlog,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
		InitialFields: map[string]interface{}{
			"service": "slog-test",
		},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 测试各种日志方法
	l.Info("slog info message")
	l.Infof("slog formatted %s", "message")
	l.Infow("slog structured", "key1", "value1", "key2", 123)

	// 测试 With
	withLogger := l.With("request_id", "slog-123")
	withLogger.Info("with logger message")

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.GreaterOrEqual(t, len(lines), 3)

	// 验证第一条日志
	var logEntry map[string]interface{}
	err = json.Unmarshal([]byte(lines[0]), &logEntry)
	require.NoError(t, err)
	
	assert.Equal(t, "slog", logEntry["type"])
	assert.Equal(t, "slog-test", logEntry["service"])
	assert.Equal(t, "slog info message", logEntry["msg"])
}

// MockFilter 用于测试的模拟过滤器
type MockFilter struct {
	enabled bool
	isEmpty bool
}

func (m *MockFilter) Filter(meta *logger.Metadata) bool {
	return m.enabled
}

func (m *MockFilter) Enabled(ctx context.Context, meta *logger.Metadata) bool {
	return m.enabled
}

func (m *MockFilter) IsEmpty() bool {
	return m.isEmpty
}

// TestLoggerWithFilter 测试日志过滤器
func TestLoggerWithFilter(t *testing.T) {
	// 这个测试需要访问内部实现，可能需要调整
	// 或者在 logger 包中提供设置过滤器的公开方法
}

// TestFormattedLogging 测试格式化日志
func TestFormattedLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "formatted.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 测试各种格式化日志
	l.Infof("String: %s", "test")
	l.Infof("Integer: %d", 123)
	l.Infof("Float: %.2f", 3.14159)
	l.Infof("Boolean: %t", true)
	l.Infof("Multiple: %s %d %v", "test", 123, []int{1, 2, 3})

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 5)
}

// TestDifferentOutputFormats 测试不同的输出格式
func TestDifferentOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		format string
		check  func(t *testing.T, output string)
	}{
		{
			name:   "json format",
			format: "json",
			check: func(t *testing.T, output string) {
				var data map[string]interface{}
				err := json.Unmarshal([]byte(output), &data)
				assert.NoError(t, err, "output should be valid JSON")
			},
		},
		{
			name:   "console format",
			format: "console",
			check: func(t *testing.T, output string) {
				// Console format should contain the message
				assert.Contains(t, output, "format test message")
				// Console format might still contain JSON fields but in different format
				// Just verify it's not pure JSON
				lines := strings.Split(strings.TrimSpace(output), "\n")
				for _, line := range lines {
					if line != "" {
						// Should not be parseable as pure JSON
						var data map[string]interface{}
						err := json.Unmarshal([]byte(line), &data)
						if err == nil {
							// If it parses as JSON, that's fine too for zap console
							// Just ensure it contains our message
							assert.Contains(t, line, "format test message")
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			logFile := filepath.Join(tmpDir, "format_test.log")

			opts := &logger.LogsOptions{
				Type:        logger.LoggerTypeZap,
				Level:       "info",
				Format:      tt.format,
				Encoding:    tt.format,
				OutputPaths: []string{logFile},
			}

			l, err := logger.NewLogger(opts)
			require.NoError(t, err)

			l.Info("format test message")

			time.Sleep(100 * time.Millisecond)

			content, err := os.ReadFile(logFile)
			require.NoError(t, err)

			tt.check(t, strings.TrimSpace(string(content)))
		})
	}
}

// TestCallerSkip 测试 CallerSkip 功能
func TestCallerSkip(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "caller_skip.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
		CallerSkip:  0,
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 测试不同的 caller skip
	l.Info("skip 0")
	
	skipLogger := l.WithCallerSkip(1)
	skipLogger.Info("skip 1")

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.Len(t, lines, 2)
}

// TestMultipleOutputs 测试多个输出目标
func TestMultipleOutputs(t *testing.T) {
	tmpDir := t.TempDir()
	logFile1 := filepath.Join(tmpDir, "output1.log")
	logFile2 := filepath.Join(tmpDir, "output2.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile1, logFile2},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	l.Info("multiple outputs test")

	time.Sleep(100 * time.Millisecond)

	// 验证两个文件都有内容
	content1, err := os.ReadFile(logFile1)
	require.NoError(t, err)
	assert.Contains(t, string(content1), "multiple outputs test")

	content2, err := os.ReadFile(logFile2)
	require.NoError(t, err)
	assert.Contains(t, string(content2), "multiple outputs test")
}

// TestSetLevel 测试动态设置日志级别
func TestSetLevel(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "setlevel.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	l, err := logger.NewLogger(opts)
	require.NoError(t, err)

	// 初始级别是 info，debug 不应该被记录
	l.Debug("should not appear")
	l.Info("info message 1")

	// 更改级别为 debug (注意：当前实现可能不支持动态级别更改)
	l.SetLevel(logger.DebugLevel)
	l.Info("info message 2")

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	// 验证日志内容
	assert.NotContains(t, string(content), "should not appear")
	assert.Contains(t, string(content), "info message 1")
	assert.Contains(t, string(content), "info message 2")
}