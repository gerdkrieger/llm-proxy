// Package api provides admin HTTP handlers.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/application/caching"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// AdminHandler handles admin API requests
type AdminHandler struct {
	clientRepo      *repositories.OAuthClientRepository
	tokenRepo       *repositories.OAuthTokenRepository
	requestLogRepo  *repositories.RequestLogRepository
	filterMatchRepo *repositories.FilterMatchRepository
	cacheService    *caching.Service
	providerMgr     *providers.ProviderManager
	logger          *logger.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	clientRepo *repositories.OAuthClientRepository,
	tokenRepo *repositories.OAuthTokenRepository,
	requestLogRepo *repositories.RequestLogRepository,
	filterMatchRepo *repositories.FilterMatchRepository,
	cacheService *caching.Service,
	providerMgr *providers.ProviderManager,
	log *logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		clientRepo:      clientRepo,
		tokenRepo:       tokenRepo,
		requestLogRepo:  requestLogRepo,
		filterMatchRepo: filterMatchRepo,
		cacheService:    cacheService,
		providerMgr:     providerMgr,
		logger:          log,
	}
}

// ============================================================================
// CLIENT MANAGEMENT
// ============================================================================

// CreateClientRequest represents a request to create an OAuth client
type CreateClientRequest struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	Name         string   `json:"name"`
	RedirectURIs []string `json:"redirect_uris"`
	GrantTypes   []string `json:"grant_types"`
	DefaultScope string   `json:"default_scope"`
	RateLimitRPM *int     `json:"rate_limit_rpm,omitempty"`
	RateLimitRPD *int     `json:"rate_limit_rpd,omitempty"`
}

// UpdateClientRequest represents a request to update an OAuth client
type UpdateClientRequest struct {
	Name         *string  `json:"name,omitempty"`
	RedirectURIs []string `json:"redirect_uris,omitempty"`
	GrantTypes   []string `json:"grant_types,omitempty"`
	DefaultScope *string  `json:"default_scope,omitempty"`
	RateLimitRPM *int     `json:"rate_limit_rpm,omitempty"`
	RateLimitRPD *int     `json:"rate_limit_rpd,omitempty"`
	Enabled      *bool    `json:"enabled,omitempty"`
}

