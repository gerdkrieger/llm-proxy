# OpenWebUI Header Configuration Fix

**Date:** 2026-02-07  
**Issue:** OpenWebUI using wrong header name for API key  
**Status:** CRITICAL - This is why OpenWebUI can't connect!

---

## Problem Identified

**Current OpenWebUI Configuration:**
```json
{
  "X-Admin-API-Key": "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
}
```

**Why This Fails:**

1. **OpenWebUI sends:** `X-Admin-API-Key: sk-llm-proxy-...`
2. **APIKeyMiddleware expects:** `Authorization: Bearer sk-llm-proxy-...`
3. **Result:** Key is ignored → 401 Unauthorized

**Technical Explanation:**

```go
// APIKeyMiddleware (internal/interfaces/middleware/apikey.go)
// Only reads Authorization header!
authHeader := r.Header.Get("Authorization")  // ← Looking here
if !strings.HasPrefix(authHeader, "Bearer ") {
    next.ServeHTTP(w, r)  // Not found → pass to OAuth
    return
}
```

The `X-Admin-API-Key` header is **only read by AdminMiddleware**, which is **only active on `/admin/*` routes**!

For `/v1/*` routes (LLM API), you **must use** `Authorization: Bearer`.

---

## The Fix

### Option 1: Change Header in OpenWebUI (Recommended)

**OpenWebUI Interface:**

1. Open `https://chat.aitrail.ch`
2. Login as admin
3. Go to: **Admin Panel → Settings → Connections**
4. Find **OpenAI API** connection section
5. Look for **Custom Headers** or **API Configuration**

**Change from:**
```json
{
  "X-Admin-API-Key": "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
}
```

**To:**
```json
{
  "Authorization": "Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
}
```

**Or simpler - use the standard API Key field:**

Most likely, OpenWebUI has a dedicated **API Key** field. Use that instead of custom headers:

```
API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

OpenWebUI will automatically format it as: `Authorization: Bearer sk-llm-proxy-...`

---

### Option 2: Remove Custom Header, Use Standard Field

**Best Practice:**

Don't use custom headers at all! OpenWebUI should have standard fields:

**Configuration:**
```
Base URL: https://scrubgate.tech/v1
API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

**Custom Headers:** (Leave empty or remove)

OpenWebUI will automatically send:
```
Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

---

## Testing the Fix

### Test 1: Manual curl Test

**Test what OpenWebUI is currently sending (WRONG):**
```bash
# This FAILS because X-Admin-API-Key is not recognized on /v1/*
curl -v -H "X-Admin-API-Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789" \
  https://scrubgate.tech/v1/models

# Result: 401 Unauthorized
```

**Test what OpenWebUI SHOULD send (CORRECT):**
```bash
# This WORKS because Authorization Bearer is correct
curl -v -H "Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789" \
  https://scrubgate.tech/v1/models

# Result: 200 OK with model list
```

### Test 2: Check OpenWebUI Logs

**After applying fix, watch logs:**
```bash
ssh openweb "docker logs -f llm-proxy-backend 2>&1 | grep '172.18.0.2'"
```

**Should see:**
```json
{
  "level": "info",
  "message": "API request authenticated",
  "key_name": "OpenWebUI",
  "endpoint": "/v1/models",
  "status": 200
}
```

**Instead of:**
```json
{
  "level": "warn",
  "message": "Token validation failed",
  "status": 401
}
```

### Test 3: Live Monitor

After fix:
1. Open `https://scrubgate.tech:3005`
2. Go to **Live Monitor**
3. Should show: ✅ **Connected** (green)
4. Recent requests showing 200 status codes

---

## Why This Happened

### Common Misconception

The key name `sk-llm-proxy-openwebui-...` contains "proxy" which might make users think it should use a custom proxy header like `X-Admin-API-Key`.

**But that's not how it works!**

- The prefix `sk-llm-proxy-` is just the **key format requirement**
- It still needs to be sent in the **standard OpenAI header**: `Authorization: Bearer`

### OpenWebUI Confusion

OpenWebUI allows custom headers for proxy configurations. Users might add:
```json
{
  "X-Admin-API-Key": "..."
}
```

Thinking this is needed for the proxy. **It's not!**

For OpenAI-compatible APIs (which LLM-Proxy is), use the standard:
```
Authorization: Bearer <api-key>
```

---

## Correct Configuration Guide

### Standard OpenAI-Compatible Setup

**OpenWebUI Settings:**
```
Provider Type: OpenAI
Base URL: https://scrubgate.tech/v1
API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
Custom Headers: (empty)
```

**What OpenWebUI Sends:**
```http
POST /v1/chat/completions HTTP/1.1
Host: scrubgate.tech
Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
Content-Type: application/json
```

**What LLM-Proxy Expects:**
```http
Authorization: Bearer sk-llm-proxy-*
                      └─ Must start with this prefix
```

✅ **Perfect match!**

---

### Wrong Configuration (Current)

**OpenWebUI Settings:**
```
Provider Type: OpenAI
Base URL: https://scrubgate.tech/v1
API Key: (empty or other)
Custom Headers:
{
  "X-Admin-API-Key": "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
}
```

**What OpenWebUI Sends:**
```http
POST /v1/chat/completions HTTP/1.1
Host: scrubgate.tech
X-Admin-API-Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
Content-Type: application/json
```

**What LLM-Proxy Sees:**
- APIKeyMiddleware: "No Authorization header → skip"
- OAuthMiddleware: "No OAuth token → 401"

❌ **Mismatch!**

---

## Alternative: Modify LLM-Proxy (Not Recommended)

If for some reason you can't change OpenWebUI, we could modify the APIKeyMiddleware to also accept `X-Admin-API-Key` header for client keys.

**But this is confusing because:**
- `X-Admin-API-Key` suggests it's for admin operations
- Client keys should use standard `Authorization` header
- Breaks OpenAI compatibility

**Only do this if OpenWebUI absolutely can't be configured properly.**

---

## After Fix Checklist

- [ ] OpenWebUI using `Authorization: Bearer` header (not `X-Admin-API-Key`)
- [ ] API Key field contains: `sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789`
- [ ] Custom Headers field is empty
- [ ] Base URL is: `https://scrubgate.tech/v1`
- [ ] Test connection in OpenWebUI shows success
- [ ] Live Monitor shows green ✅ Connected
- [ ] Can send chat messages in OpenWebUI
- [ ] Logs show 200 status codes from 172.18.0.2

---

## Quick Commands

**Test current (wrong) configuration:**
```bash
curl -s -o /dev/null -w "%{http_code}\n" \
  -H "X-Admin-API-Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789" \
  https://scrubgate.tech/v1/models
# Expected: 401
```

**Test correct configuration:**
```bash
curl -s -o /dev/null -w "%{http_code}\n" \
  -H "Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789" \
  https://scrubgate.tech/v1/models
# Expected: 200
```

**Watch for OpenWebUI requests:**
```bash
ssh openweb "docker logs -f llm-proxy-backend 2>&1 | grep -E '(172.18.0.2|OpenWebUI)'"
```

---

## Summary

**Problem:** OpenWebUI uses `X-Admin-API-Key` header  
**Solution:** Change to `Authorization: Bearer` header  
**Reason:** APIKeyMiddleware only reads `Authorization` header  

**The key is correct, just the header name is wrong!**

**Fix:** In OpenWebUI settings, use the standard API Key field instead of custom headers.

**Priority:** 🔴 **CRITICAL** - This is why OpenWebUI can't connect!
