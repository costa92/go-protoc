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

// setRedisDefaults set available default values for some fields.
func setRedisDefaults(opts *RedisOptions) {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:6379"
	}
	if opts.PoolSize == 0 {
		opts.PoolSize = 100 // 设置合理的连接池大小
	}
	if opts.MinIdleConns == 0 {
		opts.MinIdleConns = 10
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}
	if opts.DialTimeout == 0 {
		opts.DialTimeout = 5 * time.Second
	}
	if opts.ReadTimeout == 0 {
		opts.ReadTimeout = 3 * time.Second
	}
	if opts.WriteTimeout == 0 {
		opts.WriteTimeout = 3 * time.Second
	}
	if opts.PoolTimeout == 0 {
		opts.PoolTimeout = 4 * time.Second
	}
}
