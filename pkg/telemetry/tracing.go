package telemetry

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TracerProvider 全局 tracer provider
var TracerProvider *sdktrace.TracerProvider

// Tracer 全局 tracer
var Tracer trace.Tracer

// InitTracer 初始化 tracer
func InitTracer(serviceName, endpoint string) (func(context.Context) error, error) {
	// 检查是否使用 stdout 作为导出器
	useStdout := endpoint == "stdout"

	// 如果不是 stdout 且 endpoint 为空，则不导出
	if endpoint == "" && !useStdout {
		log.Println("Warning: OTLP_ENDPOINT not set, tracing will not be exported")
		// 返回空函数
		return func(context.Context) error { return nil }, nil
	}

	ctx := context.Background()

	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			// 服务名称
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var bsp sdktrace.SpanProcessor

	if useStdout {
		// 使用 stdout 导出器
		stdoutExporter, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
			stdouttrace.WithWriter(os.Stdout),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout exporter: %w", err)
		}
		bsp = sdktrace.NewBatchSpanProcessor(stdoutExporter)
	} else {
		// 连接到 OTLP endpoint
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		conn, err := grpc.DialContext(ctx, endpoint,
			// 注意：在生产环境中应使用 TLS
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
		}

		// 设置 trace exporter
		traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
		if err != nil {
			return nil, fmt.Errorf("failed to create trace exporter: %w", err)
		}

		bsp = sdktrace.NewBatchSpanProcessor(traceExporter)
	}

	// 注册 trace exporter 到 TracerProvider
	TracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(TracerProvider)

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 创建全局 tracer
	Tracer = TracerProvider.Tracer(serviceName)

	// 返回清理函数
	return TracerProvider.Shutdown, nil
}
