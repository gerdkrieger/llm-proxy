// Package middleware provides HTTP middleware components.
package middleware

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// APIKeyMiddleware validates static API keys for OpenAI-compatible clients
type APIKeyMiddleware struct {
	apiKeys []config.ClientAPIKeyConfig
	logger  *logger.Logger
}

// NewAPIKeyMiddleware creates a new API key middleware
func NewAPIKeyMiddleware(cfg *config.Config, log *logger.Logger) *APIKeyMiddleware {
	return &APIKeyMiddleware{
		apiKeys: cfg.ClientAPIKeys,
		logger:  log,
	}
}

// Authenticate validates static API keys and adds client context
// If the key is not a static API key (doesn't start with "sk-llm-proxy-"),
// this middleware passes through to allow OAuth middleware to handle it
func (m *APIKeyMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			// No auth header - let OAuth middleware handle it
			next.ServeHTTP(w, r)
			return
		}

		// Parse Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format - let OAuth middleware handle it
			next.ServeHTTP(w, r)
			return
		}

		token := parts[1]

		// Check if this is a static API key (starts with "sk-llm-proxy-")
		if !strings.HasPrefix(token, "sk-llm-proxy-") {
			// Not a static API key - pass through to next middleware
			next.ServeHTTP(w, r)
			return
		}

		tokenPrefix := token
		if len(tokenPrefix) > 20 {
			tokenPrefix = token[:20]
		}
		m.logger.Debugf("Detected static API key format, validating: %s...", tokenPrefix)

		// Validate static API key
		clientConfig := m.validateAPIKey(token)
		if clientConfig == nil {
			m.logger.Warnf("Invalid static API key attempted: %s...", tokenPrefix)
			m.respondError(w, http.StatusUnauthorized, "invalid API key")
			return
		}

		// Check if key is enabled
		if !clientConfig.Enabled {
			m.logger.Warnf("Disabled API key attempted: %s", clientConfig.Name)
			m.respondError(w, http.StatusForbidden, "API key is disabled")
			return
		}

		// Add client info to context (compatible with OAuth context keys)
		ctx := r.Context()
		ctx = context.WithValue(ctx, ClientIDKey, clientConfig.Name)
		ctx = context.WithValue(ctx, ScopeKey, strings.Join(clientConfig.Scopes, " "))
		// Also set for RequestLoggerMiddleware
		ctx = context.WithValue(ctx, "api_key_name", clientConfig.Name)

		m.logger.Infof("Static API key authenticated successfully: %s", clientConfig.Name)

		// Update the mutable AuthInfo struct (set by RequestLoggerMiddleware) so the
		// request logger can read auth info even though we create a new request via WithContext.
		SetAuthInfo(r.Context(), "static_api_key", clientConfig.Name, nil)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateAPIKey checks if the provided key is valid and enabled
func (m *APIKeyMiddleware) validateAPIKey(key string) *config.ClientAPIKeyConfig {
	for _, apiKey := range m.apiKeys {
		if subtle.ConstantTimeCompare([]byte(apiKey.Key), []byte(key)) == 1 {
			return &apiKey
		}
	}
	return nil
}

// respondError sends an error response
func (m *APIKeyMiddleware) respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	resp, _ := json.Marshal(map[string]string{"error": message})
	w.Write(resp)
}
