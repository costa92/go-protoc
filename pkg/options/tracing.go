package options

import (
	"context"
	"fmt"

	"github.com/costa92/go-protoc/pkg/telemetry"
	"github.com/spf13/pflag"
)

// TracerShutdownFunc 用于在应用程序关闭时关闭 tracer
var TracerShutdownFunc func(context.Context) error

type TracingOptions struct {
	ServiceName  string `json:"service_name"    mapstructure:"service_name"`
	Enabled      bool   `json:"enabled"         mapstructure:"enabled"`
	Exporter     string `json:"exporter"        mapstructure:"exporter"`
	OTLPEndpoint string `json:"otlp_endpoint"  mapstructure:"otlp_endpoint"`
}

func NewTracingOptions() *TracingOptions {
	return &TracingOptions{
		ServiceName:  "go-protoc-service",
		Enabled:      false,
		Exporter:     "stdout",
		OTLPEndpoint: "localhost:4317",
	}
}

func (o *TracingOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.ServiceName, "tracing.service-name", o.ServiceName, "服务名称")
	fs.BoolVar(&o.Enabled, "tracing.enabled", o.Enabled, "是否启用链路追踪")
	fs.StringVar(&o.Exporter, "tracing.exporter", o.Exporter, "导出器类型 (stdout, jaeger, otlp)")
	fs.StringVar(&o.OTLPEndpoint, "tracing.otlp-endpoint", o.OTLPEndpoint, "OTLP导出器端点")
}

func (o *TracingOptions) Validate() []error {
	var errs []error
	if o.ServiceName == "" {
		errs = append(errs, fmt.Errorf("tracing service name cannot be empty"))
	}
	return errs
}

func (o *TracingOptions) Complete() error {
	// 如果未启用，则跳过初始化
	if !o.Enabled {
		return nil
	}

	// 根据导出器类型设置端点
	endpoint := o.OTLPEndpoint
	if o.Exporter == "stdout" {
		endpoint = "stdout"
	} else if o.Exporter == "jaeger" {
		// 可以设置默认的 Jaeger 端点
		// 这里暂时使用 OTLP 端点
	}

	// 初始化 OpenTelemetry Tracer
	shutdown, err := telemetry.InitTracer(o.ServiceName, endpoint)
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// 保存 shutdown 函数以便在应用程序关闭时使用
	TracerShutdownFunc = shutdown

	fmt.Println("Tracer initialized successfully, shutdown function registered")

	return nil
}
