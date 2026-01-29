// Package caching provides response caching functionality.
package caching

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/cache"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// Service handles caching of LLM responses
type Service struct {
	redis  *cache.RedisClient
	config config.CacheConfig
	logger *logger.Logger
	stats  *Stats
}

// Stats tracks cache performance metrics
type Stats struct {
	Hits   int64
	Misses int64
	Errors int64
}

// NewService creates a new caching service
func NewService(redis *cache.RedisClient, cfg config.CacheConfig, log *logger.Logger) *Service {
	return &Service{
		redis:  redis,
		config: cfg,
		logger: log,
		stats:  &Stats{},
	}
}

// GenerateCacheKey generates a unique cache key for a request
func (s *Service) GenerateCacheKey(req *models.OpenAIRequest) string {
	// Create a deterministic representation of the request
	// We include: model, messages, temperature, max_tokens, top_p
	// We exclude: stream, user, and other metadata that don't affect output

	keyData := struct {
		Model       string                 `json:"model"`
		Messages    []models.OpenAIMessage `json:"messages"`
		Temperature *float64               `json:"temperature,omitempty"`
		MaxTokens   *int                   `json:"max_tokens,omitempty"`
		TopP        *float64               `json:"top_p,omitempty"`
		Stop        interface{}            `json:"stop,omitempty"`
	}{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stop:        req.Stop,
	}

	// Serialize to JSON (deterministic)
	jsonBytes, err := json.Marshal(keyData)
	if err != nil {
		s.logger.Warnf("Failed to marshal cache key data: %v", err)
		return ""
	}

	// Hash the JSON to create a compact key
	hash := sha256.Sum256(jsonBytes)
	hashStr := hex.EncodeToString(hash[:])

	// Prefix with model for easier debugging/invalidation
	return fmt.Sprintf("llm:response:%s:%s", req.Model, hashStr)
}

// Get retrieves a cached response
func (s *Service) Get(ctx context.Context, key string) (*models.OpenAIResponse, bool) {
	if !s.config.Enabled {
		return nil, false
	}

	// Try to get from cache
	data, err := s.redis.Get(ctx, key)
	if err != nil {
		s.stats.Misses++
		s.logger.Debugf("Cache miss: %s", key)
		return nil, false
	}

	// Deserialize response
	var response models.OpenAIResponse
	if err := json.Unmarshal([]byte(data), &response); err != nil {
		s.stats.Errors++
		s.logger.Warnf("Failed to unmarshal cached response: %v", err)
		return nil, false
	}

	s.stats.Hits++
	s.logger.Debugf("Cache hit: %s", key)
	return &response, true
}

// Set stores a response in cache
func (s *Service) Set(ctx context.Context, key string, response *models.OpenAIResponse) error {
	if !s.config.Enabled {
		return nil
	}

	// Serialize response
	data, err := json.Marshal(response)
	if err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Calculate TTL
	ttl := time.Duration(s.config.TTL) * time.Second

	// Store in cache
	if err := s.redis.Set(ctx, key, string(data), ttl); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to set cache: %w", err)
	}

	s.logger.Debugf("Cached response: %s (TTL: %v)", key, ttl)
	return nil
}

// Delete removes a cache entry
func (s *Service) Delete(ctx context.Context, key string) error {
	if !s.config.Enabled {
		return nil
	}

	if err := s.redis.Delete(ctx, key); err != nil {
		s.stats.Errors++
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	s.logger.Debugf("Deleted cache entry: %s", key)
	return nil
}

// InvalidateByPattern invalidates all cache entries matching a pattern
func (s *Service) InvalidateByPattern(ctx context.Context, pattern string) (int, error) {
	if !s.config.Enabled {
		return 0, nil
	}

	// Use Redis SCAN to find matching keys
	keys, err := s.redis.Scan(ctx, pattern)
	if err != nil {
		s.stats.Errors++
		return 0, fmt.Errorf("failed to scan keys: %w", err)
	}

	// Delete all matching keys
	deleted := 0
	for _, key := range keys {
		if err := s.redis.Delete(ctx, key); err != nil {
			s.logger.Warnf("Failed to delete key %s: %v", key, err)
		} else {
			deleted++
		}
	}

	s.logger.Infof("Invalidated %d cache entries matching pattern: %s", deleted, pattern)
	return deleted, nil
}

// InvalidateByModel invalidates all cache entries for a specific model
func (s *Service) InvalidateByModel(ctx context.Context, model string) (int, error) {
	pattern := fmt.Sprintf("llm:response:%s:*", model)
	return s.InvalidateByPattern(ctx, pattern)
}

// Clear clears all cache entries
func (s *Service) Clear(ctx context.Context) error {
	if !s.config.Enabled {
		return nil
	}

	count, err := s.InvalidateByPattern(ctx, "llm:response:*")
	if err != nil {
		return err
	}

	s.logger.Infof("Cleared all cache entries: %d deleted", count)
	return nil
}

// GetStats returns cache performance statistics
func (s *Service) GetStats() *Stats {
	return s.stats
}

// GetHitRate returns the cache hit rate as a percentage
func (s *Service) GetHitRate() float64 {
	total := s.stats.Hits + s.stats.Misses
	if total == 0 {
		return 0.0
	}
	return float64(s.stats.Hits) / float64(total) * 100.0
}

// ResetStats resets cache statistics
func (s *Service) ResetStats() {
	s.stats = &Stats{}
	s.logger.Info("Cache statistics reset")
}
