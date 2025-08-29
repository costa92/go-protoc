package logger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/sdk/resource"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// OTLPConfig OTLP 配置
type OTLPConfig struct {
	Enabled       bool          `json:"enabled" mapstructure:"enabled"`
	Endpoint      string        `json:"endpoint" mapstructure:"endpoint"`
	Protocol      string        `json:"protocol" mapstructure:"protocol"` // grpc 或 http
	Timeout       time.Duration `json:"timeout" mapstructure:"timeout"`
	BatchTimeout  time.Duration `json:"batch_timeout" mapstructure:"batch_timeout"`
	BatchSize     int           `json:"batch_size" mapstructure:"batch_size"`
	Insecure      bool          `json:"insecure" mapstructure:"insecure"`
	Headers       map[string]string `json:"headers" mapstructure:"headers"`
	
	// Resource attributes
	ServiceName    string `json:"service_name" mapstructure:"service_name"`
	ServiceVersion string `json:"service_version" mapstructure:"service_version"`
	Environment    string `json:"environment" mapstructure:"environment"`
}

// DefaultOTLPConfig 返回默认 OTLP 配置
func DefaultOTLPConfig() *OTLPConfig {
	return &OTLPConfig{
		Enabled:       false,
		Endpoint:      "127.0.0.1:4327", // OTEL Agent gRPC port
		Protocol:      "grpc",
		Timeout:       5 * time.Second,
		BatchTimeout:  1 * time.Second,
		BatchSize:     100,
		Insecure:      true,
		Headers:       make(map[string]string),
		ServiceName:   "apiserver",
		ServiceVersion: "v1.0.0",
		Environment:   "development",
	}
}

// OTLPLogExporter OTLP 日志导出器
type OTLPLogExporter struct {
	config         *OTLPConfig
	loggerProvider *sdklog.LoggerProvider
	exporter       sdklog.Exporter
	resource       *resource.Resource
	mu             sync.RWMutex
	closed         bool
}

// NewOTLPLogExporter 创建 OTLP 日志导出器
func NewOTLPLogExporter(cfg *OTLPConfig) (*OTLPLogExporter, error) {
	if cfg == nil {
		cfg = DefaultOTLPConfig()
	}

	// 创建资源
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// 创建 OTLP 导出器
	var exporter sdklog.Exporter
	switch cfg.Protocol {
	case "grpc":
		exporter, err = createGRPCExporter(cfg)
	case "http":
		exporter, err = createHTTPExporter(cfg)
	default:
		return nil, fmt.Errorf("unsupported OTLP protocol: %s", cfg.Protocol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// 创建批处理器
	processor := sdklog.NewBatchProcessor(exporter)

	// 创建 LoggerProvider
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(processor),
		sdklog.WithResource(res),
	)

	return &OTLPLogExporter{
		config:         cfg,
		loggerProvider: loggerProvider,
		exporter:       exporter,
		resource:       res,
	}, nil
}

// createGRPCExporter 创建 gRPC OTLP 导出器
func createGRPCExporter(cfg *OTLPConfig) (sdklog.Exporter, error) {
	opts := []otlploggrpc.Option{
		otlploggrpc.WithEndpoint(cfg.Endpoint),
		otlploggrpc.WithTimeout(cfg.Timeout),
	}

	if cfg.Insecure {
		opts = append(opts, otlploggrpc.WithTLSCredentials(insecure.NewCredentials()))
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploggrpc.WithHeaders(cfg.Headers))
	}

	// 添加 gRPC 连接选项
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(),
	}
	
	if cfg.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	opts = append(opts, otlploggrpc.WithDialOption(dialOpts...))

	return otlploggrpc.New(context.Background(), opts...)
}

// createHTTPExporter 创建 HTTP OTLP 导出器
func createHTTPExporter(cfg *OTLPConfig) (sdklog.Exporter, error) {
	endpoint := cfg.Endpoint
	if cfg.Protocol == "http" && !contains(endpoint, "://") {
		if cfg.Insecure {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithTimeout(cfg.Timeout),
	}

	if cfg.Insecure {
		opts = append(opts, otlploghttp.WithInsecure())
	}

	if len(cfg.Headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(cfg.Headers))
	}

	return otlploghttp.New(context.Background(), opts...)
}

