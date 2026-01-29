-- =============================================================================
-- LLM-PROXY DATABASE SCHEMA - INITIAL MIGRATION
-- =============================================================================
-- Version: 1.0.0
-- Description: Initial database schema for LLM-Proxy MVP
-- Author: LLM-Proxy Team
-- Date: 2024-01-29
-- =============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =============================================================================
-- OAUTH & AUTHENTICATION
-- =============================================================================

-- OAuth Clients (Applications that use the proxy)
CREATE TABLE oauth_clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uris TEXT[],
    grant_types TEXT[] DEFAULT ARRAY['authorization_code', 'client_credentials', 'refresh_token'],
    default_scope TEXT DEFAULT 'read,write',
    rate_limit_rpm INTEGER DEFAULT 1000,  -- NULL = unlimited
    rate_limit_rpd INTEGER,               -- NULL = unlimited
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE oauth_clients IS 'OAuth 2.0 client applications';
COMMENT ON COLUMN oauth_clients.client_id IS 'Public client identifier';
COMMENT ON COLUMN oauth_clients.client_secret_hash IS 'Hashed client secret (bcrypt)';
COMMENT ON COLUMN oauth_clients.rate_limit_rpm IS 'Requests per minute (NULL = unlimited)';
COMMENT ON COLUMN oauth_clients.rate_limit_rpd IS 'Requests per day (NULL = unlimited)';

-- Indexes
CREATE INDEX idx_oauth_clients_client_id ON oauth_clients(client_id);
CREATE INDEX idx_oauth_clients_enabled ON oauth_clients(enabled);

-- OAuth Tokens (Access & Refresh Tokens)
CREATE TABLE oauth_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID REFERENCES oauth_clients(id) ON DELETE CASCADE,
    access_token VARCHAR(512) UNIQUE NOT NULL,
    refresh_token VARCHAR(512) UNIQUE,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMP,
    scope TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE oauth_tokens IS 'OAuth 2.0 access and refresh tokens';
COMMENT ON COLUMN oauth_tokens.access_token IS 'JWT access token';
COMMENT ON COLUMN oauth_tokens.refresh_token IS 'Refresh token for obtaining new access tokens';

-- Indexes
CREATE INDEX idx_oauth_tokens_client_id ON oauth_tokens(client_id);
CREATE INDEX idx_oauth_tokens_access_token ON oauth_tokens(access_token);
CREATE INDEX idx_oauth_tokens_expires_at ON oauth_tokens(expires_at);

-- =============================================================================
-- ADMIN USERS
-- =============================================================================

CREATE TABLE admin_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    api_key_hash VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);

COMMENT ON TABLE admin_users IS 'Administrative users with API key access';
COMMENT ON COLUMN admin_users.api_key_hash IS 'Hashed admin API key (bcrypt)';

-- Indexes
CREATE INDEX idx_admin_users_api_key_hash ON admin_users(api_key_hash);
CREATE INDEX idx_admin_users_enabled ON admin_users(enabled);

-- =============================================================================
-- ACCESS CONTROL (Simplified for MVP)
-- =============================================================================

CREATE TABLE access_lists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    list_type VARCHAR(20) NOT NULL CHECK (list_type IN ('whitelist', 'blacklist')),
    category VARCHAR(50) NOT NULL CHECK (category IN ('client', 'ip', 'model')),
    value TEXT NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(list_type, category, value)
);

COMMENT ON TABLE access_lists IS 'Access control lists (whitelist/blacklist)';
COMMENT ON COLUMN access_lists.list_type IS 'Type of list: whitelist or blacklist';
COMMENT ON COLUMN access_lists.category IS 'What is being filtered: client, ip, or model';
COMMENT ON COLUMN access_lists.value IS 'The value to match (client_id, IP address, or model name)';

-- Indexes
CREATE INDEX idx_access_lists_type_category ON access_lists(list_type, category);
CREATE INDEX idx_access_lists_enabled ON access_lists(enabled);

-- =============================================================================
-- REQUEST LOGGING & ANALYTICS
-- =============================================================================

CREATE TABLE request_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID REFERENCES oauth_clients(id) ON DELETE SET NULL,
    request_id VARCHAR(255) UNIQUE NOT NULL,
    method VARCHAR(10),
    path VARCHAR(255),
    model VARCHAR(100),
    provider VARCHAR(50),
    prompt_tokens INTEGER,
    completion_tokens INTEGER,
    total_tokens INTEGER,
    cost_usd NUMERIC(10, 6),
    duration_ms INTEGER,
    status_code INTEGER,
    cached BOOLEAN DEFAULT false,
    ip_address INET,
    user_agent TEXT,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE request_logs IS 'Detailed logs of all API requests';
COMMENT ON COLUMN request_logs.request_id IS 'Unique request identifier (X-Request-ID)';
COMMENT ON COLUMN request_logs.cached IS 'Whether the response was served from cache';
COMMENT ON COLUMN request_logs.cost_usd IS 'Calculated cost in USD based on token usage';

-- Indexes (optimized for queries)
CREATE INDEX idx_request_logs_client_id ON request_logs(client_id);
CREATE INDEX idx_request_logs_created_at ON request_logs(created_at DESC);
CREATE INDEX idx_request_logs_model ON request_logs(model);
CREATE INDEX idx_request_logs_status_code ON request_logs(status_code);
CREATE INDEX idx_request_logs_cached ON request_logs(cached);

