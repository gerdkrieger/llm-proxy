// Package repositories provides filter match logging repository.
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// FilterMatch represents a logged filter match event
type FilterMatch struct {
	ID          uuid.UUID
	RequestID   string
	ClientID    *uuid.UUID
	FilterID    *int // NULL for attachment redactions
	Model       string
	Provider    string
	Pattern     string
	Replacement string
	FilterType  string
	MatchCount  int
	MatchedText *string
	IPAddress   *string
	UserAgent   *string
	CreatedAt   time.Time
}

// FilterMatchRepository handles filter match database operations
type FilterMatchRepository struct {
	db *database.DB
}

// NewFilterMatchRepository creates a new filter match repository
func NewFilterMatchRepository(db *database.DB) *FilterMatchRepository {
	return &FilterMatchRepository{db: db}
}

// Create creates a new filter match log entry
func (r *FilterMatchRepository) Create(ctx context.Context, match *FilterMatch) error {
	query := `
		INSERT INTO filter_matches (
			id, request_id, client_id, filter_id, model, provider,
			pattern, replacement, filter_type, match_count, matched_text,
			ip_address, user_agent
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at
	`

	err := r.db.Pool.QueryRow(
		ctx, query,
		match.ID,
		match.RequestID,
		match.ClientID,
		match.FilterID,
		match.Model,
		match.Provider,
		match.Pattern,
		match.Replacement,
		match.FilterType,
		match.MatchCount,
		match.MatchedText,
		match.IPAddress,
		match.UserAgent,
	).Scan(&match.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create filter match: %w", err)
	}

	return nil
}

// GetRecentMatches retrieves recent filter matches
func (r *FilterMatchRepository) GetRecentMatches(ctx context.Context, limit int) ([]*FilterMatch, error) {
	query := `
		SELECT 
			fm.id, fm.request_id, fm.client_id, fm.filter_id,
			fm.model, fm.provider, fm.pattern, fm.replacement,
			fm.filter_type, fm.match_count, fm.matched_text,
			fm.ip_address::text, fm.user_agent, fm.created_at
		FROM filter_matches fm
		ORDER BY fm.created_at DESC
		LIMIT $1
	`

	rows, err := r.db.Pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query filter matches: %w", err)
	}
	defer rows.Close()

	matches := make([]*FilterMatch, 0, limit)
	for rows.Next() {
		match := &FilterMatch{}
		err := rows.Scan(
			&match.ID,
			&match.RequestID,
			&match.ClientID,
			&match.FilterID,
			&match.Model,
			&match.Provider,
			&match.Pattern,
			&match.Replacement,
			&match.FilterType,
			&match.MatchCount,
			&match.MatchedText,
			&match.IPAddress,
			&match.UserAgent,
			&match.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan filter match: %w", err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating filter matches: %w", err)
	}

	return matches, nil
}

// GetMatchesByClient retrieves filter matches for a specific client
func (r *FilterMatchRepository) GetMatchesByClient(ctx context.Context, clientID uuid.UUID, limit int) ([]*FilterMatch, error) {
	query := `
		SELECT 
			fm.id, fm.request_id, fm.client_id, fm.filter_id,
			fm.model, fm.provider, fm.pattern, fm.replacement,
			fm.filter_type, fm.match_count, fm.matched_text,
			fm.ip_address::text, fm.user_agent, fm.created_at
		FROM filter_matches fm
		WHERE fm.client_id = $1
		ORDER BY fm.created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, clientID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query filter matches by client: %w", err)
	}
	defer rows.Close()

	matches := make([]*FilterMatch, 0, limit)
	for rows.Next() {
		match := &FilterMatch{}
		err := rows.Scan(
			&match.ID,
			&match.RequestID,
			&match.ClientID,
			&match.FilterID,
			&match.Model,
			&match.Provider,
			&match.Pattern,
			&match.Replacement,
			&match.FilterType,
			&match.MatchCount,
			&match.MatchedText,
			&match.IPAddress,
			&match.UserAgent,
			&match.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan filter match: %w", err)
		}
		matches = append(matches, match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating filter matches: %w", err)
	}

	return matches, nil
}

// GetTotalMatches returns the total count of filter matches
func (r *FilterMatchRepository) GetTotalMatches(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM filter_matches`

	var count int
	err := r.db.Pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count filter matches: %w", err)
	}

	return count, nil
}
