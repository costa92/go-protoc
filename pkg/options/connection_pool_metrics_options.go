package options

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/pflag"

	"github.com/costa92/go-protoc/v2/pkg/db"
)

var _ IOptions = (*ConnectionPoolMetricsOptions)(nil)

// ConnectionPoolMetricsOptions 连接池监控配置选项
type ConnectionPoolMetricsOptions struct {
	// 全局配置
	Enabled         bool          `json:"enabled" mapstructure:"enabled"`
	CollectInterval time.Duration `json:"collect-interval" mapstructure:"collect-interval"`

	// 数据库监控配置
	Database DatabaseMetricsOptions `json:"database" mapstructure:"database"`

	// Redis监控配置
	Redis RedisMetricsOptions `json:"redis" mapstructure:"redis"`
}

// DatabaseMetricsOptions 数据库监控选项
type DatabaseMetricsOptions struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Name    string `json:"name" mapstructure:"name"`
}

// RedisMetricsOptions Redis监控选项
type RedisMetricsOptions struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Name    string `json:"name" mapstructure:"name"`
}

// NewConnectionPoolMetricsOptions 创建默认连接池监控选项
func NewConnectionPoolMetricsOptions() *ConnectionPoolMetricsOptions {
	return &ConnectionPoolMetricsOptions{
		Enabled:         false,
		CollectInterval: 30 * time.Second,
		Database: DatabaseMetricsOptions{
			Enabled: false,
			Name:    "default",
		},
		Redis: RedisMetricsOptions{
			Enabled: false,
			Name:    "default",
		},
	}
}

// Validate 验证连接池监控选项
func (o *ConnectionPoolMetricsOptions) Validate() []error {
	var errs []error

	if o.CollectInterval <= 0 {
		o.CollectInterval = 30 * time.Second
	}

	if o.Database.Enabled && o.Database.Name == "" {
		o.Database.Name = "default"
	}

	if o.Redis.Enabled && o.Redis.Name == "" {
		o.Redis.Name = "default"
	}

	return errs
}

// AddFlags 添加连接池监控相关的命令行标志
func (o *ConnectionPoolMetricsOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	// 全局监控配置
	fs.BoolVar(&o.Enabled, join(prefixes...)+"pool-metrics.enabled", o.Enabled, "Enable connection pool metrics collection.")
	fs.DurationVar(&o.CollectInterval, join(prefixes...)+"pool-metrics.collect-interval", o.CollectInterval, "Connection pool metrics collection interval.")

	// 数据库监控配置
	fs.BoolVar(&o.Database.Enabled, join(prefixes...)+"pool-metrics.database.enabled", o.Database.Enabled, "Enable database connection pool metrics.")
	fs.StringVar(&o.Database.Name, join(prefixes...)+"pool-metrics.database.name", o.Database.Name, "Database connection pool metrics name for identification.")

	// Redis监控配置
	fs.BoolVar(&o.Redis.Enabled, join(prefixes...)+"pool-metrics.redis.enabled", o.Redis.Enabled, "Enable Redis connection pool metrics.")
	fs.StringVar(&o.Redis.Name, join(prefixes...)+"pool-metrics.redis.name", o.Redis.Name, "Redis connection pool metrics name for identification.")
}

// ToDBConfig 转换为数据库层的监控配置
func (o *ConnectionPoolMetricsOptions) ToDBConfig() *db.MetricsConfig {
	config := db.NewMetricsConfig()

	config.Enabled = o.Enabled
	config.CollectInterval = o.CollectInterval
	config.Registry = prometheus.DefaultRegisterer

	// 数据库配置
	config.Database.Enabled = o.Database.Enabled
	config.Database.Name = o.Database.Name
	config.Database.Type = "mysql" // 默认类型，可以根据实际数据库类型设置

	// Redis配置
	config.Redis.Enabled = o.Redis.Enabled
	config.Redis.Name = o.Redis.Name

	return config
}

// IsAnyEnabled 检查是否启用了任何连接池监控
func (o *ConnectionPoolMetricsOptions) IsAnyEnabled() bool {
	return o.Enabled && (o.Database.Enabled || o.Redis.Enabled)
}

// ApplyToMySQLOptions 应用到MySQL选项
func (o *ConnectionPoolMetricsOptions) ApplyToMySQLOptions(mysqlOpts *MySQLOptions) {
	if o.Database.Enabled {
		mysqlOpts.EnableMetrics = true
		if o.Database.Name != "" {
			mysqlOpts.MetricsName = o.Database.Name
		}
	}
}

// ApplyToRedisOptions 应用到Redis选项
func (o *ConnectionPoolMetricsOptions) ApplyToRedisOptions(redisOpts *RedisOptions) {
	if o.Redis.Enabled {
		redisOpts.EnableMetrics = true
		if o.Redis.Name != "" {
			redisOpts.MetricsName = o.Redis.Name
		}
	}
}
