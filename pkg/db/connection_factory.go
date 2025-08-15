package db

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// ConnectionType 连接类型
type ConnectionType string

const (
	ConnectionTypeMySQL ConnectionType = "mysql"
	ConnectionTypeRedis ConnectionType = "redis"
)

// ConnectionFactory 连接工厂，统一创建和管理数据库连接
type ConnectionFactory struct {
	config *MetricsConfig
}

// NewConnectionFactory 创建连接工厂
func NewConnectionFactory(config *MetricsConfig) *ConnectionFactory {
	if config == nil {
		config = NewMetricsConfig()
	}
	return &ConnectionFactory{
		config: config,
	}
}

// CreateMySQLConnection 创建MySQL连接
func (cf *ConnectionFactory) CreateMySQLConnection(opts *MySQLOptions) (*gorm.DB, error) {
	if opts == nil {
		return nil, fmt.Errorf("mysql options cannot be nil")
	}

	// 应用监控配置
	if cf.config.Database.Enabled {
		opts.EnableMetrics = true
		if opts.MetricsName == "" {
			opts.MetricsName = cf.config.Database.Name
		}
	}

	// 创建连接
	db, err := NewMySQL(opts)
	if err != nil {
		return nil, err
	}

	// 注册到监控收集器
	if opts.EnableMetrics {
		collector := GetGlobalOptimizedCollector()
		endpoint := fmt.Sprintf("%s/%s", opts.Addr, opts.Database)
		collector.RegisterDatabase(opts.MetricsName, endpoint, db)
	}

	return db, nil
}

// CreateRedisConnection 创建Redis连接
func (cf *ConnectionFactory) CreateRedisConnection(opts *RedisOptions) (*redis.Client, error) {
	if opts == nil {
		return nil, fmt.Errorf("redis options cannot be nil")
	}

	// 应用监控配置
	if cf.config.Redis.Enabled {
		opts.EnableMetrics = true
		if opts.MetricsName == "" {
			opts.MetricsName = cf.config.Redis.Name
		}
	}

	// 创建连接
	client, err := NewRedis(opts)
	if err != nil {
		return nil, err
	}

	// 注册到监控收集器
	if opts.EnableMetrics {
		collector := GetGlobalOptimizedCollector()
		collector.RegisterRedis(opts.MetricsName, opts.Addr, client)
	}

	return client, nil
}

// CreateMonitoredMySQLConnection 创建带监控的MySQL连接
func (cf *ConnectionFactory) CreateMonitoredMySQLConnection(opts *MySQLOptions) (*MonitoredDB, error) {
	if opts == nil {
		return nil, fmt.Errorf("mysql options cannot be nil")
	}

	// 确保启用监控
	opts.EnableMetrics = true
	if opts.MetricsName == "" {
		opts.MetricsName = cf.config.Database.Name
	}

	return NewMySQLWithMetrics(opts)
}

// CreateMonitoredRedisConnection 创建带监控的Redis连接
func (cf *ConnectionFactory) CreateMonitoredRedisConnection(opts *RedisOptions) (*MonitoredRedis, error) {
	if opts == nil {
		return nil, fmt.Errorf("redis options cannot be nil")
	}

	// 确保启用监控
	opts.EnableMetrics = true
	if opts.MetricsName == "" {
		opts.MetricsName = cf.config.Redis.Name
	}

	return NewRedisWithMetrics(opts)
}

// GetSupportedConnectionTypes 获取支持的连接类型
func (cf *ConnectionFactory) GetSupportedConnectionTypes() []ConnectionType {
	return []ConnectionType{
		ConnectionTypeMySQL,
		ConnectionTypeRedis,
	}
}

// ValidateConnectionType 验证连接类型
func (cf *ConnectionFactory) ValidateConnectionType(connType ConnectionType) error {
	supported := cf.GetSupportedConnectionTypes()
	for _, t := range supported {
		if t == connType {
			return nil
		}
	}
	return fmt.Errorf("unsupported connection type: %s", connType)
}

// 全局连接工厂实例
var globalConnectionFactory *ConnectionFactory

// GetGlobalConnectionFactory 获取全局连接工厂
func GetGlobalConnectionFactory() *ConnectionFactory {
	if globalConnectionFactory == nil {
		globalConnectionFactory = NewConnectionFactory(NewMetricsConfig())
	}
	return globalConnectionFactory
}

// SetGlobalConnectionFactory 设置全局连接工厂
func SetGlobalConnectionFactory(factory *ConnectionFactory) {
	globalConnectionFactory = factory
}

// InitializeGlobalConnectionFactory 初始化全局连接工厂
func InitializeGlobalConnectionFactory(config *MetricsConfig) {
	factory := NewConnectionFactory(config)
	SetGlobalConnectionFactory(factory)
}
