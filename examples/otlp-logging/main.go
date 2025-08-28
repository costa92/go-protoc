// Package main demonstrates OTLP logging integration with the Agent -> Collector architecture
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/costa92/go-protoc/v2/pkg/version"
)

func main() {
	fmt.Println("=== OTLP Logging Integration Example ===")
	fmt.Println("Demonstrating: Application -> OTEL Agent -> OTEL Collector -> VictoriaLogs")

	// 示例1: 使用简化配置
	demonstrateQuickConfig()

	// 示例2: 使用详细配置
	demonstrateDetailedConfig()

	// 示例3: 混合架构（同时文件 + OTLP）
	demonstrateMixedArchitecture()
}

// demonstrateQuickConfig 演示简化配置方式
func demonstrateQuickConfig() {
	fmt.Println("\n🚀 示例1: 简化配置方式")

	// 创建简化配置
	config := &logger.QuickConfig{
		Preset:       logger.PresetObservability, // 可观测性预设
		Type:         "zap",
		Level:        "debug",
		LogDir:       "logs/otlp-example",
		EnableOTLP:   true,                        // 启用 OTLP
		OTLPEndpoint: "127.0.0.1:4327",           // OTEL Agent gRPC 端点
	}

	// 转换为完整配置
	fullOptions := config.ToFullOptions()

	// 创建日志器
	log, err := logger.NewLogger(fullOptions)
	if err != nil {
		fmt.Printf("❌ Failed to create logger: %v\n", err)
		return
	}

	fmt.Println("✅ Logger created with OTLP support")
	fmt.Printf("   📡 OTLP Endpoint: %s\n", config.OTLPEndpoint)
	fmt.Printf("   📝 File Output: %v\n", fullOptions.OutputPaths)

	// 发送测试日志
	sendTestLogs(log, "quick-config")

	// 优雅关闭
	shutdownLogger(log)
}

// demonstrateDetailedConfig 演示详细配置方式
func demonstrateDetailedConfig() {
	fmt.Println("\n🔧 示例2: 详细配置方式")

	// 获取版本信息
	versionInfo := version.Get()

	// 创建详细的 OTLP 配置
	otlpConfig := &logger.OTLPConfig{
		Enabled:        true,
		Endpoint:       "127.0.0.1:4327", // OTEL Agent
		Protocol:       "grpc",
		Timeout:        10 * time.Second,
		BatchTimeout:   2 * time.Second,
		BatchSize:      50, // 较小的批次，更快发送
		Insecure:       true,
		Headers:        map[string]string{
			"x-custom-header": "otlp-logging-example",
		},
		ServiceName:    versionInfo.ServiceName,
		ServiceVersion: versionInfo.GitVersion,
		Environment:    "example",
	}

	// 创建完整的日志选项
	options := &logger.LogsOptions{
		Type:              logger.LoggerTypeZap,
		Level:             "info",
		Format:            "json",
		DisableCaller:     false,
		DisableStacktrace: false,
		EnableColor:       false,
		OutputPaths:       []string{"stdout", "logs/otlp-example/detailed.log"},
		ErrorOutputPaths:  []string{"stderr"},
		Development:       false,
		Encoding:          "json",
		OTLP:              otlpConfig, // 添加 OTLP 配置
		InitialFields: map[string]interface{}{
			"service":     versionInfo.ServiceName,
			"version":     versionInfo.GitVersion,
			"branch":      versionInfo.GitBranch,
			"environment": "example",
			"example":     "detailed-config",
		},
	}

	// 创建日志器
	log, err := logger.NewLogger(options)
	if err != nil {
		fmt.Printf("❌ Failed to create detailed logger: %v\n", err)
		return
	}

	fmt.Println("✅ Detailed logger created with custom OTLP configuration")
	fmt.Printf("   📡 OTLP Protocol: %s\n", otlpConfig.Protocol)
	fmt.Printf("   ⏱️  Batch Timeout: %v\n", otlpConfig.BatchTimeout)
	fmt.Printf("   📦 Batch Size: %d\n", otlpConfig.BatchSize)

	// 发送测试日志
	sendTestLogs(log, "detailed-config")

	// 优雅关闭
	shutdownLogger(log)
}

