package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// 全局指标变量
var (
	// HTTPRequestsTotal 记录HTTP请求总数
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "HTTP请求总数",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration 记录HTTP请求耗时
	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP请求耗时（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// GRPCRequestsTotal 记录gRPC请求总数
	GRPCRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "gRPC请求总数",
		},
		[]string{"method", "status"},
	)

	// GRPCRequestDuration 记录gRPC请求耗时
	GRPCRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "gRPC请求耗时（秒）",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)
