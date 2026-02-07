-- Migration: Enhance request_logs table for Live Monitor
-- Created: 2026-02-07
-- Purpose: Add authentication tracking and filtering info to existing request_logs table

-- Add new columns for authentication tracking
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS auth_type VARCHAR(20);
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS api_key_name VARCHAR(100);
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS was_filtered BOOLEAN DEFAULT FALSE;
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS filter_reason TEXT;

-- Add comments for new columns
COMMENT ON COLUMN request_logs.auth_type IS 'Type of authentication used: api_key, oauth, admin, or none';
COMMENT ON COLUMN request_logs.api_key_name IS 'Name of the API key used (never store the actual key!)';
COMMENT ON COLUMN request_logs.was_filtered IS 'Whether the request was blocked by content filters';
COMMENT ON COLUMN request_logs.filter_reason IS 'Reason why request was filtered (if applicable)';

-- Create additional indexes for Live Monitor queries
CREATE INDEX IF NOT EXISTS idx_request_logs_api_key_name ON request_logs(api_key_name);
CREATE INDEX IF NOT EXISTS idx_request_logs_was_filtered ON request_logs(was_filtered);
CREATE INDEX IF NOT EXISTS idx_request_logs_ip_created ON request_logs(ip_address, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_request_logs_auth_type ON request_logs(auth_type);
