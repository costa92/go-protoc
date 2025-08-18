package logger

import (
	"testing"
	"time"
)

// TestDynamicParameterFalse 测试 Dynamic=false 时返回静态 logger
func TestDynamicParameterFalse(t *testing.T) {
	tests := []struct {
		name   string
		opts   *LogsOptions
		verify func(t *testing.T, logger Logger)
	}{
		{
			name: "Static Zap Logger",
			opts: &LogsOptions{
				Type:        LoggerTypeZap,
				Dynamic:     false,
				Level:       "info",
				Format:      "json",
				OutputPaths: []string{"stdout"},
			},
			verify: func(t *testing.T, logger Logger) {
				// 静态 Zap logger 应该返回 ZapLogger 类型
				if _, ok := logger.(*ZapLogger); !ok {
					t.Errorf("Expected ZapLogger when Dynamic=false, got %T", logger)
				}
			},
		},
		{
			name: "Static Slog Logger",
			opts: &LogsOptions{
				Type:        LoggerTypeSlog,
				Dynamic:     false,
				Level:       "info",
				Format:      "json",
				OutputPaths: []string{"stdout"},
			},
			verify: func(t *testing.T, logger Logger) {
				// 静态 Slog logger 应该返回 SlogLogger 类型
				if _, ok := logger.(*SlogLogger); !ok {
					t.Errorf("Expected SlogLogger when Dynamic=false, got %T", logger)
				}
			},
		},
		{
			name: "Default Options (Dynamic=false)",
			opts: DefaultOptions(),
			verify: func(t *testing.T, logger Logger) {
				// 默认配置应该返回静态 ZapLogger
				if _, ok := logger.(*ZapLogger); !ok {
					t.Errorf("Expected ZapLogger with default options, got %T", logger)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.opts)
			if err != nil {
				t.Fatalf("Failed to create logger: %v", err)
			}
			
			tt.verify(t, logger)
			
			// 验证logger基本功能
			logger.Info("Test message")
		})
	}
}

// TestDynamicParameterTrue 测试 Dynamic=true 时返回 DynamicLogger
func TestDynamicParameterTrue(t *testing.T) {
	tests := []struct {
		name string
		opts *LogsOptions
	}{
		{
			name: "Dynamic Zap Logger",
			opts: &LogsOptions{
				Type:        LoggerTypeZap,
				Dynamic:     true,
				Level:       "info",
				Format:      "json",
				OutputPaths: []string{"stdout"},
			},
		},
		{
			name: "Dynamic Slog Logger",
			opts: &LogsOptions{
				Type:        LoggerTypeSlog,
				Dynamic:     true,
				Level:       "info",
				Format:      "json",
				OutputPaths: []string{"stdout"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.opts)
			if err != nil {
				t.Fatalf("Failed to create dynamic logger: %v", err)
			}
			
			// 验证返回的是 DynamicLogger
			dynamicLogger, ok := logger.(*DynamicLogger)
			if !ok {
				t.Errorf("Expected DynamicLogger when Dynamic=true, got %T", logger)
			}
			
			// 验证基本日志功能
			logger.Info("Test dynamic logger")
			
			// 验证可以获取配置
			options := dynamicLogger.GetOptions()
			if options.Type != tt.opts.Type {
				t.Errorf("Expected type %s, got %s", tt.opts.Type, options.Type)
			}
			if !options.Dynamic {
				t.Error("Expected Dynamic=true in options")
			}
		})
	}
}

// TestDynamicLoggerRuntimeUpdate 测试 DynamicLogger 的运行时更新功能
func TestDynamicLoggerRuntimeUpdate(t *testing.T) {
	// 创建动态 logger
	opts := &LogsOptions{
		Type:        LoggerTypeZap,
		Dynamic:     true,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}
	
	logger, err := NewLogger(opts)
	if err != nil {
		t.Fatalf("Failed to create dynamic logger: %v", err)
	}
	
	dynamicLogger, ok := logger.(*DynamicLogger)
	if !ok {
		t.Fatal("Expected DynamicLogger")
	}
	
	// 测试更新级别
	err = dynamicLogger.UpdateLevel("debug")
	if err != nil {
		t.Fatalf("Failed to update level: %v", err)
	}
	
	// 验证级别已更新
	options := dynamicLogger.GetOptions()
	if options.Level != "debug" {
		t.Errorf("Expected level debug, got %s", options.Level)
	}
	
	// 测试日志输出（debug 现在应该可见）
	dynamicLogger.Debug("This debug message should be visible")
	
	// 测试完整配置更新
	newOpts := &LogsOptions{
		Type:        LoggerTypeSlog, // 切换到 Slog
		Level:       "warn",
		Format:      "text",
		OutputPaths: []string{"stdout"},
	}
	
	err = dynamicLogger.UpdateConfig(newOpts)
	if err != nil {
		t.Fatalf("Failed to update config: %v", err)
	}
	
	// 验证配置已更新
	options = dynamicLogger.GetOptions()
	if options.Type != LoggerTypeSlog {
		t.Errorf("Expected type slog, got %s", options.Type)
	}
	if options.Level != "warn" {
		t.Errorf("Expected level warn, got %s", options.Level)
	}
	if options.Format != "text" {
		t.Errorf("Expected format text, got %s", options.Format)
	}
	
	// 测试日志输出
	dynamicLogger.Warn("This warning message should be visible")
	dynamicLogger.Info("This info message should NOT be visible (below warn level)")
}

