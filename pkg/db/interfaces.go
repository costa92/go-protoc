package db

import (
	"context"
)

// Logger 日志记录器接口
type Logger interface {
	// Debug 调试日志
	Debug(args ...interface{})

	// Info 信息日志
	Info(args ...interface{})

	// Warn 警告日志
	Warn(args ...interface{})

	// Error 错误日志
	Error(args ...interface{})

	// With 添加字段
	With(key string, value interface{}) Logger
}

// DatabaseConnection 数据库连接接口
type DatabaseConnection interface {
	// Close 关闭连接
	Close() error

	// Ping 测试连接
	Ping(ctx context.Context) error
}

// ConfigValidator 配置验证器接口
type ConfigValidator interface {
	// Validate 验证配置
	Validate() error

	// GetDefaults 获取默认配置
	GetDefaults() interface{}

	// ApplyDefaults 应用默认配置
	ApplyDefaults() error
}

// LifecycleManager 生命周期管理器接口
type LifecycleManager interface {
	// Initialize 初始化
	Initialize(ctx context.Context) error

	// Start 启动
	Start(ctx context.Context) error

	// Stop 停止
	Stop(ctx context.Context) error

	// Shutdown 关闭
	Shutdown(ctx context.Context) error

	// GetState 获取状态
	GetState() string
}

// 注意：PoolMonitor 接口已移至 pkg/monitor 包以避免循环导入
// 使用: import "github.com/costa92/go-protoc/v2/pkg/monitor"
