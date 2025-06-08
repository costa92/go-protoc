// Package tracing 提供了 OpenTelemetry 的初始化功能。
package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTracer 初始化 OpenTelemetry 追踪器并返回一个关闭函数。
func InitTracer(serviceName string) (*trace.TracerProvider, error) {
	// 为了演示，我们将使用一个标准的 stdout 导出器。
	// 在生产环境中，您应该使用 OTLP 导出器将数据发送到如 Jaeger, Zipkin 或 OpenTelemetry Collector。
	exporter, err := stdouttrace.New(
		stdouttrace.WithWriter(os.Stdout), // 将追踪信息打印到标准输出
		stdouttrace.WithPrettyPrint(),     // 美化输出格式
	)
	if err != nil {
		return nil, err
	}

	// resource 属性用于标识您的服务。
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, err
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

	// 返回关闭函数，以便在程序退出时调用。
	return tp, nil
}
