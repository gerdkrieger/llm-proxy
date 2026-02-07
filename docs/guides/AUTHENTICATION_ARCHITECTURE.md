# Authentication Architecture Guide

**Last Updated:** 2026-02-07  
**Status:** Production  
**Security Level:** Critical

---

## Table of Contents

1. [Overview](#overview)
2. [Three-Tier Authentication System](#three-tier-authentication-system)
3. [Admin Authentication](#admin-authentication)
4. [Client API Key Authentication](#client-api-key-authentication)
5. [OAuth Token Authentication](#oauth-token-authentication)
6. [Route Protection Matrix](#route-protection-matrix)
7. [Security Best Practices](#security-best-practices)
8. [Common Misconfigurations](#common-misconfigurations)
9. [Key Management](#key-management)
10. [Testing & Validation](#testing--validation)

---

## Overview

LLM-Proxy implements a **three-tier authentication system** with strict separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     LLM-Proxy Server                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  /admin/*  Routes          /v1/*  Routes                   │
│  ┌──────────────────┐      ┌──────────────────────────┐   │
│  │ Admin Middleware │      │ API Key Middleware       │   │
│  │ (Admin Keys)     │      │ (Client Keys: sk-llm-*)  │   │
│  └──────────────────┘      │                          │   │
│         ↓                  │ ↓ Fallback if not sk-    │   │
│    Admin API               │                          │   │
│    - Clients               │ OAuth Middleware         │   │
│    - Filters               │ (JWT Tokens)             │   │
│    - System Stats          └──────────────────────────┘   │
│                                   ↓                        │
│                              LLM API                       │
│                              - Chat Completions            │
│                              - Models                      │
│                              - Embeddings                  │
└─────────────────────────────────────────────────────────────┘
```

**Key Principle:** Admin keys and client keys are completely separate and non-interchangeable.

---

## Three-Tier Authentication System

### Tier 1: Admin Authentication
- **Purpose:** Administrative operations only
- **Key Format:** `admin_*` (no strict prefix requirement)
- **Header:** `X-Admin-API-Key` or `Authorization: Bearer <admin_key>`
- **Routes:** `/admin/*` only
- **Cannot Access:** `/v1/*` routes (LLM API)
- **Configuration:** `config.yaml` → `admin.api_keys[]`

### Tier 2: Client API Key Authentication
- **Purpose:** Static keys for OpenAI-compatible clients
- **Key Format:** MUST start with `sk-llm-proxy-`
- **Header:** `Authorization: Bearer <client_key>`
- **Routes:** `/v1/*` only
- **Cannot Access:** `/admin/*` routes
- **Configuration:** `config.yaml` → `client_api_keys[]`

### Tier 3: OAuth Token Authentication
- **Purpose:** Dynamic JWT tokens for web applications
- **Token Format:** JWT (JSON Web Token)
- **Header:** `Authorization: Bearer <jwt_token>`
- **Routes:** `/v1/*` only
- **Cannot Access:** `/admin/*` routes
- **Configuration:** Database (`oauth_clients` table)

---

## Admin Authentication

### Purpose

Admin authentication is **exclusively** for:
- Accessing Admin UI (`https://llmproxy.aitrail.ch:3005`)
- Managing system configuration via Admin API
- Viewing statistics and logs
- Managing OAuth clients and filters

### Implementation

**File:** `internal/interfaces/middleware/admin.go`

```go
func (m *AdminMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Accept both X-Admin-API-Key header and Authorization Bearer
        token := r.Header.Get("X-Admin-API-Key")
        if token == "" {
            authHeader := r.Header.Get("Authorization")
            if strings.HasPrefix(authHeader, "Bearer ") {
                token = strings.TrimPrefix(authHeader, "Bearer ")
            }
        }
        
        // Validate against config.Admin.APIKeys
        for _, key := range m.config.Admin.APIKeys {
            if key.Key == token && key.Enabled {
                next.ServeHTTP(w, r)
                return
            }
        }
        
        // Reject if not valid
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
    })
}
```

### Configuration Example

**File:** `configs/config.yaml`

```yaml
admin:
  api_keys:
    - key: "admin_dev_local_key_12345678901234567890123456789012"
      name: "Local Development"
      enabled: true
    
    - key: "admin_prod_key_98765432109876543210987654321098"
      name: "Production Admin"
      enabled: true
```

### Protected Routes

All routes under `/admin/*`:
- `GET /admin/clients` - List OAuth clients
- `GET /admin/clients/:id` - Get client details
- `POST /admin/clients` - Create OAuth client
- `PUT /admin/clients/:id` - Update OAuth client
- `DELETE /admin/clients/:id` - Delete OAuth client
- `GET /admin/filters/stats` - Filter statistics
- `GET /admin/filters/blocked-requests` - Blocked requests
- `GET /admin/system/stats` - System statistics

### Usage Examples

**Admin UI Login:**
```bash
# The Admin UI automatically adds X-Admin-API-Key header
curl -H "X-Admin-API-Key: admin_dev_local_key_12345..." \
  https://llmproxy.aitrail.ch/admin/clients
```

**Direct Admin API Access:**
```bash
# Using X-Admin-API-Key header (recommended)
curl -H "X-Admin-API-Key: admin_dev_local_key_12345..." \
  https://llmproxy.aitrail.ch/admin/system/stats

# Using Authorization Bearer (alternative)
curl -H "Authorization: Bearer admin_dev_local_key_12345..." \
  https://llmproxy.aitrail.ch/admin/clients
```

### ⚠️ Critical Security Warning

**NEVER use admin keys for LLM API access!**

```bash
# This will FAIL (admin key cannot access /v1/*)
curl -H "Authorization: Bearer admin_dev_local_key_12345..." \
  https://llmproxy.aitrail.ch/v1/models

# Response: 401 Unauthorized
# Reason: Admin middleware is NOT applied to /v1/* routes
```

---

## Client API Key Authentication

### Purpose

Client API keys are for **LLM API access only**:
- OpenAI-compatible chat completions
- Model listing
- Embeddings
- Text-to-speech
- Any `/v1/*` endpoint

**Common Use Cases:**
- OpenWebUI integration
- Cursor IDE integration
- Custom applications using OpenAI SDK
- Third-party tools expecting OpenAI API

### Implementation

**File:** `internal/interfaces/middleware/apikey.go`

```go
func (m *APIKeyMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            next.ServeHTTP(w, r)  // Pass to next middleware (OAuth)
            return
        }
        
        token := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Only handle tokens starting with "sk-llm-proxy-"
        if !strings.HasPrefix(token, "sk-llm-proxy-") {
            next.ServeHTTP(w, r)  // Pass to OAuth middleware
            return
        }
        
        // Validate against config.ClientAPIKeys
        for _, key := range m.config.ClientAPIKeys {
            if key.Key == token && key.Enabled {
                // Store key info in context for logging
                ctx := context.WithValue(r.Context(), "api_key_name", key.Name)
                ctx = context.WithValue(ctx, "api_key_scopes", key.Scopes)
                next.ServeHTTP(w, r.WithContext(ctx))
                return
            }
        }
        
        // Reject if invalid
        http.Error(w, "Invalid API key", http.StatusUnauthorized)
    })
}
```

### Key Format Requirements

**MUST start with:** `sk-llm-proxy-`

**Recommended Format:**
```
sk-llm-proxy-{client-name}-{date}-{description}-{random}

Examples:
✅ sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
✅ sk-llm-proxy-cursor-2026-01-30-secure-key-def456uvw012
✅ sk-llm-proxy-mobile-app-2026-02-01-prod-key-ghi789jkl345
```

### Configuration Example

**File:** `configs/config.yaml`

```yaml
client_api_keys:
  - key: "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
    name: "OpenWebUI"
    scopes: ["read", "write"]
    enabled: true
  
  - key: "sk-llm-proxy-cursor-2026-01-30-secure-key-def456uvw012"
    name: "Cursor IDE"
    scopes: ["read", "write"]
    enabled: true
  
  - key: "sk-llm-proxy-readonly-2026-02-01-demo-key-mno012pqr345"
    name: "Demo App (Read-Only)"
    scopes: ["read"]
    enabled: true
  
  - key: "sk-llm-proxy-legacy-2025-12-01-old-key-stu345vwx678"
    name: "Legacy App (Disabled)"
    scopes: ["read", "write"]
    enabled: false
```

### Scopes

**Read Scope:** `["read"]`
- `GET /v1/models` - List available models
- `GET /v1/models/{model}` - Get model details

**Write Scope:** `["write"]`
- `POST /v1/chat/completions` - Chat completions
- `POST /v1/completions` - Text completions
- `POST /v1/embeddings` - Generate embeddings
- `POST /v1/audio/speech` - Text-to-speech

**Full Access:** `["read", "write"]`
- All endpoints (recommended for most clients)

### Usage Examples

**OpenWebUI Configuration:**
```bash
# In OpenWebUI Admin Panel → Settings → Connections
Base URL: https://llmproxy.aitrail.ch/v1
API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

**Cursor IDE Configuration:**
```json
{
  "openai": {
    "baseURL": "https://llmproxy.aitrail.ch/v1",
    "apiKey": "sk-llm-proxy-cursor-2026-01-30-secure-key-def456uvw012"
  }
}
```

**Direct API Access:**
```bash
# List models
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  https://llmproxy.aitrail.ch/v1/models

# Chat completion
curl -X POST https://llmproxy.aitrail.ch/v1/chat/completions \
  -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### ⚠️ Critical Security Warning

**Client keys CANNOT access admin endpoints:**

```bash
# This will FAIL (client key cannot access /admin/*)
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  https://llmproxy.aitrail.ch/admin/clients

# Response: 401 Unauthorized
# Reason: APIKeyMiddleware is NOT applied to /admin/* routes
```

---

## OAuth Token Authentication

### Purpose

OAuth tokens provide **dynamic, time-limited access** for:
- Web applications with user login
- Mobile applications
- Third-party integrations requiring token refresh
- Applications needing fine-grained user permissions

### Implementation

**File:** `internal/interfaces/middleware/oauth.go`

```go
func (m *OAuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }
        
        token := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Skip if it's a static API key (starts with sk-llm-proxy-)
        if strings.HasPrefix(token, "sk-llm-proxy-") {
            next.ServeHTTP(w, r)  // Already handled by APIKeyMiddleware
            return
        }
        
        // Validate JWT token
        claims, err := m.validateJWT(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Store claims in context
        ctx := context.WithValue(r.Context(), "oauth_client_id", claims.ClientID)
        ctx = context.WithValue(ctx, "oauth_scopes", claims.Scopes)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### OAuth Flow

1. **Client Registration** (via Admin API):
   ```bash
   curl -X POST https://llmproxy.aitrail.ch/admin/clients \
     -H "X-Admin-API-Key: admin_dev_local_key_12345..." \
     -H "Content-Type: application/json" \
     -d '{
       "name": "My Web App",
       "redirect_uris": ["https://myapp.com/callback"],
       "scopes": ["read", "write"]
     }'
   
   # Response:
   {
     "client_id": "client_abc123xyz789",
     "client_secret": "secret_def456uvw012",
     "redirect_uris": ["https://myapp.com/callback"],
     "scopes": ["read", "write"]
   }
   ```

2. **Authorization Request** (user redirected to):
   ```
   https://llmproxy.aitrail.ch/oauth/authorize?
     client_id=client_abc123xyz789&
     redirect_uri=https://myapp.com/callback&
     response_type=code&
     scope=read+write
   ```

3. **Token Exchange** (backend):
   ```bash
   curl -X POST https://llmproxy.aitrail.ch/oauth/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=authorization_code" \
     -d "code=auth_code_from_callback" \
     -d "client_id=client_abc123xyz789" \
     -d "client_secret=secret_def456uvw012" \
     -d "redirect_uri=https://myapp.com/callback"
   
   # Response:
   {
     "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
     "token_type": "Bearer",
     "expires_in": 3600,
     "refresh_token": "refresh_ghi789jkl345mno678",
     "scope": "read write"
   }
   ```

4. **API Access** (using access token):
   ```bash
   curl -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
     https://llmproxy.aitrail.ch/v1/chat/completions \
     -H "Content-Type: application/json" \
     -d '{"model": "gpt-4", "messages": [...]}'
   ```

5. **Token Refresh** (when expired):
   ```bash
   curl -X POST https://llmproxy.aitrail.ch/oauth/token \
     -H "Content-Type: application/x-www-form-urlencoded" \
     -d "grant_type=refresh_token" \
     -d "refresh_token=refresh_ghi789jkl345mno678" \
     -d "client_id=client_abc123xyz789" \
     -d "client_secret=secret_def456uvw012"
   ```

### Token Structure (JWT)

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "user_12345",
    "client_id": "client_abc123xyz789",
    "scopes": ["read", "write"],
    "exp": 1738934400,
    "iat": 1738930800
  },
  "signature": "..."
}
```

### ⚠️ Important Notes

- OAuth tokens **cannot access** `/admin/*` routes
- Tokens expire (default: 1 hour)
- Refresh tokens valid for longer (default: 30 days)
- Clients stored in PostgreSQL database

---

## Route Protection Matrix

| Route Pattern | Admin Key | Client Key (`sk-llm-proxy-*`) | OAuth Token |
|--------------|-----------|-------------------------------|-------------|
| `/admin/clients` | ✅ YES | ❌ NO | ❌ NO |
| `/admin/filters/*` | ✅ YES | ❌ NO | ❌ NO |
| `/admin/system/*` | ✅ YES | ❌ NO | ❌ NO |
| `/v1/models` | ❌ NO | ✅ YES (read scope) | ✅ YES (read scope) |
| `/v1/chat/completions` | ❌ NO | ✅ YES (write scope) | ✅ YES (write scope) |
| `/v1/embeddings` | ❌ NO | ✅ YES (write scope) | ✅ YES (write scope) |
| `/oauth/authorize` | ❌ NO | ❌ NO | 🔓 Public |
| `/oauth/token` | ❌ NO | ❌ NO | 🔓 Public |

**Legend:**
- ✅ **YES** - Authenticated and authorized
- ❌ **NO** - Returns 401 Unauthorized
- 🔓 **Public** - No authentication required

---

## Security Best Practices

### 1. Key Separation

**DO:**
- ✅ Use admin keys ONLY for Admin UI and admin operations
- ✅ Use client keys for OpenWebUI, Cursor, and other LLM clients
- ✅ Use OAuth for web applications with user login
- ✅ Generate unique keys for each application
- ✅ Use descriptive key names in configuration

**DON'T:**
- ❌ Use admin keys in OpenWebUI or other LLM clients
- ❌ Use client keys for administrative operations
- ❌ Share keys between multiple applications
- ❌ Hardcode keys in client-side code
- ❌ Commit keys to version control

### 2. Key Naming Conventions

**Admin Keys:**
```
admin_{environment}_{description}_key_{random}

Examples:
admin_prod_key_98765432109876543210987654321098
admin_dev_local_key_12345678901234567890123456789012
admin_staging_key_11111111111111111111111111111111
```

**Client Keys:**
```
sk-llm-proxy-{client}-{date}-{description}-{random}

Examples:
sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
sk-llm-proxy-cursor-2026-01-30-secure-key-def456uvw012
sk-llm-proxy-mobile-2026-02-01-prod-key-ghi789jkl345
```

### 3. Scope Restrictions

**Principle of Least Privilege:**
- Only grant `["write"]` scope if needed for chat/completions
- Use `["read"]` for applications that only list models
- Regularly audit scope assignments

**Example:**
```yaml
client_api_keys:
  # Full access for interactive chat applications
  - key: "sk-llm-proxy-openwebui-..."
    name: "OpenWebUI"
    scopes: ["read", "write"]
  
  # Read-only for monitoring tools
  - key: "sk-llm-proxy-monitor-..."
    name: "Monitoring Dashboard"
    scopes: ["read"]
```

### 4. Key Rotation

**Recommended Schedule:**
- Admin keys: Every 90 days
- Client keys: Every 180 days
- OAuth client secrets: Every 365 days

**Rotation Process:**
1. Generate new key in `config.yaml`
2. Update client configuration with new key
3. Test new key works correctly
4. Set old key to `enabled: false`
5. Monitor logs for usage of old key
6. Remove old key after 30 days

### 5. Key Storage

**Server-Side (Production):**
```bash
# Store config.yaml with restrictive permissions
sudo chown llm-proxy:llm-proxy /opt/llm-proxy/configs/config.yaml
sudo chmod 600 /opt/llm-proxy/configs/config.yaml
```

**Environment Variables (Alternative):**
```bash
# For sensitive keys, use environment variables
export LLM_PROXY_ADMIN_KEY="admin_prod_key_..."
export LLM_PROXY_CLIENT_KEY_OPENWEBUI="sk-llm-proxy-openwebui-..."
```

**Client-Side:**
- Use environment variables (`.env` files with `.gitignore`)
- Use secret management tools (Vault, AWS Secrets Manager)
- Never commit keys to Git repositories

### 6. Logging & Monitoring

**What to Log:**
- ✅ Authentication failures (wrong key, expired token)
- ✅ Key usage patterns (detect anomalies)
- ✅ Scope violations (write attempt with read-only key)
- ✅ Admin operations (client creation, config changes)

**What NOT to Log:**
- ❌ Full API keys (log only last 4 characters)
- ❌ Client secrets
- ❌ Password hashes

**Example Log Entry:**
```json
{
  "level": "info",
  "timestamp": "2026-02-07T10:15:30Z",
  "message": "API request authenticated",
  "key_name": "OpenWebUI",
  "key_last4": "z789",
  "endpoint": "/v1/chat/completions",
  "remote_addr": "172.18.0.2",
  "scopes": ["read", "write"]
}
```

---

## Common Misconfigurations

### Problem 1: OpenWebUI Using Admin Key

**Symptoms:**
- OpenWebUI shows "Connection failed"
- Logs show `401 Unauthorized` for `/v1/models`
- Live Monitor shows ⚠️ Auth Failed

**Cause:**
```yaml
# WRONG: Using admin key in OpenWebUI
OpenWebUI API Key: admin_dev_local_key_12345...
```

**Fix:**
```yaml
# CORRECT: Use client key
OpenWebUI API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

**Why It Happens:**
- User copied admin key from Admin UI documentation
- Misunderstood key separation
- Admin key was easier to find in config

**Prevention:**
- Clear documentation emphasizing key separation
- Admin UI warning when admin key used for LLM API
- Live Monitor showing key type being used

---

### Problem 2: Client Key Without `sk-llm-proxy-` Prefix

**Symptoms:**
- API returns `401 Unauthorized`
- Logs show "Token validation failed"
- Key exists in config but not recognized

**Cause:**
```yaml
# WRONG: Missing required prefix
client_api_keys:
  - key: "openwebui-2026-01-30-secure-key-abc123xyz789"
```

**Fix:**
```yaml
# CORRECT: Must start with sk-llm-proxy-
client_api_keys:
  - key: "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
```

**Why It Happens:**
- User didn't read format requirements
- Copied key format from different proxy
- Thought prefix was optional

**Prevention:**
- Validation on config load (warn about invalid format)
- Key generation script that enforces prefix
- Documentation with clear examples

---

### Problem 3: Disabled Key Still in Use

**Symptoms:**
- Intermittent `401 Unauthorized` errors
- Works locally but fails in production
- Old application still connecting

**Cause:**
```yaml
# Key disabled but client not updated
client_api_keys:
  - key: "sk-llm-proxy-legacy-2025-12-01-old-key-stu345vwx678"
    name: "Legacy App"
    enabled: false  # ← Client still using this!
```

**Fix:**
1. Identify clients using disabled key (check logs)
2. Update clients with new key
3. Remove disabled key from config

**Prevention:**
- 30-day grace period before disabling keys
- Email notifications to key owners
- Automated key rotation scripts

---

### Problem 4: OAuth Token Used for Admin Operations

**Symptoms:**
- Web app can't access admin endpoints
- Returns `401 Unauthorized` for `/admin/*`

**Cause:**
```javascript
// WRONG: OAuth token cannot access admin routes
fetch('https://llmproxy.aitrail.ch/admin/clients', {
  headers: {
    'Authorization': `Bearer ${oauthAccessToken}`
  }
})
```

**Fix:**
```javascript
// CORRECT: Use admin key for admin operations
fetch('https://llmproxy.aitrail.ch/admin/clients', {
  headers: {
    'X-Admin-API-Key': adminKey
  }
})
```

**Why It Happens:**
- Developer assumes same token works for all routes
- Misunderstood route separation
- Frontend needs both admin and LLM API access

**Prevention:**
- Clear API documentation with route examples
- Separate admin SDK from LLM SDK
- OpenAPI spec showing auth per endpoint

---

## Key Management

### Generating Keys

**Admin Keys:**
```bash
# Generate secure random key (64 characters)
openssl rand -hex 32 | awk '{print "admin_prod_key_" $0}'

# Output:
# admin_prod_key_a1b2c3d4e5f6...
```

**Client Keys:**
```bash
# Generate with required prefix
openssl rand -hex 20 | awk '{print "sk-llm-proxy-openwebui-2026-02-07-secure-key-" $0}'

# Output:
# sk-llm-proxy-openwebui-2026-02-07-secure-key-a1b2c3d4e5f6...
```

### Adding Keys to Configuration

**File:** `configs/config.yaml`

```yaml
admin:
  api_keys:
    - key: "admin_prod_key_a1b2c3d4e5f6..."
      name: "Production Admin"
      enabled: true

client_api_keys:
  - key: "sk-llm-proxy-openwebui-2026-02-07-secure-key-a1b2c3d4e5f6..."
    name: "OpenWebUI"
    scopes: ["read", "write"]
    enabled: true
```

**Restart Service:**
```bash
cd /opt/llm-proxy
docker compose -f docker-compose.openwebui.yml restart backend
```

### Revoking Keys

**Immediate Revocation:**
```yaml
# Set enabled: false
client_api_keys:
  - key: "sk-llm-proxy-compromised-key-..."
    name: "Compromised Key"
    enabled: false  # ← Immediately blocks access
```

**Permanent Removal:**
```yaml
# Remove from config entirely
# (Do this after 30-day grace period)
```

### Key Auditing

**Check Key Usage:**
```bash
# See which keys are being used
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep 'API request' | tail -100"
```

**Find Last Usage:**
```bash
# Check last time a specific key was used
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep 'key_name.*OpenWebUI' | tail -1"
```

**Identify Unused Keys:**
```bash
# Keys not used in last 30 days (candidates for removal)
# Compare config against logs
```

---

## Testing & Validation

### Automated Security Tests

**Script:** `scripts/maintenance/test-auth-separation.sh`

```bash
#!/bin/bash
set -euo pipefail

# Test authentication separation
# Ensures admin keys can't access LLM API and vice versa

BASE_URL="https://llmproxy.aitrail.ch"
ADMIN_KEY="admin_dev_local_key_12345678901234567890123456789012"
CLIENT_KEY="sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"

echo "🔒 Testing Authentication Separation"
echo "===================================="
echo ""

# Test 1: Admin key should access admin routes
echo "Test 1: Admin key → /admin/clients"
status=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "X-Admin-API-Key: $ADMIN_KEY" \
  "$BASE_URL/admin/clients")

if [ "$status" = "200" ]; then
  echo "✅ PASS: Admin key can access /admin/* (expected)"
else
  echo "❌ FAIL: Admin key returned $status for /admin/*"
  exit 1
fi

# Test 2: Admin key should NOT access LLM API
echo "Test 2: Admin key → /v1/models"
status=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $ADMIN_KEY" \
  "$BASE_URL/v1/models")

if [ "$status" = "401" ]; then
  echo "✅ PASS: Admin key rejected from /v1/* (expected)"
else
  echo "❌ FAIL: Admin key returned $status for /v1/* (should be 401)"
  exit 1
fi

# Test 3: Client key should access LLM API
echo "Test 3: Client key → /v1/models"
status=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $CLIENT_KEY" \
  "$BASE_URL/v1/models")

if [ "$status" = "200" ]; then
  echo "✅ PASS: Client key can access /v1/* (expected)"
else
  echo "❌ FAIL: Client key returned $status for /v1/*"
  exit 1
fi

# Test 4: Client key should NOT access admin routes
echo "Test 4: Client key → /admin/clients"
status=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer $CLIENT_KEY" \
  "$BASE_URL/admin/clients")

if [ "$status" = "401" ]; then
  echo "✅ PASS: Client key rejected from /admin/* (expected)"
else
  echo "❌ FAIL: Client key returned $status for /admin/* (should be 401)"
  exit 1
fi

# Test 5: Invalid key should be rejected
echo "Test 5: Invalid key → /v1/models"
status=$(curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer invalid-key-12345" \
  "$BASE_URL/v1/models")

if [ "$status" = "401" ]; then
  echo "✅ PASS: Invalid key rejected (expected)"
else
  echo "❌ FAIL: Invalid key returned $status (should be 401)"
  exit 1
fi

echo ""
echo "🎉 All authentication separation tests passed!"
```

**Run Tests:**
```bash
chmod +x scripts/maintenance/test-auth-separation.sh
./scripts/maintenance/test-auth-separation.sh
```

---

### Manual Testing Checklist

**Admin Authentication:**
- [ ] Admin key can access Admin UI
- [ ] Admin key can list OAuth clients
- [ ] Admin key can view filter statistics
- [ ] Admin key CANNOT access `/v1/models`
- [ ] Admin key CANNOT make chat completions
- [ ] Invalid admin key returns 401

**Client Key Authentication:**
- [ ] Client key can list models
- [ ] Client key can make chat completions
- [ ] Client key with `read` scope can list models
- [ ] Client key with `read` scope CANNOT make completions
- [ ] Client key CANNOT access `/admin/clients`
- [ ] Invalid client key returns 401
- [ ] Key without `sk-llm-proxy-` prefix returns 401

**OAuth Authentication:**
- [ ] Valid JWT token can access `/v1/*`
- [ ] Expired JWT token returns 401
- [ ] JWT token CANNOT access `/admin/*`
- [ ] Refresh token can obtain new access token

---

### Debugging Authentication Issues

**Enable Debug Logging:**
```yaml
# config.yaml
logging:
  level: debug  # Show detailed auth logs
```

**Check Logs:**
```bash
# Follow logs in real-time
ssh openweb "docker logs -f llm-proxy-backend 2>&1 | grep -E '(auth|token|401)'"

# Search for specific key (last 4 characters)
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep 'z789'"
```

**Test with curl:**
```bash
# Verbose output shows exact HTTP exchange
curl -v -H "Authorization: Bearer sk-llm-proxy-..." \
  https://llmproxy.aitrail.ch/v1/models
```

**Check Configuration:**
```bash
# Verify key exists in config
ssh openweb "grep -A 3 'sk-llm-proxy-openwebui' /opt/llm-proxy/configs/config.yaml"

# Check key is enabled
ssh openweb "grep -A 5 'sk-llm-proxy-openwebui' /opt/llm-proxy/configs/config.yaml | grep enabled"
```

---

## Summary

### Key Principles

1. **Three separate authentication systems:**
   - Admin keys for `/admin/*`
   - Client keys for `/v1/*`
   - OAuth tokens for `/v1/*`

2. **Complete isolation:**
   - Admin keys CANNOT access LLM API
   - Client keys CANNOT access admin endpoints
   - No mixing or fallback between key types

3. **Strict format requirements:**
   - Client keys MUST start with `sk-llm-proxy-`
   - Admin keys have no prefix requirement
   - OAuth tokens are JWT format

4. **Security best practices:**
   - One key per application
   - Regular key rotation
   - Least privilege scopes
   - Secure storage
   - Comprehensive logging

### Quick Reference

**Use Admin Key For:**
- Admin UI login
- Creating OAuth clients
- Viewing system statistics
- Managing filters

**Use Client Key For:**
- OpenWebUI
- Cursor IDE
- Custom LLM applications
- OpenAI SDK integrations

**Use OAuth Token For:**
- Web applications with users
- Mobile applications
- Third-party integrations

---

## Related Documentation

- [OpenWebUI Connection Fix](../OPENWEBUI_CONNECTION_FIX.md)
- [Complete Monitoring Guide](./COMPLETE_MONITORING_GUIDE.md)
- [OpenWebUI Setup Guide](./OPENWEBUI_SETUP.md)

---

**Document Version:** 1.0  
**Last Updated:** 2026-02-07  
**Author:** LLM-Proxy Development Team
