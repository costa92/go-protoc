package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是应用程序的配置结构
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Observability ObservabilityConfig `mapstructure:"observability"`
	Middleware    MiddlewareConfig    `mapstructure:"middleware"`
}

// ServerConfig 包含服务器相关配置
type ServerConfig struct {
	HTTP HTTPConfig `mapstructure:"http"`
	GRPC GRPCConfig `mapstructure:"grpc"`
}

// HTTPConfig 包含HTTP服务相关配置
type HTTPConfig struct {
	Addr    string `mapstructure:"addr"`
	Timeout int    `mapstructure:"timeout"`
}

// GRPCConfig 包含gRPC服务相关配置
type GRPCConfig struct {
	Addr            string `mapstructure:"addr"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

// ObservabilityConfig 包含可观测性相关配置
type ObservabilityConfig struct {
	Tracing   TracingConfig `mapstructure:"tracing"`
	Metrics   MetricsConfig `mapstructure:"metrics"`
	SkipPaths []string      `mapstructure:"skip_paths"`
}

// TracingConfig 包含链路追踪相关配置
type TracingConfig struct {
	ServiceName  string `mapstructure:"service_name"`
	Enabled      bool   `mapstructure:"enabled"`
	Exporter     string `mapstructure:"exporter"`
	OTLPEndpoint string `mapstructure:"otlp_endpoint"`
}

// MetricsConfig 包含指标监控相关配置
type MetricsConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// MiddlewareConfig 包含中间件相关配置
type MiddlewareConfig struct {
	Timeout   time.Duration   `mapstructure:"timeout"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
}

// CORSConfig 定义跨域配置
type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

// RateLimitConfig 定义限流配置
type RateLimitConfig struct {
	Enable bool          `mapstructure:"enable"`
	Limit  int           `mapstructure:"limit"`
	Burst  int           `mapstructure:"burst"`
	Window time.Duration `mapstructure:"window"`
}

// LoadConfig 从指定的文件加载配置
func LoadConfig(configPath string) (*Config, error) {
	// 创建一个新的Viper实例
	v := viper.New()

	// 设置配置文件路径
	v.SetConfigFile(configPath)

	// 尝试读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 自动加载环境变量覆盖配置
	// 环境变量格式为：GO_PROTOC_SERVER_HTTP_ADDR=:8080
	v.SetEnvPrefix("GO_PROTOC")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 解析配置到结构体
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &config, nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTP: HTTPConfig{
				Addr:    ":8090",
				Timeout: 5,
			},
			GRPC: GRPCConfig{
				Addr:            ":8091",
				ShutdownTimeout: 10,
			},
		},
		Observability: ObservabilityConfig{
			Tracing: TracingConfig{
				ServiceName:  "go-protoc-service",
				Enabled:      true,
				Exporter:     "stdout",
				OTLPEndpoint: "localhost:4317",
			},
			Metrics: MetricsConfig{
				Enabled: true,
				Path:    "/metrics",
			},
			SkipPaths: []string{
				"/metrics",
				"/debug/",
				"/swagger/",
				"/healthz",
				"/favicon.ico",
			},
		},
		Middleware: MiddlewareConfig{
			Timeout: 30 * time.Second,
			CORS: CORSConfig{
				AllowOrigins: []string{"*"},
				AllowMethods: []string{
					"GET",
					"POST",
					"PUT",
					"DELETE",
					"OPTIONS",
					"HEAD",
				},
				AllowHeaders: []string{
					"Authorization",
					"Content-Type",
					"X-Request-ID",
					"X-Real-IP",
				},
				ExposeHeaders:    []string{},
				AllowCredentials: true,
				MaxAge:           12 * time.Hour,
			},
			RateLimit: RateLimitConfig{
				Enable: true,
				Limit:  100,
				Burst:  200,
				Window: time.Minute,
			},
		},
	}
}
