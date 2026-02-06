#!/bin/bash
# =============================================================================
# LLM-PROXY FILTER REPORT
# =============================================================================
# Zeigt einen detaillierten Report über alle gefilterten Inhalte
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
DAYS="${2:-7}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  LLM-PROXY FILTER REPORT${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${CYAN}Server:${NC} ${SERVER}"
echo -e "${CYAN}Zeitraum:${NC} Letzte ${DAYS} Tage"
echo ""

# =============================================================================
# 1. ZUSAMMENFASSUNG
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  1. ZUSAMMENFASSUNG${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -c \"
SELECT 
  COUNT(*) as total_matches,
  COUNT(DISTINCT request_id) as unique_requests,
  COUNT(DISTINCT pattern) as unique_patterns
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '${DAYS} days';
\"" | while read -r line; do
  if [[ -n "$line" ]]; then
    TOTAL=$(echo "$line" | awk '{print $1}')
    REQUESTS=$(echo "$line" | awk '{print $3}')
    PATTERNS=$(echo "$line" | awk '{print $5}')
    
    echo -e "${GREEN}✓ Total Filter-Matches:${NC}     ${TOTAL}"
    echo -e "${GREEN}✓ Betroffene Requests:${NC}      ${REQUESTS}"
    echo -e "${GREEN}✓ Verschiedene Pattern:${NC}     ${PATTERNS}"
  fi
done

echo ""

# =============================================================================
# 2. FILTER-MATCHES NACH TYP
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  2. FILTER-MATCHES NACH TYP${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \"
SELECT 
  pattern as 'PII-Typ',
  COUNT(*) as 'Anzahl',
  COUNT(DISTINCT request_id) as 'Requests',
  TO_CHAR(MAX(created_at), 'YYYY-MM-DD HH24:MI') as 'Letzter Match'
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '${DAYS} days'
GROUP BY pattern
ORDER BY COUNT(*) DESC;
\""

echo ""

# =============================================================================
# 3. TÄGLICHE ÜBERSICHT
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  3. TÄGLICHE ÜBERSICHT${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \"
SELECT 
  TO_CHAR(DATE(created_at), 'YYYY-MM-DD') as 'Datum',
  COUNT(*) as 'Matches',
  COUNT(DISTINCT request_id) as 'Requests',
  COUNT(DISTINCT pattern) as 'Patterns'
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '${DAYS} days'
GROUP BY DATE(created_at)
ORDER BY DATE(created_at) DESC;
\""

echo ""

# =============================================================================
# 4. NACH PROVIDER
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  4. NACH PROVIDER${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \"
SELECT 
  COALESCE(provider, 'Unknown') as 'Provider',
  COUNT(*) as 'Matches',
  COUNT(DISTINCT request_id) as 'Requests'
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '${DAYS} days'
GROUP BY provider
ORDER BY COUNT(*) DESC;
\""

echo ""

# =============================================================================
# 5. LETZTE 20 FILTER-MATCHES (DETAILLIERT)
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  5. LETZTE 20 FILTER-MATCHES (DETAILLIERT)${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \"
SELECT 
  TO_CHAR(created_at, 'MM-DD HH24:MI:SS') as 'Zeit',
  pattern as 'Pattern',
  replacement as 'Ersetzt',
  SUBSTRING(request_id, 1, 20) as 'Request-ID',
  model as 'Model'
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '${DAYS} days'
ORDER BY created_at DESC
LIMIT 20;
\""

echo ""

# =============================================================================
# 6. STATISTIK: REQUESTS MIT VS OHNE FILTERUNG
# =============================================================================
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}  6. REQUESTS MIT VS OHNE FILTERUNG${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

TOTAL_REQUESTS=$(ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -A -c \"
SELECT COUNT(*) FROM request_logs WHERE created_at > NOW() - INTERVAL '${DAYS} days';
\"")

FILTERED_REQUESTS=$(ssh "${SERVER}" "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -A -c \"
SELECT COUNT(DISTINCT request_id) FROM filter_matches WHERE created_at > NOW() - INTERVAL '${DAYS} days';
\"")

UNFILTERED_REQUESTS=$((TOTAL_REQUESTS - FILTERED_REQUESTS))

if [[ "$TOTAL_REQUESTS" -gt 0 ]]; then
  PERCENTAGE=$(echo "scale=1; $FILTERED_REQUESTS * 100 / $TOTAL_REQUESTS" | bc)
else
  PERCENTAGE="0.0"
fi

echo -e "${CYAN}Total Requests:${NC}          ${TOTAL_REQUESTS}"
echo -e "${GREEN}Gefilterte Requests:${NC}     ${FILTERED_REQUESTS} (${PERCENTAGE}%)"
echo -e "${BLUE}Ungefilterte Requests:${NC}   ${UNFILTERED_REQUESTS}"

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✓ Report abgeschlossen${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
