# Testing Guide for LLM-Proxy

## Quick Start

### 1. Start the Server
```bash
make dev
```

### 2. Run the Test Suite
```bash
./test_api.sh
```

## Manual Testing

### OAuth Token Generation

Get an access token:
```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "test_client",
    "client_secret": "test_secret_123456",
    "scope": "read write"
  }' | jq .
```

Response:
```json
{
  "access_token": "eyJhbGci...",
  "token_type": "Bearer",
  "expires_in": 3599,
  "refresh_token": "eyJhbGci...",
  "scope": "read write"
}
```

### List Available Models

```bash
TOKEN="<your_access_token>"

curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer $TOKEN" | jq .
```

Response:
```json
{
  "object": "list",
  "data": [
    {
      "id": "claude-3-opus-20240229",
      "object": "model",
      "created": 1769692238,
      "owned_by": "anthropic"
    },
    ...
  ]
}
```

### Chat Completion

```bash
TOKEN="<your_access_token>"

curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [
      {
        "role": "user",
        "content": "Hello! Please respond with a brief greeting."
      }
    ],
    "max_tokens": 100,
    "temperature": 0.7
  }' | jq .
```

Response:
```json
{
  "id": "msg_abc123",
  "object": "chat.completion",
  "created": 1769692238,
  "model": "claude-3-haiku-20240307",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "Hello! It's nice to meet you."
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 12,
    "completion_tokens": 8,
    "total_tokens": 20
  }
}
```

## Test OAuth Client

A test OAuth client is pre-configured in the database:

- **Client ID**: `test_client`
- **Client Secret**: `test_secret_123456`
- **Scopes**: `read write`
- **Grant Types**: `client_credentials`, `refresh_token`

## Configuration

Before testing chat completions, ensure you have a valid Claude API key:

1. Open `.env` file
2. Update `CLAUDE_API_KEY` with your real Anthropic API key
3. Restart the server: `make restart`

## Error Handling Tests

### Invalid Credentials
```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "invalid",
    "client_secret": "invalid"
  }' | jq .
```

Expected: `401 Unauthorized` with error `invalid_client`

### Invalid Token
```bash
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer invalid_token" | jq .
```

Expected: `401 Unauthorized` with error `invalid_token`

### Missing Scope
```bash
# Get token with only 'read' scope
TOKEN=$(curl -s -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "test_client",
    "client_secret": "test_secret_123456",
    "scope": "read"
  }' | jq -r '.access_token')

# Try to use chat endpoint (requires 'write' scope)
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [{"role": "user", "content": "Hello"}]
  }' | jq .
```

Expected: `403 Forbidden` with error about insufficient scope

## Database Verification

Check OAuth tokens in database:
```bash
make db-psql
\x
SELECT * FROM oauth_tokens ORDER BY created_at DESC LIMIT 5;
```

Check request logs:
```bash
make db-psql
\x
SELECT * FROM request_logs ORDER BY created_at DESC LIMIT 5;
```

## Performance Testing

Use Apache Bench for load testing:
```bash
# Get a token first
TOKEN="<your_access_token>"

# Test models endpoint
ab -n 1000 -c 10 \
  -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/v1/models
```

## Troubleshooting

### Server won't start
- Check if ports 8080, 5433, 6380 are available
- Verify Docker services are running: `cd deployments/docker && docker compose ps`
- Check logs: `make logs`

### OAuth token invalid
- Verify `OAUTH_JWT_SECRET` is set in `.env`
- Check token hasn't expired (default: 1 hour)
- Ensure client exists in database

### Chat completion fails
- Verify `CLAUDE_API_KEY` is valid
- Check provider health: `curl http://localhost:8080/health/detailed | jq .`
- Review server logs for API errors

## Next Steps

1. **Add your Claude API key** to `.env`
2. **Create additional OAuth clients** for different teams/projects
3. **Monitor request logs** to track usage and costs
4. **Set up rate limits** for OAuth clients in the database
5. **Configure caching** for frequently used prompts
