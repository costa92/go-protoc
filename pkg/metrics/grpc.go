package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// GRPCCollector gRPC指标收集器
type GRPCCollector struct {
	config   *Config
	registry prometheus.Registerer
	enabled  atomic.Bool
	running  atomic.Bool

	// 指标
	callsTotal          *prometheus.CounterVec
	callDuration        *prometheus.HistogramVec
	messageSent         *prometheus.CounterVec
	messageReceived     *prometheus.CounterVec
	messageSentSize     *prometheus.HistogramVec
	messageReceivedSize *prometheus.HistogramVec
	activeStreams       *prometheus.GaugeVec

	// 内部状态
	mu sync.RWMutex
}

// NewGRPCCollector 创建gRPC指标收集器
func NewGRPCCollector(config *Config) *GRPCCollector {
	if config == nil {
		config = NewConfig()
	}

	collector := &GRPCCollector{
		config:   config,
		registry: config.Registry,
	}

	collector.initMetrics()
	return collector
}

// initMetrics 初始化指标
func (g *GRPCCollector) initMetrics() {
	namespace := g.config.Namespace
	subsystem := "grpc"

	// 调用总数
	g.callsTotal = promauto.With(g.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "calls_total",
			Help:      "Total number of gRPC calls",
		},
		[]string{"service", "method", "code", "type"},
	)

	// 调用持续时间
	g.callDuration = promauto.With(g.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "call_duration_seconds",
			Help:      "gRPC call duration in seconds",
			Buckets:   g.config.GRPC.Buckets,
		},
		[]string{"service", "method", "code", "type"},
	)

	// 发送消息数
	g.messageSent = promauto.With(g.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "messages_sent_total",
			Help:      "Total number of gRPC messages sent",
		},
		[]string{"service", "method", "type"},
	)

	// 接收消息数
	g.messageReceived = promauto.With(g.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "messages_received_total",
			Help:      "Total number of gRPC messages received",
		},
		[]string{"service", "method", "type"},
	)

	// 发送消息大小
	g.messageSentSize = promauto.With(g.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "message_sent_size_bytes",
			Help:      "Size of gRPC messages sent in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 512MB
		},
		[]string{"service", "method", "type"},
	)

	// 接收消息大小
	g.messageReceivedSize = promauto.With(g.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "message_received_size_bytes",
			Help:      "Size of gRPC messages received in bytes",
			Buckets:   prometheus.ExponentialBuckets(1024, 2, 10), // 1KB to 512MB
		},
		[]string{"service", "method", "type"},
	)

	// 活跃流数量
	g.activeStreams = promauto.With(g.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "active_streams",
			Help:      "Number of active gRPC streams",
		},
		[]string{"service", "method", "type"},
	)
}

// Start 启动收集器
func (g *GRPCCollector) Start() error {
	if !g.running.CompareAndSwap(false, true) {
		return nil // 已经运行
	}

	if !g.config.GRPC.Enabled {
		g.running.Store(false)
		return nil
	}

	g.enabled.Store(true)
	return nil
}

// Stop 停止收集器
func (g *GRPCCollector) Stop() error {
	if !g.running.CompareAndSwap(true, false) {
		return nil // 未运行
	}

	g.enabled.Store(false)
	return nil
}

// IsRunning 检查是否运行
func (g *GRPCCollector) IsRunning() bool {
	return g.running.Load()
}

// GetType 获取收集器类型
func (g *GRPCCollector) GetType() MetricType {
	return MetricTypeGRPC
}

// GetStatus 获取状态
func (g *GRPCCollector) GetStatus() map[string]interface{} {
	return map[string]interface{}{
		"type":    string(MetricTypeGRPC),
		"enabled": g.enabled.Load(),
		"running": g.running.Load(),
		"config":  g.config.GRPC,
	}
}

// RegisterTarget 注册监控目标
func (g *GRPCCollector) RegisterTarget(name string, target interface{}) error {
	// gRPC收集器通常通过拦截器自动收集，不需要手动注册
	return nil
}

// UnregisterTarget 取消注册监控目标
func (g *GRPCCollector) UnregisterTarget(name string) error {
	// gRPC收集器通常通过拦截器自动收集，不需要手动注册
	return nil
}

