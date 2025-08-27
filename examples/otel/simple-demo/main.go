// Simple OTEL demo without complex imports
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

func main() {
	fmt.Println("🚀 OTEL Simple Demo")
	fmt.Println("==================")

	// Check Gateway health first
	fmt.Println("📡 Checking OTEL Gateway health...")
	if err := checkGatewayHealth(); err != nil {
		log.Fatalf("Gateway health check failed: %v", err)
	}
	fmt.Println("   ✓ Gateway is healthy")

	// Initialize OTEL tracing
	fmt.Println("\n🔧 Initializing OTEL tracing...")
	shutdown, err := initTracing()
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer func() {
		fmt.Println("\n🔄 Shutting down OTEL...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown OTEL: %v", err)
		} else {
			fmt.Println("   ✓ OTEL shutdown completed")
		}
	}()
	fmt.Println("   ✓ Tracing initialized")

	// Generate sample traces
	fmt.Println("\n📊 Generating sample telemetry data...")
	generateSampleTraces()

	// Wait for data to be exported
	fmt.Println("\n⏳ Waiting for data export...")
	time.Sleep(10 * time.Second)

	fmt.Println("\n✅ Demo completed!")
	printVerificationInstructions()
}

func checkGatewayHealth() error {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://127.0.0.1:13133")
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("gateway returned status: %d", resp.StatusCode)
	}
	return nil
}

func initTracing() (func(context.Context) error, error) {
	ctx := context.Background()

	// Create OTLP HTTP trace exporter
	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("127.0.0.1:4328"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("otel-simple-demo"),
			semconv.ServiceVersionKey.String("v1.0.0"),
			semconv.DeploymentEnvironmentKey.String("demo"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
			sdktrace.WithMaxExportBatchSize(100),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	return tp.Shutdown, nil
}

func generateSampleTraces() {
	tracer := otel.Tracer("otel-simple-demo")

	scenarios := []string{
		"user-login",
		"order-processing",
		"inventory-check",
		"payment-processing",
		"notification-sending",
	}

	for i, scenario := range scenarios {
		fmt.Printf("   📋 Generating trace for: %s\n", scenario)

		ctx := context.Background()
		ctx, rootSpan := tracer.Start(ctx, scenario)

		rootSpan.SetAttributes(
			attribute.String("scenario", scenario),
			attribute.Int("iteration", i+1),
			attribute.String("component", "demo"),
		)

		// Simulate sub-operations
		simulateOperation(ctx, "validate-input", 50*time.Millisecond)
		simulateOperation(ctx, "database-query", 100*time.Millisecond)
		simulateOperation(ctx, "business-logic", 75*time.Millisecond)
		simulateOperation(ctx, "response-formatting", 25*time.Millisecond)

		rootSpan.End()
		time.Sleep(200 * time.Millisecond)
	}
}

func simulateOperation(ctx context.Context, operationName string, duration time.Duration) {
	tracer := otel.Tracer("otel-simple-demo")
	_, span := tracer.Start(ctx, operationName)
	defer span.End()

	span.SetAttributes(
		attribute.String("operation.type", operationName),
		attribute.Int64("duration.ms", duration.Milliseconds()),
	)

	// Simulate work
	time.Sleep(duration)
}

func printVerificationInstructions() {
	fmt.Println("\n🔍 Verification Instructions:")
	fmt.Println("==============================")

	fmt.Println("\n📊 1. Check Gateway Processing:")
	fmt.Println("   docker logs proj-otel-collector --tail 20")
	fmt.Println("   curl http://127.0.0.1:8888/metrics | grep traces")

	fmt.Println("\n🔍 2. View Traces in Jaeger:")
	fmt.Println("   Open: http://127.0.0.1:16686")
	fmt.Println("   Service: otel-simple-demo")
	fmt.Println("   Look for operations: user-login, order-processing, etc.")

	fmt.Println("\n📝 3. Query Logs (if configured):")
	fmt.Println("   Open: http://127.0.0.1:9428/select/vmui/")
	fmt.Println("   Query: service.name:otel-simple-demo")

	fmt.Println("\n🐳 4. Check Container Status:")
	fmt.Println("   docker ps | grep proj-")

	fmt.Println("\n💡 Tips:")
	fmt.Println("   - Wait 1-2 minutes for data to appear in Jaeger")
	fmt.Println("   - Refresh the Jaeger UI if traces don't appear immediately")
	fmt.Println("   - Check Gateway logs for any processing errors")
}
