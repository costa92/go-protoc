package config

import "strings"

// ObservabilityConfig 定义可观察性配置
type ObservabilityConfig struct {
	// SkipPaths 定义不需要记录的路径前缀
	SkipPaths []string
}

// DefaultObservabilityConfig 返回默认的可观察性配置
func DefaultObservabilityConfig() *ObservabilityConfig {
	return &ObservabilityConfig{
		SkipPaths: []string{
			"/metrics",     // Prometheus 指标
			"/debug/",      // Debug 端点
			"/swagger/",    // Swagger UI
			"/healthz",     // 健康检查
			"/favicon.ico", // 浏览器图标请求
		},
	}
}

// ShouldSkip 检查给定路径是否应该跳过观察
func (c *ObservabilityConfig) ShouldSkip(path string) bool {
	for _, skipPath := range c.SkipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}
