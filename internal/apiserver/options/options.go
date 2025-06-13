package options

import (
	"github.com/costa92/go-protoc/pkg/logger"
	generOptions "github.com/costa92/go-protoc/pkg/options"
	"github.com/google/wire"
	"github.com/spf13/pflag"
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
		Name:    "apiserver",
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
}

// Validate 校验所有选项。
func (o *Options) Validate() []error {
	var errs []error
	errs = append(errs, o.HTTP.Validate()...)
	errs = append(errs, o.GRPC.Validate()...)
	errs = append(errs, o.Metrics.Validate()...)
	errs = append(errs, o.Tracing.Validate()...)
	errs = append(errs, o.Log.Validate()...)
	return errs
}

// Complete 完成选项配置，从配置文件加载配置
func (o *Options) Complete() error {
	return o.Tracing.Complete()
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

// ProviderSet 是 options 的 Wire 提供者集合
var ProviderSet = wire.NewSet(
	NewOptions,
)
