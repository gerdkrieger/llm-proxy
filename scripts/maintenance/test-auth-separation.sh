#!/bin/bash
set -euo pipefail

##############################################################################
# LLM-Proxy Authentication Separation Test
##############################################################################
# Purpose: Verify that admin keys and client keys are properly isolated
# - Admin keys should access /admin/* but NOT /v1/*
# - Client keys should access /v1/* but NOT /admin/*
# - Invalid keys should be rejected everywhere
##############################################################################

# Configuration
BASE_URL="${LLM_PROXY_URL:-https://llmproxy.aitrail.ch}"
ADMIN_KEY="${LLM_PROXY_ADMIN_KEY:-admin_dev_key_12345678901234567890123456789012}"
CLIENT_KEY="${LLM_PROXY_CLIENT_KEY:-sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
PASSED=0
FAILED=0
TOTAL=0

##############################################################################
# Helper Functions
##############################################################################

print_header() {
  echo ""
  echo -e "${BLUE}========================================${NC}"
  echo -e "${BLUE}$1${NC}"
  echo -e "${BLUE}========================================${NC}"
  echo ""
}

print_test() {
  echo -e "${YELLOW}Test $1:${NC} $2"
}

print_pass() {
  echo -e "${GREEN}✅ PASS:${NC} $1"
  ((PASSED++))
  ((TOTAL++))
  echo ""
}

print_fail() {
  echo -e "${RED}❌ FAIL:${NC} $1"
  ((FAILED++))
  ((TOTAL++))
  echo ""
}

print_summary() {
  echo ""
  echo -e "${BLUE}========================================${NC}"
  echo -e "${BLUE}Test Summary${NC}"
  echo -e "${BLUE}========================================${NC}"
  echo -e "Total Tests: $TOTAL"
  echo -e "${GREEN}Passed: $PASSED${NC}"
  if [ $FAILED -gt 0 ]; then
    echo -e "${RED}Failed: $FAILED${NC}"
  else
    echo -e "Failed: $FAILED"
  fi
  
  if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}🎉 All tests passed! Authentication separation is working correctly.${NC}"
    return 0
  else
    echo ""
    echo -e "${RED}⚠️  Some tests failed. Please review authentication configuration.${NC}"
    return 1
  fi
}

test_endpoint() {
  local description="$1"
  local method="$2"
  local url="$3"
  local header="$4"
  local expected_status="$5"
  local reason="$6"
  
  print_test "$TOTAL" "$description"
  
  # Make request and capture status code (with 10 second timeout)
  local status
  status=$(timeout 10 curl -s -o /dev/null -w "%{http_code}" -X "$method" -H "$header" "$url" 2>/dev/null || echo "000")
  
  # Check result
  if [ "$status" = "$expected_status" ]; then
    print_pass "$reason (got $status)"
  else
    print_fail "Expected $expected_status but got $status - $reason"
  fi
}

##############################################################################
# Main Test Suite
##############################################################################

print_header "LLM-Proxy Authentication Separation Test"

echo "Configuration:"
echo "  Base URL: $BASE_URL"
echo "  Admin Key: ${ADMIN_KEY:0:20}...${ADMIN_KEY: -8}"
echo "  Client Key: ${CLIENT_KEY:0:25}...${CLIENT_KEY: -8}"
echo ""

##############################################################################
# Section 1: Admin Key Tests
##############################################################################

print_header "Section 1: Admin Key Authentication"

# Test 1: Admin key should access /admin/clients
test_endpoint \
  "Admin key → GET /admin/clients" \
  "GET" \
  "$BASE_URL/admin/clients" \
  "X-Admin-API-Key: $ADMIN_KEY" \
  "200" \
  "Admin key can access admin endpoints"

# Test 2: Admin key with Authorization Bearer should also work
test_endpoint \
  "Admin key (Bearer) → GET /admin/clients" \
  "GET" \
  "$BASE_URL/admin/clients" \
  "Authorization: Bearer $ADMIN_KEY" \
  "200" \
  "Admin key works with Authorization header"

# Test 3: Admin key should NOT access /v1/models
test_endpoint \
  "Admin key → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: Bearer $ADMIN_KEY" \
  "401" \
  "Admin key rejected from LLM API endpoints"

# Test 4: Admin key should NOT access /v1/chat/completions
test_endpoint \
  "Admin key → POST /v1/chat/completions" \
  "POST" \
  "$BASE_URL/v1/chat/completions" \
  "Authorization: Bearer $ADMIN_KEY" \
  "401" \
  "Admin key rejected from chat completions"

