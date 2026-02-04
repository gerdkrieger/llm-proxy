#!/bin/bash

# Test script to demonstrate all content filters working together

echo "========================================"
echo "Content Filter Integration Test"
echo "========================================"
echo ""

# API Key
ADMIN_KEY="admin_dev_key_12345678901234567890123456789012"
BASE_URL="http://localhost:8080"

# Test text with all filter patterns
TEST_TEXT="This is a comprehensive test message. First, we have some badword and damn language, plus shit content. I work on Project Phoenix which is top secret and contains confidential information. You can reach me at john.doe@example.com or call 0123-456789. My password is admin123 and the secret key is sk-abc123. Payment can be made to card 1234 5678 9012 3456. We're competing with CompetitorX in the market."

echo "Original Text:"
echo "----------------------------------------"
echo "$TEST_TEXT"
echo ""
echo "========================================"
echo ""

# Test each filter individually to show what it catches
echo "Testing Individual Filters:"
echo "----------------------------------------"

# Filter 1: badword
echo "1. Word Filter (badword):"
curl -s -X POST "$BASE_URL/admin/filters/1/test" \
  -H "Content-Type: application/json" \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -d "{\"text\": \"$TEST_TEXT\"}" | jq -r '.filtered_text'
echo ""

# Filter 2: damn
echo "2. Word Filter (damn):"
curl -s -X POST "$BASE_URL/admin/filters/2/test" \
  -H "Content-Type: application/json" \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -d "{\"text\": \"$TEST_TEXT\"}" | jq -r '.filtered_text'
echo ""

# Filter 7: Email
echo "3. Regex Filter (Email):"
curl -s -X POST "$BASE_URL/admin/filters/7/test" \
  -H "Content-Type: application/json" \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -d "{\"text\": \"$TEST_TEXT\"}" | jq -r '.filtered_text'
echo ""

# Filter 9: Credit Card
echo "4. Regex Filter (Credit Card):"
curl -s -X POST "$BASE_URL/admin/filters/9/test" \
  -H "Content-Type: application/json" \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -d "{\"text\": \"$TEST_TEXT\"}" | jq -r '.filtered_text'
echo ""

# Filter 5: Project Phoenix (phrase)
echo "5. Phrase Filter (Project Phoenix):"
curl -s -X POST "$BASE_URL/admin/filters/5/test" \
  -H "Content-Type: application/json" \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  -d "{\"text\": \"$TEST_TEXT\"}" | jq -r '.filtered_text'
echo ""

echo "========================================"
echo ""
echo "Current Filter Statistics:"
echo "----------------------------------------"
curl -s "$BASE_URL/admin/filters/stats" \
  -H "X-Admin-API-Key: $ADMIN_KEY" | jq .
echo ""

echo "========================================"
echo "All Filters List (Count: $(curl -s "$BASE_URL/admin/filters" -H "X-Admin-API-Key: $ADMIN_KEY" | jq '.count')):"
echo "----------------------------------------"
curl -s "$BASE_URL/admin/filters" \
  -H "X-Admin-API-Key: $ADMIN_KEY" | jq -r '.filters[] | "\(.id). [\(.filter_type)] \(.pattern) -> \(.replacement) (Priority: \(.priority))"'
echo ""

echo "========================================"
echo "Test Complete!"
echo "========================================"
