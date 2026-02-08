// Package api provides admin HTTP handlers.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
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
	clientRepo         *repositories.OAuthClientRepository
	tokenRepo          *repositories.OAuthTokenRepository
	requestLogRepo     *repositories.RequestLogRepository
	filterMatchRepo    *repositories.FilterMatchRepository
	providerModelRepo  *repositories.ProviderModelRepository
	systemSettingsRepo *repositories.SystemSettingsRepository
	cacheService       *caching.Service
	providerMgr        *providers.ProviderManager
	logger             *logger.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	clientRepo *repositories.OAuthClientRepository,
	tokenRepo *repositories.OAuthTokenRepository,
	requestLogRepo *repositories.RequestLogRepository,
	filterMatchRepo *repositories.FilterMatchRepository,
	providerModelRepo *repositories.ProviderModelRepository,
	systemSettingsRepo *repositories.SystemSettingsRepository,
	cacheService *caching.Service,
	providerMgr *providers.ProviderManager,
	log *logger.Logger,
) *AdminHandler {
	return &AdminHandler{
		clientRepo:         clientRepo,
		tokenRepo:          tokenRepo,
		requestLogRepo:     requestLogRepo,
		filterMatchRepo:    filterMatchRepo,
		providerModelRepo:  providerModelRepo,
		systemSettingsRepo: systemSettingsRepo,
		cacheService:       cacheService,
		providerMgr:        providerMgr,
		logger:             log,
	}
}

// ============================================================================
// CLIENT MANAGEMENT
// ============================================================================

// CreateClientRequest represents a request to create an API client
type CreateClientRequest struct {
	ClientID      string   `json:"client_id"`
	ClientSecret  string   `json:"client_secret"`
	Name          string   `json:"name"`
	RedirectURIs  []string `json:"redirect_uris"`
	GrantTypes    []string `json:"grant_types"`
	DefaultScope  string   `json:"default_scope"`
	AllowedModels []string `json:"allowed_models,omitempty"` // null = all models, [] = none, ["model"] = specific
	RateLimitRPM  *int     `json:"rate_limit_rpm,omitempty"`
	RateLimitRPD  *int     `json:"rate_limit_rpd,omitempty"`
}

// UpdateClientRequest represents a request to update an API client
type UpdateClientRequest struct {
	Name          *string  `json:"name,omitempty"`
	RedirectURIs  []string `json:"redirect_uris,omitempty"`
	GrantTypes    []string `json:"grant_types,omitempty"`
	DefaultScope  *string  `json:"default_scope,omitempty"`
	AllowedModels []string `json:"allowed_models,omitempty"` // null = all models, [] = none, ["model"] = specific
	RateLimitRPM  *int     `json:"rate_limit_rpm,omitempty"`
	RateLimitRPD  *int     `json:"rate_limit_rpd,omitempty"`
	Enabled       *bool    `json:"enabled,omitempty"`
}