// ClientResponse represents an OAuth client response
type ClientResponse struct {
	ID           string    `json:"id"`
	ClientID     string    `json:"client_id"`
	Name         string    `json:"name"`
	RedirectURIs []string  `json:"redirect_uris"`
	GrantTypes   []string  `json:"grant_types"`
	DefaultScope string    `json:"default_scope"`
	RateLimitRPM *int      `json:"rate_limit_rpm"`
	RateLimitRPD *int      `json:"rate_limit_rpd"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListClients lists all OAuth clients
// GET /admin/clients
func (h *AdminHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// TODO: Add proper pagination with query params (limit/offset)
	// For now, use a safe limit of 1000 clients
	const maxClients = 1000

	h.logger.Info("Admin: Listing all OAuth clients")

	// Get clients from database
	dbClients, err := h.clientRepo.List(ctx, maxClients, 0)
	if err != nil {
		h.logger.Errorf(err, "Failed to list clients")
		h.respondError(w, http.StatusInternalServerError, "failed to list clients")
		return
	}

	// Convert to response format
	clients := make([]ClientResponse, 0, len(dbClients))
	for _, client := range dbClients {
		clients = append(clients, ClientResponse{
			ID:           client.ID.String(),
			ClientID:     client.ClientID,
			Name:         client.Name,
			RedirectURIs: client.RedirectURIs,
			GrantTypes:   client.GrantTypes,
			DefaultScope: client.DefaultScope,
			RateLimitRPM: client.RateLimitRPM,
			RateLimitRPD: client.RateLimitRPD,
			Enabled:      client.Enabled,
			CreatedAt:    client.CreatedAt,
			UpdatedAt:    client.UpdatedAt,
		})
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"clients": clients,
		"total":   len(clients),
	})
}

// GetClient gets a specific OAuth client
// GET /admin/clients/{client_id}
func (h *AdminHandler) GetClient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clientID := chi.URLParam(r, "client_id")

	h.logger.Infof("Admin: Getting OAuth client: %s", clientID)

	client, err := h.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "client not found")
		return
	}

	response := ClientResponse{
		ID:           client.ID.String(),
		ClientID:     client.ClientID,
		Name:         client.Name,
		RedirectURIs: client.RedirectURIs,
		GrantTypes:   client.GrantTypes,
		DefaultScope: client.DefaultScope,
		RateLimitRPM: client.RateLimitRPM,
		RateLimitRPD: client.RateLimitRPD,
		Enabled:      client.Enabled,
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// CreateClient creates a new OAuth client
// POST /admin/clients
func (h *AdminHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate request
	if req.ClientID == "" || req.ClientSecret == "" || req.Name == "" {
		h.respondError(w, http.StatusBadRequest, "client_id, client_secret, and name are required")
		return
	}

	h.logger.Infof("Admin: Creating OAuth client: %s", req.ClientID)

	// Create client
	client := &repositories.OAuthClient{
		ID:           uuid.New(),
		ClientID:     req.ClientID,
		ClientSecret: req.ClientSecret, // Will be hashed by repository
		Name:         req.Name,
		RedirectURIs: req.RedirectURIs,
		GrantTypes:   req.GrantTypes,
		DefaultScope: req.DefaultScope,
		RateLimitRPM: req.RateLimitRPM,
		RateLimitRPD: req.RateLimitRPD,
		Enabled:      true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.clientRepo.Create(ctx, client); err != nil {
		h.logger.Errorf(err, "Failed to create client")
		h.respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	response := ClientResponse{
		ID:           client.ID.String(),
		ClientID:     client.ClientID,
		Name:         client.Name,
		RedirectURIs: client.RedirectURIs,
		GrantTypes:   client.GrantTypes,
		DefaultScope: client.DefaultScope,
		RateLimitRPM: client.RateLimitRPM,
		RateLimitRPD: client.RateLimitRPD,
		Enabled:      client.Enabled,
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// UpdateClient updates an OAuth client
// PATCH /admin/clients/{client_id}
func (h *AdminHandler) UpdateClient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clientID := chi.URLParam(r, "client_id")

	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.Infof("Admin: Updating OAuth client: %s", clientID)

	// Get existing client
	client, err := h.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "client not found")
		return
	}

	// Update fields
	if req.Name != nil {
		client.Name = *req.Name
	}
	if req.RedirectURIs != nil {
		client.RedirectURIs = req.RedirectURIs
	}
	if req.GrantTypes != nil {
		client.GrantTypes = req.GrantTypes
	}
	if req.DefaultScope != nil {
		client.DefaultScope = *req.DefaultScope
	}
	if req.RateLimitRPM != nil {
		client.RateLimitRPM = req.RateLimitRPM
	}
	if req.RateLimitRPD != nil {
		client.RateLimitRPD = req.RateLimitRPD
	}
	if req.Enabled != nil {
		client.Enabled = *req.Enabled
	}
	client.UpdatedAt = time.Now()

	// Update in database
	if err := h.clientRepo.Update(ctx, client); err != nil {
		h.logger.Errorf(err, "Failed to update client")
		h.respondError(w, http.StatusInternalServerError, "failed to update client")
		return
	}

	response := ClientResponse{
		ID:           client.ID.String(),
		ClientID:     client.ClientID,
		Name:         client.Name,
		RedirectURIs: client.RedirectURIs,
		GrantTypes:   client.GrantTypes,
		DefaultScope: client.DefaultScope,
		RateLimitRPM: client.RateLimitRPM,
		RateLimitRPD: client.RateLimitRPD,
		Enabled:      client.Enabled,
		CreatedAt:    client.CreatedAt,
		UpdatedAt:    client.UpdatedAt,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// DeleteClient deletes an OAuth client
// DELETE /admin/clients/{client_id}
func (h *AdminHandler) DeleteClient(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clientID := chi.URLParam(r, "client_id")

	h.logger.Infof("Admin: Deleting OAuth client: %s", clientID)

	// Get client first
	client, err := h.clientRepo.GetByClientID(ctx, clientID)
	if err != nil {
		h.respondError(w, http.StatusNotFound, "client not found")
		return
	}

	// Delete client
	if err := h.clientRepo.Delete(ctx, client.ID); err != nil {
		h.logger.Errorf(err, "Failed to delete client")
		h.respondError(w, http.StatusInternalServerError, "failed to delete client")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "client deleted successfully",
	})
}

// ============================================================================
// CACHE MANAGEMENT
// ============================================================================

// GetCacheStats returns cache statistics
// GET /admin/cache/stats
func (h *AdminHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.cacheService.GetStats()
	hitRate := h.cacheService.GetHitRate()

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"hits":     stats.Hits,
		"misses":   stats.Misses,
		"errors":   stats.Errors,
		"hit_rate": hitRate,
	})
}

// ClearCache clears all cache entries
// POST /admin/cache/clear
func (h *AdminHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Warn("Admin: Clearing all cache")

	if err := h.cacheService.Clear(ctx); err != nil {
		h.logger.Errorf(err, "Failed to clear cache")
		h.respondError(w, http.StatusInternalServerError, "failed to clear cache")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]string{
		"message": "cache cleared successfully",
	})
}

// InvalidateCacheByModel invalidates cache for a specific model
// POST /admin/cache/invalidate/{model}
func (h *AdminHandler) InvalidateCacheByModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	model := chi.URLParam(r, "model")

	h.logger.Infof("Admin: Invalidating cache for model: %s", model)

	count, err := h.cacheService.InvalidateByModel(ctx, model)
	if err != nil {
		h.logger.Errorf(err, "Failed to invalidate cache")
		h.respondError(w, http.StatusInternalServerError, "failed to invalidate cache")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":         "cache invalidated successfully",
		"entries_removed": count,
	})
}

// ============================================================================
// USAGE STATISTICS
// ============================================================================

// GetUsageStats returns usage statistics
// GET /admin/stats/usage
func (h *AdminHandler) GetUsageStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameters
	clientID := r.URL.Query().Get("client_id")
	model := r.URL.Query().Get("model")

	h.logger.Info("Admin: Getting usage statistics")

	stats, err := h.requestLogRepo.GetStatistics(ctx, clientID, model, time.Time{}, time.Time{})
	if err != nil {
		h.logger.Errorf(err, "Failed to get statistics")
		h.respondError(w, http.StatusInternalServerError, "failed to get statistics")
		return
	}

	h.respondJSON(w, http.StatusOK, stats)
}

// ============================================================================
// PROVIDER STATUS
// ============================================================================

// GetProviderStatus returns provider health status
// GET /admin/providers/status
func (h *AdminHandler) GetProviderStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Admin: Getting provider status")

	// Check provider health
	err := h.providerMgr.Health(ctx)
	healthy := err == nil

	status := map[string]interface{}{
		"healthy":        healthy,
		"provider_count": h.providerMgr.GetProviderCount(),
		"models":         h.providerMgr.GetAvailableModels(),
	}

	if err != nil {
		status["error"] = err.Error()
	}

	h.respondJSON(w, http.StatusOK, status)
}

// GetProviderDetails returns detailed information about all configured providers
// GET /admin/providers
func (h *AdminHandler) GetProviderDetails(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Admin: Getting detailed provider information")

	providers := make([]map[string]interface{}, 0)

	// Claude Provider Info
	if h.providerMgr != nil {
		ctx := r.Context()

		// Try to get config (we'll need to add a method to provider manager)
		claudeInfo := map[string]interface{}{
			"id":       "claude",
			"name":     "Anthropic Claude",
			"type":     "claude",
			"enabled":  true,
			"status":   "unknown",
			"models":   []string{},
			"api_keys": 0,
		}

		// Get Claude models
		allModels := h.providerMgr.GetAvailableModels()
		claudeModels := make([]string, 0)
		openaiModels := make([]string, 0)

		for _, model := range allModels {
			if strings.Contains(model, "claude") {
				claudeModels = append(claudeModels, model)
			} else if strings.Contains(model, "gpt") {
				openaiModels = append(openaiModels, model)
			}
		}

		claudeInfo["models"] = claudeModels

		// Test Claude health
		if err := h.providerMgr.Health(ctx); err == nil {
			claudeInfo["status"] = "healthy"
		} else {
			claudeInfo["status"] = "unhealthy"
			claudeInfo["error"] = err.Error()
		}

		// Get provider count
		if len(claudeModels) > 0 {
			claudeInfo["api_keys"] = 1 // At least one Claude key
			providers = append(providers, claudeInfo)
		}

		// OpenAI Provider Info
		if len(openaiModels) > 0 {
			openaiInfo := map[string]interface{}{
				"id":       "openai",
				"name":     "OpenAI",
				"type":     "openai",
				"enabled":  true,
				"status":   "healthy", // We assume healthy if models are configured
				"models":   openaiModels,
				"api_keys": 1,
			}
			providers = append(providers, openaiInfo)
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"providers": providers,
		"total":     len(providers),
	})
}

// ============================================================================
// FILTER MATCHES
// ============================================================================

// FilterMatchResponse represents a filter match in API responses
type FilterMatchResponse struct {
	ID          string    `json:"id"`
	RequestID   string    `json:"request_id"`
	ClientID    *string   `json:"client_id,omitempty"`
	ClientName  *string   `json:"client_name,omitempty"`
	FilterID    *int      `json:"filter_id,omitempty"` // NULL for attachment redactions
	Model       string    `json:"model"`
	Provider    string    `json:"provider"`
	Pattern     string    `json:"pattern"`
	Replacement string    `json:"replacement"`
	FilterType  string    `json:"filter_type"`
	MatchCount  int       `json:"match_count"`
	MatchedText *string   `json:"matched_text,omitempty"`
	IPAddress   *string   `json:"ip_address,omitempty"`
	UserAgent   *string   `json:"user_agent,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// GetFilterMatches returns recent filter matches (blocked content)
// GET /admin/filters/matches
func (h *AdminHandler) GetFilterMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Info("Admin: Getting filter matches")

	// Get limit from query params (default 100, max 1000)
	limit := 100
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
			if limit > 1000 {
				limit = 1000
			}
			if limit < 1 {
				limit = 100
			}
		}
	}

	// Get matches from repository
	matches, err := h.filterMatchRepo.GetRecentMatches(ctx, limit)
	if err != nil {
		h.logger.Errorf(err, "Failed to get filter matches")
		h.respondError(w, http.StatusInternalServerError, "failed to get filter matches")
		return
	}

	// Convert to response format and enrich with client names
	responses := make([]FilterMatchResponse, 0, len(matches))
	for _, match := range matches {
		resp := FilterMatchResponse{
			ID:          match.ID.String(),
			RequestID:   match.RequestID,
			FilterID:    match.FilterID,
			Model:       match.Model,
			Provider:    match.Provider,
			Pattern:     match.Pattern,
			Replacement: match.Replacement,
			FilterType:  match.FilterType,
			MatchCount:  match.MatchCount,
			MatchedText: match.MatchedText,
			IPAddress:   match.IPAddress,
			UserAgent:   match.UserAgent,
			CreatedAt:   match.CreatedAt,
		}

		// Get client name if client_id is present
		if match.ClientID != nil {
			resp.ClientID = new(string)
			*resp.ClientID = match.ClientID.String()

			// Try to fetch client name
			if client, err := h.clientRepo.GetByID(ctx, *match.ClientID); err == nil {
				resp.ClientName = &client.Name
			}
		}

		responses = append(responses, resp)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"matches": responses,
		"total":   len(responses),
	})
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// respondJSON sends a JSON response
func (h *AdminHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func (h *AdminHandler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "error",
		"message": message,
	})
}
