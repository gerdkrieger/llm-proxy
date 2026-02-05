-- Remove allowed_models column
DROP INDEX IF EXISTS idx_oauth_clients_allowed_models;
ALTER TABLE oauth_clients DROP COLUMN IF EXISTS allowed_models;

-- Restore original table comment
COMMENT ON TABLE oauth_clients IS 'OAuth 2.0 client applications';
