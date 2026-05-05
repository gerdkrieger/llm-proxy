// Package api provides provider management HTTP handlers.
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/application/providers"
	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	providersInfra "github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// maskAPIKey masks an API key for safe logging.
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// ProviderManagementHandler handles provider management requests
type ProviderManagementHandler struct {
	providerSettingsRepo *repositories.ProviderSettingsRepository
	providerConfigRepo   *repositories.ProviderConfigRepository
	providerModelRepo    *repositories.ProviderModelRepository
	providerAPIKeyRepo   *repositories.ProviderAPIKeyRepository // nil if encryption not configured
	providerMgr          *providersInfra.ProviderManager
	config               *config.Config
	logger               *logger.Logger
}

// NewProviderManagementHandler creates a new provider management handler
func NewProviderManagementHandler(
	providerSettingsRepo *repositories.ProviderSettingsRepository,
	providerConfigRepo *repositories.ProviderConfigRepository,
	providerModelRepo *repositories.ProviderModelRepository,
	providerAPIKeyRepo *repositories.ProviderAPIKeyRepository,
	providerMgr *providersInfra.ProviderManager,
	cfg *config.Config,
	log *logger.Logger,
) *ProviderManagementHandler {
	return &ProviderManagementHandler{
		providerSettingsRepo: providerSettingsRepo,
		providerConfigRepo:   providerConfigRepo,
		providerModelRepo:    providerModelRepo,
		providerAPIKeyRepo:   providerAPIKeyRepo,
		providerMgr:          providerMgr,
		config:               cfg,
		logger:               log,
	}
}

// GetProviderConfig returns the configuration for a specific provider
// GET /admin/providers/{id}/config
func (h *ProviderManagementHandler) GetProviderConfig(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	ctx := r.Context()
	h.logger.Infof("Admin: Getting config for provider: %s", providerID)

	var config map[string]interface{}

	// First check if it's a built-in provider
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
		// Check if it's a custom provider in the database
		providerConfig, err := h.providerConfigRepo.GetByProviderID(ctx, providerID)
		if err != nil {
			h.logger.Warnf("Provider %s not found in database: %v", providerID, err)
			h.respondError(w, http.StatusNotFound, "provider not found")
			return
		}

		// Count API keys for this provider
		var apiKeys []*repositories.ProviderAPIKey
		if h.providerAPIKeyRepo != nil {
			apiKeys, _ = h.providerAPIKeyRepo.ListByProvider(ctx, providerID)
		}

		// Get enabled models count
		models, _ := h.providerModelRepo.GetByProvider(ctx, providerID)
		enabledModels := []string{}
		for _, model := range models {
			if model.Enabled {
				enabledModels = append(enabledModels, model.ModelID)
			}
		}

		config = map[string]interface{}{
			"provider_id":   providerConfig.ProviderID,
			"provider_name": providerConfig.ProviderName,
			"provider_type": providerConfig.ProviderType,
			"enabled":       providerConfig.Enabled,
			"api_keys":      len(apiKeys),
			"models":        enabledModels,
			"config":        providerConfig.Config,
		}
	}

	h.respondJSON(w, http.StatusOK, config)
}

// TestProvider tests the connection to a specific provider
// POST /admin/providers/{id}/test
func (h *ProviderManagementHandler) TestProvider(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	ctx := r.Context()

	h.logger.Infof("Admin: Testing connection for provider: %s", providerID)

	// Get provider config from database
	providerConfig, err := h.providerConfigRepo.GetByProviderID(ctx, providerID)
	if err != nil {
		h.logger.Errorf(err, "Failed to get provider config for %s", providerID)
		h.respondError(w, http.StatusNotFound, "provider not found")
		return
	}

	// Test based on provider type
	var testResult map[string]interface{}

	switch providerConfig.ProviderType {
	case "claude", "openai":
		// Test built-in providers via ProviderManager
		testResult = h.testBuiltInProvider(ctx, providerID)
	case "gemini":
		testResult = h.testGeminiProvider(providerConfig)
	case "openrouter":
		testResult = h.testOpenRouterProvider(providerConfig)
	case "ollama", "openai-compatible":
		testResult = h.testOpenAICompatibleProvider(providerConfig)
	default:
		testResult = map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("unsupported provider type: %s", providerConfig.ProviderType),
			"models": []string{},
		}
	}

	statusCode := http.StatusOK
	if testResult["status"] == "failed" {
		statusCode = http.StatusServiceUnavailable
	}

	h.respondJSON(w, statusCode, testResult)
}

