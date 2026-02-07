# Auto-Deployment System - Complete Guide

## 📋 Table of Contents

1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Installation](#installation)
4. [How It Works](#how-it-works)
5. [Usage Guide](#usage-guide)
6. [Deployment Scripts](#deployment-scripts)
7. [Git Hooks](#git-hooks)
8. [Makefile Targets](#makefile-targets)
9. [Safety Features](#safety-features)
10. [Troubleshooting](#troubleshooting)
11. [Advanced Usage](#advanced-usage)

---

## Overview

The **Auto-Deployment System** for LLM-Proxy automatically detects and deploys database migrations and content filters to production whenever you commit changes to the repository.

### Key Features

✅ **Automatic Detection** - Detects migration and filter changes in Git commits  
✅ **Safe Deployment** - Automatic database backups before every change  
✅ **Interactive** - Prompts for confirmation before deploying to production  
✅ **Comprehensive** - Supports migrations, filters, and full deployments  
✅ **Verified** - Checks deployment success and verifies data integrity  
✅ **Reversible** - Provides rollback instructions on failure  

### What Gets Auto-Deployed?

| File Pattern | Description | Auto-Deploy |
|--------------|-------------|-------------|
| `migrations/*.up.sql` | Database schema migrations | ✅ Yes |
| `migrations/*.down.sql` | Migration rollbacks | ✅ Yes |
| `migrations/filters/*.sql` | Content filter SQL imports | ✅ Yes |
| `migrations/filters/*.csv` | Content filter CSV imports | ✅ Yes |

---

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Developer Workflow                                          │
└─────────────────────────────────────────────────────────────┘
         │
         │ 1. Create migration file
         │    migrations/000008_add_column.up.sql
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│  git commit -m "Add new column"                              │
└─────────────────────────────────────────────────────────────┘
         │
         │ 2. Git post-commit hook triggers
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│  .git/hooks/post-commit                                      │
│  • Detects migration files in commit                         │
│  • Shows list of detected changes                            │
│  • Calls auto-deploy-migrations.sh                           │
└─────────────────────────────────────────────────────────────┘
         │
         │ 3. Prompts user
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│  Deploy to production? [y/N]                                 │
│  > y                                                          │
└─────────────────────────────────────────────────────────────┘
         │
         │ 4. If confirmed
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│  deploy-migrations.sh                                        │
│  • SSH to production server (openweb)                        │
│  • Backup database                                            │
│  • Copy migration files                                       │
│  • Run migrations                                             │
│  • Restart backend                                            │
│  • Verify deployment                                          │
└─────────────────────────────────────────────────────────────┘
         │
         │ 5. Success
         │
         ▼
┌─────────────────────────────────────────────────────────────┐
│  ✓ Migration 000008 deployed successfully                    │
│  ✓ Backend restarted                                         │
│  ✓ Health check passed                                       │
└─────────────────────────────────────────────────────────────┘
```

---

## Installation

### Step 1: Install Git Hooks

Run the setup script to install the Git hooks:

```bash
./scripts/setup-git-hooks.sh
```

Or use the Makefile target:

```bash
make setup-hooks
```

### Step 2: Verify Installation

Check that the post-commit hook is installed:

```bash
ls -la .git/hooks/post-commit
```

Expected output:
```
-rwxr-xr-x 1 user user 2352 Feb  7 15:28 .git/hooks/post-commit
```

### Step 3: Test the System

Create a test migration and commit it:

```bash
# Create a test migration
echo "-- Test migration" > migrations/000999_test.up.sql

# Commit it
git add migrations/000999_test.up.sql
git commit -m "test: Test auto-deployment system"

# The hook should trigger and prompt you
# Answer 'N' to skip actual deployment during testing
```

---

## How It Works

### Workflow Overview

1. **Developer creates migration** → Writes SQL file in `migrations/` directory
2. **Developer commits changes** → Runs `git commit`
3. **Git hook triggers** → `.git/hooks/post-commit` executes automatically
4. **Changes detected** → Hook scans commit for migration/filter files
5. **User prompted** → "Deploy to production? [y/N]"
6. **If confirmed** → Deployment script executes
7. **Database backed up** → Full backup created before changes
8. **Changes deployed** → Migrations run, filters imported
9. **Backend restarted** → Services restarted to apply changes
10. **Verification** → Health checks confirm successful deployment

### What Happens When You Skip?

If you answer **'N'** to the deployment prompt:

- Changes are committed to Git normally
- No production deployment occurs
- You can deploy later manually using `make deploy-migrations`
- The hook will continue to work for future commits

### What Happens on Failure?

If deployment fails:

1. **Backup is preserved** → Located in `/opt/llm-proxy/backups/`
2. **Error message shown** → Detailed error information
3. **Rollback instructions** → How to restore from backup
4. **Git commit remains** → Your code is safe in Git
5. **Next steps provided** → How to fix and retry

---

## Usage Guide

### Basic Usage - Auto-Deploy After Commit

```bash
# 1. Create a migration
cat > migrations/000008_add_feature.up.sql << 'EOF'
ALTER TABLE api_keys ADD COLUMN last_used_at TIMESTAMP;
EOF

# 2. Commit the changes
git add migrations/000008_add_feature.up.sql
git commit -m "feat: Add last_used_at column to api_keys"

# 3. Hook triggers automatically
# You'll see:
#
# ═══════════════════════════════════════════════════════════
#   DATABASE CHANGES DETECTED IN COMMIT
# ═══════════════════════════════════════════════════════════
#
# Migration files:
#   • migrations/000008_add_feature.up.sql
#
# Deploy to production? [y/N]: 

# 4. Type 'y' and press Enter to deploy
```

### Manual Deployment (Without Hook)

If you want to deploy manually without using the auto-deploy hook:

```bash
# Deploy pending migrations
make deploy-migrations

# Deploy content filters
make deploy-filters

# Full deployment (code + migrations + filters)
make deploy-full
```

### Check What Would Be Deployed

Run the auto-deploy script manually to see what would be deployed:

```bash
make deploy-auto
```

Or directly:

```bash
./scripts/deployment/auto-deploy-migrations.sh
```

---

## Deployment Scripts

### 1. `deploy-migrations.sh`

**Purpose:** Deploy database migrations to production server

**Usage:**
```bash
./scripts/deployment/deploy-migrations.sh [OPTIONS] [MIGRATION_NUMBER]

# Deploy all pending migrations
./scripts/deployment/deploy-migrations.sh

# Deploy specific migration
./scripts/deployment/deploy-migrations.sh 000008

# Skip backup (not recommended)
./scripts/deployment/deploy-migrations.sh --skip-backup

# Skip backend restart
./scripts/deployment/deploy-migrations.sh --skip-restart
```

**What it does:**
1. Connects to production server via SSH
2. Creates full database backup
3. Copies migration files to server
4. Runs migrations using golang-migrate
5. Restarts backend container
6. Verifies deployment success
7. Shows migration status

**Options:**
- `--skip-backup` - Skip database backup (dangerous!)
- `--skip-restart` - Don't restart backend after migration
- `--help` - Show help message

---

### 2. `deploy-filters.sh`

**Purpose:** Deploy content filters to production database

**Usage:**
```bash
./scripts/deployment/deploy-filters.sh [OPTIONS]

# Deploy filters
./scripts/deployment/deploy-filters.sh

# Dry-run (test without changes)
./scripts/deployment/deploy-filters.sh --dry-run

# Show help
./scripts/deployment/deploy-filters.sh --help
```

**What it does:**
1. Backs up existing filters
2. Copies filter SQL file to server
3. Imports filters into database
4. Verifies filter count
5. Shows filter statistics by type
6. Tests critical filters

**Options:**
- `--dry-run` - Test deployment without making changes
- `--help` - Show help message

---

### 3. `deploy-full.sh`

**Purpose:** Full production deployment (code, migrations, filters, restart)

**Usage:**
```bash
./scripts/deployment/deploy-full.sh [OPTIONS]

# Full deployment
./scripts/deployment/deploy-full.sh

# Skip Docker build
./scripts/deployment/deploy-full.sh --skip-build

# Skip migrations
./scripts/deployment/deploy-full.sh --skip-migrations

# Skip filters
./scripts/deployment/deploy-full.sh --skip-filters
```

**What it does:**
1. Checks Git status (warns if uncommitted changes)
2. Syncs code to production via rsync
3. Rebuilds Docker containers
4. Runs database migrations
5. Deploys content filters
6. Restarts all services
7. Runs health checks
8. Shows deployment summary

**Options:**
- `--skip-build` - Don't rebuild Docker images
- `--skip-migrations` - Skip database migrations
- `--skip-filters` - Skip filter deployment
- `--help` - Show help message

---

### 4. `auto-deploy-migrations.sh`

**Purpose:** Detect and prompt for deployment of uncommitted migrations

**Usage:**
```bash
./scripts/deployment/auto-deploy-migrations.sh

# Or via Makefile
make deploy-auto
```

**What it does:**
1. Scans for new migration files in last commit
2. Scans for filter changes
3. Shows list of detected changes
4. Prompts: "Deploy to production? [y/N]"
5. If 'yes': Calls appropriate deployment script
6. If 'no': Exits gracefully

**When it runs:**
- Automatically after every `git commit` (via post-commit hook)
- Manually via `make deploy-auto`
- Manually via direct script execution

---

## Git Hooks

### Post-Commit Hook

**Location:** `.git/hooks/post-commit`

**Purpose:** Automatically detect database changes in commits and prompt for deployment

**Trigger Patterns:**

| Pattern | Triggers Hook |
|---------|---------------|
| `migrations/*.up.sql` | ✅ Yes |
| `migrations/*.down.sql` | ✅ Yes |
| `migrations/filters/*.sql` | ✅ Yes |
| `migrations/filters/*.csv` | ✅ Yes |
| `internal/**/*.go` | ❌ No |
| `configs/*.yaml` | ❌ No |
| `README.md` | ❌ No |

**Example Output:**

```
═══════════════════════════════════════════════════════════
  DATABASE CHANGES DETECTED IN COMMIT
═══════════════════════════════════════════════════════════

Migration files:
  • migrations/000008_add_last_used.up.sql

Filter files:
  • migrations/filters/custom_filters.sql

Deploy to production? [y/N]:
```

**Disable Temporarily:**

To skip auto-deployment for one commit:
```bash
# Just answer 'N' when prompted
git commit -m "feat: Add migration"
# > Deploy to production? [y/N]: N
```

**Disable Permanently:**

To completely disable the hook:
```bash
rm .git/hooks/post-commit
```

To re-enable later:
```bash
./scripts/setup-git-hooks.sh
```

---

## Makefile Targets

### Deployment Targets

```bash
make deploy-migrations   # Deploy database migrations to production
make deploy-filters      # Deploy content filters to production  
make deploy-full         # Full deployment (code + DB + filters)
make deploy-auto         # Check for pending deployments
make setup-hooks         # Install Git hooks for auto-deployment
```

### Usage Examples

```bash
# Install Git hooks (one-time setup)
make setup-hooks

# Deploy only migrations
make deploy-migrations

# Deploy only filters
make deploy-filters

# Full deployment (everything)
make deploy-full

# Check what would be deployed
make deploy-auto
```

---

## Safety Features

### 1. Automatic Backups

**Before every deployment:**
- Full PostgreSQL database dump created
- Backup stored in `/opt/llm-proxy/backups/`
- Timestamp-based filenames (e.g., `backup_20260207_153045.sql`)
- Retention: Backups kept indefinitely (manual cleanup required)

**Restore from backup:**
```bash
# On production server
ssh openweb
cd /opt/llm-proxy/backups
psql -U proxy_user -d llm_proxy < backup_20260207_153045.sql
```

### 2. Verification Checks

**Post-deployment verification:**
- ✅ Migration version check
- ✅ Table row counts
- ✅ Backend health check
- ✅ Container status check
- ✅ Filter count verification

### 3. Rollback Instructions

**If migration fails:**
```bash
# 1. Restore from backup
ssh openweb
cd /opt/llm-proxy/backups
psql -U proxy_user -d llm_proxy < backup_TIMESTAMP.sql

# 2. Restart backend
docker restart llm-proxy-backend

# 3. Verify restoration
docker exec llm-proxy-backend /app/llm-proxy --version
```

### 4. Confirmation Prompts

**Interactive prompts before:**
- Production deployments
- Database backups
- Backend restarts
- Destructive operations

### 5. Dry-Run Mode

**Test deployments without changes:**
```bash
./scripts/deployment/deploy-filters.sh --dry-run
```

Shows what would happen without actually modifying anything.

---

## Troubleshooting

### Problem: Hook doesn't trigger after commit

**Symptoms:**
- No prompt after `git commit`
- Hook seems to be ignored

**Solutions:**
```bash
# 1. Check if hook exists
ls -la .git/hooks/post-commit

# 2. Check if hook is executable
chmod +x .git/hooks/post-commit

# 3. Reinstall hooks
./scripts/setup-git-hooks.sh

# 4. Test manually
./.git/hooks/post-commit
```

---

### Problem: "Permission denied" when connecting to server

**Symptoms:**
- SSH connection fails
- "Permission denied (publickey)" error

**Solutions:**
```bash
# 1. Check SSH key
ssh -T openweb

# 2. Add SSH key to ssh-agent
ssh-add ~/.ssh/id_rsa

# 3. Test connection
ssh openweb "echo 'Connection successful'"

# 4. Check SSH config
cat ~/.ssh/config | grep openweb
```

---

### Problem: Migration fails on production

**Symptoms:**
- Migration script reports error
- Database changes not applied

**Solutions:**
```bash
# 1. Check error message in output
# Look for SQL error details

# 2. Restore from backup
ssh openweb
cd /opt/llm-proxy/backups
ls -lth | head -5  # Find latest backup
psql -U proxy_user -d llm_proxy < backup_TIMESTAMP.sql

# 3. Fix migration SQL locally
# Edit migrations/XXXXXX_name.up.sql

# 4. Test locally first
make migrate-up

# 5. Retry deployment
make deploy-migrations
```

---

### Problem: Backend doesn't restart after migration

**Symptoms:**
- Migration succeeds but backend not restarted
- Old code still running

**Solutions:**
```bash
# 1. Manually restart backend
ssh openweb "docker restart llm-proxy-backend"

# 2. Check container status
ssh openweb "docker ps | grep llm-proxy"

# 3. Check container logs
ssh openweb "docker logs llm-proxy-backend --tail 50"

# 4. Verify migration applied
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 5;'"
```

---

### Problem: Filters not importing correctly

**Symptoms:**
- Filter deployment succeeds but filters missing
- Filter count doesn't match expected

**Solutions:**
```bash
# 1. Check filter file syntax
cat migrations/filters/enterprise_standard_filters.sql

# 2. Test filter import locally
psql -U proxy_user -d llm_proxy < migrations/filters/enterprise_standard_filters.sql

# 3. Check for SQL errors
# Look for duplicate key violations, constraint errors

# 4. Verify filter count on production
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT COUNT(*) FROM content_filters;'"

# 5. Retry with verbose output
./scripts/deployment/deploy-filters.sh
```

---

## Advanced Usage

### Custom Deployment Workflow

**Scenario:** You want to deploy only specific migrations, not all pending ones.

```bash
# Deploy specific migration number
./scripts/deployment/deploy-migrations.sh 000008

# Or deploy up to specific version
ssh openweb
cd /opt/llm-proxy
docker exec llm-proxy-backend migrate -path /app/migrations -database "$DB_URL" goto 8
```

---

### Batch Filter Updates

**Scenario:** You have 100+ new filters to deploy.

```bash
# 1. Create filters in CSV format (faster)
cat > migrations/filters/batch_update.csv << 'EOF'
name,description,type,pattern,enabled,priority,action,category
"IBAN DE","German IBAN","regex","DE\\d{2}\\s?(?:[0-9]{4}\\s?){4}[0-9]{2}",true,100,"redact","financial"
"Email","Email addresses","regex","[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",true,90,"redact","pii"
EOF

# 2. Import via COPY (faster than INSERT)
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy << 'EOSQL'
COPY content_filters (name, description, type, pattern, enabled, priority, action, category)
FROM STDIN WITH CSV HEADER;
$(cat migrations/filters/batch_update.csv)
\.
EOSQL"

# 3. Verify import
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT COUNT(*) FROM content_filters;'"
```

---

### Zero-Downtime Deployment

**Scenario:** You need to deploy without service interruption.

```bash
# 1. Run migration without restart
./scripts/deployment/deploy-migrations.sh --skip-restart

# 2. Verify migration success
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;'"

# 3. Deploy new code with rolling restart
ssh openweb "cd /opt/llm-proxy && docker compose up -d --no-deps --build backend"

# 4. Wait for health check
for i in {1..30}; do
  curl -sf http://68.183.208.213:8080/health && break
  sleep 2
done
```

---

### Scheduled Deployments

**Scenario:** Deploy migrations at specific time (e.g., 2 AM maintenance window).

```bash
# 1. Create deployment script
cat > /tmp/scheduled-deploy.sh << 'EOF'
#!/bin/bash
cd /home/krieger/Sites/golang-projekte/llm-proxy
make deploy-full > /tmp/deploy-$(date +%Y%m%d-%H%M%S).log 2>&1
EOF

chmod +x /tmp/scheduled-deploy.sh

# 2. Schedule with 'at' command
echo "/tmp/scheduled-deploy.sh" | at 02:00

# 3. Verify scheduled job
atq

# 4. Cancel if needed
atrm JOB_NUMBER
```

---

## Best Practices

### ✅ DO

- **Always test migrations locally first** using `make migrate-up`
- **Read the backup confirmation** before proceeding
- **Review migration SQL** before committing
- **Use descriptive migration names** (e.g., `000008_add_user_last_login.up.sql`)
- **Keep migrations small** - One logical change per migration
- **Write rollback migrations** (`.down.sql`) for every `.up.sql`
- **Test rollbacks** locally before deploying

### ❌ DON'T

- **Don't skip backups** unless absolutely necessary
- **Don't force-push** after auto-deployment has occurred
- **Don't edit deployed migrations** - Create new ones instead
- **Don't run migrations manually** on production without backup
- **Don't commit broken SQL** - Always test locally first
- **Don't ignore warnings** from deployment scripts

---

## Summary

The Auto-Deployment System provides:

✅ **Automatic detection** of database changes in Git commits  
✅ **Safe deployment** with backups and verification  
✅ **Interactive prompts** for production deployments  
✅ **Comprehensive scripts** for migrations, filters, and full deployments  
✅ **Git hooks** for seamless workflow integration  
✅ **Makefile targets** for convenient access  
✅ **Rollback support** for failed deployments  

**Quick Start:**
```bash
# 1. Install hooks (one-time)
make setup-hooks

# 2. Create migration
echo "ALTER TABLE api_keys ADD COLUMN notes TEXT;" > migrations/000008_add_notes.up.sql

# 3. Commit
git add migrations/000008_add_notes.up.sql
git commit -m "feat: Add notes column to api_keys"

# 4. Deploy when prompted
# > Deploy to production? [y/N]: y
```

**Need Help?**
- Check logs: `ssh openweb "docker logs llm-proxy-backend --tail 100"`
- Test connection: `ssh openweb "echo 'OK'"`
- Run verification: `make deploy-auto`
- Restore backup: See [Safety Features → Rollback Instructions](#3-rollback-instructions)

---

**Last Updated:** 2026-02-07  
**Version:** 1.0.0  
**Maintainer:** LLM-Proxy Team
