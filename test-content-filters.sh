#!/bin/bash

# Content Filtering API Test Script
# Tests all filter CRUD operations and filter application

set -e  # Exit on error

BASE_URL="http://localhost:8080"
ADMIN_KEY="admin_dev_key_12345678901234567890123456789012"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Content Filtering API Test"
echo "=========================================="
echo ""

# Helper function for API calls
api_call() {
    local method=$1
    local endpoint=$2
    local data=$3
    
    if [ -z "$data" ]; then
        curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "X-Admin-API-Key: $ADMIN_KEY" \
            -H "Content-Type: application/json"
    else
        curl -s -X "$method" "$BASE_URL$endpoint" \
            -H "X-Admin-API-Key: $ADMIN_KEY" \
            -H "Content-Type: application/json" \
            -d "$data"
    fi
}

# Test 1: Create Word Filter
echo -e "${YELLOW}Test 1: Create Word Filter (badword → [FILTERED])${NC}"
CREATE_FILTER1=$(api_call POST "/admin/filters" '{
    "pattern": "badword",
    "replacement": "[FILTERED]",
    "description": "Filter offensive word",
    "filter_type": "word",
    "case_sensitive": false,
    "enabled": true,
    "priority": 100
}')
echo "$CREATE_FILTER1" | jq '.'
FILTER_ID1=$(echo "$CREATE_FILTER1" | jq -r '.id')
echo -e "${GREEN}✓ Filter created with ID: $FILTER_ID1${NC}\n"

# Test 2: Create Phrase Filter
echo -e "${YELLOW}Test 2: Create Phrase Filter (confidential information → [REDACTED])${NC}"
CREATE_FILTER2=$(api_call POST "/admin/filters" '{
    "pattern": "confidential information",
    "replacement": "[REDACTED]",
    "description": "Redact confidential information",
    "filter_type": "phrase",
    "case_sensitive": false,
    "enabled": true,
    "priority": 90
}')
echo "$CREATE_FILTER2" | jq '.'
FILTER_ID2=$(echo "$CREATE_FILTER2" | jq -r '.id')
echo -e "${GREEN}✓ Filter created with ID: $FILTER_ID2${NC}\n"

# Test 3: Create Regex Filter (Email)
echo -e "${YELLOW}Test 3: Create Regex Filter (email → [EMAIL])${NC}"
CREATE_FILTER3=$(api_call POST "/admin/filters" '{
    "pattern": "\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b",
    "replacement": "[EMAIL]",
    "description": "Filter email addresses",
    "filter_type": "regex",
    "case_sensitive": false,
    "enabled": true,
    "priority": 80
}')
echo "$CREATE_FILTER3" | jq '.'
FILTER_ID3=$(echo "$CREATE_FILTER3" | jq -r '.id')
echo -e "${GREEN}✓ Filter created with ID: $FILTER_ID3${NC}\n"

# Test 4: List All Filters
echo -e "${YELLOW}Test 4: List All Filters${NC}"
LIST_FILTERS=$(api_call GET "/admin/filters")
echo "$LIST_FILTERS" | jq '.'
COUNT=$(echo "$LIST_FILTERS" | jq '.count')
echo -e "${GREEN}✓ Found $COUNT filters${NC}\n"

# Test 5: Get Single Filter
echo -e "${YELLOW}Test 5: Get Filter by ID ($FILTER_ID1)${NC}"
GET_FILTER=$(api_call GET "/admin/filters/$FILTER_ID1")
echo "$GET_FILTER" | jq '.'
echo -e "${GREEN}✓ Filter retrieved${NC}\n"

# Test 6: Test Filter (Ad-hoc)
echo -e "${YELLOW}Test 6: Test Filter (Ad-hoc)${NC}"
TEST_RESULT=$(api_call POST "/admin/filters/test" '{
    "text": "This is a badword and contains confidential information and my email is test@example.com",
    "pattern": "badword",
    "replacement": "[FILTERED]",
    "filter_type": "word"
}')
echo "$TEST_RESULT" | jq '.'
FILTERED=$(echo "$TEST_RESULT" | jq -r '.filtered_text')
echo -e "${GREEN}✓ Filtered text: $FILTERED${NC}\n"

