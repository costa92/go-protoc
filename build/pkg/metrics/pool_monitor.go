package metrics

import (
	"github.com/costa92/go-protoc/v2/pkg/monitor"
)

// PoolMonitorImpl 连接池监控器实现
type PoolMonitorImpl struct {
	dbCollector    *DatabaseCollector
	redisCollector *RedisCollector
}

// NewPoolMonitor 创建新的连接池监控器
func NewPoolMonitor(config *Config) monitor.FullMonitor {
	return &PoolMonitorImpl{
		dbCollector:    NewDatabaseCollector(config),
		redisCollector: NewRedisCollector(config),
	}
}

// NewPoolMonitorWithCollectors 使用现有收集器创建监控器
func NewPoolMonitorWithCollectors(dbCollector *DatabaseCollector, redisCollector *RedisCollector) monitor.FullMonitor {
	return &PoolMonitorImpl{
		dbCollector:    dbCollector,
		redisCollector: redisCollector,
	}
}

// RecordConnection 记录连接数变化
func (p *PoolMonitorImpl) RecordConnection(poolName string, idle, open, inuse int) {
	if p.dbCollector != nil && p.dbCollector.enabled.Load() {
		// 从数据库收集器中找到对应的连接
		if conn, exists := p.dbCollector.connections.Load(poolName); exists {
			dbConn := conn.(*DatabaseConnection)

			// 记录连接状态
			p.dbCollector.poolConnections.WithLabelValues(
				dbConn.Name, dbConn.Type, "idle", dbConn.Endpoint,
			).Set(float64(idle))

			p.dbCollector.poolConnections.WithLabelValues(
				dbConn.Name, dbConn.Type, "open", dbConn.Endpoint,
			).Set(float64(open))

			p.dbCollector.poolConnections.WithLabelValues(
				dbConn.Name, dbConn.Type, "in_use", dbConn.Endpoint,
			).Set(float64(inuse))
		}
	}
}

// RecordOperation 记录操作
func (p *PoolMonitorImpl) RecordOperation(poolName string, operation string, success bool, duration float64) {
	if p.dbCollector != nil && p.dbCollector.enabled.Load() {
		if conn, exists := p.dbCollector.connections.Load(poolName); exists {
			dbConn := conn.(*DatabaseConnection)

			// 记录操作计数
			p.dbCollector.poolOperations.WithLabelValues(
				dbConn.Name, dbConn.Type, operation, dbConn.Endpoint,
			).Inc()

			// 记录操作耗时
			p.dbCollector.poolDuration.WithLabelValues(
				dbConn.Name, dbConn.Type, operation, dbConn.Endpoint,
			).Observe(duration)
		}
	}

	// Redis操作记录
	if p.redisCollector != nil && p.redisCollector.enabled.Load() {
		if conn, exists := p.redisCollector.connections.Load(poolName); exists {
			redisConn := conn.(*RedisConnection)

			status := "success"
			if !success {
				status = "error"
			}

			p.redisCollector.commandTotal.WithLabelValues(
				redisConn.Name, operation, status,
			).Inc()

			p.redisCollector.commandDuration.WithLabelValues(
				redisConn.Name, operation,
			).Observe(duration)
		}
	}
}

// RecordError 记录错误
func (p *PoolMonitorImpl) RecordError(poolName string, operation string, err error) {
	// 错误记录通过 RecordOperation 的 success=false 来处理
	p.RecordOperation(poolName, operation, false, 0)
}

// Start 启动监控器
func (p *PoolMonitorImpl) Start() error {
	if p.dbCollector != nil {
		if err := p.dbCollector.Start(); err != nil {
			return err
		}
	}

	if p.redisCollector != nil {
		if err := p.redisCollector.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Stop 停止监控器
func (p *PoolMonitorImpl) Stop() error {
	if p.dbCollector != nil {
		if err := p.dbCollector.Stop(); err != nil {
			return err
		}
	}

	if p.redisCollector != nil {
		if err := p.redisCollector.Stop(); err != nil {
			return err
		}
	}

	return nil
}

// RegisterDatabase 注册数据库连接以供监控
func (p *PoolMonitorImpl) RegisterDatabase(name string, target interface{}) error {
	if p.dbCollector != nil {
		return p.dbCollector.RegisterTarget(name, target)
	}
	return nil
}

// RegisterRedis 注册Redis连接以供监控
func (p *PoolMonitorImpl) RegisterRedis(name string, target interface{}) error {
	if p.redisCollector != nil {
		return p.redisCollector.RegisterTarget(name, target)
	}
	return nil
}

// UnregisterDatabase 取消注册数据库连接
func (p *PoolMonitorImpl) UnregisterDatabase(name string) error {
	if p.dbCollector != nil {
		return p.dbCollector.UnregisterTarget(name)
	}
	return nil
}

// UnregisterRedis 取消注册Redis连接
func (p *PoolMonitorImpl) UnregisterRedis(name string) error {
	if p.redisCollector != nil {
		return p.redisCollector.UnregisterTarget(name)
	}
	return nil
}

// IsRunning 检查是否运行
func (p *PoolMonitorImpl) IsRunning() bool {
	dbRunning := p.dbCollector != nil && p.dbCollector.IsRunning()
	redisRunning := p.redisCollector != nil && p.redisCollector.IsRunning()
	return dbRunning || redisRunning
}
