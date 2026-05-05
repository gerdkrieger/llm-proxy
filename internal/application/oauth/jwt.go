// Package oauth provides OAuth 2.0 implementation.
package oauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/config"
)

const (
	// MinJWTSecretLength defines the minimum required length for JWT secrets.
	// Using 32 characters ensures 256 bits of entropy.
	// NOTE: This MUST stay in sync with config.go validate() which also enforces >=32.
	MinJWTSecretLength = 32
)

// TokenGenerator handles JWT token generation and validation
type TokenGenerator struct {
	config config.OAuthConfig
}

// NewTokenGenerator creates a new token generator with JWT secret validation
func NewTokenGenerator(cfg config.OAuthConfig) (*TokenGenerator, error) {
	// Validate JWT secret length for security
	if len(cfg.JWTSecret) < MinJWTSecretLength {
		return nil, fmt.Errorf(
			"JWT secret is too short: got %d characters, minimum required is %d characters (512 bits). "+
				"Generate a secure secret with: openssl rand -base64 64 | tr -d '\\n'",
			len(cfg.JWTSecret),
			MinJWTSecretLength,
		)
	}

	return &TokenGenerator{
		config: cfg,
	}, nil
}

// Claims represents JWT claims
type Claims struct {
	ClientID string `json:"client_id"`
	Scope    string `json:"scope"`
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token
func (tg *TokenGenerator) GenerateAccessToken(clientID, scope string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(tg.config.AccessTokenTTL)

	claims := Claims{
		ClientID: clientID,
		Scope:    scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tg.config.Issuer,
			Subject:   clientID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tg.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// GenerateRefreshToken generates a new refresh token
func (tg *TokenGenerator) GenerateRefreshToken(clientID string) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(tg.config.RefreshTokenTTL)

	claims := Claims{
		ClientID: clientID,
		Scope:    "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    tg.config.Issuer,
			Subject:   clientID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(tg.config.JWTSecret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token and returns the claims
func (tg *TokenGenerator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tg.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Additional validation
	if claims.Issuer != tg.config.Issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	return claims, nil
}

// HasScope checks if the token has a specific scope
func (c *Claims) HasScope(requiredScope string) bool {
	// Simple scope check (can be extended for complex scope logic)
	scopes := parseScopes(c.Scope)
	for _, scope := range scopes {
		if scope == requiredScope {
			return true
		}
	}
	return false
}

// parseScopes splits scope string into individual scopes
func parseScopes(scopeString string) []string {
	if scopeString == "" {
		return []string{}
	}

	// Scopes are comma or space separated
	scopes := []string{}
	for _, s := range splitByCommaOrSpace(scopeString) {
		if s != "" {
			scopes = append(scopes, s)
		}
	}
	return scopes
}

// splitByCommaOrSpace splits a string by comma or space
func splitByCommaOrSpace(s string) []string {
	result := []string{}
	current := ""

	for _, ch := range s {
		if ch == ',' || ch == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
