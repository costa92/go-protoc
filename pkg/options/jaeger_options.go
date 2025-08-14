package options

import (
	"strings"

	"github.com/costa92/go-protoc/v2/pkg/trace"
	"github.com/spf13/pflag"
)

var _ IOptions = (*JaegerOptions)(nil)

// JaegerOptions defines options for consul client.
type JaegerOptions struct {
	// Server is the url of the Jaeger server
	Server      string `json:"server,omitempty" mapstructure:"server"`
	ServiceName string `json:"service-name,omitempty" mapstructure:"service-name"`
	Env         string `json:"env,omitempty" mapstructure:"env"`
}

// NewJaegerOptions create a `zero` value instance.
func NewJaegerOptions() *JaegerOptions {
	return &JaegerOptions{
		Server: "http://127.0.0.1:14268/api/traces",
		Env:    "dev",
	}
}

// Validate verifies flags passed to JaegerOptions.
func (o *JaegerOptions) Validate() []error {
	errs := []error{}

	return errs
}

// AddFlags adds flags related to mysql storage for a specific APIServer to the specified FlagSet.
func (o *JaegerOptions) AddFlags(fs *pflag.FlagSet, prefixes ...string) {
	fs.StringVar(&o.Server, "jaeger.server", o.Server, ""+
		"Server is the url of the Jaeger server.")
	fs.StringVar(&o.ServiceName, "jaeger.service-name", o.ServiceName, ""+
		"Specify the service name for jaeger resource.")
	fs.StringVar(&o.Env, "jaeger.env", o.Env, "Specify the deployment environment(dev/test/staging/prod).")
}

func (o *JaegerOptions) SetTracerProvider(serviceName string) error {
	// 构造正确的 Jaeger 收集器端点 URL
	jaegerURL := o.Server
	if !strings.HasPrefix(jaegerURL, "http://") && !strings.HasPrefix(jaegerURL, "https://") {
		jaegerURL = "http://" + jaegerURL
	}
	if !strings.HasSuffix(jaegerURL, "/api/traces") {
		jaegerURL = jaegerURL + "/api/traces"
	}

	// 使用统一的追踪管理器初始化
	cfg := trace.TracerConfig{
		ServiceName: serviceName,
		Endpoint:    jaegerURL,
		Environment: o.Env,
	}

	return trace.InitializeGlobalTracer(cfg)
}
