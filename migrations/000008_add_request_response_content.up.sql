-- Migration: Add request and response content to request_logs
-- Created: 2026-02-07
-- Purpose: Store full request/response details for detailed inspection in Live Monitor

-- Add columns for request/response content
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS request_headers JSONB;
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS request_body TEXT;
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS response_headers JSONB;
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS response_body TEXT;
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS response_size_bytes BIGINT;

-- Add comments for documentation
COMMENT ON COLUMN request_logs.request_headers IS 'Request HTTP headers as JSON (sanitized - no auth tokens)';
COMMENT ON COLUMN request_logs.request_body IS 'Full request body (may contain user prompts)';
COMMENT ON COLUMN request_logs.response_headers IS 'Response HTTP headers as JSON';
COMMENT ON COLUMN request_logs.response_body IS 'Full response body from LLM provider';
COMMENT ON COLUMN request_logs.response_size_bytes IS 'Size of response body in bytes';

-- Create index for faster queries on requests with bodies
CREATE INDEX IF NOT EXISTS idx_request_logs_has_body ON request_logs(id) WHERE request_body IS NOT NULL;

-- Add constraint to limit body size (prevent abuse)
-- Note: TEXT columns in PostgreSQL can store up to 1GB, but we may want to limit this
-- For now, we'll handle size limits in the application layer
