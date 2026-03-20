# 🧹 Server Cleanup Guide

## Overview

This guide explains how to clean up unnecessary files from your production server, keeping only essential configuration files.

---

## Why Clean Up?

### Problems with Source Code on Server

❌ **Security Risk:** Source code exposes application logic  
❌ **Attack Surface:** More files = more potential vulnerabilities  
❌ **Confusion:** Outdated code can cause debugging issues  
❌ **Disk Space:** Unnecessary files waste storage  
❌ **Wrong Workflow:** Editing code on server is bad practice

### Benefits of Minimal Server

✅ **Security:** Only configs, no source code  
✅ **Clarity:** Clear what's needed  
✅ **Speed:** Faster deployments  
✅ **Best Practice:** Production = runtime only  
✅ **Easy Recovery:** Simple backup/restore

---

## Before & After

### ❌ BEFORE (Messy Server - ~150+ files)

```
/opt/llm-proxy/
├── .git/                     ← 500+ files!
├── .gitignore
├── .gitlab-ci.yml
├── README.md
├── Makefile
├── go.mod, go.sum
├── Dockerfile*
│
├── internal/                 ← Go source code
│   ├── application/
│   ├── domain/
│   ├── infrastructure/
│   └── interfaces/
│
├── cmd/                      ← More Go code
├── pkg/                      ← More Go code
├── tests/                    ← Test code
│
├── admin-ui/                 ← Frontend source
│   ├── src/                  ← 50+ Svelte files
│   ├── node_modules/         ← 15,000+ files!
│   ├── package.json
│   └── dist/                 ← Build output
│
├── landing/                  ← Landing page source
│   ├── *.html
│   ├── *.png
│   └── Dockerfile
│
├── docs/                     ← Documentation
├── scripts/                  ← Build scripts
├── migrations/               ← SQL migrations
└── bin/                      ← Compiled binaries
```

**Total:** ~16,000+ files, ~500 MB

---

### ✅ AFTER (Clean Server - ~20 files)

```
/opt/llm-proxy/
├── deployments/
│   └── docker/
│       ├── .env                               ← Config + Secrets
│       ├── docker-compose.registry-deploy.yml ← Container definitions
│       ├── prometheus.yml                     ← Metrics
│       └── grafana/
│           ├── provisioning/
│           └── dashboards/
│
└── scripts/
    └── deployment/
        └── server-deploy.sh                   ← Deployment automation

/etc/caddy/
└── Caddyfile                                  ← Reverse proxy

Docker Volumes:
├── llm-proxy-postgres-data/                   ← Database (persisted)
├── llm-proxy-redis-data/                      ← Cache (persisted)
├── llm-proxy-prometheus-data/                 ← Metrics (persisted)
└── llm-proxy-grafana-data/                    ← Dashboards (persisted)
```

**Total:** ~20 files, ~2 MB

**Savings:** 15,980+ files removed, 498 MB saved!

---

## Cleanup Process

### Step 1: Pre-Cleanup Checklist

```bash
# Verify you have access
ssh openweb "whoami"

# Check current disk usage
ssh openweb "du -sh /opt/llm-proxy"

# List what will be removed (dry-run)
ssh openweb "cd /opt/llm-proxy && ls -la"
```

### Step 2: Backup Current State

```bash
# Create full backup before cleanup
ssh openweb "cd /opt && tar -czf /tmp/llm-proxy-backup-$(date +%Y%m%d).tar.gz llm-proxy/"
scp openweb:/tmp/llm-proxy-backup-*.tar.gz ./backups/
```

### Step 3: Run Cleanup Script

```bash
# Run automated cleanup
cd /home/krieger/Sites/golang-projekte/llm-proxy
./scripts/deployment/cleanup-server.sh

# Follow prompts:
# - Enter server hostname: openweb
# - Confirm: yes
```

### Step 4: Verify Cleanup

