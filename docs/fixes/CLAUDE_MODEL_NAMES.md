# Claude Model Names Fix

**Date:** 2026-02-07  
**Issue:** `model: claude-3-opus-latest` not found  
**Error:** HTTP 404 from Anthropic API

---

## Problem

**Error Message:**
```json
{
  "type": "error",
  "error": {
    "type": "not_found_error",
    "message": "model: claude-3-opus-latest"
  }
}
```

**Reason:** The model name `claude-3-opus-latest` **does not exist**!

---

## Understanding Anthropic Model Names

**Anthropic does NOT use "-latest" suffix!**

Every Claude model has a **specific date version**:

### ❌ WRONG Model Names
```
claude-3-opus-latest        # Does not exist!
claude-3-sonnet-latest      # Does not exist!
claude-3-5-sonnet-latest    # Does not exist!
claude-opus                 # Does not exist!
```

### ✅ CORRECT Model Names

```yaml
# Claude 3.5 Sonnet (Latest & Best)
claude-3-5-sonnet-20241022  # ⭐ RECOMMENDED (Oct 2024 release)
claude-3-5-sonnet-20240620  # Previous version (Jun 2024)

# Claude 3 Opus (Most Capable)
claude-3-opus-20240229      # ⭐ Use this instead of "opus-latest"

# Claude 3 Sonnet (Balanced)
claude-3-sonnet-20240229

# Claude 3 Haiku (Fast & Cheap)
claude-3-haiku-20240307
```

---

## Solution

### Option 1: Use Claude 3.5 Sonnet (Recommended) ⭐

**Replace:**
```
claude-3-opus-latest
```

**With:**
```
claude-3-5-sonnet-20241022
```

**Why:**
- ✅ Latest model (October 2024)
- ✅ Best performance
- ✅ Better than Opus for most tasks
- ✅ Faster than Opus
- ✅ Cheaper than Opus

---

### Option 2: Use Claude 3 Opus (If you specifically need Opus)

**Replace:**
```
claude-3-opus-latest
```

**With:**
```
claude-3-opus-20240229
```

**Why:**
- ✅ Most capable Claude 3 model
- ✅ Best for complex reasoning
- ⚠️ Slower than Sonnet
- ⚠️ More expensive

---

## Quick Fix

### If using OpenWebUI

1. Open OpenWebUI
2. Go to model selector
3. Change from: `claude-3-opus-latest`
4. To: `claude-3-5-sonnet-20241022` (recommended)
5. OR: `claude-3-opus-20240229` (if you need Opus)

### If using API directly

**Before:**
```json
{
  "model": "claude-3-opus-latest",
  "messages": [...]
}
```

**After:**
```json
{
  "model": "claude-3-5-sonnet-20241022",
  "messages": [...]
}
```

### If using Cursor IDE

1. Open Cursor Settings
2. Find Model Configuration
3. Change to: `claude-3-5-sonnet-20241022`

---

## Updated LLM-Proxy Configuration

**File:** `configs/config.yaml`

**Before:**
```yaml
providers:
  claude:
    models:
      - claude-3-haiku-20240307
```

**After:**
```yaml
providers:
  claude:
    models:
      - claude-3-5-sonnet-20241022  # Latest & Best (empfohlen)
      - claude-3-5-sonnet-20240620  # Previous Sonnet
      - claude-3-opus-20240229      # Most capable (NOT "latest"!)
      - claude-3-sonnet-20240229    # Balanced
      - claude-3-haiku-20240307     # Fast & cheap
```

**After changing config:**
```bash
# Restart backend to load new models
docker compose -f deployments/docker-compose.openwebui.yml restart backend
```

---

## Model Comparison

| Model | Release | Best For | Speed | Cost | Context |
|-------|---------|----------|-------|------|---------|
| **claude-3-5-sonnet-20241022** ⭐ | Oct 2024 | Most tasks | Fast | Medium | 200K |
| claude-3-5-sonnet-20240620 | Jun 2024 | General | Fast | Medium | 200K |
| claude-3-opus-20240229 | Feb 2024 | Complex reasoning | Slow | High | 200K |
| claude-3-sonnet-20240229 | Feb 2024 | Balanced | Medium | Low | 200K |
| claude-3-haiku-20240307 | Mar 2024 | Simple tasks | Very Fast | Very Low | 200K |

**Recommendation:** Use `claude-3-5-sonnet-20241022` for 90% of use cases!

---

## How to Find Available Models

### Method 1: Check Anthropic Documentation

