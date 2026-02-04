#!/bin/bash
# =============================================================================
# FIX PROVIDER_MODELS TABLE - Add missing description column
# =============================================================================
# Run on LIVE server to add missing description column
# =============================================================================

echo "========================================="
echo "FIXING provider_models TABLE SCHEMA"
echo "========================================="

echo ""
echo "Step 1: Checking current schema..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\d provider_models"

echo ""
echo "Step 2: Adding missing description column..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
ALTER TABLE provider_models 
ADD COLUMN IF NOT EXISTS description TEXT;
"

echo ""
echo "Step 3: Verifying updated schema..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\d provider_models"

echo ""
echo "Step 4: Restarting backend to sync models..."
docker restart llm-proxy-backend

echo ""
echo "Step 5: Waiting for backend to start..."
sleep 10

echo ""
echo "Step 6: Checking backend logs for errors..."
docker logs llm-proxy-backend --tail 30 | grep -iE "error|ERROR|model sync" || echo "No errors found"

echo ""
echo "========================================="
echo "FIX COMPLETE"
echo "========================================="
echo ""
echo "Verification:"
echo "curl https://llmproxy.aitrail.ch/admin/providers -H 'X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012'"
echo ""
