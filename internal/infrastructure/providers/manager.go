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
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers/abacus"
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

// clientState tracks metadata for load balancing and rate limiting
type clientState struct {
	client       interface{} // *claude.Client, *openai.Client, or *abacus.Client
	weight       int         // Weight for load balancing (higher = more requests)
	maxRPM       int         // Max requests per minute (0 = unlimited)
	requestCount int         // Current request count in time window
	windowStart  time.Time   // Start of current rate limit window
}

// resetRateLimitIfNeeded resets the request counter if the time window has expired
func (cs *clientState) resetRateLimitIfNeeded() {
	if cs.maxRPM == 0 {
		return // No rate limit
	}

	now := time.Now()
	if now.Sub(cs.windowStart) >= time.Minute {
		cs.requestCount = 0
		cs.windowStart = now
	}
}

// canAcceptRequest checks if this client can accept a new request
func (cs *clientState) canAcceptRequest() bool {
	cs.resetRateLimitIfNeeded()

	if cs.maxRPM == 0 {
		return true // No rate limit
	}

	return cs.requestCount < cs.maxRPM
}

// incrementRequestCount increments the request counter
func (cs *clientState) incrementRequestCount() {
	cs.resetRateLimitIfNeeded()
	cs.requestCount++
}

// ProviderManager manages multiple provider instances with load balancing
type ProviderManager struct {
	providers     []Provider
	claudeStates  []*clientState
	openaiStates  []*clientState
	abacusStates  []*clientState
	currentIndex  int
	mu            sync.RWMutex
	logger        *logger.Logger
	retryConfig   config.RetryConfig
	config        config.ProvidersConfig
	dbKeyProvider DBKeyProvider // Store reference for hot-reload
}

// DBKeyProvider provides decrypted API keys from the database.
// This is used to load provider keys from the DB instead of config.yaml.
type DBKeyProvider interface {
	GetDecryptedKeysByProvider(ctx context.Context, providerID string) ([]DBProviderKey, error)
}

// DBProviderKey holds a decrypted API key loaded from the database.
type DBProviderKey struct {
	APIKey string
	Weight int
	MaxRPM int
}