// testBuiltInProvider tests Claude/OpenAI via ProviderManager
func (h *ProviderManagementHandler) testBuiltInProvider(ctx context.Context, providerID string) map[string]interface{} {
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

	// Get models from ProviderManager
	models := h.providerMgr.GetAvailableModels()
	providerModels := []string{}

	for _, model := range models {
		if providerID == "claude" && len(model) >= 6 && model[:6] == "claude" {
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

	return response
}

// testGeminiProvider tests Google Gemini API
func (h *ProviderManagementHandler) testGeminiProvider(config *repositories.ProviderConfig) map[string]interface{} {
	ctx := context.Background()

	// Try to get API key from provider_api_keys table first
	apiKey := ""
	if h.providerAPIKeyRepo != nil {
		decryptedKeys, err := h.providerAPIKeyRepo.GetDecryptedKeysByProvider(ctx, config.ProviderID)
		if err == nil && len(decryptedKeys) > 0 {
			// Use the first enabled key
			for _, key := range decryptedKeys {
				apiKey = key.APIKey
				h.logger.Debugf("Using API key from database for %s", config.ProviderID)
				break
			}
		}
	}

	// Fallback to config if no key in database
	if apiKey == "" {
		if key, ok := config.Config["api_key"].(string); ok && key != "" {
			apiKey = key
			h.logger.Debugf("Using API key from config for %s", config.ProviderID)
		}
	}

	if apiKey == "" {
		return map[string]interface{}{
			"status": "failed",
			"error":  "API key not configured. Please add an API key via 'Manage API Keys'.",
			"models": []string{},
		}
	}

	// Test Gemini API by listing models
	url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + apiKey

	h.logger.Infof("Testing Gemini API with key %s...", maskAPIKey(apiKey))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		h.logger.Errorf(err, "Gemini API request failed")
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Connection failed: %v", err),
			"models": []string{},
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		// Only log body on non-success to avoid verbose logs
		h.logger.Warnf("Gemini API non-success status: %d, body: %s", resp.StatusCode, string(body))
	} else {
		h.logger.Infof("Gemini API response status: %d", resp.StatusCode)
	}

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"status":        "failed",
			"error":         fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)),
			"models":        []string{},
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		}
	}

	// Parse models from response
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Errorf(err, "Failed to parse Gemini models response")
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Failed to parse response: %v", err),
			"models": []string{},
		}
	}

	models := make([]string, 0)
	for _, model := range result.Models {
		// Extract model ID from name (e.g. "models/gemini-pro" -> "gemini-pro")
		if len(model.Name) > 7 && model.Name[:7] == "models/" {
			models = append(models, model.Name[7:])
		} else {
			models = append(models, model.Name)
		}
	}

	return map[string]interface{}{
		"status": "success",
		"models": models,
		"count":  len(models),
	}
}

// testOpenRouterProvider tests OpenRouter API
func (h *ProviderManagementHandler) testOpenRouterProvider(config *repositories.ProviderConfig) map[string]interface{} {
	ctx := context.Background()

	// Try to get API key from provider_api_keys table first
	apiKey := ""
	if h.providerAPIKeyRepo != nil {
		decryptedKeys, err := h.providerAPIKeyRepo.GetDecryptedKeysByProvider(ctx, config.ProviderID)
		if err == nil && len(decryptedKeys) > 0 {
			for _, key := range decryptedKeys {
				apiKey = key.APIKey
				h.logger.Debugf("Using API key from database for %s", config.ProviderID)
				break
			}
		}
	}

	// Fallback to config
	if apiKey == "" {
		if key, ok := config.Config["api_key"].(string); ok && key != "" {
			apiKey = key
			h.logger.Debugf("Using API key from config for %s", config.ProviderID)
		}
	}

	if apiKey == "" {
		return map[string]interface{}{
			"status": "failed",
			"error":  "API key not configured. Please add an API key via 'Manage API Keys'.",
			"models": []string{},
		}
	}

	// Test OpenRouter API by listing models
	url := "https://openrouter.ai/api/v1/models"

	h.logger.Infof("Testing OpenRouter API with key %s...", maskAPIKey(apiKey))

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Failed to create request: %v", err),
			"models": []string{},
		}
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		h.logger.Errorf(err, "OpenRouter API request failed")
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Connection failed: %v", err),
			"models": []string{},
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	h.logger.Infof("OpenRouter API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"status":        "failed",
			"error":         fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)),
			"models":        []string{},
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		}
	}

	// Parse models from response
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Errorf(err, "Failed to parse OpenRouter models response")
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Failed to parse response: %v", err),
			"models": []string{},
		}
	}

	models := make([]string, len(result.Data))
	for i, model := range result.Data {
		models[i] = model.ID
	}

	return map[string]interface{}{
		"status": "success",
		"models": models[:min(10, len(models))], // Show first 10 models
		"count":  len(models),
	}
}

