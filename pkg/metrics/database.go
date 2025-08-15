package metrics

import (
	"database/sql"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// DatabaseCollector 数据库指标收集器
type DatabaseCollector struct {
	config   *Config
	registry prometheus.Registerer
	enabled  atomic.Bool
	running  atomic.Bool
	sampler  Sampler // 优化: 添加采样器降低监控开销

	// 数据库连接池指标
	poolConnections *prometheus.GaugeVec
	poolOperations  *prometheus.CounterVec
	poolDuration    *prometheus.HistogramVec

	// 查询指标
	queryTotal    *prometheus.CounterVec
	queryDuration *prometheus.HistogramVec
	slowQueries   *prometheus.CounterVec

	// 事务指标
	transactionTotal    *prometheus.CounterVec
	transactionDuration *prometheus.HistogramVec

	// 连接管理
	connections sync.Map // map[string]*DatabaseConnection
	mu          sync.RWMutex
}

// DatabaseConnection 数据库连接信息
type DatabaseConnection struct {
	Name     string
	Type     string
	Endpoint string
	DB       *gorm.DB
	SqlDB    *sql.DB
}

// NewDatabaseCollector 创建数据库指标收集器
func NewDatabaseCollector(config *Config) *DatabaseCollector {
	if config == nil {
		config = NewConfig()
	}

	collector := &DatabaseCollector{
		config:   config,
		registry: config.Registry,
		sampler:  NewDefaultSampler(&config.Sampling), // 优化: 初始化采样器
	}

	collector.initMetrics()
	return collector
}

// initMetrics 初始化指标
func (d *DatabaseCollector) initMetrics() {
	namespace := d.config.Namespace
	subsystem := "database"

	// 连接池指标
	d.poolConnections = promauto.With(d.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pool_connections",
			Help:      "Number of database pool connections in various states",
		},
		[]string{"database", "type", "state", "endpoint"},
	)

	d.poolOperations = promauto.With(d.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pool_operations_total",
			Help:      "Total number of database pool operations",
		},
		[]string{"database", "type", "operation", "endpoint"},
	)

	d.poolDuration = promauto.With(d.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pool_operation_duration_seconds",
			Help:      "Duration of database pool operations in seconds",
			Buckets:   d.config.Database.Buckets,
		},
		[]string{"database", "type", "operation", "endpoint"},
	)

	// 查询指标
	d.queryTotal = promauto.With(d.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "queries_total",
			Help:      "Total number of database queries",
		},
		[]string{"database", "operation", "table", "status"},
	)

	d.queryDuration = promauto.With(d.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "query_duration_seconds",
			Help:      "Database query duration in seconds",
			Buckets:   d.config.Database.Buckets,
		},
		[]string{"database", "operation", "table"},
	)

	d.slowQueries = promauto.With(d.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "slow_queries_total",
			Help:      "Total number of slow database queries",
		},
		[]string{"database", "operation", "table"},
	)

	// 事务指标
	d.transactionTotal = promauto.With(d.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "transactions_total",
			Help:      "Total number of database transactions",
		},
		[]string{"database", "operation", "status"},
	)

	d.transactionDuration = promauto.With(d.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "transaction_duration_seconds",
			Help:      "Database transaction duration in seconds",
			Buckets:   d.config.Database.Buckets,
		},
		[]string{"database", "operation"},
	)
}

// Start 启动收集器
func (d *DatabaseCollector) Start() error {
	if !d.running.CompareAndSwap(false, true) {
		return nil // 已经运行
	}

	if !d.config.Database.Enabled {
		d.running.Store(false)
		return nil
	}

	d.enabled.Store(true)

	// 启动定期收集连接池统计
	go d.collectLoop()

	return nil
}

// Stop 停止收集器
func (d *DatabaseCollector) Stop() error {
	if !d.running.CompareAndSwap(true, false) {
		return nil // 未运行
	}

	d.enabled.Store(false)
	return nil
}

// IsRunning 检查是否运行
func (d *DatabaseCollector) IsRunning() bool {
	return d.running.Load()
}

// GetType 获取收集器类型
func (d *DatabaseCollector) GetType() MetricType {
	return MetricTypeDatabase
}

