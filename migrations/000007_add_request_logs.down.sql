-- Rollback migration: Remove columns added in 000007
-- This removes authentication tracking and filtering columns

DROP INDEX IF EXISTS idx_request_logs_auth_type;
DROP INDEX IF EXISTS idx_request_logs_ip_created;
DROP INDEX IF EXISTS idx_request_logs_was_filtered;
DROP INDEX IF EXISTS idx_request_logs_api_key_name;

ALTER TABLE request_logs DROP COLUMN IF EXISTS filter_reason;
ALTER TABLE request_logs DROP COLUMN IF EXISTS was_filtered;
ALTER TABLE request_logs DROP COLUMN IF EXISTS api_key_name;
ALTER TABLE request_logs DROP COLUMN IF EXISTS auth_type;