// ClientResponse represents an API client response
type ClientResponse struct {
	ID            string    `json:"id"`
	ClientID      string    `json:"client_id"`
	Name          string    `json:"name"`
	RedirectURIs  []string  `json:"redirect_uris"`
	GrantTypes    []string  `json:"grant_types"`
	DefaultScope  string    `json:"default_scope"`
	AllowedModels []string  `json:"allowed_models"` // Always include: null = all models, [] = none, ["model"] = specific
	RateLimitRPM  *int      `json:"rate_limit_rpm"`
	RateLimitRPD  *int      `json:"rate_limit_rpd"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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
			ID:            client.ID.String(),
			ClientID:      client.ClientID,
			Name:          client.Name,
			RedirectURIs:  client.RedirectURIs,
			GrantTypes:    client.GrantTypes,
			DefaultScope:  client.DefaultScope,
			AllowedModels: client.AllowedModels,
			RateLimitRPM:  client.RateLimitRPM,
			RateLimitRPD:  client.RateLimitRPD,
			Enabled:       client.Enabled,
			CreatedAt:     client.CreatedAt,
			UpdatedAt:     client.UpdatedAt,
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
		ID:            uuid.New(),
		ClientID:      req.ClientID,
		ClientSecret:  req.ClientSecret, // Will be hashed by repository
		Name:          req.Name,
		RedirectURIs:  req.RedirectURIs,
		GrantTypes:    req.GrantTypes,
		DefaultScope:  req.DefaultScope,
		AllowedModels: req.AllowedModels, // nil = all models allowed
		RateLimitRPM:  req.RateLimitRPM,
		RateLimitRPD:  req.RateLimitRPD,
		Enabled:       true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := h.clientRepo.Create(ctx, client); err != nil {
		h.logger.Errorf(err, "Failed to create client")
		h.respondError(w, http.StatusInternalServerError, "failed to create client")
		return
	}

	response := ClientResponse{
		ID:            client.ID.String(),
		ClientID:      client.ClientID,
		Name:          client.Name,
		RedirectURIs:  client.RedirectURIs,
		GrantTypes:    client.GrantTypes,
		DefaultScope:  client.DefaultScope,
		AllowedModels: client.AllowedModels,
		RateLimitRPM:  client.RateLimitRPM,
		RateLimitRPD:  client.RateLimitRPD,
		Enabled:       client.Enabled,
		CreatedAt:     client.CreatedAt,
		UpdatedAt:     client.UpdatedAt,
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
	if req.AllowedModels != nil {
		client.AllowedModels = req.AllowedModels
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
		ID:            client.ID.String(),
		ClientID:      client.ClientID,
		Name:          client.Name,
		RedirectURIs:  client.RedirectURIs,
		GrantTypes:    client.GrantTypes,
		DefaultScope:  client.DefaultScope,
		AllowedModels: client.AllowedModels,
		RateLimitRPM:  client.RateLimitRPM,
		RateLimitRPD:  client.RateLimitRPD,
		Enabled:       client.Enabled,
		CreatedAt:     client.CreatedAt,
		UpdatedAt:     client.UpdatedAt,
	}

	h.respondJSON(w, http.StatusOK, response)
}

// DeleteClient deletes an API client
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

	// Get enabled models from database (not config!)
	enabledModels := make([]string, 0)
	for _, providerID := range []string{"claude", "openai"} {
		models, err := h.providerModelRepo.GetEnabledByProvider(ctx, providerID)
		if err != nil {
			h.logger.Warnf("Failed to get enabled models for %s: %v", providerID, err)
			continue
		}
		for _, model := range models {
			enabledModels = append(enabledModels, model.ModelID)
		}
	}

	status := map[string]interface{}{
		"healthy":        healthy,
		"provider_count": h.providerMgr.GetProviderCount(),
		"models":         enabledModels,
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

	ctx := r.Context()
	providers := make([]map[string]interface{}, 0)

	// Get active provider IDs
	activeProviders := h.providerMgr.GetActiveProviderIDs()
	activeProviderMap := make(map[string]bool)
	for _, pid := range activeProviders {
		activeProviderMap[pid] = true
	}

	// Claude Provider Info
	if activeProviderMap["claude"] {
		claudeInfo := map[string]interface{}{
			"id":       "claude",
			"name":     "Anthropic Claude",
			"type":     "claude",
			"enabled":  true,
			"status":   "unknown",
			"models":   []string{},
			"api_keys": 1,
		}

		// Get ENABLED models from database
		enabledModels, err := h.providerModelRepo.GetEnabledByProvider(ctx, "claude")
		if err != nil {
			h.logger.Warnf("Failed to get enabled Claude models: %v", err)
		} else {
			modelIDs := make([]string, len(enabledModels))
			for i, model := range enabledModels {
				modelIDs[i] = model.ModelID
			}
			claudeInfo["models"] = modelIDs
		}

		// Test Claude health
		if err := h.providerMgr.Health(ctx); err == nil {
			claudeInfo["status"] = "healthy"
		} else {
			claudeInfo["status"] = "unhealthy"
			claudeInfo["error"] = err.Error()
		}

		providers = append(providers, claudeInfo)
	}

	// OpenAI Provider Info
	if activeProviderMap["openai"] {
		openaiInfo := map[string]interface{}{
			"id":       "openai",
			"name":     "OpenAI",
			"type":     "openai",
			"enabled":  true,
			"status":   "healthy",
			"models":   []string{},
			"api_keys": 1,
		}

		// Get ENABLED models from database
		enabledModels, err := h.providerModelRepo.GetEnabledByProvider(ctx, "openai")
		if err != nil {
			h.logger.Warnf("Failed to get enabled OpenAI models: %v", err)
		} else {
			modelIDs := make([]string, len(enabledModels))
			for i, model := range enabledModels {
				modelIDs[i] = model.ModelID
			}
			openaiInfo["models"] = modelIDs
		}

		providers = append(providers, openaiInfo)
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

// ============================================================================
// REQUEST LOGS (for Live Monitor)
// ============================================================================

// GetRequestLogs retrieves recent API request logs for monitoring
// GET /admin/requests?limit=50
func (h *AdminHandler) GetRequestLogs(w http.ResponseWriter, r *http.Request) {
	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 && parsedLimit <= 500 {
			limit = parsedLimit
		}
	}

	// Build filters
	filters := repositories.RequestLogFilters{
		Limit: limit,
	}

	// Get logs from repository
	logs, err := h.requestLogRepo.List(r.Context(), filters)
	if err != nil {
		h.logger.Warnf("Failed to retrieve request logs: %v", err)
		h.respondError(w, http.StatusInternalServerError, "failed to retrieve request logs")
		return
	}

	// Convert to response format (keep structure compatible with Live Monitor)
	type RequestLogResponse struct {
		ID           string    `json:"id"`
		ClientID     *string   `json:"client_id,omitempty"`
		ClientName   *string   `json:"client_name,omitempty"`
		CreatedAt    time.Time `json:"created_at"`
		Method       string    `json:"method"`
		Endpoint     string    `json:"endpoint"`
		StatusCode   int       `json:"status_code"`
		IPAddress    *string   `json:"ip_address"`
		UserAgent    *string   `json:"user_agent"`
		AuthType     *string   `json:"auth_type"`
		APIKeyName   *string   `json:"api_key_name"`
		DurationMS   int       `json:"duration_ms"`
		Model        string    `json:"model,omitempty"`
		Provider     string    `json:"provider,omitempty"`
		WasFiltered  bool      `json:"was_filtered"`
		FilterReason *string   `json:"filter_reason,omitempty"`
		ErrorMessage *string   `json:"error_message,omitempty"`
	}

	// Build a client name lookup map for resolving client_id to names
	clientNameMap := make(map[string]string)
	clients, err := h.clientRepo.List(r.Context(), 1000, 0)
	if err == nil {
		for _, c := range clients {
			clientNameMap[c.ID.String()] = c.Name
		}
	}

	response := make([]RequestLogResponse, 0, len(logs))
	for _, log := range logs {
		entry := RequestLogResponse{
			ID:           log.ID.String(),
			CreatedAt:    log.CreatedAt,
			Method:       log.Method,
			Endpoint:     log.Path,
			StatusCode:   log.StatusCode,
			IPAddress:    log.IPAddress,
			UserAgent:    log.UserAgent,
			AuthType:     log.AuthType,
			APIKeyName:   log.APIKeyName,
			DurationMS:   log.DurationMS,
			Model:        log.Model,
			Provider:     log.Provider,
			WasFiltered:  log.WasFiltered,
			FilterReason: log.FilterReason,
			ErrorMessage: log.ErrorMessage,
		}

		// Resolve client_id to client_name
		if log.ClientID != nil {
			clientIDStr := log.ClientID.String()
			entry.ClientID = &clientIDStr
			if name, ok := clientNameMap[clientIDStr]; ok {
				entry.ClientName = &name
			}
		}

		response = append(response, entry)
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  response,
		"total": len(response),
	})
}

// GetRequestLogDetails retrieves detailed information for a specific request
// GET /admin/requests/{id}
func (h *AdminHandler) GetRequestLogDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := chi.URLParam(r, "id")

	// Parse UUID
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request ID format")
		return
	}

	// Get full log details from repository
	log, err := h.requestLogRepo.GetByID(ctx, id)
	if err != nil {
		h.logger.Warnf("Failed to retrieve request log details for ID %s: %v", idStr, err)
		h.respondError(w, http.StatusNotFound, "request log not found")
		return
	}

	// Return full log with all details including request/response bodies
	type RequestLogDetailResponse struct {
		ID                string                 `json:"id"`
		ClientID          *string                `json:"client_id,omitempty"`
		ClientName        *string                `json:"client_name,omitempty"`
		RequestID         string                 `json:"request_id"`
		CreatedAt         time.Time              `json:"created_at"`
		Method            string                 `json:"method"`
		Path              string                 `json:"path"`
		StatusCode        int                    `json:"status_code"`
		DurationMS        int                    `json:"duration_ms"`
		IPAddress         *string                `json:"ip_address"`
		UserAgent         *string                `json:"user_agent"`
		AuthType          *string                `json:"auth_type"`
		APIKeyName        *string                `json:"api_key_name"`
		Model             string                 `json:"model,omitempty"`
		Provider          string                 `json:"provider,omitempty"`
		PromptTokens      int                    `json:"prompt_tokens,omitempty"`
		CompletionTokens  int                    `json:"completion_tokens,omitempty"`
		TotalTokens       int                    `json:"total_tokens,omitempty"`
		CostUSD           float64                `json:"cost_usd,omitempty"`
		WasFiltered       bool                   `json:"was_filtered"`
		FilterReason      *string                `json:"filter_reason,omitempty"`
		ErrorMessage      *string                `json:"error_message,omitempty"`
		RequestHeaders    map[string]interface{} `json:"request_headers,omitempty"`
		RequestBody       *string                `json:"request_body,omitempty"`
		ResponseHeaders   map[string]interface{} `json:"response_headers,omitempty"`
		ResponseBody      *string                `json:"response_body,omitempty"`
		ResponseSizeBytes *int64                 `json:"response_size_bytes,omitempty"`
	}

	response := RequestLogDetailResponse{
		ID:                log.ID.String(),
		RequestID:         log.RequestID,
		CreatedAt:         log.CreatedAt,
		Method:            log.Method,
		Path:              log.Path,
		StatusCode:        log.StatusCode,
		DurationMS:        log.DurationMS,
		IPAddress:         log.IPAddress,
		UserAgent:         log.UserAgent,
		AuthType:          log.AuthType,
		APIKeyName:        log.APIKeyName,
		Model:             log.Model,
		Provider:          log.Provider,
		PromptTokens:      log.PromptTokens,
		CompletionTokens:  log.CompletionTokens,
		TotalTokens:       log.TotalTokens,
		CostUSD:           log.CostUSD,
		WasFiltered:       log.WasFiltered,
		FilterReason:      log.FilterReason,
		ErrorMessage:      log.ErrorMessage,
		RequestHeaders:    log.RequestHeaders,
		RequestBody:       log.RequestBody,
		ResponseHeaders:   log.ResponseHeaders,
		ResponseBody:      log.ResponseBody,
		ResponseSizeBytes: log.ResponseSizeBytes,
	}

	// Resolve client_id to client_name
	if log.ClientID != nil {
		clientIDStr := log.ClientID.String()
		response.ClientID = &clientIDStr
		client, clientErr := h.clientRepo.GetByID(ctx, *log.ClientID)
		if clientErr == nil && client != nil {
			response.ClientName = &client.Name
		}
	}

	h.respondJSON(w, http.StatusOK, response)
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

// GetSettings returns all system settings
func (h *AdminHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.systemSettingsRepo.GetAll(r.Context())
	if err != nil {
		h.logger.Errorf(err, "Failed to get system settings")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve settings")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"settings": settings,
	})
}

// UpdateSetting updates a single system setting
func (h *AdminHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Key == "" {
		h.respondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	if err := h.systemSettingsRepo.Set(r.Context(), req.Key, req.Value); err != nil {
		h.logger.Errorf(err, "Failed to update setting: %s", req.Key)
		h.respondError(w, http.StatusInternalServerError, "Failed to update setting")
		return
	}

	h.logger.Infof("Setting updated: %s = %s", req.Key, req.Value)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Setting updated successfully",
		"key":     req.Key,
		"value":   req.Value,
	})
}
