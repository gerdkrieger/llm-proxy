# OpenWebUI Authentication Fix Guide

**Date:** 2026-02-07  
**Issue:** OpenWebUI not connecting to LLM-Proxy  
**Root Cause:** Incorrect or missing API key configuration  
**Status:** Ready to Fix

---

## Problem Summary

OpenWebUI is currently NOT successfully connecting to the LLM-Proxy. Analysis shows:

✅ **LLM-Proxy is running** (backend, database, Redis all healthy)  
✅ **Filtering is working** (OCR dependencies installed, tested successfully)  
✅ **Authentication separation is correct** (admin keys ≠ client keys)  
❌ **OpenWebUI has wrong/missing API key**

**Evidence:**
- No recent requests from OpenWebUI IP (172.18.0.2) in backend logs
- OpenWebUI environment variables show empty `OPENAI_API_KEY`
- Live Monitor shows no active connections

---

## Solution: Configure Correct Client API Key

### Step 1: Verify Current Configuration

**Check OpenWebUI Admin Panel:**
1. Open: `https://chat.aitrail.ch`
2. Login as admin
3. Navigate to: **Admin Panel → Settings → Connections**
4. Look for **OpenAI API** section

**Expected to see:**
- Base URL: `https://llmproxy.aitrail.ch/v1` ✅
- API Key: `sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789` ✅

**Likely problem:**
- API Key is empty, wrong, or using admin key

---

### Step 2: Get the Correct Client API Key

**The correct key for OpenWebUI is:**
```
sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

**Verify it exists in config:**
```bash
ssh openweb "grep -A 2 'sk-llm-proxy-openwebui' /opt/llm-proxy/configs/config.yaml"
```

**Expected output:**
```yaml
- key: "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
  name: "OpenWebUI"
  scopes: ["read", "write"]
```

---

### Step 3: Fix OpenWebUI Configuration

**Option A: Via Web Interface (Recommended)**

1. Open `https://chat.aitrail.ch`
2. Login as admin
3. Go to **Admin Panel → Settings → Connections**
4. Find **OpenAI API** section
5. Configure:
   ```
   Base URL: https://llmproxy.aitrail.ch/v1
   API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
   ```
6. Click **Save**
7. Test connection by clicking **Test**

**Option B: Via Environment Variables**

If web interface doesn't work, update Docker Compose:

```bash
ssh openweb
cd /opt/open-webui

# Edit docker-compose.yml
nano docker-compose.yml
```

Add these environment variables:
```yaml
services:
  open-webui:
    environment:
      - OPENAI_API_BASE_URL=https://llmproxy.aitrail.ch/v1
      - OPENAI_API_KEY=sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

Restart OpenWebUI:
```bash
docker compose restart
```

---

### Step 4: Verify Fix

**Test 1: Check Logs**
```bash
# Watch for new requests from OpenWebUI
ssh openweb "docker logs -f llm-proxy-backend 2>&1 | grep '172.18.0.2'"
```

**Expected to see:**
```json
{
  "level": "info",
  "message": "API request authenticated",
  "key_name": "OpenWebUI",
  "endpoint": "/v1/models",
  "remote_addr": "172.18.0.2",
  "status": 200
}
```

**Test 2: Check Live Monitor**
1. Open `https://llmproxy.aitrail.ch:3005`
2. Login with admin key
3. Click **🔴 Live Monitor**
4. Should see:
   ```
   ✅ Connected
   Client: OpenWebUI (172.18.0.2)
   Recent requests showing 200 status codes
   ```

**Test 3: Try Chat in OpenWebUI**
1. Open `https://chat.aitrail.ch`
2. Start a new chat
3. Select a model (e.g., GPT-4)
4. Send a message: "Hello, please respond"
5. Should receive response from LLM

**Test 4: Verify Filtering**
1. In OpenWebUI, try sending: "My email is test@example.com"
2. Check Admin UI → Filters → Blocked Requests
3. Should see the request with EMAIL filter triggered

---

## Common Mistakes to Avoid

### ❌ DON'T: Use Admin Key for OpenWebUI

**WRONG:**
```yaml
OPENAI_API_KEY: admin_dev_key_12345...
```

**Why it fails:**
- Admin keys are for `/admin/*` routes only
- Admin keys CANNOT access `/v1/*` (LLM API)
- OpenWebUI needs `/v1/models` and `/v1/chat/completions`

**Result:** All requests return `401 Unauthorized`

---

### ❌ DON'T: Forget `sk-llm-proxy-` Prefix

**WRONG:**
```yaml
OPENAI_API_KEY: openwebui-2026-01-30-secure-key-abc123xyz789
```

**Why it fails:**
- Client keys MUST start with `sk-llm-proxy-`
- APIKeyMiddleware only processes keys with this prefix
- Without prefix, request goes to OAuth middleware (which also fails)

**Result:** `401 Invalid API key`

---

### ❌ DON'T: Use Wrong Base URL

**WRONG:**
```yaml
OPENAI_API_BASE_URL: https://llmproxy.aitrail.ch
```

**Why it fails:**
- Missing `/v1` path prefix
- OpenWebUI appends `/chat/completions` → `https://llmproxy.aitrail.ch/chat/completions`
- LLM-Proxy expects: `https://llmproxy.aitrail.ch/v1/chat/completions`

**Result:** `404 Not Found`

---

### ❌ DON'T: Use Disabled Key

