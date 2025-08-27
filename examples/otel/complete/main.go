// Package main provides comprehensive OpenTelemetry end-to-end integration examples
// demonstrating the complete Agent -> Gateway -> SaaS architecture pipeline
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sync"
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

// CompleteOTELDemo represents the complete end-to-end OTEL demonstration
type CompleteOTELDemo struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	logger         *slog.Logger
	serviceName    string
	serviceVersion string
	environment    string
}

// WorkflowScenario represents a business workflow scenario
type WorkflowScenario struct {
	Name       string
	Duration   time.Duration
	Operations int
	Type       string
}

// CompleteOTELExample demonstrates the full Agent -> Gateway -> SaaS architecture
// This example shows how all three layers work together in a production scenario
func CompleteOTELExample() {
	fmt.Println("=== Complete OTEL Architecture Example ===")
	fmt.Println("Demonstrating Agent -> Gateway -> SaaS pipeline")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Initialize complete OTEL demo
	demo, err := NewCompleteOTELDemo()
	if err != nil {
		log.Printf("❌ Failed to initialize complete OTEL demo: %v", err)
		log.Println("   💡 Ensure OTEL Gateway is running: docker restart proj-otel-collector")
		return
	}
	defer demo.Shutdown(ctx)

	fmt.Println("\n✅ Complete OTEL Demo initialized successfully!")
	fmt.Printf("   📝 Service: %s v%s (%s)\n", demo.serviceName, demo.serviceVersion, demo.environment)
	fmt.Println("   🔄 Data Flow: Application -> Agent -> Gateway -> SaaS platforms")

	// Run comprehensive workflow
	runCompleteWorkflow(ctx, demo)

	fmt.Println("\n🎉 Complete OTEL architecture demonstration finished!")
	printVerificationInstructions()
}

// NewCompleteOTELDemo creates a new complete OTEL demonstration
func NewCompleteOTELDemo() (*CompleteOTELDemo, error) {
	demo := &CompleteOTELDemo{
		serviceName:    "complete-otel-demo",
		serviceVersion: "v1.0.0",
		environment:    "demo",
	}

	ctx := context.Background()

	// Create shared resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(demo.serviceName),
			semconv.ServiceVersionKey.String(demo.serviceVersion),
			semconv.DeploymentEnvironmentKey.String(demo.environment),
			attribute.String("demo.type", "complete-workflow"),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize traces
	if err := demo.initTraces(ctx, res); err != nil {
		return nil, fmt.Errorf("failed to initialize traces: %w", err)
	}

	// Initialize metrics
	if err := demo.initMetrics(ctx, res); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Initialize logs
	if err := demo.initLogs(ctx, res); err != nil {
		return nil, fmt.Errorf("failed to initialize logs: %w", err)
	}

	return demo, nil
}

// initTraces initializes tracing
func (d *CompleteOTELDemo) initTraces(ctx context.Context, res *resource.Resource) error {
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("127.0.0.1:4328"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	d.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(d.tracerProvider)
	return nil
}

// initMetrics initializes metrics
func (d *CompleteOTELDemo) initMetrics(ctx context.Context, res *resource.Resource) error {
	metricExporter, err := otlpmetrichttp.New(ctx,
		otlpmetrichttp.WithEndpoint("127.0.0.1:4328"),
		otlpmetrichttp.WithInsecure(),
	)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	d.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter,
			sdkmetric.WithInterval(10*time.Second),
		)),
		sdkmetric.WithResource(res),
	)

	otel.SetMeterProvider(d.meterProvider)
	return nil
}

// initLogs initializes file-based logging
func (d *CompleteOTELDemo) initLogs(ctx context.Context, res *resource.Resource) error {
	// Create log file for OTEL Collector
	logDir := "/home/hellotalk/code/go/src/github.com/costa92/go-protoc/logs/complete-demo"
	logFile := logDir + "/complete.log"

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
	d.logger = slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	fmt.Printf("📝 Complete demo logs writing to: %s\n", logFile)
	return nil
}

