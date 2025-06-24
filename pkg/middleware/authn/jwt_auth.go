package authn

import (
	"context"
	"strings"

	"github.com/costa92/go-protoc/v2/pkg/api/errno" // For creating standard Kratos errors
	"github.com/costa92/go-protoc/v2/pkg/authn"    // For AppClaims and context operations
	"github.com/costa92/go-protoc/v2/pkg/options" // For JWTOptions
	"github.com/go-kratos/kratos/v2/errors"       // Kratos errors
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport" // For HeaderCarrier
	"github.com/golang-jwt/jwt/v5"
	"github.com/go-kratos/kratos/v2/transport/http" // For http.RequestFromServerContext
)

// PublicPaths is a list of paths that do not require JWT authentication.
// This should ideally be configurable.
var PublicPaths = map[string]bool{
	"/login":          true, // Example: Login path
	"/healthz":        true, // Example: Health check
	"/metrics":        true, // Prometheus metrics
	"/debug/pprof":    true, // Base pprof path
	"/debug/pprof/":   true, // Also handle trailing slash for prefixes
	"/openapi/":       true, // OpenAPI docs
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
				return nil, errno.ErrorUnauthorized("missing_token: Authorization header is missing")
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
				return nil, errno.ErrorUnauthorized("invalid_token_format: Authorization header format must be Bearer {token}")
			}
			tokenString := parts[1]

			// Parse and validate the token
			claims := &authn.AppClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Validate the alg is what we expect:
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					// This check is for HMAC. If using RSA/ECDSA, check for *jwt.SigningMethodRSA/*jwt.SigningMethodECDSA
					// And ensure jwtOpts.SigningMethod matches token.Header["alg"]
					return nil, errors.BadRequest("JWT_VALIDATION", "Unexpected signing method: %v", token.Header["alg"])
				}
				// jwtOpts.Key is the secret key for HS256/HS512 etc.
				return []byte(jwtOpts.Key), nil
			}, jwt.WithIssuer(jwtOpts.Issuer)) // Add WithIssuer if Issuer is configured in JWTOptions and needs validation.
			// Note: JWTOptions currently doesn't have an Issuer field. If added, uncomment and use it.
			// jwt.WithAudience(audience), jwt.WithSubject(subject) can also be added if needed.

			if err != nil {
				if errors.Is(err, jwt.ErrTokenMalformed) {
					return nil, errno.ErrorUnauthorized("token_malformed: %s", err.Error())
				} else if errors.Is(err, jwt.ErrTokenExpired) {
					return nil, errno.ErrorUnauthorized("token_expired: %s", err.Error())
				} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
					return nil, errno.ErrorUnauthorized("token_not_yet_valid: %s", err.Error())
				} else if verr, ok := err.(*jwt.ValidationError); ok && errors.Is(verr.Inner, jwt.ErrSignatureInvalid) {
					return nil, errno.ErrorUnauthorized("token_signature_invalid: %s", err.Error())
				}
				return nil, errno.ErrorUnauthorized("token_invalid: %s", err.Error())
			}

			if !token.Valid || claims == nil {
				return nil, errno.ErrorUnauthorized("token_claims_invalid: Token or claims are invalid")
			}

			// Token is valid. Store UserID (and other claims if needed) in context.
			if claims.CustomClaims.UserID == "" {
				// Depending on requirements, an empty UserID in a valid token might be an error.
				return nil, errno.ErrorForbidden("claims_missing_userid: UserID is missing from token claims")
			}
			newCtx := authn.SetUserIDInContext(ctx, claims.CustomClaims.UserID)

			return handler(newCtx, req)
		}
	}
}
