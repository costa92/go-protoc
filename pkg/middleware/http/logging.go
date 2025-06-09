package http

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/costa92/go-protoc/pkg/config"
	"github.com/costa92/go-protoc/pkg/log"
	"github.com/costa92/go-protoc/pkg/metrics"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
)

// LoggingMiddleware 创建一个 HTTP 日志中间件
func LoggingMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{w, http.StatusOK}

			spanCtx := trace.SpanContextFromContext(r.Context())
			traceID := "unknown"
			if spanCtx.IsValid() {
				traceID = spanCtx.TraceID().String()
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
					traceID := "unknown"
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

// TimeoutMiddleware 创建一个超时中间件
func TimeoutMiddleware(cfg *config.Config) mux.MiddlewareFunc {
	return TimeoutMiddlewareWithConfig(cfg.Middleware.Timeout, &cfg.Observability)
}

// TimeoutMiddlewareWithConfig 创建一个带自定义配置的超时中间件
func TimeoutMiddlewareWithConfig(timeout time.Duration, cfg *config.ObservabilityConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否在白名单中
			if shouldSkip(cfg.SkipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			r = r.WithContext(ctx)
			done := make(chan struct{})

			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				w.WriteHeader(http.StatusGatewayTimeout)
				return
			}
		})
	}
}

// shouldSkip 检查是否应该跳过该路径
func shouldSkip(skipPaths []string, path string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}
