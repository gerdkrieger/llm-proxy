// Package repositories provides data access layer for database entities.
package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
	"golang.org/x/crypto/bcrypt"
)

// OAuthClient represents an OAuth 2.0 client (now used as generic API client)
type OAuthClient struct {
	ID            uuid.UUID
	ClientID      string
	ClientSecret  string // This will be hashed
	Name          string
	RedirectURIs  []string
	GrantTypes    []string
	DefaultScope  string
	AllowedModels []string // NULL = all models allowed, empty array = no models, specific array = only those models
	RateLimitRPM  *int     // NULL = unlimited
	RateLimitRPD  *int
	Enabled       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// OAuthClientRepository handles OAuth client database operations
type OAuthClientRepository struct {
	db *database.DB
}

// NewOAuthClientRepository creates a new OAuth client repository
func NewOAuthClientRepository(db *database.DB) *OAuthClientRepository {
	return &OAuthClientRepository{db: db}
}

// Create creates a new OAuth client
func (r *OAuthClientRepository) Create(ctx context.Context, client *OAuthClient) error {
	// Hash the client secret
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(client.ClientSecret), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash client secret: %w", err)
	}

	query := `
		INSERT INTO oauth_clients (
			id, client_id, client_secret_hash, name, redirect_uris, grant_types, 
			default_scope, allowed_models, rate_limit_rpm, rate_limit_rpd, enabled
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at
	`

	// Convert allowed_models to JSONB (nil stays nil, empty slice becomes [], populated slice becomes JSON array)
	var allowedModelsJSON interface{}
	if client.AllowedModels != nil {
		allowedModelsJSON = client.AllowedModels
	}

	err = r.db.Pool.QueryRow(
		ctx, query,
		client.ID,
		client.ClientID,
		string(hashedSecret),
		client.Name,
		client.RedirectURIs,
		client.GrantTypes,
		client.DefaultScope,
		allowedModelsJSON,
		client.RateLimitRPM,
		client.RateLimitRPD,
		client.Enabled,
	).Scan(&client.CreatedAt, &client.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create OAuth client: %w", err)
	}

	return nil
}

// GetByClientID retrieves an OAuth client by client ID
func (r *OAuthClientRepository) GetByClientID(ctx context.Context, clientID string) (*OAuthClient, error) {
	query := `
		SELECT 
			id, client_id, client_secret_hash, name, redirect_uris, grant_types,
			default_scope, allowed_models, rate_limit_rpm, rate_limit_rpd, enabled, created_at, updated_at
		FROM oauth_clients
		WHERE client_id = $1
	`

	client := &OAuthClient{}
	var secretHash string
	var allowedModelsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, clientID).Scan(
		&client.ID,
		&client.ClientID,
		&secretHash,
		&client.Name,
		&client.RedirectURIs,
		&client.GrantTypes,
		&client.DefaultScope,
		&allowedModelsJSON,
		&client.RateLimitRPM,
		&client.RateLimitRPD,
		&client.Enabled,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	// Parse allowed_models JSON if present
	if len(allowedModelsJSON) > 0 {
		// unmarshal JSON array into []string
		var models []string
		if err := json.Unmarshal(allowedModelsJSON, &models); err == nil {
			client.AllowedModels = models
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth client: %w", err)
	}

	client.ClientSecret = secretHash // Store hash for verification

	return client, nil
}

// ValidateSecret validates the client secret against the stored hash
func (r *OAuthClientRepository) ValidateSecret(client *OAuthClient, secret string) bool {
	hash := strings.TrimSpace(client.ClientSecret)
	key := strings.TrimSpace(secret)
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(key))
	return err == nil
}

// GetByID retrieves an OAuth client by ID
func (r *OAuthClientRepository) GetByID(ctx context.Context, id uuid.UUID) (*OAuthClient, error) {
	query := `
		SELECT 
			id, client_id, client_secret_hash, name, redirect_uris, grant_types,
			default_scope, allowed_models, rate_limit_rpm, rate_limit_rpd, enabled, created_at, updated_at
		FROM oauth_clients
		WHERE id = $1
	`

	client := &OAuthClient{}
	var secretHash string
	var allowedModelsJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&client.ID,
		&client.ClientID,
		&secretHash,
		&client.Name,
		&client.RedirectURIs,
		&client.GrantTypes,
		&client.DefaultScope,
		&allowedModelsJSON,
		&client.RateLimitRPM,
		&client.RateLimitRPD,
		&client.Enabled,
		&client.CreatedAt,
		&client.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth client: %w", err)
	}

	// Parse allowed_models JSON if present
	if len(allowedModelsJSON) > 0 {
		var models []string
		if err := json.Unmarshal(allowedModelsJSON, &models); err == nil {
			client.AllowedModels = models
		}
	}

	client.ClientSecret = secretHash

	return client, nil
}

