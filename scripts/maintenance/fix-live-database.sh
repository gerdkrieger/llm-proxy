#!/bin/bash
# =============================================================================
# FIX LIVE DATABASE - Initialize Database Schema
# =============================================================================
# Run this on LIVE server to create missing database tables
# =============================================================================

echo "========================================="
echo "FIXING LLM-PROXY DATABASE ON LIVE SERVER"
echo "========================================="

echo ""
echo "Step 1: Checking current database state..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\dt" 2>&1

echo ""
echo "Step 2: Running database migrations..."
echo "Checking if backend has migration files..."

# Option A: If backend binary has embedded migrations
echo "Trying to run migrations via backend binary..."
docker exec llm-proxy-backend /app/llm-proxy migrate 2>&1 || echo "Binary doesn't have migrate command"

# Option B: Check if SQL migration files exist in deployments/docker/migrations
echo ""
echo "Checking for SQL migration files..."
docker exec llm-proxy-backend sh -c 'ls -la /app/migrations/*.sql 2>/dev/null || ls -la /deployments/docker/migrations/*.sql 2>/dev/null || echo "No migration files found in container"'

# Option C: Apply migrations from host if they exist
echo ""
echo "Step 3: Creating schema manually if migrations not available..."

# Create schema SQL
cat > /tmp/llm_proxy_schema.sql <<'EOSQL'
-- =============================================================================
-- LLM-PROXY DATABASE SCHEMA
-- =============================================================================

-- Schema Migrations tracking
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- OAuth Clients
CREATE TABLE IF NOT EXISTS oauth_clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uris TEXT[],
    grant_types VARCHAR(50)[] DEFAULT ARRAY['client_credentials'],
    default_scope VARCHAR(255) DEFAULT 'read write',
    rate_limit_rpm INTEGER,
    rate_limit_rpd INTEGER,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- OAuth Tokens
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(255) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    access_token VARCHAR(512) UNIQUE NOT NULL,
    refresh_token VARCHAR(512) UNIQUE,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    scope VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    refresh_expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Provider Configurations
CREATE TABLE IF NOT EXISTS provider_configs (
    id SERIAL PRIMARY KEY,
    provider_id VARCHAR(50) UNIQUE NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    provider_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Provider API Keys (Encrypted)
CREATE TABLE IF NOT EXISTS provider_settings (
    id SERIAL PRIMARY KEY,
    provider_id VARCHAR(50) REFERENCES provider_configs(provider_id) ON DELETE CASCADE,
    key_name VARCHAR(255) NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, key_name)
);

-- Provider Models
CREATE TABLE IF NOT EXISTS provider_models (
    id SERIAL PRIMARY KEY,
    provider_id VARCHAR(50) REFERENCES provider_configs(provider_id) ON DELETE CASCADE,
    model_id VARCHAR(255) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    pricing JSONB,
    capabilities JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, model_id)
);

-- Content Filters
CREATE TABLE IF NOT EXISTS content_filters (
    id SERIAL PRIMARY KEY,
    pattern TEXT NOT NULL,
    replacement TEXT,
    filter_type VARCHAR(50) NOT NULL,
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    description TEXT,
    category VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Filter Matches (for tracking)
CREATE TABLE IF NOT EXISTS filter_matches (
    id SERIAL PRIMARY KEY,
    filter_id INTEGER REFERENCES content_filters(id) ON DELETE CASCADE,
    client_id VARCHAR(255),
    matched_text TEXT,
    replacement_text TEXT,
    matched_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Request Logs
CREATE TABLE IF NOT EXISTS request_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(255),
    provider_id VARCHAR(50),
    model VARCHAR(255),
    request_tokens INTEGER,
    response_tokens INTEGER,
    total_tokens INTEGER,
    latency_ms INTEGER,
    status_code INTEGER,
    error_message TEXT,
    cached BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Usage Analytics
CREATE TABLE IF NOT EXISTS usage_analytics (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(255),
    provider_id VARCHAR(50),
    model VARCHAR(255),
    date DATE NOT NULL,
    total_requests INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    cache_hits INTEGER DEFAULT 0,
    cache_misses INTEGER DEFAULT 0,
    total_cost DECIMAL(10,4) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(client_id, provider_id, model, date)
);

-- Access Control Lists
CREATE TABLE IF NOT EXISTS access_lists (
    id SERIAL PRIMARY KEY,
    client_id VARCHAR(255) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    provider_id VARCHAR(50),
    model VARCHAR(255),
    allowed BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(client_id, provider_id, model)
);

-- Admin Users
CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    api_key VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) DEFAULT 'admin',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Billing Accounts
CREATE TABLE IF NOT EXISTS billing_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id VARCHAR(255) REFERENCES oauth_clients(client_id) ON DELETE CASCADE,
    balance DECIMAL(10,4) DEFAULT 0,
    credit_limit DECIMAL(10,4) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Billing Transactions
CREATE TABLE IF NOT EXISTS billing_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID REFERENCES billing_accounts(id) ON DELETE CASCADE,
    amount DECIMAL(10,4) NOT NULL,
    transaction_type VARCHAR(50) NOT NULL,
    description TEXT,
    request_log_id UUID REFERENCES request_logs(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- System Settings
CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_client_id ON oauth_tokens(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_access_token ON oauth_tokens(access_token);
CREATE INDEX IF NOT EXISTS idx_request_logs_client_id ON request_logs(client_id);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_usage_analytics_date ON usage_analytics(date);
CREATE INDEX IF NOT EXISTS idx_filter_matches_filter_id ON filter_matches(filter_id);
CREATE INDEX IF NOT EXISTS idx_billing_transactions_account_id ON billing_transactions(account_id);

-- Mark migration as applied
INSERT INTO schema_migrations (version) VALUES (1) ON CONFLICT DO NOTHING;

EOSQL

echo "Applying schema to database..."
docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/llm_proxy_schema.sql

echo ""
echo "Step 4: Verifying tables were created..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\dt"

echo ""
echo "Step 5: Checking if clients exist..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT COUNT(*) as client_count FROM oauth_clients;"

echo ""
echo "========================================="
echo "DATABASE INITIALIZATION COMPLETE"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Restart backend: docker restart llm-proxy-backend"
echo "2. Test API: curl http://localhost:8080/admin/clients -H 'X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012'"
echo ""
