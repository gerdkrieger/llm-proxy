// Package repositories provides request logging repository.
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// RequestLog represents a logged API request
type RequestLog struct {
	ID               uuid.UUID
	ClientID         *uuid.UUID
	RequestID        string
	Method           string
	Path             string
	Model            string
	Provider         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CostUSD          float64
	DurationMS       int
	StatusCode       int
	Cached           bool
	IPAddress        *string
	UserAgent        *string
	ErrorMessage     *string
	CreatedAt        time.Time

	// New fields for Live Monitor (added in migration 000007)
	AuthType     *string `json:"auth_type"`     // 'api_key', 'oauth', 'admin', 'none'
	APIKeyName   *string `json:"api_key_name"`  // Name of API key (not the actual key!)
	WasFiltered  bool    `json:"was_filtered"`  // Whether content was filtered
	FilterReason *string `json:"filter_reason"` // Reason for filtering

	// Request/Response content (added in migration 000008)
	RequestHeaders    map[string]interface{} `json:"request_headers,omitempty"`     // Request HTTP headers
	RequestBody       *string                `json:"request_body,omitempty"`        // Request body
	ResponseHeaders   map[string]interface{} `json:"response_headers,omitempty"`    // Response HTTP headers
	ResponseBody      *string                `json:"response_body,omitempty"`       // Response body
	ResponseSizeBytes *int64                 `json:"response_size_bytes,omitempty"` // Response size
}

// RequestLogRepository handles request log database operations
type RequestLogRepository struct {
	db *database.DB
}

// NewRequestLogRepository creates a new request log repository
func NewRequestLogRepository(db *database.DB) *RequestLogRepository {
	return &RequestLogRepository{db: db}
}

// Create creates a new request log entry
func (r *RequestLogRepository) Create(ctx context.Context, log *RequestLog) error {
	query := `
		INSERT INTO request_logs (
			id, client_id, request_id, method, path, model, provider,
			prompt_tokens, completion_tokens, total_tokens, cost_usd,
			duration_ms, status_code, cached, ip_address, user_agent, error_message,
			auth_type, api_key_name, was_filtered, filter_reason,
			request_headers, request_body, response_headers, response_body, response_size_bytes
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26)
		RETURNING created_at
	`

	err := r.db.Pool.QueryRow(
		ctx, query,
		log.ID,
		log.ClientID,
		log.RequestID,
		log.Method,
		log.Path,
		log.Model,
		log.Provider,
		log.PromptTokens,
		log.CompletionTokens,
		log.TotalTokens,
		log.CostUSD,
		log.DurationMS,
		log.StatusCode,
		log.Cached,
		log.IPAddress,
		log.UserAgent,
		log.ErrorMessage,
		log.AuthType,
		log.APIKeyName,
		log.WasFiltered,
		log.FilterReason,
		log.RequestHeaders,
		log.RequestBody,
		log.ResponseHeaders,
		log.ResponseBody,
		log.ResponseSizeBytes,
	).Scan(&log.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create request log: %w", err)
	}

	return nil
}

// GetByID retrieves a log entry by ID with full details
func (r *RequestLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*RequestLog, error) {
	query := `
		SELECT 
			id, client_id, request_id, method, path, model, provider,
			prompt_tokens, completion_tokens, total_tokens, cost_usd,
			duration_ms, status_code, cached, ip_address, user_agent, error_message, created_at,
			auth_type, api_key_name, was_filtered, filter_reason,
			request_headers, request_body, response_headers, response_body, response_size_bytes
		FROM request_logs
		WHERE id = $1
	`

	log := &RequestLog{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.ClientID,
		&log.RequestID,
		&log.Method,
		&log.Path,
		&log.Model,
		&log.Provider,
		&log.PromptTokens,
		&log.CompletionTokens,
		&log.TotalTokens,
		&log.CostUSD,
		&log.DurationMS,
		&log.StatusCode,
		&log.Cached,
		&log.IPAddress,
		&log.UserAgent,
		&log.ErrorMessage,
		&log.CreatedAt,
		&log.AuthType,
		&log.APIKeyName,
		&log.WasFiltered,
		&log.FilterReason,
		&log.RequestHeaders,
		&log.RequestBody,
		&log.ResponseHeaders,
		&log.ResponseBody,
		&log.ResponseSizeBytes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get request log by ID: %w", err)
	}

	return log, nil
}

// GetByRequestID retrieves a log entry by request ID
func (r *RequestLogRepository) GetByRequestID(ctx context.Context, requestID string) (*RequestLog, error) {
	query := `
		SELECT 
			id, client_id, request_id, method, path, model, provider,
			prompt_tokens, completion_tokens, total_tokens, cost_usd,
			duration_ms, status_code, cached, ip_address, user_agent, error_message, created_at,
			auth_type, api_key_name, was_filtered, filter_reason,
			request_headers, request_body, response_headers, response_body, response_size_bytes
		FROM request_logs
		WHERE request_id = $1
	`

	log := &RequestLog{}
	err := r.db.Pool.QueryRow(ctx, query, requestID).Scan(
		&log.ID,
		&log.ClientID,
		&log.RequestID,
		&log.Method,
		&log.Path,
		&log.Model,
		&log.Provider,
		&log.PromptTokens,
		&log.CompletionTokens,
		&log.TotalTokens,
		&log.CostUSD,
		&log.DurationMS,
		&log.StatusCode,
		&log.Cached,
		&log.IPAddress,
		&log.UserAgent,
		&log.ErrorMessage,
		&log.CreatedAt,
		&log.AuthType,
		&log.APIKeyName,
		&log.WasFiltered,
		&log.FilterReason,
		&log.RequestHeaders,
		&log.RequestBody,
		&log.ResponseHeaders,
		&log.ResponseBody,
		&log.ResponseSizeBytes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get request log: %w", err)
	}

	return log, nil
}

