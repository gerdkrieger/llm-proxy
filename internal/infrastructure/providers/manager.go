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
	providers    []Provider
	currentIndex int
	mu           sync.RWMutex
	logger       *logger.Logger
	retryConfig  config.RetryConfig
}

// NewProviderManager creates a new provider manager
func NewProviderManager(cfg config.ProvidersConfig, log *logger.Logger) *ProviderManager {
	pm := &ProviderManager{
		providers:   make([]Provider, 0),
		logger:      log,
		retryConfig: cfg.Claude.Retry,
	}

	// Initialize Claude providers
	if cfg.Claude.Enabled {
		for i, apiKeyConfig := range cfg.Claude.APIKeys {
			client := claude.NewClient(apiKeyConfig.Key, cfg.Claude, log)
			pm.providers = append(pm.providers, client)
			log.Infof("Initialized Claude provider %d with key: %s", i+1, client.GetAPIKey())
		}
	}

	if len(pm.providers) == 0 {
		log.Warn("No providers initialized!")
	} else {
		log.Infof("Provider manager initialized with %d provider(s)", len(pm.providers))
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

	if len(pm.providers) == 0 {
		return fmt.Errorf("no providers configured")
	}

	var healthyCount int
	for i, provider := range pm.providers {
		if err := provider.Health(ctx); err != nil {
			pm.logger.Warnf("Provider %d unhealthy: %v", i+1, err)
		} else {
			healthyCount++
		}
	}

	if healthyCount == 0 {
		return fmt.Errorf("all providers unhealthy")
	}

	pm.logger.Debugf("Provider health: %d/%d healthy", healthyCount, len(pm.providers))
	return nil
}

// GetProviderCount returns the number of configured providers
func (pm *ProviderManager) GetProviderCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.providers)
}

// GetAvailableModels returns list of available models
func (pm *ProviderManager) GetAvailableModels() []string {
	// For now, return Claude models
	// This could be extended to query providers dynamically
	return []string{
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}
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
