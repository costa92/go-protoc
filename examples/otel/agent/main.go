// Package main provides OpenTelemetry Agent implementation examples
// demonstrating the Agent -> Gateway -> SaaS architecture pattern
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// OTELAgent represents the Agent layer in OTEL architecture
// It collects telemetry data from the application and forwards to Gateway
type OTELAgent struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	logger         *slog.Logger
	serviceName    string
	serviceVersion string
	environment    string
}

// OTELAgentConfig holds configuration for OTEL Agent
type OTELAgentConfig struct {
	ServiceName         string
	ServiceVersion      string
	Environment         string
	GatewayHTTPEndpoint string // OTEL Gateway HTTP endpoint (4328)
	GatewayGRPCEndpoint string // OTEL Gateway gRPC endpoint (4327)
	EnableLogs          bool
	EnableMetrics       bool
	EnableTraces        bool
}

// NewOTELAgent creates a new OTEL Agent instance
func NewOTELAgent(config OTELAgentConfig) (*OTELAgent, error) {
	agent := &OTELAgent{
		serviceName:    config.ServiceName,
		serviceVersion: config.ServiceVersion,
		environment:    config.Environment,
	}

	ctx := context.Background()

	// Create shared resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
			semconv.ServiceVersionKey.String(config.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(config.Environment),
			semconv.TelemetrySDKLanguageGo,
			semconv.TelemetrySDKVersionKey.String("1.21.0"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize Traces (if enabled)
	if config.EnableTraces {
		if err := agent.initTraces(ctx, config.GatewayHTTPEndpoint, res); err != nil {
			return nil, fmt.Errorf("failed to initialize traces: %w", err)
		}
	}

	// Initialize Metrics (if enabled)
	if config.EnableMetrics {
		if err := agent.initMetrics(ctx, config.GatewayHTTPEndpoint, res); err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
	}

	// Initialize Logs (if enabled)
	if config.EnableLogs {
		if err := agent.initLogs(ctx, config.GatewayHTTPEndpoint, res); err != nil {
			return nil, fmt.Errorf("failed to initialize logs: %w", err)
		}
	}

	return agent, nil
}

// initTraces initializes OpenTelemetry tracing
func (a *OTELAgent) initTraces(ctx context.Context, endpoint string, res *resource.Resource) error {
	// Create OTLP HTTP trace exporter
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	a.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter,
			sdktrace.WithBatchTimeout(time.Second*5),
			sdktrace.WithMaxExportBatchSize(100),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.TraceIDRatioBased(0.1)), // 10% sampling
	)

	// Set global tracer provider
	otel.SetTracerProvider(a.tracerProvider)

	return nil
}

// initMetrics initializes OpenTelemetry metrics
func (a *OTELAgent) initMetrics(ctx context.Context, endpoint string, res *resource.Resource) error {
	// Create OTLP HTTP metric exporter
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint(endpoint),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	// Create meter provider
	a.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(30*time.Second))),
		sdkmetric.WithResource(res),
	)

	// Set global meter provider
	otel.SetMeterProvider(a.meterProvider)

	return nil
}

// initLogs initializes file-based logging (for OTEL Collector filelog receiver)
func (a *OTELAgent) initLogs(ctx context.Context, endpoint string, res *resource.Resource) error {
	// Create log file for OTEL Collector
	logDir := "/home/hellotalk/code/go/src/github.com/costa92/go-protoc/logs/agent-demo"
	logFile := logDir + "/agent.log"
	
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	
	// Create multi-writer: both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	
	// Create structured logger
	a.logger = slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	
	fmt.Printf("📝 Agent logs writing to: %s\n", logFile)
	return nil
}