// List retrieves request logs with filters
func (r *RequestLogRepository) List(ctx context.Context, filters RequestLogFilters) ([]*RequestLog, error) {
	query := `
		SELECT 
			id, client_id, request_id, method, path, model, provider,
			prompt_tokens, completion_tokens, total_tokens, cost_usd,
			duration_ms, status_code, cached, ip_address, user_agent, error_message, created_at,
			auth_type, api_key_name, was_filtered, filter_reason
		FROM request_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if filters.ClientID != nil {
		query += fmt.Sprintf(" AND client_id = $%d", argCount)
		args = append(args, filters.ClientID)
		argCount++
	}

	if filters.Model != "" {
		query += fmt.Sprintf(" AND model = $%d", argCount)
		args = append(args, filters.Model)
		argCount++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filters.StartTime)
		argCount++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filters.EndTime)
		argCount++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
		argCount++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filters.Offset)
	}

	rows, err := r.db.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list request logs: %w", err)
	}
	defer rows.Close()

	var logs []*RequestLog
	for rows.Next() {
		log := &RequestLog{}
		err := rows.Scan(
			&log.ID,
			&log.ClientID,
			&log.RequestID,
			&log.Method,
			&log.Path,
			&log.Model,
			&log.Provider,
			&log.PromptTokens,
			&log.CompletionTokens,
			&log.TotalTokens,
			&log.CostUSD,
			&log.DurationMS,
			&log.StatusCode,
			&log.Cached,
			&log.IPAddress,
			&log.UserAgent,
			&log.ErrorMessage,
			&log.CreatedAt,
			&log.AuthType,
			&log.APIKeyName,
			&log.WasFiltered,
			&log.FilterReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan request log: %w", err)
		}
		logs = append(logs, log)
	}

	return logs, nil
}

// RequestLogFilters represents filters for request log queries
type RequestLogFilters struct {
	ClientID  *uuid.UUID
	Model     string
	StartTime *time.Time
	EndTime   *time.Time
	Limit     int
	Offset    int
}

// GetStats retrieves aggregate statistics
func (r *RequestLogRepository) GetStats(ctx context.Context, filters RequestLogFilters) (*RequestLogStats, error) {
	query := `
		SELECT 
			COUNT(*) as total_requests,
			SUM(total_tokens) as total_tokens,
			SUM(cost_usd) as total_cost,
			AVG(duration_ms) as avg_duration,
			SUM(CASE WHEN cached THEN 1 ELSE 0 END) as cached_requests,
			SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END) as error_requests
		FROM request_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if filters.ClientID != nil {
		query += fmt.Sprintf(" AND client_id = $%d", argCount)
		args = append(args, filters.ClientID)
		argCount++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filters.StartTime)
		argCount++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filters.EndTime)
	}

	stats := &RequestLogStats{}
	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&stats.TotalRequests,
		&stats.TotalTokens,
		&stats.TotalCost,
		&stats.AvgDuration,
		&stats.CachedRequests,
		&stats.ErrorRequests,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Calculate cache hit rate
	if stats.TotalRequests > 0 {
		stats.CacheHitRate = float64(stats.CachedRequests) / float64(stats.TotalRequests) * 100
	}

	return stats, nil
}

// RequestLogStats represents aggregate statistics
type RequestLogStats struct {
	TotalRequests  int64
	TotalTokens    int64
	TotalCost      float64
	AvgDuration    float64
	CachedRequests int64
	ErrorRequests  int64
	CacheHitRate   float64
}

// GetStatistics retrieves aggregate statistics
func (r *RequestLogRepository) GetStatistics(ctx context.Context, clientID, model string, startTime, endTime time.Time) (*RequestLogStats, error) {
	query := `
		SELECT
			COUNT(*) as total_requests,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(cost_usd), 0) as total_cost,
			COALESCE(AVG(duration_ms), 0) as avg_duration,
			COALESCE(SUM(CASE WHEN cached THEN 1 ELSE 0 END), 0) as cached_requests,
			COALESCE(SUM(CASE WHEN status_code >= 400 THEN 1 ELSE 0 END), 0) as error_requests
		FROM request_logs
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	// Add filters
	if clientID != "" {
		query += fmt.Sprintf(" AND client_id = (SELECT id FROM oauth_clients WHERE client_id = $%d)", argCount)
		args = append(args, clientID)
		argCount++
	}

	if model != "" {
		query += fmt.Sprintf(" AND model = $%d", argCount)
		args = append(args, model)
		argCount++
	}

	if !startTime.IsZero() {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, startTime)
		argCount++
	}

	if !endTime.IsZero() {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, endTime)
		argCount++
	}

	stats := &RequestLogStats{}

	err := r.db.Pool.QueryRow(ctx, query, args...).Scan(
		&stats.TotalRequests,
		&stats.TotalTokens,
		&stats.TotalCost,
		&stats.AvgDuration,
		&stats.CachedRequests,
		&stats.ErrorRequests,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// Calculate cache hit rate
	if stats.TotalRequests > 0 {
		stats.CacheHitRate = float64(stats.CachedRequests) / float64(stats.TotalRequests) * 100.0
	}

	return stats, nil
}
