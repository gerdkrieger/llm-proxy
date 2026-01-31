// Package api provides provider management HTTP handlers.
package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

// Available models for each provider (hardcoded for MVP)
var providerModels = map[string][]ModelInfo{
	"claude": {
		{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet (Latest)", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku (Latest)", Capabilities: []string{"vision"}},
		{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-3-sonnet-20240229", Name: "Claude 3 Sonnet", Capabilities: []string{"vision", "function_calling"}},
		{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", Capabilities: []string{"vision"}},
		{ID: "claude-2.1", Name: "Claude 2.1", Capabilities: []string{}},
		{ID: "claude-2.0", Name: "Claude 2.0", Capabilities: []string{}},
		{ID: "claude-instant-1.2", Name: "Claude Instant 1.2", Capabilities: []string{}},
	},
	"openai": {
		{ID: "gpt-4-turbo", Name: "GPT-4 Turbo", Capabilities: []string{"vision", "function_calling", "json_mode"}},
		{ID: "gpt-4-turbo-preview", Name: "GPT-4 Turbo Preview", Capabilities: []string{"vision", "function_calling"}},
		{ID: "gpt-4-1106-preview", Name: "GPT-4 Turbo 1106", Capabilities: []string{"vision", "function_calling"}},
		{ID: "gpt-4-vision-preview", Name: "GPT-4 Vision Preview", Capabilities: []string{"vision"}},
		{ID: "gpt-4", Name: "GPT-4", Capabilities: []string{"function_calling"}},
		{ID: "gpt-4-0613", Name: "GPT-4 (0613)", Capabilities: []string{"function_calling"}},
		{ID: "gpt-4-32k", Name: "GPT-4 32K", Capabilities: []string{"function_calling"}},
		{ID: "gpt-4-32k-0613", Name: "GPT-4 32K (0613)", Capabilities: []string{"function_calling"}},
		{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo", Capabilities: []string{"function_calling"}},
		{ID: "gpt-3.5-turbo-16k", Name: "GPT-3.5 Turbo 16K", Capabilities: []string{"function_calling"}},
		{ID: "gpt-3.5-turbo-1106", Name: "GPT-3.5 Turbo 1106", Capabilities: []string{"function_calling", "json_mode"}},
		{ID: "gpt-3.5-turbo-0613", Name: "GPT-3.5 Turbo (0613)", Capabilities: []string{"function_calling"}},
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

	// Get available models for this provider
	availableModels, ok := providerModels[providerID]
	if !ok {
		h.respondError(w, http.StatusNotFound, "Provider not found")
		return
	}

	// Create a map of enabled model IDs
	enabledSet := make(map[string]bool)
	for _, modelID := range req.EnabledModels {
		enabledSet[modelID] = true
	}

	// Update all models in the database
	ctx := r.Context()
	updatedCount := 0
	for _, model := range availableModels {
		enabled := enabledSet[model.ID]

		// Create or update model in DB
		dbModel := &repositories.ProviderModel{
			ID:         uuid.New(),
			ProviderID: providerID,
			ModelID:    model.ID,
			ModelName:  model.Name,
			Enabled:    enabled,
			Capabilities: map[string]interface{}{
				"features": model.Capabilities,
			},
			Pricing: map[string]interface{}{},
		}

		if err := h.providerModelRepo.Create(ctx, dbModel); err != nil {
			h.logger.Warnf("Failed to update model %s: %v", model.ID, err)
			continue
		}
		updatedCount++
	}

	h.logger.Infof("Updated %d models for provider %s", updatedCount, providerID)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"provider_id":   providerID,
		"updated_count": updatedCount,
		"enabled_count": len(req.EnabledModels),
	})
}
