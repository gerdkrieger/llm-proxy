// Package api provides models endpoint handler.
package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ModelsHandler handles model listing requests
type ModelsHandler struct {
	providerManager   *providers.ProviderManager
	providerModelRepo *repositories.ProviderModelRepository
	logger            *logger.Logger
}

// NewModelsHandler creates a new models handler
func NewModelsHandler(
	providerManager *providers.ProviderManager,
	providerModelRepo *repositories.ProviderModelRepository,
	log *logger.Logger,
) *ModelsHandler {
	return &ModelsHandler{
		providerManager:   providerManager,
		providerModelRepo: providerModelRepo,
		logger:            log,
	}
}

// ListModels returns available models (filtered by enabled status)
// GET /v1/models
func (h *ModelsHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().Unix()

	// Get all enabled models from database for all providers
	var allEnabledModels []*repositories.ProviderModel

	// Get enabled models for Claude provider
	claudeModels, err := h.providerModelRepo.GetEnabledByProvider(ctx, "claude")
	if err != nil {
		h.logger.Warnf("Error fetching Claude models: %v", err)
	} else {
		allEnabledModels = append(allEnabledModels, claudeModels...)
	}

	// Get enabled models for OpenAI provider
	openaiModels, err := h.providerModelRepo.GetEnabledByProvider(ctx, "openai")
	if err != nil {
		h.logger.Warnf("Error fetching OpenAI models: %v", err)
	} else {
		allEnabledModels = append(allEnabledModels, openaiModels...)
	}

	// Check which providers are healthy/available
	activeProviders := h.providerManager.GetActiveProviderIDs()
	activeProviderMap := make(map[string]bool)
	for _, pid := range activeProviders {
		activeProviderMap[pid] = true
	}

	// Filter to only include models from active providers
	modelList := make([]models.OpenAIModel, 0, len(allEnabledModels))
	totalEnabled := len(allEnabledModels)
	skipped := 0

	for _, dbModel := range allEnabledModels {
		// Skip if provider is not active
		if !activeProviderMap[dbModel.ProviderID] {
			h.logger.Debugf("Skipping model %s - provider %s not active", dbModel.ModelID, dbModel.ProviderID)
			skipped++
			continue
		}

		modelList = append(modelList, models.OpenAIModel{
			ID:      dbModel.ModelID,
			Object:  "model",
			Created: now,
			OwnedBy: dbModel.ProviderID,
		})
	}

	h.logger.Infof("Returning %d models (from %d enabled in DB, %d skipped - provider inactive)",
		len(modelList), totalEnabled, skipped)

	response := models.OpenAIModelList{
		Object: "list",
		Data:   modelList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// getProviderFromModelName determines the provider ID from model name
func (h *ModelsHandler) getProviderFromModelName(modelName string) string {
	lowerName := strings.ToLower(modelName)

	if strings.Contains(lowerName, "claude") {
		return "claude"
	}
	if strings.Contains(lowerName, "gpt") {
		return "openai"
	}

	// Default to anthropic for backward compatibility
	return "anthropic"
}

// GetModel returns a specific model (if enabled)
// GET /v1/models/{model}
func (h *ModelsHandler) GetModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract model ID from URL
	modelID := r.URL.Path[len("/v1/models/"):]

	// Check if model exists in provider manager
	availableModels := h.providerManager.GetAvailableModels()
	found := false
	for _, m := range availableModels {
		if m == modelID {
			found = true
			break
		}
	}

	if !found {
		h.respondError(w, http.StatusNotFound, "model_not_found", "Model not found: "+modelID)
		return
	}

	// Check if model is enabled
	providerID := h.getProviderFromModelName(modelID)
	isEnabled, err := h.providerModelRepo.IsModelEnabled(ctx, providerID, modelID)
	if err != nil {
		h.logger.Warnf("Error checking model status for %s: %v, defaulting to enabled", modelID, err)
		isEnabled = true
	}

	if !isEnabled {
		h.respondError(w, http.StatusNotFound, "model_not_found", "Model is disabled: "+modelID)
		return
	}

	model := models.OpenAIModel{
		ID:      modelID,
		Object:  "model",
		Created: time.Now().Unix(),
		OwnedBy: providerID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model)
}

// respondError sends an error response
func (h *ModelsHandler) respondError(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(models.OpenAIErrorResponse{
		Error: models.OpenAIError{
			Message: message,
			Type:    errorType,
		},
	})
}
