package options

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/spf13/pflag"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/costa92/go-protoc/v2/pkg/db"
	loggerPkg "github.com/costa92/go-protoc/v2/pkg/logger"
	"github.com/costa92/go-protoc/v2/pkg/monitor"
)

var _ IOptions = (*MySQLOptions)(nil)

// MySQLOptions defines options for mysql database.
type MySQLOptions struct {
	Addr                  string        `json:"addr,omitempty" mapstructure:"addr"`
	Username              string        `json:"username,omitempty" mapstructure:"username"`
	Password              string        `json:"-" mapstructure:"password"`
	Database              string        `json:"database" mapstructure:"database"`
	MaxIdleConnections    int           `json:"max-idle-connections,omitempty" mapstructure:"max-idle-connections,omitempty"`
	MaxOpenConnections    int           `json:"max-open-connections,omitempty" mapstructure:"max-open-connections"`
	MaxConnectionLifeTime time.Duration `json:"max-connection-life-time,omitempty" mapstructure:"max-connection-life-time"`
	LogLevel              int           `json:"log-level" mapstructure:"log-level"`
}

// NewMySQLOptions create a `zero` value instance.
func NewMySQLOptions() *MySQLOptions {
	return &MySQLOptions{
		Addr:                  "127.0.0.1:3306",
		Username:              "onex",
		Password:              "onex(#)666",
		Database:              "onex",
		MaxIdleConnections:    100,
		MaxOpenConnections:    100,
		MaxConnectionLifeTime: time.Duration(10) * time.Second,
		LogLevel:              1, // Silent
	}
}

// Validate verifies flags passed to MySQLOptions.
func (o *MySQLOptions) Validate() []error {
	errs := []error{}

	if o.Addr == "" {
		errs = append(errs, fmt.Errorf("mysql addr cannot be empty"))
	}
	if o.Username == "" {
		errs = append(errs, fmt.Errorf("mysql username cannot be empty"))
	}
	if o.Database == "" {
		errs = append(errs, fmt.Errorf("mysql database cannot be empty"))
	}

	return errs
}

// TestConnection tests the actual connection to MySQL database.
func (o *MySQLOptions) TestConnection() error {
	if o.Addr == "" || o.Username == "" || o.Database == "" {
		return fmt.Errorf("MySQL connection parameters not configured")
	}

	// 构建DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		o.Username, o.Password, o.Addr, o.Database)

	// 尝试连接数据库
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}
	defer db.Close()

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试连接
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	return nil
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *MySQLOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Addr, join(prefixes...)+"mysql.host", o.Addr, ""+
		"MySQL service host address. If left blank, the following related mysql options will be ignored.")
	fs.StringVar(&o.Username, join(prefixes...)+"mysql.username", o.Username, "Username for access to mysql service.")
	fs.StringVar(&o.Password, join(prefixes...)+"mysql.password", o.Password, ""+
		"Password for access to mysql, should be used pair with password.")
	fs.StringVar(&o.Database, join(prefixes...)+"mysql.database", o.Database, ""+
		"Database name for the server to use.")
	fs.IntVar(&o.MaxIdleConnections, join(prefixes...)+"mysql.max-idle-connections", o.MaxIdleConnections, ""+
		"Maximum idle connections allowed to connect to mysql.")
	fs.IntVar(&o.MaxOpenConnections, join(prefixes...)+"mysql.max-open-connections", o.MaxOpenConnections, ""+
		"Maximum open connections allowed to connect to mysql.")
	fs.DurationVar(&o.MaxConnectionLifeTime, join(prefixes...)+"mysql.max-connection-life-time", o.MaxConnectionLifeTime, ""+
		"Maximum connection life time allowed to connect to mysql.")
	fs.IntVar(&o.LogLevel, join(prefixes...)+"mysql.log-mode", o.LogLevel, ""+
		"Specify gorm log level.")
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

// NewDB create mysql store with the given config.
func (o *MySQLOptions) NewDB() (*gorm.DB, error) {
	opts := &db.MySQLOptions{
		Addr:                  o.Addr,
		Username:              o.Username,
		Password:              o.Password,
		Database:              o.Database,
		MaxIdleConnections:    o.MaxIdleConnections,
		MaxOpenConnections:    o.MaxOpenConnections,
		MaxConnectionLifeTime: o.MaxConnectionLifeTime,
		Logger:                loggerPkg.NewGormLogger(loggerPkg.DefaultOptions(), gormlogger.LogLevel(o.LogLevel)),
	}

	return db.NewMySQL(opts)
}

// NewDBWithMonitor create mysql store with the given config and monitor.
func (o *MySQLOptions) NewDBWithMonitor(poolMonitor monitor.PoolMonitor) (*gorm.DB, error) {
	opts := &db.MySQLOptions{
		Addr:                  o.Addr,
		Username:              o.Username,
		Password:              o.Password,
		Database:              o.Database,
		MaxIdleConnections:    o.MaxIdleConnections,
		MaxOpenConnections:    o.MaxOpenConnections,
		MaxConnectionLifeTime: o.MaxConnectionLifeTime,
		Logger:                loggerPkg.NewGormLogger(loggerPkg.DefaultOptions(), gormlogger.LogLevel(o.LogLevel)),
		Monitor:               poolMonitor,
	}

	return db.NewMySQL(opts)
}
