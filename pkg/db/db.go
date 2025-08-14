package db

import (
	"github.com/google/wire"
	redis "github.com/redis/go-redis/v9"
)

// ProviderSet is db providers.
// Note: This ProviderSet is now mainly for backward compatibility.
// The preferred approach is to use options.XxxOptions.NewClient() methods directly.
var ProviderSet = wire.NewSet(
	wire.Bind(new(redis.UniversalClient), new(*redis.Client)), // 正确绑定接口和实现
)
