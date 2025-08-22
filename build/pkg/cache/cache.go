package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache 统一缓存接口
type Cache interface {
	// Get 获取缓存
	Get(ctx context.Context, key string, dest interface{}) error
	// Set 设置缓存
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	// Delete 删除缓存
	Delete(ctx context.Context, key string) error
	// Exists 检查缓存是否存在
	Exists(ctx context.Context, key string) (bool, error)
	// Clear 清空缓存
	Clear(ctx context.Context) error
}

// MultiLevelCache 多级缓存实现
type MultiLevelCache struct {
	localCache *LocalCache
	redisCache *RedisCache
	localTTL   time.Duration // 本地缓存TTL
	redisTTL   time.Duration // Redis缓存TTL
}

// LocalCache 本地内存缓存
type LocalCache struct {
	data map[string]*cacheItem
	mu   sync.RWMutex
	ttl  time.Duration
}

type cacheItem struct {
	value     interface{}
	expiredAt time.Time
}

// RedisCache Redis缓存
type RedisCache struct {
	client redis.Cmdable
	prefix string
}

// NewMultiLevelCache 创建多级缓存
func NewMultiLevelCache(redisClient redis.Cmdable, localTTL, redisTTL time.Duration) *MultiLevelCache {
	return &MultiLevelCache{
		localCache: NewLocalCache(localTTL),
		redisCache: NewRedisCache(redisClient, "app:cache:"),
		localTTL:   localTTL,
		redisTTL:   redisTTL,
	}
}

// NewLocalCache 创建本地缓存
func NewLocalCache(ttl time.Duration) *LocalCache {
	cache := &LocalCache{
		data: make(map[string]*cacheItem, 1000), // 预分配1000个条目
		ttl:  ttl,
	}

	// 启动清理器
	go cache.startCleaner()

	return cache
}

// NewRedisCache 创建Redis缓存
func NewRedisCache(client redis.Cmdable, prefix string) *RedisCache {
	return &RedisCache{
		client: client,
		prefix: prefix,
	}
}

// Get 多级缓存获取
func (c *MultiLevelCache) Get(ctx context.Context, key string, dest interface{}) error {
	// L1: 本地缓存
	if err := c.localCache.Get(ctx, key, dest); err == nil {
		return nil
	}

	// L2: Redis缓存
	if err := c.redisCache.Get(ctx, key, dest); err == nil {
		// 回写到本地缓存
		_ = c.localCache.Set(ctx, key, dest, c.localTTL)
		return nil
	}

	return ErrCacheNotFound
}

// Set 多级缓存设置
func (c *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 设置到两级缓存
	err1 := c.localCache.Set(ctx, key, value, c.localTTL)
	err2 := c.redisCache.Set(ctx, key, value, ttl)

	// 只要有一个成功就算成功
	if err1 != nil && err2 != nil {
		return fmt.Errorf("both cache levels failed: local=%v, redis=%v", err1, err2)
	}

	return nil
}

// Delete 多级缓存删除
func (c *MultiLevelCache) Delete(ctx context.Context, key string) error {
	err1 := c.localCache.Delete(ctx, key)
	err2 := c.redisCache.Delete(ctx, key)

	// 只要有一个成功就算成功
	if err1 != nil && err2 != nil {
		return fmt.Errorf("both cache levels failed: local=%v, redis=%v", err1, err2)
	}

	return nil
}

// Exists 检查缓存是否存在
func (c *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	// 先检查本地缓存
	if exists, err := c.localCache.Exists(ctx, key); err == nil && exists {
		return true, nil
	}

	// 再检查Redis缓存
	return c.redisCache.Exists(ctx, key)
}

// Clear 清空缓存
func (c *MultiLevelCache) Clear(ctx context.Context) error {
	err1 := c.localCache.Clear(ctx)
	err2 := c.redisCache.Clear(ctx)

	if err1 != nil && err2 != nil {
		return fmt.Errorf("both cache levels failed: local=%v, redis=%v", err1, err2)
	}

	return nil
}

// LocalCache 实现

func (c *LocalCache) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return ErrCacheNotFound
	}

	if time.Now().After(item.expiredAt) {
		c.Delete(ctx, key)
		return ErrCacheExpired
	}

	// 复制值到目标
	return copyValue(item.value, dest)
}

func (c *LocalCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.ttl
	}

	item := &cacheItem{
		value:     value,
		expiredAt: time.Now().Add(ttl),
	}

	c.mu.Lock()
	c.data[key] = item
	c.mu.Unlock()

	return nil
}

func (c *LocalCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()

	return nil
}

func (c *LocalCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	item, exists := c.data[key]
	c.mu.RUnlock()

	if !exists {
		return false, nil
	}

	if time.Now().After(item.expiredAt) {
		c.Delete(ctx, key)
		return false, nil
	}

	return true, nil
}

func (c *LocalCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	c.data = make(map[string]*cacheItem, 1000)
	c.mu.Unlock()

	return nil
}

func (c *LocalCache) startCleaner() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		c.cleanExpired()
	}
}

func (c *LocalCache) cleanExpired() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.data {
		if now.After(item.expiredAt) {
			delete(c.data, key)
		}
	}
}

// RedisCache 实现

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.prefix + key
	data, err := c.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheNotFound
		}
		return err
	}

	return json.Unmarshal(data, dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.prefix + key
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, fullKey, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := c.prefix + key
	return c.client.Del(ctx, fullKey).Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.prefix + key
	count, err := c.client.Exists(ctx, fullKey).Result()
	return count > 0, err
}

func (c *RedisCache) Clear(ctx context.Context) error {
	pattern := c.prefix + "*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()

	keys := make([]string, 0, 100)
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())

		// 批量删除，避免一次性删除太多
		if len(keys) >= 100 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0] // 重置切片
		}
	}

	// 删除剩余的key
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return iter.Err()
}

// 工具函数

func copyValue(src, dest interface{}) error {
	// 使用JSON序列化/反序列化进行深拷贝
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}