// GetStatus 获取状态
func (d *DatabaseCollector) GetStatus() map[string]interface{} {
	var connectionCount int
	d.connections.Range(func(_, _ interface{}) bool {
		connectionCount++
		return true
	})

	return map[string]interface{}{
		"type":             string(MetricTypeDatabase),
		"enabled":          d.enabled.Load(),
		"running":          d.running.Load(),
		"connection_count": connectionCount,
		"config":           d.config.Database,
	}
}

// RegisterTarget 注册监控目标
func (d *DatabaseCollector) RegisterTarget(name string, target interface{}) error {
	switch db := target.(type) {
	case *gorm.DB:
		return d.RegisterGormDB(name, db)
	case *sql.DB:
		return d.RegisterSqlDB(name, db)
	default:
		return ErrUnsupportedTarget
	}
}

// UnregisterTarget 取消注册监控目标
func (d *DatabaseCollector) UnregisterTarget(name string) error {
	d.connections.Delete(name)
	return nil
}

// RegisterGormDB 注册GORM数据库连接
func (d *DatabaseCollector) RegisterGormDB(name string, db *gorm.DB) error {
	if db == nil {
		return ErrNilTarget
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	dbType := db.Name()

	// 从配置中找到对应的数据库信息
	var endpoint string
	for _, dbInfo := range d.config.Database.Databases {
		if dbInfo.Name == name {
			endpoint = dbInfo.Endpoint
			break
		}
	}

	conn := &DatabaseConnection{
		Name:     name,
		Type:     dbType,
		Endpoint: endpoint,
		DB:       db,
		SqlDB:    sqlDB,
	}

	d.connections.Store(name, conn)
	return nil
}

// RegisterSqlDB 注册SQL数据库连接
func (d *DatabaseCollector) RegisterSqlDB(name string, db *sql.DB) error {
	if db == nil {
		return ErrNilTarget
	}

	// 从配置中找到对应的数据库信息
	var dbType, endpoint string
	for _, dbInfo := range d.config.Database.Databases {
		if dbInfo.Name == name {
			dbType = dbInfo.Type
			endpoint = dbInfo.Endpoint
			break
		}
	}

	conn := &DatabaseConnection{
		Name:     name,
		Type:     dbType,
		Endpoint: endpoint,
		SqlDB:    db,
	}

	d.connections.Store(name, conn)
	return nil
}

// collectLoop 收集循环
func (d *DatabaseCollector) collectLoop() {
	ticker := time.NewTicker(d.config.CollectInterval)
	defer ticker.Stop()

	for d.running.Load() {
		select {
		case <-ticker.C:
			d.collectConnectionPoolStats()
		}
	}
}

// collectConnectionPoolStats 收集连接池统计
func (d *DatabaseCollector) collectConnectionPoolStats() {
	if !d.enabled.Load() {
		return
	}

	d.connections.Range(func(key, value interface{}) bool {
		conn := value.(*DatabaseConnection)
		if conn.SqlDB != nil {
			d.collectSingleConnectionStats(conn)
		}
		return true
	})
}

// collectSingleConnectionStats 收集单个连接的统计
func (d *DatabaseCollector) collectSingleConnectionStats(conn *DatabaseConnection) {
	stats := conn.SqlDB.Stats()

	// 连接状态指标
	connectionStates := map[string]float64{
		"max_open": float64(stats.MaxOpenConnections),
		"open":     float64(stats.OpenConnections),
		"in_use":   float64(stats.InUse),
		"idle":     float64(stats.Idle),
	}

	for state, value := range connectionStates {
		d.poolConnections.WithLabelValues(conn.Name, conn.Type, state, conn.Endpoint).Set(value)
	}

	// 操作计数指标
	operationCounts := map[string]float64{
		"wait":                 float64(stats.WaitCount),
		"max_idle_closed":      float64(stats.MaxIdleClosed),
		"max_idle_time_closed": float64(stats.MaxIdleTimeClosed),
		"max_lifetime_closed":  float64(stats.MaxLifetimeClosed),
	}

	for operation, count := range operationCounts {
		d.poolOperations.WithLabelValues(conn.Name, conn.Type, operation, conn.Endpoint).Add(count)
	}

	// 等待时长
	if stats.WaitDuration > 0 {
		d.poolDuration.WithLabelValues(conn.Name, conn.Type, "wait", conn.Endpoint).Observe(stats.WaitDuration.Seconds())
	}
}

// RecordQuery 记录查询
func (d *DatabaseCollector) RecordQuery(database, operation, table string, duration time.Duration, success bool) {
	if !d.enabled.Load() {
		return
	}

	// 优化: 使用采样器减少监控开销
	sampleType := SampleTypeDatabase
	if !success {
		sampleType = SampleTypeError // 错误使用更高的采样率
	} else if duration > d.config.Database.SlowQuery {
		sampleType = SampleTypeSlowQuery // 慢查询使用更高的采样率
	}

	if !d.sampler.ShouldSample(sampleType) {
		return
	}

	status := "success"
	if !success {
		status = "error"
	}

	d.queryTotal.WithLabelValues(database, operation, table, status).Inc()
	d.queryDuration.WithLabelValues(database, operation, table).Observe(duration.Seconds())

	// 慢查询检测
	if duration > d.config.Database.SlowQuery {
		d.slowQueries.WithLabelValues(database, operation, table).Inc()
	}
}

// RecordConnection 记录连接状态
func (d *DatabaseCollector) RecordConnection(dbName string, state string, count int) {
	if !d.enabled.Load() {
		return
	}

	// 通过连接池统计自动处理，这里暂时不需要额外实现
}

// RecordTransaction 记录事务
func (d *DatabaseCollector) RecordTransaction(database, operation string, duration time.Duration, success bool) {
	if !d.enabled.Load() {
		return
	}

	status := "success"
	if !success {
		status = "error"
	}

	d.transactionTotal.WithLabelValues(database, operation, status).Inc()
	d.transactionDuration.WithLabelValues(database, operation).Observe(duration.Seconds())
}

// GetDatabaseMetrics 获取数据库指标接口
func (d *DatabaseCollector) GetDatabaseMetrics() DatabaseMetrics {
	return d
}

// RedisCollector Redis指标收集器
type RedisCollector struct {
	config   *Config
	registry prometheus.Registerer
	enabled  atomic.Bool
	running  atomic.Bool

	// Redis连接池指标
	poolConnections *prometheus.GaugeVec
	poolOperations  *prometheus.CounterVec

	// Redis命令指标
	commandTotal    *prometheus.CounterVec
	commandDuration *prometheus.HistogramVec

	// Redis键操作指标
	keyOperations *prometheus.CounterVec

	// 连接管理
	connections sync.Map // map[string]*RedisConnection
	mu          sync.RWMutex
}

// RedisConnection Redis连接信息
type RedisConnection struct {
	Name     string
	Endpoint string
	Client   *redis.Client
}

// NewRedisCollector 创建Redis指标收集器
func NewRedisCollector(config *Config) *RedisCollector {
	if config == nil {
		config = NewConfig()
	}

	collector := &RedisCollector{
		config:   config,
		registry: config.Registry,
	}

	collector.initMetrics()
	return collector
}

// initMetrics 初始化Redis指标
func (r *RedisCollector) initMetrics() {
	namespace := r.config.Namespace
	subsystem := "redis"

	// 连接池指标
	r.poolConnections = promauto.With(r.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pool_connections",
			Help:      "Number of Redis pool connections in various states",
		},
		[]string{"instance", "state", "endpoint"},
	)

	r.poolOperations = promauto.With(r.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "pool_operations_total",
			Help:      "Total number of Redis pool operations",
		},
		[]string{"instance", "operation", "endpoint"},
	)

	// 命令指标
	r.commandTotal = promauto.With(r.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "commands_total",
			Help:      "Total number of Redis commands",
		},
		[]string{"instance", "command", "status"},
	)

	r.commandDuration = promauto.With(r.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "command_duration_seconds",
			Help:      "Redis command duration in seconds",
			Buckets:   r.config.Redis.Buckets,
		},
		[]string{"instance", "command"},
	)

	// 键操作指标
	r.keyOperations = promauto.With(r.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "key_operations_total",
			Help:      "Total number of Redis key operations",
		},
		[]string{"instance", "operation", "key_type"},
	)
}

