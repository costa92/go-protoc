package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 是应用程序的配置结构
type Config struct {
	Server        ServerConfig        `mapstructure:"server"`
	Observability ObservabilityConfig `mapstructure:"observability"`
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
	Tracing TracingConfig `mapstructure:"tracing"`
	Metrics MetricsConfig `mapstructure:"metrics"`
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
