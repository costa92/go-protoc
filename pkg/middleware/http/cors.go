package http

import (
	"net/http"

	"github.com/costa92/go-protoc/pkg/config"
	"github.com/gorilla/mux"
)

// CORSMiddleware 创建一个 CORS 中间件
func CORSMiddleware(cfg *config.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置 CORS 头
			if len(cfg.Middleware.CORS.AllowOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", cfg.Middleware.CORS.AllowOrigins[0])
			}
			if len(cfg.Middleware.CORS.AllowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", joinStrings(cfg.Middleware.CORS.AllowMethods))
			}
			if len(cfg.Middleware.CORS.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(cfg.Middleware.CORS.AllowHeaders))
			}
			if len(cfg.Middleware.CORS.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinStrings(cfg.Middleware.CORS.ExposeHeaders))
			}
			if cfg.Middleware.CORS.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if cfg.Middleware.CORS.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", cfg.Middleware.CORS.MaxAge.String())
			}

			// 处理预检请求
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
