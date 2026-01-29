// Package middleware provides HTTP middleware for the LLM-Proxy.
package middleware

import (
	"net/http"
	"strings"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// AdminMiddleware handles admin API key authentication
type AdminMiddleware struct {
	config *config.Config
	logger *logger.Logger
}

// NewAdminMiddleware creates a new admin middleware
func NewAdminMiddleware(cfg *config.Config, log *logger.Logger) *AdminMiddleware {
	return &AdminMiddleware{
		config: cfg,
		logger: log,
	}
}

// Authenticate validates the admin API key
func (m *AdminMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get API key from header
		apiKey := r.Header.Get("X-Admin-API-Key")

		// Also check Authorization header as fallback
		if apiKey == "" {
			authHeader := r.Header.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				apiKey = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// Validate API key
		if apiKey == "" {
			m.logger.Warn("Admin request without API key")
			m.respondUnauthorized(w, "missing admin API key")
			return
		}

		// Check if API key is valid (check against all configured admin keys)
		validKey := false
		for _, validAPIKey := range m.config.Admin.APIKeys {
			if apiKey == validAPIKey {
				validKey = true
				break
			}
		}

		if !validKey {
			m.logger.Warnf("Invalid admin API key attempt: %s", maskAPIKey(apiKey))
			m.respondUnauthorized(w, "invalid admin API key")
			return
		}

		// API key is valid
		m.logger.Debugf("Admin authenticated successfully")
		next.ServeHTTP(w, r)
	})
}

// respondUnauthorized sends an unauthorized response
func (m *AdminMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`{"error":"unauthorized","message":"` + message + `"}`))
}

// maskAPIKey masks an API key for logging
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
