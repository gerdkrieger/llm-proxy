#!/bin/bash
# =============================================================================
# LLM-PROXY QUICK CHECK
# =============================================================================
# Schneller Gesundheitscheck und Status-Übersicht
# =============================================================================

set -euo pipefail

# Farben
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Server
SERVER="${1:-openweb}"

clear
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  LLM-PROXY QUICK CHECK${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${CYAN}Server:${NC} ${SERVER}"
echo -e "${CYAN}Zeit:${NC}   $(date '+%Y-%m-%d %H:%M:%S')"
echo ""

# =============================================================================
# 1. CONTAINER STATUS
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  1. CONTAINER STATUS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

BACKEND_STATUS=$(ssh "${SERVER}" "docker ps --filter name=llm-proxy-backend --format '{{.Status}}' 2>/dev/null || echo 'NOT_RUNNING'")
POSTGRES_STATUS=$(ssh "${SERVER}" "docker ps --filter name=llm-proxy-postgres --format '{{.Status}}' 2>/dev/null || echo 'NOT_RUNNING'")
REDIS_STATUS=$(ssh "${SERVER}" "docker ps --filter name=llm-proxy-redis --format '{{.Status}}' 2>/dev/null || echo 'NOT_RUNNING'")
ADMIN_STATUS=$(ssh "${SERVER}" "docker ps --filter name=llm-proxy-admin --format '{{.Status}}' 2>/dev/null || echo 'NOT_RUNNING'")

check_status() {
  if [[ "$1" == *"Up"* ]]; then
    echo -e "${GREEN}✓ $2${NC} - Running ($1)"
  else
    echo -e "${RED}✗ $2${NC} - NOT RUNNING"
  fi
}

check_status "$BACKEND_STATUS" "Backend"
check_status "$POSTGRES_STATUS" "PostgreSQL"
check_status "$REDIS_STATUS" "Redis"
check_status "$ADMIN_STATUS" "Admin UI"

echo ""

# =============================================================================
# 2. LETZTE 1 STUNDE ÜBERSICHT
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  2. LETZTE 1 STUNDE${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

STATS=$(ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -A -F'|' -c \"
SELECT 
  COUNT(*) as total,
  COUNT(CASE WHEN status_code = 200 THEN 1 END) as success,
  COUNT(CASE WHEN status_code >= 400 THEN 1 END) as errors,
  ROUND(AVG(duration_ms)) as avg_duration
FROM request_logs
WHERE created_at > NOW() - INTERVAL '1 hour';
\" 2>/dev/null || echo '0|0|0|0'")

TOTAL=$(echo "$STATS" | cut -d'|' -f1 | tr -d ' ')
SUCCESS=$(echo "$STATS" | cut -d'|' -f2 | tr -d ' ')
ERRORS=$(echo "$STATS" | cut -d'|' -f3 | tr -d ' ')
AVG_DURATION=$(echo "$STATS" | cut -d'|' -f4 | tr -d ' ')

echo -e "${CYAN}Total Requests:${NC}      ${TOTAL}"
echo -e "${GREEN}Erfolgreiche:${NC}        ${SUCCESS}"
echo -e "${RED}Fehler:${NC}              ${ERRORS}"
echo -e "${BLUE}Ø Duration:${NC}          ${AVG_DURATION} ms"

echo ""

# =============================================================================
# 3. FILTERUNG LETZTE 1 STUNDE
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  3. FILTERUNG (LETZTE 1 STUNDE)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

FILTER_STATS=$(ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -A -F'|' -c \"
SELECT 
  COUNT(*) as total_matches,
  COUNT(DISTINCT request_id) as filtered_requests
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '1 hour';
\" 2>/dev/null || echo '0|0'")

FILTER_MATCHES=$(echo "$FILTER_STATS" | cut -d'|' -f1 | tr -d ' ')
FILTERED_REQUESTS=$(echo "$FILTER_STATS" | cut -d'|' -f2 | tr -d ' ')

if [[ "$FILTER_MATCHES" -gt 0 ]]; then
  echo -e "${GREEN}✓ Filterung aktiv${NC}"
  echo -e "${CYAN}  Filter-Matches:${NC}       ${FILTER_MATCHES}"
  echo -e "${CYAN}  Gefilterte Requests:${NC}  ${FILTERED_REQUESTS}"
  
  # Top gefilterte Pattern
  echo ""
  echo -e "${CYAN}  Top gefilterte Pattern:${NC}"
  ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -c \"
SELECT 
  '    - ' || pattern || ': ' || COUNT(*) 
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '1 hour'
GROUP BY pattern
ORDER BY COUNT(*) DESC
LIMIT 5;
\" 2>/dev/null"
else
  echo -e "${BLUE}ℹ Keine Filterung in letzter Stunde${NC}"
fi

echo ""

# =============================================================================
# 4. LETZTE 5 REQUESTS
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  4. LETZTE 5 REQUESTS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \"
SELECT 
  TO_CHAR(created_at, 'HH24:MI:SS') as 'Zeit',
  CASE 
    WHEN status_code = 200 THEN '✓'
    WHEN status_code >= 400 THEN '✗'
    ELSE '?'
  END as 'S',
  SUBSTRING(model, 1, 20) as 'Model',
  duration_ms as 'ms',
  total_tokens as 'Tokens'
FROM request_logs
ORDER BY created_at DESC
LIMIT 5;
\" 2>/dev/null"

echo ""

# =============================================================================
# 5. SYSTEMRESOURCEN
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  5. SYSTEMRESOURCEN${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# CPU & Memory
BACKEND_STATS=$(ssh "${SERVER}" "docker stats llm-proxy-backend --no-stream --format '{{.CPUPerc}}|{{.MemUsage}}' 2>/dev/null || echo 'N/A|N/A'")
CPU=$(echo "$BACKEND_STATS" | cut -d'|' -f1)
MEM=$(echo "$BACKEND_STATS" | cut -d'|' -f2)

echo -e "${CYAN}Backend CPU:${NC}     ${CPU}"
echo -e "${CYAN}Backend Memory:${NC}  ${MEM}"

# Disk Space
DB_SIZE=$(ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -A -c \"
SELECT pg_size_pretty(pg_database_size('llm_proxy'));
\" 2>/dev/null || echo 'N/A'")

echo -e "${CYAN}Database Size:${NC}   ${DB_SIZE}"

echo ""

# =============================================================================
# 6. QUICK ACTIONS
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  QUICK ACTIONS${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${BLUE}Weitere Befehle:${NC}"
echo -e "  ${CYAN}./scripts/maintenance/monitor-requests.sh${NC}    - Live Request Monitor"
echo -e "  ${CYAN}./scripts/maintenance/filter-report.sh${NC}       - Detaillierter Filter Report"
echo -e "  ${CYAN}ssh ${SERVER} \"docker logs llm-proxy-backend -f\"${NC}  - Live Logs"
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Check abgeschlossen${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
