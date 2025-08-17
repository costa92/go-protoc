package trace

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	globalTracer trace.Tracer
	mu           sync.RWMutex
	initialized  bool
)

// TracerConfig holds the configuration for the tracer.
type TracerConfig struct {
	ServiceName string
	Endpoint    string
	Environment string
}

// InitializeGlobalTracer initializes the global tracer with the given configuration.
// This should be called once during application startup.
func InitializeGlobalTracer(cfg TracerConfig) error {
	mu.Lock()
	defer mu.Unlock()

	if initialized {
		return fmt.Errorf("global tracer already initialized")
	}

	// 添加调试日志
	fmt.Printf("[Trace] 正在初始化全局追踪器...\n")
	fmt.Printf("[Trace] 服务名称: %s\n", cfg.ServiceName)
	fmt.Printf("[Trace] 使用 Agent 端点: 127.0.0.1:6831\n")
	fmt.Printf("[Trace] 环境: %s\n", cfg.Environment)

	// Create Jaeger exporter using agent endpoint (UDP)
	exporter, err := jaeger.New(jaeger.WithAgentEndpoint(
		jaeger.WithAgentHost("127.0.0.1"),
		jaeger.WithAgentPort("6831"),
	))
	if err != nil {
		return fmt.Errorf("创建 Jaeger 导出器失败: %w", err)
	}

	// Create resource
	res, err := resource.New(context.Background(), resource.WithAttributes(
		semconv.ServiceNameKey.String(cfg.ServiceName),
		attribute.String("env", cfg.Environment),
		attribute.String("exporter", "jaeger"),
	))
	if err != nil {
		return fmt.Errorf("创建资源失败: %w", err)
	}

	// Create TracerProvider
	bsp := tracesdk.NewBatchSpanProcessor(exporter)
	tp := tracesdk.NewTracerProvider(
		// Set the sampling rate to 100%
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(1.0))),
		// Always be sure to batch in production
		tracesdk.WithSpanProcessor(bsp),
		// Record information about this application in a Resource
		tracesdk.WithResource(res),
	)

	// Set global TracerProvider
	otel.SetTracerProvider(tp)

	// Create global tracer
	globalTracer = otel.Tracer(cfg.ServiceName)
	initialized = true

	fmt.Printf("[Trace] 全局追踪器初始化成功! 采样率: 100%%\n")
	fmt.Printf("[Trace] 使用 UDP Agent 协议发送到 Jaeger\n")
	fmt.Printf("[Trace] 请访问 Jaeger UI: http://localhost:16686\n")

	return nil
}

// GetTracer returns a tracer for the given component.
// If name is empty, it returns the global tracer.
func GetTracer(name string) trace.Tracer {
	mu.RLock()
	defer mu.RUnlock()

	if !initialized {
		fmt.Printf("[Trace] 警告: 全局追踪器未初始化，返回空追踪器\n")
		return trace.NewNoopTracerProvider().Tracer("")
	}

	if name == "" {
		return globalTracer
	}

	return otel.Tracer(name)
}

// IsInitialized returns whether the global tracer has been initialized.
func IsInitialized() bool {
	mu.RLock()
	defer mu.RUnlock()
	return initialized
}

// StartSpan starts a new span with the given name and options.
// It uses the global tracer if available, otherwise returns a noop span.
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := GetTracer("")
	return tracer.Start(ctx, name, opts...)
}

// WithAttributes is a helper function to create span attributes.
func WithAttributes(attrs ...attribute.KeyValue) trace.SpanStartOption {
	return trace.WithAttributes(attrs...)
}

// WithSpanKind is a helper function to set span kind.
func WithSpanKind(kind trace.SpanKind) trace.SpanStartOption {
	return trace.WithSpanKind(kind)
}

// SetNoOpTracer sets a no-op tracer provider for when tracing is disabled.
func SetNoOpTracer() {
	mu.Lock()
	defer mu.Unlock()

	// 设置空操作追踪器
	otel.SetTracerProvider(trace.NewNoopTracerProvider())
	globalTracer = trace.NewNoopTracerProvider().Tracer("")
	initialized = true

	fmt.Printf("[Trace] Jaeger 追踪已禁用，使用空操作追踪器\n")
}

// Shutdown gracefully shuts down the tracer provider.
// This should be called during application shutdown.
func Shutdown(ctx context.Context) error {
	if tp, ok := otel.GetTracerProvider().(*tracesdk.TracerProvider); ok {
		return tp.Shutdown(ctx)
	}
	return nil
}
