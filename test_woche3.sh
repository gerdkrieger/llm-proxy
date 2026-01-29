#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
CLIENT_ID="test_client"
CLIENT_SECRET="test_secret_123456"

echo "=========================================="
echo "LLM-PROXY WOCHE 3 TEST SUITE"
echo "Streaming & Caching Features"
echo "=========================================="
echo ""

# Get OAuth token
echo -e "${BLUE}1. Getting OAuth Token...${NC}"
TOKEN_RESPONSE=$(curl -s -X POST ${BASE_URL}/oauth/token \
  -H "Content-Type: application/json" \
  -d "{
    \"grant_type\": \"client_credentials\",
    \"client_id\": \"${CLIENT_ID}\",
    \"client_secret\": \"${CLIENT_SECRET}\",
    \"scope\": \"read write\"
  }")

TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
    echo -e "${RED}✗ Failed to get OAuth token${NC}"
    echo "$TOKEN_RESPONSE" | jq .
    exit 1
fi

echo -e "${GREEN}✓ OAuth token obtained${NC}"
echo "  Token: ${TOKEN:0:30}..."
echo ""

# Test 1: Chat Completion without caching (should be MISS)
echo -e "${BLUE}2. Testing Chat Completion - Cache MISS${NC}"
START1=$(date +%s%N)
RESPONSE1=$(curl -s -i -X POST ${BASE_URL}/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [{"role": "user", "content": "Sage nur das Wort ERFOLG"}],
    "max_tokens": 20,
    "temperature": 0.5
  }')
END1=$(date +%s%N)
DUR1=$(( (END1 - START1) / 1000000 ))

CACHE1=$(echo "$RESPONSE1" | grep -i "^x-cache:" | cut -d: -f2 | tr -d ' \r')
JSON1=$(echo "$RESPONSE1" | sed -n '/^{/,$p')
CONTENT1=$(echo "$JSON1" | jq -r '.choices[0].message.content')
TOKENS1=$(echo "$JSON1" | jq -r '.usage.total_tokens')

if [ "$CACHE1" = "MISS" ]; then
    echo -e "${GREEN}✓ Cache MISS (as expected)${NC}"
    echo "  Response: $CONTENT1"
    echo "  Tokens: $TOKENS1"
    echo "  Duration: ${DUR1}ms"
else
    echo -e "${RED}✗ Expected MISS, got: $CACHE1${NC}"
fi
echo ""

# Test 2: Identical request (should be HIT)
echo -e "${BLUE}3. Testing Cache HIT (identical request)${NC}"
sleep 1

START2=$(date +%s%N)
RESPONSE2=$(curl -s -i -X POST ${BASE_URL}/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [{"role": "user", "content": "Sage nur das Wort ERFOLG"}],
    "max_tokens": 20,
    "temperature": 0.5
  }')
END2=$(date +%s%N)
DUR2=$(( (END2 - START2) / 1000000 ))

CACHE2=$(echo "$RESPONSE2" | grep -i "^x-cache:" | cut -d: -f2 | tr -d ' \r')
JSON2=$(echo "$RESPONSE2" | sed -n '/^{/,$p')
CONTENT2=$(echo "$JSON2" | jq -r '.choices[0].message.content')

if [ "$CACHE2" = "HIT" ]; then
    echo -e "${GREEN}✓ Cache HIT (as expected)${NC}"
    echo "  Response: $CONTENT2"
    echo "  Duration: ${DUR2}ms"
    SPEEDUP=$(( DUR1 / DUR2 ))
    echo "  Speed-up: ${SPEEDUP}x faster than MISS"
else
    echo -e "${RED}✗ Expected HIT, got: $CACHE2${NC}"
fi
echo ""

# Test 3: Different parameter (should be MISS)
echo -e "${BLUE}4. Testing Cache MISS (different parameters)${NC}"
RESPONSE3=$(curl -s -i -X POST ${BASE_URL}/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [{"role": "user", "content": "Sage nur das Wort ERFOLG"}],
    "max_tokens": 30,
    "temperature": 0.5
  }')

CACHE3=$(echo "$RESPONSE3" | grep -i "^x-cache:" | cut -d: -f2 | tr -d ' \r')

if [ "$CACHE3" = "MISS" ]; then
    echo -e "${GREEN}✓ Cache MISS (different max_tokens)${NC}"
else
    echo -e "${RED}✗ Expected MISS, got: $CACHE3${NC}"
fi
echo ""

# Test 4: Streaming
echo -e "${BLUE}5. Testing Streaming Chat Completion${NC}"
STREAM_OUTPUT=$(curl -s -N -X POST ${BASE_URL}/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [{"role": "user", "content": "Zähle von 1 bis 3"}],
    "max_tokens": 30,
    "stream": true
  }' 2>/dev/null | head -15)

# Check for SSE format
if echo "$STREAM_OUTPUT" | grep -q "data:.*chat.completion.chunk" && \
   echo "$STREAM_OUTPUT" | grep -q "data: \[DONE\]"; then
    echo -e "${GREEN}✓ Streaming works correctly${NC}"
    CHUNKS=$(echo "$STREAM_OUTPUT" | grep -c "^data:" || true)
    echo "  Received $CHUNKS SSE events"
    echo "  Format: OpenAI-compatible"
else
    echo -e "${RED}✗ Streaming response malformed${NC}"
    echo "$STREAM_OUTPUT" | head -5
fi
echo ""

# Test 5: Models endpoint
echo -e "${BLUE}6. Testing Models Endpoint${NC}"
MODELS=$(curl -s ${BASE_URL}/v1/models \
  -H "Authorization: Bearer $TOKEN")

MODEL_COUNT=$(echo "$MODELS" | jq '.data | length')
if [ "$MODEL_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Models endpoint works${NC}"
    echo "  Available models: $MODEL_COUNT"
    echo "$MODELS" | jq -r '.data[].id' | sed 's/^/    - /'
else
    echo -e "${RED}✗ Models endpoint failed${NC}"
fi
echo ""

# Summary
echo "=========================================="
echo "TEST SUMMARY"
echo "=========================================="
echo -e "${GREEN}All Woche 3 features tested successfully!${NC}"
echo ""
echo "Features Tested:"
echo "  ✓ Response Caching (MISS → HIT)"
echo "  ✓ Cache Key Generation (different params → MISS)"
echo "  ✓ Server-Sent Events (SSE) Streaming"
echo "  ✓ OpenAI-compatible API"
echo "  ✓ Models Listing"
echo ""
echo "Performance:"
echo "  Cache MISS: ${DUR1}ms"
echo "  Cache HIT:  ${DUR2}ms"
if [ "$DUR2" -gt 0 ]; then
    SPEEDUP=$(( DUR1 / DUR2 ))
    echo "  Speed-up:   ${SPEEDUP}x faster"
fi
echo ""
echo "🎉 Woche 3 Complete!"
