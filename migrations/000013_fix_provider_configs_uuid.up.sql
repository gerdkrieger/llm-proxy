-- ============================================================================
-- Migration 000011: Fix provider_configs to use UUID instead of INTEGER
-- ============================================================================
-- Created: 2026-02-28
-- Purpose: Convert provider_configs.id from INTEGER to UUID for consistency
--          with new provider management code
-- ============================================================================

-- Step 1: Truncate dependent tables (data will be re-synced from ProviderManager)
TRUNCATE TABLE provider_models CASCADE;
TRUNCATE TABLE provider_api_keys CASCADE;
TRUNCATE TABLE provider_settings CASCADE;
TRUNCATE TABLE provider_configs CASCADE;

-- Step 2: Convert id column from INTEGER to UUID
ALTER TABLE provider_configs ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS provider_configs_id_seq CASCADE;
ALTER TABLE provider_configs ALTER COLUMN id TYPE UUID USING gen_random_uuid();
ALTER TABLE provider_configs ALTER COLUMN id SET DEFAULT gen_random_uuid();

-- Step 3: Add health_status columns if they don't exist
ALTER TABLE provider_configs 
  ADD COLUMN IF NOT EXISTS health_status VARCHAR(20) DEFAULT 'unknown',
  ADD COLUMN IF NOT EXISTS last_health_check TIMESTAMP;

-- Step 4: Add api_key_encrypted column if it doesn't exist (for legacy compatibility)
ALTER TABLE provider_configs 
  ADD COLUMN IF NOT EXISTS api_key_encrypted TEXT DEFAULT '';

-- Step 5: Insert built-in providers (will be managed by ProviderManager)
INSERT INTO provider_configs (provider_id, provider_name, provider_type, enabled)
VALUES 
    ('claude', 'Anthropic Claude', 'claude', true),
    ('openai', 'OpenAI', 'openai', true)
ON CONFLICT (provider_id) DO NOTHING;

-- Step 6: Verify the migration
DO $$
DECLARE
    id_type TEXT;
BEGIN
    SELECT data_type INTO id_type 
    FROM information_schema.columns 
    WHERE table_name = 'provider_configs' AND column_name = 'id';
    
    IF id_type = 'uuid' THEN
        RAISE NOTICE 'Migration 000011: SUCCESS - provider_configs.id is now UUID';
    ELSE
        RAISE EXCEPTION 'Migration 000011: FAILED - provider_configs.id is still %', id_type;
    END IF;
END $$;
