-- Add provider_models table for model-level configuration
CREATE TABLE IF NOT EXISTS provider_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id VARCHAR(50) NOT NULL,
    model_id VARCHAR(255) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    capabilities JSONB DEFAULT '{}',
    pricing JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, model_id)
);

-- Index for faster lookups
CREATE INDEX idx_provider_models_provider_id ON provider_models(provider_id);
CREATE INDEX idx_provider_models_enabled ON provider_models(enabled);
CREATE INDEX idx_provider_models_provider_enabled ON provider_models(provider_id, enabled);

-- Comments
COMMENT ON TABLE provider_models IS 'Configuration for individual models per provider';
COMMENT ON COLUMN provider_models.provider_id IS 'Provider identifier (claude, openai, etc.)';
COMMENT ON COLUMN provider_models.model_id IS 'Model identifier from provider API';
COMMENT ON COLUMN provider_models.model_name IS 'Human-readable model name';
COMMENT ON COLUMN provider_models.enabled IS 'Whether this model is available for use';
COMMENT ON COLUMN provider_models.capabilities IS 'JSON object with model capabilities (vision, function_calling, etc.)';
COMMENT ON COLUMN provider_models.pricing IS 'JSON object with pricing information (input_price, output_price per 1M tokens)';
