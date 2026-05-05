package repositories

import (
	"context"

	"github.com/llm-proxy/llm-proxy/internal/domain/models"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// ContentFilterRepository handles content filter data operations
type ContentFilterRepository struct {
	db *database.DB
}

// NewContentFilterRepository creates a new repository
func NewContentFilterRepository(db *database.DB) *ContentFilterRepository {
	return &ContentFilterRepository{db: db}
}

// Create creates a new content filter
func (r *ContentFilterRepository) Create(ctx context.Context, filter *models.ContentFilter) error {
	query := `
		INSERT INTO content_filters (
			pattern, replacement, description, filter_type, 
			case_sensitive, enabled, priority, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.Pool.QueryRow(ctx, query,
		filter.Pattern,
		filter.Replacement,
		filter.Description,
		filter.FilterType,
		filter.CaseSensitive,
		filter.Enabled,
		filter.Priority,
		filter.CreatedBy,
	).Scan(&filter.ID, &filter.CreatedAt, &filter.UpdatedAt)
}

// GetByID retrieves a filter by ID
func (r *ContentFilterRepository) GetByID(ctx context.Context, id int) (*models.ContentFilter, error) {
	var filter models.ContentFilter
	query := `SELECT * FROM content_filters WHERE id = $1`
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&filter.ID, &filter.Pattern, &filter.Replacement, &filter.Description,
		&filter.FilterType, &filter.CaseSensitive, &filter.Enabled, &filter.Priority,
		&filter.CreatedAt, &filter.UpdatedAt, &filter.CreatedBy,
		&filter.MatchCount, &filter.LastMatchedAt,
	)
	if err != nil {
		return nil, err
	}
	return &filter, nil
}

// List retrieves all content filters, ordered by priority
func (r *ContentFilterRepository) List(ctx context.Context, enabledOnly bool) ([]*models.ContentFilter, error) {
	query := `SELECT * FROM content_filters`
	if enabledOnly {
		query += ` WHERE enabled = TRUE`
	}
	query += ` ORDER BY priority DESC, id ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var filters []*models.ContentFilter
	for rows.Next() {
		var filter models.ContentFilter
		err := rows.Scan(
			&filter.ID, &filter.Pattern, &filter.Replacement, &filter.Description,
			&filter.FilterType, &filter.CaseSensitive, &filter.Enabled, &filter.Priority,
			&filter.CreatedAt, &filter.UpdatedAt, &filter.CreatedBy,
			&filter.MatchCount, &filter.LastMatchedAt,
		)
		if err != nil {
			return nil, err
		}
		filters = append(filters, &filter)
	}
	return filters, nil
}

// Update updates a content filter
func (r *ContentFilterRepository) Update(ctx context.Context, id int, req *models.UpdateContentFilterRequest) error {
	query := `UPDATE content_filters SET updated_at = NOW()`
	args := []interface{}{}
	argCount := 1

	if req.Pattern != nil {
		query += `, pattern = $` + string(rune(argCount+'0'))
		args = append(args, *req.Pattern)
		argCount++
	}
	if req.Replacement != nil {
		query += `, replacement = $` + string(rune(argCount+'0'))
		args = append(args, *req.Replacement)
		argCount++
	}
	if req.Description != nil {
		query += `, description = $` + string(rune(argCount+'0'))
		args = append(args, *req.Description)
		argCount++
	}
	if req.Enabled != nil {
		query += `, enabled = $` + string(rune(argCount+'0'))
		args = append(args, *req.Enabled)
		argCount++
	}
	if req.Priority != nil {
		query += `, priority = $` + string(rune(argCount+'0'))
		args = append(args, *req.Priority)
		argCount++
	}

	query += ` WHERE id = $` + string(rune(argCount+'0'))
	args = append(args, id)

	_, err := r.db.Pool.Exec(ctx, query, args...)
	return err
}

// Delete deletes a content filter
func (r *ContentFilterRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM content_filters WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id)
	return err
}

// IncrementMatchCount increments the match counter for a filter by 1.
func (r *ContentFilterRepository) IncrementMatchCount(ctx context.Context, id int) error {
	return r.IncrementMatchCountBy(ctx, id, 1)
}

// IncrementMatchCountBy increments the match counter for a filter by a given amount.
func (r *ContentFilterRepository) IncrementMatchCountBy(ctx context.Context, id int, count int) error {
	query := `UPDATE content_filters SET match_count = match_count + $2, last_matched_at = NOW() WHERE id = $1`
	_, err := r.db.Pool.Exec(ctx, query, id, count)
	return err
}