// testOpenAICompatibleProvider tests Ollama or other OpenAI-compatible APIs
func (h *ProviderManagementHandler) testOpenAICompatibleProvider(config *repositories.ProviderConfig) map[string]interface{} {
	baseURL, ok := config.Config["base_url"].(string)
	if !ok || baseURL == "" {
		return map[string]interface{}{
			"status": "failed",
			"error":  "Base URL not configured",
			"models": []string{},
		}
	}

	// Test by calling /v1/models endpoint
	url := baseURL + "/models"

	h.logger.Infof("Testing OpenAI-compatible API: %s", url)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Failed to create request: %v", err),
			"models": []string{},
		}
	}

	// Add API key if configured
	if apiKey, ok := config.Config["api_key"].(string); ok && apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		h.logger.Errorf(err, "API request failed")
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Connection failed: %v. Make sure the service is running at %s", err, baseURL),
			"models": []string{},
		}
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	h.logger.Infof("API response status: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{
			"status":        "failed",
			"error":         fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)),
			"models":        []string{},
			"status_code":   resp.StatusCode,
			"response_body": string(body),
		}
	}

	// Parse models from OpenAI-compatible response
	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		h.logger.Errorf(err, "Failed to parse models response")
		return map[string]interface{}{
			"status": "failed",
			"error":  fmt.Sprintf("Failed to parse response: %v", err),
			"models": []string{},
		}
	}

	models := make([]string, len(result.Data))
	for i, model := range result.Data {
		models[i] = model.ID
	}

	return map[string]interface{}{
		"status": "success",
		"models": models,
		"count":  len(models),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// SyncProviderModels triggers a manual model synchronization to database
// POST /admin/providers/sync-models
func (h *ProviderManagementHandler) SyncProviderModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Info("Admin: Triggering manual model sync to database...")

	// Import the model sync service to get all known models
	allModels := providers.GetAllKnownModels()
	syncedCount := 0
	errorCount := 0
	skippedCount := 0

	for _, model := range allModels {
		// Check if model already exists
		existing, err := h.providerModelRepo.GetByProviderAndModel(ctx, model.ProviderID, model.ID)

		if err != nil || existing == nil {
			// Model doesn't exist, create it as enabled by default
			desc := model.Description
			dbModel := &repositories.ProviderModel{
				ID:          uuid.New(),
				ProviderID:  model.ProviderID,
				ModelID:     model.ID,
				ModelName:   model.Name,
				Enabled:     true, // Default: enabled
				Description: &desc,
				Capabilities: map[string]interface{}{
					"features": model.Capabilities,
				},
				Pricing: map[string]interface{}{},
			}

			if err := h.providerModelRepo.Create(ctx, dbModel); err != nil {
				h.logger.Warnf("Failed to sync model %s: %v", model.ID, err)
				errorCount++
				continue
			}

			h.logger.Debugf("Synced new model: %s (%s)", model.Name, model.ID)
			syncedCount++
		} else {
			h.logger.Debugf("Model already exists: %s (skipped)", model.ID)
			skippedCount++
		}
	}

	h.logger.Infof("Model sync completed: %d new models synced, %d skipped (already exist), %d errors, %d total models",
		syncedCount, skippedCount, errorCount, len(allModels))

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":       true,
		"message":       "Model synchronization completed",
		"synced_count":  syncedCount,
		"skipped_count": skippedCount,
		"error_count":   errorCount,
		"total_models":  len(allModels),
	})
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

	// Get models from database first
	dbModels, err := h.providerModelRepo.GetByProvider(r.Context(), providerID)
	if err != nil {
		h.logger.Errorf(err, "Failed to get models from DB for provider %s", providerID)
		h.respondError(w, http.StatusInternalServerError, "Failed to retrieve models")
		return
	}

	// If no models in DB, check if we have hardcoded models for built-in providers
	if len(dbModels) == 0 {
		availableModels, ok := providerModels[providerID]
		if !ok {
			h.respondError(w, http.StatusNotFound, "No models found for provider. Import models first.")
			return
		}

		// Use hardcoded models with default enabled status
		modelsWithStatus := make([]ModelWithStatus, len(availableModels))
		for i, model := range availableModels {
			modelsWithStatus[i] = ModelWithStatus{
				ModelInfo: model,
				Enabled:   true, // Default to enabled
			}
		}

		h.respondJSON(w, http.StatusOK, map[string]interface{}{
			"provider_id": providerID,
			"models":      modelsWithStatus,
			"total":       len(modelsWithStatus),
		})
		return
	}

	// Convert DB models to ModelWithStatus format
	modelsWithStatus := make([]ModelWithStatus, len(dbModels))
	for i, dbModel := range dbModels {
		capabilities := []string{}
		if caps, ok := dbModel.Capabilities["features"].([]interface{}); ok {
			for _, cap := range caps {
				if capStr, ok := cap.(string); ok {
					capabilities = append(capabilities, capStr)
				}
			}
		}

		modelsWithStatus[i] = ModelWithStatus{
			ModelInfo: ModelInfo{
				ID:           dbModel.ModelID,
				Name:         dbModel.ModelName,
				Capabilities: capabilities,
			},
			Enabled: dbModel.Enabled,
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

// ============================================================================
// API KEY MANAGEMENT
// ============================================================================

// ListProviderAPIKeys returns all API keys for a provider (hints only, not decrypted).
// GET /admin/providers/{id}/keys
func (h *ProviderManagementHandler) ListProviderAPIKeys(w http.ResponseWriter, r *http.Request) {
	if h.providerAPIKeyRepo == nil {
		h.respondError(w, http.StatusNotImplemented, "encryption not configured — set ENCRYPTION_KEY env var")
		return
	}

	providerID := chi.URLParam(r, "id")
	h.logger.Infof("Admin: Listing API keys for provider: %s", providerID)

	keys, err := h.providerAPIKeyRepo.ListByProvider(r.Context(), providerID)
	if err != nil {
		h.logger.Errorf(err, "Failed to list API keys for provider %s", providerID)
		h.respondError(w, http.StatusInternalServerError, "failed to list API keys")
		return
	}

	// Count config.yaml keys for this provider (fallback info)
	var configKeyCount int
	switch providerID {
	case "claude":
		configKeyCount = len(h.config.Providers.Claude.APIKeys)
	case "openai":
		configKeyCount = len(h.config.Providers.OpenAI.APIKeys)
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"provider_id":      providerID,
		"keys":             keys,
		"total":            len(keys),
		"config_key_count": configKeyCount,
	})
}

// AddProviderAPIKeyRequest is the request body for adding a provider API key.
type AddProviderAPIKeyRequest struct {
	KeyName string `json:"key_name"`
	APIKey  string `json:"api_key"`
	Weight  int    `json:"weight"`
	MaxRPM  int    `json:"max_rpm"`
}

// AddProviderAPIKey stores a new encrypted API key for a provider.
// POST /admin/providers/{id}/keys
func (h *ProviderManagementHandler) AddProviderAPIKey(w http.ResponseWriter, r *http.Request) {
	if h.providerAPIKeyRepo == nil {
		h.respondError(w, http.StatusNotImplemented, "encryption not configured — set ENCRYPTION_KEY env var")
		return
	}

	providerID := chi.URLParam(r, "id")

	var req AddProviderAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate
	if req.KeyName == "" {
		h.respondError(w, http.StatusBadRequest, "key_name is required")
		return
	}
	if req.APIKey == "" {
		h.respondError(w, http.StatusBadRequest, "api_key is required")
		return
	}
	if req.Weight <= 0 {
		req.Weight = 1
	}
	if req.MaxRPM <= 0 {
		req.MaxRPM = 60
	}

	h.logger.Infof("Admin: Adding API key '%s' for provider: %s", req.KeyName, providerID)

	key, err := h.providerAPIKeyRepo.Create(r.Context(), providerID, req.KeyName, req.APIKey, req.Weight, req.MaxRPM)
	if err != nil {
		h.logger.Errorf(err, "Failed to add API key for provider %s", providerID)
		h.respondError(w, http.StatusInternalServerError, "failed to store API key")
		return
	}

	// Hot-reload provider keys immediately
	if err := h.providerMgr.ReloadKeys(r.Context()); err != nil {
		h.logger.Warnf("Failed to reload provider keys after adding key: %v", err)
		// Don't fail the request - key is stored, just needs manual restart
	}

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"message": "API key stored successfully. The key value will not be shown again.",
		"key":     key,
	})
}

