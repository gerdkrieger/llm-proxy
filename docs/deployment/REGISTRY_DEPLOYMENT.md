# 🐳 Registry-Based Deployment Strategy

## Overview

This deployment strategy uses **pre-built Docker images** from container registries. No source code is needed on the production server - only configuration files.

### Key Principles

- ✅ **Build locally** or on CI/CD server
- ✅ **Push to registries** (GitHub/GitLab Container Registries)
- ✅ **Production server** only pulls images and runs them
- ✅ **Zero source code** on production server
- ✅ **Easy rollbacks** by changing version tags
- ✅ **Consistent builds** across environments

---

## Registry Strategy

### Images → Registries

| Component | Registry | Image URL |
|-----------|----------|-----------|
| **Backend** | GitHub CR | `ghcr.io/gerdkrieger/llm-proxy-backend` |
| **Admin-UI** | GitHub CR | `ghcr.io/gerdkrieger/llm-proxy-admin-ui` |
| **Landing Page** | GitLab CR | `registry.gitlab.com/krieger-engineering/llm-proxy-landing` |

> **⚠️ IMPORTANT:** Landing Page goes **ONLY** to GitLab Registry!

---

## Quick Start

### 1. Initial Setup (One-Time)

#### A. Login to Registries (Local Machine)

```bash
# GitHub Container Registry
docker login ghcr.io -u gerdkrieger

# GitLab Container Registry
docker login registry.gitlab.com
```

#### B. Server Setup

```bash
# On production server
ssh openweb

# Login to registries on server
docker login ghcr.io -u gerdkrieger
docker login registry.gitlab.com

# Create network and volumes
docker network create llm-proxy-network
docker volume create llm-proxy-postgres-data
docker volume create llm-proxy-redis-data
docker volume create llm-proxy-prometheus-data
docker volume create llm-proxy-grafana-data

# Create necessary directories
mkdir -p /opt/llm-proxy/scripts/deployment
mkdir -p /opt/llm-proxy/deployments/docker
mkdir -p /opt/llm-proxy-backups
```

### 2. Build & Deploy

#### Using Makefile (Recommended)

```bash
# Full release workflow
make release VERSION=v1.0.0

# Or step-by-step:
make build VERSION=v1.0.0         # Build images
make push VERSION=v1.0.0          # Push to registries
make deploy-prod VERSION=v1.0.0   # Deploy to server
```

#### Manual Commands

```bash
# Build and push images
./scripts/deployment/build-and-push.sh v1.0.0

# Deploy on server
ssh openweb "cd /opt/llm-proxy && ./scripts/deployment/server-deploy.sh v1.0.0"
```

---

## Deployment Workflow

### Development → Production

```
┌─────────────────┐
│ LOCAL MACHINE   │
├─────────────────┤
│ 1. Code changes │
│ 2. Git commit   │
│ 3. Build images │
│ 4. Push to      │
│    registries   │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│ CONTAINER REGISTRIES            │
├─────────────────────────────────┤
│ ghcr.io:                        │
│  - llm-proxy-backend:v1.0.0     │
│  - llm-proxy-admin-ui:v1.0.0    │
│                                 │
│ registry.gitlab.com:            │
│  - llm-proxy-landing:v1.0.0     │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────┐
│ PROD SERVER     │
├─────────────────┤
│ 1. Pull images  │
│ 2. Backup DB    │
│ 3. Deploy       │
│ 4. Health check │
└─────────────────┘
```

---

## Common Operations

### Deploy New Version

```bash
# Build and push locally
make build-and-push VERSION=v1.2.0

# Deploy to production
make deploy-prod VERSION=v1.2.0
```

### Rollback to Previous Version

```bash
# Rollback to last working version
make rollback VERSION=v1.1.0
```

### Check Production Status

```bash
make status
```

Output:
```
NAMES                  STATUS                        PORTS
llm-proxy-backend      Up 5 minutes (healthy)        127.0.0.1:8080->8080/tcp
llm-proxy-admin-ui     Up 5 minutes (healthy)        127.0.0.1:3005->80/tcp
llm-proxy-landing      Up 5 minutes (healthy)        127.0.0.1:8090->80/tcp
...
```

### View Logs

```bash
# Backend logs
make logs SERVICE=backend

# Admin UI logs
make logs SERVICE=admin-ui

# Landing page logs
make logs SERVICE=landing
```

### Backup Database

```bash
make backup-db
```

---

## File Structure on Production Server

```
/opt/llm-proxy/
├── deployments/
│   └── docker/
│       ├── .env                               # ONLY config file!
│       ├── docker-compose.registry-deploy.yml # Deployment config
│       ├── prometheus.yml                     # Prometheus config
│       └── grafana/                           # Grafana dashboards
├── scripts/
│   └── deployment/
│       └── server-deploy.sh                   # Deployment script
└── /opt/llm-proxy-backups/                    # Database backups
    └── 20260320/
        ├── postgres-backup-120530.sql
        └── images-before-120530.txt
```

> **⚠️ IMPORTANT:** No source code, no build tools, no Git repository on server!

---

## Versioning Strategy

### Semantic Versioning

Use semantic versioning for releases:

