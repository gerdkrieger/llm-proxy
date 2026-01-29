// Package api provides models endpoint handler.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ModelsHandler handles model listing requests
type ModelsHandler struct {
	providerManager *providers.ProviderManager
	logger          *logger.Logger
}

// NewModelsHandler creates a new models handler
func NewModelsHandler(providerManager *providers.ProviderManager, log *logger.Logger) *ModelsHandler {
	return &ModelsHandler{
		providerManager: providerManager,
		logger:          log,
	}
}

// ListModels returns available models
// GET /v1/models
func (h *ModelsHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	// Get available models from provider manager
	modelNames := h.providerManager.GetAvailableModels()

	// Convert to OpenAI model format
	modelList := make([]models.OpenAIModel, 0, len(modelNames))
	now := time.Now().Unix()

	for _, modelName := range modelNames {
		modelList = append(modelList, models.OpenAIModel{
			ID:      modelName,
			Object:  "model",
			Created: now,
			OwnedBy: "anthropic",
		})
	}

	response := models.OpenAIModelList{
		Object: "list",
		Data:   modelList,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetModel returns a specific model
// GET /v1/models/{model}
func (h *ModelsHandler) GetModel(w http.ResponseWriter, r *http.Request) {
	// For simplicity, just return the model if it's in our list
	// In a real implementation, you'd validate against available models

	// Extract model ID from URL
	modelID := r.URL.Path[len("/v1/models/"):]

	// Check if model exists
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

	model := models.OpenAIModel{
		ID:      modelID,
		Object:  "model",
		Created: time.Now().Unix(),
		OwnedBy: "anthropic",
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