// Shutdown gracefully shuts down the demo
func (d *CompleteOTELDemo) Shutdown(ctx context.Context) error {
	var errors []error

	if d.tracerProvider != nil {
		if err := d.tracerProvider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown tracer provider: %w", err))
		}
	}

	if d.meterProvider != nil {
		if err := d.meterProvider.Shutdown(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to shutdown meter provider: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("shutdown errors: %v", errors)
	}

	return nil
}

// runCompleteWorkflow executes the complete end-to-end workflow demonstration
func runCompleteWorkflow(ctx context.Context, demo *CompleteOTELDemo) {
	fmt.Println("\n🚀 Starting Complete Workflow Demonstration...")

	// Define business scenarios
	scenarios := []WorkflowScenario{
		{
			Name:       "E-commerce Order Processing",
			Duration:   15 * time.Second,
			Operations: 8,
			Type:       "transaction",
		},
		{
			Name:       "User Authentication Flow",
			Duration:   10 * time.Second,
			Operations: 5,
			Type:       "authentication",
		},
		{
			Name:       "Data Analytics Pipeline",
			Duration:   20 * time.Second,
			Operations: 12,
			Type:       "analytics",
		},
		{
			Name:       "API Gateway Traffic",
			Duration:   12 * time.Second,
			Operations: 10,
			Type:       "api",
		},
	}

	var wg sync.WaitGroup

	// Run scenarios concurrently
	for i, scenario := range scenarios {
		wg.Add(1)
		go func(idx int, s WorkflowScenario) {
			defer wg.Done()
			runWorkflowScenario(ctx, demo, idx+1, s)
		}(i, scenario)
	}

	// Monitor progress
	go monitorWorkflowProgress(ctx, demo, len(scenarios))

	// Wait for all scenarios to complete
	wg.Wait()

	fmt.Println("\n✅ All workflow scenarios completed!")
	fmt.Println("📊 Telemetry data has been sent through the complete pipeline:")
	fmt.Println("   Application → Agent → Gateway → SaaS Platforms")
}

// runWorkflowScenario executes a single workflow scenario
func runWorkflowScenario(ctx context.Context, demo *CompleteOTELDemo, scenarioID int, scenario WorkflowScenario) {
	tracer := otel.Tracer("complete-demo")
	meter := otel.Meter("complete-demo")

	// Create metrics
	operationCounter, _ := meter.Int64Counter(
		"workflow_operations_total",
		metric.WithDescription("Total number of workflow operations"),
	)

	operationDuration, _ := meter.Float64Histogram(
		"workflow_operation_duration_seconds",
		metric.WithDescription("Duration of workflow operations"),
	)

	// Create root span for the scenario
	ctx, rootSpan := tracer.Start(ctx, scenario.Name)
	defer rootSpan.End()

	rootSpan.SetAttributes(
		attribute.String("scenario.name", scenario.Name),
		attribute.String("scenario.type", scenario.Type),
		attribute.Int("scenario.id", scenarioID),
		attribute.Int("scenario.operations", scenario.Operations),
		attribute.String("scenario.duration", scenario.Duration.String()),
	)

	demo.logger.InfoContext(ctx, "Starting workflow scenario",
		"scenario_id", scenarioID,
		"scenario_name", scenario.Name,
		"scenario_type", scenario.Type,
		"expected_operations", scenario.Operations,
		"expected_duration", scenario.Duration.String(),
	)

	// Execute operations within the scenario
	operationInterval := scenario.Duration / time.Duration(scenario.Operations)
	for i := 0; i < scenario.Operations; i++ {
		operationStart := time.Now()

		// Create operation span
		_, opSpan := tracer.Start(ctx, fmt.Sprintf("%s-operation-%d", scenario.Type, i+1))

		// Simulate different types of operations based on scenario type
		switch scenario.Type {
		case "transaction":
			simulateTransactionOperation(ctx, demo, i+1, scenario.Name)
		case "authentication":
			simulateAuthenticationOperation(ctx, demo, i+1, scenario.Name)
		case "analytics":
			simulateAnalyticsOperation(ctx, demo, i+1, scenario.Name)
		case "api":
			simulateAPIOperation(ctx, demo, i+1, scenario.Name)
		default:
			simulateGenericOperation(ctx, demo, i+1, scenario.Name)
		}

		// Record metrics
		operationDuration.Record(ctx, time.Since(operationStart).Seconds(),
			metric.WithAttributes(
				attribute.String("scenario", scenario.Name),
				attribute.String("operation", fmt.Sprintf("op-%d", i+1)),
			),
		)

		operationCounter.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("scenario", scenario.Name),
				attribute.String("status", "success"),
			),
		)

		opSpan.End()

		// Wait before next operation
		time.Sleep(operationInterval)
	}

	demo.logger.InfoContext(ctx, "Completed workflow scenario",
		"scenario_id", scenarioID,
		"scenario_name", scenario.Name,
		"actual_operations", scenario.Operations,
		"actual_duration", scenario.Duration.String(),
		"status", "success",
	)
}

