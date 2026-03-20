# 🔄 Complete Deployment Flow

## Overview Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    LOKALE ENTWICKLUNG (Dein Computer)                    │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                          ┌─────────▼─────────┐
                          │  Code Änderungen  │
                          │  (Go, Svelte, HTML)│
                          └─────────┬─────────┘
                                    │
                          ┌─────────▼─────────┐
                          │   git commit      │
                          │   git push        │
                          └─────────┬─────────┘
                                    │
                    ┌───────────────┼───────────────┐
                    │               │               │
           ┌────────▼────────┐      │      ┌───────▼────────┐
           │  GitLab Repo    │      │      │  GitHub Repo   │
           │  (Primary)      │      │      │  (Mirror)      │
           └─────────────────┘      │      └────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────┐
│                    BUILD PHASE (Dein Computer)                           │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────▼───────────────┐
                    │  make release VERSION=v1.2.0  │
                    └───────────────┬───────────────┘
                                    │
            ┌───────────────────────┼───────────────────────┐
            │                       │                       │
    ┌───────▼────────┐     ┌───────▼────────┐     ┌───────▼────────┐
    │  Build Backend │     │ Build Admin-UI │     │ Build Landing  │
    │  (Go → Binary) │     │ (Svelte → JS)  │     │  (HTML → Nginx)│
    └───────┬────────┘     └───────┬────────┘     └───────┬────────┘
            │                      │                       │
            │              ┌───────▼────────┐              │
            │              │  npm run build │              │
            │              └───────┬────────┘              │
            │                      │                       │
    ┌───────▼────────┐     ┌───────▼────────┐     ┌───────▼────────┐
    │ Docker Image   │     │ Docker Image   │     │ Docker Image   │
    │ backend:v1.2.0 │     │ admin-ui:v1.2.0│     │ landing:v1.2.0 │
    └───────┬────────┘     └───────┬────────┘     └───────┬────────┘
            │                      │                       │
┌─────────────────────────────────────────────────────────────────────────┐
│                    REGISTRY PUSH (Internet)                              │
└─────────────────────────────────────────────────────────────────────────┘
            │                      │                       │
            │                      │                       │
    ┌───────▼────────────┐  ┌──────▼───────────┐  ┌──────▼──────────────┐
    │ GitHub Registry    │  │ GitHub Registry  │  │ GitLab Registry     │
    │ ghcr.io/backend    │  │ ghcr.io/admin-ui │  │ gitlab.com/landing  │
    │ :v1.2.0 + :latest  │  │ :v1.2.0 + :latest│  │ :v1.2.0 + :latest   │
    └────────────────────┘  └──────────────────┘  └─────────────────────┘
            │                      │                       │
            └──────────────────────┼───────────────────────┘
                                   │
┌─────────────────────────────────────────────────────────────────────────┐
│                    SERVER DEPLOYMENT (Production)                        │
└─────────────────────────────────────────────────────────────────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │ make deploy-prod v1.2.0     │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │ SSH: server-deploy.sh       │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │ 1. Backup PostgreSQL        │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │ 2. docker-compose pull      │
                    │    - Pull backend:v1.2.0    │
                    │    - Pull admin-ui:v1.2.0   │
                    │    - Pull landing:v1.2.0    │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │ 3. docker-compose up -d     │
                    │    (Zero-Downtime)          │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │ 4. Health Checks            │
                    │    Wait for: healthy        │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │ 5. Cleanup old images       │
                    └──────────────┬──────────────┘
                                   │
                                   ▼
                    ┌──────────────────────────────┐
                    │   ✅ DEPLOYMENT SUCCESS      │
                    └──────────────────────────────┘
                                   │
┌─────────────────────────────────────────────────────────────────────────┐
│                    RUNNING STATE (Production)                            │
└─────────────────────────────────────────────────────────────────────────┘
                                   │
        ┌──────────────────────────┼──────────────────────────┐
        │                          │                          │
┌───────▼────────┐      ┌──────────▼─────────┐      ┌───────▼────────┐
│ Backend        │      │ Admin-UI           │      │ Landing        │
│ 127.0.0.1:8080 │      │ 127.0.0.1:3005     │      │ 127.0.0.1:8090 │
│ (healthy)      │      │ (healthy)          │      │ (healthy)      │
└───────┬────────┘      └──────────┬─────────┘      └───────┬────────┘
        │                          │                        │
        └──────────────────────────┼────────────────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │  Caddy Reverse Proxy        │
                    │  :80 :443 (HTTPS)           │
                    └──────────────┬──────────────┘
                                   │
                    ┌──────────────▼──────────────┐
                    │  Public Internet            │
                    │  scrubgate.tech             │
                    │  scrubgate.com              │
                    └─────────────────────────────┘
```

---

## Phase-by-Phase Breakdown

### Phase 1: Development (Local)

**Duration:** Variable (hours/days)

```bash
# You work on your local machine
cd /home/krieger/Sites/golang-projekte/llm-proxy

