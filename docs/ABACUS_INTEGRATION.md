# Abacus.ai Integration Guide

## Overview

This document describes the integration of **Abacus.ai** as a new LLM provider in the LLM-Proxy project. The integration allows OpenWebUI and other OpenAI-compatible clients to use Abacus.ai deployments through the proxy.

## Architecture

The Abacus.ai integration follows the existing provider architecture pattern:

```
internal/infrastructure/providers/abacus/
├── client.go        # Abacus.ai HTTP client
└── mapper.go        # OpenAI ↔ Abacus.ai format conversion
```

### Key Components

1. **Client** (`client.go`): Handles HTTP communication with Abacus.ai API
2. **Mapper** (`mapper.go`): Converts between OpenAI and Abacus.ai request/response formats
3. **Provider Manager** (`manager.go`): Registers and manages Abacus.ai clients
4. **Model Sync Service** (`model_sync_service.go`): Defines available Abacus.ai models

## Abacus.ai API Specifics

### Authentication
- Uses `apiKey` header (not `Authorization: Bearer`)
- API key format: Custom Abacus.ai key

### Base URL
```
https://api.abacus.ai/api/v0
```

### Endpoint
```
POST /api/v0/chatLLM
```

### Request Format
```json
{
  "deploymentId": "your-deployment-id",
  "messages": [
    {"role": "user", "content": "Hello"},
    {"role": "assistant", "content": "Hi there!"}
  ],
  "systemMessage": "You are a helpful assistant",
  "temperature": 0.7,
  "maxTokens": 1000,
  "topP": 0.9,
  "stream": false,
  "llmName": "gpt-4"
}
```

### Response Format
```json
{
  "success": true,
  "result": {
    "messages": [
      {"role": "user", "content": "Hello"},
      {"role": "assistant", "content": "Hi there! How can I help?"}
    ],
    "conversationId": "conv-123",
    "deploymentId": "deployment-456",
    "llmDisplayName": "GPT-4"
  }
}
```

## Configuration

### 1. Update `config.yaml`

Add Abacus.ai configuration to your `configs/config.yaml`:

```yaml
providers:
  abacus:
    enabled: true
    api_keys:
      - key: "YOUR-ABACUS-API-KEY"
        weight: 1
        max_rpm: 1000
    
    # Default deployment ID (required if not in model name)
    deployment_id: "YOUR-DEPLOYMENT-ID"
    
    # Available models
    models:
      - abacus:gpt-4
      - abacus:gpt-4-turbo
      - abacus:claude-3-opus
      - abacus:claude-3-sonnet
      - abacus:llama-3-70b
    
    timeout: 120s
```

### 2. Environment Variables (Alternative)

You can also configure via `.env`:

```bash
ABACUS_ENABLED=true
ABACUS_API_KEY=your-api-key-here
ABACUS_DEPLOYMENT_ID=your-deployment-id
```

## Model Naming Convention

Abacus.ai uses deployment IDs instead of model names. The proxy supports two formats:

### Format 1: Explicit Deployment ID
```
abacus:deployment-id-here
```

Example:
```
abacus:abc123def456
```

### Format 2: LLM Name (uses default deployment_id from config)
```
abacus:gpt-4
abacus:claude-3-opus
```

The mapper will:
1. Extract deployment ID from model name if present
2. Fall back to `deployment_id` from config
3. Use `llmName` field to specify the underlying LLM

## Usage Examples

### OpenWebUI Configuration

1. **Add Model in OpenWebUI**:
   - Go to Settings → Models
   - Add new model: `abacus:gpt-4`
   - Set base URL: `http://localhost:8080/v1`
   - Set API key: Your client API key from config

2. **Start Chatting**:
   - Select `abacus:gpt-4` from model dropdown
   - Send messages as normal
   - Proxy translates to Abacus.ai format automatically

### cURL Example

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR-CLIENT-API-KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "abacus:gpt-4",
    "messages": [
      {"role": "system", "content": "You are a helpful assistant"},
      {"role": "user", "content": "What is the capital of France?"}
    ],
    "temperature": 0.7,
    "max_tokens": 1000
  }'
```

## Request Flow

```
OpenWebUI/Client
    ↓ (OpenAI format)
LLM-Proxy API Handler
    ↓
Provider Manager (determines provider: "abacus")
    ↓
Abacus Mapper (MapOpenAIToAbacus)
    ↓ (Abacus.ai format)
Abacus Client (HTTP POST to api.abacus.ai)
    ↓
Abacus.ai API
    ↓ (Abacus.ai response)
Abacus Mapper (MapAbacusToOpenAI)
    ↓ (OpenAI format)