// DeleteProviderAPIKey deletes an API key by ID.
// DELETE /admin/providers/{id}/keys/{key_id}
func (h *ProviderManagementHandler) DeleteProviderAPIKey(w http.ResponseWriter, r *http.Request) {
	if h.providerAPIKeyRepo == nil {
		h.respondError(w, http.StatusNotImplemented, "encryption not configured — set ENCRYPTION_KEY env var")
		return
	}

	providerID := chi.URLParam(r, "id")
	keyIDStr := chi.URLParam(r, "key_id")

	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid key ID format")
		return
	}

	h.logger.Infof("Admin: Deleting API key %s for provider: %s", keyIDStr, providerID)

	if err := h.providerAPIKeyRepo.Delete(r.Context(), keyID); err != nil {
		h.logger.Errorf(err, "Failed to delete API key %s", keyIDStr)
		h.respondError(w, http.StatusInternalServerError, "failed to delete API key")
		return
	}

	// Hot-reload provider keys immediately
	if err := h.providerMgr.ReloadKeys(r.Context()); err != nil {
		h.logger.Warnf("Failed to reload provider keys after deleting key: %v", err)
		// Don't fail the request - key is deleted, just needs manual restart
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "API key deleted successfully",
	})
}

// ToggleProviderAPIKey enables or disables an API key.
// PUT /admin/providers/{id}/keys/{key_id}/toggle
func (h *ProviderManagementHandler) ToggleProviderAPIKey(w http.ResponseWriter, r *http.Request) {
	if h.providerAPIKeyRepo == nil {
		h.respondError(w, http.StatusNotImplemented, "encryption not configured — set ENCRYPTION_KEY env var")
		return
	}

	providerID := chi.URLParam(r, "id")
	keyIDStr := chi.URLParam(r, "key_id")

	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid key ID format")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.Infof("Admin: Toggling API key %s for provider %s to enabled=%v", keyIDStr, providerID, req.Enabled)

	if err := h.providerAPIKeyRepo.SetEnabled(r.Context(), keyID, req.Enabled); err != nil {
		h.logger.Errorf(err, "Failed to toggle API key %s", keyIDStr)
		h.respondError(w, http.StatusInternalServerError, "failed to toggle API key")
		return
	}

	// Hot-reload provider keys immediately
	if err := h.providerMgr.ReloadKeys(r.Context()); err != nil {
		h.logger.Warnf("Failed to reload provider keys after toggling key: %v", err)
		// Don't fail the request - key is toggled, just needs manual restart
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "API key status updated",
		"enabled": req.Enabled,
	})
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

