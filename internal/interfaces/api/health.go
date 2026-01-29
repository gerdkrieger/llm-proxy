// Package api provides HTTP handlers for the API.
package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/infrastructure/cache"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
	"github.com/llm-proxy/llm-proxy/pkg/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db     *database.DB
	redis  *cache.RedisClient
	logger *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *database.DB, redis *cache.RedisClient, log *logger.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		logger: log,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Version   string            `json:"version,omitempty"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// SimpleHealth handles the simple health check endpoint (for liveness)
// GET /health
func (h *HealthHandler) SimpleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DetailedHealth handles the detailed health check endpoint (for readiness)
// GET /health/ready
func (h *HealthHandler) DetailedHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checks := make(map[string]string)
	overallStatus := "healthy"

	// Check database
	if err := h.db.Health(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
		h.logger.Error(err, "Database health check failed")
	} else {
		checks["database"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		checks["redis"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
		h.logger.Error(err, "Redis health check failed")
	} else {
		checks["redis"] = "healthy"
	}

	// Determine HTTP status code
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// DetailedHealthWithStats handles detailed health check with statistics
// GET /health/detailed
func (h *HealthHandler) DetailedHealthWithStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	checks := make(map[string]string)
	overallStatus := "healthy"

	// Check database
	if err := h.db.Health(ctx); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
	} else {
		dbStats := h.db.Stats()
		checks["database"] = "healthy"
		checks["database_total_conns"] = string(rune(dbStats.TotalConns()))
		checks["database_idle_conns"] = string(rune(dbStats.IdleConns()))
	}

	// Check Redis
	if err := h.redis.Health(ctx); err != nil {
		checks["redis"] = "unhealthy: " + err.Error()
		overallStatus = "unhealthy"
	} else {
		redisStats := h.redis.Stats()
		checks["redis"] = "healthy"
		checks["redis_total_conns"] = string(rune(redisStats.TotalConns))
		checks["redis_idle_conns"] = string(rune(redisStats.IdleConns))
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC(),
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
