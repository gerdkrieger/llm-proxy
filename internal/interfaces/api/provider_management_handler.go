// Package api provides provider management HTTP handlers.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ProviderManagementHandler handles provider management requests
type ProviderManagementHandler struct {
	providerSettingsRepo *repositories.ProviderSettingsRepository
	providerModelRepo    *repositories.ProviderModelRepository
	providerMgr          *providers.ProviderManager
	config               *config.Config
	logger               *logger.Logger
}

// NewProviderManagementHandler creates a new provider management handler
func NewProviderManagementHandler(
	providerSettingsRepo *repositories.ProviderSettingsRepository,
	providerModelRepo *repositories.ProviderModelRepository,
	providerMgr *providers.ProviderManager,
	cfg *config.Config,
	log *logger.Logger,
) *ProviderManagementHandler {
	return &ProviderManagementHandler{
		providerSettingsRepo: providerSettingsRepo,
		providerModelRepo:    providerModelRepo,
		providerMgr:          providerMgr,
		config:               cfg,
		logger:               log,
	}
}

// GetProviderConfig returns the configuration for a specific provider
// GET /admin/providers/{id}/config
func (h *ProviderManagementHandler) GetProviderConfig(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	h.logger.Infof("Admin: Getting config for provider: %s", providerID)

	var config map[string]interface{}

	switch providerID {
	case "claude":
		config = map[string]interface{}{
			"provider_id":   "claude",
			"provider_name": "Anthropic Claude",
			"provider_type": "claude",
			"enabled":       h.config.Providers.Claude.Enabled,
			"api_keys":      len(h.config.Providers.Claude.APIKeys),
			"models":        h.config.Providers.Claude.Models,
		}

	case "openai":
		config = map[string]interface{}{
			"provider_id":   "openai",
			"provider_name": "OpenAI",
			"provider_type": "openai",
			"enabled":       h.config.Providers.OpenAI.Enabled,
			"api_keys":      len(h.config.Providers.OpenAI.APIKeys),
			"models":        h.config.Providers.OpenAI.Models,
		}

	default:
		h.respondError(w, http.StatusNotFound, "provider not found")
		return
	}

	h.respondJSON(w, http.StatusOK, config)
}

// TestProvider tests the connection to a specific provider
// POST /admin/providers/{id}/test
func (h *ProviderManagementHandler) TestProvider(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	ctx := r.Context()

	h.logger.Infof("Admin: Testing connection for provider: %s", providerID)

	// Test provider health
	err := h.providerMgr.Health(ctx)

	var status string
	var errorMsg *string
	if err != nil {
		status = "failed"
		errStr := err.Error()
		errorMsg = &errStr
	} else {
		status = "success"
	}

	// Get models to verify connection
	models := h.providerMgr.GetAvailableModels()
	providerModels := []string{}

	for _, model := range models {
		if providerID == "claude" && (model[:6] == "claude" || len(model) > 6 && model[:6] == "claude") {
			providerModels = append(providerModels, model)
		} else if providerID == "openai" && len(model) >= 3 && model[:3] == "gpt" {
			providerModels = append(providerModels, model)
		}
	}

	response := map[string]interface{}{
		"status": status,
		"models": providerModels,
	}

	if errorMsg != nil {
		response["error"] = *errorMsg
	}

	statusCode := http.StatusOK
	if status == "failed" {
		statusCode = http.StatusServiceUnavailable
	}

	h.respondJSON(w, statusCode, response)
}

// ToggleProvider enables or disables a provider
// PUT /admin/providers/{id}/toggle
func (h *ProviderManagementHandler) ToggleProvider(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	ctx := r.Context()

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.Infof("Admin: Toggling provider %s to enabled=%v", providerID, req.Enabled)

	// Try to get existing setting
	setting, err := h.providerSettingsRepo.GetByProviderID(ctx, providerID)
	if err != nil {
		// Create new setting
		setting = &repositories.ProviderSetting{
			ProviderID:   providerID,
			ProviderName: getProviderName(providerID),
			ProviderType: providerID,
			Enabled:      req.Enabled,
			Config:       make(map[string]interface{}),
		}

		if err := h.providerSettingsRepo.Upsert(ctx, setting); err != nil {
			h.logger.Errorf(err, "Failed to create provider setting")
			h.respondError(w, http.StatusInternalServerError, "failed to update provider status")
			return
		}
	} else {
		// Update existing
		if err := h.providerSettingsRepo.SetEnabled(ctx, providerID, req.Enabled); err != nil {
			h.logger.Errorf(err, "Failed to update provider enabled status")
			h.respondError(w, http.StatusInternalServerError, "failed to update provider status")
			return
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"provider_id": providerID,
		"enabled":     req.Enabled,
		"message":     "Provider status updated successfully",
	})
}

