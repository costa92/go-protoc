package db

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisOptions defines options for redis database.
type RedisOptions struct {
	Addr         string
	Username     string
	Password     string
	Database     int
	MaxRetries   int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	PoolTimeout  time.Duration
	PoolSize     int
	// +optional
	EnableMetrics bool
	// +optional
	MetricsName string
}

// NewRedis create a new redis db instance with the given options.
func NewRedis(opts *RedisOptions) (*redis.Client, error) {
	// Set default values to ensure all fields in opts are available.
	setRedisDefaults(opts)

	options := &redis.Options{
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

	rdb := redis.NewClient(options)

	// check redis if is ok
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		return nil, err
	}

	return rdb, nil
}

// NewRedisWithMetrics create a new redis client instance with metrics monitoring.
func NewRedisWithMetrics(opts *RedisOptions) (*MonitoredRedis, error) {
	client, err := NewRedis(opts)
	if err != nil {
		return nil, err
	}

	if opts.EnableMetrics {
		metrics := GetGlobalMetrics()
		metricsName := opts.MetricsName
		if metricsName == "" {
			metricsName = "redis_default"
		}

		monitoredRedis := NewMonitoredRedis(client, metrics, metricsName)

		// 初始收集一次指标
		monitoredRedis.CollectMetrics()

		return monitoredRedis, nil
	}

	// 如果未启用监控，返回包装的未监控实例
	return &MonitoredRedis{Client: client, redisName: "redis_default"}, nil
}

// setRedisDefaults set available default values for some fields.
func setRedisDefaults(opts *RedisOptions) {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:6379"
	}
	if opts.PoolSize == 0 {
		// 优化: 增加连接池大小以支持更高并发
		opts.PoolSize = 200
	}
	if opts.MinIdleConns == 0 {
		// 优化: 增加最小空闲连接数，减少连接建立开销
		opts.MinIdleConns = 50
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}
	if opts.DialTimeout == 0 {
		// 优化: 减少连接超时时间
		opts.DialTimeout = 2 * time.Second
	}
	if opts.ReadTimeout == 0 {
		// 优化: 减少读超时时间，提高响应速度
		opts.ReadTimeout = 1 * time.Second
	}
	if opts.WriteTimeout == 0 {
		// 优化: 减少写超时时间，提高响应速度
		opts.WriteTimeout = 1 * time.Second
	}
	if opts.PoolTimeout == 0 {
		// 优化: 减少连接池等待超时
		opts.PoolTimeout = 2 * time.Second
	}
	if opts.MetricsName == "" && opts.EnableMetrics {
		opts.MetricsName = "redis_default"
	}
}