// Start 启动Redis收集器
func (r *RedisCollector) Start() error {
	if !r.running.CompareAndSwap(false, true) {
		return nil
	}

	if !r.config.Redis.Enabled {
		r.running.Store(false)
		return nil
	}

	r.enabled.Store(true)
	go r.collectLoop()
	return nil
}

// Stop 停止Redis收集器
func (r *RedisCollector) Stop() error {
	if !r.running.CompareAndSwap(true, false) {
		return nil
	}

	r.enabled.Store(false)
	return nil
}

// IsRunning 检查是否运行
func (r *RedisCollector) IsRunning() bool {
	return r.running.Load()
}

// GetType 获取收集器类型
func (r *RedisCollector) GetType() MetricType {
	return MetricTypeRedis
}

// GetStatus 获取状态
func (r *RedisCollector) GetStatus() map[string]interface{} {
	var connectionCount int
	r.connections.Range(func(_, _ interface{}) bool {
		connectionCount++
		return true
	})

	return map[string]interface{}{
		"type":             string(MetricTypeRedis),
		"enabled":          r.enabled.Load(),
		"running":          r.running.Load(),
		"connection_count": connectionCount,
		"config":           r.config.Redis,
	}
}

// RegisterTarget 注册监控目标
func (r *RedisCollector) RegisterTarget(name string, target interface{}) error {
	client, ok := target.(*redis.Client)
	if !ok {
		return ErrUnsupportedTarget
	}

	return r.RegisterRedisClient(name, client)
}

