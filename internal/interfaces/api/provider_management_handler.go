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
	providerMgr          *providers.ProviderManager
	config               *config.Config
	logger               *logger.Logger
}

// NewProviderManagementHandler creates a new provider management handler
func NewProviderManagementHandler(
	providerSettingsRepo *repositories.ProviderSettingsRepository,
	providerMgr *providers.ProviderManager,
	cfg *config.Config,
	log *logger.Logger,
) *ProviderManagementHandler {
	return &ProviderManagementHandler{
		providerSettingsRepo: providerSettingsRepo,
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
