package metrics

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HealthStatus 健康状态
type HealthStatus int

const (
	HealthStatusUnknown HealthStatus = iota
	HealthStatusHealthy
	HealthStatusUnhealthy
	HealthStatusDegraded
)

// String 返回健康状态字符串
func (h HealthStatus) String() string {
	switch h {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusUnhealthy:
		return "unhealthy"
	case HealthStatusDegraded:
		return "degraded"
	default:
		return "unknown"
	}
}

// HealthCheck 健康检查项
type HealthCheck struct {
	Name       string
	Status     HealthStatus
	Message    string
	LastCheck  time.Time
	CheckCount int64
	ErrorCount int64
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// Check 执行健康检查
	Check(ctx context.Context) error

	// RecordHealthy 记录健康状态
	RecordHealthy(component string)

	// RecordUnhealthy 记录不健康状态
	RecordUnhealthy(component string, message string)

	// RecordDegraded 记录降级状态
	RecordDegraded(component string, message string)

	// GetStatus 获取健康状态
	GetStatus() map[string]*HealthCheck

	// GetOverallStatus 获取整体健康状态
	GetOverallStatus() HealthStatus

	// IsHealthy 检查是否健康
	IsHealthy() bool
}

// defaultHealthChecker 默认健康检查器实现
type defaultHealthChecker struct {
	config   HealthCheckConfig
	registry prometheus.Registerer
	enabled  atomic.Bool

	// 健康检查项
	checks sync.Map // map[string]*HealthCheck

	// 指标
	healthStatus  *prometheus.GaugeVec
	checkTotal    *prometheus.CounterVec
	checkDuration *prometheus.HistogramVec
	lastCheckTime *prometheus.GaugeVec

	mu sync.RWMutex
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(config HealthCheckConfig) HealthChecker {
	checker := &defaultHealthChecker{
		config:   config,
		registry: prometheus.DefaultRegisterer,
	}

	checker.initMetrics()
	checker.enabled.Store(config.Enabled)

	return checker
}

// initMetrics 初始化指标
func (h *defaultHealthChecker) initMetrics() {
	namespace := "health"

	// 健康状态指标
	h.healthStatus = promauto.With(h.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "status",
			Help:      "Health status of components (0=unknown, 1=healthy, 2=unhealthy, 3=degraded)",
		},
		[]string{"component"},
	)

	// 检查总数指标
	h.checkTotal = promauto.With(h.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "checks_total",
			Help:      "Total number of health checks performed",
		},
		[]string{"component", "status"},
	)

	// 检查持续时间指标
	h.checkDuration = promauto.With(h.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "check_duration_seconds",
			Help:      "Duration of health checks in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5},
		},
		[]string{"component"},
	)

	// 最后检查时间指标
	h.lastCheckTime = promauto.With(h.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "last_check_timestamp",
			Help:      "Timestamp of the last health check",
		},
		[]string{"component"},
	)
}

// Check 执行健康检查
func (h *defaultHealthChecker) Check(ctx context.Context) error {
	if !h.enabled.Load() {
		return nil
	}

	start := time.Now()
	defer func() {
		duration := time.Since(start)
		h.checkDuration.WithLabelValues("overall").Observe(duration.Seconds())
	}()

	// 这里可以添加具体的健康检查逻辑
	// 例如检查数据库连接、外部服务等

	return nil
}

// RecordHealthy 记录健康状态
func (h *defaultHealthChecker) RecordHealthy(component string) {
	if !h.enabled.Load() {
		return
	}

	h.recordStatus(component, HealthStatusHealthy, "")
}

// RecordUnhealthy 记录不健康状态
func (h *defaultHealthChecker) RecordUnhealthy(component string, message string) {
	if !h.enabled.Load() {
		return
	}

	h.recordStatus(component, HealthStatusUnhealthy, message)
}