// UnregisterTarget 取消注册监控目标
func (r *RedisCollector) UnregisterTarget(name string) error {
	r.connections.Delete(name)
	return nil
}

// RegisterRedisClient 注册Redis客户端
func (r *RedisCollector) RegisterRedisClient(name string, client *redis.Client) error {
	if client == nil {
		return ErrNilTarget
	}

	endpoint := client.Options().Addr

	conn := &RedisConnection{
		Name:     name,
		Endpoint: endpoint,
		Client:   client,
	}

	r.connections.Store(name, conn)
	return nil
}

// collectLoop Redis收集循环
func (r *RedisCollector) collectLoop() {
	ticker := time.NewTicker(r.config.CollectInterval)
	defer ticker.Stop()

	for r.running.Load() {
		select {
		case <-ticker.C:
			r.collectConnectionPoolStats()
		}
	}
}

// collectConnectionPoolStats 收集Redis连接池统计
func (r *RedisCollector) collectConnectionPoolStats() {
	if !r.enabled.Load() {
		return
	}

	r.connections.Range(func(key, value interface{}) bool {
		conn := value.(*RedisConnection)
		r.collectSingleRedisStats(conn)
		return true
	})
}

// collectSingleRedisStats 收集单个Redis连接统计
func (r *RedisCollector) collectSingleRedisStats(conn *RedisConnection) {
	stats := conn.Client.PoolStats()

	// 连接状态指标
	connectionStates := map[string]float64{
		"total":  float64(stats.TotalConns),
		"active": float64(stats.TotalConns - stats.IdleConns),
		"idle":   float64(stats.IdleConns),
		"stale":  float64(stats.StaleConns),
	}

	for state, value := range connectionStates {
		r.poolConnections.WithLabelValues(conn.Name, state, conn.Endpoint).Set(value)
	}

	// 操作计数指标
	operationCounts := map[string]float64{
		"hits":     float64(stats.Hits),
		"misses":   float64(stats.Misses),
		"timeouts": float64(stats.Timeouts),
	}

	for operation, count := range operationCounts {
		r.poolOperations.WithLabelValues(conn.Name, operation, conn.Endpoint).Add(count)
	}
}

// RecordCommand 记录Redis命令
func (r *RedisCollector) RecordCommand(instance, command string, duration time.Duration, success bool) {
	if !r.enabled.Load() {
		return
	}

	status := "success"
	if !success {
		status = "error"
	}

	r.commandTotal.WithLabelValues(instance, command, status).Inc()
	r.commandDuration.WithLabelValues(instance, command).Observe(duration.Seconds())
}

// RecordConnection 记录Redis连接状态
func (r *RedisCollector) RecordConnection(state string, count int) {
	if !r.enabled.Load() {
		return
	}

	// 通过连接池统计自动处理
}

// RecordKeyOperation 记录Redis键操作
func (r *RedisCollector) RecordKeyOperation(instance, operation, keyType string, count int) {
	if !r.enabled.Load() {
		return
	}

	r.keyOperations.WithLabelValues(instance, operation, keyType).Add(float64(count))
}

// GetRedisMetrics 获取Redis指标接口
func (r *RedisCollector) GetRedisMetrics() RedisMetrics {
	return r
}
