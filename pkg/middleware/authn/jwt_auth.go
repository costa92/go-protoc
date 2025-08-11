package authn

import (
	"context"
	"strings"
	"sync"

	"github.com/costa92/go-protoc/v2/internal/apiserver/pkg/locales"
	v1 "github.com/costa92/go-protoc/v2/pkg/api/apiserver/v1" // For JWT business errors
	"github.com/costa92/go-protoc/v2/pkg/api/errno"           // For creating standard Kratos errors
	"github.com/costa92/go-protoc/v2/pkg/authn"               // For AppClaims and context operations
	"github.com/costa92/go-protoc/v2/pkg/i18n"
	"github.com/costa92/go-protoc/v2/pkg/options" // For JWTOptions
	"github.com/go-kratos/kratos/v2/errors"       // Kratos errors
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"      // For HeaderCarrier
	"github.com/go-kratos/kratos/v2/transport/http" // For http.RequestFromServerContext
	"github.com/golang-jwt/jwt/v5"
)

// PathTrie 前缀树用于高效路径匹配
type PathTrie struct {
	isEnd    bool
	children map[byte]*PathTrie
}

// pathMatcher 全局路径匹配器
var (
	pathMatcher *PathTrie
	once        sync.Once
)

// initPathMatcher 初始化路径匹配器（只执行一次）
func initPathMatcher() {
	once.Do(func() {
		pathMatcher = &PathTrie{children: make(map[byte]*PathTrie)}
		
		// 优化: 预定义的公共路径列表
		publicPaths := []string{
			"/login",
			"/healthz",
			"/metrics",
			"/debug/pprof",
			"/openapi/",
		}
		
		for _, path := range publicPaths {
			pathMatcher.Insert(path)
		}
	})
}

// Insert 插入路径到前缀树
func (pt *PathTrie) Insert(path string) {
	current := pt
	for i := 0; i < len(path); i++ {
		char := path[i]
		if current.children[char] == nil {
			current.children[char] = &PathTrie{children: make(map[byte]*PathTrie)}
		}
		current = current.children[char]
	}
	current.isEnd = true
}

// IsPublicPath 检查路径是否为公共路径
func (pt *PathTrie) IsPublicPath(path string) bool {
	current := pt
	for i := 0; i < len(path); i++ {
		char := path[i]
		if current.children[char] == nil {
			return false
		}
		current = current.children[char]
		// 检查是否匹配前缀路径（以/结尾的路径）
		if current.isEnd {
			// 完全匹配或前缀匹配
			if i == len(path)-1 || (i < len(path)-1 && path[i] == '/') {
				return true
			}
		}
	}
	return current.isEnd
}

// ServerJWTAuth is the JWT authentication middleware for Kratos HTTP server.
func ServerJWTAuth(jwtOpts *options.JWTOptions) middleware.Middleware {
	// 优化: 初始化路径匹配器
	initPathMatcher()
	
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Access the HTTP request to check the path
			if httpReq, ok := http.RequestFromServerContext(ctx); ok {
				path := httpReq.URL.Path
				// 优化: 使用前缀树进行高效路径匹配
				if pathMatcher.IsPublicPath(path) {
					return handler(ctx, req) // Skip auth for public paths
				}
			}

			// Try to get a HeaderCarrier from the context
			header, ok := transport.FromServerContext(ctx)
			if !ok {
				// This should not happen in a normal Kratos HTTP flow
				return nil, errno.ErrorUnauthorized("missing_header_carrier: Request header not found")
			}

			authHeader := header.RequestHeader().Get("Authorization")
			if authHeader == "" {
				message := i18n.FromContext(ctx).T(locales.JWTTokenMissing)
				return nil, v1.ErrorJWTTokenMissing("%s", message)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
				message := i18n.FromContext(ctx).T(locales.JWTTokenFormatInvalid)
				return nil, v1.ErrorJWTTokenFormatInvalid("%s", message)
			}
			tokenString := parts[1]

			// Parse and validate JWT token
			token, err := jwt.ParseWithClaims(tokenString, &authn.AppClaims{}, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(jwtOpts.Key), nil
			})

			if err != nil {
				if errors.Is(err, jwt.ErrTokenExpired) {
					message := i18n.FromContext(ctx).T(locales.JWTTokenExpired)
					return nil, v1.ErrorJWTTokenExpired("%s", message)
				} else if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
					message := i18n.FromContext(ctx).T(locales.JWTTokenInvalid)
					return nil, v1.ErrorJWTTokenInvalid("%s", message)
				} else if errors.Is(err, jwt.ErrTokenMalformed) {
					message := i18n.FromContext(ctx).T(locales.JWTTokenMalformed)
					return nil, v1.ErrorJWTTokenMalformed("%s", message)
				} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
					message := i18n.FromContext(ctx).T(locales.JWTTokenNotValidYet)
					return nil, v1.ErrorJWTTokenNotValidYet("%s", message)
				} else {
					message := i18n.FromContext(ctx).T(locales.JWTTokenInvalid)
					return nil, v1.ErrorJWTTokenInvalid("%s", message)
				}
			}

			// Extract claims from token
			claims, ok := token.Claims.(*authn.AppClaims)
			if !token.Valid || !ok || claims == nil {
				message := i18n.FromContext(ctx).T(locales.JWTTokenInvalid)
				return nil, v1.ErrorJWTTokenInvalid("%s", message)
			}

			// Token is valid. Store UserID (and other claims if needed) in context.
			if claims.CustomClaims.UserID == "" {
				// Depending on requirements, an empty UserID in a valid token might be an error.
				message := i18n.FromContext(ctx).T(locales.JWTTokenInvalid)
				return nil, v1.ErrorJWTTokenInvalid("%s", message)
			}
			newCtx := authn.SetUserIDInContext(ctx, claims.CustomClaims.UserID)

			return handler(newCtx, req)
		}
	}
}
