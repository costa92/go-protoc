package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/time/rate"
)

// rateLimiter 限流器
type rateLimiter struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// rateLimiters 存储所有限流器
var (
	rateLimiters = make(map[string]*rateLimiter)
	mu           sync.Mutex
)

// cleanup 清理过期的限流器
func cleanup(cleanupInterval time.Duration) {
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for range ticker.C {
			mu.Lock()
			for ip, rl := range rateLimiters {
				if time.Since(rl.lastSeen) > cleanupInterval {
					delete(rateLimiters, ip)
				}
			}
			mu.Unlock()
		}
	}()
}

// RateLimitMiddleware 创建一个限流中间件，接受明确的配置参数而不是依赖全局配置
func RateLimitMiddleware(
	enable bool,
	limit float64,
	burst int,
	skipPaths []string, // 可选的跳过路径列表
) mux.MiddlewareFunc {
	if !enable {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// 启动清理过期限流器的协程
	cleanup(time.Hour)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否在白名单中
			for _, path := range skipPaths {
				if r.URL.Path == path {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 获取客户端 IP
			ip := r.RemoteAddr
			if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
				ip = forwardedFor
			}

			// 获取或创建限流器
			mu.Lock()
			limiter, exists := rateLimiters[ip]
			if !exists {
				limiter = &rateLimiter{
					limiter: rate.NewLimiter(rate.Limit(limit), burst),
				}
				rateLimiters[ip] = limiter
			}
			limiter.lastSeen = time.Now()
			mu.Unlock()

			// 检查是否允许请求
			if !limiter.limiter.Allow() {
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
