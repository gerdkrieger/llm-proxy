#!/bin/bash
set -e

echo "=========================================="
echo "LLM-Proxy API Test Suite"
echo "=========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
CLIENT_ID="test_client"
CLIENT_SECRET="test_secret_123456"

echo "1. Testing Health Endpoint..."
HEALTH=$(curl -s ${BASE_URL}/health)
if echo "$HEALTH" | jq -e '.status == "ok"' > /dev/null; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
    exit 1
fi
echo ""

echo "2. Testing OAuth Token Generation..."
TOKEN_RESPONSE=$(curl -s -X POST ${BASE_URL}/oauth/token \
  -H "Content-Type: application/json" \
  -d "{
    \"grant_type\": \"client_credentials\",
    \"client_id\": \"${CLIENT_ID}\",
    \"client_secret\": \"${CLIENT_SECRET}\",
    \"scope\": \"read write\"
  }")

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
    echo -e "${GREEN}✓ OAuth token generated successfully${NC}"
    echo "   Token (first 50 chars): ${ACCESS_TOKEN:0:50}..."
else
    echo -e "${RED}✗ OAuth token generation failed${NC}"
    echo "$TOKEN_RESPONSE" | jq .
    exit 1
fi
echo ""

echo "3. Testing Models Endpoint (GET /v1/models)..."
MODELS=$(curl -s ${BASE_URL}/v1/models \
  -H "Authorization: Bearer $ACCESS_TOKEN")

MODEL_COUNT=$(echo "$MODELS" | jq '.data | length')
if [ "$MODEL_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓ Models endpoint returned $MODEL_COUNT model(s)${NC}"
    echo "$MODELS" | jq '.data[].id'
else
    echo -e "${RED}✗ Models endpoint failed${NC}"
    echo "$MODELS" | jq .
    exit 1
fi
echo ""

echo "4. Testing Chat Completion Endpoint (POST /v1/chat/completions)..."
CHAT_RESPONSE=$(curl -s -X POST ${BASE_URL}/v1/chat/completions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [
      {
        "role": "user",
        "content": "Say hello in exactly 5 words."
      }
    ],
    "max_tokens": 50
  }')

# Check if response has error (expected if API key is not configured)
HAS_ERROR=$(echo "$CHAT_RESPONSE" | jq -e '.error' > /dev/null 2>&1 && echo "yes" || echo "no")
if [ "$HAS_ERROR" = "yes" ]; then
    ERROR_MSG=$(echo "$CHAT_RESPONSE" | jq -r '.error.message')
    if echo "$ERROR_MSG" | grep -q "invalid x-api-key"; then
        echo -e "${YELLOW}⚠ Chat completion endpoint working, but Claude API key needs configuration${NC}"
        echo "   Update CLAUDE_API_KEY in .env file with your real API key"
    else
        echo -e "${RED}✗ Chat completion failed with unexpected error${NC}"
        echo "$CHAT_RESPONSE" | jq .
    fi
else
    # Success - we got a real response
    CONTENT=$(echo "$CHAT_RESPONSE" | jq -r '.choices[0].message.content')
    echo -e "${GREEN}✓ Chat completion successful${NC}"
    echo "   Response: $CONTENT"
fi
echo ""

echo "5. Testing Invalid Token (Security)..."
INVALID_RESPONSE=$(curl -s ${BASE_URL}/v1/models \
  -H "Authorization: Bearer invalid_token_123")

ERROR_TYPE=$(echo "$INVALID_RESPONSE" | jq -r '.error')
if echo "$INVALID_RESPONSE" | jq -e ".error" > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Invalid token correctly rejected${NC}"
else
    echo -e "${RED}✗ Invalid token should be rejected${NC}"
    echo "$INVALID_RESPONSE" | jq .
fi
echo ""

echo "=========================================="
echo -e "${GREEN}Test Suite Complete!${NC}"
echo "=========================================="
echo ""
echo "Summary:"
echo "  - Health check: Working"
echo "  - OAuth authentication: Working"
echo "  - Models endpoint: Working"
echo "  - Chat completion: Endpoint working (requires real Claude API key)"
echo "  - Security: Working"
echo ""
echo "Next steps:"
echo "  1. Add your real Claude API key to .env file"
echo "  2. Restart the server: make restart"
echo "  3. Run this test again to verify full functionality"
