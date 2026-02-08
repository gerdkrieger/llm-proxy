// Package repositories provides the provider API key repository.
package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/llm-proxy/llm-proxy/internal/infrastructure/database"
	"github.com/llm-proxy/llm-proxy/pkg/crypto"
)

// ProviderAPIKey represents an encrypted LLM provider API key stored in the database.
type ProviderAPIKey struct {
	ID           uuid.UUID `json:"id"`
	ProviderID   string    `json:"provider_id"`
	KeyName      string    `json:"key_name"`
	EncryptedKey string    `json:"-"`        // Never exposed via JSON
	KeyHint      string    `json:"key_hint"` // e.g. "...r5PZ"
	Weight       int       `json:"weight"`
	MaxRPM       int       `json:"max_rpm"`
	Enabled      bool      `json:"enabled"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ProviderAPIKeyRepository handles provider API key database operations
// with application-level AES-256-GCM encryption.
type ProviderAPIKeyRepository struct {
	db        *database.DB
	encryptor *crypto.KeyEncryptor
}

// NewProviderAPIKeyRepository creates a new provider API key repository.
func NewProviderAPIKeyRepository(db *database.DB, encryptor *crypto.KeyEncryptor) *ProviderAPIKeyRepository {
	return &ProviderAPIKeyRepository{db: db, encryptor: encryptor}
}

// Create stores a new API key (encrypting it first).
func (r *ProviderAPIKeyRepository) Create(ctx context.Context, providerID, keyName, plaintextKey string, weight, maxRPM int) (*ProviderAPIKey, error) {
	encrypted, err := r.encryptor.Encrypt(plaintextKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API key: %w", err)
	}

	// Build hint from last 4 chars
	hint := ""
	if len(plaintextKey) > 4 {
		hint = "..." + plaintextKey[len(plaintextKey)-4:]
	}

	id := uuid.New()
	query := `
		INSERT INTO provider_api_keys (id, provider_id, key_name, encrypted_key, key_hint, weight, max_rpm, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
		RETURNING created_at, updated_at
	`

	key := &ProviderAPIKey{
		ID:         id,
		ProviderID: providerID,
		KeyName:    keyName,
		KeyHint:    hint,
		Weight:     weight,
		MaxRPM:     maxRPM,
		Enabled:    true,
	}

	err = r.db.Pool.QueryRow(ctx, query, id, providerID, keyName, encrypted, hint, weight, maxRPM).
		Scan(&key.CreatedAt, &key.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert provider API key: %w", err)
	}

	return key, nil
}

// ListByProvider returns all API keys for a provider (without decrypted values).
func (r *ProviderAPIKeyRepository) ListByProvider(ctx context.Context, providerID string) ([]*ProviderAPIKey, error) {
	query := `
		SELECT id, provider_id, key_name, key_hint, weight, max_rpm, enabled, created_at, updated_at
		FROM provider_api_keys
		WHERE provider_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Pool.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list provider API keys: %w", err)
	}
	defer rows.Close()

	var keys []*ProviderAPIKey
	for rows.Next() {
		key := &ProviderAPIKey{}
		err := rows.Scan(&key.ID, &key.ProviderID, &key.KeyName, &key.KeyHint,
			&key.Weight, &key.MaxRPM, &key.Enabled, &key.CreatedAt, &key.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan provider API key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, rows.Err()
}

// GetDecryptedKeysByProvider returns all enabled API keys for a provider with decrypted values.
// This should ONLY be used internally by the ProviderManager for initialization.
func (r *ProviderAPIKeyRepository) GetDecryptedKeysByProvider(ctx context.Context, providerID string) ([]DecryptedProviderKey, error) {
	query := `
		SELECT id, key_name, encrypted_key, weight, max_rpm
		FROM provider_api_keys
		WHERE provider_id = $1 AND enabled = TRUE
		ORDER BY weight DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query provider API keys: %w", err)
	}
	defer rows.Close()

	var keys []DecryptedProviderKey
	for rows.Next() {
		var id uuid.UUID
		var keyName, encryptedKey string
		var weight, maxRPM int

		if err := rows.Scan(&id, &keyName, &encryptedKey, &weight, &maxRPM); err != nil {
			return nil, fmt.Errorf("failed to scan: %w", err)
		}

		plaintext, err := r.encryptor.Decrypt(encryptedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt key %s: %w", keyName, err)
		}

		keys = append(keys, DecryptedProviderKey{
			ID:      id,
			KeyName: keyName,
			APIKey:  plaintext,
			Weight:  weight,
			MaxRPM:  maxRPM,
		})
	}

	return keys, rows.Err()
}

// DecryptedProviderKey holds a decrypted API key for internal use only.
type DecryptedProviderKey struct {
	ID      uuid.UUID
	KeyName string
	APIKey  string // Plaintext — handle with care
	Weight  int
	MaxRPM  int
}

// Delete removes an API key by ID.
func (r *ProviderAPIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Pool.Exec(ctx, `DELETE FROM provider_api_keys WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete provider API key: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("API key not found")
	}
	return nil
}

// SetEnabled enables or disables an API key.
func (r *ProviderAPIKeyRepository) SetEnabled(ctx context.Context, id uuid.UUID, enabled bool) error {
	result, err := r.db.Pool.Exec(ctx,
		`UPDATE provider_api_keys SET enabled = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`,
		enabled, id)
	if err != nil {
		return fmt.Errorf("failed to update API key enabled status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("API key not found")
	}
	return nil
}

// CountByProvider returns the number of enabled keys for a provider.
func (r *ProviderAPIKeyRepository) CountByProvider(ctx context.Context, providerID string) (int, error) {
	var count int
	err := r.db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM provider_api_keys WHERE provider_id = $1 AND enabled = TRUE`,
		providerID).Scan(&count)
	return count, err
}
