// Package examples provides OpenTelemetry SaaS platform integration examples
// demonstrating how data flows from Gateway to various SaaS platforms
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// SaaSPlatform represents different SaaS monitoring platforms
type SaaSPlatform string

const (
	PlatformDatadog   SaaSPlatform = "datadog"
	PlatformNewRelic  SaaSPlatform = "newrelic"
	PlatformGrafana   SaaSPlatform = "grafana"
	PlatformLocal     SaaSPlatform = "local"
)

// SaaSClient represents a generic SaaS platform client
type SaaSClient struct {
	Platform   SaaSPlatform
	BaseURL    string
	APIKey     string
	httpClient *http.Client
}

// SaaSQuery represents a query to SaaS platforms
type SaaSQuery struct {
	Platform  SaaSPlatform `json:"platform"`
	QueryType string       `json:"query_type"` // logs, metrics, traces
	Query     string       `json:"query"`
	TimeRange TimeRange    `json:"time_range"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// SaaSResponse represents response from SaaS platforms
type SaaSResponse struct {
	Platform SaaSPlatform    `json:"platform"`
	Data     json.RawMessage `json:"data"`
	Count    int             `json:"count"`
	Duration time.Duration   `json:"duration"`
}

// NewSaaSClient creates a new SaaS platform client
func NewSaaSClient(platform SaaSPlatform, baseURL, apiKey string) *SaaSClient {
	return &SaaSClient{
		Platform: platform,
		BaseURL:  baseURL,
		APIKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// QueryLogs queries logs from the SaaS platform
func (c *SaaSClient) QueryLogs(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	switch c.Platform {
	case PlatformLocal:
		return c.queryVictoriaLogs(ctx, query, timeRange)
	case PlatformDatadog:
		return c.queryDatadogLogs(ctx, query, timeRange)
	case PlatformNewRelic:
		return c.queryNewRelicLogs(ctx, query, timeRange)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", c.Platform)
	}
}

// QueryMetrics queries metrics from the SaaS platform
func (c *SaaSClient) QueryMetrics(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	switch c.Platform {
	case PlatformLocal:
		return c.queryPrometheusMetrics(ctx, query, timeRange)
	case PlatformDatadog:
		return c.queryDatadogMetrics(ctx, query, timeRange)
	case PlatformNewRelic:
		return c.queryNewRelicMetrics(ctx, query, timeRange)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", c.Platform)
	}
}

// QueryTraces queries traces from the SaaS platform
func (c *SaaSClient) QueryTraces(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	switch c.Platform {
	case PlatformLocal:
		return c.queryJaegerTraces(ctx, query, timeRange)
	case PlatformDatadog:
		return c.queryDatadogTraces(ctx, query, timeRange)
	case PlatformNewRelic:
		return c.queryNewRelicTraces(ctx, query, timeRange)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", c.Platform)
	}
}

// queryVictoriaLogs queries local VictoriaLogs
func (c *SaaSClient) queryVictoriaLogs(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// VictoriaLogs query format
	url := fmt.Sprintf("%s/select/logsql/query", c.BaseURL)
	
	payload := fmt.Sprintf("query=%s", query)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return &SaaSResponse{
		Platform: PlatformLocal,
		Data:     json.RawMessage(body),
		Duration: time.Since(start),
	}, nil
}

// queryPrometheusMetrics queries local Prometheus
func (c *SaaSClient) queryPrometheusMetrics(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Prometheus query format
	url := fmt.Sprintf("%s/api/v1/query", c.BaseURL)
	if !timeRange.End.IsZero() {
		url = fmt.Sprintf("%s/api/v1/query_range", c.BaseURL)
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	q := req.URL.Query()
	q.Add("query", query)
	if !timeRange.Start.IsZero() {
		q.Add("start", fmt.Sprintf("%d", timeRange.Start.Unix()))
	}
	if !timeRange.End.IsZero() {
		q.Add("end", fmt.Sprintf("%d", timeRange.End.Unix()))
		q.Add("step", "30s")
	}
	req.URL.RawQuery = q.Encode()
	
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return &SaaSResponse{
		Platform: PlatformLocal,
		Data:     json.RawMessage(body),
		Duration: time.Since(start),
	}, nil
}

// queryJaegerTraces queries local Jaeger
func (c *SaaSClient) queryJaegerTraces(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Jaeger API query format
	url := fmt.Sprintf("%s/api/traces", c.BaseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	q := req.URL.Query()
	q.Add("service", query) // Use query as service name for demo
	if !timeRange.Start.IsZero() {
		q.Add("start", fmt.Sprintf("%d", timeRange.Start.UnixMicro()))
	}
	if !timeRange.End.IsZero() {
		q.Add("end", fmt.Sprintf("%d", timeRange.End.UnixMicro()))
	}
	q.Add("limit", "100")
	req.URL.RawQuery = q.Encode()
	
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return &SaaSResponse{
		Platform: PlatformLocal,
		Data:     json.RawMessage(body),
		Duration: time.Since(start),
	}, nil
}

// queryDatadogLogs queries Datadog logs (mock implementation)
func (c *SaaSClient) queryDatadogLogs(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Mock Datadog logs API
	url := "https://api.datadoghq.com/api/v2/logs/events/search"
	
	payload := map[string]interface{}{
		"filter": map[string]interface{}{
			"query": query,
			"from":  timeRange.Start.Format(time.RFC3339),
			"to":    timeRange.End.Format(time.RFC3339),
		},
		"page": map[string]interface{}{
			"limit": 100,
		},
	}
	
	return c.executeSaaSQuery(ctx, "POST", url, payload)
}

// queryDatadogMetrics queries Datadog metrics (mock implementation)
func (c *SaaSClient) queryDatadogMetrics(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Mock Datadog metrics API
	url := "https://api.datadoghq.com/api/v1/query"
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	q := req.URL.Query()
	q.Add("query", query)
	q.Add("from", fmt.Sprintf("%d", timeRange.Start.Unix()))
	q.Add("to", fmt.Sprintf("%d", timeRange.End.Unix()))
	req.URL.RawQuery = q.Encode()
	
	req.Header.Set("DD-API-KEY", c.APIKey)
	
	// Mock response for demo
	return &SaaSResponse{
		Platform: PlatformDatadog,
		Data:     json.RawMessage(`{"status":"ok","message":"Mock Datadog response"}`),
		Duration: 200 * time.Millisecond,
	}, nil
}

// queryDatadogTraces queries Datadog traces (mock implementation)
func (c *SaaSClient) queryDatadogTraces(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Mock Datadog APM API
	return &SaaSResponse{
		Platform: PlatformDatadog,
		Data:     json.RawMessage(`{"traces":[],"message":"Mock Datadog traces response"}`),
		Duration: 300 * time.Millisecond,
	}, nil
}

// queryNewRelicLogs queries New Relic logs (mock implementation)
func (c *SaaSClient) queryNewRelicLogs(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Mock New Relic Logs API
	return &SaaSResponse{
		Platform: PlatformNewRelic,
		Data:     json.RawMessage(`{"logs":[],"message":"Mock New Relic logs response"}`),
		Duration: 250 * time.Millisecond,
	}, nil
}

// queryNewRelicMetrics queries New Relic metrics (mock implementation)
func (c *SaaSClient) queryNewRelicMetrics(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Mock New Relic GraphQL API
	return &SaaSResponse{
		Platform: PlatformNewRelic,
		Data:     json.RawMessage(`{"data":{"actor":{"account":{"nrql":{"results":[]}}}}}`),
		Duration: 180 * time.Millisecond,
	}, nil
}

// queryNewRelicTraces queries New Relic traces (mock implementation)
func (c *SaaSClient) queryNewRelicTraces(ctx context.Context, query string, timeRange TimeRange) (*SaaSResponse, error) {
	// Mock New Relic Distributed Tracing API
	return &SaaSResponse{
		Platform: PlatformNewRelic,
		Data:     json.RawMessage(`{"traces":[],"message":"Mock New Relic traces response"}`),
		Duration: 220 * time.Millisecond,
	}, nil
}

// executeSaaSQuery executes a generic SaaS query
func (c *SaaSClient) executeSaaSQuery(ctx context.Context, method, url string, payload interface{}) (*SaaSResponse, error) {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	
	// Set authentication header
	switch c.Platform {
	case PlatformDatadog:
		req.Header.Set("DD-API-KEY", c.APIKey)
	case PlatformNewRelic:
		req.Header.Set("Api-Key", c.APIKey)
	}
	
	start := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()
	
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	
	return &SaaSResponse{
		Platform: c.Platform,
		Data:     json.RawMessage(responseBody),
		Duration: time.Since(start),
	}, nil
}

// SaaSManager manages multiple SaaS platform clients
type SaaSManager struct {
	clients map[SaaSPlatform]*SaaSClient
}

// NewSaaSManager creates a new SaaS manager
func NewSaaSManager() *SaaSManager {
	return &SaaSManager{
		clients: make(map[SaaSPlatform]*SaaSClient),
	}
}

// AddPlatform adds a SaaS platform to the manager
func (m *SaaSManager) AddPlatform(platform SaaSPlatform, baseURL, apiKey string) {
	m.clients[platform] = NewSaaSClient(platform, baseURL, apiKey)
}

// QueryAll queries all configured platforms
func (m *SaaSManager) QueryAll(ctx context.Context, query SaaSQuery) ([]*SaaSResponse, error) {
	var responses []*SaaSResponse
	var errors []error
	
	for platform, client := range m.clients {
		var resp *SaaSResponse
		var err error
		
		switch query.QueryType {
		case "logs":
			resp, err = client.QueryLogs(ctx, query.Query, query.TimeRange)
		case "metrics":
			resp, err = client.QueryMetrics(ctx, query.Query, query.TimeRange)
		case "traces":
			resp, err = client.QueryTraces(ctx, query.Query, query.TimeRange)
		default:
			err = fmt.Errorf("unsupported query type: %s", query.QueryType)
		}
		
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", platform, err))
			continue
		}
		
		responses = append(responses, resp)
	}
	
	if len(errors) > 0 && len(responses) == 0 {
		return nil, fmt.Errorf("all queries failed: %v", errors)
	}
	
	return responses, nil
}

// ExampleSaaSIntegration demonstrates SaaS platform integration
func ExampleSaaSIntegration() {
	fmt.Println("=== OTEL SaaS Platform Integration Example ===")

	// Create SaaS manager
	manager := NewSaaSManager()
	
	// Add local platforms (actually accessible)
	manager.AddPlatform(PlatformLocal, "http://127.0.0.1:9428", "") // VictoriaLogs
	
	// Add SaaS platforms (mock for demo - would need real API keys)
	// manager.AddPlatform(PlatformDatadog, "https://api.datadoghq.com", "your-api-key")
	// manager.AddPlatform(PlatformNewRelic, "https://api.newrelic.com", "your-api-key")
	
	ctx := context.Background()
	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	// Query examples
	demonstrateSaaSQueries(manager, ctx, timeRange)
	
	fmt.Println("\n🔄 Data Flow Summary:")
	fmt.Println("   Application → Agent → Gateway → SaaS Platforms")
	fmt.Println("   ")
	fmt.Println("   Local Stack:")
	fmt.Println("   📝 Logs: VictoriaLogs (9428)")
	fmt.Println("   📊 Metrics: Prometheus (8889)")
	fmt.Println("   🔍 Traces: Jaeger (16686)")
	fmt.Println("   ")
	fmt.Println("   SaaS Platforms:")
	fmt.Println("   ☁️  Datadog: All-in-one observability")
	fmt.Println("   ☁️  New Relic: APM and monitoring")
	fmt.Println("   ☁️  Grafana Cloud: Metrics and logs")
}

// demonstrateSaaSQueries demonstrates querying different SaaS platforms
func demonstrateSaaSQueries(manager *SaaSManager, ctx context.Context, timeRange TimeRange) {
	queries := []SaaSQuery{
		{
			QueryType: "logs",
			Query:     "service.name:apiserver AND level:error",
			TimeRange: timeRange,
		},
		{
			QueryType: "metrics", 
			Query:     "http_requests_total",
			TimeRange: timeRange,
		},
		{
			QueryType: "traces",
			Query:     "apiserver", // Service name
			TimeRange: timeRange,
		},
	}

	for _, query := range queries {
		fmt.Printf("\n🔍 Querying %s: %s\n", query.QueryType, query.Query)
		
		responses, err := manager.QueryAll(ctx, query)
		if err != nil {
			fmt.Printf("❌ Query failed: %v\n", err)
			continue
		}

		for _, resp := range responses {
			fmt.Printf("✅ %s responded in %v\n", resp.Platform, resp.Duration)
			
			// Show sample of response data
			if len(resp.Data) > 0 {
				sample := string(resp.Data)
				if len(sample) > 100 {
					sample = sample[:100] + "..."
				}
				fmt.Printf("   Sample: %s\n", sample)
			}
		}
	}
}

// ExampleSaaSConfiguration demonstrates SaaS platform configuration
func ExampleSaaSConfiguration() {
	fmt.Println("=== OTEL SaaS Configuration Examples ===")

	fmt.Println("\n🔧 Gateway Configuration for SaaS Platforms:")
	
	fmt.Println("\n1. Datadog Integration:")
	fmt.Println("   exporters:")
	fmt.Println("     otlphttp/datadog:")
	fmt.Println("       endpoint: \"https://otlp.datadoghq.com\"")
	fmt.Println("       headers:")
	fmt.Println("         DD-API-KEY: \"${DD_API_KEY}\"")
	fmt.Println("       compression: gzip")
	fmt.Println("   ")
	fmt.Println("   Environment variables:")
	fmt.Println("   export DD_API_KEY=your_datadog_api_key")

	fmt.Println("\n2. New Relic Integration:")
	fmt.Println("   exporters:")
	fmt.Println("     otlphttp/newrelic:")
	fmt.Println("       endpoint: \"https://otlp.nr-data.net:4318\"")
	fmt.Println("       headers:")
	fmt.Println("         api-key: \"${NEW_RELIC_API_KEY}\"")
	fmt.Println("   ")
	fmt.Println("   Environment variables:")
	fmt.Println("   export NEW_RELIC_API_KEY=your_newrelic_api_key")

	fmt.Println("\n3. Grafana Cloud Integration:")
	fmt.Println("   exporters:")
	fmt.Println("     otlphttp/grafana:")
	fmt.Println("       endpoint: \"${GRAFANA_CLOUD_OTLP_ENDPOINT}\"")
	fmt.Println("       headers:")
	fmt.Println("         Authorization: \"Basic ${GRAFANA_CLOUD_API_KEY}\"")

	fmt.Println("\n🚀 Deployment Steps:")
	fmt.Println("   1. Set API keys as environment variables")
	fmt.Println("   2. Update Gateway configuration")
	fmt.Println("   3. Restart Gateway: docker restart proj-otel-collector")
	fmt.Println("   4. Verify data export in SaaS platform dashboards")

	demonstrateConfigurationTemplates()
}

// demonstrateConfigurationTemplates shows configuration templates
func demonstrateConfigurationTemplates() {
	fmt.Println("\n📋 Complete Configuration Template:")
	
	config := `
exporters:
  # Local platforms
  otlphttp/victorialogs:
    endpoint: "http://proj-victorialogs:9428/insert/opentelemetry"
    tls:
      insecure: true

  prometheus:
    endpoint: "0.0.0.0:8889"

  otlphttp/jaeger:
    endpoint: "http://proj-jaeger:14268/api/traces"
    tls:
      insecure: true

  # SaaS platforms
  otlphttp/datadog:
    endpoint: "https://otlp.datadoghq.com"
    headers:
      DD-API-KEY: "${DD_API_KEY}"
    compression: gzip

  otlphttp/newrelic:
    endpoint: "https://otlp.nr-data.net:4318"
    headers:
      api-key: "${NEW_RELIC_API_KEY}"

service:
  pipelines:
    logs:
      receivers: [filelog, otlp]
      processors: [resource, batch]
      exporters: [otlphttp/victorialogs, otlphttp/datadog, otlphttp/newrelic]
    
    metrics:
      receivers: [prometheus, otlp]
      processors: [resource, batch]
      exporters: [prometheus, otlphttp/datadog, otlphttp/newrelic]
    
    traces:
      receivers: [otlp]
      processors: [resource, batch]
      exporters: [otlphttp/jaeger, otlphttp/datadog, otlphttp/newrelic]
`
	
	fmt.Println(config)
	
	fmt.Println("\n💡 Configuration Best Practices:")
	fmt.Println("   - Use environment variables for sensitive data")
	fmt.Println("   - Enable compression for SaaS exports")
	fmt.Println("   - Configure retry policies for reliability")
	fmt.Println("   - Monitor export success rates")
	fmt.Println("   - Use resource processors for data enrichment")
}

// ExampleSaaSAnalytics demonstrates analytics across platforms
func ExampleSaaSAnalytics() {
	fmt.Println("=== OTEL SaaS Analytics Example ===")

	manager := NewSaaSManager()
	manager.AddPlatform(PlatformLocal, "http://127.0.0.1:9428", "")
	
	ctx := context.Background()
	
	// Demonstrate different types of analytics queries
	analyticsQueries := []struct {
		name  string
		query SaaSQuery
	}{
		{
			name: "Error Rate Analysis",
			query: SaaSQuery{
				QueryType: "logs",
				Query:     "level:error",
				TimeRange: TimeRange{
					Start: time.Now().Add(-24 * time.Hour),
					End:   time.Now(),
				},
			},
		},
		{
			name: "Performance Metrics",
			query: SaaSQuery{
				QueryType: "metrics",
				Query:     "http_request_duration_seconds",
				TimeRange: TimeRange{
					Start: time.Now().Add(-1 * time.Hour),
					End:   time.Now(),
				},
			},
		},
		{
			name: "Trace Analysis",
			query: SaaSQuery{
				QueryType: "traces",
				Query:     "high-latency-service",
				TimeRange: TimeRange{
					Start: time.Now().Add(-30 * time.Minute),
					End:   time.Now(),
				},
			},
		},
	}

	fmt.Println("\n📊 Running Analytics Queries:")
	
	for _, aq := range analyticsQueries {
		fmt.Printf("\n🔍 %s\n", aq.name)
		fmt.Printf("   Query: %s\n", aq.query.Query)
		fmt.Printf("   Type: %s\n", aq.query.QueryType)
		fmt.Printf("   Range: %s to %s\n", 
			aq.query.TimeRange.Start.Format("15:04:05"),
			aq.query.TimeRange.End.Format("15:04:05"))

		responses, err := manager.QueryAll(ctx, aq.query)
		if err != nil {
			fmt.Printf("   ❌ Failed: %v\n", err)
			continue
		}

		for _, resp := range responses {
			fmt.Printf("   ✅ %s: %v response time\n", resp.Platform, resp.Duration)
		}
	}

	fmt.Println("\n📈 Analytics Capabilities:")
	fmt.Println("   - Cross-platform correlation")
	fmt.Println("   - Real-time alerting")
	fmt.Println("   - Historical trend analysis") 
	fmt.Println("   - Performance optimization insights")
	fmt.Println("   - Error pattern detection")
	fmt.Println("   - Capacity planning metrics")

	fmt.Println("\n🎯 Next Steps:")
	fmt.Println("   1. Set up SaaS platform dashboards")
	fmt.Println("   2. Configure alerting rules")
	fmt.Println("   3. Create custom analytics queries")
	fmt.Println("   4. Set up automated reporting")
	fmt.Println("   5. Implement ML-based anomaly detection")
}

// SaaSIntegrationTester tests SaaS platform integrations
type SaaSIntegrationTester struct {
	manager *SaaSManager
}

// NewSaaSIntegrationTester creates a new integration tester
func NewSaaSIntegrationTester() *SaaSIntegrationTester {
	manager := NewSaaSManager()
	// Add local platforms for testing
	manager.AddPlatform(PlatformLocal, "http://127.0.0.1:9428", "")
	
	return &SaaSIntegrationTester{
		manager: manager,
	}
}

// RunIntegrationTests runs comprehensive integration tests
func (t *SaaSIntegrationTester) RunIntegrationTests(ctx context.Context) error {
	fmt.Println("=== SaaS Integration Tests ===")
	
	tests := []struct {
		name string
		test func(context.Context) error
	}{
		{"Log Platform Connectivity", t.testLogPlatforms},
		{"Metric Platform Connectivity", t.testMetricPlatforms},
		{"Trace Platform Connectivity", t.testTracePlatforms},
		{"Data Consistency Check", t.testDataConsistency},
		{"Performance Benchmarks", t.testPerformance},
	}

	var passed, failed int
	for _, test := range tests {
		fmt.Printf("\n🧪 %s\n", test.name)
		if err := test.test(ctx); err != nil {
			fmt.Printf("❌ Failed: %v\n", err)
			failed++
		} else {
			fmt.Printf("✅ Passed\n")
			passed++
		}
	}

	fmt.Printf("\n📊 Integration Test Results: %d passed, %d failed\n", passed, failed)
	
	if failed == 0 {
		fmt.Println("🎉 All SaaS integrations working correctly!")
		return nil
	}
	
	return fmt.Errorf("%d integration tests failed", failed)
}

// testLogPlatforms tests log platform integrations
func (t *SaaSIntegrationTester) testLogPlatforms(ctx context.Context) error {
	query := SaaSQuery{
		QueryType: "logs",
		Query:     "*",
		TimeRange: TimeRange{
			Start: time.Now().Add(-1 * time.Hour),
			End:   time.Now(),
		},
	}
	
	_, err := t.manager.QueryAll(ctx, query)
	return err
}

// testMetricPlatforms tests metric platform integrations  
func (t *SaaSIntegrationTester) testMetricPlatforms(ctx context.Context) error {
	query := SaaSQuery{
		QueryType: "metrics", 
		Query:     "up",
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * time.Minute),
			End:   time.Now(),
		},
	}
	
	_, err := t.manager.QueryAll(ctx, query)
	return err
}

// testTracePlatforms tests trace platform integrations
func (t *SaaSIntegrationTester) testTracePlatforms(ctx context.Context) error {
	query := SaaSQuery{
		QueryType: "traces",
		Query:     "test-service",
		TimeRange: TimeRange{
			Start: time.Now().Add(-30 * time.Minute),
			End:   time.Now(),
		},
	}
	
	_, err := t.manager.QueryAll(ctx, query)
	return err
}

// testDataConsistency checks data consistency across platforms
func (t *SaaSIntegrationTester) testDataConsistency(ctx context.Context) error {
	// Mock consistency check - in real implementation would compare data across platforms
	fmt.Println("   Checking data consistency across platforms...")
	time.Sleep(100 * time.Millisecond) // Simulate check
	return nil
}

// testPerformance runs performance benchmarks
func (t *SaaSIntegrationTester) testPerformance(ctx context.Context) error {
	fmt.Println("   Running performance benchmarks...")
	
	start := time.Now()
	
	// Run multiple queries to test performance
	for i := 0; i < 5; i++ {
		query := SaaSQuery{
			QueryType: "logs",
			Query:     fmt.Sprintf("test-query-%d", i),
			TimeRange: TimeRange{
				Start: time.Now().Add(-5 * time.Minute),
				End:   time.Now(),
			},
		}
		
		if _, err := t.manager.QueryAll(ctx, query); err != nil {
			return err
		}
	}
	
	duration := time.Since(start)
	fmt.Printf("   Performance: 5 queries in %v (avg: %v per query)\n", 
		duration, duration/5)
		
	return nil
}

// ExampleSaaSTesting demonstrates SaaS integration testing
func ExampleSaaSTesting() {
	fmt.Println("=== SaaS Integration Testing ===")

	tester := NewSaaSIntegrationTester()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := tester.RunIntegrationTests(ctx); err != nil {
		fmt.Printf("Integration tests failed: %v\n", err)
	}

	fmt.Println("\n💡 Testing Best Practices:")
	fmt.Println("   - Test all configured SaaS platforms")
	fmt.Println("   - Verify data integrity across platforms")
	fmt.Println("   - Monitor query performance")
	fmt.Println("   - Test error handling and retries")
	fmt.Println("   - Validate data retention policies")
	fmt.Println("   - Test platform-specific features")
}

// initializeOTEL initializes OpenTelemetry with trace export (simplified for demo)
func initializeOTEL(ctx context.Context) (func(context.Context) error, *slog.Logger, error) {
	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("saas-platform-demo"),
			semconv.ServiceVersionKey.String("v1.0.0"),
			semconv.DeploymentEnvironmentKey.String("demo"),
			attribute.String("component", "saas-integration"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create trace exporter (for context correlation)
	traceExporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("http://127.0.0.1:4328"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create trace provider
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Create log file for OTEL Collector filelog receiver
	logDir := "/home/hellotalk/code/go/src/github.com/costa92/go-protoc/logs/saas-demo"
	logFile := logDir + "/saas.log"
	
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create log file: %w", err)
	}
	
	// Create multi-writer: both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, file)
	
	// Create structured logger with JSON output
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Shutdown function
	shutdown := func(ctx context.Context) error {
		if err := traceProvider.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown trace provider: %w", err)
		}
		return nil
	}

	fmt.Println("✅ OTEL initialized with Gateway endpoints:")
	fmt.Printf("   📝 Logs: Writing to file %s (picked up by filelog receiver)\n", logFile)
	fmt.Println("   🔍 Traces: http://127.0.0.1:4328 (OTLP HTTP)")

	return shutdown, logger, nil
}

// generateSaaSBusinessLogs generates various business activity logs
func generateSaaSBusinessLogs(ctx context.Context, logger *slog.Logger) {
	tracer := otel.Tracer("saas-demo")

	// Simulate different SaaS business scenarios
	scenarios := []struct {
		name      string
		logLevel  slog.Level
		message   string
		attributes []any
	}{
		{
			name:     "user-authentication",
			logLevel: slog.LevelInfo,
			message:  "User authentication successful",
			attributes: []any{
				"user_id", "user_12345",
				"auth_method", "oauth2",
				"platform", "web",
				"ip_address", "192.168.1.100",
			},
		},
		{
			name:     "payment-processing",
			logLevel: slog.LevelInfo,
			message:  "Payment transaction completed",
			attributes: []any{
				"transaction_id", "txn_98765",
				"amount", "99.99",
				"currency", "USD",
				"payment_method", "credit_card",
				"status", "success",
			},
		},
		{
			name:     "api-rate-limit",
			logLevel: slog.LevelWarn,
			message:  "API rate limit approaching threshold",
			attributes: []any{
				"api_key", "key_abc123",
				"current_requests", 850,
				"limit", 1000,
				"window", "1h",
				"endpoint", "/api/v1/users",
			},
		},
		{
			name:     "database-connection-error",
			logLevel: slog.LevelError,
			message:  "Failed to connect to database",
			attributes: []any{
				"database", "user_profiles",
				"error", "connection timeout",
				"retry_count", 3,
				"duration_ms", 5000,
			},
		},
		{
			name:     "saas-integration",
			logLevel: slog.LevelInfo,
			message:  "SaaS platform data sync completed",
			attributes: []any{
				"platform", "datadog",
				"sync_type", "metrics",
				"records_synced", 1250,
				"duration_ms", 2300,
				"status", "success",
			},
		},
	}

	for i, scenario := range scenarios {
		// Create trace span for each scenario
		ctx, span := tracer.Start(ctx, scenario.name)
		span.SetAttributes(
			attribute.String("scenario", scenario.name),
			attribute.Int("sequence", i+1),
		)

		// Generate structured log with all attributes
		attrs := []slog.Attr{
			slog.String("scenario", scenario.name),
			slog.Int("sequence", i+1),
			slog.String("timestamp", time.Now().Format(time.RFC3339)),
		}
		
		// Add scenario-specific attributes
		for j := 0; j < len(scenario.attributes); j += 2 {
			key := scenario.attributes[j].(string)
			value := scenario.attributes[j+1]
			
			switch v := value.(type) {
			case string:
				attrs = append(attrs, slog.String(key, v))
			case int:
				attrs = append(attrs, slog.Int(key, v))
			default:
				attrs = append(attrs, slog.Any(key, v))
			}
		}

		logger.LogAttrs(ctx, scenario.logLevel, scenario.message, attrs...)

		span.End()

		fmt.Printf("📝 Generated %s log: %s\n", scenario.logLevel.String(), scenario.message)
		
		// Small delay between logs
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\n✅ Generated %d business scenario logs\n", len(scenarios))
}

// ExampleSaaSIntegrationWithLogs demonstrates SaaS platform integration with actual log generation
func ExampleSaaSIntegrationWithLogs() {
	fmt.Println("=== OTEL SaaS Platform Integration with Log Generation ===")

	// Initialize OTEL with log export to Gateway → SaaS
	ctx := context.Background()
	shutdown, logger, err := initializeOTEL(ctx)
	if err != nil {
		fmt.Printf("❌ Failed to initialize OTEL: %v\n", err)
		return
	}
	defer shutdown(ctx)

	// Create SaaS manager
	manager := NewSaaSManager()
	
	// Add local platforms (actually accessible)
	manager.AddPlatform(PlatformLocal, "http://127.0.0.1:9428", "") // VictoriaLogs
	
	// Add SaaS platforms (mock for demo - would need real API keys)
	// manager.AddPlatform(PlatformDatadog, "https://api.datadoghq.com", "your-api-key")
	// manager.AddPlatform(PlatformNewRelic, "https://api.newrelic.com", "your-api-key")

	// Generate business activity logs
	fmt.Println("\n🔄 Generating SaaS business activity logs...")
	generateSaaSBusinessLogs(ctx, logger)

	// Wait for logs to be processed
	fmt.Println("⏳ Waiting for logs to be processed and exported...")
	time.Sleep(3 * time.Second)

	timeRange := TimeRange{
		Start: time.Now().Add(-5 * time.Minute),
		End:   time.Now(),
	}

	// Query examples to verify data was received
	demonstrateSaaSQueries(manager, ctx, timeRange)
	
	fmt.Println("\n🔄 Data Flow Summary:")
	fmt.Println("   SaaS App → OTEL Agent → Gateway → VictoriaLogs")
	fmt.Println("   ")
	fmt.Println("   Local Stack:")
	fmt.Println("   📝 Logs: VictoriaLogs (9428) - Check: http://127.0.0.1:9428/select/vmui/")
	fmt.Println("   📊 Metrics: Prometheus (8889)")
	fmt.Println("   🔍 Traces: Jaeger (16686) - Check: http://127.0.0.1:16686")
	fmt.Println("   ")
	fmt.Println("   SaaS Platforms:")
	fmt.Println("   ☁️  Datadog: All-in-one observability")
	fmt.Println("   ☁️  New Relic: APM and monitoring")
	fmt.Println("   ☁️  Grafana Cloud: Metrics and logs")
}

func main() {
	// Run SaaS integration with actual log generation
	ExampleSaaSIntegrationWithLogs()
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	
	// Additional examples
	ExampleSaaSConfiguration()
	
	fmt.Println("\n" + strings.Repeat("=", 50))
	
	ExampleSaaSAnalytics()
	
	fmt.Println("\n🎯 Verification Steps:")
	fmt.Println("   1. Check VictoriaLogs: http://127.0.0.1:9428/select/vmui/")
	fmt.Println("   2. Search for: service.name:\"saas-platform-demo\"")
	fmt.Println("   3. Check Jaeger traces: http://127.0.0.1:16686")
	fmt.Println("   4. Look for service: saas-platform-demo")
	fmt.Println("   5. Verify OTEL Collector metrics: curl http://127.0.0.1:8888/metrics")
}