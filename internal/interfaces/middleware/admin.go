// Package middleware provides HTTP middleware for the LLM-Proxy.
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
			if subtle.ConstantTimeCompare([]byte(apiKey), []byte(validAPIKey)) == 1 {
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

		// Add admin auth info to context for RequestLoggerMiddleware
		ctx := context.WithValue(r.Context(), "admin_authenticated", true)
		// Update the mutable AuthInfo struct so the request logger middleware can read it
		SetAuthInfo(r.Context(), "admin", "admin", nil)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// respondUnauthorized sends an unauthorized response
func (m *AdminMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	resp, err := json.Marshal(map[string]string{"error": "unauthorized", "message": message})
	if err != nil {
		m.logger.Errorf(err, "Failed to marshal admin unauthorized response")
		w.Write([]byte(`{"error":"unauthorized"}`))
		return
	}
	if _, err := w.Write(resp); err != nil {
		m.logger.Errorf(err, "Failed to write admin unauthorized response")
	}
}

// maskAPIKey masks an API key for logging
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
