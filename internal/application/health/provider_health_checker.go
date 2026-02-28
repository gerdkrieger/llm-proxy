// Package health provides health check services for providers.
package health

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database/repositories"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// ProviderHealthChecker periodically checks the health of custom providers
type ProviderHealthChecker struct {
	providerConfigRepo *repositories.ProviderConfigRepository
	providerAPIKeyRepo *repositories.ProviderAPIKeyRepository
	logger             *logger.Logger
	checkInterval      time.Duration
	stopCh             chan struct{}
}

// NewProviderHealthChecker creates a new provider health checker
func NewProviderHealthChecker(
	providerConfigRepo *repositories.ProviderConfigRepository,
	providerAPIKeyRepo *repositories.ProviderAPIKeyRepository,
	checkInterval time.Duration,
	log *logger.Logger,
) *ProviderHealthChecker {
	return &ProviderHealthChecker{
		providerConfigRepo: providerConfigRepo,
		providerAPIKeyRepo: providerAPIKeyRepo,
		logger:             log,
		checkInterval:      checkInterval,
		stopCh:             make(chan struct{}),
	}
}

// Start begins the periodic health check routine
func (h *ProviderHealthChecker) Start() {
	h.logger.Infof("Starting provider health checker (interval: %v)", h.checkInterval)

	// Run initial check after 30 seconds
	time.AfterFunc(30*time.Second, func() {
		h.runHealthChecks()
	})

	// Start periodic checks
	ticker := time.NewTicker(h.checkInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				h.runHealthChecks()
			case <-h.stopCh:
				ticker.Stop()
				h.logger.Info("Provider health checker stopped")
				return
			}
		}
	}()
}

// Stop stops the health checker
func (h *ProviderHealthChecker) Stop() {
	close(h.stopCh)
}

// runHealthChecks checks all custom providers
func (h *ProviderHealthChecker) runHealthChecks() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	h.logger.Debug("Running provider health checks...")

	// Get all providers from database
	providers, err := h.providerConfigRepo.ListAll(ctx)
	if err != nil {
		h.logger.Errorf(err, "Failed to list providers for health check")
		return
	}

	// Check each custom provider (skip built-in claude/openai)
	for _, provider := range providers {
		// Skip built-in providers (managed by ProviderManager)
		if provider.ProviderType == "claude" || provider.ProviderType == "openai" {
			continue
		}

		h.logger.Debugf("Health checking provider: %s (%s)", provider.ProviderID, provider.ProviderType)

		status := h.checkProvider(ctx, provider)

		// Update database
		if err := h.providerConfigRepo.UpdateHealthStatus(ctx, provider.ProviderID, status); err != nil {
			h.logger.Warnf("Failed to update health status for %s: %v", provider.ProviderID, err)
		} else {
			h.logger.Debugf("Provider %s health status: %s", provider.ProviderID, status)
		}
	}

	h.logger.Debug("Provider health checks completed")
}

// checkProvider performs a health check on a specific provider
func (h *ProviderHealthChecker) checkProvider(ctx context.Context, provider *repositories.ProviderConfig) string {
	// Get API key
	apiKey := h.getProviderAPIKey(ctx, provider)
	if apiKey == "" {
		return "unhealthy" // No API key configured
	}

	// Check based on provider type
	switch provider.ProviderType {
	case "gemini":
		return h.checkGemini(apiKey)
	case "openrouter":
		return h.checkOpenRouter(apiKey)
	case "ollama", "openai-compatible":
		return h.checkOpenAICompatible(provider, apiKey)
	default:
		return "unknown"
	}
}

// getProviderAPIKey retrieves the API key for a provider
func (h *ProviderHealthChecker) getProviderAPIKey(ctx context.Context, provider *repositories.ProviderConfig) string {
	// First check encrypted keys in database
	keys, err := h.providerAPIKeyRepo.GetDecryptedKeysByProvider(ctx, provider.ProviderID)
	if err == nil && len(keys) > 0 {
		// Use first enabled key
		for _, key := range keys {
			return key.APIKey
		}
	}

	// Fallback to config JSONB
	if provider.Config != nil {
		if key, ok := provider.Config["api_key"].(string); ok && key != "" {
			return key
		}
	}

	return ""
}

// checkGemini checks Google Gemini API
func (h *ProviderHealthChecker) checkGemini(apiKey string) string {
	url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + apiKey

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "healthy"
	}

	return "unhealthy"
}

// checkOpenRouter checks OpenRouter API
func (h *ProviderHealthChecker) checkOpenRouter(apiKey string) string {
	url := "https://openrouter.ai/api/v1/models"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "unhealthy"
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "healthy"
	}

	return "unhealthy"
}

// checkOpenAICompatible checks Ollama or OpenAI-compatible endpoints
func (h *ProviderHealthChecker) checkOpenAICompatible(provider *repositories.ProviderConfig, apiKey string) string {
	// Get base URL from config
	baseURL, ok := provider.Config["base_url"].(string)
	if !ok || baseURL == "" {
		return "unhealthy"
	}

	url := baseURL + "/models"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "unhealthy"
	}

	// Add authorization if API key exists (not needed for Ollama)
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "healthy"
	}

	return "unhealthy"
}

// RunHealthCheck manually triggers a health check for a specific provider
// This is called when the user clicks "Test Connection" in the UI
func (h *ProviderHealthChecker) RunHealthCheck(ctx context.Context, providerID string) (string, error) {
	provider, err := h.providerConfigRepo.GetByProviderID(ctx, providerID)
	if err != nil {
		return "", fmt.Errorf("provider not found: %w", err)
	}

	status := h.checkProvider(ctx, provider)

	// Update database
	if err := h.providerConfigRepo.UpdateHealthStatus(ctx, providerID, status); err != nil {
		return status, fmt.Errorf("failed to update health status: %w", err)
	}

	return status, nil
}
