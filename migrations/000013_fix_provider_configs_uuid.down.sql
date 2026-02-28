-- ============================================================================
-- Migration 000011: Rollback provider_configs UUID change
-- ============================================================================
-- WARNING: This rollback will truncate all provider_configs data!
-- ============================================================================

-- Step 1: Truncate dependent tables
TRUNCATE TABLE provider_models CASCADE;
TRUNCATE TABLE provider_api_keys CASCADE;
TRUNCATE TABLE provider_settings CASCADE;
TRUNCATE TABLE provider_configs CASCADE;

-- Step 2: Convert id column back to INTEGER
ALTER TABLE provider_configs ALTER COLUMN id DROP DEFAULT;
ALTER TABLE provider_configs ALTER COLUMN id TYPE INTEGER USING 0;
CREATE SEQUENCE IF NOT EXISTS provider_configs_id_seq;
ALTER TABLE provider_configs ALTER COLUMN id SET DEFAULT nextval('provider_configs_id_seq'::regclass);

-- Step 3: Remove health_status columns
ALTER TABLE provider_configs 
  DROP COLUMN IF EXISTS health_status,
  DROP COLUMN IF EXISTS last_health_check;

-- Step 4: Remove api_key_encrypted column
ALTER TABLE provider_configs 
  DROP COLUMN IF EXISTS api_key_encrypted;

-- Step 5: Reinsert built-in providers with INTEGER ids
INSERT INTO provider_configs (provider_id, provider_name, provider_type, enabled)
VALUES 
    ('claude', 'Anthropic Claude', 'claude', true),
    ('openai', 'OpenAI', 'openai', true);
