package options

import (
	"strings"

	"github.com/costa92/go-protoc/v2/pkg/trace"
	"github.com/spf13/pflag"
)

var _ IOptions = (*JaegerOptions)(nil)

// JaegerOptions defines options for Jaeger tracing client.
type JaegerOptions struct {
	// Enabled determines whether Jaeger tracing is enabled
	Enabled     bool   `json:"enabled,omitempty" mapstructure:"enabled"`
	// Server is the url of the Jaeger server
	Server      string `json:"server,omitempty" mapstructure:"server"`
	ServiceName string `json:"service-name,omitempty" mapstructure:"service-name"`
	Env         string `json:"env,omitempty" mapstructure:"env"`
}

// NewJaegerOptions create a `zero` value instance.
func NewJaegerOptions() *JaegerOptions {
	return &JaegerOptions{
		Enabled: false, // 默认关闭，需要显式开启
		Server:  "127.0.0.1:6831",
		Env:     "dev",
	}
}

// Validate verifies flags passed to JaegerOptions.
func (o *JaegerOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to Jaeger tracing for a specific APIServer to the specified FlagSet.
func (o *JaegerOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.BoolVar(&o.Enabled, "jaeger.enabled", o.Enabled, ""+
		"Enable or disable Jaeger tracing.")
	fs.StringVar(&o.Server, "jaeger.server", o.Server, ""+
		"Server is the url of the Jaeger server.")
	fs.StringVar(&o.ServiceName, "jaeger.service-name", o.ServiceName, ""+
		"Specify the service name for jaeger resource.")
	fs.StringVar(&o.Env, "jaeger.env", o.Env, "Specify the deployment environment(dev/test/staging/prod).")
}

func (o *JaegerOptions) SetTracerProvider(serviceName string) error {
	// 如果未启用追踪，直接返回
	if !o.Enabled {
		trace.SetNoOpTracer()
		return nil
	}

	// 构造正确的 OTLP 端点 URL
	otlpURL := o.Server
	if !strings.HasPrefix(otlpURL, "http://") && !strings.HasPrefix(otlpURL, "https://") {
		otlpURL = "http://" + otlpURL
	}

	// 使用统一的追踪管理器初始化
	cfg := trace.TracerConfig{
		ServiceName: serviceName,
		Endpoint:    otlpURL,
		Environment: o.Env,
	}

	return trace.InitializeGlobalTracer(cfg)
}
