-- Migration 000012: Add provider_api_keys table for DB-managed LLM provider API keys
-- Keys are encrypted with AES-256-GCM at the application level

CREATE TABLE IF NOT EXISTS provider_api_keys (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id     VARCHAR(50) NOT NULL,           -- 'claude', 'openai'
    key_name        VARCHAR(255) NOT NULL,           -- Human-readable label (e.g. "Production Key 1")
    encrypted_key   TEXT NOT NULL,                    -- AES-256-GCM encrypted API key (base64 encoded)
    key_hint        VARCHAR(20),                      -- Last 4 chars of key for identification (e.g. "...r5PZ")
    weight          INT NOT NULL DEFAULT 1,           -- Load balancing weight
    max_rpm         INT NOT NULL DEFAULT 1000,        -- Max requests per minute
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_provider_api_keys_provider FOREIGN KEY (provider_id)
        REFERENCES provider_configs(provider_id) ON DELETE CASCADE
);

-- Index for quick lookups by provider
CREATE INDEX IF NOT EXISTS idx_provider_api_keys_provider_id ON provider_api_keys(provider_id);
CREATE INDEX IF NOT EXISTS idx_provider_api_keys_enabled ON provider_api_keys(provider_id, enabled);