// List retrieves all OAuth clients
func (r *OAuthClientRepository) List(ctx context.Context, limit, offset int) ([]*OAuthClient, error) {
	query := `
		SELECT 
			id, client_id, name, redirect_uris, grant_types,
			default_scope, allowed_models, rate_limit_rpm, rate_limit_rpd, enabled, created_at, updated_at
		FROM oauth_clients
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list OAuth clients: %w", err)
	}
	defer rows.Close()

	var clients []*OAuthClient
	for rows.Next() {
		client := &OAuthClient{}
		var allowedModelsJSON []byte
		err := rows.Scan(
			&client.ID,
			&client.ClientID,
			&client.Name,
			&client.RedirectURIs,
			&client.GrantTypes,
			&client.DefaultScope,
			&allowedModelsJSON,
			&client.RateLimitRPM,
			&client.RateLimitRPD,
			&client.Enabled,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth client: %w", err)
		}

		// Parse allowed_models JSON if present
		if len(allowedModelsJSON) > 0 {
			var models []string
			if err := json.Unmarshal(allowedModelsJSON, &models); err == nil {
				client.AllowedModels = models
			}
		}

		clients = append(clients, client)
	}

	return clients, nil
}

// ListWithSecrets retrieves all OAuth clients including their secret hashes.
// Used internally for authentication validation only - never expose hashes via API.
func (r *OAuthClientRepository) ListWithSecrets(ctx context.Context) ([]*OAuthClient, error) {
	query := `
		SELECT 
			id, client_id, client_secret_hash, name, redirect_uris, grant_types,
			default_scope, allowed_models, rate_limit_rpm, rate_limit_rpd, enabled, created_at, updated_at
		FROM oauth_clients
		WHERE enabled = true
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list OAuth clients: %w", err)
	}
	defer rows.Close()

	var clients []*OAuthClient
	for rows.Next() {
		client := &OAuthClient{}
		var secretHash string
		var allowedModelsJSON []byte
		err := rows.Scan(
			&client.ID,
			&client.ClientID,
			&secretHash,
			&client.Name,
			&client.RedirectURIs,
			&client.GrantTypes,
			&client.DefaultScope,
			&allowedModelsJSON,
			&client.RateLimitRPM,
			&client.RateLimitRPD,
			&client.Enabled,
			&client.CreatedAt,
			&client.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan OAuth client: %w", err)
		}

		client.ClientSecret = secretHash

		if len(allowedModelsJSON) > 0 {
			var models []string
			if err := json.Unmarshal(allowedModelsJSON, &models); err == nil {
				client.AllowedModels = models
			}
		}

		clients = append(clients, client)
	}

	return clients, nil
}

// Update updates an OAuth client
func (r *OAuthClientRepository) Update(ctx context.Context, client *OAuthClient) error {
	query := `
		UPDATE oauth_clients
		SET name = $1, redirect_uris = $2, grant_types = $3, default_scope = $4,
		    allowed_models = $5, rate_limit_rpm = $6, rate_limit_rpd = $7, enabled = $8, updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
		RETURNING updated_at
	`

	// Convert allowed_models to JSONB
	var allowedModelsJSON interface{}
	if client.AllowedModels != nil {
		allowedModelsJSON = client.AllowedModels
	}

	err := r.db.Pool.QueryRow(
		ctx, query,
		client.Name,
		client.RedirectURIs,
		client.GrantTypes,
		client.DefaultScope,
		allowedModelsJSON,
		client.RateLimitRPM,
		client.RateLimitRPD,
		client.Enabled,
		client.ID,
	).Scan(&client.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update OAuth client: %w", err)
	}

	return nil
}

// UpdateSecret updates only the client secret hash
func (r *OAuthClientRepository) UpdateSecret(ctx context.Context, id uuid.UUID, newSecret string) error {
	hashedSecret, err := bcrypt.GenerateFromPassword([]byte(newSecret), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash client secret: %w", err)
	}

	query := `
		UPDATE oauth_clients
		SET client_secret_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`

	result, err := r.db.Pool.Exec(ctx, query, string(hashedSecret), id)
	if err != nil {
		return fmt.Errorf("failed to update client secret: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("client not found")
	}

	return nil
}

// Delete deletes an OAuth client
func (r *OAuthClientRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM oauth_clients WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete OAuth client: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("OAuth client not found")
	}

	return nil
}
