# Request Logging Deployment Guide

**Date:** 2026-02-07  
**Feature:** Request Logging & Live Monitor Integration  
**Status:** Ready for Deployment  
**Priority:** HIGH (Fixes Live Monitor)

---

## Overview

This deployment adds **Production-Grade Request Logging** to LLM-Proxy:

✅ **All API requests logged to database** (not just LLM requests)  
✅ **Live Monitor now works** (previously showed 404)  
✅ **Authentication tracking** (API key name, auth type)  
✅ **Filter tracking** (which requests were blocked)  
✅ **Performance metrics** (response time, status codes)  

---

## What Changed

### 1. Database Changes
- **Migration 000007:** Adds columns to existing `request_logs` table
- **New columns:** `auth_type`, `api_key_name`, `was_filtered`, `filter_reason`
- **New indexes:** For efficient Live Monitor queries

### 2. Backend Changes
- **New Middleware:** `RequestLoggerMiddleware` (logs all requests)
- **New Endpoint:** `GET /admin/requests` (for Live Monitor)
- **Updated Repository:** `RequestLogRepository` (supports new fields)

### 3. Frontend Changes
- **Live Monitor:** Now fetches from `/admin/requests` instead of `/admin/logs`
- **Field mapping:** Uses `endpoint` instead of `path`

---

## Deployment Steps

### Step 1: Backup Database

**CRITICAL:** Always backup before migrations!

```bash
ssh openweb

# Backup PostgreSQL database
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > \
  /opt/llm-proxy/backups/llm_proxy_pre_migration_007_$(date +%Y%m%d_%H%M%S).sql

# Verify backup
ls -lh /opt/llm-proxy/backups/
```

---

### Step 2: Pull Latest Code

```bash
ssh openweb
cd /opt/llm-proxy

# Pull latest changes
git fetch origin
git checkout develop
git pull origin develop

# Verify you're on the right commit
git log -1 --oneline
# Should show: feat: Add request logging for Live Monitor
```

---

### Step 3: Run Database Migration

```bash
cd /opt/llm-proxy

# Check current migration status
docker exec llm-proxy-backend migrate -database "postgres://proxy_user:dev_password_2024@llm-proxy-postgres:5432/llm_proxy?sslmode=disable" -path /app/migrations version

# Run migration 000007
docker exec llm-proxy-backend migrate -database "postgres://proxy_user:dev_password_2024@llm-proxy-postgres:5432/llm_proxy?sslmode=disable" -path /app/migrations up

# Verify migration succeeded
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\d request_logs"
# Should show new columns: auth_type, api_key_name, was_filtered, filter_reason
```

**Expected Output:**
```
000007/u add_request_logs (1.234s)
```

---

### Step 4: Rebuild Backend

```bash
cd /opt/llm-proxy

# Rebuild backend with new code
docker compose -f deployments/docker-compose.openwebui.yml build backend

# Or use full path if needed
docker compose -f /opt/llm-proxy/deployments/docker-compose.openwebui.yml build backend
```

---

### Step 5: Rebuild Admin UI

```bash
cd /opt/llm-proxy

# Rebuild Admin UI with updated Live Monitor
docker compose -f deployments/docker-compose.openwebui.yml build admin-ui
```

---

### Step 6: Restart Services

```bash
cd /opt/llm-proxy

# Restart backend (includes new middleware)
docker compose -f deployments/docker-compose.openwebui.yml restart backend

# Restart admin-ui (updated Live Monitor)
docker compose -f deployments/docker-compose.openwebui.yml restart admin-ui

# Wait 10 seconds for startup
sleep 10

# Check all services are running
docker compose -f deployments/docker-compose.openwebui.yml ps
```

**Expected:** All services showing "Up"

---

### Step 7: Verify Deployment

#### Test 1: Check Migration

```bash
# Verify new columns exist
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "SELECT column_name FROM information_schema.columns WHERE table_name='request_logs' AND column_name IN ('auth_type', 'api_key_name', 'was_filtered', 'filter_reason');"
```

**Expected:** 4 rows showing the new columns

---

#### Test 2: Make a Test Request

