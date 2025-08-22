package log

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
)

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// VictoriaLogsWriter 是一个直接写入VictoriaLogs的Writer实现
type VictoriaLogsWriter struct {
	endpoint    string
	client      *http.Client
	buffer      chan LogEntry
	bufferSize  int
	batchSize   int
	flushTicker *time.Ticker
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	closed      bool
}

// LogEntry 表示发送到VictoriaLogs的日志条目
// VictoriaLogs要求使用特定的字段名：_msg, _time, _stream
type LogEntry struct {
	Timestamp   string                 `json:"_time"` // VictoriaLogs标准时间字段
	Message     string                 `json:"_msg"`  // VictoriaLogs标准消息字段
	Level       string                 `json:"level"`
	Logger      string                 `json:"logger,omitempty"`
	Caller      string                 `json:"caller,omitempty"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version,omitempty"`
	Environment string                 `json:"environment,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"` // 分布式追踪ID
	SpanID      string                 `json:"span_id,omitempty"`  // Span ID
	Fields      map[string]interface{} `json:"fields,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Stack       string                 `json:"stack,omitempty"`
}

// VictoriaLogsOptions 配置选项
type VictoriaLogsOptions struct {
	Endpoint      string        // VictoriaLogs API端点
	Service       string        // 服务名称
	Version       string        // 服务版本
	Environment   string        // 环境标识
	BufferSize    int           // 缓冲区大小
	BatchSize     int           // 批量发送大小
	FlushInterval time.Duration // 强制刷新间隔
	Timeout       time.Duration // HTTP请求超时
}

// NewVictoriaLogsWriter 创建新的VictoriaLogs写入器
func NewVictoriaLogsWriter(opts VictoriaLogsOptions) *VictoriaLogsWriter {
	if opts.BufferSize == 0 {
		opts.BufferSize = 1000
	}
	if opts.BatchSize == 0 {
		opts.BatchSize = 100
	}
	if opts.FlushInterval == 0 {
		opts.FlushInterval = 5 * time.Second
	}
	if opts.Timeout == 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.Endpoint == "" {
		opts.Endpoint = "http://127.0.0.1:9428"
	}

	ctx, cancel := context.WithCancel(context.Background())

	writer := &VictoriaLogsWriter{
		endpoint: opts.Endpoint + "/insert/jsonline",
		client: &http.Client{
			Timeout: opts.Timeout,
		},
		buffer:      make(chan LogEntry, opts.BufferSize),
		bufferSize:  opts.BufferSize,
		batchSize:   opts.BatchSize,
		flushTicker: time.NewTicker(opts.FlushInterval),
		ctx:         ctx,
		cancel:      cancel,
	}

	// 启动后台goroutine处理日志发送
	writer.wg.Add(1)
	go writer.worker(opts)

	return writer
}

// Write 实现io.Writer接口
func (w *VictoriaLogsWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return 0, fmt.Errorf("writer is closed")
	}
	w.mu.Unlock()

	// 解析JSON格式的日志
	var logData map[string]interface{}
	if err := json.Unmarshal(p, &logData); err != nil {
		// 如果不是JSON格式，创建简单的日志条目
		entry := LogEntry{
			Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
			Level:     "INFO",
			Message:   string(bytes.TrimSpace(p)),
		}
		select {
		case w.buffer <- entry:
		default:
			// 缓冲区满，丢弃日志（或可以选择阻塞）
		}
		return len(p), nil
	}

	// 构建标准化的日志条目
	// 按优先级提取消息字段：message > msg > operation > caller > 其他有意义字段
	message := w.extractString(logData, "message", "")
	if message == "" {
		message = w.extractString(logData, "msg", "")
	}
	if message == "" {
		// 尝试从operation构建有意义的消息
		if operation := w.extractString(logData, "operation", ""); operation != "" {
			component := w.extractString(logData, "component", "")
			if component != "" {
				message = fmt.Sprintf("[%s] %s", component, operation)
			} else {
				message = operation
			}
		}
	}
	if message == "" {
		if caller := w.extractString(logData, "caller", ""); caller != "" {
			message = "Log from " + caller
		} else {
			message = "Unknown log entry"
		}
	}

	entry := LogEntry{
		Timestamp: w.convertToRFC3339(logData),
		Message:   message,
		Level:     w.extractString(logData, "level", "INFO"),
		Logger:    w.extractString(logData, "logger", ""),
		Caller:    w.extractString(logData, "caller", ""),
		Error:     w.extractString(logData, "error", ""),
		Stack:     w.extractString(logData, "stack", ""),
		TraceID:   w.extractTraceID(logData),
		SpanID:    w.extractSpanID(logData),
		Fields:    make(map[string]interface{}),
	}

	// 提取其他字段，保留更多有用信息
	for k, v := range logData {
		switch k {
		case "timestamp", "level", "logger", "message", "msg", "caller", "error", "stack":
			// 已处理的标准字段，跳过
		case "_time", "_msg", "_stream", "_stream_id":
			// VictoriaLogs内部字段，跳过
		case "x-trace-id", "trace_id", "traceId", "trace-id", "TraceId", "TRACE_ID", "traceid":
			// trace_id 相关字段已单独处理，跳过
		case "x-span-id", "span_id", "spanId", "span-id", "SpanId", "SPAN_ID", "spanid":
			// span_id 相关字段已单独处理，跳过
		default:
			// 保留所有其他字段，包括service.id, service.name等
			entry.Fields[k] = v
		}
	}

	// 发送到缓冲区
	select {
	case w.buffer <- entry:
	default:
		// 缓冲区满，可以选择丢弃或阻塞
	}

	return len(p), nil
}

// extractString 从map中提取字符串值
func (w *VictoriaLogsWriter) extractString(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return defaultValue
}

// extractTraceID 提取 trace_id，支持多种常见字段名
func (w *VictoriaLogsWriter) extractTraceID(data map[string]interface{}) string {
	// 常见的 trace_id 字段名，优先使用项目标准字段名
	traceFields := []string{
		"x-trace-id", // 项目标准字段名（最高优先级）
		"trace_id", "traceId", "trace-id", "TraceId", "TRACE_ID",
		"traceid", "span.trace_id", "dd.trace_id",
	}

	for _, field := range traceFields {
		if val := w.extractString(data, field, ""); val != "" {
			return val
		}
	}

	// 检查嵌套字段（如 span.trace_id）
	if span, ok := data["span"].(map[string]interface{}); ok {
		if traceID := w.extractString(span, "trace_id", ""); traceID != "" {
			return traceID
		}
	}

	return ""
}

// extractSpanID 提取 span_id，支持多种常见字段名
func (w *VictoriaLogsWriter) extractSpanID(data map[string]interface{}) string {
	// 常见的 span_id 字段名，优先使用项目标准字段名
	spanFields := []string{
		"x-span-id", // 项目标准字段名（最高优先级）
		"span_id", "spanId", "span-id", "SpanId", "SPAN_ID",
		"spanid", "span.span_id", "dd.span_id",
	}

	for _, field := range spanFields {
		if val := w.extractString(data, field, ""); val != "" {
			return val
		}
	}

	// 检查嵌套字段
	if span, ok := data["span"].(map[string]interface{}); ok {
		if spanID := w.extractString(span, "span_id", ""); spanID != "" {
			return spanID
		}
	}

	return ""
}

// convertToRFC3339 将时间戳转换为VictoriaLogs兼容的RFC3339格式（UTC）
func (w *VictoriaLogsWriter) convertToRFC3339(data map[string]interface{}) string {
	timestampStr := w.extractString(data, "timestamp", "")

	if timestampStr == "" {
		// 发送UTC时间，让VictoriaLogs UI处理时区显示
		return time.Now().UTC().Format(time.RFC3339Nano)
	}

	// 尝试解析不同的时间格式
	timeFormats := []string{
		"2006-01-02 15:04:05.000",          // 项目自定义格式
		"2006-01-02T15:04:05.000Z07:00",    // RFC3339Nano
		"2006-01-02T15:04:05Z07:00",        // RFC3339
		"2006-01-02 15:04:05",              // Simple format
		time.RFC3339Nano,                   // Standard RFC3339Nano
		time.RFC3339,                       // Standard RFC3339
		"2006-01-02T15:04:05.000000Z07:00", // RFC3339 with microseconds
		"2006-01-02T15:04:05.000000Z",      // UTC with microseconds
	}

	for _, format := range timeFormats {
		if t, err := time.Parse(format, timestampStr); err == nil {
			// 如果解析的时间没有时区信息，假设它是本地时间
			if t.Location() == time.UTC && format == "2006-01-02 15:04:05.000" {
				// 项目自定义格式假设为本地时间，需要获取系统时区
				localTime := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
				return localTime.UTC().Format(time.RFC3339Nano)
			}
			// 转换为UTC发送
			return t.UTC().Format(time.RFC3339Nano)
		}
	}

	// 如果所有格式都解析失败，返回当前UTC时间
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// worker 后台工作协程，批量发送日志到VictoriaLogs
func (w *VictoriaLogsWriter) worker(opts VictoriaLogsOptions) {
	defer w.wg.Done()

	batch := make([]LogEntry, 0, w.batchSize)

	for {
		select {
		case <-w.ctx.Done():
			// 处理剩余的日志
			if len(batch) > 0 {
				w.sendBatch(batch, opts)
			}
			return

		case entry := <-w.buffer:
			// 添加服务信息
			entry.Service = opts.Service
			entry.Version = opts.Version
			entry.Environment = opts.Environment

			batch = append(batch, entry)

			// 检查是否需要发送批次
			if len(batch) >= w.batchSize {
				w.sendBatch(batch, opts)
				batch = batch[:0] // 重置切片
			}

		case <-w.flushTicker.C:
			// 定时刷新
			if len(batch) > 0 {
				w.sendBatch(batch, opts)
				batch = batch[:0] // 重置切片
			}
		}
	}
}

// sendBatch 批量发送日志到VictoriaLogs
func (w *VictoriaLogsWriter) sendBatch(batch []LogEntry, opts VictoriaLogsOptions) {
	if len(batch) == 0 {
		return
	}

	// 构建JSON Lines格式的数据
	var buf bytes.Buffer
	for _, entry := range batch {
		data, err := json.Marshal(entry)
		if err != nil {
			fmt.Printf("VictoriaLogs JSON marshal error: %v\n", err)
			continue // 跳过无法序列化的条目
		}
		buf.Write(data)
		buf.WriteByte('\n')
	}

	// 发送HTTP请求
	req, err := http.NewRequestWithContext(w.ctx, "POST", w.endpoint, &buf)
	if err != nil {
		fmt.Printf("VictoriaLogs create request failed: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/x-ndjson")
	req.Header.Set("User-Agent", "apiserver-victoria-writer/v2.0.0")

	resp, err := w.client.Do(req)
	if err != nil {
		// 发送失败，记录到标准错误输出
		fmt.Printf("VictoriaLogs HTTP request failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode >= 400 {
		// 读取错误响应
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("VictoriaLogs insert failed: %d %s\n", resp.StatusCode, string(body))
		fmt.Printf("Failed payload (first entry): %s\n", buf.String()[:min(500, buf.Len())])
	}
	// 成功时静默运行
}

// Sync 同步刷新所有待发送的日志
func (w *VictoriaLogsWriter) Sync() error {
	// 等待缓冲区清空或超时
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("sync timeout")
		case <-ticker.C:
			if len(w.buffer) == 0 {
				return nil
			}
		}
	}
}

// Close 关闭写入器
func (w *VictoriaLogsWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()

	// 停止ticker
	w.flushTicker.Stop()

	// 取消context
	w.cancel()

	// 等待worker完成
	w.wg.Wait()

	// 关闭缓冲区
	close(w.buffer)

	return nil
}

// WriteSyncer 返回zapcore.WriteSyncer接口
func (w *VictoriaLogsWriter) WriteSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(w)
}
