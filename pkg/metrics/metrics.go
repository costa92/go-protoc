package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Package metrics 提供统一的指标收集和管理功能
// 支持 HTTP, gRPC, 数据库, Redis 和自定义指标

// Quick Start Example:
//
//	// 1. 创建配置
//	config := metrics.NewConfig(
//		metrics.WithEnabled(true),
//		metrics.WithServiceName("my-service"),
//		metrics.WithHTTP(true, 8080),
//		metrics.WithGRPC(true),
//		metrics.WithDatabase(true, metrics.DatabaseInfo{
//			Name: "main", Type: "mysql", Endpoint: "localhost:3306",
//		}),
//	)
//
//	// 2. 初始化全局管理器
//	err := metrics.InitializeGlobalManager(config)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 3. 获取指标接口
//	manager := metrics.GetGlobalManager()
//	httpMetrics, _ := manager.GetHTTPMetrics()
//	dbMetrics, _ := manager.GetDatabaseMetrics()
//
//	// 4. 记录指标
//	httpMetrics.RecordRequest("GET", "/api/users", 200, time.Millisecond*150)
//	dbMetrics.RecordQuery("main", "SELECT", "users", time.Millisecond*50, true)

// StartMetrics 快速启动指标收集
// 这是一个便捷函数，用于快速设置和启动指标收集
func StartMetrics(serviceName string, options ...ConfigOption) (*Manager, error) {
	// 创建默认配置
	config := NewConfig()

	// 应用默认选项
	config.ApplyOptions(
		WithEnabled(true),
		WithServiceName(serviceName),
		WithHTTP(true, 8080),
	)

	// 应用用户选项
	config.ApplyOptions(options...)

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 创建并启动管理器
	manager := NewManager(config)
	if err := manager.Start(); err != nil {
		return nil, err
	}

	return manager, nil
}

// StopMetrics 停止指标收集
func StopMetrics(manager *Manager) error {
	if manager != nil {
		return manager.Stop()
	}
	return nil
}

// GetMetrics 获取指标收集器
func GetMetrics() *Manager {
	return GetGlobalManager()
}

// RecordHTTPRequest 记录HTTP请求 - 便捷函数
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration) {
	manager := GetGlobalManager()
	if httpMetrics, err := manager.GetHTTPMetrics(); err == nil {
		httpMetrics.RecordRequest(method, path, statusCode, duration)
	}
}

// RecordGRPCCall 记录gRPC调用 - 便捷函数
func RecordGRPCCall(service, method, code string, duration time.Duration) {
	manager := GetGlobalManager()
	if grpcMetrics, err := manager.GetGRPCMetrics(); err == nil {
		grpcMetrics.RecordCall(service, method, code, duration)
	}
}

// RecordDatabaseQuery 记录数据库查询 - 便捷函数
func RecordDatabaseQuery(database, operation, table string, duration time.Duration, success bool) {
	manager := GetGlobalManager()
	if dbMetrics, err := manager.GetDatabaseMetrics(); err == nil {
		dbMetrics.RecordQuery(database, operation, table, duration, success)
	}
}

// RecordRedisCommand 记录Redis命令 - 便捷函数
func RecordRedisCommand(instance, command string, duration time.Duration, success bool) {
	manager := GetGlobalManager()
	if redisMetrics, err := manager.GetRedisMetrics(); err == nil {
		redisMetrics.RecordCommand(instance, command, duration, success)
	}
}

// RegisterDatabase 注册数据库连接 - 便捷函数
func RegisterDatabase(name string, db *gorm.DB) error {
	manager := GetGlobalManager()
	return manager.RegisterTarget(MetricTypeDatabase, name, db)
}

// RegisterRedis 注册Redis连接 - 便捷函数
func RegisterRedis(name string, client *redis.Client) error {
	manager := GetGlobalManager()
	return manager.RegisterTarget(MetricTypeRedis, name, client)
}

// GetCustomMetrics 获取自定义指标 - 便捷函数
func GetCustomMetrics() (CustomMetrics, error) {
	manager := GetGlobalManager()
	return manager.GetCustomMetrics()
}

// CreateCounter 创建计数器 - 便捷函数
func CreateCounter(name, help string, labels []string) (*prometheus.CounterVec, error) {
	customMetrics, err := GetCustomMetrics()
	if err != nil {
		return nil, err
	}
	return customMetrics.CreateCounter(name, help, labels)
}

