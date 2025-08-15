package db

import (
	"fmt"
	"time"

	"database/sql"

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
	// +optional
	EnableMetrics bool
	// +optional
	MetricsName string
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

	return db, nil
}

// NewMySQLWithMetrics create a new gorm db instance with metrics monitoring.
func NewMySQLWithMetrics(opts *MySQLOptions) (*MonitoredDB, error) {
	db, err := NewMySQL(opts)
	if err != nil {
		return nil, err
	}

	if opts.EnableMetrics {
		metrics := GetGlobalMetrics()
		metricsName := opts.MetricsName
		if metricsName == "" {
			metricsName = opts.Database
		}

		monitoredDB := NewMonitoredDB(db, metrics, metricsName)

		// 初始收集一次指标
		if err := monitoredDB.CollectMetrics(); err != nil {
			return nil, fmt.Errorf("failed to collect initial metrics: %w", err)
		}

		return monitoredDB, nil
	}

	// 如果未启用监控，返回包装的未监控实例
	return &MonitoredDB{DB: db, databaseName: opts.Database}, nil
}

// setMySQLDefaults set available default values for some fields.
func setMySQLDefaults(opts *MySQLOptions) {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:3306"
	}
	if opts.MaxIdleConnections == 0 {
		// 优化: 降低空闲连接数，减少资源占用
		opts.MaxIdleConnections = 25
	}
	if opts.MaxOpenConnections == 0 {
		// 优化: 增加最大连接数，支持更高并发
		opts.MaxOpenConnections = 200
	}
	if opts.MaxConnectionLifeTime == 0 {
		// 优化: 增加连接生命周期，减少连接重建开销
		opts.MaxConnectionLifeTime = time.Duration(30) * time.Minute
	}
	if opts.Logger == nil {
		opts.Logger = logger.Default
	}
	if opts.MetricsName == "" && opts.EnableMetrics {
		opts.MetricsName = opts.Database
		if opts.MetricsName == "" {
			opts.MetricsName = "mysql_default"
		}
	}
}

func MustRawDB(db *gorm.DB) *sql.DB {
	raw, err := db.DB()
	if err != nil {
		panic(err)
	}
	return raw
}