```bash
# Check remaining files
ssh openweb "cd /opt/llm-proxy && find . -type f | sort"

# Should show ONLY:
# ./deployments/docker/.env
# ./deployments/docker/docker-compose.registry-deploy.yml
# ./deployments/docker/prometheus.yml
# ./deployments/docker/grafana/...
# ./scripts/deployment/server-deploy.sh
```

### Step 5: Test Deployment

```bash
# Test that deployment still works
make deploy-prod VERSION=latest

# Verify containers are healthy
make status
```

---

## Manual Cleanup (Alternative)

If you prefer to clean up manually:

```bash
ssh openweb

cd /opt/llm-proxy

# Remove source code
rm -rf internal/ cmd/ pkg/ api/ tests/
rm -f go.mod go.sum *.go

# Remove admin-ui source
rm -rf admin-ui/src/ admin-ui/node_modules/ admin-ui/dist/
rm -f admin-ui/package*.json admin-ui/*.config.js
rm -f admin-ui/Dockerfile*

# Remove landing page
rm -rf landing/

# Remove Git
rm -rf .git/ .gitignore .gitlab*

# Remove docs
rm -rf docs/ README*.md

# Remove build tools
rm -f Makefile* Dockerfile*

# Remove old scripts
rm -rf scripts/setup/ scripts/maintenance/
rm -f scripts/deployment/build-and-push.sh

# Keep ONLY:
# - deployments/docker/.env
# - deployments/docker/docker-compose.registry-deploy.yml
# - deployments/docker/prometheus.yml
# - deployments/docker/grafana/
# - scripts/deployment/server-deploy.sh
```

---

## What Gets Removed

### Source Code (All!)

```
❌ internal/        # Backend Go code
❌ cmd/             # Main packages
❌ pkg/             # Shared packages
❌ api/             # API definitions
❌ tests/           # Test code
❌ *.go             # All Go files
❌ go.mod, go.sum   # Go dependencies
```

### Frontend Source (All!)

```
❌ admin-ui/src/          # Svelte source
❌ admin-ui/node_modules/ # 15,000+ files!
❌ admin-ui/dist/         # Build output
❌ admin-ui/package.json  # NPM config
❌ admin-ui/*.config.js   # Build configs
❌ landing/*.html         # Landing HTML
❌ landing/*.png          # Images
```

### Build Tools

```
❌ Dockerfile*           # Build instructions
❌ Makefile*             # Build automation
❌ .dockerignore         # Docker ignore
❌ deploy.sh             # Old scripts
```

### Git & Docs

```
❌ .git/                 # Git repository
❌ .gitignore            # Git ignore
❌ .gitlab-ci.yml        # CI/CD
❌ docs/                 # Documentation
❌ README*.md            # Readme files
```

### Misc Files

```
❌ migrations/           # SQL migrations (in container!)
❌ bin/                  # Compiled binaries
❌ logs/                 # Logs (in containers!)
❌ example-*.csv         # Examples
❌ test-*.json           # Test data
```

---

## What Gets Kept

### Essential Files ONLY

```
✅ deployments/docker/.env                          # Configuration + Secrets
✅ deployments/docker/docker-compose.registry-deploy.yml  # Container definitions
✅ deployments/docker/prometheus.yml                # Metrics config
✅ deployments/docker/grafana/                      # Grafana dashboards
✅ scripts/deployment/server-deploy.sh              # Deployment script
✅ /etc/caddy/Caddyfile                             # Reverse proxy config
```

### Docker Volumes (Never Touched!)

```
✅ llm-proxy-postgres-data/    # Database (persisted)
✅ llm-proxy-redis-data/       # Cache (persisted)
✅ llm-proxy-prometheus-data/  # Metrics (persisted)
✅ llm-proxy-grafana-data/     # Dashboards (persisted)
```

---

## Safety Features

### Automatic Backup

The cleanup script automatically creates a backup before removing anything:

```bash
./backups/server-backup-20260320-120530.tar.gz
```