// RecordCall 记录调用
func (g *GRPCCollector) RecordCall(service, method string, code string, duration time.Duration) {
	if !g.enabled.Load() {
		return
	}

	callType := g.getCallType(service, method)

	g.callsTotal.WithLabelValues(service, method, code, callType).Inc()
	g.callDuration.WithLabelValues(service, method, code, callType).Observe(duration.Seconds())
}

// RecordMessageSent 记录发送消息
func (g *GRPCCollector) RecordMessageSent(service, method string, size int64) {
	if !g.enabled.Load() {
		return
	}

	callType := g.getCallType(service, method)

	g.messageSent.WithLabelValues(service, method, callType).Inc()
	g.messageSentSize.WithLabelValues(service, method, callType).Observe(float64(size))
}

// RecordMessageReceived 记录接收消息
func (g *GRPCCollector) RecordMessageReceived(service, method string, size int64) {
	if !g.enabled.Load() {
		return
	}

	callType := g.getCallType(service, method)

	g.messageReceived.WithLabelValues(service, method, callType).Inc()
	g.messageReceivedSize.WithLabelValues(service, method, callType).Observe(float64(size))
}

// IncrementActiveStreams 增加活跃流数量
func (g *GRPCCollector) IncrementActiveStreams(service, method string) {
	if !g.enabled.Load() {
		return
	}

	callType := g.getCallType(service, method)
	g.activeStreams.WithLabelValues(service, method, callType).Inc()
}

// DecrementActiveStreams 减少活跃流数量
func (g *GRPCCollector) DecrementActiveStreams(service, method string) {
	if !g.enabled.Load() {
		return
	}

	callType := g.getCallType(service, method)
	g.activeStreams.WithLabelValues(service, method, callType).Dec()
}

// getCallType 根据方法名推断调用类型
func (g *GRPCCollector) getCallType(service, method string) string {
	// 这里可以根据实际需要实现更复杂的逻辑
	// 例如通过方法签名或者配置来确定调用类型
	return "unary" // 默认为一元调用
}

// GetGRPCMetrics 获取gRPC指标接口
func (g *GRPCCollector) GetGRPCMetrics() GRPCMetrics {
	return g
}

// GRPCInterceptor gRPC拦截器接口
type GRPCInterceptor interface {
	// UnaryServerInterceptor 一元服务器拦截器
	UnaryServerInterceptor() interface{}

	// StreamServerInterceptor 流服务器拦截器
	StreamServerInterceptor() interface{}

	// UnaryClientInterceptor 一元客户端拦截器
	UnaryClientInterceptor() interface{}

	// StreamClientInterceptor 流客户端拦截器
	StreamClientInterceptor() interface{}
}

// grpcInterceptor gRPC拦截器实现
type grpcInterceptor struct {
	collector *GRPCCollector
}

// NewGRPCInterceptor 创建gRPC拦截器
func NewGRPCInterceptor(collector *GRPCCollector) GRPCInterceptor {
	return &grpcInterceptor{
		collector: collector,
	}
}

// UnaryServerInterceptor 一元服务器拦截器
func (i *grpcInterceptor) UnaryServerInterceptor() interface{} {
	// 返回具体的gRPC一元服务器拦截器
	// 这里需要根据具体的gRPC版本实现
	return nil
}

// StreamServerInterceptor 流服务器拦截器
func (i *grpcInterceptor) StreamServerInterceptor() interface{} {
	// 返回具体的gRPC流服务器拦截器
	// 这里需要根据具体的gRPC版本实现
	return nil
}

// UnaryClientInterceptor 一元客户端拦截器
func (i *grpcInterceptor) UnaryClientInterceptor() interface{} {
	// 返回具体的gRPC一元客户端拦截器
	// 这里需要根据具体的gRPC版本实现
	return nil
}

// StreamClientInterceptor 流客户端拦截器
func (i *grpcInterceptor) StreamClientInterceptor() interface{} {
	// 返回具体的gRPC流客户端拦截器
	// 这里需要根据具体的gRPC版本实现
	return nil
}
