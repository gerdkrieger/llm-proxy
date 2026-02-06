#!/bin/bash
# =============================================================================
# LLM-PROXY REQUEST MONITOR
# =============================================================================
# Zeigt alle Requests in Echtzeit mit Filterung
# =============================================================================

set -euo pipefail

# Farben
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Server
SERVER="${1:-openweb}"

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  LLM-PROXY REQUEST MONITOR${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo -e "${YELLOW}Überwache Requests auf: ${SERVER}${NC}"
echo -e "${YELLOW}Drücke Ctrl+C zum Beenden${NC}"
echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Live Logs mit farbiger Ausgabe
ssh "${SERVER}" "docker logs llm-proxy-backend -f --tail 50" 2>&1 | while IFS= read -r line; do
  # Parse JSON log entry
  if echo "$line" | jq -e . >/dev/null 2>&1; then
    LEVEL=$(echo "$line" | jq -r '.level // empty')
    MESSAGE=$(echo "$line" | jq -r '.message // empty')
    REQUEST_ID=$(echo "$line" | jq -r '.request_id // empty')
    METHOD=$(echo "$line" | jq -r '.method // empty')
    PATH=$(echo "$line" | jq -r '.path // empty')
    STATUS=$(echo "$line" | jq -r '.status // empty')
    DURATION=$(echo "$line" | jq -r '.duration_ms // empty')
    MODEL=$(echo "$line" | jq -r '.model // empty')
    
    # Timestamp
    TIMESTAMP=$(date +"%H:%M:%S")
    
    # Color based on level and status
    if [[ "$LEVEL" == "error" ]]; then
      COLOR=$RED
      PREFIX="❌ ERROR"
    elif [[ "$LEVEL" == "warn" ]]; then
      COLOR=$YELLOW
      PREFIX="⚠️  WARN"
    elif [[ "$STATUS" == "200" ]]; then
      COLOR=$GREEN
      PREFIX="✅ SUCCESS"
    elif [[ "$STATUS" != "" ]] && [[ "$STATUS" != "null" ]]; then
      COLOR=$RED
      PREFIX="❌ FAILED"
    else
      COLOR=$BLUE
      PREFIX="ℹ️  INFO"
    fi
    
    # Check for filter activity
    if echo "$MESSAGE" | grep -iq "filter\|redact\|pii"; then
      COLOR=$YELLOW
      PREFIX="🔒 FILTERED"
    fi
    
    # Output formatted line
    if [[ "$REQUEST_ID" != "" && "$REQUEST_ID" != "null" ]]; then
      echo -e "${COLOR}[${TIMESTAMP}] ${PREFIX}${NC}"
      echo -e "  ${COLOR}├─${NC} Request: ${REQUEST_ID}"
      [[ "$METHOD" != "" && "$METHOD" != "null" ]] && echo -e "  ${COLOR}├─${NC} Method: ${METHOD} ${PATH}"
      [[ "$MODEL" != "" && "$MODEL" != "null" ]] && echo -e "  ${COLOR}├─${NC} Model: ${MODEL}"
      [[ "$STATUS" != "" && "$STATUS" != "null" ]] && echo -e "  ${COLOR}├─${NC} Status: ${STATUS}"
      [[ "$DURATION" != "" && "$DURATION" != "null" ]] && echo -e "  ${COLOR}├─${NC} Duration: ${DURATION}ms"
      echo -e "  ${COLOR}└─${NC} ${MESSAGE}"
      echo ""
    elif [[ "$MESSAGE" != "" && "$MESSAGE" != "null" ]]; then
      echo -e "${COLOR}[${TIMESTAMP}] ${PREFIX}${NC} ${MESSAGE}"
    fi
  else
    # Non-JSON log line
    echo "$line"
  fi
done
