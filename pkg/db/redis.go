package db

import (
	"context"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/monitor"
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
	// +optional 可选的连接池监控器
	Monitor monitor.PoolMonitor
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

	// 如果提供了监控器，注册Redis连接进行监控
	if opts.Monitor != nil {
		// 使用地址作为监控标识
		monitorName := opts.Addr
		if monitorName == "" {
			monitorName = "redis_default"
		}

		// 注册到监控器 - 监控器会自动开始收集连接池统计
		if redisMonitor, ok := opts.Monitor.(monitor.RedisMonitor); ok {
			if err := redisMonitor.RegisterRedis(monitorName, rdb); err != nil {
				// 监控注册失败不应该影响Redis连接创建
				// 只记录错误但继续返回Redis连接
				// 注意：这里没有logger，可以考虑添加到RedisOptions中或使用全局logger
			}
		}
	}

	return rdb, nil
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
}
