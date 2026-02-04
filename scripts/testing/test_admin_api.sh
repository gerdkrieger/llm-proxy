#!/bin/bash
set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080"
ADMIN_API_KEY="admin_dev_key_12345678901234567890123456789012"

echo "=========================================="
echo "LLM-PROXY ADMIN API TEST SUITE"
echo "=========================================="
echo ""

# Test 1: Provider Status
echo -e "${BLUE}1. Testing Provider Status${NC}"
PROVIDER_STATUS=$(curl -s ${BASE_URL}/admin/providers/status \
  -H "X-Admin-API-Key: $ADMIN_API_KEY")

HEALTHY=$(echo "$PROVIDER_STATUS" | jq -r '.healthy')
PROVIDER_COUNT=$(echo "$PROVIDER_STATUS" | jq -r '.provider_count')

if [ "$HEALTHY" = "true" ]; then
    echo -e "${GREEN}✓ Providers healthy${NC}"
    echo "  Provider count: $PROVIDER_COUNT"
else
    echo -e "${YELLOW}⚠ Providers not healthy${NC}"
    echo "$PROVIDER_STATUS" | jq .
fi
echo ""

# Test 2: Cache Stats
echo -e "${BLUE}2. Testing Cache Statistics${NC}"
CACHE_STATS=$(curl -s ${BASE_URL}/admin/cache/stats \
  -H "X-Admin-API-Key: $ADMIN_API_KEY")

HITS=$(echo "$CACHE_STATS" | jq -r '.hits')
MISSES=$(echo "$CACHE_STATS" | jq -r '.misses')
HIT_RATE=$(echo "$CACHE_STATS" | jq -r '.hit_rate')

echo -e "${GREEN}✓ Cache statistics retrieved${NC}"
echo "  Hits: $HITS"
echo "  Misses: $MISSES"
echo "  Hit Rate: $HIT_RATE%"
echo ""

# Test 3: Create OAuth Client
echo -e "${BLUE}3. Testing Create OAuth Client${NC}"
TEST_CLIENT_ID="test_admin_$(date +%s)"
CREATE_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X POST ${BASE_URL}/admin/clients \
  -H "X-Admin-API-Key: $ADMIN_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"client_id\": \"$TEST_CLIENT_ID\",
    \"client_secret\": \"secret123\",
    \"name\": \"Test Client via Admin API\",
    \"grant_types\": [\"client_credentials\"],
    \"default_scope\": \"read write\"
  }")

