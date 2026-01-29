// Package api provides OAuth HTTP handlers.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/llm-proxy/llm-proxy/internal/application/oauth"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// OAuthHandler handles OAuth 2.0 endpoints
type OAuthHandler struct {
	service *oauth.Service
	logger  *logger.Logger
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(service *oauth.Service, log *logger.Logger) *OAuthHandler {
	return &OAuthHandler{
		service: service,
		logger:  log,
	}
}

// Token handles token requests
// POST /oauth/token
func (h *OAuthHandler) Token(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req oauth.TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.GrantType == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "grant_type is required")
		return
	}

	if req.ClientID == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "client_id is required")
		return
	}

	// Handle token request
	resp, err := h.service.HandleTokenRequest(ctx, &req)
	if err != nil {
		h.logger.Error(err, "Token request failed")
		h.respondError(w, http.StatusUnauthorized, "invalid_client", err.Error())
		return
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// Revoke handles token revocation
// POST /oauth/revoke
func (h *OAuthHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req struct {
		Token string `json:"token"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Token == "" {
		h.respondError(w, http.StatusBadRequest, "invalid_request", "token is required")
		return
	}

	// Revoke token
	if err := h.service.RevokeToken(ctx, req.Token); err != nil {
		h.logger.Error(err, "Token revocation failed")
		// OAuth spec: return 200 even if token doesn't exist
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "revoked"})
}

// respondError sends an OAuth error response
func (h *OAuthHandler) respondError(w http.ResponseWriter, statusCode int, errorCode, errorDescription string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": errorDescription,
	})
}