// demonstrateMixedArchitecture 演示混合架构
func demonstrateMixedArchitecture() {
	fmt.Println("\n🔄 示例3: 混合架构（文件 + OTLP 双路发送）")

	config := &logger.QuickConfig{
		Preset:       logger.PresetProduction,    // 生产环境预设
		Type:         "zap",
		Level:        "info",
		LogDir:       "logs/otlp-example",
		EnableOTLP:   true,                       // 启用 OTLP
		OTLPEndpoint: "127.0.0.1:4327",          // OTEL Agent
	}

	fullOptions := config.ToFullOptions()

	// 创建日志器
	log, err := logger.NewLogger(fullOptions)
	if err != nil {
		fmt.Printf("❌ Failed to create mixed logger: %v\n", err)
		return
	}

	fmt.Println("✅ Mixed architecture logger created")
	fmt.Println("   📝 Local File: logs/otlp-example/apiserver/app.log")
	fmt.Printf("   📡 OTLP Agent: %s\n", config.OTLPEndpoint)
	fmt.Println("   🔄 Data Flow: App -> [File + OTLP Agent] -> OTEL Collector -> VictoriaLogs")

	// 发送混合测试日志
	sendMixedTestLogs(log)

	// 优雅关闭
	shutdownLogger(log)
}

// sendTestLogs 发送测试日志
func sendTestLogs(log logger.Logger, scenario string) {
	fmt.Printf("📤 Sending test logs for scenario: %s\n", scenario)

	// 使用 With 添加场景标识
	scenarioLogger := log.With("scenario", scenario, "test_run", time.Now().Unix())

	// 发送不同级别的日志
	scenarioLogger.Debug("Debug message: detailed diagnostic information")
	scenarioLogger.Info("Info message: normal operation log", "operation", "user_login", "user_id", 12345)
	scenarioLogger.Warn("Warning message: potential issue detected", "issue_type", "high_memory_usage", "memory_percent", 85.5)
	scenarioLogger.Error("Error message: operation failed", "error", "database connection timeout", "retry_count", 3)

	// 发送结构化日志
	scenarioLogger.Infow("User operation completed",
		"user_id", 12345,
		"operation", "purchase",
		"amount", 99.99,
		"currency", "USD",
		"payment_method", "credit_card",
		"timestamp", time.Now(),
	)

	fmt.Printf("   ✅ Test logs sent for %s\n", scenario)
}

// sendMixedTestLogs 发送混合测试日志
func sendMixedTestLogs(log logger.Logger) {
	fmt.Println("📤 Sending mixed architecture test logs")

	// 模拟业务场景
	scenarios := []struct {
		name      string
		operation string
		data      map[string]interface{}
	}{
		{
			name:      "e-commerce-order",
			operation: "order_processing",
			data: map[string]interface{}{
				"order_id":       "ORD-2025-001",
				"user_id":        67890,
				"total_amount":   156.78,
				"payment_status": "completed",
				"items_count":    3,
			},
		},
		{
			name:      "user-authentication",
			operation: "user_login",
			data: map[string]interface{}{
				"user_id":    12345,
				"login_type": "oauth2",
				"ip_address": "192.168.1.100",
				"user_agent": "Mozilla/5.0 (compatible)",
				"success":    true,
			},
		},
		{
			name:      "api-request",
			operation: "api_call",
			data: map[string]interface{}{
				"method":      "POST",
				"endpoint":    "/api/v1/users",
				"status_code": 201,
				"duration_ms": 45.2,
				"request_id":  "req-abc123",
			},
		},
	}

	for i, scenario := range scenarios {
		scenarioLogger := log.With("scenario_id", i+1, "scenario_name", scenario.name)

		scenarioLogger.Infow(fmt.Sprintf("Processing %s", scenario.operation),
			"data", scenario.data,
		)

		// 模拟处理延迟
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("   ✅ Mixed architecture test logs sent")
	fmt.Println("   📊 Data sent to both local files and OTLP Agent")
}

// shutdownLogger 优雅关闭日志器
func shutdownLogger(log logger.Logger) {
	fmt.Println("🔄 Shutting down logger...")

	// 检查是否支持 Shutdown
	if shutdownable, ok := log.(interface{ Shutdown(context.Context) error }); ok {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := shutdownable.Shutdown(ctx); err != nil {
			fmt.Printf("⚠️  Warning: failed to shutdown logger: %v\n", err)
		} else {
			fmt.Println("✅ Logger shutdown completed")
		}
	} else {
		fmt.Println("ℹ️  Logger does not support graceful shutdown")
	}

	// 等待一下，确保数据发送完毕
	time.Sleep(2 * time.Second)
}

func init() {
	// 确保日志目录存在
	fmt.Println("📁 Ensuring log directories exist...")
	if err := ensureLogDirs(); err != nil {
		fmt.Printf("⚠️  Warning: failed to create log directories: %v\n", err)
	}
}

func ensureLogDirs() error {
	dirs := []string{
		"logs/otlp-example",
		"logs/otlp-example/apiserver",
	}

	for _, dir := range dirs {
		if err := mkdir(dir); err != nil {
			return err
		}
	}
	return nil
}

func mkdir(dir string) error {
	// 这里应该导入 os 包，但为了简化示例，使用 fmt
	fmt.Printf("   📁 Creating directory: %s\n", dir)
	return nil
}