##############################################################################
# Section 2: Client Key Tests
##############################################################################

print_header "Section 2: Client Key Authentication"

# Test 5: Client key should access /v1/models
test_endpoint \
  "Client key → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: Bearer $CLIENT_KEY" \
  "200" \
  "Client key can access LLM API endpoints"

# Test 6: Client key should NOT access /admin/clients
test_endpoint \
  "Client key → GET /admin/clients" \
  "GET" \
  "$BASE_URL/admin/clients" \
  "Authorization: Bearer $CLIENT_KEY" \
  "401" \
  "Client key rejected from admin endpoints"

# Test 7: Client key should NOT work with X-Admin-API-Key header
test_endpoint \
  "Client key (X-Admin-API-Key) → GET /admin/clients" \
  "GET" \
  "$BASE_URL/admin/clients" \
  "X-Admin-API-Key: $CLIENT_KEY" \
  "401" \
  "Client key doesn't work as admin key"

# Test 8: Client key should NOT access /admin/filters/stats
test_endpoint \
  "Client key → GET /admin/filters/stats" \
  "GET" \
  "$BASE_URL/admin/filters/stats" \
  "Authorization: Bearer $CLIENT_KEY" \
  "401" \
  "Client key rejected from filter stats"

##############################################################################
# Section 3: Invalid Key Tests
##############################################################################

print_header "Section 3: Invalid Key Tests"

# Test 9: Invalid key should be rejected from /v1/models
test_endpoint \
  "Invalid key → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: Bearer invalid-key-12345" \
  "401" \
  "Invalid key rejected from LLM API"

# Test 10: Invalid key should be rejected from /admin/clients
test_endpoint \
  "Invalid key → GET /admin/clients" \
  "GET" \
  "$BASE_URL/admin/clients" \
  "X-Admin-API-Key: invalid-key-12345" \
  "401" \
  "Invalid key rejected from admin API"

# Test 11: Client key without sk-llm-proxy- prefix should be rejected
test_endpoint \
  "Malformed client key → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: Bearer openwebui-2026-01-30-key-abc123" \
  "401" \
  "Key without required prefix rejected"

# Test 12: Empty Authorization header should be rejected
test_endpoint \
  "Empty auth → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: Bearer" \
  "401" \
  "Empty authorization rejected"

##############################################################################
# Section 4: Cross-Contamination Tests
##############################################################################

print_header "Section 4: Cross-Contamination Tests"

# Test 13: Admin key in Authorization + Client endpoint
test_endpoint \
  "Admin key → GET /v1/chat/completions" \
  "GET" \
  "$BASE_URL/v1/chat/completions" \
  "Authorization: Bearer $ADMIN_KEY" \
  "401" \
  "Admin key cannot be used for chat API"

# Test 14: Client key in X-Admin-API-Key + Admin endpoint
test_endpoint \
  "Client key (X-Admin) → GET /admin/system/stats" \
  "GET" \
  "$BASE_URL/admin/system/stats" \
  "X-Admin-API-Key: $CLIENT_KEY" \
  "401" \
  "Client key cannot be used for admin API"

# Test 15: Verify admin key doesn't accidentally work with sk-llm-proxy- prefix
FAKE_CLIENT_KEY="sk-llm-proxy-${ADMIN_KEY:6}"
test_endpoint \
  "Admin key with fake prefix → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: Bearer $FAKE_CLIENT_KEY" \
  "401" \
  "Adding prefix to admin key doesn't grant access"

##############################################################################
# Section 5: Scope Tests (if applicable)
##############################################################################

print_header "Section 5: Additional Security Tests"

# Test 16: No Authorization header at all
test_endpoint \
  "No auth → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "User-Agent: TestScript" \
  "401" \
  "Missing authorization rejected"

# Test 17: Wrong auth scheme
test_endpoint \
  "Basic auth → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: Basic dGVzdDp0ZXN0" \
  "401" \
  "Wrong auth scheme rejected"

# Test 18: Case sensitivity check (Bearer vs bearer)
test_endpoint \
  "Lowercase bearer → GET /v1/models" \
  "GET" \
  "$BASE_URL/v1/models" \
  "Authorization: bearer $CLIENT_KEY" \
  "401" \
  "Case-sensitive Bearer keyword"

##############################################################################
# Print Summary and Exit
##############################################################################

print_summary
exit $?
