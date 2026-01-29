-- =============================================================================
-- LLM-PROXY DATABASE SCHEMA - ROLLBACK MIGRATION
-- =============================================================================
-- Version: 1.0.0
-- Description: Rollback initial database schema
-- WARNING: This will delete ALL data!
-- =============================================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_system_settings_updated_at ON system_settings;
DROP TRIGGER IF EXISTS update_provider_configs_updated_at ON provider_configs;
DROP TRIGGER IF EXISTS update_billing_accounts_updated_at ON billing_accounts;
DROP TRIGGER IF EXISTS update_usage_analytics_updated_at ON usage_analytics;
DROP TRIGGER IF EXISTS update_access_lists_updated_at ON access_lists;
DROP TRIGGER IF EXISTS update_oauth_clients_updated_at ON oauth_clients;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS system_settings CASCADE;
DROP TABLE IF EXISTS provider_configs CASCADE;
DROP TABLE IF EXISTS billing_transactions CASCADE;
DROP TABLE IF EXISTS billing_accounts CASCADE;
DROP TABLE IF EXISTS usage_analytics CASCADE;
DROP TABLE IF EXISTS request_logs CASCADE;
DROP TABLE IF EXISTS access_lists CASCADE;
DROP TABLE IF EXISTS admin_users CASCADE;
DROP TABLE IF EXISTS oauth_tokens CASCADE;
DROP TABLE IF EXISTS oauth_clients CASCADE;

-- Drop extensions (optional - comment out if shared with other databases)
-- DROP EXTENSION IF EXISTS "pgcrypto";
-- DROP EXTENSION IF EXISTS "uuid-ossp";

-- =============================================================================
-- END OF ROLLBACK
-- =============================================================================
