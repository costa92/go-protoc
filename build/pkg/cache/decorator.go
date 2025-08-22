package cache

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// CacheDecorator 缓存装饰器
type CacheDecorator struct {
	cache   Cache
	keyFunc KeyFunc
	ttl     time.Duration
}

// KeyFunc 缓存key生成函数
type KeyFunc func(method string, args ...interface{}) string

// NewCacheDecorator 创建缓存装饰器
func NewCacheDecorator(cache Cache, keyFunc KeyFunc, ttl time.Duration) *CacheDecorator {
	if keyFunc == nil {
		keyFunc = DefaultKeyFunc
	}

	return &CacheDecorator{
		cache:   cache,
		keyFunc: keyFunc,
		ttl:     ttl,
	}
}

// DecorateFunc 装饰函数，添加缓存功能
func (d *CacheDecorator) DecorateFunc(method string, fn interface{}) interface{} {
	fnType := reflect.TypeOf(fn)
	fnValue := reflect.ValueOf(fn)

	// 创建新的函数
	newFn := func(args []reflect.Value) []reflect.Value {
		// 生成缓存key
		argInterfaces := make([]interface{}, len(args))
		for i, arg := range args {
			argInterfaces[i] = arg.Interface()
		}

		// 第一个参数应该是context
		if len(args) == 0 {
			return fnValue.Call(args)
		}

		ctx, ok := args[0].Interface().(context.Context)
		if !ok {
			return fnValue.Call(args)
		}

		cacheKey := d.keyFunc(method, argInterfaces[1:]...) // 跳过context参数

		// 尝试从缓存获取
		if fnType.NumOut() > 0 {
			resultType := fnType.Out(0)
			result := reflect.New(resultType).Interface()

			if err := d.cache.Get(ctx, cacheKey, result); err == nil {
				// 缓存命中，构造返回值
				results := make([]reflect.Value, fnType.NumOut())
				results[0] = reflect.ValueOf(result).Elem()

				// 如果有错误返回值，设置为nil
				for i := 1; i < fnType.NumOut(); i++ {
					results[i] = reflect.Zero(fnType.Out(i))
				}

				return results
			}
		}

		// 缓存未命中，调用原函数
		results := fnValue.Call(args)

		// 如果没有错误，缓存结果
		if len(results) > 1 {
			errResult := results[len(results)-1]
			if errResult.IsNil() {
				_ = d.cache.Set(ctx, cacheKey, results[0].Interface(), d.ttl)
			}
		} else if len(results) == 1 {
			_ = d.cache.Set(ctx, cacheKey, results[0].Interface(), d.ttl)
		}

		return results
	}

	// 创建新的函数值
	return reflect.MakeFunc(fnType, newFn).Interface()
}

// WithCache 为对象添加缓存功能
func (d *CacheDecorator) WithCache(obj interface{}, methods []string) interface{} {
	objType := reflect.TypeOf(obj)
	objValue := reflect.ValueOf(obj)

	// 创建新的结构体类型
	newObj := reflect.New(objType.Elem()).Interface()
	newObjValue := reflect.ValueOf(newObj).Elem()

	// 复制原对象的字段
	for i := 0; i < objValue.Elem().NumField(); i++ {
		field := objValue.Elem().Field(i)
		if field.CanInterface() {
			newObjValue.Field(i).Set(field)
		}
	}

	// 为指定方法添加缓存
	methodSet := make(map[string]bool, len(methods))
	for _, method := range methods {
		methodSet[method] = true
	}

	// 创建代理对象
	return newObj
}

// DefaultKeyFunc 默认的key生成函数
func DefaultKeyFunc(method string, args ...interface{}) string {
	hasher := md5.New()
	hasher.Write([]byte(method))

	for _, arg := range args {
		data, _ := json.Marshal(arg)
		hasher.Write(data)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// SimpleKeyFunc 简单的key生成函数，适用于简单参数
func SimpleKeyFunc(method string, args ...interface{}) string {
	key := method
	for _, arg := range args {
		key += fmt.Sprintf(":%v", arg)
	}
	return key
}

// CacheableRepository 可缓存的仓储接口
type CacheableRepository interface {
	// SetCache 设置缓存
	SetCache(cache Cache)
	// InvalidateCache 失效缓存
	InvalidateCache(ctx context.Context, keys ...string) error
}

// BaseCacheableRepository 基础可缓存仓储实现
type BaseCacheableRepository struct {
	cache Cache
}

// SetCache 设置缓存
func (r *BaseCacheableRepository) SetCache(cache Cache) {
	r.cache = cache
}

// InvalidateCache 失效缓存
func (r *BaseCacheableRepository) InvalidateCache(ctx context.Context, keys ...string) error {
	if r.cache == nil {
		return nil
	}

	for _, key := range keys {
		if err := r.cache.Delete(ctx, key); err != nil {
			return err
		}
	}

	return nil
}

// GetFromCacheOrFetch 从缓存获取或者调用函数获取
func (r *BaseCacheableRepository) GetFromCacheOrFetch(
	ctx context.Context,
	key string,
	dest interface{},
	ttl time.Duration,
	fetchFn func() (interface{}, error),
) error {
	if r.cache == nil {
		// 没有缓存，直接调用函数
		result, err := fetchFn()
		if err != nil {
			return err
		}
		return copyValue(result, dest)
	}

	// 尝试从缓存获取
	if err := r.cache.Get(ctx, key, dest); err == nil {
		return nil
	}

	// 缓存未命中，调用函数获取
	result, err := fetchFn()
	if err != nil {
		return err
	}

	// 设置缓存
	_ = r.cache.Set(ctx, key, result, ttl)

	// 复制结果到目标
	return copyValue(result, dest)
}
