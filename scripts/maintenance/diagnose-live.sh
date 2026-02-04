#!/bin/bash
# ============================================================================= 
# LLM-PROXY LIVE SERVER DIAGNOSTICS
# =============================================================================
# Run this on the LIVE server (68.183.208.213) to diagnose connectivity issues
# Usage: ssh user@68.183.208.213 'bash -s' < diagnose-live.sh
# =============================================================================

echo "========================================="
echo "1. CONTAINER STATUS"
echo "========================================="
docker ps --filter "name=llm-proxy" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "========================================="
echo "2. DOCKER NETWORK INSPECTION"
echo "========================================="
docker network ls | grep llm-proxy
echo ""
echo "Containers in llm-proxy-network:"
docker network inspect llm-proxy-network --format '{{range .Containers}}{{.Name}}: {{.IPv4Address}}{{"\n"}}{{end}}' 2>/dev/null || echo "Network not found or no containers"

echo ""
echo "========================================="
echo "3. BACKEND ENVIRONMENT VARIABLES"
echo "========================================="
docker inspect llm-proxy-backend --format='{{range .Config.Env}}{{println .}}{{end}}' 2>/dev/null | grep -E "DATABASE|REDIS" | sort

echo ""
echo "========================================="
echo "4. TEST POSTGRES CONNECTION FROM BACKEND"
echo "========================================="
echo "Testing PostgreSQL connection from backend container..."
docker exec llm-proxy-backend sh -c 'nc -zv postgres 5432 2>&1' 2>/dev/null || \
docker exec llm-proxy-backend sh -c 'wget -qO- --timeout=2 --tries=1 http://postgres:5432 2>&1' 2>/dev/null || \
echo "ERROR: Cannot connect to postgres:5432 from backend"

echo ""
echo "Testing PostgreSQL with psql from backend..."
docker exec llm-proxy-backend sh -c 'command -v psql && psql -h postgres -U proxy_user -d llm_proxy -c "SELECT 1" 2>&1' 2>/dev/null || \
echo "psql not available in backend container (this is OK if Go uses native driver)"

echo ""
echo "========================================="
echo "5. TEST REDIS CONNECTION FROM BACKEND"
echo "========================================="
echo "Testing Redis connection from backend container..."
docker exec llm-proxy-backend sh -c 'nc -zv redis 6379 2>&1' 2>/dev/null || \
docker exec llm-proxy-backend sh -c 'wget -qO- --timeout=2 --tries=1 http://redis:6379 2>&1' 2>/dev/null || \
echo "ERROR: Cannot connect to redis:6379 from backend"

echo ""
echo "========================================="
echo "6. TEST POSTGRES DIRECTLY"
echo "========================================="
echo "Testing PostgreSQL directly..."
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT COUNT(*) as client_count FROM oauth_clients;" 2>&1

echo ""
echo "========================================="
echo "7. TEST REDIS DIRECTLY"
echo "========================================="
echo "Testing Redis directly..."
docker exec llm-proxy-redis redis-cli ping 2>&1
docker exec llm-proxy-redis redis-cli INFO stats | grep -E "total_connections|keyspace" 2>&1

echo ""
echo "========================================="
echo "8. BACKEND LOGS (Last 30 lines)"
echo "========================================="
docker logs llm-proxy-backend --tail 30 2>&1 | grep -E "error|Error|ERROR|warn|Warn|WARN|database|Database|redis|Redis|connect|Connect"

echo ""
echo "========================================="
echo "9. TEST BACKEND API ENDPOINTS"
echo "========================================="
echo "Testing /health endpoint..."
docker exec llm-proxy-backend wget -qO- http://localhost:8080/health 2>&1

echo ""
echo "Testing /admin/clients endpoint..."
docker exec llm-proxy-backend wget -qO- --header='X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012' http://localhost:8080/admin/clients 2>&1 | head -20

echo ""
echo "========================================="
echo "10. CHECK DNS RESOLUTION IN BACKEND"
echo "========================================="
echo "Checking if backend can resolve 'postgres' hostname..."
docker exec llm-proxy-backend sh -c 'nslookup postgres 2>&1 || getent hosts postgres 2>&1' 2>/dev/null || echo "DNS tools not available"

echo ""
echo "Checking if backend can resolve 'redis' hostname..."
docker exec llm-proxy-backend sh -c 'nslookup redis 2>&1 || getent hosts redis 2>&1' 2>/dev/null || echo "DNS tools not available"

echo ""
echo "========================================="
echo "11. NETWORK CONNECTIVITY MATRIX"
echo "========================================="
echo "Can backend reach postgres?"
docker exec llm-proxy-backend sh -c 'timeout 2 cat < /dev/null > /dev/tcp/postgres/5432 && echo "✓ YES" || echo "✗ NO"' 2>/dev/null || echo "bash not available"

echo ""
echo "Can backend reach redis?"
docker exec llm-proxy-backend sh -c 'timeout 2 cat < /dev/null > /dev/tcp/redis/6379 && echo "✓ YES" || echo "✗ NO"' 2>/dev/null || echo "bash not available"

echo ""
echo "========================================="
echo "DIAGNOSIS COMPLETE"
echo "========================================="
