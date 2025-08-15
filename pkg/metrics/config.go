package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Config 统一的指标配置
type Config struct {
	// 全局配置
	Enabled         bool                  `json:"enabled" mapstructure:"enabled"`
	CollectInterval time.Duration         `json:"collect_interval" mapstructure:"collect_interval"`
	Registry        prometheus.Registerer `json:"-"`
	Namespace       string                `json:"namespace" mapstructure:"namespace"`
	ServiceName     string                `json:"service_name" mapstructure:"service_name"`

	// HTTP配置
	HTTP HTTPConfig `json:"http" mapstructure:"http"`

	// gRPC配置
	GRPC GRPCConfig `json:"grpc" mapstructure:"grpc"`

	// 数据库配置
	Database DatabaseConfig `json:"database" mapstructure:"database"`

	// Redis配置
	Redis RedisConfig `json:"redis" mapstructure:"redis"`

	// 自定义指标配置
	Custom CustomConfig `json:"custom" mapstructure:"custom"`

	// 健康检查配置
	HealthCheck HealthCheckConfig `json:"health_check" mapstructure:"health_check"`
}

// HTTPConfig HTTP指标配置
type HTTPConfig struct {
	Enabled      bool      `json:"enabled" mapstructure:"enabled"`
	Port         int       `json:"port" mapstructure:"port"`
	Path         string    `json:"path" mapstructure:"path"`
	Buckets      []float64 `json:"buckets" mapstructure:"buckets"`
	ExcludePaths []string  `json:"exclude_paths" mapstructure:"exclude_paths"`
}

// GRPCConfig gRPC指标配置
type GRPCConfig struct {
	Enabled bool      `json:"enabled" mapstructure:"enabled"`
	Buckets []float64 `json:"buckets" mapstructure:"buckets"`
}

// DatabaseConfig 数据库指标配置
type DatabaseConfig struct {
	Enabled   bool           `json:"enabled" mapstructure:"enabled"`
	Databases []DatabaseInfo `json:"databases" mapstructure:"databases"`
	SlowQuery time.Duration  `json:"slow_query" mapstructure:"slow_query"`
	Buckets   []float64      `json:"buckets" mapstructure:"buckets"`
}

// DatabaseInfo 数据库信息
type DatabaseInfo struct {
	Name     string `json:"name" mapstructure:"name"`
	Type     string `json:"type" mapstructure:"type"` // mysql, postgres, sqlite
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`
}

// RedisConfig Redis指标配置
type RedisConfig struct {
	Enabled   bool        `json:"enabled" mapstructure:"enabled"`
	Instances []RedisInfo `json:"instances" mapstructure:"instances"`
	Buckets   []float64   `json:"buckets" mapstructure:"buckets"`
}

// RedisInfo Redis实例信息
type RedisInfo struct {
	Name     string `json:"name" mapstructure:"name"`
	Endpoint string `json:"endpoint" mapstructure:"endpoint"`
}

// CustomConfig 自定义指标配置
type CustomConfig struct {
	Enabled bool `json:"enabled" mapstructure:"enabled"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	Enabled  bool          `json:"enabled" mapstructure:"enabled"`
	Interval time.Duration `json:"interval" mapstructure:"interval"`
	Timeout  time.Duration `json:"timeout" mapstructure:"timeout"`
}

// NewConfig 创建默认配置
func NewConfig() *Config {
	return &Config{
		Enabled:         false,
		CollectInterval: 30 * time.Second,
		Registry:        prometheus.DefaultRegisterer,
		Namespace:       "app",
		ServiceName:     "unknown",

		HTTP: HTTPConfig{
			Enabled:      false,
			Port:         8080,
			Path:         "/metrics",
			Buckets:      []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
			ExcludePaths: []string{"/health", "/metrics"},
		},

		GRPC: GRPCConfig{
			Enabled: false,
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		},

		Database: DatabaseConfig{
			Enabled:   false,
			Databases: []DatabaseInfo{},
			SlowQuery: 1 * time.Second,
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		},

		Redis: RedisConfig{
			Enabled:   false,
			Instances: []RedisInfo{},
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10},
		},

		Custom: CustomConfig{
			Enabled: false,
		},

		HealthCheck: HealthCheckConfig{
			Enabled:  false,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		},
	}
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.CollectInterval <= 0 {
		return fmt.Errorf("collect_interval must be positive, got %v", c.CollectInterval)
	}

	if c.ServiceName == "" {
		return fmt.Errorf("service_name cannot be empty")
	}

	if c.Namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}

	// 验证HTTP配置
	if c.HTTP.Enabled {
		if c.HTTP.Port <= 0 || c.HTTP.Port > 65535 {
			return fmt.Errorf("HTTP port must be between 1 and 65535, got %d", c.HTTP.Port)
		}
		if c.HTTP.Path == "" {
			return fmt.Errorf("HTTP metrics path cannot be empty")
		}
	}

	// 验证数据库配置
	if c.Database.Enabled {
		for _, db := range c.Database.Databases {
			if db.Name == "" {
				return fmt.Errorf("database name cannot be empty")
			}
			if db.Type == "" {
				return fmt.Errorf("database type cannot be empty for database %s", db.Name)
			}
		}
	}

	// 验证Redis配置
	if c.Redis.Enabled {
		for _, redis := range c.Redis.Instances {
			if redis.Name == "" {
				return fmt.Errorf("redis name cannot be empty")
			}
			if redis.Endpoint == "" {
				return fmt.Errorf("redis endpoint cannot be empty for instance %s", redis.Name)
			}
		}
	}

	// 验证健康检查配置
	if c.HealthCheck.Enabled {
		if c.HealthCheck.Interval <= 0 {
			return fmt.Errorf("health check interval must be positive")
		}
		if c.HealthCheck.Timeout <= 0 {
			return fmt.Errorf("health check timeout must be positive")
		}
	}

	return nil
}

