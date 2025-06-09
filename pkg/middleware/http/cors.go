package http

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// CORSMiddleware 创建一个 CORS 中间件，接受明确的配置参数而不是依赖全局配置
func CORSMiddleware(
	allowOrigins []string,
	allowMethods []string,
	allowHeaders []string,
	exposeHeaders []string,
	allowCredentials bool,
	maxAge time.Duration,
) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 设置 CORS 头
			if len(allowOrigins) > 0 {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigins[0])
			}
			if len(allowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", joinStrings(allowMethods))
			}
			if len(allowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", joinStrings(allowHeaders))
			}
			if len(exposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", joinStrings(exposeHeaders))
			}
			if allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if maxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", maxAge.String())
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