// ImportProviderModels imports models discovered from provider test into database
// POST /admin/providers/{id}/models/import
func (h *ProviderManagementHandler) ImportProviderModels(w http.ResponseWriter, r *http.Request) {
	providerID := chi.URLParam(r, "id")
	ctx := r.Context()

	h.logger.Infof("Admin: Importing models for provider: %s", providerID)

	// Get provider config
	providerConfig, err := h.providerConfigRepo.GetByProviderID(ctx, providerID)
	if err != nil {
		h.logger.Errorf(err, "Failed to get provider config")
		h.respondError(w, http.StatusNotFound, "provider not found")
		return
	}

	// Test provider to get available models
	var testResult map[string]interface{}
	switch providerConfig.ProviderType {
	case "gemini":
		testResult = h.testGeminiProvider(providerConfig)
	case "openrouter":
		testResult = h.testOpenRouterProvider(providerConfig)
	case "ollama", "openai-compatible":
		testResult = h.testOpenAICompatibleProvider(providerConfig)
	default:
		h.respondError(w, http.StatusBadRequest, fmt.Sprintf("model import not supported for provider type: %s", providerConfig.ProviderType))
		return
	}

	// Check if test was successful
	if testResult["status"] != "success" {
		h.respondError(w, http.StatusServiceUnavailable, fmt.Sprintf("provider test failed: %v", testResult["error"]))
		return
	}

	// Get models from test result
	modelsInterface, ok := testResult["models"].([]string)
	if !ok {
		h.respondError(w, http.StatusInternalServerError, "failed to get models from test result")
		return
	}

	// Import each model into database
	importedCount := 0
	skippedCount := 0
	errorCount := 0

	for _, modelID := range modelsInterface {
		// Check if model already exists
		existing, err := h.providerModelRepo.GetByProviderAndModel(ctx, providerID, modelID)
		if err == nil && existing != nil {
			h.logger.Debugf("Model %s already exists for provider %s, skipping", modelID, providerID)
			skippedCount++
			continue
		}

		// Create new model entry
		dbModel := &repositories.ProviderModel{
			ID:          uuid.New(),
			ProviderID:  providerID,
			ModelID:     modelID,
			ModelName:   modelID, // Use model ID as name by default
			Enabled:     true,    // Enable by default
			Description: nil,
			Capabilities: map[string]interface{}{
				"features": []string{}, // No capabilities info available yet
			},
			Pricing: map[string]interface{}{},
		}

		if err := h.providerModelRepo.Create(ctx, dbModel); err != nil {
			h.logger.Warnf("Failed to import model %s: %v", modelID, err)
			errorCount++
			continue
		}

		h.logger.Debugf("Imported model: %s for provider %s", modelID, providerID)
		importedCount++
	}

	h.logger.Infof("Model import completed for %s: %d imported, %d skipped, %d errors",
		providerID, importedCount, skippedCount, errorCount)

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"message":        "Models imported successfully",
		"imported_count": importedCount,
		"skipped_count":  skippedCount,
		"error_count":    errorCount,
		"total_models":   len(modelsInterface),
	})
}

