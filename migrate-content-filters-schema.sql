-- Migration: Update content_filters schema to match backend expectations
-- Adds missing columns and removes deprecated ones

BEGIN;

-- Add missing columns if they don't exist
ALTER TABLE content_filters 
  ADD COLUMN IF NOT EXISTS case_sensitive BOOLEAN DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS created_by VARCHAR(255),
  ADD COLUMN IF NOT EXISTS match_count INTEGER DEFAULT 0,
  ADD COLUMN IF NOT EXISTS last_matched_at TIMESTAMP;

-- Remove category column if you want to keep schema minimal (optional)
-- ALTER TABLE content_filters DROP COLUMN IF EXISTS category;

-- Update existing rows to have default values
UPDATE content_filters 
SET case_sensitive = FALSE 
WHERE case_sensitive IS NULL;

UPDATE content_filters 
SET match_count = 0 
WHERE match_count IS NULL;

-- Verify schema
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'content_filters'
ORDER BY ordinal_position;

COMMIT;
