# Admin API Documentation

## Overview

The Admin API provides management endpoints for the LLM-Proxy system. It uses API Key authentication (separate from OAuth) and provides full CRUD operations for OAuth clients, cache management, usage statistics, and provider monitoring.

**Base URL:** `http://localhost:8080/admin`

**Authentication:** API Key via `X-Admin-API-Key` header

---

## Authentication

All Admin API requests require an API Key in the header:

```bash
X-Admin-API-Key: your_admin_api_key_here
```

**Default Admin API Key** (configured in `configs/config.yaml`):
```
admin_dev_key_12345678901234567890123456789012
```

**Security:** Change this key in production!

---

## Endpoints

### 1. OAuth Client Management

#### List All Clients
```
GET /admin/clients
```

**Response:**
```json
{
  "clients": [
    {
      "id": "uuid",
      "client_id": "test_client",
      "name": "Test Client",
      "redirect_uris": [],
      "grant_types": ["client_credentials"],
      "default_scope": "read write",
      "rate_limit_rpm": null,
      "rate_limit_rpd": null,
      "enabled": true,
      "created_at": "2026-01-29T12:00:00Z",
      "updated_at": "2026-01-29T12:00:00Z"
    }
  ],
  "total": 1
}
```

#### Get Single Client
```
GET /admin/clients/{client_id}
```

**Response:**
```json
{
  "id": "uuid",
  "client_id": "test_client",
  "name": "Test Client",
  "redirect_uris": [],
  "grant_types": ["client_credentials"],
  "default_scope": "read write",
  "rate_limit_rpm": null,
  "rate_limit_rpd": null,
  "enabled": true,
  "created_at": "2026-01-29T12:00:00Z",
  "updated_at": "2026-01-29T12:00:00Z"
}
```

#### Create Client
```
POST /admin/clients
```

**Request Body:**
```json
{
  "client_id": "my_app",
  "client_secret": "secure_secret_123",
  "name": "My Application",
  "redirect_uris": [],
  "grant_types": ["client_credentials", "refresh_token"],
  "default_scope": "read write",
  "rate_limit_rpm": 1000,
  "rate_limit_rpd": 50000
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "client_id": "my_app",
  ...
}
```

#### Update Client
```
PATCH /admin/clients/{client_id}
```

**Request Body:** (all fields optional)
```json
{
  "name": "Updated Name",
  "default_scope": "read",
  "rate_limit_rpm": 500,
  "enabled": false
}
```

**Response:** `200 OK`

#### Delete Client
```
DELETE /admin/clients/{client_id}
```

**Response:**
```json
{
  "message": "client deleted successfully"
}
```

---

### 2. Cache Management

#### Get Cache Statistics
```
GET /admin/cache/stats
```

**Response:**
```json
{
  "hits": 150,
  "misses": 50,
  "errors": 0,
  "hit_rate": 75.0
}
```

#### Clear All Cache
```
POST /admin/cache/clear
```

**Response:**
```json
{
  "message": "cache cleared successfully"
}
```

#### Invalidate Cache by Model
```
POST /admin/cache/invalidate/{model}
```

**Example:**
```bash
POST /admin/cache/invalidate/claude-3-haiku-20240307
```

**Response:**
```json
{
  "message": "cache invalidated successfully",
  "entries_removed": 15
}
```

---

### 3. Usage Statistics

#### Get Usage Statistics
```
GET /admin/stats/usage
```

**Query Parameters:**
- `client_id` (optional): Filter by OAuth client
- `model` (optional): Filter by model name

**Response:**
```json
{
  "TotalRequests": 1250,
  "TotalTokens": 45000,
  "TotalCost": 0.85,
  "AvgDuration": 523.5,
  "CachedRequests": 300,
  "ErrorRequests": 5,
  "CacheHitRate": 24.0
}
```

**Example with filters:**
```bash
GET /admin/stats/usage?client_id=test_client&model=claude-3-haiku-20240307
```

---

### 4. Provider Management