// CreateGauge 创建仪表盘 - 便捷函数
func CreateGauge(name, help string, labels []string) (*prometheus.GaugeVec, error) {
	customMetrics, err := GetCustomMetrics()
	if err != nil {
		return nil, err
	}
	return customMetrics.CreateGauge(name, help, labels)
}

// CreateHistogram 创建直方图 - 便捷函数
func CreateHistogram(name, help string, labels []string, buckets []float64) (*prometheus.HistogramVec, error) {
	customMetrics, err := GetCustomMetrics()
	if err != nil {
		return nil, err
	}
	return customMetrics.CreateHistogram(name, help, labels, buckets)
}

// GetStatus 获取整体状态 - 便捷函数
func GetStatus() map[string]interface{} {
	manager := GetGlobalManager()
	return manager.GetStatus()
}

// IsHealthy 检查系统健康状态 - 便捷函数
func IsHealthy() bool {
	healthChecker := GetGlobalHealthChecker()
	return healthChecker.IsHealthy()
}

// SetupMiddleware 设置中间件 - 便捷函数
// 返回可用于不同框架的中间件
func SetupMiddleware() (HTTPMiddleware, GRPCInterceptor, error) {
	manager := GetGlobalManager()

	// 获取HTTP收集器
	httpCollector, exists := manager.GetCollector(MetricTypeHTTP)
	if !exists {
		return nil, nil, ErrCollectorNotFound
	}

	// 获取gRPC收集器
	grpcCollector, exists := manager.GetCollector(MetricTypeGRPC)
	if !exists {
		return nil, nil, ErrCollectorNotFound
	}

	// 创建中间件
	httpMiddleware := NewHTTPMiddleware(httpCollector.(*HTTPCollector))
	grpcInterceptor := NewGRPCInterceptor(grpcCollector.(*GRPCCollector))

	return httpMiddleware, grpcInterceptor, nil
}

// Shutdown 优雅关闭指标收集
func Shutdown(ctx context.Context) error {
	manager := GetGlobalManager()
	if !manager.IsRunning() {
		return nil
	}

	// 创建一个带超时的上下文
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// 在另一个goroutine中执行停止操作
	done := make(chan error, 1)
	go func() {
		done <- manager.Stop()
	}()

	// 等待停止完成或超时
	select {
	case err := <-done:
		return err
	case <-shutdownCtx.Done():
		return shutdownCtx.Err()
	}
}

// ============ 以下是一些实用的预设配置 ============

// WebServiceConfig 创建Web服务的标准配置
func WebServiceConfig(serviceName string, port int) *Config {
	config := NewConfig()
	config.ApplyOptions(
		WithEnabled(true),
		WithServiceName(serviceName),
		WithHTTP(true, port),
		WithGRPC(true),
		WithCustom(true),
		WithHealthCheck(true, 30*time.Second, 5*time.Second),
	)
	return config
}

// DatabaseServiceConfig 创建数据库服务的标准配置
func DatabaseServiceConfig(serviceName string, databases ...DatabaseInfo) *Config {
	config := NewConfig()
	config.ApplyOptions(
		WithEnabled(true),
		WithServiceName(serviceName),
		WithDatabase(true, databases...),
		WithCustom(true),
		WithHealthCheck(true, 30*time.Second, 5*time.Second),
	)
	return config
}

// MicroserviceConfig 创建微服务的标准配置
func MicroserviceConfig(serviceName string, port int, databases []DatabaseInfo, redisInstances []RedisInfo) *Config {
	config := NewConfig()
	config.ApplyOptions(
		WithEnabled(true),
		WithServiceName(serviceName),
		WithHTTP(true, port),
		WithGRPC(true),
		WithCustom(true),
		WithHealthCheck(true, 30*time.Second, 5*time.Second),
	)

	if len(databases) > 0 {
		config.ApplyOptions(WithDatabase(true, databases...))
	}

	if len(redisInstances) > 0 {
		config.ApplyOptions(WithRedis(true, redisInstances...))
	}

	return config
}

// MinimalConfig 创建最小化配置
func MinimalConfig(serviceName string) *Config {
	config := NewConfig()
	config.ApplyOptions(
		WithEnabled(true),
		WithServiceName(serviceName),
		WithCustom(true),
	)
	return config
}
