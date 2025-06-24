package authn

import (
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims contains custom data to be stored in the JWT.
// This can be extended with more application-specific claims like roles, permissions, etc.
type CustomClaims struct {
	UserID string `json:"userID,omitempty"` // Example: User's unique identifier
	// Username string `json:"username,omitempty"` // Example: User's username
	// Roles    []string `json:"roles,omitempty"`    // Example: User roles
}

// AppClaims represents the structure of our JWT claims, embedding standard registered claims
// from golang-jwt/jwt/v5 and our CustomClaims.
type AppClaims struct {
	jwt.RegisteredClaims
	CustomClaims
}
