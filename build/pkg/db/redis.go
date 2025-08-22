package db

import (
	"context"
	"runtime"
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
// 优化: 根据CPU核数动态调整Redis连接池配置
func setRedisDefaults(opts *RedisOptions) {
	if opts.Addr == "" {
		opts.Addr = "127.0.0.1:6379"
	}

	// 获取CPU核数用于动态配置
	numCPU := runtime.NumCPU()

	if opts.PoolSize == 0 {
		// 优化: 根据CPU核数设置连接池大小，支持更高并发
		// 公式: CPU核数 * 15，最小30，最大300
		opts.PoolSize = calculateOptimalRedisValue(numCPU*15, 30, 300)
	}
	if opts.MinIdleConns == 0 {
		// 优化: 根据连接池大小设置最小空闲连接数
		// 公式: 连接池大小的20%，最小5，最大50
		idealIdle := opts.PoolSize / 5
		opts.MinIdleConns = calculateOptimalRedisValue(idealIdle, 5, 50)
	}
	if opts.MaxRetries == 0 {
		opts.MaxRetries = 3
	}
	if opts.DialTimeout == 0 {
		// 优化: 根据系统负载调整连接超时时间
		if opts.PoolSize > 100 {
			// 高并发场景: 较短的连接超时
			opts.DialTimeout = 1 * time.Second
		} else {
			// 低并发场景: 较长的连接超时，确保连接成功
			opts.DialTimeout = 3 * time.Second
		}
	}
	if opts.ReadTimeout == 0 {
		// 优化: 设置合理的读超时时间
		opts.ReadTimeout = 2 * time.Second
	}
	if opts.WriteTimeout == 0 {
		// 优化: 设置合理的写超时时间
		opts.WriteTimeout = 2 * time.Second
	}
	if opts.PoolTimeout == 0 {
		// 优化: 根据连接池大小调整等待超时
		if opts.PoolSize > 100 {
			// 大连接池: 较短的等待时间，快速失败
			opts.PoolTimeout = 1 * time.Second
		} else {
			// 小连接池: 较长的等待时间，提高成功率
			opts.PoolTimeout = 3 * time.Second
		}
	}
}

// calculateOptimalRedisValue 计算Redis优化值，确保在合理范围内
func calculateOptimalRedisValue(calculated, min, max int) int {
	if calculated < min {
		return min
	}
	if calculated > max {
		return max
	}
	return calculated
}
