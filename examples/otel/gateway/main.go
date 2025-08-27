// Package main provides OpenTelemetry Gateway implementation examples
// demonstrating how to interact with and configure the Gateway layer
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// GatewayClient represents a client to interact with OTEL Gateway
type GatewayClient struct {
	baseURL    string
	httpClient *http.Client
}

// GatewayStatus represents the health status of OTEL Gateway
type GatewayStatus struct {
	Status   string    `json:"status"`
	UpSince  time.Time `json:"upSince"`
	UpTime   string    `json:"uptime"`
	Version  string    `json:"version,omitempty"`
	Build    string    `json:"build,omitempty"`
}

// GatewayMetrics represents metrics from OTEL Gateway
type GatewayMetrics struct {
	ProcessedSpans   int64 `json:"processed_spans"`
	ProcessedMetrics int64 `json:"processed_metrics"`  
	ProcessedLogs    int64 `json:"processed_logs"`
	DroppedData      int64 `json:"dropped_data"`
	ExportedData     int64 `json:"exported_data"`
}

// NewGatewayClient creates a new Gateway client
func NewGatewayClient(baseURL string) *GatewayClient {
	return &GatewayClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CheckHealth checks if the OTEL Gateway is healthy
func (c *GatewayClient) CheckHealth(ctx context.Context) (*GatewayStatus, error) {
	url := fmt.Sprintf("%s", c.baseURL+":13133")
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed: status %d, body: %s", resp.StatusCode, body)
	}

	var status GatewayStatus
	if err := json.Unmarshal(body, &status); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &status, nil
}

// GetMetrics retrieves metrics from the OTEL Gateway
func (c *GatewayClient) GetMetrics(ctx context.Context) (string, error) {
	url := fmt.Sprintf("%s:8888/metrics", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metrics request failed: status %d", resp.StatusCode)
	}

	return string(body), nil
}

// SendTestTrace sends a test trace to the Gateway
func (c *GatewayClient) SendTestTrace(ctx context.Context) error {
	// Create minimal OTLP trace payload
	tracePayload := map[string]interface{}{
		"resourceSpans": []map[string]interface{}{
			{
				"resource": map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"key":   "service.name",
							"value": map[string]interface{}{"stringValue": "test-service"},
						},
					},
				},
				"scopeSpans": []map[string]interface{}{
					{
						"spans": []map[string]interface{}{
							{
								"traceId":           "1234567890abcdef1234567890abcdef",
								"spanId":            "1234567890abcdef",
								"name":              "test-span",
								"startTimeUnixNano": fmt.Sprintf("%d", time.Now().UnixNano()),
								"endTimeUnixNano":   fmt.Sprintf("%d", time.Now().Add(time.Millisecond*100).UnixNano()),
								"kind":              1,
							},
						},
					},
				},
			},
		},
	}

	return c.sendOTLPData(ctx, "traces", tracePayload)
}

// SendTestMetric sends a test metric to the Gateway
func (c *GatewayClient) SendTestMetric(ctx context.Context) error {
	now := time.Now()
	metricPayload := map[string]interface{}{
		"resourceMetrics": []map[string]interface{}{
			{
				"resource": map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"key":   "service.name",
							"value": map[string]interface{}{"stringValue": "test-service"},
						},
					},
				},
				"scopeMetrics": []map[string]interface{}{
					{
						"metrics": []map[string]interface{}{
							{
								"name": "test_counter",
								"sum": map[string]interface{}{
									"dataPoints": []map[string]interface{}{
										{
											"timeUnixNano": fmt.Sprintf("%d", now.UnixNano()),
											"asInt":        42,
										},
									},
									"aggregationTemporality": 2,
									"isMonotonic":           true,
								},
							},
						},
					},
				},
			},
		},
	}

	return c.sendOTLPData(ctx, "metrics", metricPayload)
}

