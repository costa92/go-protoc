// Copyright 2024 costa92. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package options

import (
	"fmt"
	"os"

	"github.com/costa92/go-protoc/internal/apiserver"
	"github.com/costa92/go-protoc/pkg/logger"
	generOptions "github.com/costa92/go-protoc/pkg/options"
	"github.com/google/wire"
	"github.com/spf13/pflag"
)

var (
	// Name 是编译后的软件名称
	Name = "go-protoc-apiserver"

	// ID 包含主机名和检索过程中遇到的任何错误
	ID, _ = os.Hostname()
)

// Options 是 apiserver 的顶层选项结构。
type Options struct {
	Name       string                       `json:"name" mapstructure:"name"`
	HTTP       *generOptions.HTTPOptions    `json:"http" mapstructure:"http"`
	GRPC       *generOptions.GRPCOptions    `json:"grpc" mapstructure:"grpc"`
	Metrics    *generOptions.MetricsOptions `json:"metrics" mapstructure:"metrics"`
	Tracing    *generOptions.TracingOptions `json:"tracing" mapstructure:"tracing"`
	Log        *logger.LogOptions           `json:"log" mapstructure:"log"`
	Middleware *MiddlewareOptions           `json:"middleware" mapstructure:"middleware"`
}

// MiddlewareOptions 中间件配置
type MiddlewareOptions struct {
	Timeout   string       `json:"timeout" mapstructure:"timeout"`
	CORS      *CORSOptions `json:"cors" mapstructure:"cors"`
	RateLimit *RateLimit   `json:"rate_limit" mapstructure:"rate_limit"`
}

// CORSOptions CORS配置
type CORSOptions struct {
	AllowOrigins     []string `json:"allow_origins" mapstructure:"allow_origins"`
	AllowMethods     []string `json:"allow_methods" mapstructure:"allow_methods"`
	AllowHeaders     []string `json:"allow_headers" mapstructure:"allow_headers"`
	ExposeHeaders    []string `json:"expose_headers" mapstructure:"expose_headers"`
	AllowCredentials bool     `json:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           string   `json:"max_age" mapstructure:"max_age"`
}

// RateLimit 限流配置
type RateLimit struct {
	Enable bool   `json:"enable" mapstructure:"enable"`
	Limit  int    `json:"limit" mapstructure:"limit"`
	Burst  int    `json:"burst" mapstructure:"burst"`
	Window string `json:"window" mapstructure:"window"`
}

// NewOptions 创建一个带有完整默认值的顶层 Options。
func NewOptions() *Options {
	return &Options{
		Name:    Name,
		HTTP:    generOptions.NewHTTPOptions(),
		GRPC:    generOptions.NewGRPCOptions(),
		Metrics: generOptions.NewMetricsOptions(),
		Tracing: generOptions.NewTracingOptions(),
		Log:     logger.NewLogOptions(),
		Middleware: &MiddlewareOptions{
			Timeout: "30s",
			CORS: &CORSOptions{
				AllowOrigins:     []string{"*"},
				AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD"},
				AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID", "X-Real-IP"},
				ExposeHeaders:    []string{},
				AllowCredentials: true,
				MaxAge:           "12h",
			},
			RateLimit: &RateLimit{
				Enable: true,
				Limit:  100,
				Burst:  200,
				Window: "1m",
			},
		},
	}
}

// AddFlags 将所有组件的标志添加到指定的 FlagSet。
func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.Name, "name", o.Name, "服务器名称")
	o.HTTP.AddFlags(fs)
	o.GRPC.AddFlags(fs)
	o.Metrics.AddFlags(fs)
	o.Tracing.AddFlags(fs)
	o.Log.AddFlags(fs)

	// 添加中间件相关的标志
	if o.Middleware != nil {
		fs.StringVar(&o.Middleware.Timeout, "middleware.timeout", o.Middleware.Timeout, "中间件超时时间")

		if o.Middleware.RateLimit != nil {
			fs.BoolVar(&o.Middleware.RateLimit.Enable, "middleware.rate-limit.enable", o.Middleware.RateLimit.Enable, "是否启用限流")
			fs.IntVar(&o.Middleware.RateLimit.Limit, "middleware.rate-limit.limit", o.Middleware.RateLimit.Limit, "限流速率")
			fs.IntVar(&o.Middleware.RateLimit.Burst, "middleware.rate-limit.burst", o.Middleware.RateLimit.Burst, "限流突发大小")
			fs.StringVar(&o.Middleware.RateLimit.Window, "middleware.rate-limit.window", o.Middleware.RateLimit.Window, "限流窗口大小")
		}
	}
}

// Validate 校验所有选项。
func (o *Options) Validate() []error {
	var errs []error

	// 验证基本配置
	if o.Name == "" {
		errs = append(errs, fmt.Errorf("服务名称不能为空"))
	}

	// 验证各个组件的配置
	errs = append(errs, o.HTTP.Validate()...)
	errs = append(errs, o.GRPC.Validate()...)
	errs = append(errs, o.Metrics.Validate()...)
	errs = append(errs, o.Tracing.Validate()...)
	errs = append(errs, o.Log.Validate()...)

	// 验证中间件配置
	if o.Middleware != nil {
		if o.Middleware.RateLimit != nil && o.Middleware.RateLimit.Enable {
			if o.Middleware.RateLimit.Limit <= 0 {
				errs = append(errs, fmt.Errorf("限流 limit 必须大于 0"))
			}
			if o.Middleware.RateLimit.Burst <= 0 {
				errs = append(errs, fmt.Errorf("限流 burst 必须大于 0"))
			}
			if o.Middleware.RateLimit.Window == "" {
				errs = append(errs, fmt.Errorf("限流 window 不能为空"))
			}
		}
	}

	return errs
}

// Complete 完成选项配置，从配置文件加载配置
func (o *Options) Complete() error {
	if err := o.Tracing.Complete(); err != nil {
		return err
	}

	// 确保服务名称已设置
	if o.Name == "" {
		o.Name = Name
	}

	return nil
}

// GetGRPCOptions 获取 gRPC 选项
func (o *Options) GetGRPCOptions() *generOptions.GRPCOptions {
	return o.GRPC
}

// GetHTTPOptions 获取 HTTP 选项
func (o *Options) GetHTTPOptions() *generOptions.HTTPOptions {
	return o.HTTP
}

// GetLogOptions 获取日志选项
func (o *Options) GetLogOptions() *logger.LogOptions {
	return o.Log
}

// GetMiddlewareOptions 获取中间件选项
func (o *Options) GetMiddlewareOptions() *MiddlewareOptions {
	return o.Middleware
}

// ProviderSet 是 options 的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewOptions,
)

func (o *Options) ApplyTo(c *apiserver.Config) error {
	c.GRPCOptions = o.GRPC
	c.HTTPOptions = o.HTTP
	return nil
}

// Config 返回 apiserver.Config 的实例
func (o *Options) Config() (*apiserver.Config, error) {
	c := &apiserver.Config{}
	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}
	return c, nil
}
