-- Drop provider_models table
DROP INDEX IF EXISTS idx_provider_models_provider_enabled;
DROP INDEX IF EXISTS idx_provider_models_enabled;
DROP INDEX IF EXISTS idx_provider_models_provider_id;
DROP TABLE IF EXISTS provider_models;
