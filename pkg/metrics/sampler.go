package metrics

import (
	"math/rand"
	"sync/atomic"
	"time"
)

// SamplingConfig 采样配置
type SamplingConfig struct {
	// HTTPSampleRate HTTP请求采样率 (0.0-1.0)
	HTTPSampleRate float64 `json:"http_sample_rate" yaml:"http_sample_rate"`

	// GRPCSampleRate gRPC调用采样率 (0.0-1.0)
	GRPCSampleRate float64 `json:"grpc_sample_rate" yaml:"grpc_sample_rate"`

	// DatabaseSampleRate 数据库操作采样率 (0.0-1.0)
	DatabaseSampleRate float64 `json:"database_sample_rate" yaml:"database_sample_rate"`

	// RedisSampleRate Redis操作采样率 (0.0-1.0)
	RedisSampleRate float64 `json:"redis_sample_rate" yaml:"redis_sample_rate"`

	// ErrorSampleRate 错误采样率 (0.0-1.0，错误通常需要更高采样率)
	ErrorSampleRate float64 `json:"error_sample_rate" yaml:"error_sample_rate"`

	// SlowQuerySampleRate 慢查询采样率 (0.0-1.0)
	SlowQuerySampleRate float64 `json:"slow_query_sample_rate" yaml:"slow_query_sample_rate"`

	// HighVolumeOperationRate 高频操作采样率 (0.0-1.0)
	HighVolumeOperationRate float64 `json:"high_volume_operation_rate" yaml:"high_volume_operation_rate"`
}

// DefaultSamplingConfig 默认采样配置
func DefaultSamplingConfig() *SamplingConfig {
	return &SamplingConfig{
		HTTPSampleRate:          0.1,  // 10% HTTP请求
		GRPCSampleRate:          0.1,  // 10% gRPC调用
		DatabaseSampleRate:      0.05, // 5% 数据库操作
		RedisSampleRate:         0.02, // 2% Redis操作
		ErrorSampleRate:         1.0,  // 100% 错误记录
		SlowQuerySampleRate:     1.0,  // 100% 慢查询
		HighVolumeOperationRate: 0.01, // 1% 高频操作
	}
}

// ProductionSamplingConfig 生产环境采样配置
func ProductionSamplingConfig() *SamplingConfig {
	return &SamplingConfig{
		HTTPSampleRate:          0.01,  // 1% HTTP请求
		GRPCSampleRate:          0.01,  // 1% gRPC调用
		DatabaseSampleRate:      0.005, // 0.5% 数据库操作
		RedisSampleRate:         0.001, // 0.1% Redis操作
		ErrorSampleRate:         1.0,   // 100% 错误记录
		SlowQuerySampleRate:     1.0,   // 100% 慢查询
		HighVolumeOperationRate: 0.001, // 0.1% 高频操作
	}
}

// DevelopmentSamplingConfig 开发环境采样配置
func DevelopmentSamplingConfig() *SamplingConfig {
	return &SamplingConfig{
		HTTPSampleRate:          1.0, // 100% HTTP请求
		GRPCSampleRate:          1.0, // 100% gRPC调用
		DatabaseSampleRate:      1.0, // 100% 数据库操作
		RedisSampleRate:         1.0, // 100% Redis操作
		ErrorSampleRate:         1.0, // 100% 错误记录
		SlowQuerySampleRate:     1.0, // 100% 慢查询
		HighVolumeOperationRate: 1.0, // 100% 高频操作
	}
}

// Sampler 采样器接口
type Sampler interface {
	// ShouldSample 是否应该采样
	ShouldSample(sampleType SampleType) bool
	// ShouldSampleWithRate 使用指定采样率采样
	ShouldSampleWithRate(rate float64) bool
	// UpdateConfig 更新采样配置
	UpdateConfig(config *SamplingConfig)
	// GetConfig 获取当前采样配置
	GetConfig() *SamplingConfig
}

// SampleType 采样类型
type SampleType int

const (
	SampleTypeHTTP SampleType = iota
	SampleTypeGRPC
	SampleTypeDatabase
	SampleTypeRedis
	SampleTypeError
	SampleTypeSlowQuery
	SampleTypeHighVolumeOperation
)

// DefaultSampler 默认采样器实现
type DefaultSampler struct {
	config atomic.Value // *SamplingConfig
	rng    *rand.Rand
}

