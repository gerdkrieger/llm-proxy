-- Rollback: Remove request and response content columns

-- Drop index
DROP INDEX IF EXISTS idx_request_logs_has_body;

-- Remove columns
ALTER TABLE request_logs DROP COLUMN IF EXISTS request_headers;
ALTER TABLE request_logs DROP COLUMN IF EXISTS request_body;
ALTER TABLE request_logs DROP COLUMN IF EXISTS response_headers;
ALTER TABLE request_logs DROP COLUMN IF EXISTS response_body;
ALTER TABLE request_logs DROP COLUMN IF EXISTS response_size_bytes;