https://docs.anthropic.com/en/docs/models-overview

### Method 2: Check LLM-Proxy Admin UI

1. Open: `https://scrubgate.tech:3005`
2. Login with admin key
3. Go to: **Providers** or **Models** section
4. See: List of configured models

### Method 3: API Call

```bash
# List available models via LLM-Proxy
curl -H "Authorization: Bearer sk-llm-proxy-..." \
  https://scrubgate.tech/v1/models | jq '.data[] | select(.id | startswith("claude"))'
```

---

## Deployment

If you updated `configs/config.yaml`:

```bash
# On production server
ssh openweb
cd /opt/llm-proxy

# Pull latest config
git pull origin develop

# Restart backend to load new models
docker compose -f deployments/docker-compose.openwebui.yml restart backend

# Verify models are loaded
docker logs llm-proxy-backend 2>&1 | grep -i "claude"
```

---

## Testing

### Test 1: Check Model is Available

```bash
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  https://scrubgate.tech/v1/models | jq '.data[] | select(.id == "claude-3-5-sonnet-20241022")'
```

**Expected:** Model object returned

### Test 2: Make Request with Correct Model

```bash
curl -X POST https://scrubgate.tech/v1/chat/completions \
  -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-5-sonnet-20241022",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

**Expected:** 200 OK with response

### Test 3: Try Wrong Model (Should Fail)

```bash
curl -X POST https://scrubgate.tech/v1/chat/completions \
  -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-latest",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

**Expected:** 404 Not Found (model doesn't exist)

---

## Common Mistakes

### Mistake 1: Using "-latest" suffix

```yaml
❌ WRONG: claude-3-opus-latest
✅ RIGHT: claude-3-opus-20240229
```

### Mistake 2: Forgetting the date

```yaml
❌ WRONG: claude-opus
❌ WRONG: claude-3-opus
✅ RIGHT: claude-3-opus-20240229
```

### Mistake 3: Using old model names

```yaml
❌ OUTDATED: claude-2.1
❌ OUTDATED: claude-instant-1.2
✅ CURRENT: claude-3-5-sonnet-20241022
```

### Mistake 4: Case sensitivity

```yaml
❌ WRONG: Claude-3-Opus-20240229  (uppercase)
❌ WRONG: CLAUDE-3-OPUS-20240229  (uppercase)
✅ RIGHT: claude-3-opus-20240229  (lowercase)
```

---

## Model Naming Convention

Anthropic uses this pattern:

```
claude-{generation}-{variant}-{YYYYMMDD}

Examples:
claude-3-5-sonnet-20241022
  |   | |   |      |
  |   | |   |      └─ Release date (Oct 22, 2024)
  |   | |   └──────── Variant (sonnet, opus, haiku)
  |   | └──────────── Sub-generation (5)
  |   └────────────── Generation (3)
  └────────────────── Brand (claude)
```

**Never:**
- No "-latest" suffix
- No version numbers like "v2"
- No shortcuts like "opus" without generation

---

## Why Anthropic Uses Dates

**Benefits:**
- ✅ Clear versioning
- ✅ No ambiguity about which model
- ✅ Reproducible results
- ✅ Can run older versions if needed

**Example:**
- App tested with: `claude-3-5-sonnet-20240620`
- New release: `claude-3-5-sonnet-20241022`
- Can keep using old version if needed!

---

## Model Aliases (Future Feature)

**Note:** Some proxies support aliases like:

```yaml
aliases:
  claude-opus-latest: claude-3-opus-20240229
  claude-sonnet-latest: claude-3-5-sonnet-20241022
```

**But LLM-Proxy does not (yet).** Use exact names!

---

## Summary

### The Problem
```
❌ You tried: claude-3-opus-latest
❌ Error: model not found
```

### The Solution
```
✅ Use: claude-3-5-sonnet-20241022 (recommended)
✅ Or: claude-3-opus-20240229 (if you need Opus)
```

### How to Fix
1. Update client to use correct model name
2. OR update config to include all Claude models
3. Restart backend
4. Test with curl or OpenWebUI

---

## Related Documentation

- [Anthropic Models](https://docs.anthropic.com/en/docs/models-overview)
- [OpenAI API Compatibility](https://docs.anthropic.com/en/api/openai-api-compatibility)
- [LLM-Proxy Provider Configuration](../guides/PROVIDER_CONFIGURATION.md)

---

**Quick Fix:** Change `claude-3-opus-latest` to `claude-3-5-sonnet-20241022` ✅
