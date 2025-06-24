package authn

import "context"

// Define a context key type to avoid collisions.
type contextKey string

const (
	// UserIDKey is the key used to store the UserID in the context after JWT validation.
	UserIDKey contextKey = "UserID"
	// UsernameKey is the key used to store the Username in the context.
	// UsernameKey contextKey = "Username" // Example if username is also in claims
	// ClaimsKey is the key used to store the full AppClaims struct in the context.
	// ClaimsKey contextKey = "AppClaims" // Example if full claims are needed
)

// GetUserIDFromContext retrieves the UserID from the context.
// Returns the UserID and true if found, otherwise empty string and false.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// SetUserIDInContext sets the UserID in the context.
func SetUserIDInContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// Example for full claims if needed later:
// func GetClaimsFromContext(ctx context.Context) (*AppClaims, bool) {
// 	claims, ok := ctx.Value(ClaimsKey).(*AppClaims)
// 	return claims, ok
// }

// func SetClaimsInContext(ctx context.Context, claims *AppClaims) context.Context {
// 	return context.WithValue(ctx, ClaimsKey, claims)
// }
