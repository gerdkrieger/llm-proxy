# Database Migration System

**Automated database migration execution for LLM-Proxy deployment**

---

## Table of Contents

- [Overview](#overview)
- [Why This Matters](#why-this-matters)
- [Migration Flow](#migration-flow)
- [Quick Start](#quick-start)
- [Commands Reference](#commands-reference)
- [How It Works](#how-it-works)
- [Troubleshooting](#troubleshooting)
- [Recovery Procedures](#recovery-procedures)
- [Creating New Migrations](#creating-new-migrations)

---

## Overview

The LLM-Proxy deployment system now includes **automatic database migration execution** to prevent production outages caused by schema mismatches. Migrations run automatically before deploying new container images.

### Key Features

✅ **Automatic execution** - Migrations run before container deployment  
✅ **Auto-rollback** - Database restored from backup if migrations fail  
✅ **Version tracking** - Uses `schema_migrations` table to track state  
✅ **Dirty state detection** - Identifies and helps recover from failed migrations  
✅ **Manual control** - Tools for checking status, applying, and rolling back migrations  
✅ **Zero downtime** - Migrations run while old containers are still serving traffic  

---

## Why This Matters

### The Problem We Solved

**2026-02-04 Production Outage** (2 hours downtime):
- Backend deployed with new code expecting new database columns
- Migrations were NOT run before deployment
- Application crashed with errors: `column client_secret_hash does not exist`
- Required emergency SSH access and manual SQL execution to fix

### The Solution

**Automated Migration Flow**:
```
1. Backup Database (automatic)
2. Run Migrations (automatic)
3. Deploy Containers (automatic)
4. Health Checks (automatic)

If migrations fail → Restore backup + Abort deployment
```

**Result**: No more schema mismatch outages. Migrations are ALWAYS run before deployment.

---

## Migration Flow

### Full Deployment with Migrations

```
┌─────────────────────────────────────────────────────────────┐
│                    make release VERSION=v1.2.0              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  LOCAL: Build & Push Images                                 │
│  ├─ Build backend, frontend, nginx images                   │
│  ├─ Tag with version                                         │
│  └─ Push to GitHub CR + GitLab CR                           │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  SERVER: Deploy with Migrations                              │
│                                                              │
│  1. PRE-CHECKS                                               │
│     ├─ Verify compose file exists                           │
│     ├─ Verify .env file exists                              │
│     ├─ Check PostgreSQL is running                          │
│     ├─ Test database connectivity                           │
│     └─ Ensure network exists                                │
│                                                              │
│  2. BACKUP DATABASE                                          │
│     ├─ Create timestamped backup                            │
│     ├─ Store in /opt/llm-proxy-backups/YYYYMMDD/            │
│     └─ Verify backup size                                   │
│                                                              │
│  3. RUN MIGRATIONS ⭐ NEW!                                   │
│     ├─ Check current version                                │
│     ├─ List pending migrations                              │
│     ├─ Run golang-migrate in Docker                         │
│     ├─ Verify new version                                   │
│     │                                                        │
│     └─ IF FAIL:                                              │
│         ├─ Restore database from backup                     │
│         ├─ Abort deployment                                 │
│         └─ Exit with error                                  │
│                                                              │
│  4. PULL NEW IMAGES                                          │
│     ├─ Pull backend:VERSION                                 │
│     ├─ Pull frontend:VERSION                                │
│     └─ Pull nginx:VERSION                                   │
│                                                              │
│  5. DEPLOY CONTAINERS                                        │
│     ├─ docker compose up -d                                 │
│     └─ Wait for startup                                     │
│                                                              │
│  6. HEALTH CHECKS                                            │
│     ├─ Check backend health endpoint                        │
│     ├─ Check frontend accessibility                         │
│     └─ Verify all containers running                        │
│                                                              │
│  7. CLEANUP                                                  │
│     └─ Remove old images                                    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ✅ DEPLOYMENT COMPLETE
```

### Migration Execution Details

```
┌─────────────────────────────────────────────────────────────┐
│  MIGRATION EXECUTION (Step 3)                                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Check Current Version                                       │
│  └─ Query: SELECT version FROM schema_migrations            │
│     Result: 000013 (current version)                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  List Pending Migrations                                     │
│  └─ Compare /opt/llm-proxy/migrations/*.up.sql              │
│     with current version                                     │
│     Result: 000014, 000015 (pending)                         │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  Run golang-migrate                                          │
│  └─ docker run --rm --network llm-proxy-network \            │
│       -v /opt/llm-proxy/migrations:/migrations \             │
│       migrate/migrate:v4.17.0 \                              │
│       -path=/migrations \                                    │
│       -database "postgres://user:pass@host/db?sslmode=..." \ │
│       up                                                     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌─────────┴─────────┐
                    │                   │
                SUCCESS              FAILURE
                    │                   │
                    ▼                   ▼
        ┌───────────────────┐   ┌──────────────────┐
        │ Verify New Version│   │ Restore Backup   │
        │ 000015 ✓          │   │ Abort Deployment │
        │ Continue Deploy   │   │ Exit with Error  │
        └───────────────────┘   └──────────────────┘
```

---

## Quick Start

### Check Migration Status

```bash
# Check current migration version on server
make migrate-status

# List pending migrations
make migrate-pending
```

### Apply Migrations Manually

```bash
# Apply all pending migrations
make migrate-up

# Apply only the next migration
make migrate-up-one
```

### Full Deployment (Automatic Migrations)

```bash
# Build, push, and deploy with automatic migrations
make release VERSION=v1.2.0
```

### Sync Migrations to Server

```bash
# After creating new migrations, sync to server
make migrate-sync
```

---

## Commands Reference

### Makefile Commands

| Command | Description |
|---------|-------------|
| `make migrate-status` | Check current migration version on server |
| `make migrate-pending` | List pending migrations on server |
| `make migrate-up` | Apply all pending migrations |
| `make migrate-up-one` | Apply only the next migration |
| `make migrate-down` | Rollback last migration (CAREFUL!) |
| `make migrate-create NAME=...` | Create new migration files |
| `make migrate-sync` | Sync migrations to production server |

### Direct Script Usage

```bash
# Check status
./scripts/deployment/migrate.sh status

# List pending
./scripts/deployment/migrate.sh pending

# Apply migrations
./scripts/deployment/migrate.sh up
./scripts/deployment/migrate.sh up 1    # Apply next 1 migration

# Rollback migrations
./scripts/deployment/migrate.sh down
./scripts/deployment/migrate.sh down 1  # Rollback last 1 migration

# Create new migration
./scripts/deployment/migrate.sh create add_user_roles

# Sync to server
./scripts/deployment/migrate.sh sync

# Force version (dirty state recovery)
./scripts/deployment/migrate.sh force 000013
```

---

## How It Works

### Migration Files

Migrations are stored in `migrations/` directory:

```
migrations/
├── 000001_init.up.sql                    # Initial schema
├── 000001_init.down.sql                  # Rollback for 000001
├── 000002_add_users.up.sql               # Add users table
├── 000002_add_users.down.sql             # Rollback for 000002
├── ...
└── 000013_fix_provider_configs_uuid.up.sql
```

**Naming Convention**: `{version}_{description}.{up|down}.sql`

- `version`: 6-digit number (000001, 000002, ...)
- `description`: Snake_case description
- `up.sql`: Forward migration (apply changes)
- `down.sql`: Reverse migration (rollback changes)

### Version Tracking

Migrations are tracked in the `schema_migrations` table:

```sql
CREATE TABLE schema_migrations (
    version BIGINT PRIMARY KEY,
    dirty BOOLEAN NOT NULL
);
```

- **version**: Current migration version (e.g., 000013)
- **dirty**: `true` if migration failed mid-execution, `false` if clean

### Migration Tool

We use **golang-migrate** (v4.17.0) in a Docker container:

```bash
docker run --rm \
  --network llm-proxy-network \
  -v /opt/llm-proxy/migrations:/migrations \
  migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database "postgres://user:pass@host/db?sslmode=disable" \
  up
```

**Why Docker?**
- Consistent environment (no local installation needed)
- Same tool version everywhere (v4.17.0)
- Network access to PostgreSQL container

### Automatic Backup

Before running migrations, the system automatically:

1. Creates timestamped backup: `/opt/llm-proxy-backups/YYYYMMDD/postgres-backup-HHMMSS.sql`
2. Verifies backup size
3. Stores backup path for potential rollback

**Backup retention**: Backups are kept indefinitely. Clean up old backups manually.

### Auto-Rollback

If migrations fail:

1. **Detect failure**: Check exit code of `migrate up`
2. **Restore database**: `psql < backup.sql`
3. **Abort deployment**: Exit with error, don't deploy containers
4. **Log error**: Show which migration failed

---

## Troubleshooting

### Problem: "Dirty database version"

**Symptom**:
```
error: Dirty database version 000014. Fix and force version.
```

**Cause**: A migration failed mid-execution, leaving the database in an inconsistent state.

**Solution**:

1. **Check what failed**:
   ```bash
   make migrate-status
   # Shows: Version: 000014 (dirty)
   ```

2. **Manually inspect database**:
   ```bash
   ssh user@server
   docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
   
   # Check what was partially applied
   \dt  # List tables
   \d table_name  # Describe table
   ```

3. **Fix the issue**:
   - If migration was partially applied, manually complete it or revert it
   - Check migration file: `migrations/000014_*.up.sql`

4. **Force version**:
   ```bash
   # If migration was completed manually
   make migrate-force VERSION=000014
   
   # If migration was reverted manually
   make migrate-force VERSION=000013
   ```

5. **Verify**:
   ```bash
   make migrate-status
   # Should show: Version: 000014 (clean)
   ```

### Problem: Migration fails during deployment

**Symptom**:
```
✗ Migration failed!
Restoring database from backup...
```

**Cause**: SQL error in migration file, or database constraint violation.

**Solution**:

1. **Check logs**:
   ```bash
   # Deployment script shows error output
   # Look for SQL error messages
   ```

2. **Review migration file**:
   ```bash
   cat migrations/000014_*.up.sql
   # Check for syntax errors, missing columns, etc.
   ```

3. **Fix migration file**:
   - Edit the `.up.sql` file
   - Test locally first (if possible)

4. **Sync and retry**:
   ```bash
   make migrate-sync
   make release VERSION=v1.2.0
   ```

### Problem: "Cannot connect to database"

**Symptom**:
```
✗ Cannot connect to database
```

**Cause**: PostgreSQL container not running, or network issue.

**Solution**:

1. **Check PostgreSQL**:
   ```bash
   ssh user@server
   docker ps | grep postgres
   ```

2. **Check network**:
   ```bash
   docker network ls | grep llm-proxy-network
   ```

3. **Check credentials**:
   ```bash
   cat /opt/llm-proxy/deployments/docker/.env | grep DB_
   ```

4. **Test connection**:
   ```bash
   docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT 1;"
   ```

### Problem: Migration applied but version not updated

**Symptom**: Migration ran successfully, but `migrate-status` shows old version.

**Cause**: `schema_migrations` table not updated (rare).

**Solution**:

1. **Check table**:
   ```bash
   ssh user@server
   docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
   
   SELECT * FROM schema_migrations;
   ```

2. **Manually update**:
   ```sql
   UPDATE schema_migrations SET version = 000014, dirty = false;
   ```

3. **Verify**:
   ```bash
   make migrate-status
   ```

---

## Recovery Procedures

### Scenario 1: Deployment Failed, Need to Rollback

**Situation**: Deployment completed, but application is broken. Need to rollback.

**Steps**:

1. **Identify backup**:
   ```bash
   ssh user@server
   ls -lh /opt/llm-proxy-backups/$(date +%Y%m%d)/
   # Find the backup created before deployment
   ```

2. **Rollback database**:
   ```bash
   # Stop backend to prevent writes
   docker stop llm-proxy-backend
   
   # Restore backup
   docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < \
     /opt/llm-proxy-backups/20260320/postgres-backup-143000.sql
   
   # Verify
   docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
     "SELECT version FROM schema_migrations;"
   ```

3. **Rollback containers**:
   ```bash
   cd /opt/llm-proxy/deployments/docker
   
   # Edit docker-compose.registry-deploy.yml
   # Change image tags back to previous version
   
   docker compose -f docker-compose.registry-deploy.yml up -d
   ```

4. **Verify**:
   ```bash
   docker ps
   curl http://localhost:8080/health
   ```

### Scenario 2: Migration Stuck in Dirty State

**Situation**: Migration failed, database is dirty, can't apply new migrations.

**Steps**:

1. **Assess damage**:
   ```bash
   make migrate-status
   # Shows: Version: 000014 (dirty)
   
   ssh user@server
   docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
   
   # Check what was applied
   \dt  # List tables
   ```

2. **Option A: Complete the migration manually**:
   ```bash
   # If migration was partially applied, complete it
   cat migrations/000014_*.up.sql
   # Copy remaining SQL statements
   
   docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
   # Paste and execute remaining statements
   
   # Force version to clean
   ./scripts/deployment/migrate.sh force 000014
   ```

3. **Option B: Revert the migration manually**:
   ```bash
   # If migration should be reverted
   cat migrations/000014_*.down.sql
   # Copy SQL statements
   
   docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
   # Paste and execute statements
   
   # Force version to previous
   ./scripts/deployment/migrate.sh force 000013
   ```

4. **Verify**:
   ```bash
   make migrate-status
   # Should show: Version: 000014 (clean) or 000013 (clean)
   ```

### Scenario 3: Complete Database Corruption

**Situation**: Database is completely broken, need to restore from backup.

**Steps**:

1. **Stop all services**:
   ```bash
   ssh user@server
   cd /opt/llm-proxy/deployments/docker
   docker compose -f docker-compose.registry-deploy.yml down
   ```

2. **Backup current state** (for forensics):
   ```bash
   docker start llm-proxy-postgres
   docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > \
     /opt/llm-proxy-backups/corrupted-$(date +%Y%m%d-%H%M%S).sql
   docker stop llm-proxy-postgres
   ```

3. **Drop and recreate database**:
   ```bash
   docker start llm-proxy-postgres
   docker exec -it llm-proxy-postgres psql -U proxy_user -d postgres
   
   DROP DATABASE llm_proxy;
   CREATE DATABASE llm_proxy OWNER proxy_user;
   \q
   ```

4. **Restore from backup**:
   ```bash
   # Find latest good backup
   ls -lh /opt/llm-proxy-backups/*/
   
   # Restore
   docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < \
     /opt/llm-proxy-backups/20260320/postgres-backup-120000.sql
   ```

5. **Verify**:
   ```bash
   docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
     "SELECT version FROM schema_migrations;"
   
   docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \
     "SELECT COUNT(*) FROM users;"
   ```

6. **Restart services**:
   ```bash
   cd /opt/llm-proxy/deployments/docker
   docker compose -f docker-compose.registry-deploy.yml up -d
   ```

---

## Creating New Migrations

### Step 1: Create Migration Files

```bash
# Create new migration
make migrate-create NAME=add_user_roles

# This creates:
# migrations/000014_add_user_roles.up.sql
# migrations/000014_add_user_roles.down.sql
```

### Step 2: Write SQL

**migrations/000014_add_user_roles.up.sql**:
```sql
-- Add roles table
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Add role_id to users
ALTER TABLE users ADD COLUMN role_id UUID REFERENCES roles(id);

-- Insert default roles
INSERT INTO roles (name, description) VALUES
    ('admin', 'Administrator with full access'),
    ('user', 'Regular user with limited access');
```

**migrations/000014_add_user_roles.down.sql**:
```sql
-- Remove role_id from users
ALTER TABLE users DROP COLUMN role_id;

-- Drop roles table
DROP TABLE roles;
```

### Step 3: Test Locally (Optional)

```bash
# If you have local PostgreSQL
docker run --rm \
  -v $(pwd)/migrations:/migrations \
  migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database "postgres://user:pass@localhost:5432/llm_proxy_test?sslmode=disable" \
  up

# Verify
psql -U user -d llm_proxy_test -c "\dt"

# Rollback test
docker run --rm \
  -v $(pwd)/migrations:/migrations \
  migrate/migrate:v4.17.0 \
  -path=/migrations \
  -database "postgres://user:pass@localhost:5432/llm_proxy_test?sslmode=disable" \
  down 1
```

### Step 4: Sync to Server

```bash
# Sync migrations to production server
make migrate-sync
```

### Step 5: Apply Migration

```bash
# Option A: Apply manually
make migrate-up

# Option B: Apply during next deployment
make release VERSION=v1.2.0
```

### Best Practices

✅ **Always write `.down.sql`** - Every migration must be reversible  
✅ **Test locally first** - If possible, test on local database  
✅ **Small migrations** - One logical change per migration  
✅ **Idempotent when possible** - Use `IF NOT EXISTS`, `IF EXISTS`  
✅ **No data loss** - Avoid `DROP COLUMN` without backup  
✅ **Add indexes separately** - Large indexes can timeout  
✅ **Use transactions** - Wrap in `BEGIN; ... COMMIT;` when possible  

### Example: Safe Column Removal

**Bad**:
```sql
-- migrations/000015_remove_old_column.up.sql
ALTER TABLE users DROP COLUMN old_field;
```

**Good**:
```sql
-- migrations/000015_remove_old_column.up.sql
-- Step 1: Make column nullable (if not already)
ALTER TABLE users ALTER COLUMN old_field DROP NOT NULL;

-- Step 2: Add comment for future removal
COMMENT ON COLUMN users.old_field IS 'DEPRECATED: Remove after 2026-04-01';

-- migrations/000016_actually_remove_old_column.up.sql (later)
-- After verifying no code uses this column
ALTER TABLE users DROP COLUMN old_field;
```

---

## Related Documentation

- [Registry Deployment Guide](REGISTRY_DEPLOYMENT.md) - Full deployment process
- [Deployment Flow](DEPLOYMENT_FLOW.md) - Phase-by-phase explanation
- [Cheatsheet](CHEATSHEET.md) - Quick command reference
- [Server Cleanup](SERVER_CLEANUP.md) - Maintenance procedures

---

## Summary

✅ **Migrations run automatically** before container deployment  
✅ **Auto-rollback** if migrations fail  
✅ **Manual control** via `make migrate-*` commands  
✅ **Version tracking** via `schema_migrations` table  
✅ **Automatic backups** before migrations  
✅ **Dirty state detection** and recovery procedures  

**No more schema mismatch outages!** 🎉
