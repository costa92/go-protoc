package db

import (
	"fmt"
	"runtime"
	"time"

	"database/sql"

	"github.com/costa92/go-protoc/v2/pkg/monitor"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	Addr                  string
	Username              string
	Password              string
	Database              string
	MaxIdleConnections    int
	MaxOpenConnections    int
	MaxConnectionLifeTime time.Duration
	// +optional
	Logger logger.Interface
	// +optional 可选的连接池监控器
	Monitor monitor.PoolMonitor
}

// DSN return DSN from MySQLOptions.
func (o *MySQLOptions) DSN() string {
	return fmt.Sprintf(`%s:%s@tcp(%s)/%s?charset=utf8&parseTime=%t&loc=%s`,
		o.Username,
		o.Password,
		o.Addr,
		o.Database,
		true,
		"Local")
}

// NewMySQL create a new gorm db instance with the given options.
func NewMySQL(opts *MySQLOptions) (*gorm.DB, error) {
	// Set default values to ensure all fields in opts are available.
	setMySQLDefaults(opts)

	db, err := gorm.Open(mysql.Open(opts.DSN()), &gorm.Config{
		// PrepareStmt executes the given query in cached statement.
		// This can improve performance.
		PrepareStmt: true,
		Logger:      opts.Logger,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(opts.MaxOpenConnections)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(opts.MaxConnectionLifeTime)

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(opts.MaxIdleConnections)

	// 如果提供了监控器，注册数据库连接进行监控
	if opts.Monitor != nil {
		// 使用数据库名称作为监控标识，如果为空则使用默认名称
		monitorName := opts.Database
		if monitorName == "" {
			monitorName = "mysql_default"
		}

		// 注册到监控器 - 监控器会自动开始收集连接池统计
		if dbMonitor, ok := opts.Monitor.(monitor.DatabaseMonitor); ok {
			if err := dbMonitor.RegisterDatabase(monitorName, db); err != nil {
				// 监控注册失败不应该影响数据库连接创建
				// 只记录错误但继续返回数据库连接
				if opts.Logger != nil {
					opts.Logger.Error(nil, "Failed to register database monitor", "error", err, "database", monitorName)
				}
			}
		}
	}

	return db, nil
}

// setMySQLDefaults set available default values for some fields.
// 优化: 根据CPU核数和系统负载动态调整连接池配置
func setMySQLDefaults(opts *MySQLOptions) {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:3306"
	}

	// 获取CPU核数用于动态配置
	numCPU := runtime.NumCPU()

	if opts.MaxIdleConnections == 0 {
		// 优化: 根据CPU核数设置空闲连接数，减少资源占用同时保证性能
		// 公式: CPU核数 * 5，最小10，最大50
		opts.MaxIdleConnections = calculateOptimalValue(numCPU*5, 10, 50)
	}
	if opts.MaxOpenConnections == 0 {
		// 优化: 根据CPU核数设置最大连接数，支持更高并发
		// 公式: CPU核数 * 25，最小50，最大500
		opts.MaxOpenConnections = calculateOptimalValue(numCPU*25, 50, 500)
	}
	if opts.MaxConnectionLifeTime == 0 {
		// 优化: 设置合理的连接生命周期，平衡性能和资源使用
		// 高负载环境使用较短的生命周期，低负载环境使用较长的生命周期
		if opts.MaxOpenConnections > 200 {
			// 高并发场景: 较短的连接生命周期，避免连接堆积
			opts.MaxConnectionLifeTime = 10 * time.Minute
		} else {
			// 低并发场景: 较长的连接生命周期，减少重建开销
			opts.MaxConnectionLifeTime = 30 * time.Minute
		}
	}
	if opts.Logger == nil {
		opts.Logger = logger.Default
	}
}

// calculateOptimalValue 计算优化值，确保在合理范围内
func calculateOptimalValue(calculated, min, max int) int {
	if calculated < min {
		return min
	}
	if calculated > max {
		return max
	}
	return calculated
}

func MustRawDB(db *gorm.DB) *sql.DB {
	raw, err := db.DB()
	if err != nil {
		panic(err)
	}
	return raw
}