// IsAnyEnabled 检查是否有任何指标启用
func (c *Config) IsAnyEnabled() bool {
	return c.Enabled && (c.HTTP.Enabled || c.GRPC.Enabled || c.Database.Enabled || c.Redis.Enabled || c.Custom.Enabled)
}

// GetEnabledServices 获取启用的服务列表
func (c *Config) GetEnabledServices() []string {
	var services []string

	if c.HTTP.Enabled {
		services = append(services, "http")
	}

	if c.GRPC.Enabled {
		services = append(services, "grpc")
	}

	if c.Database.Enabled {
		for _, db := range c.Database.Databases {
			services = append(services, fmt.Sprintf("database:%s", db.Name))
		}
	}

	if c.Redis.Enabled {
		for _, redis := range c.Redis.Instances {
			services = append(services, fmt.Sprintf("redis:%s", redis.Name))
		}
	}

	if c.Custom.Enabled {
		services = append(services, "custom")
	}

	return services
}

// Clone 克隆配置
func (c *Config) Clone() *Config {
	cloned := *c

	// 深拷贝切片
	cloned.Database.Databases = make([]DatabaseInfo, len(c.Database.Databases))
	copy(cloned.Database.Databases, c.Database.Databases)

	cloned.Redis.Instances = make([]RedisInfo, len(c.Redis.Instances))
	copy(cloned.Redis.Instances, c.Redis.Instances)

	cloned.HTTP.ExcludePaths = make([]string, len(c.HTTP.ExcludePaths))
	copy(cloned.HTTP.ExcludePaths, c.HTTP.ExcludePaths)

	cloned.HTTP.Buckets = make([]float64, len(c.HTTP.Buckets))
	copy(cloned.HTTP.Buckets, c.HTTP.Buckets)

	cloned.GRPC.Buckets = make([]float64, len(c.GRPC.Buckets))
	copy(cloned.GRPC.Buckets, c.GRPC.Buckets)

	cloned.Database.Buckets = make([]float64, len(c.Database.Buckets))
	copy(cloned.Database.Buckets, c.Database.Buckets)

	cloned.Redis.Buckets = make([]float64, len(c.Redis.Buckets))
	copy(cloned.Redis.Buckets, c.Redis.Buckets)

	return &cloned
}

// ConfigOption 配置选项函数
type ConfigOption func(*Config)

// WithEnabled 启用指标
func WithEnabled(enabled bool) ConfigOption {
	return func(c *Config) {
		c.Enabled = enabled
	}
}

// WithServiceName 设置服务名称
func WithServiceName(name string) ConfigOption {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithNamespace 设置命名空间
func WithNamespace(namespace string) ConfigOption {
	return func(c *Config) {
		c.Namespace = namespace
	}
}

// WithCollectInterval 设置收集间隔
func WithCollectInterval(interval time.Duration) ConfigOption {
	return func(c *Config) {
		c.CollectInterval = interval
	}
}

// WithRegistry 设置Prometheus注册器
func WithRegistry(registry prometheus.Registerer) ConfigOption {
	return func(c *Config) {
		c.Registry = registry
	}
}

// WithHTTP 启用HTTP指标
func WithHTTP(enabled bool, port int) ConfigOption {
	return func(c *Config) {
		c.HTTP.Enabled = enabled
		if port > 0 {
			c.HTTP.Port = port
		}
	}
}

// WithGRPC 启用gRPC指标
func WithGRPC(enabled bool) ConfigOption {
	return func(c *Config) {
		c.GRPC.Enabled = enabled
	}
}

// WithDatabase 添加数据库配置
func WithDatabase(enabled bool, databases ...DatabaseInfo) ConfigOption {
	return func(c *Config) {
		c.Database.Enabled = enabled
		c.Database.Databases = append(c.Database.Databases, databases...)
	}
}

// WithRedis 添加Redis配置
func WithRedis(enabled bool, instances ...RedisInfo) ConfigOption {
	return func(c *Config) {
		c.Redis.Enabled = enabled
		c.Redis.Instances = append(c.Redis.Instances, instances...)
	}
}

// WithCustom 启用自定义指标
func WithCustom(enabled bool) ConfigOption {
	return func(c *Config) {
		c.Custom.Enabled = enabled
	}
}

// WithHealthCheck 启用健康检查
func WithHealthCheck(enabled bool, interval, timeout time.Duration) ConfigOption {
	return func(c *Config) {
		c.HealthCheck.Enabled = enabled
		if interval > 0 {
			c.HealthCheck.Interval = interval
		}
		if timeout > 0 {
			c.HealthCheck.Timeout = timeout
		}
	}
}

// ApplyOptions 应用配置选项
func (c *Config) ApplyOptions(opts ...ConfigOption) {
	for _, opt := range opts {
		opt(c)
	}
}