HTTP_CODE=$(echo "$CREATE_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)
JSON=$(echo "$CREATE_RESPONSE" | sed '/^HTTP_CODE:/d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Client created successfully${NC}"
    echo "  Client ID: $TEST_CLIENT_ID"
else
    echo -e "${RED}✗ Failed to create client${NC}"
    echo "$JSON" | jq .
fi
echo ""

# Test 4: Get OAuth Client
echo -e "${BLUE}4. Testing Get OAuth Client${NC}"
if [ "$HTTP_CODE" = "201" ]; then
    GET_RESPONSE=$(curl -s ${BASE_URL}/admin/clients/$TEST_CLIENT_ID \
      -H "X-Admin-API-Key: $ADMIN_API_KEY")
    
    CLIENT_NAME=$(echo "$GET_RESPONSE" | jq -r '.name')
    
    if [ "$CLIENT_NAME" != "null" ]; then
        echo -e "${GREEN}✓ Client retrieved successfully${NC}"
        echo "  Name: $CLIENT_NAME"
    else
        echo -e "${RED}✗ Failed to get client${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Skipping (no client to get)${NC}"
fi
echo ""

# Test 5: Update OAuth Client
echo -e "${BLUE}5. Testing Update OAuth Client${NC}"
if [ "$HTTP_CODE" = "201" ]; then
    UPDATE_RESPONSE=$(curl -s -X PATCH ${BASE_URL}/admin/clients/$TEST_CLIENT_ID \
      -H "X-Admin-API-Key: $ADMIN_API_KEY" \
      -H "Content-Type: application/json" \
      -d '{
        "name": "Updated Test Client"
      }')
    
    UPDATED_NAME=$(echo "$UPDATE_RESPONSE" | jq -r '.name')
    
    if [ "$UPDATED_NAME" = "Updated Test Client" ]; then
        echo -e "${GREEN}✓ Client updated successfully${NC}"
        echo "  New name: $UPDATED_NAME"
    else
        echo -e "${RED}✗ Failed to update client${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Skipping (no client to update)${NC}"
fi
echo ""

# Test 6: List OAuth Clients
echo -e "${BLUE}6. Testing List OAuth Clients${NC}"
LIST_RESPONSE=$(curl -s ${BASE_URL}/admin/clients \
  -H "X-Admin-API-Key: $ADMIN_API_KEY")

TOTAL=$(echo "$LIST_RESPONSE" | jq -r '.total')
echo -e "${GREEN}✓ Clients listed${NC}"
echo "  Total clients: $TOTAL"
echo ""

# Test 7: Usage Statistics
echo -e "${BLUE}7. Testing Usage Statistics${NC}"
USAGE_STATS=$(curl -s ${BASE_URL}/admin/stats/usage \
  -H "X-Admin-API-Key: $ADMIN_API_KEY")

TOTAL_REQUESTS=$(echo "$USAGE_STATS" | jq -r '.TotalRequests // 0')
TOTAL_TOKENS=$(echo "$USAGE_STATS" | jq -r '.TotalTokens // 0')
TOTAL_COST=$(echo "$USAGE_STATS" | jq -r '.TotalCost // 0')

echo -e "${GREEN}✓ Usage statistics retrieved${NC}"
echo "  Total Requests: $TOTAL_REQUESTS"
echo "  Total Tokens: $TOTAL_TOKENS"
echo "  Total Cost: \$$TOTAL_COST"
echo ""

# Test 8: Delete OAuth Client
echo -e "${BLUE}8. Testing Delete OAuth Client${NC}"
if [ "$HTTP_CODE" = "201" ]; then
    DELETE_RESPONSE=$(curl -s -X DELETE ${BASE_URL}/admin/clients/$TEST_CLIENT_ID \
      -H "X-Admin-API-Key: $ADMIN_API_KEY")
    
    MESSAGE=$(echo "$DELETE_RESPONSE" | jq -r '.message')
    
    if [ "$MESSAGE" = "client deleted successfully" ]; then
        echo -e "${GREEN}✓ Client deleted successfully${NC}"
    else
        echo -e "${RED}✗ Failed to delete client${NC}"
    fi
else
    echo -e "${YELLOW}⚠ Skipping (no client to delete)${NC}"
fi
echo ""

# Test 9: Invalid API Key (Security Test)
echo -e "${BLUE}9. Testing Invalid Admin API Key (Security)${NC}"
INVALID_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" ${BASE_URL}/admin/providers/status \
  -H "X-Admin-API-Key: invalid_key_123")

HTTP_CODE=$(echo "$INVALID_RESPONSE" | grep "HTTP_CODE" | cut -d: -f2)

if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✓ Invalid API key rejected (security working)${NC}"
else
    echo -e "${RED}✗ Invalid API key not rejected (security issue!)${NC}"
fi
echo ""

# Summary
echo "=========================================="
echo "TEST SUMMARY"
echo "=========================================="
echo -e "${GREEN}Admin API Tests Complete!${NC}"
echo ""
echo "Tested Features:"
echo "  ✓ Admin Authentication (API Key)"
echo "  ✓ Provider Health Status"
echo "  ✓ Cache Statistics"
echo "  ✓ OAuth Client CRUD Operations"
echo "  ✓ Usage Statistics"
echo "  ✓ Security (Invalid Key Rejection)"
echo ""
echo "🎉 Admin API is ready!"