// SendTestLog sends a test log to the Gateway
func (c *GatewayClient) SendTestLog(ctx context.Context) error {
	logPayload := map[string]interface{}{
		"resourceLogs": []map[string]interface{}{
			{
				"resource": map[string]interface{}{
					"attributes": []map[string]interface{}{
						{
							"key":   "service.name",
							"value": map[string]interface{}{"stringValue": "test-service"},
						},
					},
				},
				"scopeLogs": []map[string]interface{}{
					{
						"logRecords": []map[string]interface{}{
							{
								"timeUnixNano":   fmt.Sprintf("%d", time.Now().UnixNano()),
								"severityNumber": 9,
								"severityText":   "INFO",
								"body":          map[string]interface{}{"stringValue": "Test log message"},
								"attributes": []map[string]interface{}{
									{
										"key":   "log.level",
										"value": map[string]interface{}{"stringValue": "info"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return c.sendOTLPData(ctx, "logs", logPayload)
}

// sendOTLPData sends OTLP data to the Gateway
func (c *GatewayClient) sendOTLPData(ctx context.Context, endpoint string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s:4328/v1/%s", c.baseURL, endpoint)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send %s: status %d, body: %s", endpoint, resp.StatusCode, body)
	}

	return nil
}

// GatewayConfigManager helps manage Gateway configuration
type GatewayConfigManager struct {
	configPath string
}

// NewGatewayConfigManager creates a new config manager
func NewGatewayConfigManager(configPath string) *GatewayConfigManager {
	return &GatewayConfigManager{
		configPath: configPath,
	}
}

// ExampleGatewayInteraction demonstrates Gateway interaction
func ExampleGatewayInteraction() {
	fmt.Println("=== OTEL Gateway Interaction Example ===")

	// Create Gateway client
	client := NewGatewayClient("http://127.0.0.1")
	ctx := context.Background()

	// Check Gateway health
	fmt.Println("\n🔍 Checking Gateway Health...")
	status, err := client.CheckHealth(ctx)
	if err != nil {
		fmt.Printf("❌ Health check failed: %v\n", err)
		fmt.Println("   Make sure OTEL Gateway is running: make deploy.install.docker.otel")
		return
	}

	fmt.Printf("✅ Gateway is healthy\n")
	fmt.Printf("   Status: %s\n", status.Status)
	fmt.Printf("   Uptime: %s\n", status.UpTime)
	fmt.Printf("   Started: %s\n", status.UpSince.Format(time.RFC3339))

	// Send test telemetry data
	fmt.Println("\n📤 Sending Test Data to Gateway...")

	if err := client.SendTestTrace(ctx); err != nil {
		fmt.Printf("❌ Failed to send trace: %v\n", err)
	} else {
		fmt.Println("✅ Test trace sent successfully")
	}

	if err := client.SendTestMetric(ctx); err != nil {
		fmt.Printf("❌ Failed to send metric: %v\n", err)
	} else {
		fmt.Println("✅ Test metric sent successfully")
	}

	if err := client.SendTestLog(ctx); err != nil {
		fmt.Printf("❌ Failed to send log: %v\n", err)
	} else {
		fmt.Println("✅ Test log sent successfully")
	}

	// Get Gateway metrics
	fmt.Println("\n📊 Gateway Processing Metrics...")
	metrics, err := client.GetMetrics(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to get metrics: %v\n", err)
		return
	}

	// Parse and display relevant metrics
	displayGatewayMetrics(metrics)

	fmt.Println("\n🎯 Gateway -> SaaS Data Flow:")
	fmt.Println("   📝 Logs → VictoriaLogs")
	fmt.Println("   📊 Metrics → Prometheus")
	fmt.Println("   🔍 Traces → Jaeger")
	fmt.Println("   ☁️  All Data → SaaS Platforms (if configured)")
}

// displayGatewayMetrics parses and displays key Gateway metrics
func displayGatewayMetrics(metricsData string) {
	fmt.Println("   Key Gateway Metrics:")
	
	// Simple parsing for demo - in production use proper Prometheus client
	lines := bytes.Split([]byte(metricsData), []byte("\n"))
	
	relevantMetrics := map[string]string{
		"otelcol_processor_batch_batch_send_size_sum":     "Batch Send Size",
		"otelcol_receiver_accepted_spans_total":           "Accepted Spans",
		"otelcol_receiver_accepted_metric_points_total":   "Accepted Metrics", 
		"otelcol_receiver_accepted_log_records_total":     "Accepted Logs",
		"otelcol_exporter_sent_spans_total":              "Exported Spans",
		"otelcol_exporter_sent_metric_points_total":      "Exported Metrics",
		"otelcol_exporter_sent_log_records_total":        "Exported Logs",
	}

	for _, line := range lines {
		lineStr := string(line)
		for metricName, displayName := range relevantMetrics {
			if bytes.Contains(line, []byte(metricName)) && !bytes.Contains(line, []byte("#")) {
				fmt.Printf("   - %s: %s\n", displayName, extractMetricValue(lineStr))
				break
			}
		}
	}
}

// extractMetricValue extracts the numeric value from a Prometheus metric line
func extractMetricValue(line string) string {
	parts := bytes.Fields([]byte(line))
	if len(parts) >= 2 {
		return string(parts[len(parts)-1])
	}
	return "N/A"
}

// ExampleGatewayConfiguration demonstrates Gateway configuration management
func ExampleGatewayConfiguration() {
	fmt.Println("=== OTEL Gateway Configuration Example ===")

	configPath := "/home/hellotalk/code/go/src/github.com/costa92/go-protoc/_thirdparty/otel/config/otel-collector.yaml"
	
	fmt.Printf("\n📋 Gateway Configuration Location:\n")
	fmt.Printf("   Path: %s\n", configPath)
	
	fmt.Printf("\n🔧 Configuration Structure:\n")
	fmt.Println("   receivers:")
	fmt.Println("     - filelog (monitors application log files)")
	fmt.Println("     - otlp (receives data from Agents)")
	fmt.Println("     - prometheus (scrapes metrics)")
	fmt.Println("   ")
	fmt.Println("   processors:")
	fmt.Println("     - batch (batches data for efficiency)")
	fmt.Println("     - resource (adds metadata)")
	fmt.Println("     - memory_limiter (prevents OOM)")
	fmt.Println("   ")
	fmt.Println("   exporters:")
	fmt.Println("     - file (local backup)")
	fmt.Println("     - otlphttp/victorialogs (log storage)")
	fmt.Println("     - prometheus (metric storage)")
	fmt.Println("     - otlphttp/jaeger (trace storage)")
	
	fmt.Printf("\n📊 Data Processing Flow:\n")
	fmt.Println("   Agent Data → Gateway Receivers → Processors → Exporters → SaaS")
	fmt.Println("   ")
	fmt.Println("   Receivers collect from:")
	fmt.Println("   - OTLP endpoints (4327 gRPC, 4328 HTTP)")
	fmt.Println("   - File monitoring (/host/logs/**/*.log)")
	fmt.Println("   - Prometheus scraping (host.docker.internal:8080)")
	fmt.Println("   ")
	fmt.Println("   Exporters forward to:")
	fmt.Println("   - VictoriaLogs (http://proj-victorialogs:9428)")
	fmt.Println("   - Prometheus (0.0.0.0:8889)")
	fmt.Println("   - Jaeger (http://proj-jaeger:14268)")

	demonstrateConfigurationCustomization()
}

// demonstrateConfigurationCustomization shows how to customize Gateway config
func demonstrateConfigurationCustomization() {
	fmt.Printf("\n⚙️  Configuration Customization Examples:\n")
	
	fmt.Println("\n1. Adding SaaS Platform Export:")
	fmt.Println("   exporters:")
	fmt.Println("     otlphttp/datadog:")
	fmt.Println("       endpoint: \"https://otlp.datadoghq.com\"")
	fmt.Println("       headers:")
	fmt.Println("         DD-API-KEY: \"${DD_API_KEY}\"")
	fmt.Println("       compression: gzip")

	fmt.Println("\n2. Custom Sampling Configuration:")
	fmt.Println("   processors:")
	fmt.Println("     probabilistic_sampler:")
	fmt.Println("       sampling_percentage: 10  # 10% sampling")

	fmt.Println("\n3. Enhanced Resource Attributes:")
	fmt.Println("   processors:")
	fmt.Println("     resource:")
	fmt.Println("       attributes:")
	fmt.Println("         - key: deployment.environment")
	fmt.Println("           value: \"${PROJ_ENVIRONMENT}\"")
	fmt.Println("         - key: k8s.cluster.name")
	fmt.Println("           value: \"production-cluster\"")

	fmt.Println("\n4. Batch Optimization:")
	fmt.Println("   processors:")
	fmt.Println("     batch:")
	fmt.Println("       timeout: 1s")
	fmt.Println("       send_batch_size: 512")
	fmt.Println("       send_batch_max_size: 1024")

	fmt.Printf("\n💡 Configuration Tips:\n")
	fmt.Println("   - Restart Gateway after config changes: docker restart proj-otel-collector")
	fmt.Println("   - Test config with: docker logs proj-otel-collector")
	fmt.Println("   - Monitor metrics at: http://127.0.0.1:8888/metrics")
	fmt.Println("   - Health check at: http://127.0.0.1:13133")
}

// ExampleGatewayMonitoring demonstrates Gateway monitoring and observability
func ExampleGatewayMonitoring() {
	fmt.Println("=== OTEL Gateway Monitoring Example ===")
	
	client := NewGatewayClient("http://127.0.0.1")
	ctx := context.Background()

	fmt.Println("\n📊 Gateway Monitoring Dashboard:")
	
	// Continuous monitoring loop (simplified for demo)
	for i := 0; i < 3; i++ {
		fmt.Printf("\n--- Monitoring Check #%d ---\n", i+1)
		
		// Health status
		if status, err := client.CheckHealth(ctx); err == nil {
			fmt.Printf("✅ Status: %s (uptime: %s)\n", status.Status, status.UpTime)
		} else {
			fmt.Printf("❌ Health check failed: %v\n", err)
		}

		// Process metrics
		if metrics, err := client.GetMetrics(ctx); err == nil {
			fmt.Println("📈 Processing Metrics:")
			displayGatewayMetrics(metrics)
		} else {
			fmt.Printf("❌ Failed to get metrics: %v\n", err)
		}

		if i < 2 { // Don't sleep on last iteration
			time.Sleep(2 * time.Second)
		}
	}

	fmt.Println("\n🔍 Monitoring Best Practices:")
	fmt.Println("   - Set up alerts on Gateway health (13133 endpoint)")
	fmt.Println("   - Monitor memory usage (prevent OOM)")
	fmt.Println("   - Track data processing rates")
	fmt.Println("   - Monitor export success rates")
	fmt.Println("   - Set up dashboard for key metrics")
	
	fmt.Println("\n📱 Available Monitoring Endpoints:")
	fmt.Println("   Health Check: http://127.0.0.1:13133")
	fmt.Println("   Internal Metrics: http://127.0.0.1:8888/metrics")
	fmt.Println("   Performance Profiling: http://127.0.0.1:1777/debug/pprof/")
}

// GatewayTestSuite provides comprehensive testing for Gateway functionality
type GatewayTestSuite struct {
	client *GatewayClient
}

// NewGatewayTestSuite creates a new test suite
func NewGatewayTestSuite(baseURL string) *GatewayTestSuite {
	return &GatewayTestSuite{
		client: NewGatewayClient(baseURL),
	}
}

// RunComprehensiveTest runs a comprehensive test of Gateway functionality
func (ts *GatewayTestSuite) RunComprehensiveTest(ctx context.Context) error {
	fmt.Println("=== Gateway Comprehensive Test Suite ===")

	tests := []struct {
		name string
		test func() error
	}{
		{"Health Check", ts.testHealth},
		{"Trace Processing", ts.testTraceProcessing},
		{"Metric Processing", ts.testMetricProcessing},
		{"Log Processing", ts.testLogProcessing},
		{"Metrics Collection", ts.testMetricsCollection},
	}

	var passed, failed int
	for _, test := range tests {
		fmt.Printf("\n🧪 Running: %s\n", test.name)
		if err := test.test(); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			failed++
		} else {
			fmt.Printf("✅ Passed\n")
			passed++
		}
	}

	fmt.Printf("\n📊 Test Results: %d passed, %d failed\n", passed, failed)
	if failed == 0 {
		fmt.Println("🎉 All tests passed! Gateway is working correctly.")
		return nil
	}
	return fmt.Errorf("%d tests failed", failed)
}

// testHealth tests Gateway health endpoint
func (ts *GatewayTestSuite) testHealth() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := ts.client.CheckHealth(ctx)
	return err
}

// testTraceProcessing tests trace data processing
func (ts *GatewayTestSuite) testTraceProcessing() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return ts.client.SendTestTrace(ctx)
}

// testMetricProcessing tests metric data processing
func (ts *GatewayTestSuite) testMetricProcessing() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return ts.client.SendTestMetric(ctx)
}

// testLogProcessing tests log data processing
func (ts *GatewayTestSuite) testLogProcessing() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	return ts.client.SendTestLog(ctx)
}

// testMetricsCollection tests Gateway metrics collection
func (ts *GatewayTestSuite) testMetricsCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err := ts.client.GetMetrics(ctx)
	return err
}

// ExampleGatewayTesting demonstrates comprehensive Gateway testing
func ExampleGatewayTesting() {
	fmt.Println("=== OTEL Gateway Testing Example ===")

	testSuite := NewGatewayTestSuite("http://127.0.0.1")
	ctx := context.Background()

	if err := testSuite.RunComprehensiveTest(ctx); err != nil {
		log.Printf("Test suite failed: %v", err)
	}

	fmt.Println("\n💡 Testing Tips:")
	fmt.Println("   - Run tests before deploying configuration changes")
	fmt.Println("   - Use automated testing in CI/CD pipelines")
	fmt.Println("   - Test with realistic data volumes")
	fmt.Println("   - Verify data reaches all configured exporters")
}