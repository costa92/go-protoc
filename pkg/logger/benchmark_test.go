package logger_test

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/logger"
)

// 基准测试辅助函数
func createBenchmarkLogger(b *testing.B, loggerType logger.LoggerType, output io.Writer) logger.Logger {
	opts := &logger.LogsOptions{
		Type:              loggerType,
		Level:             "info",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		DisableCaller:     true, // 禁用 caller 以获得更纯粹的性能数据
		DisableStacktrace: true,
		Development:       false,
	}

	l, err := logger.NewLogger(opts)
	if err != nil {
		b.Fatal(err)
	}
	return l
}

// BenchmarkZapLogger 测试 Zap logger 性能
func BenchmarkZapLogger(b *testing.B) {
	l := createBenchmarkLogger(b, logger.LoggerTypeZap, io.Discard)

	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Info("benchmark message")
		}
	})

	b.Run("Infof", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Infof("benchmark message %d", i)
		}
	})

	b.Run("Infow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Infow("benchmark message",
				"iteration", i,
				"timestamp", time.Now().Unix(),
				"status", "ok",
			)
		}
	})

	b.Run("With", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = l.With("key", "value")
		}
	})

	b.Run("WithMultipleFields", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = l.With(
				"field1", "value1",
				"field2", "value2",
				"field3", "value3",
				"field4", "value4",
				"field5", "value5",
			)
		}
	})
}

// BenchmarkSlogLogger 测试 Slog logger 性能
func BenchmarkSlogLogger(b *testing.B) {
	l := createBenchmarkLogger(b, logger.LoggerTypeSlog, io.Discard)

	b.Run("Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Info("benchmark message")
		}
	})

	b.Run("Infof", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Infof("benchmark message %d", i)
		}
	})

	b.Run("Infow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Infow("benchmark message",
				"iteration", i,
				"timestamp", time.Now().Unix(),
				"status", "ok",
			)
		}
	})
}

// BenchmarkLoggerComparison 比较不同 logger 的性能
func BenchmarkLoggerComparison(b *testing.B) {
	zapLogger := createBenchmarkLogger(b, logger.LoggerTypeZap, io.Discard)
	slogLogger := createBenchmarkLogger(b, logger.LoggerTypeSlog, io.Discard)

	b.Run("Zap-StructuredLogging", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			zapLogger.Infow("structured log",
				"user_id", 12345,
				"action", "benchmark",
				"duration", 1.23,
				"success", true,
			)
		}
	})

	b.Run("Slog-StructuredLogging", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			slogLogger.Infow("structured log",
				"user_id", 12345,
				"action", "benchmark",
				"duration", 1.23,
				"success", true,
			)
		}
	})
}

// BenchmarkConcurrentLogging 测试并发日志性能
func BenchmarkConcurrentLogging(b *testing.B) {
	l := createBenchmarkLogger(b, logger.LoggerTypeZap, io.Discard)

	b.Run("Concurrent-1", func(b *testing.B) {
		b.SetParallelism(1)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("concurrent benchmark")
			}
		})
	})

	b.Run("Concurrent-10", func(b *testing.B) {
		b.SetParallelism(10)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("concurrent benchmark")
			}
		})
	})

	b.Run("Concurrent-100", func(b *testing.B) {
		b.SetParallelism(100)
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				l.Info("concurrent benchmark")
			}
		})
	})
}

// BenchmarkLoggerWithContext 测试带上下文的日志性能
func BenchmarkLoggerWithContext(b *testing.B) {
	l := createBenchmarkLogger(b, logger.LoggerTypeZap, io.Discard)
	ctx := context.Background()

	b.Run("WithCtx", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctxLogger := l.WithCtx(ctx, "request_id", fmt.Sprintf("req-%d", i))
			ctxLogger.Info("context benchmark")
		}
	})

	b.Run("WithCtx-Reuse", func(b *testing.B) {
		ctxLogger := l.WithCtx(ctx, "request_id", "req-123")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctxLogger.Info("context benchmark")
		}
	})
}

// BenchmarkLoggerCreation 测试 logger 创建性能
func BenchmarkLoggerCreation(b *testing.B) {
	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}

	b.Run("NewLogger-Zap", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := logger.NewLogger(opts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	slogOpts := &logger.LogsOptions{
		Type:        logger.LoggerTypeSlog,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{"stdout"},
	}

	b.Run("NewLogger-Slog", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := logger.NewLogger(slogOpts)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkLoggerWithFile 测试文件输出性能
func BenchmarkLoggerWithFile(b *testing.B) {
	tmpDir := b.TempDir()
	logFile := filepath.Join(tmpDir, "benchmark.log")

	opts := &logger.LogsOptions{
		Type:        logger.LoggerTypeZap,
		Level:       "info",
		Format:      "json",
		OutputPaths: []string{logFile},
	}

	l, err := logger.NewLogger(opts)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("FileOutput", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Infow("file benchmark",
				"iteration", i,
				"data", "some test data",
			)
		}
	})

	// 清理
	b.StopTimer()
	os.Remove(logFile)
}

// BenchmarkDynamicLogger 测试动态 logger 性能
func BenchmarkDynamicLogger(b *testing.B) {
	opts := &logger.LogsOptions{
		Type:          logger.LoggerTypeZap,
		Level:         "info",
		Format:        "json",
		OutputPaths:   []string{"stdout"},
		DisableCaller: true,
	}

	dl, err := logger.NewDynamicLogger(opts)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("DynamicLogger-Info", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dl.Info("dynamic logger benchmark")
		}
	})

	b.Run("DynamicLogger-UpdateLevel", func(b *testing.B) {
		levels := []string{"debug", "info", "warn", "error"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			level := levels[i%len(levels)]
			err := dl.UpdateLevel(level)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkGlobalLogger 测试全局 logger 函数性能
func BenchmarkGlobalLogger(b *testing.B) {
	// 设置一个优化的默认 logger
	opts := &logger.LogsOptions{
		Type:          logger.LoggerTypeZap,
		Level:         "info",
		Format:        "json",
		OutputPaths:   []string{"stdout"},
		DisableCaller: true,
	}

	l, _ := logger.NewLogger(opts)
	logger.SetDefaultLogger(l)

	b.Run("GlobalInfo", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("global benchmark")
		}
	})

	b.Run("GlobalInfow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Infow("global structured",
				"iteration", i,
				"benchmark", true,
			)
		}
	})
}

// BenchmarkMemoryAllocation 测试内存分配
func BenchmarkMemoryAllocation(b *testing.B) {
	l := createBenchmarkLogger(b, logger.LoggerTypeZap, io.Discard)

	b.Run("SimpleLog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Info("memory benchmark")
		}
	})

	b.Run("StructuredLog", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			l.Infow("memory benchmark",
				"field1", "value1",
				"field2", 123,
				"field3", true,
				"field4", 3.14,
			)
		}
	})

	b.Run("WithFields", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			withLogger := l.With("request_id", i)
			withLogger.Info("with fields benchmark")
		}
	})
}
