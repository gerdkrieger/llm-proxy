package repositories

import (
	"context"

	"github.com/llm-proxy/llm-proxy/internal/infrastructure/providers"
)

// ProviderAPIKeyDBAdapter adapts ProviderAPIKeyRepository to the providers.DBKeyProvider interface.
type ProviderAPIKeyDBAdapter struct {
	repo *ProviderAPIKeyRepository
}

// NewProviderAPIKeyDBAdapter creates a new adapter.
func NewProviderAPIKeyDBAdapter(repo *ProviderAPIKeyRepository) *ProviderAPIKeyDBAdapter {
	return &ProviderAPIKeyDBAdapter{repo: repo}
}

// GetDecryptedKeysByProvider implements providers.DBKeyProvider.
func (a *ProviderAPIKeyDBAdapter) GetDecryptedKeysByProvider(ctx context.Context, providerID string) ([]providers.DBProviderKey, error) {
	keys, err := a.repo.GetDecryptedKeysByProvider(ctx, providerID)
	if err != nil {
		return nil, err
	}

	result := make([]providers.DBProviderKey, len(keys))
	for i, k := range keys {
		result[i] = providers.DBProviderKey{
			APIKey: k.APIKey,
			Weight: k.Weight,
			MaxRPM: k.MaxRPM,
		}
	}
	return result, nil
}
