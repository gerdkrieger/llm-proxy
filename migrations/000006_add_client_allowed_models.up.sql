-- Add allowed_models column to oauth_clients table
-- This allows restricting which models a client can access
-- NULL = all models allowed (default for backward compatibility)
-- Empty array = no models allowed (effectively disabled)
-- Array of model IDs = only those models allowed

ALTER TABLE oauth_clients 
ADD COLUMN allowed_models JSONB DEFAULT NULL;

COMMENT ON COLUMN oauth_clients.allowed_models IS 'Array of allowed model IDs (NULL = all models, [] = none, ["model-id"] = specific models)';

-- Index for querying by allowed models
CREATE INDEX idx_oauth_clients_allowed_models ON oauth_clients USING GIN (allowed_models);

-- Update table comment to reflect API client usage
COMMENT ON TABLE oauth_clients IS 'API clients (OAuth 2.0 compatible) with model access control';