// Helper functions
func (h *ProviderManagementHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *ProviderManagementHandler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "error",
		"message": message,
	})
}

func getProviderName(providerID string) string {
	switch providerID {
	case "claude":
		return "Anthropic Claude"
	case "openai":
		return "OpenAI"
	default:
		return providerID
	}
}

// ============================================================================
// MODEL MANAGEMENT
// ============================================================================

// Available models for each provider (synced with model_sync_service.go)
var providerModels = map[string][]ModelInfo{
	"claude": {
		// Claude 4.5 Series (LATEST - Jan 2026)
		{ID: "claude-sonnet-4-5", Name: "Claude Sonnet 4.5 (Alias)", Capabilities: []string{"vision", "function_calling", "extended_thinking", "200k_context", "1m_context_beta"}},
		{ID: "claude-sonnet-4-5-20250929", Name: "Claude Sonnet 4.5 (Sep 2025)", Capabilities: []string{"vision", "function_calling", "extended_thinking"}},
		{ID: "claude-haiku-4-5", Name: "Claude Haiku 4.5 (Alias)", Capabilities: []string{"vision", "function_calling", "extended_thinking"}},
		{ID: "claude-haiku-4-5-20251001", Name: "Claude Haiku 4.5 (Oct 2025)", Capabilities: []string{"vision", "function_calling", "extended_thinking"}},
		{ID: "claude-opus-4-5", Name: "Claude Opus 4.5 (Alias)", Capabilities: []string{"vision", "function_calling", "extended_thinking"}},
		{ID: "claude-opus-4-5-20251101", Name: "Claude Opus 4.5 (Nov 2025)", Capabilities: []string{"vision", "function_calling", "extended_thinking"}},
		// Claude 4.1 and 4 Series (Legacy but available)
		{ID: "claude-opus-4-1", Name: "Claude Opus 4.1 (Alias)", Capabilities: []string{"vision", "function_calling", "extended_thinking"}},
		{ID: "claude-opus-4-1-20250805", Name: "Claude Opus 4.1 (Aug 2025)", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-sonnet-4-0", Name: "Claude Sonnet 4 (Alias)", Capabilities: []string{"vision", "function_calling", "extended_thinking"}},
		{ID: "claude-sonnet-4-20250514", Name: "Claude Sonnet 4 (May 2025)", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-3-7-sonnet-latest", Name: "Claude 3.7 Sonnet (Alias)", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet (Feb 2025)", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-opus-4-0", Name: "Claude Opus 4 (Alias)", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-opus-4-20250514", Name: "Claude Opus 4 (May 2025)", Capabilities: []string{"vision", "function_calling"}},
		// Claude 3 Series (Legacy)
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", Capabilities: []string{"vision"}},
	},
	"openai": {
		// GPT-5 Series (LATEST - Jan 2026)
		{ID: "gpt-5.2", Name: "GPT-5.2", Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning", "coding"}},
		{ID: "gpt-5.2-pro", Name: "GPT-5.2 Pro", Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning"}},
		{ID: "gpt-5.2-codex", Name: "GPT-5.2 Codex", Capabilities: []string{"coding", "reasoning", "agentic"}},
		{ID: "gpt-5.1", Name: "GPT-5.1", Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning"}},
		{ID: "gpt-5.1-codex", Name: "GPT-5.1 Codex", Capabilities: []string{"coding", "reasoning", "agentic"}},
		{ID: "gpt-5", Name: "GPT-5", Capabilities: []string{"vision", "function_calling", "json_mode", "reasoning"}},
		{ID: "gpt-5-pro", Name: "GPT-5 Pro", Capabilities: []string{"vision", "function_calling", "reasoning"}},
		{ID: "gpt-5-mini", Name: "GPT-5 Mini", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-5-nano", Name: "GPT-5 Nano", Capabilities: []string{"function_calling", "json_mode"}},
		// GPT-4.1 Series (Latest Non-Reasoning)
		{ID: "gpt-4.1", Name: "GPT-4.1", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4.1-mini", Name: "GPT-4.1 Mini", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4.1-nano", Name: "GPT-4.1 Nano", Capabilities: []string{"function_calling", "json_mode"}},
		// o-Series Reasoning Models
		{ID: "o3", Name: "o3", Capabilities: []string{"reasoning"}},
		{ID: "o3-pro", Name: "o3 Pro", Capabilities: []string{"reasoning"}},
		{ID: "o3-mini", Name: "o3 Mini", Capabilities: []string{"reasoning"}},
		{ID: "o4-mini", Name: "o4 Mini", Capabilities: []string{"reasoning"}},
		{ID: "o1", Name: "o1", Capabilities: []string{"reasoning"}},
		{ID: "o1-pro", Name: "o1 Pro", Capabilities: []string{"reasoning"}},
		{ID: "o1-mini", Name: "o1 Mini", Capabilities: []string{"reasoning"}},
		{ID: "o1-preview", Name: "o1 Preview", Capabilities: []string{"reasoning"}},
		// Deep Research
		{ID: "o3-deep-research", Name: "o3 Deep Research", Capabilities: []string{"reasoning", "research"}},
		{ID: "o4-mini-deep-research", Name: "o4 Mini Deep Research", Capabilities: []string{"reasoning", "research"}},
		// GPT-4o Series
		{ID: "gpt-4o", Name: "GPT-4o", Capabilities: []string{"vision", "function_calling", "json_mode", "audio"}},
		{ID: "gpt-4o-2024-11-20", Name: "GPT-4o (Nov 2024)", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4o-2024-08-06", Name: "GPT-4o (Aug 2024)", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4o-2024-05-13", Name: "GPT-4o (May 2024)", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4o-mini", Name: "GPT-4o Mini", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4o-mini-2024-07-18", Name: "GPT-4o Mini (Jul 2024)", Capabilities: []string{"vision", "function_calling"}},
		// Realtime/Audio
		{ID: "gpt-realtime", Name: "GPT Realtime", Capabilities: []string{"realtime", "audio", "text"}},
		{ID: "gpt-realtime-mini", Name: "GPT Realtime Mini", Capabilities: []string{"realtime", "audio"}},
		{ID: "gpt-audio", Name: "GPT Audio", Capabilities: []string{"audio"}},
		{ID: "gpt-audio-mini", Name: "GPT Audio Mini", Capabilities: []string{"audio"}},
		// GPT-4 Turbo (Legacy)
		{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4-turbo-2024-04-09", Name: "GPT-4 Turbo (Apr 2024)", Capabilities: []string{"vision", "function_calling"}},
		{ID: "gpt-4-turbo-preview", Name: "GPT-4 Turbo Preview", Capabilities: []string{"function_calling", "json_mode"}},
		{ID: "gpt-4-vision-preview", Name: "GPT-4 Vision Preview", Capabilities: []string{"vision"}},
		// GPT-4 (Legacy)
		{ID: "gpt-4", Name: "GPT-4", Capabilities: []string{"function_calling"}},
		{ID: "gpt-4-0613", Name: "GPT-4 (Jun 2023)", Capabilities: []string{"function_calling"}},
		{ID: "gpt-4-32k", Name: "GPT-4 32K", Capabilities: []string{"function_calling"}},
		{ID: "gpt-4-32k-0613", Name: "GPT-4 32K (Jun 2023)", Capabilities: []string{"function_calling"}},
		// GPT-3.5 Turbo (Legacy)
		{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Capabilities: []string{"function_calling", "json_mode"}},
		{ID: "gpt-3.5-turbo-0125", Name: "GPT-3.5 Turbo (Jan 2024)", Capabilities: []string{"function_calling"}},
		{ID: "gpt-3.5-turbo-1106", Name: "GPT-3.5 Turbo (Nov 2023)", Capabilities: []string{"function_calling"}},
		{ID: "gpt-3.5-turbo-0613", Name: "GPT-3.5 Turbo (Jun 2023)", Capabilities: []string{"function_calling"}},
	},
}

// ModelInfo represents information about a model
type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Capabilities []string `json:"capabilities"`
}

// ModelWithStatus represents a model with its enable/disable status
type ModelWithStatus struct {
	ModelInfo
	Enabled bool `json:"enabled"`
}

// GetProviderModels returns all available models for a provider with their status
// GET /admin/providers/{id}/models
func (h *ProviderManagementHandler) GetProviderModels(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	h.logger.Infof("Admin: Getting models for provider: %s", providerID)

	// Get available models for this provider
	availableModels, ok := providerModels[providerID]
	if !ok {
		h.respondError(w, http.StatusNotFound, "Provider not found")
		return
	}

	// Get enabled models from database
	dbModels, err := h.providerModelRepo.GetByProvider(r.Context(), providerID)
	if err != nil {
		h.logger.Warnf("Failed to get models from DB: %v, using defaults", err)
	}

	// Create a map of model ID -> enabled status
	enabledMap := make(map[string]bool)
	for _, dbModel := range dbModels {
		enabledMap[dbModel.ModelID] = dbModel.Enabled
	}

	// Combine available models with status
	modelsWithStatus := make([]ModelWithStatus, len(availableModels))
	for i, model := range availableModels {
		enabled, exists := enabledMap[model.ID]
		if !exists {
			enabled = true // Default to enabled if not in DB
		}
		modelsWithStatus[i] = ModelWithStatus{
			ModelInfo: model,
			Enabled:   enabled,
		}
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"provider_id": providerID,
		"models":      modelsWithStatus,
		"total":       len(modelsWithStatus),
	})
}

// ConfigureProviderModelsRequest represents the request to configure models
type ConfigureProviderModelsRequest struct {
	EnabledModels []string `json:"enabled_models"`
}

// ConfigureProviderModels updates which models are enabled for a provider
// POST /admin/providers/{id}/models/configure
func (h *ProviderManagementHandler) ConfigureProviderModels(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	h.logger.Infof("Admin: Configuring models for provider: %s", providerID)

	var req ConfigureProviderModelsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	ctx := r.Context()

	h.logger.Infof("Admin: Received configuration for %d enabled models for provider: %s", len(req.EnabledModels), providerID)

	// Get ALL models for this provider from database (single source of truth)
	allModels, err := h.providerModelRepo.GetByProvider(ctx, providerID)
	if err != nil {
		h.logger.Errorf(err, "Failed to get models from database")
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve models")
		return
	}

	if len(allModels) == 0 {
		h.respondError(w, http.StatusNotFound, "No models found for provider")
		return
	}

	// Create a map of enabled model IDs for fast lookup
	enabledSet := make(map[string]bool)
	for _, modelID := range req.EnabledModels {
		enabledSet[modelID] = true
	}

	// Collect all model IDs that should be enabled and disabled
	var enabledModelIDs []string
	var disabledModelIDs []string

	for _, model := range allModels {
		if enabledSet[model.ModelID] {
			enabledModelIDs = append(enabledModelIDs, model.ModelID)
		} else {
			disabledModelIDs = append(disabledModelIDs, model.ModelID)
		}
	}

	h.logger.Debugf("Enabling %d models, disabling %d models for provider %s",
		len(enabledModelIDs), len(disabledModelIDs), providerID)

	// Use batch updates for better performance
	var updateErrors []error

	// Batch enable selected models
	if len(enabledModelIDs) > 0 {
		if err := h.providerModelRepo.BulkUpdateEnabled(ctx, providerID, enabledModelIDs, true); err != nil {
			h.logger.Errorf(err, "Failed to enable models")
			updateErrors = append(updateErrors, err)
		}
	}

	// Batch disable unselected models
	if len(disabledModelIDs) > 0 {
		if err := h.providerModelRepo.BulkUpdateEnabled(ctx, providerID, disabledModelIDs, false); err != nil {
			h.logger.Errorf(err, "Failed to disable models")
			updateErrors = append(updateErrors, err)
		}
	}

	if len(updateErrors) > 0 {
		h.logger.Warnf("Model configuration completed with %d errors", len(updateErrors))
		h.respondError(w, http.StatusInternalServerError, "Failed to update some models")
		return
	}

	h.logger.Infof("Successfully configured models for provider %s: %d enabled, %d disabled",
		providerID, len(enabledModelIDs), len(disabledModelIDs))

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"provider_id":    providerID,
		"enabled_count":  len(enabledModelIDs),
		"disabled_count": len(disabledModelIDs),
		"total_count":    len(allModels),
	})
}
