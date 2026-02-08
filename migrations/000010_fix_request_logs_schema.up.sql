-- Migration: Fix missing columns in request_logs table
-- Created: 2026-02-08
-- Purpose: Add all missing columns that should have been created by migration 000001

-- Add missing core columns
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS method VARCHAR(10);
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS path VARCHAR(255);
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS user_agent TEXT;
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS prompt_tokens INTEGER;
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS completion_tokens INTEGER;

-- Add comments for clarity
COMMENT ON COLUMN request_logs.method IS 'HTTP method (GET, POST, etc.)';
COMMENT ON COLUMN request_logs.path IS 'Request path/endpoint';
COMMENT ON COLUMN request_logs.user_agent IS 'Client user agent string';
COMMENT ON COLUMN request_logs.prompt_tokens IS 'Number of tokens in the prompt/input';
COMMENT ON COLUMN request_logs.completion_tokens IS 'Number of tokens in the completion/output';

-- Create missing indexes
CREATE INDEX IF NOT EXISTS idx_request_logs_method ON request_logs(method);
CREATE INDEX IF NOT EXISTS idx_request_logs_path ON request_logs(path);
