package metrics

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// CustomCollector 自定义指标收集器
type CustomCollector struct {
	config   *Config
	registry prometheus.Registerer
	enabled  atomic.Bool
	running  atomic.Bool

	// 动态指标存储
	counters   sync.Map // map[string]*prometheus.CounterVec
	gauges     sync.Map // map[string]*prometheus.GaugeVec
	histograms sync.Map // map[string]*prometheus.HistogramVec
	summaries  sync.Map // map[string]*prometheus.SummaryVec

	// 内部状态
	mu sync.RWMutex
}

// NewCustomCollector 创建自定义指标收集器
func NewCustomCollector(config *Config) *CustomCollector {
	if config == nil {
		config = NewConfig()
	}

	return &CustomCollector{
		config:   config,
		registry: config.Registry,
	}
}

// Start 启动收集器
func (c *CustomCollector) Start() error {
	if !c.running.CompareAndSwap(false, true) {
		return nil // 已经运行
	}

	if !c.config.Custom.Enabled {
		c.running.Store(false)
		return nil
	}

	c.enabled.Store(true)
	return nil
}

// Stop 停止收集器
func (c *CustomCollector) Stop() error {
	if !c.running.CompareAndSwap(true, false) {
		return nil // 未运行
	}

	c.enabled.Store(false)
	return nil
}

// IsRunning 检查是否运行
func (c *CustomCollector) IsRunning() bool {
	return c.running.Load()
}

// GetType 获取收集器类型
func (c *CustomCollector) GetType() MetricType {
	return MetricTypeCustom
}

// GetStatus 获取状态
func (c *CustomCollector) GetStatus() map[string]interface{} {
	var counterCount, gaugeCount, histogramCount, summaryCount int

	c.counters.Range(func(_, _ interface{}) bool {
		counterCount++
		return true
	})

	c.gauges.Range(func(_, _ interface{}) bool {
		gaugeCount++
		return true
	})

	c.histograms.Range(func(_, _ interface{}) bool {
		histogramCount++
		return true
	})

	c.summaries.Range(func(_, _ interface{}) bool {
		summaryCount++
		return true
	})

	return map[string]interface{}{
		"type":            string(MetricTypeCustom),
		"enabled":         c.enabled.Load(),
		"running":         c.running.Load(),
		"counter_count":   counterCount,
		"gauge_count":     gaugeCount,
		"histogram_count": histogramCount,
		"summary_count":   summaryCount,
		"config":          c.config.Custom,
	}
}

// RegisterTarget 注册监控目标
func (c *CustomCollector) RegisterTarget(name string, target interface{}) error {
	// 自定义收集器不需要预先注册目标
	return nil
}

// UnregisterTarget 取消注册监控目标
func (c *CustomCollector) UnregisterTarget(name string) error {
	// 自定义收集器不需要预先注册目标
	return nil
}

// CreateCounter 创建计数器
func (c *CustomCollector) CreateCounter(name, help string, labels []string) (*prometheus.CounterVec, error) {
	if !c.enabled.Load() {
		return nil, ErrCollectorDisabled
	}

	if existing, ok := c.counters.Load(name); ok {
		return existing.(*prometheus.CounterVec), nil
	}

	counter := promauto.With(c.registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: c.config.Namespace,
			Subsystem: "custom",
			Name:      name,
			Help:      help,
		},
		labels,
	)

	c.counters.Store(name, counter)
	return counter, nil
}

// CreateGauge 创建仪表盘
func (c *CustomCollector) CreateGauge(name, help string, labels []string) (*prometheus.GaugeVec, error) {
	if !c.enabled.Load() {
		return nil, ErrCollectorDisabled
	}

	if existing, ok := c.gauges.Load(name); ok {
		return existing.(*prometheus.GaugeVec), nil
	}

	gauge := promauto.With(c.registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: c.config.Namespace,
			Subsystem: "custom",
			Name:      name,
			Help:      help,
		},
		labels,
	)

	c.gauges.Store(name, gauge)
	return gauge, nil
}

// CreateHistogram 创建直方图
func (c *CustomCollector) CreateHistogram(name, help string, labels []string, buckets []float64) (*prometheus.HistogramVec, error) {
	if !c.enabled.Load() {
		return nil, ErrCollectorDisabled
	}

	if existing, ok := c.histograms.Load(name); ok {
		return existing.(*prometheus.HistogramVec), nil
	}

	if buckets == nil {
		buckets = prometheus.DefBuckets
	}

	histogram := promauto.With(c.registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: c.config.Namespace,
			Subsystem: "custom",
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labels,
	)

	c.histograms.Store(name, histogram)
	return histogram, nil
}

