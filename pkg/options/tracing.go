package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

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
	return nil
}
