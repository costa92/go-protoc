package db

import (
	tracedb "github.com/costa92/go-protoc/v2/pkg/trace/db"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewMySQLWithTracing creates a new GORM MySQL DB instance with OpenTelemetry tracing.
// Uses GORM's official OpenTelemetry plugin for better integration.
func NewMySQLWithTracing(opts *MySQLOptions) (*gorm.DB, error) {
	gormDB, err := NewMySQL(opts)
	if err != nil {
		return nil, err
	}

	// Install GORM's official OpenTelemetry plugin
	plugin := tracedb.NewMySQLPluginWithOptions()
	if err := gormDB.Use(plugin); err != nil {
		return nil, err
	}

	return gormDB, nil
}

// NewRedisWithTracing creates a new Redis client with OpenTelemetry tracing.
func NewRedisWithTracing(opts *RedisOptions) (*redis.Client, error) {
	// Convert our options to redis.Options
	redisOpts := &redis.Options{
		Addr:         opts.Addr,
		Username:     opts.Username,
		Password:     opts.Password,
		DB:           opts.Database,
		MaxRetries:   opts.MaxRetries,
		MinIdleConns: opts.MinIdleConns,
		DialTimeout:  opts.DialTimeout,
		ReadTimeout:  opts.ReadTimeout,
		WriteTimeout: opts.WriteTimeout,
		PoolTimeout:  opts.PoolTimeout,
		PoolSize:     opts.PoolSize,
	}

	// Create traced Redis client using the official instrumentation
	client, err := tracedb.NewTracedRedisClient(redisOpts, redisotel.WithDBStatement(true))
	if err != nil {
		return nil, err
	}

	return client, nil
}
