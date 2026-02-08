-- Migration: Fix missing columns in request_logs table
-- Created: 2026-02-08
-- Purpose: Add request_id and ip_address columns that were missing from previous schema

-- Add request_id column (critical for request tracking)
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS request_id VARCHAR(255);

-- Add ip_address column (was referenced in migration 007 but never created)
ALTER TABLE request_logs ADD COLUMN IF NOT EXISTS ip_address VARCHAR(50);

-- Create indexes for these columns
CREATE UNIQUE INDEX IF NOT EXISTS idx_request_logs_request_id ON request_logs(request_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_ip ON request_logs(ip_address);

-- Add comments
COMMENT ON COLUMN request_logs.request_id IS 'Unique request identifier (X-Request-ID)';
COMMENT ON COLUMN request_logs.ip_address IS 'IP address of the client making the request';