// CreateSummary 创建摘要
func (c *CustomCollector) CreateSummary(name, help string, labels []string, objectives map[float64]float64) (*prometheus.SummaryVec, error) {
	if !c.enabled.Load() {
		return nil, ErrCollectorDisabled
	}

	if existing, ok := c.summaries.Load(name); ok {
		return existing.(*prometheus.SummaryVec), nil
	}

	if objectives == nil {
		objectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001}
	}

	summary := promauto.With(c.registry).NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  c.config.Namespace,
			Subsystem:  "custom",
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labels,
	)

	c.summaries.Store(name, summary)
	return summary, nil
}

// GetCounter 获取计数器
func (c *CustomCollector) GetCounter(name string) (*prometheus.CounterVec, bool) {
	if counter, ok := c.counters.Load(name); ok {
		return counter.(*prometheus.CounterVec), true
	}
	return nil, false
}

// GetGauge 获取仪表盘
func (c *CustomCollector) GetGauge(name string) (*prometheus.GaugeVec, bool) {
	if gauge, ok := c.gauges.Load(name); ok {
		return gauge.(*prometheus.GaugeVec), true
	}
	return nil, false
}

// GetHistogram 获取直方图
func (c *CustomCollector) GetHistogram(name string) (*prometheus.HistogramVec, bool) {
	if histogram, ok := c.histograms.Load(name); ok {
		return histogram.(*prometheus.HistogramVec), true
	}
	return nil, false
}

// GetSummary 获取摘要
func (c *CustomCollector) GetSummary(name string) (*prometheus.SummaryVec, bool) {
	if summary, ok := c.summaries.Load(name); ok {
		return summary.(*prometheus.SummaryVec), true
	}
	return nil, false
}

// IncrementCounter 增加计数器
func (c *CustomCollector) IncrementCounter(name string, labels prometheus.Labels, value float64) error {
	if !c.enabled.Load() {
		return ErrCollectorDisabled
	}

	counter, ok := c.GetCounter(name)
	if !ok {
		return ErrMetricNotFound
	}

	if value <= 0 {
		value = 1
	}

	counter.With(labels).Add(value)
	return nil
}

// SetGauge 设置仪表盘值
func (c *CustomCollector) SetGauge(name string, labels prometheus.Labels, value float64) error {
	if !c.enabled.Load() {
		return ErrCollectorDisabled
	}

	gauge, ok := c.GetGauge(name)
	if !ok {
		return ErrMetricNotFound
	}

	gauge.With(labels).Set(value)
	return nil
}

// ObserveHistogram 观察直方图
func (c *CustomCollector) ObserveHistogram(name string, labels prometheus.Labels, value float64) error {
	if !c.enabled.Load() {
		return ErrCollectorDisabled
	}

	histogram, ok := c.GetHistogram(name)
	if !ok {
		return ErrMetricNotFound
	}

	histogram.With(labels).Observe(value)
	return nil
}

// ObserveSummary 观察摘要
func (c *CustomCollector) ObserveSummary(name string, labels prometheus.Labels, value float64) error {
	if !c.enabled.Load() {
		return ErrCollectorDisabled
	}

	summary, ok := c.GetSummary(name)
	if !ok {
		return ErrMetricNotFound
	}

	summary.With(labels).Observe(value)
	return nil
}

// RecordDuration 记录持续时间
func (c *CustomCollector) RecordDuration(name string, labels prometheus.Labels, duration time.Duration) error {
	return c.ObserveHistogram(name, labels, duration.Seconds())
}

// RecordSize 记录大小
func (c *CustomCollector) RecordSize(name string, labels prometheus.Labels, size int64) error {
	return c.ObserveHistogram(name, labels, float64(size))
}

// DeleteMetric 删除指标
func (c *CustomCollector) DeleteMetric(metricType string, name string) error {
	if !c.enabled.Load() {
		return ErrCollectorDisabled
	}

	switch metricType {
	case "counter":
		c.counters.Delete(name)
	case "gauge":
		c.gauges.Delete(name)
	case "histogram":
		c.histograms.Delete(name)
	case "summary":
		c.summaries.Delete(name)
	default:
		return ErrUnsupportedMetricType
	}

	return nil
}

