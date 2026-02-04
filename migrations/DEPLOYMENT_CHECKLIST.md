# LLM-Proxy Production Deployment Checklist

Use this checklist for EVERY production deployment to prevent schema mismatch issues.

---

## 📋 Pre-Deployment (Developer)

### Code Changes
- [ ] All code changes committed to Git
- [ ] Git tag created for release (e.g., `v1.2.0`)
- [ ] CHANGELOG.md updated with changes
- [ ] All tests passing locally
- [ ] Code reviewed and approved

### Database Migrations
- [ ] Check if new migrations exist in `migrations/` directory
- [ ] Migrations tested locally against PostgreSQL 14
- [ ] Migration rollback script prepared (if complex change)
- [ ] Migration documented with clear comments

### Docker Images
- [ ] Backend image built successfully
- [ ] Admin-UI image built successfully
- [ ] Images tagged with Git commit SHA and version
- [ ] Images pushed to GitLab Container Registry

---

## 🔧 Deployment Steps (Operations)

### Step 1: Backup Database
```bash
# SSH to production server
ssh openweb

# Create backup
docker exec llm-proxy-postgres pg_dump -U proxy_user -d llm_proxy > /backup/llm_proxy_$(date +%Y%m%d_%H%M%S).sql

# Verify backup file exists and has content
ls -lh /backup/llm_proxy_*.sql
```

**⚠️ DO NOT PROCEED if backup failed!**

---

### Step 2: Run Database Migrations

```bash
# Copy migration file to server (from local machine)
scp migrations/XXX_migration_name.sql openweb:/tmp/

# Connect and verify current schema
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c '\d oauth_clients'"

# Run migration
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/XXX_migration_name.sql"

# Verify migration success
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c '\d oauth_clients'"
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c '\d request_logs'"
```

**✅ Verify:** New columns exist and constraints are correct

---

### Step 3: Deploy Backend

```bash
ssh openweb

# Pull latest images
cd /path/to/docker-compose
docker-compose pull backend

# Stop old backend
docker-compose stop backend

# Start new backend
docker-compose up -d backend

# Check logs for errors
docker-compose logs -f backend --tail 100
```

**✅ Verify:** Backend started without database errors

---

### Step 4: Deploy Admin-UI (if changed)

```bash
# Pull latest admin-ui image
docker-compose pull admin-ui

# Restart admin-ui
docker-compose stop admin-ui
docker-compose up -d admin-ui

# Check status
docker-compose ps admin-ui
```

---

### Step 5: Verification Tests

#### Test 1: Health Check
```bash
curl -s https://llmproxy.aitrail.ch/health | jq .
```
**Expected:** `{"status": "healthy"}`

#### Test 2: Admin Login
- Open: https://llmproxy.aitrail.ch
- Login with admin key
- **Expected:** Login successful, no console errors

#### Test 3: API Endpoints
```bash
# Providers
curl -s -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  https://llmproxy.aitrail.ch/admin/providers | jq .

# Stats (the one that failed before!)
curl -s -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  https://llmproxy.aitrail.ch/admin/stats/usage | jq .

# Clients
curl -s -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  https://llmproxy.aitrail.ch/admin/clients | jq .
```
**Expected:** All return HTTP 200 with valid JSON

#### Test 4: Create Client (the other one that failed!)
- Navigate to "Clients" tab in Admin UI
- Click "Create New Client"
- Fill in:
  - Client ID: `test-client`
  - Client Secret: `test-secret-12345`
  - Name: `Test Client`
- Click "Create Client"
- **Expected:** Success message, client appears in list

---

### Step 6: Monitor for Issues

```bash
# Watch backend logs for 5 minutes
ssh openweb "docker logs llm-proxy-backend -f --tail 100"
```

**Watch for:**
- ❌ Database errors
- ❌ 500 errors
- ❌ Stack traces
- ✅ Normal request logs

---

## 🚨 Rollback Procedure

If deployment fails:

### Rollback Backend
```bash
ssh openweb

# Stop new backend
docker-compose stop backend

# Restore old image
docker-compose up -d backend@sha256:OLD_IMAGE_HASH

# Or restore from backup tag
docker tag llm-proxy-backend:backup-20260204 llm-proxy-backend:latest
docker-compose up -d backend
```

### Rollback Database (if needed)
```bash
ssh openweb

# Restore from backup
docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /backup/llm_proxy_TIMESTAMP.sql

# Or run migration rollback script
docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/rollback_XXX.sql
```

---

## 📊 Post-Deployment

### Update Documentation
- [ ] Update deployment date in CHANGELOG.md
- [ ] Document any issues encountered
- [ ] Update version in README.md

### Notify Team
- [ ] Notify team in Slack/Discord/Email
- [ ] Share deployment notes
- [ ] Document any manual steps taken

### Monitor
- [ ] Check server metrics (CPU, RAM, disk)
- [ ] Monitor error rates in logs
- [ ] Check response times
- [ ] Verify no user complaints

---

## 🔍 Common Issues & Solutions

### Issue: "column does not exist"
**Cause:** Migration not run before backend deployment  
**Solution:** Run migration, restart backend

### Issue: "null value violates not-null constraint"
**Cause:** Migration added column without default value  
**Solution:** Add default value or make column nullable

### Issue: Backend won't start
**Cause:** Database connection failure or migration issue  
**Solution:** Check logs, verify database credentials, rollback if needed

### Issue: 502 Bad Gateway
**Cause:** Backend container not running or Caddy misconfigured  
**Solution:** Check `docker ps`, verify Caddy config, check backend logs

---

## 📈 Deployment Metrics to Track

After each deployment, record:

- Deployment date/time
- Git commit SHA / version tag
- Migrations applied
- Downtime (if any)
- Issues encountered
- Time to resolve issues

**Goal:** Zero-downtime deployments with <5 min migration time

---

## ✅ Success Criteria

Deployment is successful when:

- ✅ All health checks pass
- ✅ No 500 errors in logs
- ✅ All admin UI features work
- ✅ Can create/edit clients
- ✅ Statistics page loads
- ✅ No database errors in logs
- ✅ Response times normal (<200ms)

---

## 📞 Emergency Contacts

If deployment goes wrong:

1. **STOP** - Don't make it worse
2. **ASSESS** - Check logs, error messages
3. **ROLLBACK** - Restore previous working state
4. **NOTIFY** - Alert team
5. **DEBUG** - Fix issue in development first
6. **RETRY** - Re-deploy after fix verified

**Remember:** It's better to rollback and fix properly than to patch production!