#### Get Provider Status
```
GET /admin/providers/status
```

**Response:**
```json
{
  "healthy": true,
  "provider_count": 1,
  "models": [
    "claude-3-opus-20240229",
    "claude-3-sonnet-20240229",
    "claude-3-haiku-20240307"
  ]
}
```

**If unhealthy:**
```json
{
  "healthy": false,
  "provider_count": 1,
  "models": [...],
  "error": "error details"
}
```

---

## Error Responses

All endpoints return standard error responses:

```json
{
  "error": "error",
  "message": "descriptive error message"
}
```

**Common HTTP Status Codes:**
- `200` - Success
- `201` - Created
- `400` - Bad Request (invalid input)
- `401` - Unauthorized (invalid API key)
- `404` - Not Found
- `500` - Internal Server Error

---

## Examples

### Create a New OAuth Client

```bash
curl -X POST http://localhost:8080/admin/clients \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "mobile_app",
    "client_secret": "secret123",
    "name": "Mobile Application",
    "grant_types": ["client_credentials"],
    "default_scope": "read write",
    "rate_limit_rpm": 100
  }'
```

### Get Cache Statistics

```bash
curl http://localhost:8080/admin/cache/stats \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"
```

### Get Usage Statistics for a Specific Client

```bash
curl "http://localhost:8080/admin/stats/usage?client_id=mobile_app" \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"
```

### Clear All Cache

```bash
curl -X POST http://localhost:8080/admin/cache/clear \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"
```

### Update Client Rate Limits

```bash
curl -X PATCH http://localhost:8080/admin/clients/mobile_app \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  -H "Content-Type: application/json" \
  -d '{
    "rate_limit_rpm": 500,
    "rate_limit_rpd": 20000
  }'
```

---

## Security Best Practices

1. **Change Default Admin API Key** in production
2. **Use HTTPS** for all Admin API requests in production
3. **Restrict Admin API access** to internal networks only
4. **Rotate Admin API Keys** regularly
5. **Log all Admin API access** for audit trails
6. **Use different keys** for different environments (dev/staging/prod)

---

## Rate Limiting

Admin API endpoints are **not rate limited** by default. Ensure proper network-level protections are in place.

---

## Testing

Run the Admin API test suite:

```bash
curl -H "X-Admin-API-Key: key" http://localhost:8080/admin/filters
```

This will test all endpoints including:
- Authentication
- Client CRUD operations
- Cache management
- Usage statistics
- Provider status
- Security (invalid key rejection)

---

## Integration with Admin UI

The Admin API is designed to be consumed by the Admin UI (Svelte). All endpoints return JSON and follow RESTful conventions.

**Upcoming:** Admin UI with visual dashboards, graphs, and interactive client management.

---

## Troubleshooting

### "unauthorized" Error

**Problem:** Admin API key not accepted

**Solution:**
1. Check the `X-Admin-API-Key` header is set
2. Verify the key matches the one in `configs/config.yaml`
3. Ensure no extra whitespace in the key

### "client not found" Error

**Problem:** OAuth client doesn't exist

**Solution:**
1. Use `GET /admin/clients` to list all clients
2. Verify the `client_id` is correct
3. Create the client if it doesn't exist

### Cache Not Clearing

**Problem:** Cache entries still present after clear

**Solution:**
1. Check Redis is running
2. Verify cache is enabled in config
3. Check server logs for errors

---

## Configuration

Admin API key is configured in `configs/config.yaml`:

```yaml
admin:
  api_keys:
    - "admin_dev_key_12345678901234567890123456789012"
```

**Multiple keys supported** - add multiple keys to the array for different admins.

---

## Next Steps

- [ ] Request Logs Query Endpoints (paginated logs)
- [ ] Rate Limiting Enforcement
- [ ] Admin UI (Svelte)
- [ ] Audit Logging for Admin Actions
- [ ] Admin User Management
- [ ] Billing & Quota Management UI

---

**Admin API is production-ready!** 🎉
