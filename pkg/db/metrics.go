package db

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// MetricType 指标类型
type MetricType int

const (
	MetricTypeDatabase MetricType = iota
	MetricTypeRedis
)

// PoolMetrics 优化后的连接池监控指标
type PoolMetrics struct {
	// 指标定义 - 使用更少的指标向量，通过labels区分
	poolConnections *prometheus.GaugeVec     // 通用连接数指标
	poolOperations  *prometheus.CounterVec   // 通用操作计数指标
	poolDuration    *prometheus.HistogramVec // 通用时长指标

	registry prometheus.Registerer
	enabled  atomic.Bool // 使用原子操作避免锁

	// 预分配的labels避免重复创建
	labelPools sync.Pool
}

// LabelSet 标签集合，使用对象池减少内存分配
type LabelSet struct {
	labels prometheus.Labels
	pool   *sync.Pool
}

// NewLabelSet 创建标签集合
func NewLabelSet(pool *sync.Pool) *LabelSet {
	return &LabelSet{
		labels: make(prometheus.Labels, 4), // 预分配常用容量
		pool:   pool,
	}
}

// Set 设置标签
func (ls *LabelSet) Set(key, value string) *LabelSet {
	ls.labels[key] = value
	return ls
}

// Get 获取标签集合
func (ls *LabelSet) Get() prometheus.Labels {
	return ls.labels
}

// Release 释放回对象池
func (ls *LabelSet) Release() {
	// 清空但保留容量
	for k := range ls.labels {
		delete(ls.labels, k)
	}
	if ls.pool != nil {
		ls.pool.Put(ls)
	}
}

// NewPoolMetrics 创建优化的连接池监控指标
func NewPoolMetrics(registry prometheus.Registerer) *PoolMetrics {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}

	pm := &PoolMetrics{
		registry: registry,
	}

	// 初始化标签对象池
	pm.labelPools = sync.Pool{
		New: func() interface{} {
			return NewLabelSet(&pm.labelPools)
		},
	}

	pm.initMetrics()
	pm.enabled.Store(false)

	return pm
}

// initMetrics 初始化指标 - 使用更简化的指标结构
func (pm *PoolMetrics) initMetrics() {
	// 统一的连接数指标，通过labels区分类型和状态
	pm.poolConnections = promauto.With(pm.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "connection_pool",
			Name:      "connections",
			Help:      "Number of connections in various states",
		},
		[]string{"service_name", "service_type", "state", "endpoint"},
	)

	// 统一的操作计数指标
	pm.poolOperations = promauto.With(pm.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "connection_pool",
			Name:      "operations_total",
			Help:      "Total number of pool operations",
		},
		[]string{"service_name", "service_type", "operation", "endpoint"},
	)

	// 统一的时长指标
	pm.poolDuration = promauto.With(pm.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "connection_pool",
			Name:      "duration_seconds",
			Help:      "Duration of pool operations in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10}, // 优化的bucket
		},
		[]string{"service_name", "service_type", "operation", "endpoint"},
	)
}

// Enable 启用指标收集
func (pm *PoolMetrics) Enable() {
	pm.enabled.Store(true)
}

// Disable 禁用指标收集
func (pm *PoolMetrics) Disable() {
	pm.enabled.Store(false)
}

// IsEnabled 检查是否启用 - 无锁操作
func (pm *PoolMetrics) IsEnabled() bool {
	return pm.enabled.Load()
}

// getLabelSet 获取标签集合 - 使用对象池
func (pm *PoolMetrics) getLabelSet() *LabelSet {
	return pm.labelPools.Get().(*LabelSet)
}

// CollectDatabaseStats 收集数据库统计信息 - 优化版本
func (pm *PoolMetrics) CollectDatabaseStats(name, dbType, endpoint string, db *sql.DB) {
	if !pm.IsEnabled() || db == nil {
		return
	}

	stats := db.Stats()

	// 使用对象池获取标签集合

	// 批量设置连接状态指标
	connectionStates := map[string]float64{
		"max_open": float64(stats.MaxOpenConnections),
		"open":     float64(stats.OpenConnections),
		"in_use":   float64(stats.InUse),
		"idle":     float64(stats.Idle),
	}

	for state, value := range connectionStates {
		stateLabels := pm.getLabelSet()
		stateLabels.Set("service_name", name).
			Set("service_type", dbType).
			Set("state", state).
			Set("endpoint", endpoint)

		pm.poolConnections.With(stateLabels.Get()).Set(value)
		stateLabels.Release()
	}

	// 批量设置操作计数指标
	operationCounts := map[string]float64{
		"wait":                 float64(stats.WaitCount),
		"max_idle_closed":      float64(stats.MaxIdleClosed),
		"max_idle_time_closed": float64(stats.MaxIdleTimeClosed),
		"max_lifetime_closed":  float64(stats.MaxLifetimeClosed),
	}

	for operation, count := range operationCounts {
		opLabels := pm.getLabelSet()
		opLabels.Set("service_name", name).
			Set("service_type", dbType).
			Set("operation", operation).
			Set("endpoint", endpoint)

		pm.poolOperations.With(opLabels.Get()).Add(count)
		opLabels.Release()
	}

	// 等待时长
	if stats.WaitDuration > 0 {
		waitLabels := pm.getLabelSet()
		waitLabels.Set("service_name", name).
			Set("service_type", dbType).
			Set("operation", "wait").
			Set("endpoint", endpoint)

		pm.poolDuration.With(waitLabels.Get()).Observe(stats.WaitDuration.Seconds())
		waitLabels.Release()
	}
}

