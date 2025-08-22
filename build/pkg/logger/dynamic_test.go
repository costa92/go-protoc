package logger_test

import (
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

// TestDynamicLogger 测试动态配置的 logger
func TestDynamicLogger(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "dynamic.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	dl, err := logger.NewDynamicLogger(opts)
	require.NoError(t, err)

	// 测试初始配置
	dl.Info("initial info message")
	dl.Debug("initial debug message") // 不应该出现

	time.Sleep(100 * time.Millisecond)

	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "initial info message")
	assert.NotContains(t, string(content), "initial debug message")

	// 动态更新级别
	err = dl.UpdateLevel("debug")
	require.NoError(t, err)

	dl.Debug("debug after update")
	dl.Info("info after update")

	time.Sleep(100 * time.Millisecond)

	content, err = os.ReadFile(logFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "debug after update")
	assert.Contains(t, string(content), "info after update")
}

// TestDynamicLoggerConcurrentUpdate 测试并发更新配置
func TestDynamicLoggerConcurrentUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "dynamic_concurrent.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	dl, err := logger.NewDynamicLogger(opts)
	require.NoError(t, err)

	// 并发更新配置和写日志
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			
			// 交替更新级别
			if id%2 == 0 {
				dl.UpdateLevel("debug")
			} else {
				dl.UpdateLevel("info")
			}
			
			// 写一些日志
			for j := 0; j < 10; j++ {
				dl.Infow("concurrent log", "goroutine", id, "iteration", j)
			}
		}(i)
	}

	wg.Wait()
	time.Sleep(200 * time.Millisecond)

	// 验证没有 panic 并且日志被正确写入
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)
	
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	assert.NotEmpty(t, lines)
}

// TestDynamicLoggerFullConfigUpdate 测试完整配置更新
func TestDynamicLoggerFullConfigUpdate(t *testing.T) {
	tmpDir := t.TempDir()
	logFile1 := filepath.Join(tmpDir, "dynamic1.log")
	logFile2 := filepath.Join(tmpDir, "dynamic2.log")

	opts1 := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile1},
		InitialFields: map[string]interface{}{
			"version": "1.0",
		},
	}

	dl, err := logger.NewDynamicLogger(opts1)
	require.NoError(t, err)

	// 写初始日志
	dl.Info("message to file1")

	// 更新到新配置
	opts2 := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "debug",
		Format:      "json",
		OutputPaths: []string{logFile2},
		InitialFields: map[string]interface{}{
			"version": "2.0",
		},
	}

	err = dl.UpdateConfig(opts2)
	require.NoError(t, err)

	// 写新日志
	dl.Info("message to file2")
	dl.Debug("debug to file2")

	time.Sleep(100 * time.Millisecond)

	// 验证文件1
	content1, err := os.ReadFile(logFile1)
	require.NoError(t, err)
	assert.Contains(t, string(content1), "message to file1")
	assert.Contains(t, string(content1), `"version":"1.0"`)

	// 验证文件2
	content2, err := os.ReadFile(logFile2)
	require.NoError(t, err)
	assert.Contains(t, string(content2), "message to file2")
	assert.Contains(t, string(content2), "debug to file2")
	assert.Contains(t, string(content2), `"version":"2.0"`)
}

// TestDynamicLoggerGetOptions 测试获取当前配置
func TestDynamicLoggerGetOptions(t *testing.T) {
	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
		InitialFields: map[string]interface{}{
			"service": "test",
		},
	}

	dl, err := logger.NewDynamicLogger(opts)
	require.NoError(t, err)

	// 获取配置
	currentOpts := dl.GetOptions()
	assert.Equal(t, logger.LoggerTypeZap, currentOpts.Type)
	assert.Equal(t, "info", currentOpts.Level)
	assert.Equal(t, "json", currentOpts.Format)
	assert.Equal(t, "test", currentOpts.InitialFields["service"])

	// 更新级别
	err = dl.UpdateLevel("debug")
	require.NoError(t, err)

	// 再次获取配置
	currentOpts = dl.GetOptions()
	assert.Equal(t, "debug", currentOpts.Level)
}

// TestDynamicLoggerWithMethods 测试 With 相关方法
func TestDynamicLoggerWithMethods(t *testing.T) {
	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}

	dl, err := logger.NewDynamicLogger(opts)
	require.NoError(t, err)

	// 测试 With 方法
	withLogger := dl.With("key1", "value1")
	assert.NotNil(t, withLogger)
	
	// 测试 WithCtx 方法
	ctxLogger := dl.WithCtx(nil, "key2", "value2")
	assert.NotNil(t, ctxLogger)

	// 测试 WithCallerSkip 方法
	skipLogger := dl.WithCallerSkip(1)
	assert.NotNil(t, skipLogger)
}