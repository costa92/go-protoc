package cache

import "errors"

var (
	// ErrCacheNotFound 缓存未找到
	ErrCacheNotFound = errors.New("cache not found")

	// ErrCacheExpired 缓存已过期
	ErrCacheExpired = errors.New("cache expired")

	// ErrCacheInvalid 缓存数据无效
	ErrCacheInvalid = errors.New("cache data invalid")
)