```bash
# Make a request that will be logged
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789" \
  https://scrubgate.tech/v1/models
```

**Expected:** 200 OK with model list

---

#### Test 3: Check Request Was Logged

```bash
# Check database for the logged request
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "SELECT id, method, path, status_code, auth_type, api_key_name, created_at FROM request_logs ORDER BY created_at DESC LIMIT 5;"
```

**Expected Output:**
```
                  id                  | method |     path     | status_code | auth_type | api_key_name | created_at
--------------------------------------+--------+--------------+-------------+-----------+--------------+------------
 123e4567-e89b-12d3-a456-426614174000 | GET    | /v1/models   | 200         | api_key   | OpenWebUI    | 2026-02-07...
```

---

#### Test 4: Check Admin Endpoint

```bash
# Test the new /admin/requests endpoint
curl -s -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  "https://scrubgate.tech/admin/requests?limit=5" | jq '.logs | length'
```

**Expected:** Number (1-5) showing logged requests

---

#### Test 5: Check Live Monitor

1. Open: `https://scrubgate.tech:3005`
2. Login with admin key
3. Click: **🔴 Live Monitor** (green button at top)
4. Should see:
   - ✅ Recent requests in the table
   - ✅ Status codes (200, 401, etc.)
   - ✅ IP addresses
   - ✅ Timestamps
   - ✅ No errors in browser console

**Before this deployment:** Empty table, 404 errors  
**After this deployment:** Live requests showing!

---

#### Test 6: Check Backend Logs

```bash
# Watch backend logs for request logging
docker logs --tail=50 llm-proxy-backend 2>&1 | grep -i "logged request"
```

**Expected:**
```
DEBUG Logged request: GET /v1/models -> 200 (45ms)
DEBUG Logged request: GET /admin/requests -> 200 (12ms)
```

---

### Step 8: Monitor for Issues

Watch logs for 5-10 minutes:

```bash
# Terminal 1: Backend logs
docker logs -f llm-proxy-backend 2>&1

# Terminal 2: Make test requests
watch -n 5 'curl -s -o /dev/null -w "Status: %{http_code}\n" -H "Authorization: Bearer sk-llm-proxy-openwebui-2026..." https://scrubgate.tech/v1/models'
```

**Look for:**
- ✅ "Logged request" debug messages
- ✅ No "Failed to log request" warnings
- ✅ Response times normal (<100ms for /v1/models)
- ❌ NO database connection errors

---

## Rollback Plan

If something goes wrong:

### Rollback Step 1: Revert Code

```bash
ssh openweb
cd /opt/llm-proxy

# Find previous working commit
git log --oneline -10

# Checkout previous version
git checkout <previous-commit-hash>

# Rebuild and restart
docker compose -f deployments/docker-compose.openwebui.yml build backend admin-ui
docker compose -f deployments/docker-compose.openwebui.yml restart backend admin-ui
```

### Rollback Step 2: Revert Migration (if needed)

**ONLY if migration causes issues!**

```bash
# Rollback migration 000007
docker exec llm-proxy-backend migrate -database "postgres://proxy_user:dev_password_2024@llm-proxy-postgres:5432/llm_proxy?sslmode=disable" -path /app/migrations down 1

# Verify rollback
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\d request_logs"
# New columns should be gone
```

### Rollback Step 3: Restore Database (nuclear option)

**ONLY if database is corrupted!**

```bash
# Find backup file
ls -lh /opt/llm-proxy/backups/

# Restore from backup
cat /opt/llm-proxy/backups/llm_proxy_pre_migration_007_*.sql | \
  docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy
```

---

## Performance Impact

### Expected Performance

- **Request Logging:** ~1-2ms overhead per request (asynchronous)
- **Database Growth:** ~1KB per request
- **Disk Usage:** ~1GB per 1 million requests

### Monitoring

```bash
# Check database size
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "SELECT pg_size_pretty(pg_total_relation_size('request_logs'));"

# Check request_logs row count
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "SELECT COUNT(*) FROM request_logs;"

# Check recent log rate
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "SELECT COUNT(*) FROM request_logs WHERE created_at > NOW() - INTERVAL '1 hour';"
```

