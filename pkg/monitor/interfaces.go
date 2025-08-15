package monitor

// PoolMonitor 连接池监控接口 - 独立包避免循环导入
type PoolMonitor interface {
	// RecordConnection 记录连接数变化
	RecordConnection(poolName string, idle, open, inuse int)

	// RecordOperation 记录操作
	RecordOperation(poolName string, operation string, success bool, duration float64)

	// RecordError 记录错误
	RecordError(poolName string, operation string, err error)
}

// DatabaseMonitor 数据库监控接口扩展
type DatabaseMonitor interface {
	PoolMonitor

	// RegisterDatabase 注册数据库连接以供监控
	RegisterDatabase(name string, target interface{}) error

	// UnregisterDatabase 取消注册数据库连接
	UnregisterDatabase(name string) error
}

// RedisMonitor Redis监控接口扩展
type RedisMonitor interface {
	PoolMonitor

	// RegisterRedis 注册Redis连接以供监控
	RegisterRedis(name string, target interface{}) error

	// UnregisterRedis 取消注册Redis连接
	UnregisterRedis(name string) error
}

// FullMonitor 完整监控接口
type FullMonitor interface {
	DatabaseMonitor
	RedisMonitor

	// Start 启动监控器
	Start() error

	// Stop 停止监控器
	Stop() error

	// IsRunning 检查是否运行
	IsRunning() bool
}