// NewProviderManager creates a new provider manager.
// It first tries to load API keys from the database. If no DB keys are found
// for a provider, it falls back to the keys defined in config.yaml.
func NewProviderManager(cfg config.ProvidersConfig, log *logger.Logger, dbKeyProvider DBKeyProvider) *ProviderManager {
	pm := &ProviderManager{
		providers:     make([]Provider, 0),
		claudeStates:  make([]*clientState, 0),
		openaiStates:  make([]*clientState, 0),
		abacusStates:  make([]*clientState, 0),
		logger:        log,
		retryConfig:   cfg.Claude.Retry,
		config:        cfg,
		dbKeyProvider: dbKeyProvider, // Store for hot-reload
	}

	ctx := context.Background()

	// Initialize Claude providers (DB keys first, fallback to config)
	if cfg.Claude.Enabled {
		dbKeys := loadDBKeys(ctx, dbKeyProvider, "claude", log)
		if len(dbKeys) > 0 {
			for i, k := range dbKeys {
				client := claude.NewClient(k.APIKey, cfg.Claude, log)
				pm.providers = append(pm.providers, client)
				pm.claudeStates = append(pm.claudeStates, &clientState{
					client:      client,
					weight:      k.Weight,
					maxRPM:      k.MaxRPM,
					windowStart: time.Now(),
				})
				log.Infof("Initialized Claude provider %d from DB with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), k.Weight, k.MaxRPM)
			}
		} else {
			for i, apiKeyConfig := range cfg.Claude.APIKeys {
				client := claude.NewClient(apiKeyConfig.Key, cfg.Claude, log)
				pm.providers = append(pm.providers, client)
				pm.claudeStates = append(pm.claudeStates, &clientState{
					client:      client,
					weight:      apiKeyConfig.Weight,
					maxRPM:      apiKeyConfig.MaxRPM,
					windowStart: time.Now(),
				})
				log.Infof("Initialized Claude provider %d from config with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), apiKeyConfig.Weight, apiKeyConfig.MaxRPM)
			}
		}
	}

	// Initialize OpenAI providers (DB keys first, fallback to config)
	if cfg.OpenAI.Enabled {
		dbKeys := loadDBKeys(ctx, dbKeyProvider, "openai", log)
		if len(dbKeys) > 0 {
			for i, k := range dbKeys {
				client := openai.NewClient(k.APIKey, log)
				pm.openaiStates = append(pm.openaiStates, &clientState{
					client:      client,
					weight:      k.Weight,
					maxRPM:      k.MaxRPM,
					windowStart: time.Now(),
				})
				log.Infof("Initialized OpenAI provider %d from DB with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), k.Weight, k.MaxRPM)
			}
		} else {
			for i, apiKeyConfig := range cfg.OpenAI.APIKeys {
				client := openai.NewClient(apiKeyConfig.Key, log)
				pm.openaiStates = append(pm.openaiStates, &clientState{
					client:      client,
					weight:      apiKeyConfig.Weight,
					maxRPM:      apiKeyConfig.MaxRPM,
					windowStart: time.Now(),
				})
				log.Infof("Initialized OpenAI provider %d from config with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), apiKeyConfig.Weight, apiKeyConfig.MaxRPM)
			}
		}
	}

	// Initialize Abacus.ai providers (DB keys first, fallback to config)
	if cfg.Abacus.Enabled {
		dbKeys := loadDBKeys(ctx, dbKeyProvider, "abacus", log)
		if len(dbKeys) > 0 {
			for i, k := range dbKeys {
				client := abacus.NewClient(k.APIKey, cfg.Abacus, log)
				pm.abacusStates = append(pm.abacusStates, &clientState{
					client:      client,
					weight:      k.Weight,
					maxRPM:      k.MaxRPM,
					windowStart: time.Now(),
				})
				log.Infof("Initialized Abacus.ai provider %d from DB with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), k.Weight, k.MaxRPM)
			}
		} else {
			for i, apiKeyConfig := range cfg.Abacus.APIKeys {
				client := abacus.NewClient(apiKeyConfig.Key, cfg.Abacus, log)
				pm.abacusStates = append(pm.abacusStates, &clientState{
					client:      client,
					weight:      apiKeyConfig.Weight,
					maxRPM:      apiKeyConfig.MaxRPM,
					windowStart: time.Now(),
				})
				log.Infof("Initialized Abacus.ai provider %d from config with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), apiKeyConfig.Weight, apiKeyConfig.MaxRPM)
			}
		}
	}

	totalProviders := len(pm.providers) + len(pm.openaiStates) + len(pm.abacusStates)
	if totalProviders == 0 {
		log.Warn("No providers initialized!")
	} else {
		log.Infof("Provider manager initialized with %d provider(s) (Claude: %d, OpenAI: %d, Abacus: %d)",
			totalProviders, len(pm.providers), len(pm.openaiStates), len(pm.abacusStates))
	}

	return pm
}

// loadDBKeys tries to load keys from DB, returns nil on error or if no provider given.
func loadDBKeys(ctx context.Context, dbKeyProvider DBKeyProvider, providerID string, log *logger.Logger) []DBProviderKey {
	if dbKeyProvider == nil {
		return nil
	}
	keys, err := dbKeyProvider.GetDecryptedKeysByProvider(ctx, providerID)
	if err != nil {
		log.Warnf("Failed to load %s keys from DB (falling back to config): %v", providerID, err)
		return nil
	}
	if len(keys) == 0 {
		log.Debugf("No %s keys in DB, using config.yaml", providerID)
		return nil
	}
	log.Infof("Loaded %d %s API key(s) from database", len(keys), providerID)
	return keys
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

	totalProviders := len(pm.providers) + len(pm.openaiStates) + len(pm.abacusStates)
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
	for i, state := range pm.openaiStates {
		client := state.client.(*openai.Client)
		if err := client.Health(ctx); err != nil {
			pm.logger.Warnf("OpenAI provider %d unhealthy: %v", i+1, err)
		} else {
			healthyCount++
		}
	}

	// Check Abacus.ai providers
	for i, state := range pm.abacusStates {
		client := state.client.(*abacus.Client)
		if err := client.Health(ctx); err != nil {
			pm.logger.Warnf("Abacus.ai provider %d unhealthy: %v", i+1, err)
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
	return len(pm.providers) + len(pm.openaiStates) + len(pm.abacusStates)
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

	// Add Abacus.ai models
	if pm.config.Abacus.Enabled && len(pm.config.Abacus.Models) > 0 {
		models = append(models, pm.config.Abacus.Models...)
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
	if pm.config.OpenAI.Enabled && len(pm.openaiStates) > 0 {
		activeProviders = append(activeProviders, "openai")
	}

	// Check if Abacus.ai is enabled and has clients
	if pm.config.Abacus.Enabled && len(pm.abacusStates) > 0 {
		activeProviders = append(activeProviders, "abacus")
	}

	return activeProviders
}

// GetClaudeClient returns a Claude client using weighted load balancing and rate limiting
func (pm *ProviderManager) GetClaudeClient() *claude.Client {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.claudeStates) == 0 {
		return nil
	}

	// Try to find an available client using weighted round-robin
	selectedState := pm.selectWeightedClient(pm.claudeStates)
	if selectedState == nil {
		// All clients are rate-limited, return first one anyway
		pm.logger.Warn("All Claude clients are rate-limited, returning first client")
		return pm.claudeStates[0].client.(*claude.Client)
	}

	// Increment request count for rate limiting
	selectedState.incrementRequestCount()

	return selectedState.client.(*claude.Client)
}

// GetOpenAIClient returns an OpenAI client using weighted load balancing and rate limiting
func (pm *ProviderManager) GetOpenAIClient() *openai.Client {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.openaiStates) == 0 {
		return nil
	}

	// Try to find an available client using weighted round-robin
	selectedState := pm.selectWeightedClient(pm.openaiStates)
	if selectedState == nil {
		// All clients are rate-limited, return first one anyway
		pm.logger.Warn("All OpenAI clients are rate-limited, returning first client")
		return pm.openaiStates[0].client.(*openai.Client)
	}

	// Increment request count for rate limiting
	selectedState.incrementRequestCount()

	return selectedState.client.(*openai.Client)
}

// GetAbacusClient returns an Abacus.ai client using weighted load balancing and rate limiting
func (pm *ProviderManager) GetAbacusClient() *abacus.Client {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if len(pm.abacusStates) == 0 {
		return nil
	}

	// Try to find an available client using weighted round-robin
	selectedState := pm.selectWeightedClient(pm.abacusStates)
	if selectedState == nil {
		// All clients are rate-limited, return first one anyway
		pm.logger.Warn("All Abacus.ai clients are rate-limited, returning first client")
		return pm.abacusStates[0].client.(*abacus.Client)
	}

	// Increment request count for rate limiting
	selectedState.incrementRequestCount()

	return selectedState.client.(*abacus.Client)
}

// selectWeightedClient selects a client using weighted round-robin with rate limiting
// Returns nil if all clients are rate-limited
func (pm *ProviderManager) selectWeightedClient(states []*clientState) *clientState {
	if len(states) == 0 {
		return nil
	}

	// Calculate total weight of available (non-rate-limited) clients
	totalWeight := 0
	availableStates := make([]*clientState, 0)

	for _, state := range states {
		if state.canAcceptRequest() {
			totalWeight += state.weight
			availableStates = append(availableStates, state)
		}
	}

	// If no clients are available (all rate-limited), return nil
	if len(availableStates) == 0 {
		return nil
	}

	// If only one client available, use it
	if len(availableStates) == 1 {
		return availableStates[0]
	}

	// Weighted random selection
	// Generate a random number between 0 and totalWeight
	randWeight := time.Now().UnixNano() % int64(totalWeight)
	currentWeight := int64(0)

	for _, state := range availableStates {
		currentWeight += int64(state.weight)
		if randWeight < currentWeight {
			return state
		}
	}

	// Fallback to first available client
	return availableStates[0]
}

// DetermineProvider determines which provider to use based on model name
func (pm *ProviderManager) DetermineProvider(model string) string {
	// Simple heuristic based on model name
	if len(model) > 0 {
		// Abacus.ai models start with "abacus:"
		if len(model) >= 7 && model[:7] == "abacus:" {
			return "abacus"
		}
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

// ReloadKeys reloads API keys from the database and reinitializes provider clients.
// This allows hot-reloading of keys without restarting the backend.
func (pm *ProviderManager) ReloadKeys(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.logger.Info("Reloading provider API keys from database...")

	// Clear existing providers
	pm.providers = make([]Provider, 0)
	pm.claudeStates = make([]*clientState, 0)
	pm.openaiStates = make([]*clientState, 0)
	pm.abacusStates = make([]*clientState, 0)
	pm.currentIndex = 0

	// Reload Claude providers (DB keys first, fallback to config)
	if pm.config.Claude.Enabled {
		dbKeys := loadDBKeys(ctx, pm.dbKeyProvider, "claude", pm.logger)
		if len(dbKeys) > 0 {
			for i, k := range dbKeys {
				client := claude.NewClient(k.APIKey, pm.config.Claude, pm.logger)
				pm.providers = append(pm.providers, client)
				pm.claudeStates = append(pm.claudeStates, &clientState{
					client:      client,
					weight:      k.Weight,
					maxRPM:      k.MaxRPM,
					windowStart: time.Now(),
				})
				pm.logger.Infof("Reloaded Claude provider %d from DB with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), k.Weight, k.MaxRPM)
			}
		} else {
			for i, apiKeyConfig := range pm.config.Claude.APIKeys {
				client := claude.NewClient(apiKeyConfig.Key, pm.config.Claude, pm.logger)
				pm.providers = append(pm.providers, client)
				pm.claudeStates = append(pm.claudeStates, &clientState{
					client:      client,
					weight:      apiKeyConfig.Weight,
					maxRPM:      apiKeyConfig.MaxRPM,
					windowStart: time.Now(),
				})
				pm.logger.Infof("Reloaded Claude provider %d from config with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), apiKeyConfig.Weight, apiKeyConfig.MaxRPM)
			}
		}
	}

	// Reload OpenAI providers (DB keys first, fallback to config)
	if pm.config.OpenAI.Enabled {
		dbKeys := loadDBKeys(ctx, pm.dbKeyProvider, "openai", pm.logger)
		if len(dbKeys) > 0 {
			for i, k := range dbKeys {
				client := openai.NewClient(k.APIKey, pm.logger)
				pm.openaiStates = append(pm.openaiStates, &clientState{
					client:      client,
					weight:      k.Weight,
					maxRPM:      k.MaxRPM,
					windowStart: time.Now(),
				})
				pm.logger.Infof("Reloaded OpenAI provider %d from DB with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), k.Weight, k.MaxRPM)
			}
		} else {
			for i, apiKeyConfig := range pm.config.OpenAI.APIKeys {
				client := openai.NewClient(apiKeyConfig.Key, pm.logger)
				pm.openaiStates = append(pm.openaiStates, &clientState{
					client:      client,
					weight:      apiKeyConfig.Weight,
					maxRPM:      apiKeyConfig.MaxRPM,
					windowStart: time.Now(),
				})
				pm.logger.Infof("Reloaded OpenAI provider %d from config with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), apiKeyConfig.Weight, apiKeyConfig.MaxRPM)
			}
		}
	}

	// Reload Abacus.ai providers (DB keys first, fallback to config)
	if pm.config.Abacus.Enabled {
		dbKeys := loadDBKeys(ctx, pm.dbKeyProvider, "abacus", pm.logger)
		if len(dbKeys) > 0 {
			for i, k := range dbKeys {
				client := abacus.NewClient(k.APIKey, pm.config.Abacus, pm.logger)
				pm.abacusStates = append(pm.abacusStates, &clientState{
					client:      client,
					weight:      k.Weight,
					maxRPM:      k.MaxRPM,
					windowStart: time.Now(),
				})
				pm.logger.Infof("Reloaded Abacus.ai provider %d from DB with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), k.Weight, k.MaxRPM)
			}
		} else {
			for i, apiKeyConfig := range pm.config.Abacus.APIKeys {
				client := abacus.NewClient(apiKeyConfig.Key, pm.config.Abacus, pm.logger)
				pm.abacusStates = append(pm.abacusStates, &clientState{
					client:      client,
					weight:      apiKeyConfig.Weight,
					maxRPM:      apiKeyConfig.MaxRPM,
					windowStart: time.Now(),
				})
				pm.logger.Infof("Reloaded Abacus.ai provider %d from config with key: %s (weight: %d, maxRPM: %d)",
					i+1, client.GetAPIKey(), apiKeyConfig.Weight, apiKeyConfig.MaxRPM)
			}
		}
	}

	totalProviders := len(pm.providers) + len(pm.openaiStates) + len(pm.abacusStates)
	if totalProviders == 0 {
		pm.logger.Warn("No providers initialized after reload!")
		return fmt.Errorf("no providers available after reload")
	}

	pm.logger.Infof("Provider keys reloaded: %d provider(s) (Claude: %d, OpenAI: %d, Abacus: %d)",
		totalProviders, len(pm.providers), len(pm.openaiStates), len(pm.abacusStates))

	return nil
}
