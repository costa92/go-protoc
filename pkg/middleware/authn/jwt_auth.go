package authn

import (
	"context"
	"strings"

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

// PublicPaths is a list of paths that do not require JWT authentication.
// This should ideally be configurable.
var PublicPaths = map[string]bool{
	"/login":        true, // Example: Login path
	"/healthz":      true, // Example: Health check
	"/metrics":      true, // Prometheus metrics
	"/debug/pprof":  true, // Base pprof path
	"/debug/pprof/": true, // Also handle trailing slash for prefixes
	"/openapi/":     true, // OpenAPI docs
	// Add other public paths or path prefixes here
	// Note: For prefixes, ensure the check handles them correctly (e.g., strings.HasPrefix)
}

// ServerJWTAuth is the JWT authentication middleware for Kratos HTTP server.
func ServerJWTAuth(jwtOpts *options.JWTOptions) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// Access the HTTP request to check the path
			if httpReq, ok := http.RequestFromServerContext(ctx); ok {
				path := httpReq.URL.Path
				// Check if the path (or its prefix) is public
				if PublicPaths[path] {
					return handler(ctx, req) // Skip auth for public paths
				}
				// Handle prefix paths
				for prefix := range PublicPaths {
					if strings.HasPrefix(path, prefix) && strings.HasSuffix(prefix, "/") {
						if PublicPaths[prefix] {
							return handler(ctx, req)
						}
					}
				}
			} else {
				// If we can't get HTTP request, proceed with caution or deny.
				// For now, let's assume if no HTTP request, it might be a non-HTTP context
				// or an issue, so we proceed to auth check by default.
				// Alternatively, could return an error here.
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