// Shutdown gracefully shuts down the OTEL Agent
func (a *OTELAgent) Shutdown(ctx context.Context) error {
	var errors []error

	if a.tracerProvider != nil {
		if err := a.tracerProvider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
	}

	if a.meterProvider != nil {
		if err := a.meterProvider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown meter provider: %w", err))
		}
	}

	// Logger cleanup is handled automatically (file-based logging)

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// ExampleOTELAgentUsage demonstrates how to use the OTEL Agent
func ExampleOTELAgentUsage() {
	fmt.Println("=== OTEL Agent Example ===")

	// Create OTEL Agent configuration
	config := OTELAgentConfig{
		ServiceName:         "example-app",
		ServiceVersion:      "v1.0.0",
		Environment:         "development",
		GatewayHTTPEndpoint: "127.0.0.1:4328", // OTEL Gateway HTTP endpoint
		GatewayGRPCEndpoint: "127.0.0.1:4327", // OTEL Gateway gRPC endpoint
		EnableLogs:          true,
		EnableMetrics:       true,
		EnableTraces:        true,
	}

	// Create and initialize OTEL Agent
	agent, err := NewOTELAgent(config)
	if err != nil {
		log.Fatalf("Failed to create OTEL Agent: %v", err)
	}

	// Ensure proper shutdown
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := agent.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown OTEL Agent: %v", err)
		}
	}()

	fmt.Printf("✅ OTEL Agent initialized successfully\n")
	fmt.Printf("   Service: %s v%s (%s)\n", config.ServiceName, config.ServiceVersion, config.Environment)
	fmt.Printf("   Gateway: %s\n", config.GatewayHTTPEndpoint)

	// Demonstrate telemetry data generation
	demonstrateTracing()
	demonstrateMetrics()
	demonstrateLogs(agent)

	fmt.Println("\n🎯 Data sent to Gateway -> SaaS platforms")
	fmt.Println("   Check OTEL Collector logs for processing status")
	fmt.Println("   View data in configured SaaS platforms (VictoriaLogs, Jaeger, etc.)")
}

// demonstrateTracing shows how to create traces
func demonstrateTracing() {
	fmt.Println("\n📊 Generating Trace Data...")

	ctx := context.Background()
	tracer := otel.Tracer("example-tracer")

	// Create a root span
	ctx, rootSpan := tracer.Start(ctx, "example-operation")
	defer rootSpan.End()

	rootSpan.SetAttributes(
		attribute.String("operation.type", "demo"),
		attribute.String("user.id", "user-12345"),
	)

	// Simulate some work with child spans
	simulateDatabaseOperation(ctx)
	simulateHTTPRequest(ctx)

	fmt.Println("   ✓ Root span with 2 child spans created")
}

// simulateDatabaseOperation creates a database operation span
func simulateDatabaseOperation(ctx context.Context) {
	tracer := otel.Tracer("example-tracer")
	_, span := tracer.Start(ctx, "database-query")
	defer span.End()

	span.SetAttributes(
		attribute.String("db.system", "mysql"),
		attribute.String("db.operation", "SELECT"),
		attribute.String("db.table", "users"),
	)

	// Simulate database work
	time.Sleep(50 * time.Millisecond)
}

// simulateHTTPRequest creates an HTTP request span
func simulateHTTPRequest(ctx context.Context) {
	tracer := otel.Tracer("example-tracer")
	_, span := tracer.Start(ctx, "http-request")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", "/api/users"),
		attribute.Int("http.status_code", 200),
	)

	// Simulate HTTP request
	time.Sleep(30 * time.Millisecond)
}

// demonstrateMetrics shows how to create metrics
func demonstrateMetrics() {
	fmt.Println("\n📈 Generating Metric Data...")

	meter := otel.Meter("example-meter")

	// Create counter metric
	requestCounter, err := meter.Int64Counter(
		"http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	if err != nil {
		log.Printf("Failed to create counter: %v", err)
		return
	}

	// Create histogram metric
	requestDuration, err := meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
	)
	if err != nil {
		log.Printf("Failed to create histogram: %v", err)
		return
	}

	// Generate some metric data
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		requestCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "GET"),
				attribute.String("status", "200"),
			),
		)

		requestDuration.Record(ctx, float64(i*10+50)/1000,
			metric.WithAttributes(
				attribute.String("method", "GET"),
				attribute.String("endpoint", "/api/users"),
			),
		)
	}

	fmt.Println("   ✓ Counter and histogram metrics recorded")
}

// demonstrateLogs shows how to create structured logs using agent logger
func demonstrateLogs(agent *OTELAgent) {
	fmt.Println("\n📝 Generating Log Data...")

	if agent.logger == nil {
		fmt.Println("   ⚠️ Logger not initialized - skipping log generation")
		return
	}

	// Generate log entries using agent logger
	ctx := context.Background()

	agent.logger.InfoContext(ctx, "Agent operation successful",
		"service", agent.serviceName,
		"version", agent.serviceVersion,
		"environment", agent.environment,
		"user_id", "user-12345",
		"ip_address", "192.168.1.100",
		"session_id", "sess-abcdef",
	)

	agent.logger.WarnContext(ctx, "High memory usage detected",
		"service", agent.serviceName,
		"memory_usage_percent", 85.2,
		"threshold_percent", 80.0,
		"component", "agent",
	)

	agent.logger.ErrorContext(ctx, "Database connection failed",
		"service", agent.serviceName,
		"error", "connection timeout",
		"database", "primary",
		"retry_count", 3,
		"component", "agent",
	)

	agent.logger.InfoContext(ctx, "Agent telemetry data sent",
		"service", agent.serviceName,
		"traces", "enabled",
		"metrics", "enabled",
		"logs", "enabled",
		"gateway_endpoint", "127.0.0.1:4328",
	)

	fmt.Println("   ✓ Structured logs with agent context generated")
}

