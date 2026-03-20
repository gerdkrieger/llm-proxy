# 📝 Deployment Cheatsheet

Quick reference for common deployment tasks.

---

## 🚀 Quick Deployment

```bash
# Full release (build, push, deploy)
make release VERSION=v1.2.0

# Or step-by-step
make build VERSION=v1.2.0          # Build images
make push VERSION=v1.2.0           # Push to registries  
make deploy-prod VERSION=v1.2.0    # Deploy to server
```

---

## 🔑 Registry Login (One-Time)

```bash
# GitHub Container Registry
docker login ghcr.io -u gerdkrieger

# GitLab Container Registry
docker login registry.gitlab.com
```

---

## 📦 Build Commands

```bash
# Build all images with version tag
make build VERSION=v1.0.0

# Build and push in one step
make build-and-push VERSION=v1.0.0

# Build specific component (manual)
docker build -f deployments/docker/Dockerfile -t ghcr.io/gerdkrieger/llm-proxy-backend:v1.0.0 .
docker build -f admin-ui/Dockerfile -t ghcr.io/gerdkrieger/llm-proxy-admin-ui:v1.0.0 ./admin-ui
docker build -f landing/Dockerfile -t registry.gitlab.com/krieger-engineering/llm-proxy-landing:v1.0.0 ./landing
```

---

## 🌐 Deploy to Production

```bash
# Deploy specific version
make deploy-prod VERSION=v1.2.0

# Deploy latest
make deploy-prod VERSION=latest

# Manual deploy (SSH)
ssh openweb "cd /opt/llm-proxy && ./scripts/deployment/server-deploy.sh v1.2.0"
```

---

## 📊 Monitoring

```bash
# Check container status
make status

# View logs (tail -f)
make logs SERVICE=backend
make logs SERVICE=admin-ui
make logs SERVICE=landing

# Manual log viewing
ssh openweb "docker logs -f llm-proxy-backend"
ssh openweb "docker logs --tail 100 llm-proxy-admin-ui"
```

---

## 💾 Backup & Restore

### Backup

```bash
# Backup database
make backup-db

# Manual backup
ssh openweb "docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy" > backup.sql
```

### Restore

```bash
# Restore from backup
ssh openweb "cat backup.sql | docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy"
```

---

## ⏪ Rollback

```bash
# Rollback to previous version
make rollback VERSION=v1.1.0

# Emergency rollback (manual)
ssh openweb "cd /opt/llm-proxy && \
    export VERSION=v1.1.0 && \
    docker-compose -f deployments/docker/docker-compose.registry-deploy.yml pull && \
    docker-compose -f deployments/docker/docker-compose.registry-deploy.yml up -d"
```

---

## 🧹 Cleanup

```bash
# Clean old Docker images
make clean

# Clean on server
ssh openweb "docker image prune -f"

# Remove stopped containers
ssh openweb "docker container prune -f"

# Remove unused volumes (CAREFUL!)
ssh openweb "docker volume prune -f"
```

---

## 🔍 Troubleshooting

### Container won't start

```bash
# Check logs
ssh openweb "docker logs llm-proxy-backend"

# Check health
ssh openweb "docker inspect llm-proxy-backend | grep -A 10 Health"

# Restart specific container
ssh openweb "docker restart llm-proxy-backend"
```

### Database connection issues

```bash
# Check PostgreSQL is running
ssh openweb "docker exec llm-proxy-postgres pg_isready"

# Check connectivity from backend
ssh openweb "docker exec llm-proxy-backend ping llm-proxy-postgres"

# Check database exists
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -l"
```

### Image pull fails

```bash
# Re-login to registry
docker login ghcr.io -u gerdkrieger
docker login registry.gitlab.com

# Test pull manually
docker pull ghcr.io/gerdkrieger/llm-proxy-backend:latest

# Check network
curl -I https://ghcr.io
```

### Disk space issues

```bash
# Check disk usage
ssh openweb "df -h"

# Check Docker disk usage
ssh openweb "docker system df"

# Clean everything (CAREFUL!)
ssh openweb "docker system prune -a -f"
```

---

## 🌍 Server Access

### SSH

```bash
# SSH to production server
ssh openweb

# SSH with command
ssh openweb "docker ps"

# Copy files to server
scp file.txt openweb:/opt/llm-proxy/

# Copy files from server
scp openweb:/opt/llm-proxy/file.txt ./
```

### File Locations

