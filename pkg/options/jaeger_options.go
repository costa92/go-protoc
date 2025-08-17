package options

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/trace"
	"github.com/spf13/pflag"
)

var _ IOptions = (*JaegerOptions)(nil)

// JaegerOptions defines options for Jaeger tracing client.
type JaegerOptions struct {
	// Enabled determines whether Jaeger tracing is enabled
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
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

// TestConnection tests the actual connection to Jaeger agent.
func (o *JaegerOptions) TestConnection() error {
	if !o.Enabled {
		// 如果Jaeger未启用，跳过连接测试
		return nil
	}

	if o.Server == "" {
		return fmt.Errorf("Jaeger server address not configured")
	}

	// 解析服务器地址，支持UDP Agent端点格式
	addr := o.Server
	if !strings.Contains(addr, ":") {
		addr = addr + ":6831" // 默认Jaeger Agent UDP端口
	}

	// 移除可能的协议前缀，因为我们要测试UDP连接
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 测试UDP连接到Jaeger Agent
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to Jaeger agent at %s: %w", addr, err)
	}
	defer conn.Close()

	return nil
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
