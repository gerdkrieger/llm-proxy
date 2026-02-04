-- Migration: Add client_secret_hash and statistics columns
-- Date: 2026-02-04
-- Issue: Backend code was updated to use hashed secrets and new statistics fields,
--        but database schema was not migrated on production server
-- 
-- Symptoms:
--   - POST /admin/clients returned 500: "column client_secret_hash does not exist"
--   - GET /admin/stats/usage returned 500: "column duration_ms does not exist"
--
-- Root Cause:
--   - Backend image was deployed with new code expecting new columns
--   - Database schema was never migrated to match the new backend expectations
--   - This migration adds missing columns and adjusts constraints

-- ============================================================================
-- 1. OAuth Clients - Add hashed secret support
-- ============================================================================

-- Add new column for hashed secrets (more secure)
ALTER TABLE oauth_clients 
ADD COLUMN IF NOT EXISTS client_secret_hash VARCHAR(255);

-- Make old client_secret column optional (for backwards compatibility)
ALTER TABLE oauth_clients 
ALTER COLUMN client_secret DROP NOT NULL;

-- Make new client_secret_hash column required
-- (Backend now stores hashed secrets instead of plaintext)
ALTER TABLE oauth_clients 
ALTER COLUMN client_secret_hash SET NOT NULL;

-- ============================================================================
-- 2. Request Logs - Add cost tracking
-- ============================================================================

-- Add cost column for billing/statistics
ALTER TABLE request_logs 
ADD COLUMN IF NOT EXISTS cost_usd DECIMAL(10,6);

-- ============================================================================
-- 3. Request Logs - Rename latency_ms to duration_ms
-- ============================================================================

-- Rename column to match backend expectations
-- Check if column exists before renaming to make migration idempotent
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'request_logs' 
        AND column_name = 'latency_ms'
    ) THEN
        ALTER TABLE request_logs RENAME COLUMN latency_ms TO duration_ms;
    END IF;
END $$;

-- ============================================================================
-- Verification queries (run manually to verify migration)
-- ============================================================================

-- Verify oauth_clients columns
-- SELECT column_name, is_nullable, data_type 
-- FROM information_schema.columns 
-- WHERE table_name = 'oauth_clients' 
-- AND column_name LIKE '%secret%'
-- ORDER BY column_name;

-- Expected result:
--     column_name     | is_nullable |     data_type     
-- --------------------+-------------+-------------------
--  client_secret      | YES         | character varying
--  client_secret_hash | NO          | character varying

-- Verify request_logs columns
-- SELECT column_name, data_type 
-- FROM information_schema.columns 
-- WHERE table_name = 'request_logs' 
-- AND column_name IN ('cost_usd', 'duration_ms')
-- ORDER BY column_name;

-- Expected result:
--  column_name | data_type 
-- -------------+-----------
--  cost_usd    | numeric
--  duration_ms | integer
