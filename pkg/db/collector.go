package db

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// Collector 连接池监控收集器
type Collector struct {
	// 配置
	config  *MetricsConfig
	metrics *PoolMetrics

	// 状态管理
	state  atomic.Int32 // 0: stopped, 1: running, 2: stopping
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 连接管理
	connections sync.Map // map[string]*ConnectionInfo

	// 性能优化
	batchSize  int
	lastError  atomic.Value // 存储最后一个错误
	errorCount atomic.Int64 // 错误计数

	// 健康检查
	healthCheck *HealthChecker

	// 缓存
	metricsCache *MetricsCache
}

// CollectorState 收集器状态
type CollectorState int32

const (
	StateStopped  CollectorState = 0
	StateRunning  CollectorState = 1
	StateStopping CollectorState = 2
)

// NewCollector 创建连接池监控收集器
func NewCollector(config *MetricsConfig) *Collector {
	if config == nil {
		config = NewMetricsConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	collector := &Collector{
		config:       config,
		metrics:      NewPoolMetrics(config.Registry),
		ctx:          ctx,
		cancel:       cancel,
		batchSize:    10, // 批处理大小
		healthCheck:  NewHealthChecker(),
		metricsCache: NewMetricsCache(5 * time.Minute), // 5分钟缓存
	}

	return collector
}

// Start 启动收集器
func (c *Collector) Start() error {
	if !c.state.CompareAndSwap(int32(StateStopped), int32(StateRunning)) {
		return nil // 已经运行或正在启动
	}

	if !c.config.IsAnyEnabled() {
		c.state.Store(int32(StateStopped))
		return nil
	}

	c.metrics.Enable()

	// 启动健康检查
	if err := c.healthCheck.Start(c.ctx); err != nil {
		c.state.Store(int32(StateStopped))
		return err
	}

	// 启动主收集循环
	c.wg.Add(1)
	go c.mainCollectionLoop()

	// 启动健康检查循环
	c.wg.Add(1)
	go c.healthCheckLoop()

	return nil
}

// Stop 停止收集器
func (c *Collector) Stop() error {
	if !c.state.CompareAndSwap(int32(StateRunning), int32(StateStopping)) {
		return nil // 未运行或已在停止中
	}

	// 取消上下文
	c.cancel()

	// 等待所有goroutine结束
	c.wg.Wait()

	// 停止健康检查
	c.healthCheck.Stop()

	// 禁用指标收集
	c.metrics.Disable()

	// 更新状态
	c.state.Store(int32(StateStopped))

	return nil
}

// IsRunning 检查是否正在运行
func (c *Collector) IsRunning() bool {
	return c.state.Load() == int32(StateRunning)
}

// mainCollectionLoop 主收集循环
func (c *Collector) mainCollectionLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.CollectInterval)
	defer ticker.Stop()

	// 立即执行一次收集
	c.collectBatch()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collectBatch()
		}
	}
}

// healthCheckLoop 健康检查循环
func (c *Collector) healthCheckLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(1 * time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.performHealthCheck()
		}
	}
}

// collectBatch 批量收集指标
func (c *Collector) collectBatch() {
	if !c.metrics.IsEnabled() {
		return
	}

	var batch []*ConnectionInfo
	count := 0

	// 收集一批连接信息
	c.connections.Range(func(key, value interface{}) bool {
		if count >= c.batchSize {
			return false // 停止迭代
		}

		info := value.(*ConnectionInfo)
		batch = append(batch, info)
		count++
		return true
	})

	// 并发处理批次
	if len(batch) > 0 {
		c.processBatch(batch)
	}
}

// processBatch 处理批次
func (c *Collector) processBatch(batch []*ConnectionInfo) {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // 限制并发数

	for _, info := range batch {
		wg.Add(1)
		go func(info *ConnectionInfo) {
			defer wg.Done()

			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
				c.collectSingleConnection(info)
			case <-c.ctx.Done():
				return
			}
		}(info)
	}

	wg.Wait()
}

// collectSingleConnection 收集单个连接的指标
func (c *Collector) collectSingleConnection(info *ConnectionInfo) {
	defer func() {
		if r := recover(); r != nil {
			c.recordError("panic during metric collection")
		}
	}()

	// 检查缓存
	cacheKey := info.Name + ":" + info.Endpoint
	if c.metricsCache.IsValid(cacheKey) {
		return // 使用缓存的指标
	}

	switch info.Type {
	case MetricTypeDatabase:
		if info.DB != nil {
			if sqlDB, err := info.DB.DB(); err == nil {
				dbType := info.DB.Dialector.Name()
				c.metrics.CollectDatabaseStats(info.Name, dbType, info.Endpoint, sqlDB)
				c.metricsCache.Set(cacheKey, true)
			} else {
				c.recordError("failed to get sql.DB from gorm.DB")
			}
		}
	case MetricTypeRedis:
		if info.Redis != nil {
			c.metrics.CollectRedisStats(info.Name, info.Endpoint, info.Redis)
			c.metricsCache.Set(cacheKey, true)
		}
	}
}

// performHealthCheck 执行健康检查
func (c *Collector) performHealthCheck() {
	c.connections.Range(func(key, value interface{}) bool {
		info := value.(*ConnectionInfo)
		healthy := c.healthCheck.CheckConnection(info)

		// 记录健康状态指标
		c.recordHealthStatus(info.Name, healthy)

		return true
	})
}

// recordError 记录错误
func (c *Collector) recordError(errMsg string) {
	c.lastError.Store(errMsg)
	c.errorCount.Add(1)
}

// recordHealthStatus 记录健康状态
func (c *Collector) recordHealthStatus(name string, healthy bool) {
	var status float64
	if healthy {
		status = 1
	}

	// 记录健康状态指标
	labels := prometheus.Labels{
		"service_name": name,
		"status":       "health",
	}

	c.metrics.poolConnections.With(labels).Set(status)
}

