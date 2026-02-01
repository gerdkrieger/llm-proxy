// Package providers manages LLM provider clients with load balancing and failover.
package providers

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers/claude"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers/openai"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// Provider represents an LLM provider client
type Provider interface {
	CreateMessage(ctx context.Context, req *models.ClaudeRequest) (*models.ClaudeResponse, error)
	Health(ctx context.Context) error
	GetAPIKey() string
}

// ProviderManager manages multiple provider instances with load balancing
type ProviderManager struct {
	providers     []Provider
	openaiClients []*openai.Client
	currentIndex  int
	mu            sync.RWMutex
	logger        *logger.Logger
	retryConfig   config.RetryConfig
	config        config.ProvidersConfig
}

// NewProviderManager creates a new provider manager
func NewProviderManager(cfg config.ProvidersConfig, log *logger.Logger) *ProviderManager {
	pm := &ProviderManager{
		providers:     make([]Provider, 0),
		openaiClients: make([]*openai.Client, 0),
		logger:        log,
		retryConfig:   cfg.Claude.Retry,
		config:        cfg,
	}

	// Initialize Claude providers
	if cfg.Claude.Enabled {
		for i, apiKeyConfig := range cfg.Claude.APIKeys {
			client := claude.NewClient(apiKeyConfig.Key, cfg.Claude, log)
			pm.providers = append(pm.providers, client)
			log.Infof("Initialized Claude provider %d with key: %s", i+1, client.GetAPIKey())
		}
	}

	// Initialize OpenAI providers
	if cfg.OpenAI.Enabled {
		for i, apiKeyConfig := range cfg.OpenAI.APIKeys {
			client := openai.NewClient(apiKeyConfig.Key, log)
			pm.openaiClients = append(pm.openaiClients, client)
			log.Infof("Initialized OpenAI provider %d with key: %s", i+1, client.GetAPIKey())
		}
	}

	totalProviders := len(pm.providers) + len(pm.openaiClients)
	if totalProviders == 0 {
		log.Warn("No providers initialized!")
	} else {
		log.Infof("Provider manager initialized with %d provider(s) (Claude: %d, OpenAI: %d)",
			totalProviders, len(pm.providers), len(pm.openaiClients))
	}

	return pm
}

// CreateMessage sends a request with load balancing and retry logic
func (pm *ProviderManager) CreateMessage(ctx context.Context, req *models.ClaudeRequest) (*models.ClaudeResponse, error) {
	if len(pm.providers) == 0 {
		return nil, fmt.Errorf("no providers available")
	}

	var lastErr error
	attempts := 0
	maxAttempts := pm.retryConfig.MaxAttempts * len(pm.providers)

	for attempts < maxAttempts {
		// Get next provider (round-robin)
		provider := pm.getNextProvider()

		// Try to send request
		resp, err := provider.CreateMessage(ctx, req)
		if err == nil {
			// Success!
			if attempts > 0 {
				pm.logger.Infof("Request succeeded after %d attempt(s)", attempts+1)
			}
			return resp, nil
		}

		lastErr = err
		attempts++

		// Check if error is retryable
		if claudeErr, ok := err.(*claude.APIError); ok {
			if !claudeErr.IsRetryable() {
				// Non-retryable error (e.g., 4xx client errors)
				pm.logger.Warnf("Non-retryable error from provider: %v", err)
				return nil, err
			}

			// Rate limit error - try next provider immediately
			if claudeErr.IsRateLimitError() {
				pm.logger.Warnf("Rate limit hit, trying next provider")
				continue
			}
		}

		// Calculate backoff for retryable errors
		if attempts < maxAttempts {
			backoff := pm.calculateBackoff(attempts)
			pm.logger.Warnf("Request failed (attempt %d/%d), retrying after %v: %v", attempts, maxAttempts, backoff, err)

			// Wait before retry
			select {
			case <-time.After(backoff):
				// Continue to next attempt
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			}
		}
	}

	return nil, fmt.Errorf("all retry attempts exhausted: %w", lastErr)
}

// getNextProvider returns the next provider using round-robin
func (pm *ProviderManager) getNextProvider() Provider {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.providers) == 0 {
		return nil
	}

	provider := pm.providers[pm.currentIndex]
	pm.currentIndex = (pm.currentIndex + 1) % len(pm.providers)

	return provider
}

