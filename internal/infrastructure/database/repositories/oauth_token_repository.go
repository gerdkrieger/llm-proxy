// Package repositories provides data access for OAuth tokens.
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// OAuthToken represents an OAuth 2.0 access or refresh token
type OAuthToken struct {
	ID           uuid.UUID
	ClientID     uuid.UUID
	AccessToken  string
	RefreshToken *string
	TokenType    string
	ExpiresAt    *time.Time
	Scope        string
	CreatedAt    time.Time
}

// OAuthTokenRepository handles OAuth token database operations
type OAuthTokenRepository struct {
	db *database.DB
}

// NewOAuthTokenRepository creates a new OAuth token repository
func NewOAuthTokenRepository(db *database.DB) *OAuthTokenRepository {
	return &OAuthTokenRepository{db: db}
}

// Create stores a new OAuth token
func (r *OAuthTokenRepository) Create(ctx context.Context, token *OAuthToken) error {
	query := `
		INSERT INTO oauth_tokens (
			id, client_id, access_token, refresh_token, token_type, expires_at, scope
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at
	`

	err := r.db.Pool.QueryRow(
		ctx, query,
		token.ID,
		token.ClientID,
		token.AccessToken,
		token.RefreshToken,
		token.TokenType,
		token.ExpiresAt,
		token.Scope,
	).Scan(&token.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create OAuth token: %w", err)
	}

	return nil
}

// GetByAccessToken retrieves a token by access token string
func (r *OAuthTokenRepository) GetByAccessToken(ctx context.Context, accessToken string) (*OAuthToken, error) {
	query := `
		SELECT 
			id, client_id, access_token, refresh_token, token_type, expires_at, scope, created_at
		FROM oauth_tokens
		WHERE access_token = $1
	`

	token := &OAuthToken{}
	err := r.db.Pool.QueryRow(ctx, query, accessToken).Scan(
		&token.ID,
		&token.ClientID,
		&token.AccessToken,
		&token.RefreshToken,
		&token.TokenType,
		&token.ExpiresAt,
		&token.Scope,
		&token.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %w", err)
	}

	return token, nil
}

// GetByRefreshToken retrieves a token by refresh token string
func (r *OAuthTokenRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*OAuthToken, error) {
	query := `
		SELECT 
			id, client_id, access_token, refresh_token, token_type, expires_at, scope, created_at
		FROM oauth_tokens
		WHERE refresh_token = $1
	`

	token := &OAuthToken{}
	err := r.db.Pool.QueryRow(ctx, query, refreshToken).Scan(
		&token.ID,
		&token.ClientID,
		&token.AccessToken,
		&token.RefreshToken,
		&token.TokenType,
		&token.ExpiresAt,
		&token.Scope,
		&token.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token by refresh token: %w", err)
	}

	return token, nil
}

// DeleteByAccessToken deletes a token by access token
func (r *OAuthTokenRepository) DeleteByAccessToken(ctx context.Context, accessToken string) error {
	query := `DELETE FROM oauth_tokens WHERE access_token = $1`

	result, err := r.db.Pool.Exec(ctx, query, accessToken)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("OAuth token not found")
	}

	return nil
}

// DeleteByRefreshToken deletes a token by refresh token
func (r *OAuthTokenRepository) DeleteByRefreshToken(ctx context.Context, refreshToken string) error {
	query := `DELETE FROM oauth_tokens WHERE refresh_token = $1`

	result, err := r.db.Pool.Exec(ctx, query, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("OAuth token not found")
	}

	return nil
}

// DeleteExpired deletes all expired tokens
func (r *OAuthTokenRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM oauth_tokens WHERE expires_at < CURRENT_TIMESTAMP`

	result, err := r.db.Pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	return result.RowsAffected(), nil
}

// ListByClientID retrieves all tokens for a specific client
func (r *OAuthTokenRepository) ListByClientID(ctx context.Context, clientID uuid.UUID, limit, offset int) ([]*OAuthToken, error) {
	query := `
		SELECT 
			id, client_id, access_token, refresh_token, token_type, expires_at, scope, created_at
		FROM oauth_tokens
		WHERE client_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Pool.Query(ctx, query, clientID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list OAuth tokens: %w", err)
	}
	defer rows.Close()

	var tokens []*OAuthToken
	for rows.Next() {
		token := &OAuthToken{}
		err := rows.Scan(
			&token.ID,
			&token.ClientID,
			&token.AccessToken,
			&token.RefreshToken,
			&token.TokenType,
			&token.ExpiresAt,
			&token.Scope,
			&token.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

// IsExpired checks if a token is expired
func (t *OAuthToken) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false // No expiration
	}
	return time.Now().After(*t.ExpiresAt)
}