// TestDynamicLoggerConcurrency 测试 DynamicLogger 的并发安全性
func TestDynamicLoggerConcurrency(t *testing.T) {
	opts := &LogsOptions{
		Type:        LoggerTypeZap,
		Dynamic:     true,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}
	
	logger, err := NewLogger(opts)
	if err != nil {
		t.Fatalf("Failed to create dynamic logger: %v", err)
	}
	
	dynamicLogger, ok := logger.(*DynamicLogger)
	if !ok {
		t.Fatal("Expected DynamicLogger")
	}
	
	// 并发读写测试
	done := make(chan bool)
	
	// 启动多个 goroutine 进行日志记录
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				dynamicLogger.Infof("Concurrent log from goroutine %d, iteration %d", id, j)
			}
			done <- true
		}(i)
	}
	
	// 启动 goroutine 进行配置更新
	go func() {
		for i := 0; i < 10; i++ {
			levels := []string{"debug", "info", "warn", "error"}
			level := levels[i%len(levels)]
			dynamicLogger.UpdateLevel(level)
			time.Sleep(time.Millisecond * 10)
		}
		done <- true
	}()
	
	// 等待所有 goroutine 完成
	for i := 0; i < 11; i++ {
		select {
		case <-done:
			// OK
		case <-time.After(time.Second * 5):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

// TestDynamicLoggerInterface 测试 DynamicLogger 实现完整的 Logger 接口
func TestDynamicLoggerInterface(t *testing.T) {
	opts := &LogsOptions{
		Type:        LoggerTypeZap,
		Dynamic:     true,
		Level:       "debug",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}
	
	logger, err := NewLogger(opts)
	if err != nil {
		t.Fatalf("Failed to create dynamic logger: %v", err)
	}
	
	// 验证实现了 Logger 接口
	var _ Logger = logger
	
	// 测试所有日志级别方法
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warn message")
	logger.Error("Error message")
	
	// 测试格式化方法
	logger.Debugf("Debug message: %s", "formatted")
	logger.Infof("Info message: %d", 123)
	logger.Warnf("Warn message: %v", true)
	logger.Errorf("Error message: %f", 3.14)
	
	// 测试结构化方法
	logger.Debugw("Debug message", "key", "value")
	logger.Infow("Info message", "count", 42)
	logger.Warnw("Warn message", "active", true)
	logger.Errorw("Error message", "rate", 1.5)
	
	// 测试上下文方法
	childLogger := logger.With("service", "test")
	childLogger.Info("Message with context")
	
	skipLogger := logger.WithCallerSkip(1)
	skipLogger.Info("Message with caller skip")
}

// TestDynamicParameterValidation 测试 Dynamic 参数的配置验证
func TestDynamicParameterValidation(t *testing.T) {
	validator := NewConfigValidator()
	
	// 测试 Dynamic=true 的有效配置
	opts := &LogsOptions{
		Type:             LoggerTypeZap,
		Dynamic:          true,
		Level:            "info",
		Format:           "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	
	err := validator.Validate(opts)
	if err != nil {
		t.Errorf("Expected no error for valid dynamic config, got: %v", err)
	}
	
	// Dynamic 参数本身不应该产生警告或错误
	warnings := validator.GetWarnings()
	for _, warning := range warnings {
		if containsString(warning, "dynamic") || containsString(warning, "Dynamic") {
			t.Errorf("Unexpected warning about Dynamic parameter: %s", warning)
		}
	}
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
				containsSubstring(s[1:len(s)-1], substr))))
}

func containsSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}