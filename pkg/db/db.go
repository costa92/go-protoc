package db

import (
	"github.com/google/wire"
	redis "github.com/redis/go-redis/v9"
)

// ProviderSet is db providers.
var ProviderSet = wire.NewSet(
	NewMySQLWithTracing, // 使用带追踪的 MySQL 连接
	NewRedisWithTracing, // 使用带追踪的 Redis 连接
	wire.Bind(new(redis.UniversalClient), new(*redis.Client)), // 正确绑定接口和实现
)
