package db

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// InitializationManager 初始化管理器
type InitializationManager struct {
	config    *MetricsConfig
	collector *Collector
	factory   *ConnectionFactory

	initialized bool
	mu          sync.RWMutex

	// 依赖注入
	logger Logger
}

// InitOption 初始化选项
type InitOption func(*InitializationManager)

// WithLogger 设置日志记录器
func WithLogger(logger Logger) InitOption {
	return func(im *InitializationManager) {
		im.logger = logger
	}
}

// NewInitializationManager 创建初始化管理器
func NewInitializationManager(config *MetricsConfig, opts ...InitOption) *InitializationManager {
	if config == nil {
		config = NewMetricsConfig()
	}

	im := &InitializationManager{
		config: config,
		logger: &defaultLogger{}, // 默认日志记录器
	}

	// 应用选项
	for _, opt := range opts {
		opt(im)
	}

	return im
}

// Initialize 初始化监控系统
func (im *InitializationManager) Initialize(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if im.initialized {
		return nil // 已经初始化
	}

	// 验证配置
	if err := im.config.Validate(); err != nil {
		return fmt.Errorf("invalid metrics config: %w", err)
	}

	im.logger.Info("Initializing connection pool monitoring system")

	// 创建组件
	if err := im.createComponents(); err != nil {
		return fmt.Errorf("failed to create components: %w", err)
	}

	// 启动收集器
	if im.config.IsAnyEnabled() {
		if err := im.collector.Start(); err != nil {
			return fmt.Errorf("failed to start collector: %w", err)
		}
		im.logger.Info("Connection pool metrics collector started",
			"interval", im.config.CollectInterval,
			"services", im.config.GetEnabledServices())
	} else {
		im.logger.Info("Connection pool metrics collection is disabled")
	}

	// 设置全局实例
	im.setGlobalInstances()

	im.initialized = true
	im.logger.Info("Connection pool monitoring system initialized successfully")

	return nil
}

// createComponents 创建组件
func (im *InitializationManager) createComponents() error {
	// 创建收集器
	im.collector = NewCollector(im.config)

	// 创建连接工厂
	im.factory = NewConnectionFactory(im.config)

	return nil
}

// setGlobalInstances 设置全局实例
func (im *InitializationManager) setGlobalInstances() {
	SetGlobalConnectionFactory(im.factory)

	// 设置全局收集器
	if globalOptimizedCollector == nil {
		config := im.config.Clone()
		globalOptimizedCollector = NewOptimizedCollector(config)
	}
}

// Shutdown 关闭监控系统
func (im *InitializationManager) Shutdown(ctx context.Context) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if !im.initialized {
		return nil // 未初始化
	}

	im.logger.Info("Shutting down connection pool monitoring system")

	// 停止收集器
	if im.collector != nil {
		if err := im.collector.Stop(); err != nil {
			im.logger.Error("Failed to stop collector", "error", err)
		}
	}

	im.initialized = false
	im.logger.Info("Connection pool monitoring system shutdown completed")

	return nil
}

// GetCollector 获取收集器
func (im *InitializationManager) GetCollector() *Collector {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.collector
}

// GetFactory 获取连接工厂
func (im *InitializationManager) GetFactory() *ConnectionFactory {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.factory
}

// IsInitialized 检查是否已初始化
func (im *InitializationManager) IsInitialized() bool {
	im.mu.RLock()
	defer im.mu.RUnlock()
	return im.initialized
}

// GetStatus 获取状态
func (im *InitializationManager) GetStatus() map[string]interface{} {
	im.mu.RLock()
	defer im.mu.RUnlock()

	status := map[string]interface{}{
		"initialized": im.initialized,
		"config":      im.config,
	}

	if im.collector != nil {
		status["collector"] = im.collector.GetDiagnostics()
	}

	return status
}

// 默认日志记录器实现
type defaultLogger struct{}

func (dl *defaultLogger) Debug(args ...interface{}) {
	log.Println("[DEBUG]", args)
}

func (dl *defaultLogger) Info(args ...interface{}) {
	log.Println("[INFO]", args)
}

func (dl *defaultLogger) Warn(args ...interface{}) {
	log.Println("[WARN]", args)
}

func (dl *defaultLogger) Error(args ...interface{}) {
	log.Println("[ERROR]", args)
}

func (dl *defaultLogger) With(key string, value interface{}) Logger {
	// 简单实现，实际可以使用更复杂的日志库
	return dl
}

// InitializationBootstrap 初始化引导器
type InitializationBootstrap struct {
	manager *InitializationManager
	timeout time.Duration
}

// NewInitializationBootstrap 创建初始化引导器
func NewInitializationBootstrap(config *MetricsConfig, timeout time.Duration) *InitializationBootstrap {
	if timeout <= 0 {
		timeout = 30 * time.Second // 默认30秒超时
	}

	return &InitializationBootstrap{
		manager: NewInitializationManager(config),
		timeout: timeout,
	}
}

// Bootstrap 引导初始化
func (ib *InitializationBootstrap) Bootstrap() error {
	ctx, cancel := context.WithTimeout(context.Background(), ib.timeout)
	defer cancel()

	return ib.manager.Initialize(ctx)
}

// Shutdown 关闭
func (ib *InitializationBootstrap) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), ib.timeout)
	defer cancel()

	return ib.manager.Shutdown(ctx)
}

// GetManager 获取管理器
func (ib *InitializationBootstrap) GetManager() *InitializationManager {
	return ib.manager
}

// 全局初始化管理器
var (
	globalInitManager *InitializationManager
	initOnce          sync.Once
)

// GetGlobalInitManager 获取全局初始化管理器
func GetGlobalInitManager() *InitializationManager {
	initOnce.Do(func() {
		config := NewMetricsConfig()
		globalInitManager = NewInitializationManager(config)
	})
	return globalInitManager
}

// InitializeGlobalMonitoring 初始化全局监控
func InitializeGlobalMonitoring(config *MetricsConfig) error {
	manager := NewInitializationManager(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := manager.Initialize(ctx); err != nil {
		return err
	}

	globalInitManager = manager
	return nil
}

// ShutdownGlobalMonitoring 关闭全局监控
func ShutdownGlobalMonitoring() error {
	if globalInitManager != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		return globalInitManager.Shutdown(ctx)
	}
	return nil
}

// QuickStart 快速启动监控
func QuickStart() error {
	config := NewMetricsConfig()

	// 启用基本监控
	config.Enabled = true
	config.Database.Enabled = true
	config.Database.Name = "default"

	return InitializeGlobalMonitoring(config)
}

// QuickStartWithOptions 使用选项快速启动
func QuickStartWithOptions(opts ...MetricsOption) error {
	config := NewMetricsConfig()
	config.ApplyOptions(opts...)

	return InitializeGlobalMonitoring(config)
}