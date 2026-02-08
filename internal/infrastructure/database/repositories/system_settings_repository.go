// Package repositories provides system settings repository.
package repositories

import (
	"context"
	"time"

	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	Key         string
	Value       string
	Description *string
	UpdatedAt   time.Time
}

// SystemSettingsRepository handles system settings database operations
type SystemSettingsRepository struct {
	db *database.DB
}

// NewSystemSettingsRepository creates a new system settings repository
func NewSystemSettingsRepository(db *database.DB) *SystemSettingsRepository {
	return &SystemSettingsRepository{db: db}
}

// Get retrieves a setting by key
func (r *SystemSettingsRepository) Get(ctx context.Context, key string) (*SystemSetting, error) {
	query := `
		SELECT key, value, description, updated_at
		FROM system_settings
		WHERE key = $1
	`

	var setting SystemSetting
	err := r.db.Pool.QueryRow(ctx, query, key).Scan(
		&setting.Key,
		&setting.Value,
		&setting.Description,
		&setting.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

// GetAll retrieves all settings
func (r *SystemSettingsRepository) GetAll(ctx context.Context) ([]*SystemSetting, error) {
	query := `
		SELECT key, value, description, updated_at
		FROM system_settings
		ORDER BY key
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []*SystemSetting
	for rows.Next() {
		var setting SystemSetting
		if err := rows.Scan(
			&setting.Key,
			&setting.Value,
			&setting.Description,
			&setting.UpdatedAt,
		); err != nil {
			return nil, err
		}
		settings = append(settings, &setting)
	}

	return settings, rows.Err()
}

// Set updates or creates a setting
func (r *SystemSettingsRepository) Set(ctx context.Context, key, value string) error {
	query := `
		INSERT INTO system_settings (key, value)
		VALUES ($1, $2)
		ON CONFLICT (key) DO UPDATE
		SET value = EXCLUDED.value, updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.Pool.Exec(ctx, query, key, value)
	return err
}

// GetBool retrieves a boolean setting (defaults to false if not found or invalid)
func (r *SystemSettingsRepository) GetBool(ctx context.Context, key string) bool {
	setting, err := r.Get(ctx, key)
	if err != nil {
		return false
	}
	return setting.Value == "true" || setting.Value == "1" || setting.Value == "yes"
}
