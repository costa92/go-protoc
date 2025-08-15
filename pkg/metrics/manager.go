package metrics

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// Manager 统一指标管理器
type Manager struct {
	config   *Config
	registry Registry
	enabled  atomic.Bool
	running  atomic.Bool

	// 收集器
	collectors map[MetricType]Collector
	mu         sync.RWMutex

	// 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 健康检查
	healthChecker HealthChecker
}

// NewManager 创建指标管理器
func NewManager(config *Config, opts ...ManagerOption) *Manager {
	if config == nil {
		config = NewConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		config:     config,
		registry:   NewDefaultRegistry(config.Registry),
		collectors: make(map[MetricType]Collector),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 应用选项
	for _, opt := range opts {
		opt(manager)
	}

	// 初始化收集器
	manager.initCollectors()

	// 初始化健康检查
	if config.HealthCheck.Enabled {
		manager.healthChecker = NewHealthChecker(config.HealthCheck)
	}

	return manager
}

// ManagerOption 管理器选项
type ManagerOption func(*Manager)

// WithRegistryOption 设置注册器
func WithRegistryOption(registry Registry) ManagerOption {
	return func(m *Manager) {
		m.registry = registry
	}
}

// WithHealthChecker 设置健康检查器
func WithHealthChecker(checker HealthChecker) ManagerOption {
	return func(m *Manager) {
		m.healthChecker = checker
	}
}

// initCollectors 初始化收集器
func (m *Manager) initCollectors() {
	// HTTP收集器
	if m.config.HTTP.Enabled {
		m.collectors[MetricTypeHTTP] = NewHTTPCollector(m.config)
	}

	// gRPC收集器
	if m.config.GRPC.Enabled {
		m.collectors[MetricTypeGRPC] = NewGRPCCollector(m.config)
	}

	// 数据库收集器
	if m.config.Database.Enabled {
		m.collectors[MetricTypeDatabase] = NewDatabaseCollector(m.config)
	}

	// Redis收集器
	if m.config.Redis.Enabled {
		m.collectors[MetricTypeRedis] = NewRedisCollector(m.config)
	}

	// 自定义收集器
	if m.config.Custom.Enabled {
		m.collectors[MetricTypeCustom] = NewCustomCollector(m.config)
	}
}

// Start 启动管理器
func (m *Manager) Start() error {
	if !m.running.CompareAndSwap(false, true) {
		return nil // 已经运行
	}

	if !m.config.IsAnyEnabled() {
		m.running.Store(false)
		return nil
	}

	m.enabled.Store(true)

	// 启动所有收集器
	for _, collector := range m.collectors {
		if err := collector.Start(); err != nil {
			m.Stop()
			return fmt.Errorf("failed to start collector %s: %w", collector.GetType(), err)
		}
	}

	// 启动健康检查
	if m.healthChecker != nil {
		m.wg.Add(1)
		go m.runHealthCheck()
	}

	return nil
}

// Stop 停止管理器
func (m *Manager) Stop() error {
	if !m.running.CompareAndSwap(true, false) {
		return nil // 未运行
	}

	m.enabled.Store(false)

	// 停止所有收集器
	for _, collector := range m.collectors {
		collector.Stop()
	}

	// 停止健康检查
	m.cancel()
	m.wg.Wait()

	return nil
}

// IsRunning 检查是否运行
func (m *Manager) IsRunning() bool {
	return m.running.Load()
}

// GetCollector 获取指定类型的收集器
func (m *Manager) GetCollector(metricType MetricType) (Collector, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	collector, exists := m.collectors[metricType]
	return collector, exists
}

// GetHTTPMetrics 获取HTTP指标
func (m *Manager) GetHTTPMetrics() (HTTPMetrics, error) {
	collector, exists := m.GetCollector(MetricTypeHTTP)
	if !exists {
		return nil, ErrCollectorNotFound
	}

	httpCollector, ok := collector.(*HTTPCollector)
	if !ok {
		return nil, ErrUnsupportedCollector
	}

	return httpCollector.GetHTTPMetrics(), nil
}

// GetGRPCMetrics 获取gRPC指标
func (m *Manager) GetGRPCMetrics() (GRPCMetrics, error) {
	collector, exists := m.GetCollector(MetricTypeGRPC)
	if !exists {
		return nil, ErrCollectorNotFound
	}

	grpcCollector, ok := collector.(*GRPCCollector)
	if !ok {
		return nil, ErrUnsupportedCollector
	}

	return grpcCollector.GetGRPCMetrics(), nil
}

// GetDatabaseMetrics 获取数据库指标
func (m *Manager) GetDatabaseMetrics() (DatabaseMetrics, error) {
	collector, exists := m.GetCollector(MetricTypeDatabase)
	if !exists {
		return nil, ErrCollectorNotFound
	}

	dbCollector, ok := collector.(*DatabaseCollector)
	if !ok {
		return nil, ErrUnsupportedCollector
	}

	return dbCollector.GetDatabaseMetrics(), nil
}

// GetRedisMetrics 获取Redis指标
func (m *Manager) GetRedisMetrics() (RedisMetrics, error) {
	collector, exists := m.GetCollector(MetricTypeRedis)
	if !exists {
		return nil, ErrCollectorNotFound
	}

	redisCollector, ok := collector.(*RedisCollector)
	if !ok {
		return nil, ErrUnsupportedCollector
	}

	return redisCollector.GetRedisMetrics(), nil
}

// GetCustomMetrics 获取自定义指标
func (m *Manager) GetCustomMetrics() (CustomMetrics, error) {
	collector, exists := m.GetCollector(MetricTypeCustom)
	if !exists {
		return nil, ErrCollectorNotFound
	}

	customCollector, ok := collector.(*CustomCollector)
	if !ok {
		return nil, ErrUnsupportedCollector
	}

	return customCollector.GetCustomMetrics(), nil
}

// RegisterTarget 注册监控目标
func (m *Manager) RegisterTarget(metricType MetricType, name string, target interface{}) error {
	collector, exists := m.GetCollector(metricType)
	if !exists {
		return ErrCollectorNotFound
	}

	return collector.RegisterTarget(name, target)
}

// UnregisterTarget 取消注册监控目标
func (m *Manager) UnregisterTarget(metricType MetricType, name string) error {
	collector, exists := m.GetCollector(metricType)
	if !exists {
		return ErrCollectorNotFound
	}

	return collector.UnregisterTarget(name)
}

// GetStatus 获取管理器状态
func (m *Manager) GetStatus() map[string]interface{} {
	collectorsStatus := make(map[string]interface{})

	m.mu.RLock()
	for metricType, collector := range m.collectors {
		collectorsStatus[string(metricType)] = collector.GetStatus()
	}
	m.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":    m.enabled.Load(),
		"running":    m.running.Load(),
		"config":     m.config,
		"collectors": collectorsStatus,
	}

	if m.healthChecker != nil {
		status["health"] = m.healthChecker.GetStatus()
	}

	return status
}