// simulateTransactionOperation simulates e-commerce transaction operations
func simulateTransactionOperation(ctx context.Context, demo *CompleteOTELDemo, opID int, scenarioName string) {
	tracer := otel.Tracer("complete-demo")

	// Payment processing
	_, paymentSpan := tracer.Start(ctx, "payment-processing")
	paymentSpan.SetAttributes(
		attribute.String("payment.method", "credit_card"),
		attribute.Float64("payment.amount", 99.99),
		attribute.String("payment.currency", "USD"),
	)
	time.Sleep(30 * time.Millisecond)
	paymentSpan.End()

	// Inventory check
	_, inventorySpan := tracer.Start(ctx, "inventory-check")
	inventorySpan.SetAttributes(
		attribute.String("product.id", fmt.Sprintf("prod-%d", opID)),
		attribute.Int("inventory.quantity", 10-opID),
	)
	time.Sleep(20 * time.Millisecond)
	inventorySpan.End()

	demo.logger.InfoContext(ctx, "Transaction operation completed",
		"operation_id", opID,
		"scenario", scenarioName,
		"payment_amount", 99.99,
		"payment_method", "credit_card",
	)
}

// simulateAuthenticationOperation simulates user authentication operations
func simulateAuthenticationOperation(ctx context.Context, demo *CompleteOTELDemo, opID int, scenarioName string) {
	tracer := otel.Tracer("complete-demo")

	// User validation
	_, authSpan := tracer.Start(ctx, "user-authentication")
	authSpan.SetAttributes(
		attribute.String("auth.method", "oauth2"),
		attribute.String("user.id", fmt.Sprintf("user-%d", opID)),
		attribute.Bool("auth.success", true),
	)
	time.Sleep(25 * time.Millisecond)
	authSpan.End()

	demo.logger.InfoContext(ctx, "Authentication operation completed",
		"operation_id", opID,
		"scenario", scenarioName,
		"user_id", fmt.Sprintf("user-%d", opID),
		"auth_method", "oauth2",
		"auth_success", true,
	)
}

// simulateAnalyticsOperation simulates data analytics operations
func simulateAnalyticsOperation(ctx context.Context, demo *CompleteOTELDemo, opID int, scenarioName string) {
	tracer := otel.Tracer("complete-demo")

	// Data processing
	_, dataSpan := tracer.Start(ctx, "data-processing")
	dataSpan.SetAttributes(
		attribute.String("data.source", "clickstream"),
		attribute.Int("data.records", opID*100),
		attribute.String("processing.type", "aggregation"),
	)
	time.Sleep(40 * time.Millisecond)
	dataSpan.End()

	demo.logger.InfoContext(ctx, "Analytics operation completed",
		"operation_id", opID,
		"scenario", scenarioName,
		"data_records", opID*100,
		"processing_type", "aggregation",
	)
}

