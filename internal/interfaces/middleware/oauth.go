// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/llm-proxy/llm-proxy/internal/application/oauth"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ContextKey type for context keys
type ContextKey string

const (
	// ClientIDKey is the context key for client ID
	ClientIDKey ContextKey = "client_id"
	// ScopeKey is the context key for scope
	ScopeKey ContextKey = "scope"
	// ClaimsKey is the context key for JWT claims
	ClaimsKey ContextKey = "claims"
)

// OAuthMiddleware validates OAuth access tokens
type OAuthMiddleware struct {
	service *oauth.Service
	logger  *logger.Logger
}

// NewOAuthMiddleware creates a new OAuth middleware
func NewOAuthMiddleware(service *oauth.Service, log *logger.Logger) *OAuthMiddleware {
	return &OAuthMiddleware{
		service: service,
		logger:  log,
	}
}

// Authenticate validates the access token and adds claims to context
func (m *OAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if already authenticated by API key middleware
		clientID := GetClientID(r.Context())
		m.logger.Infof("OAuth middleware: checking context, clientID='%s'", clientID)
		if clientID != "" {
			m.logger.Infof("Request already authenticated by API key middleware: %s", clientID)
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.respondError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.respondError(w, http.StatusUnauthorized, "invalid authorization header format")
			return
		}

		accessToken := parts[1]

		// Validate token
		claims, err := m.service.ValidateAccessToken(r.Context(), accessToken)
		if err != nil {
			m.logger.Warnf("Token validation failed: %v", err)
			m.respondError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}

		// Add claims to context
		ctx := r.Context()
		ctx = context.WithValue(ctx, ClientIDKey, claims.ClientID)
		ctx = context.WithValue(ctx, ScopeKey, claims.Scope)
		ctx = context.WithValue(ctx, ClaimsKey, claims)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireScope creates middleware that requires a specific scope
func (m *OAuthMiddleware) RequireScope(requiredScope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// First check for OAuth claims (JWT token)
			claims, hasOAuthClaims := r.Context().Value(ClaimsKey).(*oauth.Claims)
			if hasOAuthClaims {
				// OAuth flow - check scope in claims
				if !claims.HasScope(requiredScope) {
					m.logger.Warnf("Client %s missing required scope: %s", claims.ClientID, requiredScope)
					m.respondError(w, http.StatusForbidden, "insufficient scope")
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// API key flow - check scope string in context
			scopeStr := GetScope(r.Context())
			clientID := GetClientID(r.Context())
			if clientID == "" {
				m.respondError(w, http.StatusUnauthorized, "no authentication context")
				return
			}

			// Parse scope string and check if it contains required scope
			scopes := strings.Fields(scopeStr)
			hasScope := false
			for _, scope := range scopes {
				if scope == requiredScope {
					hasScope = true
					break
				}
			}

			if !hasScope {
				m.logger.Warnf("Client %s missing required scope: %s", clientID, requiredScope)
				m.respondError(w, http.StatusForbidden, "insufficient scope")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// respondError sends an error response
func (m *OAuthMiddleware) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// GetClientID extracts client ID from context
func GetClientID(ctx context.Context) string {
	if clientID, ok := ctx.Value(ClientIDKey).(string); ok {
		return clientID
	}
	return ""
}

// GetScope extracts scope from context
func GetScope(ctx context.Context) string {
	if scope, ok := ctx.Value(ScopeKey).(string); ok {
		return scope
	}
	return ""
}

// GetClaims extracts JWT claims from context
func GetClaims(ctx context.Context) *oauth.Claims {
	if claims, ok := ctx.Value(ClaimsKey).(*oauth.Claims); ok {
		return claims
	}
	return nil
}