OpenWebUI/Client
```

## Mapping Details

### System Messages
- OpenAI: System messages in `messages` array with `role: "system"`
- Abacus.ai: Separate `systemMessage` field
- **Mapping**: Mapper extracts system messages and puts them in `systemMessage` field

### Message Roles
- Supported: `user`, `assistant`, `system`
- Unsupported roles are skipped

### Parameters
| OpenAI          | Abacus.ai      | Notes                    |
|-----------------|----------------|--------------------------|
| `model`         | `deploymentId` | Extracted from model name|
| `messages`      | `messages`     | Filtered (no system)     |
| `temperature`   | `temperature`  | Direct mapping           |
| `max_tokens`    | `maxTokens`    | Direct mapping           |
| `top_p`         | `topP`         | Direct mapping           |
| `stream`        | `stream`       | Direct mapping           |
| -               | `llmName`      | Set from model name      |
| -               | `systemMessage`| Extracted from messages  |

### Response Mapping
- Abacus.ai returns full conversation history
- Mapper extracts last assistant message
- Converts to OpenAI `choices` format
- Token counts set to 0 (Abacus.ai doesn't provide them)

## Available Models

The following models are registered in the database:

- `abacus:gpt-4` - GPT-4 via Abacus.ai
- `abacus:gpt-4-turbo` - GPT-4 Turbo via Abacus.ai
- `abacus:gpt-3.5-turbo` - GPT-3.5 Turbo via Abacus.ai
- `abacus:claude-3-opus` - Claude 3 Opus via Abacus.ai
- `abacus:claude-3-sonnet` - Claude 3 Sonnet via Abacus.ai
- `abacus:claude-3-haiku` - Claude 3 Haiku via Abacus.ai
- `abacus:llama-3-70b` - Llama 3 70B via Abacus.ai
- `abacus:llama-3-8b` - Llama 3 8B via Abacus.ai
- `abacus:mistral-large` - Mistral Large via Abacus.ai
- `abacus:mistral-medium` - Mistral Medium via Abacus.ai

## Health Checks

The Abacus.ai client implements a basic health check:

```go
func (c *Client) Health(ctx context.Context) error
```

Currently checks:
- API key is configured

Future improvements:
- Make lightweight API call to verify connectivity
- Check deployment availability

## Error Handling

### Abacus.ai Error Response
```json
{
  "success": false,
  "error": "Error message",
  "errorType": "ValidationError",
  "errorCode": 400
}
```

### Mapped to OpenAI Error
```json
{
  "error": {
    "message": "abacus.ai API error (status=400, type=ValidationError, code=400): Error message",
    "type": "api_error",
    "code": "abacus_error"
  }
}
```

## Load Balancing

Abacus.ai supports multiple API keys with weight-based load balancing:

```yaml
providers:
  abacus:
    api_keys:
      - key: "key-1"
        weight: 2
        max_rpm: 1000
      - key: "key-2"
        weight: 1
        max_rpm: 500
```

The provider manager will:
- Distribute requests based on weights
- Respect RPM limits per key
- Failover to next key on errors

## Database Storage

API keys can be stored in the database instead of config:

1. **Add via Admin API**:
```bash
curl -X POST http://localhost:3005/api/providers/keys \
  -H "Authorization: Bearer ADMIN-API-KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "provider_id": "abacus",
    "api_key": "YOUR-ABACUS-KEY",
    "weight": 1,
    "max_rpm": 1000
  }'
```

2. **Hot Reload**:
```bash
curl -X POST http://localhost:8080/admin/reload-keys \
  -H "Authorization: Bearer ADMIN-API-KEY"
```

Keys are encrypted in database using AES-256-GCM.

## Testing

### Unit Tests
```bash
go test ./internal/infrastructure/providers/abacus/...
```

### Integration Test
```bash
# Set environment variables
export ABACUS_API_KEY="your-key"
export ABACUS_DEPLOYMENT_ID="your-deployment-id"

# Run integration test
go test -tags=integration ./internal/infrastructure/providers/abacus/...
```

### Manual Test with cURL
```bash
# Test chat completion
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR-CLIENT-KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "abacus:gpt-4",
    "messages": [
      {"role": "user", "content": "Say hello!"}
    ]
  }'
```

## Troubleshooting

### Issue: "No deployment ID found"
**Solution**: Set `deployment_id` in config or use format `abacus:deployment-id`

### Issue: "API key not configured"
**Solution**: Add API key to `config.yaml` or database

### Issue: "Provider not found"
**Solution**: Ensure `enabled: true` in config and restart backend

### Issue: "Invalid API key"
**Solution**: Verify API key is correct in Abacus.ai dashboard

## Future Enhancements

1. **Streaming Support**: Implement SSE streaming for real-time responses
2. **Token Counting**: Parse Abacus.ai response for actual token usage
3. **Advanced Health Check**: Make API call to verify deployment availability
4. **Conversation Management**: Support `conversationId` for multi-turn conversations
5. **Document Search**: Support `searchDocuments` and filters
6. **Model Discovery**: Fetch available deployments from Abacus.ai API

## Files Modified/Created

### Created
- `internal/infrastructure/providers/abacus/client.go` (195 lines)
- `internal/infrastructure/providers/abacus/mapper.go` (155 lines)
- `docs/ABACUS_INTEGRATION.md` (this file)

### Modified
- `internal/config/config.go` - Added `AbacusConfig` and `AbacusAPIKey`
- `internal/infrastructure/providers/manager.go` - Registered Abacus.ai provider
- `internal/application/providers/model_sync_service.go` - Added Abacus.ai models
- `configs/config.example.yaml` - Added Abacus.ai configuration example

## References

- [Abacus.ai API Documentation](https://api.abacus.ai/)
- [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
- [LLM-Proxy Architecture](../README.md)

## Support

For issues or questions:
1. Check logs: `docker-compose logs -f backend`
2. Verify configuration: `configs/config.yaml`
3. Test health: `curl http://localhost:8080/health`
4. Review this guide

---

**Last Updated**: March 8, 2026  
**Integration Version**: 1.0  
**Status**: ✅ Complete (pending real API testing)
