package options

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"
)

// MiddlewareOptions 中间件配置选项
type MiddlewareOptions struct {
	// Timeout 超时时间
	Timeout time.Duration `json:"timeout" mapstructure:"timeout"`
	// RateLimit 限流配置
	RateLimit RateLimitOptions `json:"rate_limit" mapstructure:"rate_limit"`
}

// RateLimitOptions 限流配置选项
type RateLimitOptions struct {
	// Enable 是否启用限流
	Enable bool `json:"enable" mapstructure:"enable"`
	// Limit 每秒请求数限制
	Limit float64 `json:"limit" mapstructure:"limit"`
	// Burst 突发请求数限制
	Burst int `json:"burst" mapstructure:"burst"`
	// SkipPaths 跳过限流的路径
	SkipPaths []string `json:"skip_paths" mapstructure:"skip_paths"`
}

// NewMiddlewareOptions 创建中间件配置选项实例
func NewMiddlewareOptions() *MiddlewareOptions {
	return &MiddlewareOptions{
		Timeout: 30 * time.Second,
		RateLimit: RateLimitOptions{
			Enable:    true,
			Limit:     100, // 默认每秒 100 个请求
			Burst:     200, // 默认突发 200 个请求
			SkipPaths: []string{"/healthz", "/metrics"},
		},
	}
}

// Validate 实现 CliOptions 接口，验证配置选项的合法性
func (o *MiddlewareOptions) Validate() []error {
	var errs []error

	if o.Timeout <= 0 {
		errs = append(errs, fmt.Errorf("timeout must be greater than 0"))
	}

	if o.RateLimit.Enable {
		if o.RateLimit.Limit <= 0 {
			errs = append(errs, fmt.Errorf("rate limit must be greater than 0"))
		}
		if o.RateLimit.Burst <= 0 {
			errs = append(errs, fmt.Errorf("burst must be greater than 0"))
		}
		if o.RateLimit.Burst < int(o.RateLimit.Limit) {
			errs = append(errs, fmt.Errorf("burst must be greater than or equal to rate limit"))
		}
	}

	return errs
}

// AddFlags 实现 CliOptions 接口，添加命令行标志
func (o *MiddlewareOptions) AddFlags(fs *pflag.FlagSet) {
	fs.DurationVar(&o.Timeout, "middleware.timeout", o.Timeout, "Global middleware timeout")

	fs.BoolVar(&o.RateLimit.Enable, "middleware.rate-limit.enable", o.RateLimit.Enable, "Enable rate limiting")
	fs.Float64Var(&o.RateLimit.Limit, "middleware.rate-limit.limit", o.RateLimit.Limit, "Number of requests per second")
	fs.IntVar(&o.RateLimit.Burst, "middleware.rate-limit.burst", o.RateLimit.Burst, "Maximum burst size for rate limiting")
	fs.StringSliceVar(&o.RateLimit.SkipPaths, "middleware.rate-limit.skip-paths", o.RateLimit.SkipPaths, "Paths to skip rate limiting")
}
