// Package middleware provides HTTP middleware for rate limiting.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/cache"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// RateLimitMiddleware enforces per-client rate limits (RPM/RPD) using Redis.
type RateLimitMiddleware struct {
	enabled   bool
	defaultRPM int
	defaultRPD int
	redis     *cache.RedisClient
	logger    *logger.Logger
}

// NewRateLimitMiddleware creates a new rate limiting middleware.
func NewRateLimitMiddleware(cfg config.RateLimitingConfig, redis *cache.RedisClient, log *logger.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		enabled:    cfg.Enabled,
		defaultRPM: cfg.DefaultRPM,
		defaultRPD: cfg.DefaultRPD,
		redis:      redis,
		logger:     log,
	}
}

// Limit applies rate limiting to the request. It uses the client identifier
// (from context) as the rate limit key. Skips limiting if Redis is unavailable
// or the request is unauthenticated (defense in depth: auth middleware runs first).
func (m *RateLimitMiddleware) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled || m.redis == nil {
			next.ServeHTTP(w, r)
			return
		}

		clientID := GetClientID(r.Context())
		if clientID == "" {
			// No client identification — let auth middleware reject.
			// We don't rate-limit unauthenticated requests here.
			next.ServeHTTP(w, r)
			return
		}

		allowed, err := m.checkLimit(r.Context(), clientID)
		if err != nil {
			// Soft fail: log and allow (don't break service if Redis is down)
			m.logger.Warnf("Rate limit check failed for %s: %v — allowing request", clientID, err)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			m.logger.Warnf("Rate limit exceeded for client: %s", clientID)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "60")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "rate_limit_exceeded",
				"message": "Rate limit exceeded. Please reduce request frequency.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

// checkLimit verifies both RPM and RPD limits atomically via Redis Lua or pipeline.
func (m *RateLimitMiddleware) checkLimit(ctx context.Context, clientID string) (bool, error) {
	now := time.Now()
	minuteKey := fmt.Sprintf("ratelimit:rpm:%s:%s", clientID, now.Format("2006-01-02T15:04"))
	dayKey := fmt.Sprintf("ratelimit:rpd:%s:%s", clientID, now.Format("2006-01-02"))

	// Use pipeline for atomic multi-key check+increment
	pipe := m.redis.GetClient().Pipeline()

	incrRPM := pipe.Incr(ctx, minuteKey)
	_ = pipe.Expire(ctx, minuteKey, 2*time.Minute)
	incrRPD := pipe.Incr(ctx, dayKey)
	_ = pipe.Expire(ctx, dayKey, 25*time.Hour)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("redis pipeline failed: %w", err)
	}

	// Check limits (first request after window reset counts as 1)
	if m.defaultRPM > 0 && incrRPM.Val() > int64(m.defaultRPM) {
		return false, nil
	}
	if m.defaultRPD > 0 && incrRPD.Val() > int64(m.defaultRPD) {
		return false, nil
	}

	return true, nil
}
