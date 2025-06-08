// Package tracing 提供了 OpenTelemetry 的初始化功能。
package tracing

import (
	"context"
	"fmt"
	"os"

	"github.com/costa92/go-protoc/pkg/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer 初始化 OpenTelemetry 追踪器并返回一个关闭函数。
func InitTracer(cfg *config.TracingConfig) (*trace.TracerProvider, error) {
	if !cfg.Enabled {
		// 如果追踪被禁用，返回一个no-op的TracerProvider
		return trace.NewTracerProvider(), nil
	}

	var exporter trace.SpanExporter
	var err error

	// 根据配置选择导出器
	switch cfg.Exporter {
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithWriter(os.Stdout),
			stdouttrace.WithPrettyPrint(),
		)
	case "jaeger":
		// 这里使用的是Jaeger的gRPC导出器
		// 您可以根据需要配置更多选项，例如使用HTTP导出器
		exporter, err = jaeger.New(
			jaeger.WithCollectorEndpoint(),
		)
	case "otlp":
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(),
		)
		exporter, err = otlptrace.New(context.Background(), client)
	default:
		return nil, fmt.Errorf("不支持的追踪导出器类型: %s", cfg.Exporter)
	}

	if err != nil {
		return nil, fmt.Errorf("创建追踪导出器失败: %w", err)
	}

	// resource 属性用于标识您的服务。
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}

	// TracerProvider 是 OTel SDK 的核心。
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(res),
	)

	// 设置全局的 TracerProvider。
	otel.SetTracerProvider(tp)

	// 设置全局的 Propagator 以支持 W3C Trace Context 标准。
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	// 返回 TracerProvider 实例
	return tp, nil
}
