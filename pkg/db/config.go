package db

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricsConfig 统一的监控配置
type MetricsConfig struct {
	// 全局配置
	Enabled         bool                  `json:"enabled" mapstructure:"enabled"`
	CollectInterval time.Duration         `json:"collect_interval" mapstructure:"collect_interval"`
	Registry        prometheus.Registerer `json:"-"`

	// 数据库配置
	Database DatabaseMetricsConfig `json:"database" mapstructure:"database"`

	// Redis配置
	Redis RedisMetricsConfig `json:"redis" mapstructure:"redis"`
}

// DatabaseMetricsConfig 数据库监控配置
type DatabaseMetricsConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Name    string `json:"name" mapstructure:"name"`
	Type    string `json:"type" mapstructure:"type"` // mysql, postgres, sqlite
}

// RedisMetricsConfig Redis监控配置
type RedisMetricsConfig struct {
	Enabled bool   `json:"enabled" mapstructure:"enabled"`
	Name    string `json:"name" mapstructure:"name"`
	Addr    string `json:"addr" mapstructure:"addr"`
}

// NewMetricsConfig 创建默认的监控配置
func NewMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Enabled:         false,
		CollectInterval: 30 * time.Second,
		Registry:        prometheus.DefaultRegisterer,
		Database: DatabaseMetricsConfig{
			Enabled: false,
			Name:    "default",
			Type:    "mysql",
		},
		Redis: RedisMetricsConfig{
			Enabled: false,
			Name:    "default",
			Addr:    "127.0.0.1:6379",
		},
	}
}

// Validate 验证配置
func (c *MetricsConfig) Validate() error {
	if c.CollectInterval <= 0 {
		return fmt.Errorf("collect_interval must be positive, got %v", c.CollectInterval)
	}

	if c.Database.Enabled && c.Database.Name == "" {
		return fmt.Errorf("database metrics name cannot be empty when enabled")
	}

	if c.Redis.Enabled && c.Redis.Name == "" {
		return fmt.Errorf("redis metrics name cannot be empty when enabled")
	}

	return nil
}

// IsAnyEnabled 检查是否有任何监控被启用
func (c *MetricsConfig) IsAnyEnabled() bool {
	return c.Enabled && (c.Database.Enabled || c.Redis.Enabled)
}

// GetEnabledServices 获取启用的服务列表
func (c *MetricsConfig) GetEnabledServices() []string {
	var services []string
	if c.Database.Enabled {
		services = append(services, fmt.Sprintf("database:%s", c.Database.Name))
	}
	if c.Redis.Enabled {
		services = append(services, fmt.Sprintf("redis:%s", c.Redis.Name))
	}
	return services
}

// MetricsOption 监控配置选项函数
type MetricsOption func(*MetricsConfig)

// WithEnabled 启用监控
func WithEnabled(enabled bool) MetricsOption {
	return func(c *MetricsConfig) {
		c.Enabled = enabled
	}
}

// WithCollectInterval 设置收集间隔
func WithCollectInterval(interval time.Duration) MetricsOption {
	return func(c *MetricsConfig) {
		c.CollectInterval = interval
	}
}

// WithRegistry 设置Prometheus注册器
func WithRegistry(registry prometheus.Registerer) MetricsOption {
	return func(c *MetricsConfig) {
		c.Registry = registry
	}
}

// WithDatabase 配置数据库监控
func WithDatabase(enabled bool, name, dbType string) MetricsOption {
	return func(c *MetricsConfig) {
		c.Database.Enabled = enabled
		c.Database.Name = name
		c.Database.Type = dbType
	}
}

// WithRedis 配置Redis监控
func WithRedis(enabled bool, name, addr string) MetricsOption {
	return func(c *MetricsConfig) {
		c.Redis.Enabled = enabled
		c.Redis.Name = name
		c.Redis.Addr = addr
	}
}

// ApplyOptions 应用配置选项
func (c *MetricsConfig) ApplyOptions(opts ...MetricsOption) {
	for _, opt := range opts {
		opt(c)
	}
}

// Clone 克隆配置
func (c *MetricsConfig) Clone() *MetricsConfig {
	return &MetricsConfig{
		Enabled:         c.Enabled,
		CollectInterval: c.CollectInterval,
		Registry:        c.Registry,
		Database:        c.Database,
		Redis:           c.Redis,
	}
}