// simulateAPIOperation simulates API gateway operations
func simulateAPIOperation(ctx context.Context, demo *CompleteOTELDemo, opID int, scenarioName string) {
	tracer := otel.Tracer("complete-demo")

	// API request processing
	_, apiSpan := tracer.Start(ctx, "api-request")
	apiSpan.SetAttributes(
		attribute.String("http.method", "POST"),
		attribute.String("http.route", "/api/v1/data"),
		attribute.Int("http.status_code", 200),
	)
	time.Sleep(15 * time.Millisecond)
	apiSpan.End()

	demo.logger.InfoContext(ctx, "API operation completed",
		"operation_id", opID,
		"scenario", scenarioName,
		"http_method", "POST",
		"http_status", 200,
	)
}

// simulateGenericOperation simulates generic operations
func simulateGenericOperation(ctx context.Context, demo *CompleteOTELDemo, opID int, scenarioName string) {
	time.Sleep(20 * time.Millisecond)

	demo.logger.InfoContext(ctx, "Generic operation completed",
		"operation_id", opID,
		"scenario", scenarioName,
		"operation_type", "generic",
	)
}

// monitorWorkflowProgress monitors the progress of all workflow scenarios
func monitorWorkflowProgress(ctx context.Context, demo *CompleteOTELDemo, totalScenarios int) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Check Gateway health
			gatewayHealthy := checkGatewayHealth()
			
			demo.logger.InfoContext(ctx, "Workflow progress update",
				"total_scenarios", totalScenarios,
				"gateway_healthy", gatewayHealthy,
				"timestamp", time.Now().Format(time.RFC3339),
			)

			if gatewayHealthy {
				fmt.Printf("   ✅ Gateway healthy - data flowing to SaaS platforms\n")
			} else {
				fmt.Printf("   ⚠️  Gateway connection issue - check OTEL Collector status\n")
			}
		}
	}
}

// checkGatewayHealth checks if the OTEL Gateway is healthy
func checkGatewayHealth() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://127.0.0.1:13133", nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// printVerificationInstructions prints verification instructions
func printVerificationInstructions() {
	fmt.Println("\n🎯 Verification Steps:")
	fmt.Println("   1. Check complete demo logs: /home/hellotalk/code/go/src/github.com/costa92/go-protoc/logs/complete-demo/complete.log")
	fmt.Println("   2. Check VictoriaLogs: http://127.0.0.1:9428/select/vmui/")
	fmt.Println("   3. Search for: service.name:\"complete-otel-demo\"")
	fmt.Println("   4. Check Jaeger traces: http://127.0.0.1:16686")
	fmt.Println("   5. Look for service: complete-otel-demo")
	fmt.Println("   6. Verify OTEL Collector health: curl http://127.0.0.1:13133")
	fmt.Println("   7. Check Gateway metrics: curl http://127.0.0.1:8888/metrics")

	fmt.Println("\n📊 Expected Data in SaaS Platforms:")
	fmt.Println("   📝 Logs: Workflow operations, scenario progress, authentication events")
	fmt.Println("   📈 Metrics: Operation counts, duration histograms, error rates")
	fmt.Println("   🔍 Traces: End-to-end transaction flows, service dependencies")

	fmt.Println("\n🔄 Data Flow Verification:")
	fmt.Println("   Application → Agent (OTLP) → Gateway (OTEL Collector) → SaaS Platforms")
	fmt.Println("   - Agent: OTLP exporters send traces/metrics to Gateway")
	fmt.Println("   - Gateway: Processes and forwards to VictoriaLogs, Jaeger, etc.")
	fmt.Println("   - SaaS: Query data from local platforms or configured cloud services")
}

func main() {
	// Run complete OTEL architecture demonstration
	CompleteOTELExample()
}