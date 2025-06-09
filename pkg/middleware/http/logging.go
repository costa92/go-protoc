package http

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/costa92/go-protoc/pkg/metrics"
	"github.com/costa92/go-protoc/pkg/middleware/config"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// LoggingMiddleware 创建一个 HTTP 日志中间件
func LoggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	// 使用默认配置
	cfg := config.DefaultObservabilityConfig()
	return LoggingMiddlewareWithConfig(logger, cfg)
}

// LoggingMiddlewareWithConfig 创建一个带自定义配置的 HTTP 日志中间件
func LoggingMiddlewareWithConfig(logger *zap.Logger, cfg *config.ObservabilityConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否在白名单中
			if cfg.ShouldSkip(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// 创建一个自定义的 ResponseWriter 来捕获状态码
			rw := &responseWriter{w, http.StatusOK}

			// 从context中提取TraceID
			spanCtx := trace.SpanContextFromContext(r.Context())
			traceID := "unknown"
			if spanCtx.IsValid() {
				traceID = spanCtx.TraceID().String()
			}

			// 处理请求
			next.ServeHTTP(rw, r)

			// 计算请求耗时
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
			logger.Info("http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
				zap.Duration("duration", duration),
				zap.String("trace_id", traceID),
			)
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
func RecoveryMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	// 使用默认配置
	cfg := config.DefaultObservabilityConfig()
	return RecoveryMiddlewareWithConfig(logger, cfg)
}

// RecoveryMiddlewareWithConfig 创建一个带自定义配置的 HTTP 恢复中间件
func RecoveryMiddlewareWithConfig(logger *zap.Logger, cfg *config.ObservabilityConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否在白名单中
			if cfg.ShouldSkip(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			defer func() {
				if err := recover(); err != nil {
					// 从context中提取TraceID
					spanCtx := trace.SpanContextFromContext(r.Context())
					traceID := "unknown"
					if spanCtx.IsValid() {
						traceID = spanCtx.TraceID().String()
					}

					logger.Error("http panic recovered",
						zap.Any("error", err),
						zap.String("path", r.URL.Path),
						zap.String("trace_id", traceID),
					)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// TimeoutMiddleware 创建一个超时中间件
func TimeoutMiddleware(timeout time.Duration) mux.MiddlewareFunc {
	// 使用默认配置
	cfg := config.DefaultObservabilityConfig()
	return TimeoutMiddlewareWithConfig(timeout, cfg)
}

// TimeoutMiddlewareWithConfig 创建一个带自定义配置的超时中间件
func TimeoutMiddlewareWithConfig(timeout time.Duration, cfg *config.ObservabilityConfig) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否在白名单中
			if cfg.ShouldSkip(r.URL.Path) {
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
