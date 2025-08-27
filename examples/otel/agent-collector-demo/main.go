// Agent-Collector 联合使用演示
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	fmt.Println("🔄 Agent-Collector 联合使用演示")
	fmt.Println("==================================")
	
	// 演示不同的数据路径
	demos := []struct {
		name     string
		endpoint string
		desc     string
	}{
		{
			name:     "直接到Collector",
			endpoint: "127.0.0.1:4318", // Collector HTTP端口
			desc:     "应用 → Collector → SaaS",
		},
		{
			name:     "通过Agent转发",
			endpoint: "127.0.0.1:4328", // Agent HTTP端口
			desc:     "应用 → Agent → Collector → SaaS",
		},
	}
	
	for i, demo := range demos {
		fmt.Printf("\n📋 演示 %d: %s\n", i+1, demo.name)
		fmt.Printf("   数据路径: %s\n", demo.desc)
		
		if err := runDemo(demo.name, demo.endpoint); err != nil {
			log.Printf("❌ 演示失败: %v", err)
			continue
		}
		
		fmt.Printf("   ✅ 数据已通过 %s 发送\n", demo.endpoint)
		time.Sleep(2 * time.Second)
	}
	
	fmt.Println("\n🎯 验证方式:")
	fmt.Println("   1. Jaeger: http://127.0.0.1:16686")
	fmt.Println("   2. OTEL Metrics: http://127.0.0.1:8888/metrics")
	fmt.Println("   3. Agent健康: curl http://127.0.0.1:13134")
	fmt.Println("   4. Collector健康: curl http://127.0.0.1:13133")
}

func runDemo(demoName, endpoint string) error {
	ctx := context.Background()
	
	// 创建追踪导出器
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}
	
	// 创建资源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("agent-collector-demo"),
			semconv.ServiceVersionKey.String("v1.0.0"),
			attribute.String("demo.type", demoName),
			attribute.String("demo.endpoint", endpoint),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %w", err)
	}
	
	// 创建追踪提供器
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(2*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()
	
	// 创建追踪器
	tracer := tp.Tracer("demo-tracer")
	
	// 生成示例追踪
	_, span := tracer.Start(ctx, fmt.Sprintf("demo-operation-%s", demoName))
	span.SetAttributes(
		attribute.String("demo.path", demoName),
		attribute.String("demo.timestamp", time.Now().Format(time.RFC3339)),
	)
	
	// 模拟一些工作
	time.Sleep(100 * time.Millisecond)
	
	span.End()
	
	// 确保数据被导出
	time.Sleep(3 * time.Second)
	
	return nil
}