// calculateBackoff calculates exponential backoff duration
func (pm *ProviderManager) calculateBackoff(attempt int) time.Duration {
	backoff := float64(pm.retryConfig.InitialBackoff) * math.Pow(pm.retryConfig.BackoffMultiplier, float64(attempt))

	maxBackoff := float64(pm.retryConfig.MaxBackoff)
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	return time.Duration(backoff)
}

// Health checks the health of all providers
func (pm *ProviderManager) Health(ctx context.Context) error {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	totalProviders := len(pm.providers) + len(pm.openaiClients)
	if totalProviders == 0 {
		return fmt.Errorf("no providers configured")
	}

	var healthyCount int

	// Check Claude providers
	for i, provider := range pm.providers {
		if err := provider.Health(ctx); err != nil {
			pm.logger.Warnf("Claude provider %d unhealthy: %v", i+1, err)
		} else {
			healthyCount++
		}
	}

	// Check OpenAI providers
	for i, client := range pm.openaiClients {
		if err := client.Health(ctx); err != nil {
			pm.logger.Warnf("OpenAI provider %d unhealthy: %v", i+1, err)
		} else {
			healthyCount++
		}
	}

	if healthyCount == 0 {
		return fmt.Errorf("all providers unhealthy")
	}

	pm.logger.Debugf("Provider health: %d/%d healthy", healthyCount, totalProviders)
	return nil
}

// GetProviderCount returns the number of configured providers
func (pm *ProviderManager) GetProviderCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.providers) + len(pm.openaiClients)
}

// GetAvailableModels returns list of available models from all providers
func (pm *ProviderManager) GetAvailableModels() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	models := make([]string, 0)

	// Add Claude models
	if pm.config.Claude.Enabled && len(pm.config.Claude.Models) > 0 {
		models = append(models, pm.config.Claude.Models...)
	}

	// Add OpenAI models
	if pm.config.OpenAI.Enabled && len(pm.config.OpenAI.Models) > 0 {
		models = append(models, pm.config.OpenAI.Models...)
	}

	// If no models configured, return defaults
	if len(models) == 0 {
		models = []string{"claude-3-haiku-20240307"}
	}

	return models
}

// GetActiveProviderIDs returns list of provider IDs that are enabled
func (pm *ProviderManager) GetActiveProviderIDs() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	activeProviders := make([]string, 0)

	// Check if Claude is enabled and has providers
	if pm.config.Claude.Enabled && len(pm.providers) > 0 {
		activeProviders = append(activeProviders, "claude")
	}

	// Check if OpenAI is enabled and has clients
	if pm.config.OpenAI.Enabled && len(pm.openaiClients) > 0 {
		activeProviders = append(activeProviders, "openai")
	}

	return activeProviders
}

// GetClaudeClient returns the first Claude client for streaming
// TODO: Improve this to support streaming with load balancing
func (pm *ProviderManager) GetClaudeClient() *claude.Client {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.providers) == 0 {
		return nil
	}

	// Return first provider as Claude client
	if client, ok := pm.providers[0].(*claude.Client); ok {
		return client
	}

	return nil
}

// GetOpenAIClient returns the first OpenAI client
// TODO: Improve this to support load balancing
func (pm *ProviderManager) GetOpenAIClient() *openai.Client {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.openaiClients) == 0 {
		return nil
	}

	// Return first OpenAI client
	return pm.openaiClients[0]
}

// DetermineProvider determines which provider to use based on model name
func (pm *ProviderManager) DetermineProvider(model string) string {
	// Simple heuristic based on model name
	if len(model) > 0 {
		// OpenAI models start with "gpt", "o3", "o4", "text-", "dall-e", etc.
		if model[0] == 'g' && len(model) >= 3 && model[:3] == "gpt" {
			return "openai"
		}
		if model[0] == 'o' && len(model) >= 2 {
			// o3, o3-mini, o3-pro, o4, o4-mini, etc.
			if model[1] == '3' || model[1] == '4' {
				return "openai"
			}
		}
		if len(model) >= 5 && model[:5] == "text-" {
			return "openai"
		}
		if len(model) >= 7 && model[:7] == "dall-e-" {
			return "openai"
		}
		// Claude models start with "claude"
		if len(model) >= 6 && model[:6] == "claude" {
			return "claude"
		}
	}

	// Default to Claude if uncertain
	return "claude"
}