### Restore from Backup

If something goes wrong:

```bash
# Copy backup to server
scp ./backups/server-backup-20260320-120530.tar.gz openweb:/tmp/

# Restore
ssh openweb "cd /opt && tar -xzf /tmp/server-backup-20260320-120530.tar.gz"
```

### Docker Volumes Protected

The cleanup script **NEVER** touches Docker volumes. Your data is safe:

- PostgreSQL database remains intact
- Redis cache remains intact
- Prometheus metrics remain intact
- Grafana dashboards remain intact

---

## Verification

### After Cleanup, Verify:

```bash
# 1. Check file count
ssh openweb "find /opt/llm-proxy -type f | wc -l"
# Should be: ~20 files

# 2. Check disk usage
ssh openweb "du -sh /opt/llm-proxy"
# Should be: ~2 MB

# 3. Check containers still work
make status
# All should be: healthy

# 4. Check deployments still work
make deploy-prod VERSION=latest
# Should succeed

# 5. Check application works
curl https://scrubgate.tech/health
# Should return: OK
```

---

## Troubleshooting

### Deployment fails after cleanup

```bash
# Check if essential files exist
ssh openweb "ls -la /opt/llm-proxy/deployments/docker/.env"
ssh openweb "ls -la /opt/llm-proxy/deployments/docker/docker-compose.registry-deploy.yml"
ssh openweb "ls -la /opt/llm-proxy/scripts/deployment/server-deploy.sh"

# If missing, restore from backup
scp ./backups/server-backup-*.tar.gz openweb:/tmp/
ssh openweb "cd /opt && tar -xzf /tmp/server-backup-*.tar.gz"
```

### Container health check fails

```bash
# Check logs
make logs SERVICE=backend

# Verify .env file is correct
ssh openweb "cat /opt/llm-proxy/deployments/docker/.env"

# Restart containers
ssh openweb "cd /opt/llm-proxy/deployments/docker && docker-compose -f docker-compose.registry-deploy.yml restart"
```

### Accidentally deleted .env

```bash
# Restore from backup
scp ./backups/server-backup-*.tar.gz openweb:/tmp/
ssh openweb "cd /tmp && tar -xzf server-backup-*.tar.gz"
ssh openweb "cp /tmp/llm-proxy/deployments/docker/.env /opt/llm-proxy/deployments/docker/"
```

---

## Best Practices

### ✅ DO

- ✅ Always backup before cleanup
- ✅ Verify cleanup with dry-run first
- ✅ Test deployment after cleanup
- ✅ Keep backups for 30 days
- ✅ Document any manual changes

### ❌ DON'T

- ❌ Delete Docker volumes (data loss!)
- ❌ Delete .env file (secrets!)
- ❌ Delete without backup
- ❌ Run cleanup on production without testing
- ❌ Edit files directly on server after cleanup

---

## Regular Maintenance

### Monthly Cleanup Tasks

```bash
# 1. Remove old Docker images
ssh openweb "docker image prune -a -f"

# 2. Remove old backups (keep last 30 days)
find ./backups/ -name "*.tar.gz" -mtime +30 -delete

# 3. Check disk usage
ssh openweb "df -h"

# 4. Verify .env is current
ssh openweb "md5sum /opt/llm-proxy/deployments/docker/.env"
```

---

## Summary

### Before Cleanup
- ❌ 16,000+ files
- ❌ 500 MB disk space
- ❌ Source code on server
- ❌ Security risk
- ❌ Complex structure

### After Cleanup
- ✅ ~20 files
- ✅ ~2 MB disk space
- ✅ No source code
- ✅ Minimal attack surface
- ✅ Simple structure

### Deployment Still Works!
- ✅ Pull images from registries
- ✅ Deploy with zero downtime
- ✅ Automatic backups
- ✅ Easy rollbacks
- ✅ Production-ready

---

**Last Updated:** 2026-03-20  
**Version:** 1.0