// CollectRedisStats 收集Redis统计信息 - 优化版本
func (pm *PoolMetrics) CollectRedisStats(name, endpoint string, client *redis.Client) {
	if !pm.IsEnabled() || client == nil {
		return
	}

	stats := client.PoolStats()

	// 连接状态指标
	connectionStates := map[string]float64{
		"total":  float64(stats.TotalConns),
		"active": float64(stats.TotalConns - stats.IdleConns),
		"idle":   float64(stats.IdleConns),
		"stale":  float64(stats.StaleConns),
	}

	for state, value := range connectionStates {
		labels := pm.getLabelSet()
		labels.Set("service_name", name).
			Set("service_type", "redis").
			Set("state", state).
			Set("endpoint", endpoint)

		pm.poolConnections.With(labels.Get()).Set(value)
		labels.Release()
	}

	// 操作计数指标
	operationCounts := map[string]float64{
		"hits":     float64(stats.Hits),
		"misses":   float64(stats.Misses),
		"timeouts": float64(stats.Timeouts),
	}

	for operation, count := range operationCounts {
		labels := pm.getLabelSet()
		labels.Set("service_name", name).
			Set("service_type", "redis").
			Set("operation", operation).
			Set("endpoint", endpoint)

		pm.poolOperations.With(labels.Get()).Add(count)
		labels.Release()
	}
}

// OptimizedCollector 优化的收集器
type OptimizedCollector struct {
	metrics *PoolMetrics
	config  *MetricsConfig

	// 注册的连接信息
	connections sync.Map // map[string]*ConnectionInfo

	// 收集器状态
	running atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	Type     MetricType
	Name     string
	Endpoint string
	DB       *gorm.DB      // 用于数据库
	Redis    *redis.Client // 用于Redis
}

