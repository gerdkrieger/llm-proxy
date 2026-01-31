// Package repositories provides provider settings repository.
package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// ProviderSetting represents provider runtime settings
type ProviderSetting struct {
	ID             int
	ProviderID     string
	ProviderName   string
	ProviderType   string
	Enabled        bool
	Config         map[string]interface{}
	LastTestAt     *time.Time
	LastTestStatus *string
	LastTestError  *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ProviderSettingsRepository handles provider settings database operations
type ProviderSettingsRepository struct {
	db *database.DB
}

// NewProviderSettingsRepository creates a new provider settings repository
func NewProviderSettingsRepository(db *database.DB) *ProviderSettingsRepository {
	return &ProviderSettingsRepository{db: db}
}

// GetByProviderID retrieves settings for a specific provider
func (r *ProviderSettingsRepository) GetByProviderID(ctx context.Context, providerID string) (*ProviderSetting, error) {
	query := `
		SELECT 
			id, provider_id, provider_name, provider_type, enabled,
			config, last_test_at, last_test_status, last_test_error,
			created_at, updated_at
		FROM provider_settings
		WHERE provider_id = $1
	`

	setting := &ProviderSetting{}
	var configJSON []byte

	err := r.db.Pool.QueryRow(ctx, query, providerID).Scan(
		&setting.ID,
		&setting.ProviderID,
		&setting.ProviderName,
		&setting.ProviderType,
		&setting.Enabled,
		&configJSON,
		&setting.LastTestAt,
		&setting.LastTestStatus,
		&setting.LastTestError,
		&setting.CreatedAt,
		&setting.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse config JSON
	if len(configJSON) > 0 {
		if err := json.Unmarshal(configJSON, &setting.Config); err != nil {
			return nil, fmt.Errorf("failed to parse config JSON: %w", err)
		}
	}

	return setting, nil
}

// Upsert creates or updates provider settings
func (r *ProviderSettingsRepository) Upsert(ctx context.Context, setting *ProviderSetting) error {
	query := `
		INSERT INTO provider_settings (
			provider_id, provider_name, provider_type, enabled, config
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider_id)
		DO UPDATE SET
			provider_name = EXCLUDED.provider_name,
			provider_type = EXCLUDED.provider_type,
			enabled = EXCLUDED.enabled,
			config = EXCLUDED.config,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`

	// Marshal config to JSON
	configJSON, err := json.Marshal(setting.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = r.db.Pool.QueryRow(
		ctx, query,
		setting.ProviderID,
		setting.ProviderName,
		setting.ProviderType,
		setting.Enabled,
		configJSON,
	).Scan(&setting.ID, &setting.CreatedAt, &setting.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to upsert provider setting: %w", err)
	}

	return nil
}

// SetEnabled updates the enabled status of a provider
func (r *ProviderSettingsRepository) SetEnabled(ctx context.Context, providerID string, enabled bool) error {
	query := `
		UPDATE provider_settings
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

// UpdateTestStatus updates the last test status and timestamp
func (r *ProviderSettingsRepository) UpdateTestStatus(ctx context.Context, providerID, status string, errorMsg *string) error {
	query := `
		UPDATE provider_settings
		SET 
			last_test_at = CURRENT_TIMESTAMP,
			last_test_status = $1,
			last_test_error = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE provider_id = $3
	`

	_, err := r.db.Pool.Exec(ctx, query, status, errorMsg, providerID)
	if err != nil {
		return fmt.Errorf("failed to update test status: %w", err)
	}

	return nil
}

// ListAll retrieves all provider settings
func (r *ProviderSettingsRepository) ListAll(ctx context.Context) ([]*ProviderSetting, error) {
	query := `
		SELECT 
			id, provider_id, provider_name, provider_type, enabled,
			config, last_test_at, last_test_status, last_test_error,
			created_at, updated_at
		FROM provider_settings
		ORDER BY provider_id
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider settings: %w", err)
	}
	defer rows.Close()

	settings := make([]*ProviderSetting, 0)
	for rows.Next() {
		setting := &ProviderSetting{}
		var configJSON []byte

		err := rows.Scan(
			&setting.ID,
			&setting.ProviderID,
			&setting.ProviderName,
			&setting.ProviderType,
			&setting.Enabled,
			&configJSON,
			&setting.LastTestAt,
			&setting.LastTestStatus,
			&setting.LastTestError,
			&setting.CreatedAt,
			&setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider setting: %w", err)
		}

		// Parse config JSON
		if len(configJSON) > 0 {
			if err := json.Unmarshal(configJSON, &setting.Config); err != nil {
				return nil, fmt.Errorf("failed to parse config JSON: %w", err)
			}
		}

		settings = append(settings, setting)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provider settings: %w", err)
	}

	return settings, nil
}

// Delete removes a provider setting
func (r *ProviderSettingsRepository) Delete(ctx context.Context, providerID string) error {
	query := `DELETE FROM provider_settings WHERE provider_id = $1`

	result, err := r.db.Pool.Exec(ctx, query, providerID)
	if err != nil {
		return fmt.Errorf("failed to delete provider setting: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("provider not found: %s", providerID)
	}

	return nil
}