// runHealthCheck 运行健康检查
func (m *Manager) runHealthCheck() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.HealthCheck.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck 执行健康检查
func (m *Manager) performHealthCheck() {
	if m.healthChecker == nil {
		return
	}

	ctx, cancel := context.WithTimeout(m.ctx, m.config.HealthCheck.Timeout)
	defer cancel()

	// 检查所有收集器
	for metricType, collector := range m.collectors {
		if collector.IsRunning() {
			m.healthChecker.RecordHealthy(string(metricType))
		} else {
			m.healthChecker.RecordUnhealthy(string(metricType), "collector not running")
		}
	}

	// 执行健康检查
	m.healthChecker.Check(ctx)
}

// DefaultRegistry 默认注册器实现
type DefaultRegistry struct {
	registerer prometheus.Registerer
	gatherer   prometheus.Gatherer
}

// NewDefaultRegistry 创建默认注册器
func NewDefaultRegistry(registerer prometheus.Registerer) *DefaultRegistry {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	var gatherer prometheus.Gatherer
	if g, ok := registerer.(prometheus.Gatherer); ok {
		gatherer = g
	} else {
		gatherer = prometheus.DefaultGatherer
	}

	return &DefaultRegistry{
		registerer: registerer,
		gatherer:   gatherer,
	}
}

// Register 注册指标
func (r *DefaultRegistry) Register(collector prometheus.Collector) error {
	return r.registerer.Register(collector)
}

// Unregister 取消注册指标
func (r *DefaultRegistry) Unregister(collector prometheus.Collector) bool {
	return r.registerer.Unregister(collector)
}

// Gather 收集指标
func (r *DefaultRegistry) Gather() ([]*dto.MetricFamily, error) {
	return r.gatherer.Gather()
}

// MustRegister 必须注册指标
func (r *DefaultRegistry) MustRegister(collectors ...prometheus.Collector) {
	r.registerer.MustRegister(collectors...)
}

// GetRegisterer 获取注册器
func (r *DefaultRegistry) GetRegisterer() prometheus.Registerer {
	return r.registerer
}

// GetGatherer 获取收集器
func (r *DefaultRegistry) GetGatherer() prometheus.Gatherer {
	return r.gatherer
}

// 全局管理器实例
var (
	globalManager     *Manager
	globalManagerOnce sync.Once
)

// GetGlobalManager 获取全局管理器
func GetGlobalManager() *Manager {
	globalManagerOnce.Do(func() {
		config := NewConfig()
		globalManager = NewManager(config)
	})
	return globalManager
}

// SetGlobalManager 设置全局管理器
func SetGlobalManager(manager *Manager) {
	globalManager = manager
}

// InitializeGlobalManager 初始化全局管理器
func InitializeGlobalManager(config *Config, opts ...ManagerOption) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	manager := NewManager(config, opts...)
	globalManager = manager

	return manager.Start()
}