// ListMetrics 列出所有指标
func (c *CustomCollector) ListMetrics() map[string][]string {
	result := map[string][]string{
		"counters":   {},
		"gauges":     {},
		"histograms": {},
		"summaries":  {},
	}

	c.counters.Range(func(key, _ interface{}) bool {
		result["counters"] = append(result["counters"], key.(string))
		return true
	})

	c.gauges.Range(func(key, _ interface{}) bool {
		result["gauges"] = append(result["gauges"], key.(string))
		return true
	})

	c.histograms.Range(func(key, _ interface{}) bool {
		result["histograms"] = append(result["histograms"], key.(string))
		return true
	})

	c.summaries.Range(func(key, _ interface{}) bool {
		result["summaries"] = append(result["summaries"], key.(string))
		return true
	})

	return result
}

// GetCustomMetrics 获取自定义指标接口
func (c *CustomCollector) GetCustomMetrics() CustomMetrics {
	return c
}

// CustomMetricsBuilder 自定义指标构建器
type CustomMetricsBuilder struct {
	collector *CustomCollector
}

// NewCustomMetricsBuilder 创建自定义指标构建器
func NewCustomMetricsBuilder(collector *CustomCollector) *CustomMetricsBuilder {
	return &CustomMetricsBuilder{
		collector: collector,
	}
}

// Counter 构建计数器
func (b *CustomMetricsBuilder) Counter(name, help string) *CounterBuilder {
	return &CounterBuilder{
		collector: b.collector,
		name:      name,
		help:      help,
		labels:    []string{},
	}
}

// Gauge 构建仪表盘
func (b *CustomMetricsBuilder) Gauge(name, help string) *GaugeBuilder {
	return &GaugeBuilder{
		collector: b.collector,
		name:      name,
		help:      help,
		labels:    []string{},
	}
}

// Histogram 构建直方图
func (b *CustomMetricsBuilder) Histogram(name, help string) *HistogramBuilder {
	return &HistogramBuilder{
		collector: b.collector,
		name:      name,
		help:      help,
		labels:    []string{},
		buckets:   prometheus.DefBuckets,
	}
}

// Summary 构建摘要
func (b *CustomMetricsBuilder) Summary(name, help string) *SummaryBuilder {
	return &SummaryBuilder{
		collector:  b.collector,
		name:       name,
		help:       help,
		labels:     []string{},
		objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}
}

// CounterBuilder 计数器构建器
type CounterBuilder struct {
	collector *CustomCollector
	name      string
	help      string
	labels    []string
}

// WithLabels 添加标签
func (b *CounterBuilder) WithLabels(labels ...string) *CounterBuilder {
	b.labels = append(b.labels, labels...)
	return b
}

// Build 构建计数器
func (b *CounterBuilder) Build() (*prometheus.CounterVec, error) {
	return b.collector.CreateCounter(b.name, b.help, b.labels)
}

// GaugeBuilder 仪表盘构建器
type GaugeBuilder struct {
	collector *CustomCollector
	name      string
	help      string
	labels    []string
}

// WithLabels 添加标签
func (b *GaugeBuilder) WithLabels(labels ...string) *GaugeBuilder {
	b.labels = append(b.labels, labels...)
	return b
}

// Build 构建仪表盘
func (b *GaugeBuilder) Build() (*prometheus.GaugeVec, error) {
	return b.collector.CreateGauge(b.name, b.help, b.labels)
}

// HistogramBuilder 直方图构建器
type HistogramBuilder struct {
	collector *CustomCollector
	name      string
	help      string
	labels    []string
	buckets   []float64
}

// WithLabels 添加标签
func (b *HistogramBuilder) WithLabels(labels ...string) *HistogramBuilder {
	b.labels = append(b.labels, labels...)
	return b
}

// WithBuckets 设置桶
func (b *HistogramBuilder) WithBuckets(buckets []float64) *HistogramBuilder {
	b.buckets = buckets
	return b
}

// Build 构建直方图
func (b *HistogramBuilder) Build() (*prometheus.HistogramVec, error) {
	return b.collector.CreateHistogram(b.name, b.help, b.labels, b.buckets)
}

// SummaryBuilder 摘要构建器
type SummaryBuilder struct {
	collector  *CustomCollector
	name       string
	help       string
	labels     []string
	objectives map[float64]float64
}

// WithLabels 添加标签
func (b *SummaryBuilder) WithLabels(labels ...string) *SummaryBuilder {
	b.labels = append(b.labels, labels...)
	return b
}

// WithObjectives 设置目标
func (b *SummaryBuilder) WithObjectives(objectives map[float64]float64) *SummaryBuilder {
	b.objectives = objectives
	return b
}

// Build 构建摘要
func (b *SummaryBuilder) Build() (*prometheus.SummaryVec, error) {
	return b.collector.CreateSummary(b.name, b.help, b.labels, b.objectives)
}
