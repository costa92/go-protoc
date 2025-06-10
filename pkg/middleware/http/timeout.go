package http

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/costa92/go-protoc/pkg/log"
	"github.com/gorilla/mux"
)

// TimeoutMiddleware 创建一个超时中间件，接受明确的配置参数而不是依赖全局配置
func TimeoutMiddleware(timeout time.Duration) mux.MiddlewareFunc {
	return TimeoutMiddlewareWithSkipPaths(timeout, nil)
}

// TimeoutMiddlewareWithSkipPaths 创建一个带跳过路径的超时中间件
func TimeoutMiddlewareWithSkipPaths(timeout time.Duration, skipPaths []string) mux.MiddlewareFunc {
	log.Infow("TimeoutMiddlewareWithSkipPaths", "skipPaths", skipPaths)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否在白名单中
			if shouldSkipPath(skipPaths, r.URL.Path) {
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

// shouldSkipPath 检查是否应该跳过该路径
func shouldSkipPath(skipPaths []string, path string) bool {
	if skipPaths == nil {
		return false
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}
