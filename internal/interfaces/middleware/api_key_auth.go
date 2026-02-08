// Package middleware provides HTTP middleware for request processing
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// APIKeyAuthMiddleware validates API key authentication
type APIKeyAuthMiddleware struct {
	clientRepo *repositories.OAuthClientRepository
	logger     *logger.Logger
}

// NewAPIKeyAuthMiddleware creates a new API key authentication middleware
func NewAPIKeyAuthMiddleware(clientRepo *repositories.OAuthClientRepository, logger *logger.Logger) *APIKeyAuthMiddleware {
	return &APIKeyAuthMiddleware{
		clientRepo: clientRepo,
		logger:     logger,
	}
}

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// ClientContextKey is the context key for storing authenticated client
	ClientContextKey contextKey = "authenticated_client"
)

// Authenticate validates the API key and sets the client in context
func (m *APIKeyAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health and metrics endpoints
		if r.URL.Path == "/health" || r.URL.Path == "/metrics" || strings.HasPrefix(r.URL.Path, "/admin/") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if already authenticated (by previous middleware)
		if client, ok := GetClientFromContext(r.Context()); ok && client != nil {
			// Already authenticated by previous middleware (e.g., static API key)
			next.ServeHTTP(w, r)
			return
		}

		// Extract API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header - pass through to next middleware (OAuth)
			next.ServeHTTP(w, r)
			return
		}

		// Support both "Bearer <token>" and plain API key
		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		apiKey = strings.TrimSpace(apiKey)

		if apiKey == "" {
			// Empty key - pass through to next middleware
			next.ServeHTTP(w, r)
			return
		}

		// If the key looks like a JWT (starts with "eyJ"), skip DB lookup and pass to OAuth middleware
		if strings.HasPrefix(apiKey, "eyJ") {
			next.ServeHTTP(w, r)
			return
		}

		// Try to find client by validating API key against all enabled clients
		m.logger.Debugf("APIKeyAuth: Attempting DB-based key validation for key prefix: %s...", apiKey[:min(20, len(apiKey))])

		clients, err := m.clientRepo.ListWithSecrets(r.Context())
		if err != nil {
			m.logger.Error(err, "APIKeyAuth: Failed to list clients for authentication")
			http.Error(w, `{"error":"internal_error","message":"Authentication service unavailable"}`, http.StatusInternalServerError)
			return
		}

		var authenticatedClient *repositories.OAuthClient
		for _, client := range clients {
			// Validate API key against client secret hash (bcrypt)
			if m.clientRepo.ValidateSecret(client, apiKey) {
				authenticatedClient = client
				break
			}
		}

		if authenticatedClient == nil {
			m.logger.Debugf("APIKeyAuth: No matching client found for key prefix: %s..., passing to next middleware", apiKey[:min(20, len(apiKey))])
			// No matching client found - pass through to OAuth middleware
			// (might be a valid OAuth token)
			next.ServeHTTP(w, r)
			return
		}

		m.logger.Infof("APIKeyAuth: Authenticated client: %s (%s)", authenticatedClient.ClientID, authenticatedClient.Name)

		// Store client in context using the SAME keys as static APIKeyMiddleware
		// so the OAuth middleware recognizes the request as already authenticated.
		ctx := r.Context()
		ctx = context.WithValue(ctx, ClientIDKey, authenticatedClient.ClientID)
		ctx = context.WithValue(ctx, ScopeKey, authenticatedClient.DefaultScope)
		// Also store the full client object for downstream handlers (e.g., request logger)
		ctx = context.WithValue(ctx, ClientContextKey, authenticatedClient)

		// Update the mutable AuthInfo struct (set by RequestLoggerMiddleware) so the
		// request logger can read auth info even though we create a new request via WithContext.
		SetAuthInfo(r.Context(), "api_key", authenticatedClient.Name, &authenticatedClient.ID)

		// Continue with authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetClientFromContext extracts the authenticated client from context
func GetClientFromContext(ctx context.Context) (*repositories.OAuthClient, bool) {
	client, ok := ctx.Value(ClientContextKey).(*repositories.OAuthClient)
	return client, ok
}