```bash
make release VERSION=v1.0.0   # Major release
make release VERSION=v1.1.0   # Minor release  
make release VERSION=v1.1.1   # Patch release
```

### Latest Tag

Always push both version tag AND `latest` tag:

```bash
# Automatically done by build-and-push.sh
ghcr.io/gerdkrieger/llm-proxy-backend:v1.2.0
ghcr.io/gerdkrieger/llm-proxy-backend:latest
```

### Development Builds

Use branch names or commit hashes for dev builds:

```bash
make build VERSION=feature-auth
make build VERSION=$(git rev-parse --short HEAD)
```

---

## Security Considerations

### 🔒 Registry Authentication

- **GitHub:** Use Personal Access Token with `write:packages` scope
- **GitLab:** Use Deploy Token or Personal Access Token
- Store tokens securely, never commit them

### 🔒 Image Signing (Future)

```bash
# Sign images with Docker Content Trust
export DOCKER_CONTENT_TRUST=1
docker push ghcr.io/gerdkrieger/llm-proxy-backend:v1.0.0
```

### 🔒 Vulnerability Scanning

```bash
# Scan images before deployment
docker scan ghcr.io/gerdkrieger/llm-proxy-backend:v1.0.0
```

---

## Troubleshooting

### Build Fails

```bash
# Check Docker daemon
docker info

# Clean build cache
docker builder prune

# Rebuild from scratch
docker build --no-cache -f deployments/docker/Dockerfile .
```

### Push to Registry Fails

```bash
# Re-login
docker logout ghcr.io
docker login ghcr.io -u gerdkrieger

# Check network
curl -I https://ghcr.io
```

### Deployment Fails

```bash
# Check server logs
ssh openweb "docker-compose -f /opt/llm-proxy/deployments/docker/docker-compose.registry-deploy.yml logs"

# Check disk space
ssh openweb "df -h"

# Check network connectivity from server
ssh openweb "docker pull ghcr.io/gerdkrieger/llm-proxy-backend:latest"
```

### Container Unhealthy After Deployment

```bash
# Check logs
make logs SERVICE=backend

# Check health endpoint
ssh openweb "curl http://localhost:8080/health"

# Rollback
make rollback VERSION=v1.0.0
```

---

## Emergency Procedures

### Complete Rollback

```bash
# 1. Stop all containers
ssh openweb "cd /opt/llm-proxy && docker-compose -f deployments/docker/docker-compose.registry-deploy.yml down"

# 2. Restore database backup
ssh openweb "cat /opt/llm-proxy-backups/20260320/postgres-backup-120530.sql | \
    docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy"

# 3. Deploy last known good version
make deploy-prod VERSION=v1.0.0
```

### Manual Image Pull

```bash
ssh openweb "docker pull ghcr.io/gerdkrieger/llm-proxy-backend:v1.0.0"
ssh openweb "docker pull ghcr.io/gerdkrieger/llm-proxy-admin-ui:v1.0.0"
ssh openweb "docker pull registry.gitlab.com/krieger-engineering/llm-proxy-landing:v1.0.0"
```

---

## Best Practices

### ✅ DO

- ✅ Use semantic versioning
- ✅ Test locally before deploying
- ✅ Always backup database before deployment
- ✅ Tag images with both version AND latest
- ✅ Keep deployment scripts in version control
- ✅ Monitor container health after deployment
- ✅ Keep last 3-5 versions in registry for rollback

### ❌ DON'T

- ❌ Deploy `latest` tag to production (use specific versions)
- ❌ Skip backups before deployment
- ❌ Deploy without testing
- ❌ Commit `.env` files to Git
- ❌ Store source code on production server
- ❌ Use `--force` flags in production

---

## Makefile Commands Reference

```bash
make help              # Show all available commands
make login-registries  # Login to GitHub + GitLab registries
make build             # Build all images
make push              # Push images to registries
make build-and-push    # Build AND push (shortcut)
make deploy-prod       # Deploy to production server
make status            # Check container status on server
make logs SERVICE=X    # View logs from specific service
make rollback VERSION=X # Rollback to previous version
make backup-db         # Backup PostgreSQL database
make release VERSION=X # Full release workflow
make clean             # Clean up old images
```

---

## Migration from Old Deployment

### Old Method (Source Code Deployment)
```bash
# ❌ OLD WAY
ssh openweb "cd /opt/llm-proxy && git pull && docker-compose build && docker-compose up -d"
```

### New Method (Registry Deployment)
```bash
# ✅ NEW WAY
make release VERSION=v1.0.0
```

### Migration Steps

1. **One-time setup on server:**
   ```bash
   ssh openweb
   docker login ghcr.io -u gerdkrieger
   docker login registry.gitlab.com
   ```

2. **Remove source code from server:**
   ```bash
   ssh openweb "cd /opt/llm-proxy && rm -rf .git internal cmd pkg admin-ui/src landing/src"
   ```

3. **Deploy using registries:**
   ```bash
   make deploy-prod VERSION=latest
   ```

---

## Support & Troubleshooting

For issues or questions:
- Check logs: `make logs SERVICE=<name>`
- Check status: `make status`
- Rollback: `make rollback VERSION=<previous>`
- Backup: `make backup-db`

---

**Last Updated:** 2026-03-20  
**Version:** 1.0.0