// ============================================================================
// PROVIDER CRUD OPERATIONS
// ============================================================================

// CreateProviderRequest represents the request to create a new provider
type CreateProviderRequest struct {
	ProviderID   string                 `json:"provider_id"`
	ProviderName string                 `json:"provider_name"`
	ProviderType string                 `json:"provider_type"` // "ollama", "openai-compatible", "openrouter", "claude", "openai"
	Config       map[string]interface{} `json:"config"`        // Flexible config for different provider types
	Enabled      bool                   `json:"enabled"`
}

// CreateProvider creates a new provider configuration
// POST /admin/providers
func (h *ProviderManagementHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Validate required fields
	if req.ProviderID == "" {
		h.respondError(w, http.StatusBadRequest, "provider_id is required")
		return
	}
	if req.ProviderName == "" {
		h.respondError(w, http.StatusBadRequest, "provider_name is required")
		return
	}
	if req.ProviderType == "" {
		h.respondError(w, http.StatusBadRequest, "provider_type is required")
		return
	}

	// Validate provider type
	validTypes := map[string]bool{
		"ollama":            true,
		"openai-compatible": true,
		"openrouter":        true,
		"gemini":            true,
		"claude":            true,
		"openai":            true,
	}
	if !validTypes[req.ProviderType] {
		h.respondError(w, http.StatusBadRequest, "invalid provider_type. Must be one of: ollama, openai-compatible, openrouter, gemini, claude, openai")
		return
	}

	// Validate config based on provider type
	if err := h.validateProviderConfig(req.ProviderType, req.Config); err != nil {
		h.respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid config: %v", err))
		return
	}

	h.logger.Infof("Admin: Creating new provider: %s (%s)", req.ProviderName, req.ProviderType)

	// Extract API key from config if present
	var apiKey string
	if key, ok := req.Config["api_key"].(string); ok && key != "" {
		apiKey = key
		// Remove API key from config (will be stored encrypted separately)
		delete(req.Config, "api_key")
	}

	// Create provider config in database
	providerConfig := &repositories.ProviderConfig{
		ProviderID:   req.ProviderID,
		ProviderName: req.ProviderName,
		ProviderType: req.ProviderType,
		Config:       req.Config,
		Enabled:      req.Enabled,
		HealthStatus: "unknown",
	}

	if err := h.providerConfigRepo.Create(ctx, providerConfig); err != nil {
		h.logger.Errorf(err, "Failed to create provider config")
		h.respondError(w, http.StatusInternalServerError, "failed to create provider")
		return
	}

	// If API key was provided and encryption is enabled, store it encrypted
	if apiKey != "" && h.providerAPIKeyRepo != nil {
		h.logger.Infof("Storing encrypted API key for provider %s", req.ProviderID)

		keyName := req.ProviderName + " API Key"
		_, err := h.providerAPIKeyRepo.Create(ctx, req.ProviderID, keyName, apiKey, 1, 60)
		if err != nil {
			h.logger.Errorf(err, "Failed to store API key for provider %s", req.ProviderID)
			// Don't fail the whole request, provider was created
			h.logger.Warnf("Provider %s created but API key storage failed", req.ProviderID)
		} else {
			h.logger.Infof("API key stored successfully for provider %s", req.ProviderID)
		}
	}

	h.logger.Infof("Provider %s created successfully", req.ProviderID)

	h.respondJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Provider created successfully" + func() string {
			if apiKey != "" && h.providerAPIKeyRepo != nil {
				return " with encrypted API key"
			}
			return ""
		}(),
		"provider": providerConfig,
	})
}