# Make code changes
vim internal/infrastructure/providers/manager.go
vim admin-ui/src/components/Dashboard.svelte
vim landing/index.html

# Test locally
docker-compose -f docker-compose.dev.yml up

# Commit changes
git add .
git commit -m "Feature: Add weighted load balancing"
git push  # → GitLab + GitHub automatically
```

**Result:** Code is in version control (both GitLab & GitHub)

---

### Phase 2: Build (Local)

**Duration:** 2-5 minutes

```bash
# Single command builds everything
make release VERSION=v1.2.0
```

**What happens internally:**

```bash
# 1. Build Backend
docker build \
  -f deployments/docker/Dockerfile \
  -t ghcr.io/gerdkrieger/llm-proxy-backend:v1.2.0 \
  -t ghcr.io/gerdkrieger/llm-proxy-backend:latest \
  .

# 2. Build Admin-UI
docker build \
  -f admin-ui/Dockerfile \
  -t ghcr.io/gerdkrieger/llm-proxy-admin-ui:v1.2.0 \
  -t ghcr.io/gerdkrieger/llm-proxy-admin-ui:latest \
  ./admin-ui

# 3. Build Landing
docker build \
  -f landing/Dockerfile \
  -t registry.gitlab.com/krieger-engineering/llm-proxy-landing:v1.2.0 \
  -t registry.gitlab.com/krieger-engineering/llm-proxy-landing:latest \
  ./landing
```

**Result:** 3 Docker images built locally

---

### Phase 3: Registry Push (Local → Cloud)

**Duration:** 1-3 minutes (depends on internet speed)

```bash
# Automatically pushed by make release
# Or manually:
make push VERSION=v1.2.0
```

**What happens:**

```bash
# Push Backend to GitHub Registry
docker push ghcr.io/gerdkrieger/llm-proxy-backend:v1.2.0
docker push ghcr.io/gerdkrieger/llm-proxy-backend:latest

# Push Admin-UI to GitHub Registry
docker push ghcr.io/gerdkrieger/llm-proxy-admin-ui:v1.2.0
docker push ghcr.io/gerdkrieger/llm-proxy-admin-ui:latest

# Push Landing to GitLab Registry (ONLY!)
docker push registry.gitlab.com/krieger-engineering/llm-proxy-landing:v1.2.0
docker push registry.gitlab.com/krieger-engineering/llm-proxy-landing:latest
```

**Result:** Images available in cloud registries

---

### Phase 4: Server Deployment (Production)

**Duration:** 2-4 minutes

```bash
# Automatically deployed by make release
# Or manually:
make deploy-prod VERSION=v1.2.0
```

**What happens on server:**

#### Step 1: Pre-Deployment Checks
```bash
# Check if .env exists
test -f /opt/llm-proxy/deployments/docker/.env

# Check if network exists
docker network ls | grep llm-proxy-network

# Check if volumes exist
docker volume ls | grep llm-proxy
```

#### Step 2: Backup Database
```bash
# Automatic backup before deployment
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > \
    /opt/llm-proxy-backups/$(date +%Y%m%d)/postgres-backup-$(date +%H%M%S).sql
```

#### Step 3: Pull New Images
```bash
export VERSION=v1.2.0

# Pull from registries
docker pull ghcr.io/gerdkrieger/llm-proxy-backend:v1.2.0
docker pull ghcr.io/gerdkrieger/llm-proxy-admin-ui:v1.2.0
docker pull registry.gitlab.com/krieger-engineering/llm-proxy-landing:v1.2.0
```

#### Step 4: Deploy (Zero-Downtime)
```bash
# Start new containers
docker-compose -f docker-compose.registry-deploy.yml up -d

# What happens:
# 1. Old containers keep running
# 2. New containers start
# 3. Health checks run
# 4. If healthy, old containers stop
# 5. If unhealthy, rollback automatically
```

#### Step 5: Health Checks
```bash
# Wait for all containers to be healthy
while [ $(docker ps --filter "health=unhealthy" | wc -l) -gt 0 ]; do
    sleep 5
done

# Verify health
docker ps --format 'table {{.Names}}\t{{.Status}}'
```

#### Step 6: Cleanup
```bash
# Remove old unused images
docker image prune -f

# Keep last 7 days of backups
find /opt/llm-proxy-backups -type d -mtime +7 -delete
```

**Result:** New version running, old version cleaned up

---

### Phase 5: Running State (Production)

**Duration:** Indefinite (until next deployment)

**Active Components:**

```
┌─────────────────────────────────────────────────────┐
│ Docker Containers (All on 127.0.0.1)               │
├─────────────────────────────────────────────────────┤
│ llm-proxy-backend      :8080  (healthy)            │
│ llm-proxy-admin-ui     :3005  (healthy)            │
│ llm-proxy-landing      :8090  (healthy)            │
│ llm-proxy-postgres     :5433  (healthy)            │
│ llm-proxy-redis        :6380  (healthy)            │
│ llm-proxy-prometheus   :9092  (healthy)            │
│ llm-proxy-grafana      :3002  (healthy)            │
└─────────────────────────────────────────────────────┘
           │
           ▼
