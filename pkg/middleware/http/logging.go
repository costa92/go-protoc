package http

import (
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/costa92/go-protoc/pkg/metrics"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
)

const UnknownTraceID = "unknown"

// LoggingMiddleware 创建一个 HTTP 日志中间件
func LoggingMiddleware(skipPaths []string) mux.MiddlewareFunc {
	log.Infow("LoggingMiddleware", "skipPaths", skipPaths)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{w, http.StatusOK}

			log.Infow("LoggingMiddleware", "r.URL.Path", r.URL.Path)
			traceID := UnknownTraceID
			if !slices.Contains(skipPaths, r.URL.Path) {
				spanCtx := trace.SpanContextFromContext(r.Context())
				if spanCtx.IsValid() {
					traceID = spanCtx.TraceID().String()
				}
			}

			next.ServeHTTP(rw, r)

			duration := time.Since(start)

			// 记录Prometheus指标
			metrics.HTTPRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(rw.status),
			).Inc()
			metrics.HTTPRequestDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration.Seconds())

			// 记录请求信息
			log.L().WithValues(
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
				"duration", duration,
				"trace_id", traceID,
			).Infof("http request")
		})
	}
}

// responseWriter 是一个自定义的 ResponseWriter，用于捕获状态码
type responseWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader 实现 http.ResponseWriter 接口
func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// RecoveryMiddleware 创建一个 HTTP 恢复中间件
func RecoveryMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					spanCtx := trace.SpanContextFromContext(r.Context())
					traceID := UnknownTraceID
					if spanCtx.IsValid() {
						traceID = spanCtx.TraceID().String()
					}

					log.L().WithValues(
						"error", err,
						"path", r.URL.Path,
						"trace_id", traceID,
					).Errorf("http panic recovered")

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
