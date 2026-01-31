-- Migration: Add provider_settings table for runtime provider management
-- This table stores provider enable/disable state and runtime configuration

CREATE TABLE IF NOT EXISTS provider_settings (
    id SERIAL PRIMARY KEY,
    provider_id VARCHAR(50) UNIQUE NOT NULL,  -- 'claude', 'openai', etc.
    provider_name VARCHAR(100) NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    
    -- Runtime configuration (JSON for flexibility)
    config JSONB,
    
    -- Metadata
    last_test_at TIMESTAMP WITH TIME ZONE,
    last_test_status VARCHAR(20),  -- 'success', 'failed'
    last_test_error TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX idx_provider_settings_provider_id ON provider_settings(provider_id);
CREATE INDEX idx_provider_settings_enabled ON provider_settings(enabled);

-- Add comments
COMMENT ON TABLE provider_settings IS 'Runtime provider configuration and status';
COMMENT ON COLUMN provider_settings.provider_id IS 'Unique identifier matching config.yaml provider names';
COMMENT ON COLUMN provider_settings.config IS 'Runtime configuration overrides (JSON format)';