```bash
# Deployment files
/opt/llm-proxy/deployments/docker/docker-compose.registry-deploy.yml
/opt/llm-proxy/deployments/docker/.env

# Scripts
/opt/llm-proxy/scripts/deployment/server-deploy.sh

# Backups
/opt/llm-proxy-backups/YYYYMMDD/

# Logs
docker logs llm-proxy-backend
docker logs llm-proxy-admin-ui
```

---

## 📡 Service URLs

### Local (via SSH tunnel)

```bash
# Backend API
curl http://localhost:8080/health

# Admin UI
curl http://localhost:3005/health

# Landing Page
curl http://localhost:8090/health
```

### Public (via Caddy)

```bash
# Backend API + Admin UI
https://scrubgate.tech
https://scrubgate.tech/admin/clients

# Landing Page
https://scrubgate.com
```

---

## 🔐 Environment Variables

### Required in .env

```bash
# Database
DB_PASSWORD=<STRONG_PASSWORD>
DB_NAME=llm_proxy
DB_USER=proxy_user

# Backend
OAUTH_JWT_SECRET=<32_CHAR_RANDOM>
ADMIN_API_KEYS=<YOUR_KEY>

# API Keys
CLAUDE_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
```

### View on server (CAREFUL!)

```bash
# View .env (sensitive!)
ssh openweb "cat /opt/llm-proxy/deployments/docker/.env"

# Check specific variable
ssh openweb "cd /opt/llm-proxy/deployments/docker && grep DB_PASSWORD .env"
```

---

## 🔄 Update Workflow

```bash
# 1. Make code changes
git add .
git commit -m "Feature: Add new functionality"
git push

# 2. Build new version
make build VERSION=v1.2.0

# 3. Test locally (optional)
docker-compose -f deployments/docker/docker-compose.registry-deploy.yml up

# 4. Push to registries
make push VERSION=v1.2.0

# 5. Deploy to production
make deploy-prod VERSION=v1.2.0

# 6. Verify
make status
make logs SERVICE=backend
```

---

## 🆘 Emergency Commands

### Stop all containers

```bash
ssh openweb "cd /opt/llm-proxy && \
    docker-compose -f deployments/docker/docker-compose.registry-deploy.yml down"
```

### Restart all containers

```bash
ssh openweb "cd /opt/llm-proxy && \
    docker-compose -f deployments/docker/docker-compose.registry-deploy.yml restart"
```

### Force recreate

```bash
ssh openweb "cd /opt/llm-proxy && \
    docker-compose -f deployments/docker/docker-compose.registry-deploy.yml up -d --force-recreate"
```

### Nuclear option (rebuild everything)

```bash
ssh openweb "cd /opt/llm-proxy && \
    docker-compose -f deployments/docker/docker-compose.registry-deploy.yml down && \
    docker system prune -a -f && \
    ./scripts/deployment/server-deploy.sh latest"
```

---

## 📊 Health Checks

### All containers

```bash
ssh openweb "docker ps --format 'table {{.Names}}\t{{.Status}}'"
```

### Specific container

```bash
ssh openweb "curl -f http://localhost:8080/health"  # Backend
ssh openweb "curl -f http://localhost:3005/health"  # Admin-UI
ssh openweb "curl -f http://localhost:8090/health"  # Landing
```

### Database

```bash
ssh openweb "docker exec llm-proxy-postgres pg_isready"
ssh openweb "docker exec llm-proxy-redis redis-cli ping"
```

---

## 🏷️ Version Management

### List versions in registry

```bash
# GitHub (backend)
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
    https://api.github.com/user/packages/container/llm-proxy-backend/versions

# Or via Docker
docker image ls | grep llm-proxy
```

### Tag new version

```bash
# Tag locally
docker tag ghcr.io/gerdkrieger/llm-proxy-backend:latest \
    ghcr.io/gerdkrieger/llm-proxy-backend:v1.2.0

# Push both tags
docker push ghcr.io/gerdkrieger/llm-proxy-backend:v1.2.0
docker push ghcr.io/gerdkrieger/llm-proxy-backend:latest
```

---

## 🆘 Emergency Contacts

- **Makefile Help:** `make help`
- **Full Docs:** `docs/deployment/REGISTRY_DEPLOYMENT.md`
- **Quick Start:** `docs/deployment/QUICKSTART_REGISTRY.md`
- **Architecture:** `docs/deployment/DEPLOYMENT_ARCHITECTURE.md`

---

**Last Updated:** 2026-03-20  
**Cheatsheet Version:** 1.0