// UpdateProviderRequest represents the request to update a provider
type UpdateProviderRequest struct {
	ProviderName string                 `json:"provider_name"`
	Config       map[string]interface{} `json:"config"`
	Enabled      *bool                  `json:"enabled,omitempty"`
}

// UpdateProvider updates an existing provider configuration
// PUT /admin/providers/{id}
func (h *ProviderManagementHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerID := chi.URLParam(r, "id")

	var req UpdateProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.logger.Infof("Admin: Updating provider: %s", providerID)

	// Get existing provider
	existing, err := h.providerConfigRepo.GetByProviderID(ctx, providerID)
	if err != nil {
		h.logger.Errorf(err, "Failed to get provider %s", providerID)
		h.respondError(w, http.StatusNotFound, "provider not found")
		return
	}

	// Update fields
	if req.ProviderName != "" {
		existing.ProviderName = req.ProviderName
	}
	if req.Config != nil {
		// Validate config
		if err := h.validateProviderConfig(existing.ProviderType, req.Config); err != nil {
			h.respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid config: %v", err))
			return
		}
		existing.Config = req.Config
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	// Update in database
	if err := h.providerConfigRepo.Update(ctx, existing); err != nil {
		h.logger.Errorf(err, "Failed to update provider")
		h.respondError(w, http.StatusInternalServerError, "failed to update provider")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"message":  "Provider updated successfully",
		"provider": existing,
	})
}

// DeleteProvider deletes a provider configuration
// DELETE /admin/providers/{id}
func (h *ProviderManagementHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	providerID := chi.URLParam(r, "id")

	h.logger.Infof("Admin: Deleting provider: %s", providerID)

	// Prevent deletion of built-in providers
	if providerID == "claude" || providerID == "openai" {
		h.respondError(w, http.StatusForbidden, "cannot delete built-in provider")
		return
	}

	// Delete from database
	if err := h.providerConfigRepo.Delete(ctx, providerID); err != nil {
		h.logger.Errorf(err, "Failed to delete provider")
		h.respondError(w, http.StatusInternalServerError, "failed to delete provider")
		return
	}

	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Provider deleted successfully",
	})
}

// validateProviderConfig validates provider configuration based on type
func (h *ProviderManagementHandler) validateProviderConfig(providerType string, config map[string]interface{}) error {
	switch providerType {
	case "ollama", "openai-compatible":
		// Require base_url
		if _, ok := config["base_url"]; !ok {
			return fmt.Errorf("base_url is required for %s", providerType)
		}
		// api_key is optional for ollama
	case "openrouter":
		// Require api_key
		if _, ok := config["api_key"]; !ok {
			return fmt.Errorf("api_key is required for openrouter")
		}
	case "gemini":
		// Require api_key for Google Gemini
		if _, ok := config["api_key"]; !ok {
			return fmt.Errorf("api_key is required for gemini")
		}
		// Optional: project_id for enterprise usage
	case "claude", "openai":
		// api_key handled separately via provider_api_keys table
		// No specific validation here
	default:
		return fmt.Errorf("unknown provider type: %s", providerType)
	}
	return nil
}