┌─────────────────────────────────────────────────────┐
│ Caddy Reverse Proxy                                 │
├─────────────────────────────────────────────────────┤
│ scrubgate.tech   → Backend API + Admin UI          │
│ scrubgate.com    → Landing Page                     │
│ HTTPS + Security Headers                            │
└─────────────────────────────────────────────────────┘
           │
           ▼
     Public Internet
```

---

## File Locations During Each Phase

### Local Development

```
/home/krieger/Sites/golang-projekte/llm-proxy/
├── Source Code (All files)
├── Git Repository
├── Build Tools
└── Documentation
```

### After Build

```
/home/krieger/Sites/golang-projekte/llm-proxy/
├── Source Code
└── Docker Images (cached)
    ├── backend:v1.2.0
    ├── admin-ui:v1.2.0
    └── landing:v1.2.0
```

### In Registries

```
ghcr.io/gerdkrieger/
├── llm-proxy-backend:v1.2.0
├── llm-proxy-backend:latest
├── llm-proxy-admin-ui:v1.2.0
└── llm-proxy-admin-ui:latest

registry.gitlab.com/krieger-engineering/
├── llm-proxy-landing:v1.2.0
└── llm-proxy-landing:latest
```

### Production Server (AFTER CLEANUP!)

```
/opt/llm-proxy/
├── deployments/docker/
│   ├── .env
│   ├── docker-compose.registry-deploy.yml
│   ├── prometheus.yml
│   └── grafana/
└── scripts/deployment/
    └── server-deploy.sh

Docker Volumes:
├── llm-proxy-postgres-data/
├── llm-proxy-redis-data/
├── llm-proxy-prometheus-data/
└── llm-proxy-grafana-data/
```

**NO SOURCE CODE ON SERVER!**

---

## Rollback Process

If deployment fails or bugs are found:

```bash
# Rollback to previous version
make rollback VERSION=v1.1.0
```

**What happens:**

```bash
# 1. Pull old version images
docker pull ghcr.io/gerdkrieger/llm-proxy-backend:v1.1.0
docker pull ghcr.io/gerdkrieger/llm-proxy-admin-ui:v1.1.0
docker pull registry.gitlab.com/krieger-engineering/llm-proxy-landing:v1.1.0

# 2. Stop current containers
docker-compose down

# 3. Restore database backup (if needed)
cat /opt/llm-proxy-backups/20260320/postgres-backup-*.sql | \
    docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy

# 4. Start old version
export VERSION=v1.1.0
docker-compose up -d

# 5. Verify health
docker ps --format 'table {{.Names}}\t{{.Status}}'
```

**Duration:** 2-3 minutes

---

## Monitoring During Deployment

### Real-Time Monitoring

```bash
# Terminal 1: Watch container status
watch -n 2 'docker ps --format "table {{.Names}}\t{{.Status}}"'

# Terminal 2: Follow logs
docker-compose logs -f

# Terminal 3: Watch health endpoints
watch -n 5 'curl -s http://localhost:8080/health'
```

### Post-Deployment Checks

```bash
# 1. Container status
make status

# 2. Health endpoints
curl http://localhost:8080/health  # Backend
curl http://localhost:3005/health  # Admin-UI
curl http://localhost:8090/health  # Landing

# 3. Public URLs
curl https://scrubgate.tech/health
curl https://scrubgate.com/

# 4. Database connection
docker exec llm-proxy-postgres pg_isready

# 5. Application logs
make logs SERVICE=backend
```

---

## Summary Timeline

| Phase | Duration | Location | What Happens |
|-------|----------|----------|--------------|
| 1. Development | Hours/Days | Local | Code changes, testing |
| 2. Build | 2-5 min | Local | Docker images built |
| 3. Push | 1-3 min | Local→Cloud | Images uploaded to registries |
| 4. Deploy | 2-4 min | Server | Pull, backup, deploy, health check |
| 5. Running | Indefinite | Server | Production state |

**Total Deployment Time:** ~5-12 minutes from code to production

---

## Key Differences: Old vs New

### ❌ Old Way (Source Code Deployment)

```bash
# On server
cd /opt/llm-proxy
git pull                      # Pull source code
docker-compose build          # Build on server (SLOW!)
docker-compose up -d          # Deploy
```

**Problems:**
- ❌ Source code on server
- ❌ Build on production (slow, risky)
- ❌ Dependency issues on server
- ❌ No atomic deployments
- ❌ Hard to rollback

### ✅ New Way (Registry Deployment)

```bash
# On local machine
make release VERSION=v1.2.0   # Build + push + deploy
```

**Benefits:**
- ✅ No source code on server
- ✅ Build locally (fast, safe)
- ✅ Consistent images
- ✅ Atomic deployments
- ✅ Easy rollback

---

**Last Updated:** 2026-03-20  
**Version:** 1.0
