package authn

import (
	"fmt"
	"time"

	"github.com/costa92/go-protoc/v2/pkg/options" // For JWTOptions
	"github.com/golang-jwt/jwt/v5"
)

// TokenGenerator is responsible for generating JWT tokens.
type TokenGenerator struct {
	jwtOpts *options.JWTOptions
}

// NewTokenGenerator creates a new TokenGenerator.
func NewTokenGenerator(jwtOpts *options.JWTOptions) (*TokenGenerator, error) {
	if jwtOpts == nil {
		return nil, fmt.Errorf("JWTOptions cannot be nil")
	}
	if jwtOpts.Key == "" {
		return nil, fmt.Errorf("JWT key cannot be empty")
	}
	// Potentially validate signing method here if only a subset is supported by this generator
	if _, ok := jwt.GetSigningMethod(jwtOpts.SigningMethod).(*jwt.SigningMethodHMAC); !ok && jwtOpts.SigningMethod != "" {
		// Assuming HS variants for now as key is a simple string.
		// For RSA/ECDSA, key would be *rsa.PrivateKey or *ecdsa.PrivateKey.
		// The JWTOptions.Key is a string, suitable for HMAC.
		// If other methods are specified in config, this check needs to be more robust
		// or the key parsing needs to handle different types.
		// For "HS512" default, this is fine.
	}
	return &TokenGenerator{jwtOpts: jwtOpts}, nil
}

// GenerateToken creates a new JWT token with the given custom claims data.
// For now, it specifically takes userID. This can be made more generic if needed.
func (tg *TokenGenerator) GenerateToken(userID string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(tg.jwtOpts.Expired)

	claims := AppClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			// Issuer: tg.jwtOpts.Issuer, // JWTOptions doesn't have Issuer yet. Add if needed.
			// Subject: userID, // Subject can also be userID.
		},
		CustomClaims: CustomClaims{
			UserID: userID,
		},
	}

	// Get the signing method
	signingMethod := jwt.GetSigningMethod(tg.jwtOpts.SigningMethod)
	if signingMethod == nil {
		return "", fmt.Errorf("unsupported JWT signing method: %s", tg.jwtOpts.SigningMethod)
	}

	token := jwt.NewWithClaims(signingMethod, claims)

	// Sign the token with the secret key
	// JWTOptions.Key is a string, needs to be []byte for HMAC
	signedToken, err := token.SignedString([]byte(tg.jwtOpts.Key))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
