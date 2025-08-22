package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeHTTP     MetricType = "http"
	MetricTypeGRPC     MetricType = "grpc"
	MetricTypeDatabase MetricType = "database"
	MetricTypeRedis    MetricType = "redis"
	MetricTypeCustom   MetricType = "custom"
)

// Collector 指标收集器接口
type Collector interface {
	// Start 启动收集器
	Start() error

	// Stop 停止收集器
	Stop() error

	// IsRunning 检查是否运行
	IsRunning() bool

	// GetType 获取收集器类型
	GetType() MetricType

	// GetStatus 获取状态
	GetStatus() map[string]interface{}

	// RegisterTarget 注册监控目标
	RegisterTarget(name string, target interface{}) error

	// UnregisterTarget 取消注册监控目标
	UnregisterTarget(name string) error
}

// Registry 指标注册表接口
type Registry interface {
	// Register 注册指标
	Register(collector prometheus.Collector) error

	// Unregister 取消注册指标
	Unregister(collector prometheus.Collector) bool

	// Gather 收集指标
	Gather() ([]*dto.MetricFamily, error)

	// MustRegister 必须注册指标
	MustRegister(collectors ...prometheus.Collector)

	// GetRegisterer 获取注册器
	GetRegisterer() prometheus.Registerer

	// GetGatherer 获取收集器
	GetGatherer() prometheus.Gatherer
}

// MetricsManager 指标管理器接口
type MetricsManager interface {
	// Start 启动管理器
	Start() error

	// Stop 停止管理器
	Stop() error

	// IsRunning 检查是否运行
	IsRunning() bool

	// GetCollector 获取收集器
	GetCollector(metricType MetricType) (Collector, bool)

	// RegisterTarget 注册监控目标
	RegisterTarget(metricType MetricType, name string, target interface{}) error

	// UnregisterTarget 取消注册监控目标
	UnregisterTarget(metricType MetricType, name string) error

	// GetStatus 获取状态
	GetStatus() map[string]interface{}
}

// HTTPMetrics HTTP指标接口
type HTTPMetrics interface {
	// RecordRequest 记录请求
	RecordRequest(method, path string, statusCode int, duration time.Duration)

	// RecordRequestSize 记录请求大小
	RecordRequestSize(method, path string, size int64)

	// RecordResponseSize 记录响应大小
	RecordResponseSize(method, path string, statusCode int, size int64)

	// IncrementActiveConnections 增加活跃连接数
	IncrementActiveConnections()

	// DecrementActiveConnections 减少活跃连接数
	DecrementActiveConnections()
}

// GRPCMetrics gRPC指标接口
type GRPCMetrics interface {
	// RecordCall 记录调用
	RecordCall(service, method string, code string, duration time.Duration)

	// RecordMessageSent 记录发送消息
	RecordMessageSent(service, method string, size int64)

	// RecordMessageReceived 记录接收消息
	RecordMessageReceived(service, method string, size int64)

	// IncrementActiveStreams 增加活跃流数量
	IncrementActiveStreams(service, method string)

	// DecrementActiveStreams 减少活跃流数量
	DecrementActiveStreams(service, method string)
}

// DatabaseMetrics 数据库指标接口
type DatabaseMetrics interface {
	// RecordQuery 记录查询
	RecordQuery(database, operation, table string, duration time.Duration, success bool)

	// RecordConnection 记录连接状态
	RecordConnection(dbName string, state string, count int)

	// RecordTransaction 记录事务
	RecordTransaction(database, operation string, duration time.Duration, success bool)
}

// RedisMetrics Redis指标接口
type RedisMetrics interface {
	// RecordCommand 记录命令
	RecordCommand(instance, command string, duration time.Duration, success bool)

	// RecordConnection 记录连接状态
	RecordConnection(state string, count int)

	// RecordKeyOperation 记录键操作
	RecordKeyOperation(instance, operation, keyType string, count int)
}

// CustomMetrics 自定义指标接口
type CustomMetrics interface {
	// CreateCounter 创建计数器
	CreateCounter(name, help string, labels []string) (*prometheus.CounterVec, error)

	// CreateGauge 创建仪表盘
	CreateGauge(name, help string, labels []string) (*prometheus.GaugeVec, error)

	// CreateHistogram 创建直方图
	CreateHistogram(name, help string, labels []string, buckets []float64) (*prometheus.HistogramVec, error)

	// CreateSummary 创建摘要
	CreateSummary(name, help string, labels []string, objectives map[float64]float64) (*prometheus.SummaryVec, error)
}

// HealthCheckItem 健康检查项定义
type HealthCheckItem struct {
	Name       string
	Message    string
	LastCheck  time.Time
	CheckCount int64
	ErrorCount int64
}

// Logger 日志接口
type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	With(key string, value interface{}) Logger
}
