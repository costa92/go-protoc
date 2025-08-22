package metrics

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPCollector HTTP指标收集器
type HTTPCollector struct {
	config   *Config
	registry prometheus.Registerer
	enabled  atomic.Bool
	running  atomic.Bool

	// 指标
	requestsTotal     *prometheus.CounterVec
	requestDuration   *prometheus.HistogramVec
	requestSize       *prometheus.HistogramVec
	responseSize      *prometheus.HistogramVec
	activeConnections prometheus.Gauge
	requestsInFlight  prometheus.Gauge

	// 内部状态
	mu sync.RWMutex
}

// NewHTTPCollector 创建HTTP指标收集器
func NewHTTPCollector(config *Config) *HTTPCollector {
	if config == nil {
		config = NewConfig()
	}

	collector := &HTTPCollector{
		config:   config,
		registry: config.Registry,
	}

	collector.initMetrics()
	return collector
}

// initMetrics 初始化指标
func (h *HTTPCollector) initMetrics() {
	namespace := h.config.Namespace
	subsystem := "http"

	// 请求总数
	h.requestsTotal = promauto.With(h.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	// 请求持续时间
	h.requestDuration = promauto.With(h.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   h.config.HTTP.Buckets,
		},
		[]string{"method", "path", "status_code"},
	)

	// 请求大小
	h.requestSize = promauto.With(h.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 512MB
		},
		[]string{"method", "path"},
	)

	// 响应大小
	h.responseSize = promauto.With(h.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 512MB
		},
		[]string{"method", "path", "status_code"},
	)

	// 活跃连接数
	h.activeConnections = promauto.With(h.registry).NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "active_connections",
			Help:      "Number of active HTTP connections",
		},
	)

	// 正在处理的请求数
	h.requestsInFlight = promauto.With(h.registry).NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
	)
}

// Start 启动收集器
func (h *HTTPCollector) Start() error {
	if !h.running.CompareAndSwap(false, true) {
		return nil // 已经运行
	}

	if !h.config.HTTP.Enabled {
		h.running.Store(false)
		return nil
	}

	h.enabled.Store(true)
	return nil
}

// Stop 停止收集器
func (h *HTTPCollector) Stop() error {
	if !h.running.CompareAndSwap(true, false) {
		return nil // 未运行
	}

	h.enabled.Store(false)
	return nil
}

// IsRunning 检查是否运行
func (h *HTTPCollector) IsRunning() bool {
	return h.running.Load()
}

// GetType 获取收集器类型
func (h *HTTPCollector) GetType() MetricType {
	return MetricTypeHTTP
}

// GetStatus 获取状态
func (h *HTTPCollector) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"type":    string(MetricTypeHTTP),
		"enabled": h.enabled.Load(),
		"running": h.running.Load(),
		"config":  h.config.HTTP,
	}
}

// RegisterTarget 注册监控目标
func (h *HTTPCollector) RegisterTarget(name string, target interface{}) error {
	// HTTP收集器通常通过中间件自动收集，不需要手动注册
	return nil
}

// UnregisterTarget 取消注册监控目标
func (h *HTTPCollector) UnregisterTarget(name string) error {
	// HTTP收集器通常通过中间件自动收集，不需要手动注册
	return nil
}

// RecordRequest 记录请求
func (h *HTTPCollector) RecordRequest(method, path string, statusCode int, duration time.Duration) {
	if !h.enabled.Load() {
		return
	}

	// 检查是否需要排除此路径
	if h.shouldExcludePath(path) {
		return
	}

	statusStr := strconv.Itoa(statusCode)

	h.requestsTotal.WithLabelValues(method, path, statusStr).Inc()
	h.requestDuration.WithLabelValues(method, path, statusStr).Observe(duration.Seconds())
}

// RecordRequestSize 记录请求大小
func (h *HTTPCollector) RecordRequestSize(method, path string, size int64) {
	if !h.enabled.Load() {
		return
	}

	if h.shouldExcludePath(path) {
		return
	}

	h.requestSize.WithLabelValues(method, path).Observe(float64(size))
}

// RecordResponseSize 记录响应大小
func (h *HTTPCollector) RecordResponseSize(method, path string, statusCode int, size int64) {
	if !h.enabled.Load() {
		return
	}

	if h.shouldExcludePath(path) {
		return
	}

	statusStr := strconv.Itoa(statusCode)
	h.responseSize.WithLabelValues(method, path, statusStr).Observe(float64(size))
}

// IncrementActiveConnections 增加活跃连接数
func (h *HTTPCollector) IncrementActiveConnections() {
	if !h.enabled.Load() {
		return
	}
	h.activeConnections.Inc()
}

// DecrementActiveConnections 减少活跃连接数
func (h *HTTPCollector) DecrementActiveConnections() {
	if !h.enabled.Load() {
		return
	}
	h.activeConnections.Dec()
}

// IncrementRequestsInFlight 增加正在处理的请求数
func (h *HTTPCollector) IncrementRequestsInFlight() {
	if !h.enabled.Load() {
		return
	}
	h.requestsInFlight.Inc()
}

// DecrementRequestsInFlight 减少正在处理的请求数
func (h *HTTPCollector) DecrementRequestsInFlight() {
	if !h.enabled.Load() {
		return
	}
	h.requestsInFlight.Dec()
}

// shouldExcludePath 检查是否应该排除此路径
func (h *HTTPCollector) shouldExcludePath(path string) bool {
	for _, excludePath := range h.config.HTTP.ExcludePaths {
		if path == excludePath {
			return true
		}
	}
	return false
}

// GetHTTPMetrics 获取HTTP指标接口
func (h *HTTPCollector) GetHTTPMetrics() HTTPMetrics {
	return h
}

// HTTPMiddleware HTTP中间件接口
type HTTPMiddleware interface {
	// Wrap 包装HTTP处理器
	Wrap(next interface{}) interface{}
}

// httpMiddleware HTTP中间件实现
type httpMiddleware struct {
	collector *HTTPCollector
}

// NewHTTPMiddleware 创建HTTP中间件
func NewHTTPMiddleware(collector *HTTPCollector) HTTPMiddleware {
	return &httpMiddleware{
		collector: collector,
	}
}

// Wrap 包装HTTP处理器
func (m *httpMiddleware) Wrap(next interface{}) interface{} {
	// 这里可以根据具体的HTTP框架实现不同的中间件
	// 例如：gin, echo, net/http 等
	return next
}
