package apiserver

import (
	"github.com/costa92/go-protoc/v2/pkg/metrics"
	"github.com/costa92/go-protoc/v2/pkg/monitor"
	"github.com/costa92/go-protoc/v2/pkg/options"
	"github.com/google/wire"
	"gorm.io/gorm"
)

// ProviderSet 是包含所有 apiserver 相关 providers 的 wire 集合
var ProviderSet = wire.NewSet(
	ProvidePoolMonitor,
	ProvideGormDB,
	ProvideJaegerOptions,
)

// ProvidePoolMonitor 提供连接池监控器（基于配置）
func ProvidePoolMonitor(cfg *Config) monitor.PoolMonitor {
	// 从应用配置中读取监控设置
	if cfg.PoolMonitorOptions == nil {
		// 如果没有配置监控选项，返回一个禁用的监控器
		config := metrics.NewConfig()
		config.Database.Enabled = false
		config.Redis.Enabled = false
		return metrics.NewPoolMonitor(config)
	}

	// 使用配置创建监控器
	return cfg.PoolMonitorOptions.CreatePoolMonitor()
}

// ProvideGormDB provides GORM database connection using options
func ProvideGormDB(cfg *Config, poolMonitor monitor.PoolMonitor) (*gorm.DB, error) {
	return cfg.MySQLOptions.NewDBWithMonitor(poolMonitor)
}

// ProvideJaegerOptions 提供 Jaeger 配置选项
func ProvideJaegerOptions(cfg *Config) *options.JaegerOptions {
	return cfg.JaegerOptions
}
