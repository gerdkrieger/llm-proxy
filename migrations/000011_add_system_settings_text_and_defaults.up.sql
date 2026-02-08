-- Migration: Adjust system_settings value column and add new defaults
-- Created: 2026-02-08
-- Purpose: Change value column from JSONB to TEXT for simpler key-value storage,
--          and add default settings for request/response body capture

-- Step 1: Convert existing JSONB values to plain TEXT
-- First add a temporary column
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS value_text TEXT;

-- Copy data, stripping JSON quotes from simple string values
UPDATE system_settings SET value_text = CASE
    WHEN value::text LIKE '"%"' THEN TRIM(BOTH '"' FROM value::text)
    ELSE value::text
END;

-- Drop old column and rename
ALTER TABLE system_settings DROP COLUMN IF EXISTS value;
ALTER TABLE system_settings RENAME COLUMN value_text TO value;

-- Ensure NOT NULL
ALTER TABLE system_settings ALTER COLUMN value SET NOT NULL;

-- Step 2: Add default setting for body capture
INSERT INTO system_settings (key, value, description)
VALUES ('capture_request_response_bodies', 'true', 'Enable/disable capturing request and response bodies in request logs')
ON CONFLICT (key) DO NOTHING;