**Check if key is enabled:**
```bash
ssh openweb "grep -A 4 'sk-llm-proxy-openwebui' /opt/llm-proxy/configs/config.yaml"
```

**Must show:**
```yaml
enabled: true  # ← Must be true!
```

**If disabled:**
```yaml
enabled: false  # ← This blocks all requests!
```

**Result:** `401 Unauthorized`

---

## Verification Checklist

After applying the fix, verify these items:

- [ ] OpenWebUI can list models
  ```bash
  # Should show gpt-4, gpt-3.5-turbo, claude-3-opus, etc.
  ```

- [ ] OpenWebUI can send chat messages
  ```bash
  # Send "Hello" and get response
  ```

- [ ] Logs show successful requests from 172.18.0.2
  ```bash
  ssh openweb "docker logs llm-proxy-backend 2>&1 | grep '172.18.0.2' | tail -5"
  # Should show status 200
  ```

- [ ] Live Monitor shows connection
  ```bash
  # Visit https://llmproxy.aitrail.ch:3005
  # See green ✅ Connected status
  ```

- [ ] No 401 errors in OpenWebUI logs
  ```bash
  ssh openweb "docker logs open-webui 2>&1 | grep -i 'unauthorized\|401' | tail -5"
  # Should be empty or old entries only
  ```

- [ ] PII filtering is working
  ```bash
  # Send message with email in OpenWebUI
  # Check Admin UI → Filters → Blocked Requests
  # Should see the filtered request
  ```

---

## Troubleshooting

### Problem: Still getting 401 after fix

**Check 1: Correct key in config**
```bash
ssh openweb "grep -A 3 'sk-llm-proxy-openwebui' /opt/llm-proxy/configs/config.yaml"
```

**Check 2: Backend has latest config**
```bash
ssh openweb "docker compose -f docker-compose.openwebui.yml restart backend"
# Wait 10 seconds for restart
```

**Check 3: OpenWebUI using correct key**
```bash
# Test manually with curl
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  https://llmproxy.aitrail.ch/v1/models
# Should return 200 with model list
```

**Check 4: View detailed logs**
```bash
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep -A 5 '401' | tail -20"
# Look for "Token validation failed" or "Invalid API key"
```

---

### Problem: Connection timeout

**Check 1: Backend is running**
```bash
ssh openweb "docker ps | grep llm-proxy-backend"
# Should show "Up" status
```

**Check 2: Network connectivity**
```bash
ssh openweb "docker exec open-webui ping -c 3 llm-proxy-backend"
# Should succeed
```

**Check 3: Port is listening**
```bash
ssh openweb "docker exec llm-proxy-backend netstat -tlnp | grep 8080"
# Should show LISTEN on port 8080
```

---

### Problem: Models not showing

**Check 1: Models are configured**
```bash
ssh openweb "grep -A 10 'providers:' /opt/llm-proxy/configs/config.yaml | head -20"
# Should show OpenAI and Anthropic providers
```

**Check 2: Test models endpoint directly**
```bash
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  https://llmproxy.aitrail.ch/v1/models | jq .
# Should return list of models
```

**Check 3: OpenWebUI has 'read' scope**
```yaml
# In config.yaml
scopes: ["read", "write"]  # ← Must include "read"
```

---

## Quick Reference

### Correct Configuration

**OpenWebUI Settings:**
```
Base URL: https://llmproxy.aitrail.ch/v1
API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

**Testing Command:**
```bash
# Test authentication (should return 200)
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789" \
  https://llmproxy.aitrail.ch/v1/models

# Watch for OpenWebUI requests
ssh openweb "docker logs -f llm-proxy-backend 2>&1 | grep '172.18.0.2'"
```

**Live Monitor:**
```
URL: https://llmproxy.aitrail.ch:3005
Login: YOUR_ADMIN_API_KEY_HERE
Navigate to: Live Monitor (top menu)
```

---

## After Fix is Applied

Once OpenWebUI is successfully connecting:

1. **Monitor for 24 hours**
   - Check Live Monitor shows consistent requests
   - Verify no 401 errors
   - Confirm filtering is working

2. **Test all features**
   - List models
   - Chat completions
   - Streaming responses
   - PII filtering
   - PDF attachment filtering (if applicable)

3. **Document the working configuration**
   - Take screenshots of OpenWebUI settings
   - Save curl commands that work
   - Note any issues for future reference

4. **Update monitoring**
   - Add alert if OpenWebUI connection drops
   - Monitor error rate in Live Monitor
   - Check filter statistics weekly

---

## Related Documentation

- [Authentication Architecture Guide](../guides/AUTHENTICATION_ARCHITECTURE.md) - Complete auth system explanation
- [OpenWebUI Connection Fix](../OPENWEBUI_CONNECTION_FIX.md) - Original troubleshooting doc
- [Complete Monitoring Guide](../guides/COMPLETE_MONITORING_GUIDE.md) - How to monitor the system
- [OpenWebUI Setup Guide](../guides/OPENWEBUI_SETUP.md) - Initial setup instructions

---

**Priority:** HIGH  
**Complexity:** Low (simple configuration change)  
**Risk:** None (no code changes required)  
**Estimated Time:** 5-10 minutes

**Next Steps:**
1. Open OpenWebUI web interface
2. Update API key in settings
3. Test connection
4. Verify in Live Monitor
5. Done! 🎉
