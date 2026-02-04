#!/bin/bash
# =============================================================================
# FIX PROVIDER_MODELS ID TYPE - Change from SERIAL to UUID
# =============================================================================

echo "========================================="
echo "FIXING provider_models ID TYPE"
echo "========================================="

echo ""
echo "Step 1: Checking current data..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT COUNT(*) as row_count FROM provider_models;"

echo ""
echo "Step 2: Dropping and recreating provider_models table with UUID id..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
-- Drop existing table (since it's empty anyway)
DROP TABLE IF EXISTS provider_models CASCADE;

-- Recreate with UUID id
CREATE TABLE provider_models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id VARCHAR(50) REFERENCES provider_configs(provider_id) ON DELETE CASCADE,
    model_id VARCHAR(255) NOT NULL,
    model_name VARCHAR(255) NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    description TEXT,
    pricing JSONB,
    capabilities JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(provider_id, model_id)
);

CREATE INDEX IF NOT EXISTS idx_provider_models_provider_id ON provider_models(provider_id);
CREATE INDEX IF NOT EXISTS idx_provider_models_enabled ON provider_models(enabled);
"

echo ""
echo "Step 3: Verifying new schema..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\d provider_models"

echo ""
echo "Step 4: Restarting backend to sync models..."
docker restart llm-proxy-backend

echo ""
echo "Step 5: Waiting for backend to sync models..."
sleep 15

echo ""
echo "Step 6: Checking how many models were synced..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT COUNT(*) as synced_models FROM provider_models;"

echo ""
echo "Step 7: Sample of synced models..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT provider_id, model_id, model_name, enabled FROM provider_models LIMIT 10;"

echo ""
echo "Step 8: Checking backend logs..."
docker logs llm-proxy-backend --tail 20 | grep -E "Model sync completed|error|ERROR" || echo "No sync messages found"

echo ""
echo "========================================="
echo "FIX COMPLETE"
echo "========================================="
echo ""
