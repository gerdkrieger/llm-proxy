# Database Migrations

This directory contains SQL migration scripts for the LLM-Proxy database schema.

## ⚠️ IMPORTANT: Production Deployment Process

**Database migrations MUST be run BEFORE deploying new backend versions!**

### Correct Deployment Order:
1. ✅ Run database migration SQL
2. ✅ Verify migration success
3. ✅ Deploy new backend image
4. ✅ Verify endpoints work

### ❌ WRONG (causes 500 errors):
1. ❌ Deploy new backend image
2. ❌ Backend expects new columns that don't exist
3. ❌ 500 errors: "column does not exist"
4. ❌ Manual hotfix required

---

## 🐛 Incident: 2026-02-04 Production Outage

### What Happened
- **Duration**: ~2 hours
- **Impact**: Client creation and usage statistics endpoints returned 500 errors
- **Root Cause**: Backend code updated to use new database columns, but schema was not migrated

### Errors Encountered
```
ERROR: column "client_secret_hash" of relation "oauth_clients" does not exist
ERROR: column "duration_ms" does not exist  
ERROR: null value in column "client_secret" violates not-null constraint
```

### Why This Happened
1. Backend image was built and deployed with new code
2. New code expected columns: `client_secret_hash`, `duration_ms`, `cost_usd`
3. Production database still had old schema
4. No automated migration process existed
5. No deployment checklist to catch this

### Resolution
Manual SQL execution on production database (risky, not repeatable).

### Lessons Learned
- ✅ Created this migrations directory with proper SQL scripts
- ✅ Created deployment checklist
- ✅ Documented migration process
- 🔄 TODO: Add automated migration runner to backend startup
- 🔄 TODO: Add database schema version tracking

---

## 📁 Migration Files

Migrations are numbered sequentially:

- `001_add_hash_and_stats_columns.sql` - Add client_secret_hash, cost_usd, duration_ms

**Naming Convention:**
```
<number>_<descriptive_name>.sql
```

---

## 🚀 How to Run Migrations

### Option 1: Manual Execution (Production)

**Connect to database:**
```bash
# Via Docker (recommended)
docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy

# Via SSH to server
ssh openweb
docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
```

**Run migration:**
```bash
# Copy SQL file to server
scp migrations/001_add_hash_and_stats_columns.sql openweb:/tmp/

# Execute on server
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/001_add_hash_and_stats_columns.sql"
```

**Verify:**
```bash
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c '\d oauth_clients'"
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c '\d request_logs'"
```

### Option 2: Automated (Development)

```bash
cd /path/to/llm-proxy
docker-compose up -d postgres
cat migrations/*.sql | docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy
```

---

## ✅ Pre-Deployment Checklist

Before deploying a new backend version, **ALWAYS** check:

- [ ] Are there new migration files in `migrations/`?
- [ ] Have I run migrations on production database?
- [ ] Have I verified migrations succeeded?
- [ ] Does the backend log show successful database connection?
- [ ] Are all endpoints returning 200 (not 500)?

---

## 🔧 Migration Best Practices

### 1. Make Migrations Idempotent
Use `IF NOT EXISTS` and `DO $$` blocks:
```sql
ALTER TABLE my_table 
ADD COLUMN IF NOT EXISTS new_column VARCHAR(255);

DO $$ 
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns WHERE...) THEN
        -- migration logic
    END IF;
END $$;
```

### 2. Backwards Compatible Changes
When renaming or removing columns:
- Add new column first
- Migrate data
- Update backend to use new column
- (Later) Remove old column in separate migration

### 3. Test Locally First
```bash
# Start local postgres
docker run -d --name test-postgres -e POSTGRES_PASSWORD=test -p 5433:5432 postgres:14-alpine

# Run migration
psql -h localhost -p 5433 -U postgres -f migrations/001_xxx.sql

# Verify
psql -h localhost -p 5433 -U postgres -c '\d my_table'
```

### 4. Document Breaking Changes
In migration file header:
```sql
-- BREAKING CHANGE: This migration renames column X to Y
-- REQUIRES: Backend version >= v2.0.0
-- ROLLBACK: ALTER TABLE foo RENAME COLUMN Y TO X;
```

---

## 🔄 Future: Automated Migrations

### Recommended Tools:
1. **golang-migrate** - https://github.com/golang-migrate/migrate
2. **goose** - https://github.com/pressly/goose
3. **Flyway** - https://flywaydb.org/ (Java-based)

### Implementation Plan:
```go
// In main.go before starting HTTP server
import "github.com/golang-migrate/migrate/v4"

func runMigrations() error {
    m, err := migrate.New(
        "file://migrations",
        "postgres://user:pass@host/db",
    )
    if err != nil {
        return err
    }
    
    if err := m.Up(); err != nil && err != migrate.ErrNoChange {
        return err
    }
    
    return nil
}
```

---

## 📊 Schema Version Tracking

Create a `schema_migrations` table:
```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

Track applied migrations:
```sql
INSERT INTO schema_migrations (version) VALUES ('001_add_hash_and_stats_columns');
```

---

## 🆘 Emergency Rollback

If a migration causes issues:

```sql
-- Rollback 001_add_hash_and_stats_columns.sql
BEGIN;

ALTER TABLE oauth_clients DROP COLUMN IF EXISTS client_secret_hash;
ALTER TABLE oauth_clients ALTER COLUMN client_secret SET NOT NULL;
ALTER TABLE request_logs DROP COLUMN IF EXISTS cost_usd;
ALTER TABLE request_logs RENAME COLUMN duration_ms TO latency_ms;

-- Verify rollback
\d oauth_clients
\d request_logs

COMMIT;  -- or ROLLBACK if something is wrong
```

---

## 📞 Support

If migrations fail in production:

1. **DO NOT PANIC** - Database is transactional
2. Check error message carefully
3. Verify current schema: `\d table_name`
4. Check if migration was partially applied
5. If unsure, ROLLBACK and ask for help

**Contact:** Check project README for maintainer info

---

## 📚 Additional Resources

- [PostgreSQL ALTER TABLE docs](https://www.postgresql.org/docs/14/sql-altertable.html)
- [Database Migrations Best Practices](https://martinfowler.com/articles/evodb.html)
- [Zero-Downtime Migrations](https://www.braintreepayments.com/blog/safe-operations-for-high-volume-postgresql/)