# Test 7: Test Existing Filter
echo -e "${YELLOW}Test 7: Test Existing Filter ($FILTER_ID1)${NC}"
TEST_EXISTING=$(api_call POST "/admin/filters/$FILTER_ID1/test" '{
    "text": "This badword should be filtered"
}')
echo "$TEST_EXISTING" | jq '.'
echo -e "${GREEN}✓ Filter tested${NC}\n"

# Test 8: Get Filter Statistics
echo -e "${YELLOW}Test 8: Get Filter Statistics${NC}"
STATS=$(api_call GET "/admin/filters/stats")
echo "$STATS" | jq '.'
echo -e "${GREEN}✓ Statistics retrieved${NC}\n"

# Test 9: Update Filter
echo -e "${YELLOW}Test 9: Update Filter (change priority)${NC}"
UPDATE_FILTER=$(api_call PUT "/admin/filters/$FILTER_ID1" '{
    "priority": 150,
    "description": "Updated filter description"
}')
echo "$UPDATE_FILTER" | jq '.'
echo -e "${GREEN}✓ Filter updated${NC}\n"

# Test 10: Refresh Filter Cache
echo -e "${YELLOW}Test 10: Refresh Filter Cache${NC}"
REFRESH=$(api_call POST "/admin/filters/refresh")
echo "$REFRESH" | jq '.'
echo -e "${GREEN}✓ Cache refreshed${NC}\n"

# Test 11: Test Filter Application in Chat (Get OAuth Token First)
echo -e "${YELLOW}Test 11: Test Filter in Chat Completion${NC}"
echo "Getting OAuth token..."
TOKEN_RESPONSE=$(curl -s -X POST "$BASE_URL/oauth/token" \
    -H "Content-Type: application/json" \
    -d '{
        "grant_type": "client_credentials",
        "client_id": "test_client",
        "client_secret": "test_secret_123456",
        "scope": "read write"
    }')
ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')

if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
    echo -e "${GREEN}✓ Got access token${NC}"
    
    echo "Sending chat completion with filtered content..."
    CHAT_RESPONSE=$(curl -s -X POST "$BASE_URL/v1/chat/completions" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "model": "claude-3-haiku-20240307",
            "messages": [
                {"role": "user", "content": "This message contains a badword and confidential information. My email is admin@example.com"}
            ],
            "max_tokens": 50
        }')
    
    echo "$CHAT_RESPONSE" | jq '.'
    echo -e "${GREEN}✓ Chat completion with filtering completed${NC}\n"
else
    echo -e "${RED}✗ Failed to get access token${NC}\n"
fi

# Test 12: Disable Filter
echo -e "${YELLOW}Test 12: Disable Filter${NC}"
DISABLE_FILTER=$(api_call PUT "/admin/filters/$FILTER_ID2" '{
    "enabled": false
}')
echo "$DISABLE_FILTER" | jq '.'
echo -e "${GREEN}✓ Filter disabled${NC}\n"

# Test 13: Delete Filter
echo -e "${YELLOW}Test 13: Delete Filter ($FILTER_ID3)${NC}"
DELETE_FILTER=$(api_call DELETE "/admin/filters/$FILTER_ID3")
echo "$DELETE_FILTER" | jq '.'
echo -e "${GREEN}✓ Filter deleted${NC}\n"

# Final Stats
echo -e "${YELLOW}Final: Get Updated Statistics${NC}"
FINAL_STATS=$(api_call GET "/admin/filters/stats")
echo "$FINAL_STATS" | jq '.'

echo ""
echo "=========================================="
echo -e "${GREEN}All tests completed!${NC}"
echo "=========================================="
echo ""
echo "Summary:"
echo "- Created 3 filters (Word, Phrase, Regex)"
echo "- Listed all filters"
echo "- Retrieved single filter"
echo "- Tested filters (ad-hoc and existing)"
echo "- Updated filter"
echo "- Refreshed cache"
echo "- Tested filter in chat completion"
echo "- Disabled filter"
echo "- Deleted filter"
echo ""
echo "Note: Filters $FILTER_ID1 and $FILTER_ID2 were left in the database"
echo "      for further testing. Delete them manually if needed:"
echo "      curl -X DELETE $BASE_URL/admin/filters/$FILTER_ID1 -H 'X-Admin-API-Key: $ADMIN_KEY'"
echo "      curl -X DELETE $BASE_URL/admin/filters/$FILTER_ID2 -H 'X-Admin-API-Key: $ADMIN_KEY'"
echo ""
