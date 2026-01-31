package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
)

// ProviderModel represents a model configuration for a provider
type ProviderModel struct {
	ID           uuid.UUID
	ProviderID   string
	ModelID      string
	ModelName    string
	Enabled      bool
	Description  *string
	Capabilities map[string]interface{}
	Pricing      map[string]interface{}
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ProviderModelRepository handles provider model database operations
type ProviderModelRepository struct {
	db *database.DB
}

// NewProviderModelRepository creates a new provider model repository
func NewProviderModelRepository(db *database.DB) *ProviderModelRepository {
	return &ProviderModelRepository{db: db}
}

// Create creates a new provider model entry
func (r *ProviderModelRepository) Create(ctx context.Context, model *ProviderModel) error {
	query := `
		INSERT INTO provider_models (
			id, provider_id, model_id, model_name, enabled, 
			description, capabilities, pricing
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (provider_id, model_id) 
		DO UPDATE SET
			model_name = EXCLUDED.model_name,
			description = EXCLUDED.description,
			capabilities = EXCLUDED.capabilities,
			pricing = EXCLUDED.pricing,
			updated_at = CURRENT_TIMESTAMP
		RETURNING created_at, updated_at
	`

	return r.db.QueryRow(
		ctx,
		query,
		model.ID,
		model.ProviderID,
		model.ModelID,
		model.ModelName,
		model.Enabled,
		model.Description,
		model.Capabilities,
		model.Pricing,
	).Scan(&model.CreatedAt, &model.UpdatedAt)
}

// GetByProvider retrieves all models for a provider
func (r *ProviderModelRepository) GetByProvider(ctx context.Context, providerID string) ([]*ProviderModel, error) {
	query := `
		SELECT id, provider_id, model_id, model_name, enabled,
		       description, capabilities, pricing, created_at, updated_at
		FROM provider_models
		WHERE provider_id = $1
		ORDER BY model_name ASC
	`

	rows, err := r.db.Query(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*ProviderModel
	for rows.Next() {
		model := &ProviderModel{}
		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.ModelID,
			&model.ModelName,
			&model.Enabled,
			&model.Description,
			&model.Capabilities,
			&model.Pricing,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, rows.Err()
}

// GetEnabledByProvider retrieves all enabled models for a provider
func (r *ProviderModelRepository) GetEnabledByProvider(ctx context.Context, providerID string) ([]*ProviderModel, error) {
	query := `
		SELECT id, provider_id, model_id, model_name, enabled,
		       description, capabilities, pricing, created_at, updated_at
		FROM provider_models
		WHERE provider_id = $1 AND enabled = true
		ORDER BY model_name ASC
	`

	rows, err := r.db.Query(ctx, query, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var models []*ProviderModel
	for rows.Next() {
		model := &ProviderModel{}
		err := rows.Scan(
			&model.ID,
			&model.ProviderID,
			&model.ModelID,
			&model.ModelName,
			&model.Enabled,
			&model.Description,
			&model.Capabilities,
			&model.Pricing,
			&model.CreatedAt,
			&model.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		models = append(models, model)
	}

	return models, rows.Err()
}

// GetByProviderAndModel retrieves a specific model
func (r *ProviderModelRepository) GetByProviderAndModel(ctx context.Context, providerID, modelID string) (*ProviderModel, error) {
	query := `
		SELECT id, provider_id, model_id, model_name, enabled,
		       description, capabilities, pricing, created_at, updated_at
		FROM provider_models
		WHERE provider_id = $1 AND model_id = $2
	`

	model := &ProviderModel{}
	err := r.db.QueryRow(ctx, query, providerID, modelID).Scan(
		&model.ID,
		&model.ProviderID,
		&model.ModelID,
		&model.ModelName,
		&model.Enabled,
		&model.Description,
		&model.Capabilities,
		&model.Pricing,
		&model.CreatedAt,
		&model.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return model, nil
}

// UpdateEnabled updates the enabled status of a model
func (r *ProviderModelRepository) UpdateEnabled(ctx context.Context, providerID, modelID string, enabled bool) error {
	query := `
		UPDATE provider_models
		SET enabled = $1, updated_at = CURRENT_TIMESTAMP
		WHERE provider_id = $2 AND model_id = $3
	`

	err := r.db.Exec(ctx, query, enabled, providerID, modelID)
	return err
}

// BulkUpdateEnabled updates the enabled status for multiple models
func (r *ProviderModelRepository) BulkUpdateEnabled(ctx context.Context, providerID string, modelIDs []string, enabled bool) error {
	query := `
		UPDATE provider_models
		SET enabled = $1, updated_at = CURRENT_TIMESTAMP
		WHERE provider_id = $2 AND model_id = ANY($3)
	`

	err := r.db.Exec(ctx, query, enabled, providerID, modelIDs)
	return err
}

// Delete deletes a provider model entry
func (r *ProviderModelRepository) Delete(ctx context.Context, providerID, modelID string) error {
	query := `
		DELETE FROM provider_models
		WHERE provider_id = $1 AND model_id = $2
	`

	err := r.db.Exec(ctx, query, providerID, modelID)
	return err
}

// DeleteByProvider deletes all models for a provider
func (r *ProviderModelRepository) DeleteByProvider(ctx context.Context, providerID string) error {
	query := `
		DELETE FROM provider_models
		WHERE provider_id = $1
	`

	err := r.db.Exec(ctx, query, providerID)
	return err
}

// IsModelEnabled checks if a specific model is enabled
func (r *ProviderModelRepository) IsModelEnabled(ctx context.Context, providerID, modelID string) (bool, error) {
	query := `
		SELECT enabled
		FROM provider_models
		WHERE provider_id = $1 AND model_id = $2
	`

	var enabled bool
	err := r.db.QueryRow(ctx, query, providerID, modelID).Scan(&enabled)
	if err != nil {
		// If model not found in DB, assume it's enabled (backward compatibility)
		return true, nil
	}

	return enabled, nil
}