---

## Data Retention (Optional)

To prevent database from growing indefinitely:

### Option 1: Manual Cleanup

```bash
# Delete logs older than 30 days
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL '30 days';"

# Vacuum to reclaim space
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "VACUUM ANALYZE request_logs;"
```

### Option 2: Automated Cleanup (Cron)

```bash
# Add to crontab
crontab -e

# Add this line (runs daily at 3 AM)
0 3 * * * docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL '30 days'; VACUUM ANALYZE request_logs;"
```

---

## Troubleshooting

### Problem: "Failed to log request to database"

**Check 1:** Database connection

```bash
docker exec llm-proxy-backend sh -c 'echo "SELECT 1" | psql $DATABASE_URL'
```

**Check 2:** Migration ran successfully

```bash
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\d request_logs"
# Must show auth_type, api_key_name, was_filtered, filter_reason
```

**Check 3:** Repository initialized

```bash
docker logs llm-proxy-backend 2>&1 | grep -i "request.*log.*repo"
```

---

### Problem: Live Monitor shows empty

**Check 1:** Requests are being logged

```bash
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
  "SELECT COUNT(*) FROM request_logs WHERE created_at > NOW() - INTERVAL '5 minutes';"
```

**Check 2:** Admin endpoint works

```bash
curl -H "X-Admin-API-Key: admin_dev_key_12345..." \
  https://scrubgate.tech/admin/requests?limit=5
```

**Check 3:** Browser console

- Open Live Monitor
- Press F12 (Developer Tools)
- Check Console for errors
- Check Network tab for failed requests

---

### Problem: Migration fails

**Error:** `column "auth_type" already exists`

**Solution:** Migration already ran, just restart:

```bash
docker compose -f deployments/docker-compose.openwebui.yml restart backend admin-ui
```

**Error:** `relation "request_logs" does not exist`

**Solution:** Need to run init migration first:

```bash
# Run all pending migrations
docker exec llm-proxy-backend migrate -database "postgres://..." -path /app/migrations up
```

---

## Security Notes

### What is Logged

✅ **Safe to log:**
- Request method (GET, POST, etc.)
- Request path (/v1/models, etc.)
- Status code (200, 401, etc.)
- IP address
- User-Agent
- **API key NAME** (e.g., "OpenWebUI")
- Auth type ("api_key", "oauth", "admin")
- Response time
- Model name
- Provider name

❌ **NEVER logged:**
- Actual API keys (only names!)
- Request bodies (may contain PII)
- Response bodies (may contain PII)
- Client secrets
- OAuth tokens

### Privacy Compliance

**GDPR Considerations:**
- IP addresses are logged (consider as PII)
- User agents are logged
- Implement 30-day retention policy
- Document in privacy policy

---

## Summary

### Before Deployment

❌ Live Monitor shows 404 error  
❌ No visibility into API requests  
❌ Can't debug OpenWebUI connection issues  

### After Deployment

✅ Live Monitor shows real-time requests  
✅ Full request history in database  
✅ Authentication tracking (know which client made request)  
✅ Filter tracking (see what was blocked)  
✅ Performance metrics (response times)  
✅ Debugging capability (see exactly what OpenWebUI sends)  

---

## Next Steps

After successful deployment:

1. **Test OpenWebUI Fix**
   - Follow `docs/guides/OPENWEBUI_CONNECTION_SETUP.md`
   - Fix header from `X-Admin-API-Key` to `Authorization: Bearer`
   - Verify connection in Live Monitor

2. **Monitor Performance**
   - Watch database size growth
   - Check response time impact
   - Set up retention policy

3. **Use Live Monitor**
   - Debug authentication issues
   - Monitor API usage patterns
   - Identify slow requests
   - Track filter effectiveness

---

**Deployment Completed:** _______ (Date/Time)  
**Deployed By:** _______  
**Verified By:** _______  

**Issues Encountered:** (None / List)

**Rollback Required:** (Yes / No)

---

**Questions?** Contact DevOps team or check logs!