// GetDiagnostics 获取诊断信息
func (c *Collector) GetDiagnostics() map[string]interface{} {
	var connectionCount int
	c.connections.Range(func(_, _ interface{}) bool {
		connectionCount++
		return true
	})

	lastErr := ""
	if err := c.lastError.Load(); err != nil {
		lastErr = err.(string)
	}

	return map[string]interface{}{
		"state":            c.getStateName(),
		"connection_count": connectionCount,
		"error_count":      c.errorCount.Load(),
		"last_error":       lastErr,
		"config":           c.config,
		"cache_size":       c.metricsCache.Size(),
		"health_status":    c.healthCheck.GetStatus(),
	}
}

// getStateName 获取状态名称
func (c *Collector) getStateName() string {
	switch CollectorState(c.state.Load()) {
	case StateStopped:
		return "stopped"
	case StateRunning:
		return "running"
	case StateStopping:
		return "stopping"
	default:
		return "unknown"
	}
}

// RegisterDatabase 注册数据库连接
func (c *Collector) RegisterDatabase(name, endpoint string, db *gorm.DB) {
	if db == nil {
		c.recordError("attempted to register nil database")
		return
	}

	info := &ConnectionInfo{
		Type:     MetricTypeDatabase,
		Name:     name,
		Endpoint: endpoint,
		DB:       db,
	}

	c.connections.Store(name, info)
	c.healthCheck.RegisterDatabase(name, db)
}

// RegisterRedis 注册Redis连接
func (c *Collector) RegisterRedis(name, endpoint string, client *redis.Client) {
	if client == nil {
		c.recordError("attempted to register nil redis client")
		return
	}

	info := &ConnectionInfo{
		Type:     MetricTypeRedis,
		Name:     name,
		Endpoint: endpoint,
		Redis:    client,
	}

	c.connections.Store(name, info)
	c.healthCheck.RegisterRedis(name, client)
}

// UnregisterConnection 取消注册连接
func (c *Collector) UnregisterConnection(name string) {
	c.connections.Delete(name)
	c.healthCheck.UnregisterConnection(name)
	c.metricsCache.DeletePrefix(name + ":")
}

// SetBatchSize 设置批处理大小
func (c *Collector) SetBatchSize(size int) {
	if size > 0 {
		c.batchSize = size
	}
}

// MetricsCache 指标缓存
type MetricsCache struct {
	cache sync.Map
	ttl   time.Duration
}

// CacheEntry 缓存条目
type CacheEntry struct {
	value     interface{}
	timestamp time.Time
}

// NewMetricsCache 创建指标缓存
func NewMetricsCache(ttl time.Duration) *MetricsCache {
	return &MetricsCache{
		ttl: ttl,
	}
}

// Set 设置缓存
func (mc *MetricsCache) Set(key string, value interface{}) {
	entry := &CacheEntry{
		value:     value,
		timestamp: time.Now(),
	}
	mc.cache.Store(key, entry)
}

// IsValid 检查缓存是否有效
func (mc *MetricsCache) IsValid(key string) bool {
	if value, ok := mc.cache.Load(key); ok {
		entry := value.(*CacheEntry)
		return time.Since(entry.timestamp) < mc.ttl
	}
	return false
}

// Size 获取缓存大小
func (mc *MetricsCache) Size() int {
	size := 0
	mc.cache.Range(func(_, _ interface{}) bool {
		size++
		return true
	})
	return size
}

// DeletePrefix 删除指定前缀的缓存
func (mc *MetricsCache) DeletePrefix(prefix string) {
	mc.cache.Range(func(key, _ interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if len(keyStr) >= len(prefix) && keyStr[:len(prefix)] == prefix {
				mc.cache.Delete(key)
			}
		}
		return true
	})
}

// HealthChecker 健康检查器
type HealthChecker struct {
	connections sync.Map
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动健康检查器
func (hc *HealthChecker) Start(ctx context.Context) error {
	// 可以在这里启动定期清理等任务
	return nil
}

// Stop 停止健康检查器
func (hc *HealthChecker) Stop() {
	hc.cancel()
}

// RegisterDatabase 注册数据库连接
func (hc *HealthChecker) RegisterDatabase(name string, db *gorm.DB) {
	hc.connections.Store(name, &ConnectionInfo{
		Type: MetricTypeDatabase,
		Name: name,
		DB:   db,
	})
}

// RegisterRedis 注册Redis连接
func (hc *HealthChecker) RegisterRedis(name string, client *redis.Client) {
	hc.connections.Store(name, &ConnectionInfo{
		Type:  MetricTypeRedis,
		Name:  name,
		Redis: client,
	})
}

// UnregisterConnection 取消注册连接
func (hc *HealthChecker) UnregisterConnection(name string) {
	hc.connections.Delete(name)
}

// CheckConnection 检查连接健康状态
func (hc *HealthChecker) CheckConnection(info *ConnectionInfo) bool {
	switch info.Type {
	case MetricTypeDatabase:
		if info.DB != nil {
			if sqlDB, err := info.DB.DB(); err == nil {
				return sqlDB.Ping() == nil
			}
		}
	case MetricTypeRedis:
		if info.Redis != nil {
			return info.Redis.Ping(hc.ctx).Err() == nil
		}
	}
	return false
}

// GetStatus 获取健康检查状态
func (hc *HealthChecker) GetStatus() map[string]bool {
	status := make(map[string]bool)
	hc.connections.Range(func(key, value interface{}) bool {
		name := key.(string)
		info := value.(*ConnectionInfo)
		status[name] = hc.CheckConnection(info)
		return true
	})
	return status
}