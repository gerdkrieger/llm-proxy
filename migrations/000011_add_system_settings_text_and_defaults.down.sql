-- Migration: Revert system_settings value column back to JSONB
-- Created: 2026-02-08

-- Remove the body capture setting
DELETE FROM system_settings WHERE key = 'capture_request_response_bodies';

-- Convert TEXT back to JSONB
ALTER TABLE system_settings ADD COLUMN IF NOT EXISTS value_jsonb JSONB;

UPDATE system_settings SET value_jsonb = to_jsonb(value);

ALTER TABLE system_settings DROP COLUMN IF EXISTS value;
ALTER TABLE system_settings RENAME COLUMN value_jsonb TO value;

ALTER TABLE system_settings ALTER COLUMN value SET NOT NULL;
