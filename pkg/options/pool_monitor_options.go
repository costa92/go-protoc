package options

import (
	"time"

	"github.com/spf13/pflag"

	"github.com/costa92/go-protoc/v2/pkg/metrics"
	"github.com/costa92/go-protoc/v2/pkg/monitor"
)

var _ IOptions = (*PoolMonitorOptions)(nil)

// PoolMonitorOptions defines options for pool monitoring.
type PoolMonitorOptions struct {
	// Enabled 是否启用连接池监控
	Enabled bool `json:"enabled" mapstructure:"enabled"`

	// Namespace 指标命名空间
	Namespace string `json:"namespace" mapstructure:"namespace"`

	// CollectInterval 收集间隔
	CollectInterval time.Duration `json:"collect-interval" mapstructure:"collect-interval"`

	// Database 数据库监控选项
	Database DatabaseMonitorOptions `json:"database" mapstructure:"database"`

	// Redis Redis监控选项
	Redis RedisMonitorOptions `json:"redis" mapstructure:"redis"`
}

// DatabaseMonitorOptions 数据库监控选项
type DatabaseMonitorOptions struct {
	// Enabled 是否启用数据库监控
	Enabled bool `json:"enabled" mapstructure:"enabled"`

	// SlowQuery 慢查询阈值
	SlowQuery time.Duration `json:"slow-query" mapstructure:"slow-query"`

	// Buckets 延迟分布桶
	Buckets []float64 `json:"buckets" mapstructure:"buckets"`
}

// RedisMonitorOptions Redis监控选项
type RedisMonitorOptions struct {
	// Enabled 是否启用Redis监控
	Enabled bool `json:"enabled" mapstructure:"enabled"`

	// Buckets 延迟分布桶
	Buckets []float64 `json:"buckets" mapstructure:"buckets"`
}

// NewPoolMonitorOptions create a `zero` value instance.
func NewPoolMonitorOptions() *PoolMonitorOptions {
	return &PoolMonitorOptions{
		Enabled:         true, // 默认启用
		Namespace:       "apiserver",
		CollectInterval: 30 * time.Second,
		Database: DatabaseMonitorOptions{
			Enabled:   true,
			SlowQuery: 1 * time.Second,
			Buckets:   []float64{0.001, 0.01, 0.1, 1, 10},
		},
		Redis: RedisMonitorOptions{
			Enabled: true,
			Buckets: []float64{0.001, 0.01, 0.1, 1, 10},
		},
	}
}

// Validate verifies flags passed to PoolMonitorOptions.
func (o *PoolMonitorOptions) Validate() []error {
	errs := []error{}

	if o.CollectInterval <= 0 {
		o.CollectInterval = 30 * time.Second
	}

	if o.Database.SlowQuery <= 0 {
		o.Database.SlowQuery = 1 * time.Second
	}

	if len(o.Database.Buckets) == 0 {
		o.Database.Buckets = []float64{0.001, 0.01, 0.1, 1, 10}
	}

	if len(o.Redis.Buckets) == 0 {
		o.Redis.Buckets = []float64{0.001, 0.01, 0.1, 1, 10}
	}

	return errs
}

// AddFlags adds flags related to pool monitoring for a specific APIServer to the specified FlagSet.
func (o *PoolMonitorOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.BoolVar(&o.Enabled, "pool-monitor.enabled", o.Enabled, "Enable pool monitoring.")
	fs.StringVar(&o.Namespace, "pool-monitor.namespace", o.Namespace, "Pool monitor metrics namespace.")
	fs.DurationVar(&o.CollectInterval, "pool-monitor.collect-interval", o.CollectInterval, "Pool monitor collection interval.")

	// Database monitoring flags
	fs.BoolVar(&o.Database.Enabled, "pool-monitor.database.enabled", o.Database.Enabled, "Enable database pool monitoring.")
	fs.DurationVar(&o.Database.SlowQuery, "pool-monitor.database.slow-query", o.Database.SlowQuery, "Database slow query threshold.")

	// Redis monitoring flags
	fs.BoolVar(&o.Redis.Enabled, "pool-monitor.redis.enabled", o.Redis.Enabled, "Enable redis pool monitoring.")
}

// ToMetricsConfig converts PoolMonitorOptions to metrics.Config
func (o *PoolMonitorOptions) ToMetricsConfig() *metrics.Config {
	if !o.Enabled {
		// 如果监控被禁用，返回一个禁用的配置
		config := metrics.NewConfig()
		config.Database.Enabled = false
		config.Redis.Enabled = false
		return config
	}

	config := metrics.NewConfig()
	config.Namespace = o.Namespace
	config.CollectInterval = o.CollectInterval

	// 数据库监控配置
	config.Database.Enabled = o.Database.Enabled
	config.Database.SlowQuery = o.Database.SlowQuery
	config.Database.Buckets = o.Database.Buckets

	// Redis监控配置
	config.Redis.Enabled = o.Redis.Enabled
	config.Redis.Buckets = o.Redis.Buckets

	return config
}

// CreatePoolMonitor creates a pool monitor based on options
func (o *PoolMonitorOptions) CreatePoolMonitor() monitor.PoolMonitor {
	config := o.ToMetricsConfig()
	return metrics.NewPoolMonitor(config)
}