// ApplicationExample demonstrates a complete application with OTEL Agent
func ApplicationExample() {
	fmt.Println("=== Complete Application with OTEL Agent ===")

	// Initialize OTEL Agent
	config := OTELAgentConfig{
		ServiceName:         "my-microservice",
		ServiceVersion:      "v2.1.0",
		Environment:         "production",
		GatewayHTTPEndpoint: "127.0.0.1:4328",
		EnableLogs:          true,
		EnableMetrics:       true,
		EnableTraces:        true,
	}

	agent, err := NewOTELAgent(config)
	if err != nil {
		log.Fatalf("Failed to initialize OTEL Agent: %v", err)
	}

	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		agent.Shutdown(ctx)
	}()

	// Simulate application workflow
	simulateApplicationWorkflow()

	fmt.Println("\n✅ Application telemetry data sent to OTEL Gateway")
	fmt.Printf("   Gateway processes data and forwards to:\n")
	fmt.Printf("   - VictoriaLogs (logs)\n")
	fmt.Printf("   - Prometheus (metrics)\n")
	fmt.Printf("   - Jaeger (traces)\n")
	fmt.Printf("   - SaaS platforms (Datadog, New Relic, etc.)\n")
}

// simulateApplicationWorkflow simulates a typical application workflow
func simulateApplicationWorkflow() {
	ctx := context.Background()
	tracer := otel.Tracer("microservice")
	meter := otel.Meter("microservice")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Create metrics
	requestsTotal, _ := meter.Int64Counter("requests_total")
	requestDuration, _ := meter.Float64Histogram("request_duration_seconds")

	// Simulate multiple requests
	for i := 0; i < 5; i++ {
		// Start request span
		ctx, span := tracer.Start(ctx, "process_request")

		requestStart := time.Now()

		// Log request start
		logger.InfoContext(ctx, "Processing request",
			"request_id", fmt.Sprintf("req-%d", i),
			"method", "POST",
			"path", "/api/process",
		)

		// Simulate processing
		simulateProcessing(ctx)

		// Record metrics
		duration := time.Since(requestStart).Seconds()
		requestsTotal.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("method", "POST"),
				attribute.String("status", "success"),
			),
		)
		requestDuration.Record(ctx, duration,
			metric.WithAttributes(
				attribute.String("endpoint", "/api/process"),
			),
		)

		// Log completion
		logger.InfoContext(ctx, "Request completed",
			"request_id", fmt.Sprintf("req-%d", i),
			"duration_ms", int(duration*1000),
		)

		span.End()
		time.Sleep(100 * time.Millisecond)
	}
}

// simulateProcessing simulates application business logic
func simulateProcessing(ctx context.Context) {
	tracer := otel.Tracer("microservice")

	// Database operation
	_, dbSpan := tracer.Start(ctx, "database_operation")
	dbSpan.SetAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", "INSERT"),
	)
	time.Sleep(20 * time.Millisecond)
	dbSpan.End()

	// External API call
	_, apiSpan := tracer.Start(ctx, "external_api_call")
	apiSpan.SetAttributes(
		attribute.String("http.method", "GET"),
		attribute.String("http.url", "https://api.external.com/data"),
	)
	time.Sleep(30 * time.Millisecond)
	apiSpan.End()

	// Cache operation
	_, cacheSpan := tracer.Start(ctx, "cache_operation")
	cacheSpan.SetAttributes(
		attribute.String("cache.system", "redis"),
		attribute.String("cache.operation", "SET"),
	)
	time.Sleep(5 * time.Millisecond)
	cacheSpan.End()
}

func main() {
	// Run basic OTEL Agent usage example
	ExampleOTELAgentUsage()
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	
	// Run complete application example
	ApplicationExample()
	
	fmt.Println("\n🎯 Verification Steps:")
	fmt.Println("   1. Check agent logs: /home/hellotalk/code/go/src/github.com/costa92/go-protoc/logs/agent-demo/agent.log")
	fmt.Println("   2. Check VictoriaLogs: http://127.0.0.1:9428/select/vmui/")
	fmt.Println("   3. Search for: service:\"example-app\"")
	fmt.Println("   4. Check Jaeger traces: http://127.0.0.1:16686")
	fmt.Println("   5. Look for service: example-app")
	fmt.Println("   6. Verify OTEL Collector metrics: curl http://127.0.0.1:8888/metrics")
}
