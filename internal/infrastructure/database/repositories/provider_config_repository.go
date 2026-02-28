// Package repositories provides provider configuration repository.
package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// ProviderConfig represents a provider configuration in the database
type ProviderConfig struct {
	ID              uuid.UUID              `json:"id"`
	ProviderID      string                 `json:"provider_id"`
	ProviderName    string                 `json:"provider_name"`
	ProviderType    string                 `json:"provider_type"`
	Config          map[string]interface{} `json:"config"`
	Enabled         bool                   `json:"enabled"`
	HealthStatus    string                 `json:"health_status"`
	LastHealthCheck *time.Time             `json:"last_health_check,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// ProviderConfigRepository handles provider config database operations
type ProviderConfigRepository struct {
	db *database.DB
}

// NewProviderConfigRepository creates a new provider config repository
func NewProviderConfigRepository(db *database.DB) *ProviderConfigRepository {
	return &ProviderConfigRepository{db: db}
}

// GetByProviderID retrieves a provider config by its ID
func (r *ProviderConfigRepository) GetByProviderID(ctx context.Context, providerID string) (*ProviderConfig, error) {
	query := `
		SELECT 
			id, provider_id, provider_name, provider_type, config, enabled,
			health_status, last_health_check, created_at, updated_at
		FROM provider_configs
		WHERE provider_id = $1
	`

	config := &ProviderConfig{}
	var configJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, providerID).Scan(
		&config.ID,
		&config.ProviderID,
		&config.ProviderName,
		&config.ProviderType,
		&configJSON,
		&config.Enabled,
		&config.HealthStatus,
		&config.LastHealthCheck,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse config JSON
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &config.Config); err != nil {
			return nil, fmt.Errorf("failed to parse config JSON: %w", err)
		}
	}

	return config, nil
}

// ListAll retrieves all provider configs
func (r *ProviderConfigRepository) ListAll(ctx context.Context) ([]*ProviderConfig, error) {
	query := `
		SELECT 
			id, provider_id, provider_name, provider_type, config, enabled,
			health_status, last_health_check, created_at, updated_at
		FROM provider_configs
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider configs: %w", err)
	}
	defer rows.Close()

	configs := make([]*ProviderConfig, 0)
	for rows.Next() {
		config := &ProviderConfig{}
		var configJSON []byte

		err := rows.Scan(
			&config.ID,
			&config.ProviderID,
			&config.ProviderName,
			&config.ProviderType,
			&configJSON,
			&config.Enabled,
			&config.HealthStatus,
			&config.LastHealthCheck,
			&config.CreatedAt,
			&config.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider config: %w", err)
		}

		// Parse config JSON
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &config.Config); err != nil {
				return nil, fmt.Errorf("failed to parse config JSON: %w", err)
			}
		}

		configs = append(configs, config)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provider configs: %w", err)
	}

	return configs, nil
}

// Create inserts a new provider config
func (r *ProviderConfigRepository) Create(ctx context.Context, config *ProviderConfig) error {
	query := `
		INSERT INTO provider_configs (
			id, provider_id, provider_name, provider_type, config, enabled, health_status, api_key_encrypted
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at
	`

	// Marshal config to JSON
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Generate UUID if not set
	if config.ID == uuid.Nil {
		config.ID = uuid.New()
	}

	// For custom providers without separate API key management, use empty string
	// (api_key_encrypted is NOT NULL in schema, so we need a placeholder)
	apiKeyEncrypted := ""

	err = r.db.Pool.QueryRow(
		ctx, query,
		config.ID,
		config.ProviderID,
		config.ProviderName,
		config.ProviderType,
		configJSON,
		config.Enabled,
		config.HealthStatus,
		apiKeyEncrypted,
	).Scan(&config.CreatedAt, &config.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create provider config: %w", err)
	}

	return nil
}

// Update updates an existing provider config
func (r *ProviderConfigRepository) Update(ctx context.Context, config *ProviderConfig) error {
	query := `
		UPDATE provider_configs
		SET 
			provider_name = $1,
			provider_type = $2,
			config = $3,
			enabled = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE provider_id = $5
		RETURNING updated_at
	`

	// Marshal config to JSON
	configJSON, err := json.Marshal(config.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = r.db.Pool.QueryRow(
		ctx, query,
		config.ProviderName,
		config.ProviderType,
		configJSON,
		config.Enabled,
		config.ProviderID,
	).Scan(&config.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to update provider config: %w", err)
	}

	return nil
}

// Delete removes a provider config
func (r *ProviderConfigRepository) Delete(ctx context.Context, providerID string) error {
	query := `DELETE FROM provider_configs WHERE provider_id = $1`

	result, err := r.db.Pool.Exec(ctx, query, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete provider config: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %s", providerID)
	}

	return nil
}

// SetEnabled updates the enabled status of a provider
func (r *ProviderConfigRepository) SetEnabled(ctx context.Context, providerID string, enabled bool) error {
	query := `
		UPDATE provider_configs
		SET enabled = $1, updated_at = CURRENT_TIMESTAMP
		WHERE provider_id = $2
	`

	result, err := r.db.Pool.Exec(ctx, query, enabled, providerID)
	if err != nil {
		return fmt.Errorf("failed to update provider enabled status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %s", providerID)
	}

	return nil
}

// UpdateHealthStatus updates the health status and timestamp
func (r *ProviderConfigRepository) UpdateHealthStatus(ctx context.Context, providerID, status string) error {
	query := `
		UPDATE provider_configs
		SET 
			health_status = $1,
			last_health_check = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
		WHERE provider_id = $2
	`

	_, err := r.db.Pool.Exec(ctx, query, status, providerID)
	if err != nil {
		return fmt.Errorf("failed to update health status: %w", err)
	}

	return nil
}
