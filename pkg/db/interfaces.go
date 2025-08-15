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

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	// Start 启动收集器
	Start() error

	// Stop 停止收集器
	Stop() error

	// IsRunning 检查是否正在运行
	IsRunning() bool

	// RegisterDatabase 注册数据库连接
	RegisterDatabase(name, endpoint string, db interface{}) error

	// RegisterRedis 注册Redis连接
	RegisterRedis(name, endpoint string, client interface{}) error

	// UnregisterConnection 取消注册连接
	UnregisterConnection(name string) error

	// GetStatus 获取收集器状态
	GetStatus() map[string]interface{}
}

// ConnectionManager 连接管理器接口
type ConnectionManager interface {
	// CreateConnection 创建连接
	CreateConnection(connType ConnectionType, config interface{}) (interface{}, error)

	// RegisterConnection 注册连接
	RegisterConnection(name string, conn interface{}) error

	// UnregisterConnection 取消注册连接
	UnregisterConnection(name string) error

	// GetConnection 获取连接
	GetConnection(name string) (interface{}, error)

	// ListConnections 列出所有连接
	ListConnections() []string

	// HealthCheck 健康检查
	HealthCheck(ctx context.Context) error
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