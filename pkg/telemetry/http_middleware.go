package telemetry

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingMiddleware 是一个为 HTTP 请求添加 tracing 的中间件
func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 提取请求上下文中的 tracing 信息
		ctx := r.Context()
		propagator := otel.GetTextMapPropagator()
		ctx = propagator.Extract(ctx, propagation.HeaderCarrier(r.Header))

		// 创建一个新的 span
		spanName := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		ctx, span := Tracer.Start(
			ctx,
			spanName,
			trace.WithAttributes(
				semconv.HTTPMethodKey.String(r.Method),
				semconv.HTTPURLKey.String(r.URL.String()),
				semconv.HTTPTargetKey.String(r.URL.Path),
				semconv.HTTPHostKey.String(r.Host),
				semconv.HTTPSchemeKey.String(r.URL.Scheme),
				semconv.HTTPUserAgentKey.String(r.UserAgent()),
				attribute.String("http.remote_addr", r.RemoteAddr),
			),
		)
		defer span.End()

		// 创建一个包装了 ResponseWriter 的对象，用于捕获响应状态码和大小
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 记录请求开始时间
		startTime := time.Now()

		// 处理请求
		next.ServeHTTP(ww, r.WithContext(ctx))

		// 计算请求处理时间
		duration := time.Since(startTime)

		// 添加响应相关的属性
		span.SetAttributes(
			semconv.HTTPStatusCodeKey.Int(ww.statusCode),
			attribute.Int64("http.response_size", ww.size),
			attribute.Float64("http.duration_ms", float64(duration.Milliseconds())),
		)

		// 如果状态码表示错误，标记 span 为错误
		if ww.statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP status code: %d", ww.statusCode))
		}
	})
}

// responseWriter 是一个 http.ResponseWriter 的包装
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int64
}

// WriteHeader 实现 http.ResponseWriter 接口
func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write 实现 http.ResponseWriter 接口
func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += int64(size)
	return size, err
}