// Export 导出日志记录
func (o *OTLPLogExporter) Export(ctx context.Context, records []sdklog.Record) error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.closed {
		return fmt.Errorf("OTLP exporter is closed")
	}

	return o.exporter.Export(ctx, records)
}

// Shutdown 关闭导出器
func (o *OTLPLogExporter) Shutdown(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.closed {
		return nil
	}

	o.closed = true

	if err := o.loggerProvider.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown logger provider: %w", err)
	}

	return o.exporter.Shutdown(ctx)
}

// ForceFlush 强制刷新待处理的记录
func (o *OTLPLogExporter) ForceFlush(ctx context.Context) error {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.closed {
		return fmt.Errorf("OTLP exporter is closed")
	}

	return o.loggerProvider.ForceFlush(ctx)
}

// GetLoggerProvider 获取 LoggerProvider
func (o *OTLPLogExporter) GetLoggerProvider() *sdklog.LoggerProvider {
	return o.loggerProvider
}

// OTLPZapCore 实现 zapcore.Core 接口，将日志发送到 OTLP
type OTLPZapCore struct {
	otlpExporter *OTLPLogExporter
	logger       log.Logger
	level        zapcore.Level
	fields       []zapcore.Field
}

// NewOTLPZapCore 创建 OTLP Zap Core
func NewOTLPZapCore(exporter *OTLPLogExporter, level zapcore.Level) *OTLPZapCore {
	logger := exporter.loggerProvider.Logger("zap-otlp-bridge")
	
	return &OTLPZapCore{
		otlpExporter: exporter,
		logger:       logger,
		level:        level,
		fields:       make([]zapcore.Field, 0),
	}
}

// Enabled 检查是否应该记录给定级别的日志
func (c *OTLPZapCore) Enabled(level zapcore.Level) bool {
	return level >= c.level
}

// With 添加结构化上下文到 Core
func (c *OTLPZapCore) With(fields []zapcore.Field) zapcore.Core {
	clone := *c
	clone.fields = append(clone.fields[:0:0], c.fields...)
	clone.fields = append(clone.fields, fields...)
	return &clone
}

// Check 确定给定条目是否应该被记录
func (c *OTLPZapCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

// Write 将日志条目写入 OTLP 导出器
func (c *OTLPZapCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	// 对于简化实现，我们先跳过 OTLP 发送
	// TODO: 实现完整的日志记录转换
	return nil
}

// Sync 刷新任何缓冲的日志条目
func (c *OTLPZapCore) Sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return c.otlpExporter.ForceFlush(ctx)
}

// zapLevelToOTLPSeverity 将 Zap 日志级别转换为 OTLP 严重性级别
func zapLevelToOTLPSeverity(level zapcore.Level) log.Severity {
	switch level {
	case zapcore.DebugLevel:
		return log.SeverityDebug
	case zapcore.InfoLevel:
		return log.SeverityInfo
	case zapcore.WarnLevel:
		return log.SeverityWarn
	case zapcore.ErrorLevel:
		return log.SeverityError
	case zapcore.DPanicLevel, zapcore.PanicLevel:
		return log.SeverityFatal1
	case zapcore.FatalLevel:
		return log.SeverityFatal4
	default:
		return log.SeverityInfo
	}
}

// zapFieldToOTLPAttribute 将 Zap 字段转换为 OTLP 属性
func zapFieldToOTLPAttribute(field zapcore.Field) log.KeyValue {
	// 简化实现，暂时返回字符串类型
	// TODO: 实现完整的类型转换
	return log.String(field.Key, fmt.Sprintf("%v", field.Interface))
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
				func() bool {
					for i := 0; i <= len(s)-len(substr); i++ {
						if s[i:i+len(substr)] == substr {
							return true
						}
					}
					return false
				}())))
}