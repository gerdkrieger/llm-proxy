// Package cache provides Redis caching functionality.
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/config"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the redis.Client with additional functionality
type RedisClient struct {
	client *redis.Client
	config config.RedisConfig
	logger *logger.Logger
}

// New creates a new Redis client
func New(cfg config.RedisConfig, log *logger.Logger) (*RedisClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create Redis client
	log.Infof("Connecting to Redis at %s:%d", cfg.Host, cfg.Port)

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.PoolSize / 4,

		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// Pool settings
		PoolTimeout:     4 * time.Second,
		ConnMaxLifetime: time.Hour,
		ConnMaxIdleTime: 30 * time.Minute,
	})

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	log.Info("Successfully connected to Redis")

	return &RedisClient{
		client: client,
		config: cfg,
		logger: log,
	}, nil
}

// Close gracefully closes the Redis client
func (r *RedisClient) Close() error {
	if r.client != nil {
		r.logger.Info("Closing Redis client")
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis client: %w", err)
		}
		r.logger.Info("Redis client closed")
	}
	return nil
}

// Health checks the health of the Redis connection
func (r *RedisClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key does not exist
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Set stores a value in cache with TTL
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Delete removes a key from cache
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return count > 0, nil
}

// Expire sets a TTL on a key
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := r.client.Expire(ctx, key, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set expiration on key %s: %w", key, err)
	}
	return nil
}

// Increment increments a counter
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// Decrement decrements a counter
func (r *RedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return val, nil
}

// IncrementBy increments a counter by a specific amount
func (r *RedisClient) IncrementBy(ctx context.Context, key string, amount int64) (int64, error) {
	val, err := r.client.IncrBy(ctx, key, amount).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, amount, err)
	}
	return val, nil
}

// SetNX sets a key only if it doesn't exist (atomic)
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	success, err := r.client.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setnx key %s: %w", key, err)
	}
	return success, nil
}

// GetSet sets a key and returns the old value
func (r *RedisClient) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	val, err := r.client.GetSet(ctx, key, value).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to getset key %s: %w", key, err)
	}
	return val, nil
}

// MGet gets multiple keys at once
func (r *RedisClient) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to mget keys: %w", err)
	}
	return vals, nil
}

// MSet sets multiple keys at once
func (r *RedisClient) MSet(ctx context.Context, pairs ...interface{}) error {
	if err := r.client.MSet(ctx, pairs...).Err(); err != nil {
		return fmt.Errorf("failed to mset keys: %w", err)
	}
	return nil
}

// Keys finds all keys matching a pattern
func (r *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys matching pattern %s: %w", pattern, err)
	}
	return keys, nil
}

// Scan finds all keys matching a pattern using SCAN (more efficient than Keys)
func (r *RedisClient) Scan(ctx context.Context, pattern string) ([]string, error) {
	var keys []string
	var cursor uint64

	// Use SCAN to iterate through keys (avoids blocking Redis)
	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys matching pattern %s: %w", pattern, err)
		}

		keys = append(keys, scanKeys...)

		// cursor == 0 means we've scanned all keys
		if cursor == 0 {
			break
		}
	}

	return keys, nil
}

// FlushDB clears all keys in the current database (use with caution!)
func (r *RedisClient) FlushDB(ctx context.Context) error {
	if err := r.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush database: %w", err)
	}
	r.logger.Warn("Redis database flushed")
	return nil
}

// Pipeline creates a new Redis pipeline for batch operations
func (r *RedisClient) Pipeline() redis.Pipeliner {
	return r.client.Pipeline()
}

// Stats returns Redis client statistics
func (r *RedisClient) Stats() *redis.PoolStats {
	return r.client.PoolStats()
}

// LogStats logs current Redis pool statistics
func (r *RedisClient) LogStats() {
	stats := r.Stats()
	r.logger.GetZerolog().Info().
		Uint32("hits", stats.Hits).
		Uint32("misses", stats.Misses).
		Uint32("timeouts", stats.Timeouts).
		Uint32("total_conns", stats.TotalConns).
		Uint32("idle_conns", stats.IdleConns).
		Uint32("stale_conns", stats.StaleConns).
		Msg("Redis pool statistics")
}

// GetClient returns the underlying Redis client (for advanced usage)
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// MustConnect creates a new Redis client or panics
func MustConnect(cfg config.RedisConfig, log *logger.Logger) *RedisClient {
	client, err := New(cfg, log)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	return client
}
