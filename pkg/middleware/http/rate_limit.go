package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/costa92/go-protoc/pkg/config"
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

// RateLimitMiddleware 创建一个限流中间件
func RateLimitMiddleware(cfg *config.Config) mux.MiddlewareFunc {
	if !cfg.Middleware.RateLimit.Enable {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	// 启动清理过期限流器的协程
	cleanup(time.Hour)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 检查是否在白名单中
			if shouldSkip(cfg.Observability.SkipPaths, r.URL.Path) {
				next.ServeHTTP(w, r)
				return
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
					limiter: rate.NewLimiter(rate.Limit(cfg.Middleware.RateLimit.Limit), cfg.Middleware.RateLimit.Burst),
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