// RecordDegraded 记录降级状态
func (h *defaultHealthChecker) RecordDegraded(component string, message string) {
	if !h.enabled.Load() {
		return
	}

	h.recordStatus(component, HealthStatusDegraded, message)
}

// recordStatus 记录状态
func (h *defaultHealthChecker) recordStatus(component string, status HealthStatus, message string) {
	now := time.Now()

	// 获取或创建健康检查项
	checkInterface, _ := h.checks.LoadOrStore(component, &HealthCheck{
		Name: component,
	})

	check := checkInterface.(*HealthCheck)

	// 更新状态
	check.Status = status
	check.Message = message
	check.LastCheck = now
	atomic.AddInt64(&check.CheckCount, 1)

	if status != HealthStatusHealthy {
		atomic.AddInt64(&check.ErrorCount, 1)
	}

	// 更新指标
	h.healthStatus.WithLabelValues(component).Set(float64(status))
	h.checkTotal.WithLabelValues(component, status.String()).Inc()
	h.lastCheckTime.WithLabelValues(component).Set(float64(now.Unix()))
}

// GetStatus 获取健康状态
func (h *defaultHealthChecker) GetStatus() map[string]*HealthCheck {
	result := make(map[string]*HealthCheck)

	h.checks.Range(func(key, value interface{}) bool {
		component := key.(string)
		check := value.(*HealthCheck)

		// 复制健康检查项
		result[component] = &HealthCheck{
			Name:       check.Name,
			Status:     check.Status,
			Message:    check.Message,
			LastCheck:  check.LastCheck,
			CheckCount: atomic.LoadInt64(&check.CheckCount),
			ErrorCount: atomic.LoadInt64(&check.ErrorCount),
		}

		return true
	})

	return result
}

// GetOverallStatus 获取整体健康状态
func (h *defaultHealthChecker) GetOverallStatus() HealthStatus {
	if !h.enabled.Load() {
		return HealthStatusUnknown
	}

	var hasUnhealthy, hasDegraded bool

	h.checks.Range(func(key, value interface{}) bool {
		check := value.(*HealthCheck)
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
		return true
	})

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// IsHealthy 检查是否健康
func (h *defaultHealthChecker) IsHealthy() bool {
	return h.GetOverallStatus() == HealthStatusHealthy
}

// HealthCheckRegistry 健康检查注册器
type HealthCheckRegistry struct {
	checker HealthChecker
	mu      sync.RWMutex
}

// NewHealthCheckRegistry 创建健康检查注册器
func NewHealthCheckRegistry(checker HealthChecker) *HealthCheckRegistry {
	return &HealthCheckRegistry{
		checker: checker,
	}
}

// RegisterHealthy 注册健康状态
func (r *HealthCheckRegistry) RegisterHealthy(component string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checker.RecordHealthy(component)
}

// RegisterUnhealthy 注册不健康状态
func (r *HealthCheckRegistry) RegisterUnhealthy(component string, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checker.RecordUnhealthy(component, message)
}

// RegisterDegraded 注册降级状态
func (r *HealthCheckRegistry) RegisterDegraded(component string, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checker.RecordDegraded(component, message)
}

// GetChecker 获取健康检查器
func (r *HealthCheckRegistry) GetChecker() HealthChecker {
	return r.checker
}

// 全局健康检查器
var (
	globalHealthChecker     HealthChecker
	globalHealthCheckerOnce sync.Once
)

// GetGlobalHealthChecker 获取全局健康检查器
func GetGlobalHealthChecker() HealthChecker {
	globalHealthCheckerOnce.Do(func() {
		config := HealthCheckConfig{
			Enabled:  true,
			Interval: 30 * time.Second,
			Timeout:  5 * time.Second,
		}
		globalHealthChecker = NewHealthChecker(config)
	})
	return globalHealthChecker
}

// SetGlobalHealthChecker 设置全局健康检查器
func SetGlobalHealthChecker(checker HealthChecker) {
	globalHealthChecker = checker
}