-- Partitioning (for production - by month)
-- CREATE TABLE request_logs_y2024m01 PARTITION OF request_logs
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- =============================================================================
-- ANALYTICS AGGREGATIONS
-- =============================================================================

CREATE TABLE usage_analytics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID REFERENCES oauth_clients(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    model VARCHAR(100),
    total_requests INTEGER DEFAULT 0,
    successful_requests INTEGER DEFAULT 0,
    failed_requests INTEGER DEFAULT 0,
    total_tokens BIGINT DEFAULT 0,
    total_cost_usd NUMERIC(10, 2) DEFAULT 0,
    avg_duration_ms INTEGER,
    cache_hit_rate NUMERIC(5, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(client_id, date, model)
);

COMMENT ON TABLE usage_analytics IS 'Daily aggregated usage statistics per client and model';
COMMENT ON COLUMN usage_analytics.cache_hit_rate IS 'Percentage of requests served from cache (0-100)';

-- Indexes
CREATE INDEX idx_usage_analytics_client_date ON usage_analytics(client_id, date DESC);
CREATE INDEX idx_usage_analytics_date ON usage_analytics(date DESC);

-- =============================================================================
-- BILLING SYSTEM (Simplified for MVP)
-- =============================================================================

CREATE TABLE billing_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id UUID UNIQUE REFERENCES oauth_clients(id) ON DELETE CASCADE,
    credits NUMERIC(12, 4) DEFAULT 0.0000,
    currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE billing_accounts IS 'Billing accounts for OAuth clients';
COMMENT ON COLUMN billing_accounts.credits IS 'Current credit balance (prepaid)';

-- Indexes
CREATE INDEX idx_billing_accounts_client_id ON billing_accounts(client_id);

CREATE TABLE billing_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    account_id UUID REFERENCES billing_accounts(id) ON DELETE CASCADE,
    transaction_type VARCHAR(20) NOT NULL CHECK (transaction_type IN ('charge', 'topup', 'refund')),
    amount NUMERIC(12, 4) NOT NULL,
    balance_after NUMERIC(12, 4),
    request_id VARCHAR(255) REFERENCES request_logs(request_id) ON DELETE SET NULL,
    description TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE billing_transactions IS 'Transaction history for billing accounts';
COMMENT ON COLUMN billing_transactions.transaction_type IS 'Type: charge (debit), topup (credit), refund';
COMMENT ON COLUMN billing_transactions.balance_after IS 'Account balance after transaction';

-- Indexes
CREATE INDEX idx_billing_transactions_account_id ON billing_transactions(account_id);
CREATE INDEX idx_billing_transactions_created_at ON billing_transactions(created_at DESC);
CREATE INDEX idx_billing_transactions_type ON billing_transactions(transaction_type);

-- =============================================================================
-- PROVIDER CONFIGURATION
-- =============================================================================

CREATE TABLE provider_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_name VARCHAR(50) NOT NULL,
    api_key_encrypted TEXT NOT NULL,
    weight INTEGER DEFAULT 1,
    max_rpm INTEGER,
    enabled BOOLEAN DEFAULT true,
    health_status VARCHAR(20) DEFAULT 'unknown',
    last_health_check TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE provider_configs IS 'LLM provider API keys and configuration';
COMMENT ON COLUMN provider_configs.api_key_encrypted IS 'Encrypted provider API key (AES-256)';
COMMENT ON COLUMN provider_configs.weight IS 'Load balancing weight (higher = more traffic)';
COMMENT ON COLUMN provider_configs.health_status IS 'Current health: healthy, degraded, unhealthy, unknown';

-- Indexes
CREATE INDEX idx_provider_configs_enabled ON provider_configs(enabled);
CREATE INDEX idx_provider_configs_provider_name ON provider_configs(provider_name);

-- =============================================================================
-- SYSTEM SETTINGS
-- =============================================================================

CREATE TABLE system_settings (
    key VARCHAR(255) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE system_settings IS 'System-wide configuration settings';

-- =============================================================================
-- TRIGGERS & FUNCTIONS
-- =============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to relevant tables
CREATE TRIGGER update_oauth_clients_updated_at BEFORE UPDATE ON oauth_clients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_access_lists_updated_at BEFORE UPDATE ON access_lists
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_usage_analytics_updated_at BEFORE UPDATE ON usage_analytics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_billing_accounts_updated_at BEFORE UPDATE ON billing_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_provider_configs_updated_at BEFORE UPDATE ON provider_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_settings_updated_at BEFORE UPDATE ON system_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =============================================================================
-- DEFAULT DATA
-- =============================================================================

-- Insert default system settings
INSERT INTO system_settings (key, value, description) VALUES
('cache.enabled', 'true', 'Enable/disable caching'),
('cache.ttl', '3600', 'Cache TTL in seconds'),
('rate_limiting.enabled', 'true', 'Enable/disable rate limiting'),
('load_balancing.strategy', '"round-robin"', 'Load balancing strategy: round-robin, weighted, least-latency');

-- =============================================================================
-- GRANTS (for application user)
-- =============================================================================

-- All tables are owned by the database owner
-- The application connects as the same user for MVP
-- In production, create a separate app user with limited privileges:
-- CREATE USER llm_proxy_app WITH PASSWORD 'secure_password';
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO llm_proxy_app;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO llm_proxy_app;

-- =============================================================================
-- END OF MIGRATION
-- =============================================================================