// NewOptimizedCollector 创建优化的收集器
func NewOptimizedCollector(config *MetricsConfig) *OptimizedCollector {
	if config == nil {
		config = NewMetricsConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &OptimizedCollector{
		metrics: NewPoolMetrics(config.Registry),
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// RegisterDatabase 注册数据库连接
func (oc *OptimizedCollector) RegisterDatabase(name, endpoint string, db *gorm.DB) {
	if db == nil {
		return
	}

	info := &ConnectionInfo{
		Type:     MetricTypeDatabase,
		Name:     name,
		Endpoint: endpoint,
		DB:       db,
	}

	oc.connections.Store(name, info)
}

// RegisterRedis 注册Redis连接
func (oc *OptimizedCollector) RegisterRedis(name, endpoint string, client *redis.Client) {
	if client == nil {
		return
	}

	info := &ConnectionInfo{
		Type:     MetricTypeRedis,
		Name:     name,
		Endpoint: endpoint,
		Redis:    client,
	}

	oc.connections.Store(name, info)
}

// Start 启动收集器
func (oc *OptimizedCollector) Start() error {
	if !oc.running.CompareAndSwap(false, true) {
		return nil // 已经运行
	}

	if !oc.config.IsAnyEnabled() {
		return nil // 没有启用监控
	}

	oc.metrics.Enable()

	oc.wg.Add(1)
	go oc.collectLoop()

	return nil
}

// Stop 停止收集器
func (oc *OptimizedCollector) Stop() error {
	if !oc.running.CompareAndSwap(true, false) {
		return nil // 未运行
	}

	oc.cancel()
	oc.wg.Wait()
	oc.metrics.Disable()

	return nil
}

// collectLoop 收集循环 - 优化版本
func (oc *OptimizedCollector) collectLoop() {
	defer oc.wg.Done()

	ticker := time.NewTicker(oc.config.CollectInterval)
	defer ticker.Stop()

	// 立即执行一次收集
	oc.collectOnce()

	for {
		select {
		case <-oc.ctx.Done():
			return
		case <-ticker.C:
			oc.collectOnce()
		}
	}
}

// collectOnce 执行一次收集 - 并发优化
func (oc *OptimizedCollector) collectOnce() {
	if !oc.metrics.IsEnabled() {
		return
	}

	// 使用并发收集提高性能
	var wg sync.WaitGroup

	oc.connections.Range(func(key, value interface{}) bool {
		info := value.(*ConnectionInfo)

		wg.Add(1)
		go func(info *ConnectionInfo) {
			defer wg.Done()
			oc.collectConnectionInfo(info)
		}(info)

		return true
	})

	wg.Wait()
}

// collectConnectionInfo 收集单个连接信息
func (oc *OptimizedCollector) collectConnectionInfo(info *ConnectionInfo) {
	switch info.Type {
	case MetricTypeDatabase:
		if info.DB != nil {
			if sqlDB, err := info.DB.DB(); err == nil {
				dbType := info.DB.Dialector.Name()
				oc.metrics.CollectDatabaseStats(info.Name, dbType, info.Endpoint, sqlDB)
			}
		}
	case MetricTypeRedis:
		if info.Redis != nil {
			oc.metrics.CollectRedisStats(info.Name, info.Endpoint, info.Redis)
		}
	}
}

// GetStatus 获取收集器状态
func (oc *OptimizedCollector) GetStatus() map[string]interface{} {
	var connectionCount int
	oc.connections.Range(func(_, _ interface{}) bool {
		connectionCount++
		return true
	})

	return map[string]interface{}{
		"running":          oc.running.Load(),
		"enabled":          oc.metrics.IsEnabled(),
		"connection_count": connectionCount,
		"collect_interval": oc.config.CollectInterval.String(),
		"enabled_services": oc.config.GetEnabledServices(),
	}
}

// 全局优化收集器实例
var (
	globalOptimizedCollector *OptimizedCollector
	globalCollectorOnce      sync.Once
)

// GetGlobalOptimizedCollector 获取全局优化收集器
func GetGlobalOptimizedCollector() *OptimizedCollector {
	globalCollectorOnce.Do(func() {
		config := NewMetricsConfig()
		globalOptimizedCollector = NewOptimizedCollector(config)
	})
	return globalOptimizedCollector
}

// InitializeGlobalMetrics 初始化全局监控
func InitializeGlobalMetrics(config *MetricsConfig) error {
	if err := config.Validate(); err != nil {
		return err
	}

	collector := NewOptimizedCollector(config)
	globalOptimizedCollector = collector

	return collector.Start()
}

// =========================
// 简化的监控包装器
// =========================

// ConnectionPoolMetrics 连接池监控指标
type ConnectionPoolMetrics struct {
	poolMetrics *PoolMetrics
	registry    prometheus.Registerer
	enabled     bool
}

// NewConnectionPoolMetrics 创建连接池监控指标
func NewConnectionPoolMetrics(registry prometheus.Registerer) *ConnectionPoolMetrics {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}

	return &ConnectionPoolMetrics{
		poolMetrics: NewPoolMetrics(registry),
		registry:    registry,
		enabled:     false,
	}
}

// Enable 启用指标收集
func (m *ConnectionPoolMetrics) Enable() {
	m.enabled = true
	m.poolMetrics.Enable()
}

// Disable 禁用指标收集
func (m *ConnectionPoolMetrics) Disable() {
	m.enabled = false
	m.poolMetrics.Disable()
}

// IsEnabled 检查指标收集是否启用
func (m *ConnectionPoolMetrics) IsEnabled() bool {
	return m.enabled
}

// MonitoredDB 带监控的数据库连接
type MonitoredDB struct {
	*gorm.DB
	metrics      *ConnectionPoolMetrics
	databaseName string
}

// NewMonitoredDB 创建带监控的数据库连接
func NewMonitoredDB(db *gorm.DB, metrics *ConnectionPoolMetrics, databaseName string) *MonitoredDB {
	return &MonitoredDB{
		DB:           db,
		metrics:      metrics,
		databaseName: databaseName,
	}
}

// CollectMetrics 收集当前数据库的指标
func (mdb *MonitoredDB) CollectMetrics() error {
	if !mdb.metrics.IsEnabled() {
		return nil
	}

	sqlDB, err := mdb.DB.DB()
	if err != nil {
		return err
	}

	dbType := mdb.DB.Dialector.Name()
	endpoint := mdb.databaseName
	mdb.metrics.poolMetrics.CollectDatabaseStats(mdb.databaseName, dbType, endpoint, sqlDB)
	return nil
}

// MonitoredRedis 带监控的Redis客户端
type MonitoredRedis struct {
	*redis.Client
	metrics   *ConnectionPoolMetrics
	redisName string
}

// NewMonitoredRedis 创建带监控的Redis客户端
func NewMonitoredRedis(client *redis.Client, metrics *ConnectionPoolMetrics, redisName string) *MonitoredRedis {
	return &MonitoredRedis{
		Client:    client,
		metrics:   metrics,
		redisName: redisName,
	}
}

// CollectMetrics 收集当前Redis的指标
func (mr *MonitoredRedis) CollectMetrics() {
	if !mr.metrics.IsEnabled() {
		return
	}

	endpoint := mr.Client.Options().Addr
	mr.metrics.poolMetrics.CollectRedisStats(mr.redisName, endpoint, mr.Client)
}

// =========================
// 全局实例
// =========================

var (
	globalMetrics    *ConnectionPoolMetrics
	globalOnce       sync.Once
)

// GetGlobalMetrics 获取全局指标实例
func GetGlobalMetrics() *ConnectionPoolMetrics {
	globalOnce.Do(func() {
		globalMetrics = NewConnectionPoolMetrics(prometheus.DefaultRegisterer)
	})
	return globalMetrics
}

// SetGlobalMetrics 设置全局指标实例
func SetGlobalMetrics(metrics *ConnectionPoolMetrics) {
	globalMetrics = metrics
}