// NewDefaultSampler 创建默认采样器
func NewDefaultSampler(config *SamplingConfig) *DefaultSampler {
	if config == nil {
		config = DefaultSamplingConfig()
	}

	sampler := &DefaultSampler{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	sampler.config.Store(config)
	return sampler
}

// NewProductionSampler 创建生产环境采样器
func NewProductionSampler() *DefaultSampler {
	return NewDefaultSampler(ProductionSamplingConfig())
}

// NewDevelopmentSampler 创建开发环境采样器
func NewDevelopmentSampler() *DefaultSampler {
	return NewDefaultSampler(DevelopmentSamplingConfig())
}

// ShouldSample 是否应该采样
func (s *DefaultSampler) ShouldSample(sampleType SampleType) bool {
	config := s.GetConfig()

	var rate float64
	switch sampleType {
	case SampleTypeHTTP:
		rate = config.HTTPSampleRate
	case SampleTypeGRPC:
		rate = config.GRPCSampleRate
	case SampleTypeDatabase:
		rate = config.DatabaseSampleRate
	case SampleTypeRedis:
		rate = config.RedisSampleRate
	case SampleTypeError:
		rate = config.ErrorSampleRate
	case SampleTypeSlowQuery:
		rate = config.SlowQuerySampleRate
	case SampleTypeHighVolumeOperation:
		rate = config.HighVolumeOperationRate
	default:
		rate = 1.0 // 未知类型默认采样
	}

	return s.ShouldSampleWithRate(rate)
}

// ShouldSampleWithRate 使用指定采样率采样
func (s *DefaultSampler) ShouldSampleWithRate(rate float64) bool {
	if rate <= 0.0 {
		return false
	}
	if rate >= 1.0 {
		return true
	}

	return s.rng.Float64() < rate
}

// UpdateConfig 更新采样配置
func (s *DefaultSampler) UpdateConfig(config *SamplingConfig) {
	if config != nil {
		s.config.Store(config)
	}
}

// GetConfig 获取当前采样配置
func (s *DefaultSampler) GetConfig() *SamplingConfig {
	return s.config.Load().(*SamplingConfig)
}

// AdaptiveSampler 自适应采样器
type AdaptiveSampler struct {
	*DefaultSampler

	// 统计信息
	requestCount  atomic.Int64
	lastResetTime atomic.Int64

	// 自适应配置
	baseConfig     *SamplingConfig
	maxRequestRate int64 // 最大请求率（每秒）
	adaptiveMode   bool
}

// NewAdaptiveSampler 创建自适应采样器
func NewAdaptiveSampler(baseConfig *SamplingConfig, maxRequestRate int64) *AdaptiveSampler {
	if baseConfig == nil {
		baseConfig = DefaultSamplingConfig()
	}

	sampler := &AdaptiveSampler{
		DefaultSampler: NewDefaultSampler(baseConfig),
		baseConfig:     baseConfig,
		maxRequestRate: maxRequestRate,
		adaptiveMode:   true,
	}

	sampler.lastResetTime.Store(time.Now().Unix())

	// 启动自适应调整
	go sampler.startAdaptiveAdjustment()

	return sampler
}

// ShouldSample 自适应采样
func (s *AdaptiveSampler) ShouldSample(sampleType SampleType) bool {
	// 增加请求计数
	s.requestCount.Add(1)

	return s.DefaultSampler.ShouldSample(sampleType)
}

// startAdaptiveAdjustment 启动自适应调整
func (s *AdaptiveSampler) startAdaptiveAdjustment() {
	ticker := time.NewTicker(10 * time.Second) // 每10秒调整一次
	defer ticker.Stop()

	for range ticker.C {
		if s.adaptiveMode {
			s.adjustSamplingRates()
		}
	}
}

// adjustSamplingRates 调整采样率
func (s *AdaptiveSampler) adjustSamplingRates() {
	now := time.Now().Unix()
	lastReset := s.lastResetTime.Load()
	duration := now - lastReset

	if duration <= 0 {
		return
	}

	// 计算当前请求率
	currentCount := s.requestCount.Swap(0)
	currentRate := currentCount / duration
	s.lastResetTime.Store(now)

	// 根据请求率调整采样率
	currentConfig := s.GetConfig()
	newConfig := *currentConfig // 复制当前配置

	if currentRate > s.maxRequestRate {
		// 请求率过高，降低采样率
		factor := float64(s.maxRequestRate) / float64(currentRate)
		newConfig.HTTPSampleRate = adjustRate(currentConfig.HTTPSampleRate, factor, 0.001, 1.0)
		newConfig.GRPCSampleRate = adjustRate(currentConfig.GRPCSampleRate, factor, 0.001, 1.0)
		newConfig.DatabaseSampleRate = adjustRate(currentConfig.DatabaseSampleRate, factor, 0.0001, 1.0)
		newConfig.RedisSampleRate = adjustRate(currentConfig.RedisSampleRate, factor, 0.0001, 1.0)
		// 错误和慢查询保持高采样率
	} else if currentRate < s.maxRequestRate/2 {
		// 请求率较低，可以适当提高采样率
		factor := 1.2 // 缓慢增加
		newConfig.HTTPSampleRate = adjustRate(currentConfig.HTTPSampleRate, factor, 0.001, s.baseConfig.HTTPSampleRate)
		newConfig.GRPCSampleRate = adjustRate(currentConfig.GRPCSampleRate, factor, 0.001, s.baseConfig.GRPCSampleRate)
		newConfig.DatabaseSampleRate = adjustRate(currentConfig.DatabaseSampleRate, factor, 0.0001, s.baseConfig.DatabaseSampleRate)
		newConfig.RedisSampleRate = adjustRate(currentConfig.RedisSampleRate, factor, 0.0001, s.baseConfig.RedisSampleRate)
	}

	s.UpdateConfig(&newConfig)
}

// adjustRate 调整采样率
func adjustRate(current, factor, min, max float64) float64 {
	newRate := current * factor
	if newRate < min {
		return min
	}
	if newRate > max {
		return max
	}
	return newRate
}

// EnableAdaptiveMode 启用自适应模式
func (s *AdaptiveSampler) EnableAdaptiveMode() {
	s.adaptiveMode = true
}

// DisableAdaptiveMode 禁用自适应模式
func (s *AdaptiveSampler) DisableAdaptiveMode() {
	s.adaptiveMode = false
	// 恢复基础配置
	s.UpdateConfig(s.baseConfig)
}

// GetStats 获取统计信息
func (s *AdaptiveSampler) GetStats() map[string]interface{} {
	now := time.Now().Unix()
	lastReset := s.lastResetTime.Load()
	duration := now - lastReset
	currentCount := s.requestCount.Load()

	var currentRate int64
	if duration > 0 {
		currentRate = currentCount / duration
	}

	return map[string]interface{}{
		"adaptive_mode":  s.adaptiveMode,
		"current_rate":   currentRate,
		"max_rate":       s.maxRequestRate,
		"request_count":  currentCount,
		"duration":       duration,
		"current_config": s.GetConfig(),
		"base_config":    s.baseConfig,
	}
}
