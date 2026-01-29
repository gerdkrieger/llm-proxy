#!/bin/bash
# Test Script für LLM-Proxy API

echo "🧪 LLM-Proxy API Test"
echo "===================="
echo ""

# 1. Get OAuth Token
echo "1️⃣ Hole OAuth Token..."
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "test_client",
    "client_secret": "test_secret_123456",
    "scope": "read write"
  }')

ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

if [ "$ACCESS_TOKEN" != "null" ] && [ -n "$ACCESS_TOKEN" ]; then
    echo "✅ Token erhalten: ${ACCESS_TOKEN:0:50}..."
else
    echo "❌ Token-Fehler!"
    echo $TOKEN_RESPONSE | jq .
    exit 1
fi

echo ""
echo "2️⃣ Liste verfügbare Modelle..."
curl -s http://localhost:8080/v1/models \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq -r '.data[].id'

echo ""
echo "3️⃣ Sende Chat Completion Request..."
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [
      {
        "role": "user",
        "content": "Hallo! Antworte in einem kurzen Satz."
      }
    ],
    "max_tokens": 100
  }')

echo ""
echo "📝 Antwort:"
echo $RESPONSE | jq -r '.choices[0].message.content'

echo ""
echo "📊 Token Usage:"
echo $RESPONSE | jq '.usage'

echo ""
echo "✅ API Test abgeschlossen!"
