# 📋 Claude Code Übergabe-Dokument: LLM-Proxy

**Projekt:** LLM-Proxy - Universal API Gateway für LLM-Provider  
**Stand:** 20. März 2026  
**Version:** 1.0.0 (Migration System implementiert)  
**Entwickler:** Gerd Krieger (gerd.krieger@gmail.com)

---

## 📑 Inhaltsverzeichnis

1. [Projektübersicht](#projektübersicht)
2. [Architektur & Technologie-Stack](#architektur--technologie-stack)
3. [Deployment-Strategie](#deployment-strategie)
4. [Database Migration System](#database-migration-system)
5. [Wichtige Befehle](#wichtige-befehle)
6. [Verzeichnisstruktur](#verzeichnisstruktur)
7. [Konfiguration](#konfiguration)
8. [Zugänge & Credentials](#zugänge--credentials)
9. [Git-Repository Setup](#git-repository-setup)
10. [Bekannte Issues & Lessons Learned](#bekannte-issues--lessons-learned)
11. [Troubleshooting](#troubleshooting)
12. [Monitoring & Logs](#monitoring--logs)
13. [Backup & Recovery](#backup--recovery)
14. [Nächste Schritte](#nächste-schritte)

---

## Projektübersicht

### Was ist LLM-Proxy?

Ein **Universal API Gateway** für verschiedene LLM-Provider (Claude, OpenAI, etc.) mit:

- **Unified API**: Einheitliche API für alle Provider
- **OAuth 2.0**: Sichere Authentifizierung für Web-Apps
- **API Key Management**: Verwaltung von Client-API-Keys
- **Rate Limiting**: Redis-basiertes Rate Limiting
- **Request Logging**: Detaillierte Logs aller Anfragen
- **Monitoring**: Prometheus/Grafana Dashboards
- **Admin UI**: React-basiertes Admin-Interface
- **Landing Page**: Marketing-Website

### Projektziele

✅ **Zentrale LLM-Gateway**: Alle LLM-Anfragen über einen Proxy  
✅ **Multi-Tenancy**: Verschiedene Clients mit eigenen Keys  
✅ **Security**: OAuth 2.0, API Keys, Rate Limiting  
✅ **Observability**: Metrics, Logs, Traces  
✅ **Zero-Downtime Deployment**: Registry-basiertes Deployment  
✅ **Automated Migrations**: Keine Schema-Mismatch-Ausfälle mehr  

---

## Architektur & Technologie-Stack

### Komponenten

```
┌────────────────────────────────────────────────────────────┐
│                    LLM-PROXY ARCHITEKTUR                   │
└────────────────────────────────────────────────────────────┘

┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Landing   │     │  Admin UI   │     │   Clients   │
│   (Nginx)   │     │   (React)   │     │  (Web Apps) │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                    │
       └───────────────────┴────────────────────┘
                           │
                  ┌────────▼─────────┐
                  │   Nginx Proxy    │
                  │  (Port 8080)     │
                  └────────┬─────────┘
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
┌──────▼──────┐   ┌────────▼────────┐   ┌─────▼─────┐
│   Landing   │   │   Backend API   │   │ Admin UI  │
│ /           │   │   /api/v1/*     │   │  /admin/* │
└─────────────┘   └────────┬────────┘   └───────────┘
                           │
       ┌───────────────────┼───────────────────┐
       │                   │                   │
┌──────▼──────┐   ┌────────▼────────┐   ┌─────▼─────┐
│ PostgreSQL  │   │     Redis       │   │ Prometheus│
│ (Port 5432) │   │  (Port 6379)    │   │(Port 9090)│
└─────────────┘   └─────────────────┘   └─────┬─────┘
                                              │
                                        ┌─────▼─────┐
                                        │  Grafana  │
                                        │(Port 3000)│
                                        └───────────┘
```

### Technologie-Stack

#### Backend (Go)
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL (mit golang-migrate)
- **Cache**: Redis
- **Auth**: OAuth 2.0, JWT
- **Metrics**: Prometheus
- **Logging**: Structured Logging (JSON)

#### Frontend
- **Admin UI**: React 18, TypeScript, Vite
- **Landing Page**: React 18, TypeScript, Vite
- **Styling**: Tailwind CSS

#### Infrastructure
- **Container**: Docker, Docker Compose
- **Registries**: GitHub Container Registry, GitLab Container Registry
- **Proxy**: Nginx
- **Monitoring**: Prometheus + Grafana
- **Deployment**: SSH-basiert, registry-pull

#### Database Migrations
- **Tool**: golang-migrate v4.17.0
- **Location**: `migrations/` (13 Migrationen aktuell)
- **Tracking**: `schema_migrations` Tabelle

---

## Deployment-Strategie

### Registry-Based Deployment

**Prinzip**: Keine Source-Code auf dem Server, nur Images aus Registries.

### Workflow

```
LOCAL MACHINE                    REGISTRIES                 PRODUCTION SERVER
─────────────                    ──────────                 ─────────────────

1. Code ändern
   git commit
   git push
                    ┌────────────────────────┐
2. make release  ──▶│ GitHub CR (Backend/UI) │
   VERSION=v1.2.0   │ GitLab CR (Landing)    │
                    └────────────────────────┘
                                │
                    3. docker pull ◀───────────── 4. Deploy
                                                   /opt/llm-proxy/
                                                   ├── .env
                                                   ├── docker-compose.yml
                                                   ├── migrations/
                                                   └── scripts/
```

### Deployment-Reihenfolge (WICHTIG!)

```
1. Pre-Checks
   ├─ Compose file exists
   ├─ .env file exists
   ├─ PostgreSQL running
   ├─ DB connectivity
   └─ Network exists

2. Backup Database ⭐ AUTOMATIC

3. Run Migrations ⭐ AUTOMATIC
   ├─ Check current version
   ├─ Run golang-migrate up
   └─ If FAIL → Restore backup + Abort

4. Pull Images
   ├─ Backend
   ├─ Admin UI
   └─ Landing

5. Deploy Containers
   └─ docker compose up -d

6. Health Checks
   ├─ Backend /health
   ├─ Frontend accessible
   └─ All containers running

7. Cleanup
   └─ Remove old images
```

### Image-Registries

| Komponente | Registry | Image |
|------------|----------|-------|
| Backend | GitHub CR | `ghcr.io/gerdkrieger/llm-proxy-backend:VERSION` |
| Admin UI | GitHub CR | `ghcr.io/gerdkrieger/llm-proxy-admin-ui:VERSION` |
| Landing | GitLab CR | `registry.gitlab.com/krieger-engineering/llm-proxy-landing:VERSION` |

**⚠️ WICHTIG**: Landing Page geht **NUR** zu GitLab CR (nicht GitHub)!

---

## Database Migration System

### Das Problem (2026-02-04 Incident)

**2-Stunden Produktions-Ausfall**:
- Backend deployed mit neuem Code, der neue DB-Spalten erwartet
- Migrationen wurden **NICHT** vor Deployment ausgeführt
- Application crashte: `column client_secret_hash does not exist`
- Erforderte emergency SSH + manuelle SQL-Ausführung

### Die Lösung: Automatisierte Migrationen

**Seit 20. März 2026**: Migrationen laufen **automatisch** vor Container-Deployment.

#### Migration-Flow

```
make release VERSION=v1.2.0
  │
  ├─ Build images
  ├─ Push to registries
  │
  └─ Server Deployment:
      │
      ├─ 1. Backup DB (automatic)
      ├─ 2. Run Migrations (automatic) ⭐
      │    ├─ golang-migrate in Docker
      │    ├─ Check version
      │    ├─ Apply pending migrations
      │    └─ IF FAIL → Restore + Abort
      ├─ 3. Deploy containers
      └─ 4. Health checks
```

#### Migration-Dateien

```
migrations/
├── 000001_init.up.sql                    # Initial schema
├── 000001_init.down.sql                  # Rollback
├── 000002_add_users.up.sql
├── 000002_add_users.down.sql
├── ...
└── 000013_fix_provider_configs_uuid.up.sql (aktuell)
```

**Naming**: `{6-digit-version}_{description}.{up|down}.sql`

#### Version Tracking

```sql
-- Tabelle: schema_migrations
CREATE TABLE schema_migrations (
    version BIGINT PRIMARY KEY,    -- z.B. 000013
    dirty BOOLEAN NOT NULL         -- false = clean, true = failed
);
```

#### Wichtige Migration-Befehle

```bash
# Status prüfen
make migrate-status

# Pending Migrationen anzeigen
make migrate-pending

# Alle Migrationen anwenden
make migrate-up

# Neue Migration erstellen
make migrate-create NAME=add_user_roles

# Migrationen zum Server synchronisieren
make migrate-sync

# Force version (dirty state recovery)
./scripts/deployment/migrate.sh force 000013
```

### Dirty State Recovery

**Problem**: Migration failed, DB ist "dirty"

**Lösung**:
1. DB-Status manuell prüfen
2. Migration manuell komplettieren ODER zurückrollen
3. Version forcen: `./scripts/deployment/migrate.sh force 000013`

**Details**: Siehe `docs/deployment/DATABASE_MIGRATIONS.md`

---

## Wichtige Befehle

### Deployment Commands

```bash
# Full Release (Build + Push + Deploy + Migrations)
make release VERSION=v1.2.0

# Step-by-step
make build VERSION=v1.2.0         # Build images
make push VERSION=v1.2.0          # Push to registries
make deploy-prod VERSION=v1.2.0   # Deploy to server

# Rollback
make rollback VERSION=v1.1.0
```

### Migration Commands

```bash
# Status & Pending
make migrate-status               # Current version
make migrate-pending              # List pending

# Apply Migrations
make migrate-up                   # Apply all
make migrate-up-one               # Apply next 1

# Rollback (VORSICHT!)
make migrate-down                 # Rollback last

# Create & Sync
make migrate-create NAME=xyz      # Create new migration
make migrate-sync                 # Sync to server
```

### Monitoring Commands

```bash
# Status
make status                       # All containers

# Logs
make logs SERVICE=backend         # Backend logs
make logs SERVICE=admin-ui        # Admin UI logs
make logs SERVICE=landing         # Landing logs

# Database
make backup-db                    # Backup PostgreSQL
```

### Registry Commands

```bash
# Login (one-time)
make login-registries

# Manual login
docker login ghcr.io -u gerdkrieger
docker login registry.gitlab.com
```

### Server Commands (via SSH)

```bash
# SSH to server
ssh openweb

# Check containers
docker ps

# Check logs
docker logs llm-proxy-backend
docker logs llm-proxy-admin-ui

# Check database
docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy

# Check migrations
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy \
  -c "SELECT version, dirty FROM schema_migrations;"
```

---

## Verzeichnisstruktur

### Local Machine

```
llm-proxy/
├── cmd/
│   └── server/
│       └── main.go                      # Backend entry point
├── internal/
│   ├── api/                             # API handlers
│   ├── auth/                            # OAuth & JWT
│   ├── database/                        # Database layer
│   ├── llm/                             # LLM provider clients
│   └── middleware/                      # Middleware (auth, logging)
├── pkg/
│   └── ...                              # Shared packages
├── admin-ui/                            # React Admin UI
│   ├── src/
│   ├── Dockerfile
│   └── package.json
├── landing/                             # React Landing Page
│   ├── src/
│   ├── Dockerfile
│   └── package.json
├── migrations/                          # Database migrations
│   ├── 000001_init.up.sql
│   ├── 000001_init.down.sql
│   └── ...
├── scripts/
│   └── deployment/
│       ├── build-and-push.sh           # Build & push images
│       ├── server-deploy.sh            # Server deployment script
│       └── migrate.sh                  # Migration tool
├── deployments/
│   └── docker/
│       ├── docker-compose.registry-deploy.yml
│       ├── .env.example
│       ├── Dockerfile                  # Backend Dockerfile
│       ├── nginx.conf                  # Nginx config
│       └── prometheus.yml              # Prometheus config
├── docs/
│   ├── deployment/
│   │   ├── DATABASE_MIGRATIONS.md      # Migration docs ⭐
│   │   ├── REGISTRY_DEPLOYMENT.md      # Full deployment guide
│   │   ├── DEPLOYMENT_FLOW.md          # Phase-by-phase
│   │   ├── CHEATSHEET.md               # Command reference
│   │   └── QUICKSTART_REGISTRY.md      # Quick start
│   ├── GIT_DUAL_PUSH.md                # Dual-push setup
│   └── UEBERGABE_CLAUDE_CODE.md        # This document
├── Makefile.registry                    # Main Makefile
├── go.mod
├── go.sum
└── README.md
```

### Production Server (`/opt/llm-proxy/`)

```
/opt/llm-proxy/
├── deployments/
│   └── docker/
│       ├── .env                         # ⚠️ SECRET! Config
│       ├── docker-compose.registry-deploy.yml
│       ├── nginx.conf
│       └── prometheus.yml
├── migrations/                          # Synced from local
│   ├── 000001_init.up.sql
│   ├── 000001_init.down.sql
│   └── ...
├── scripts/
│   └── deployment/
│       └── server-deploy.sh             # Deployment script
└── backups/                             # Not in structure, see below

/opt/llm-proxy-backups/                  # Separate backup dir
├── 20260320/
│   ├── postgres-backup-143000.sql
│   ├── postgres-backup-150000.sql
│   └── ...
└── ...
```

**⚠️ WICHTIG**: Auf dem Server ist **KEIN Source-Code**, nur:
- Config files (`.env`, `docker-compose.yml`)
- Deployment scripts
- Migration SQL files
- Backups

---

## Konfiguration

### Environment Variables (`.env`)

**Location auf Server**: `/opt/llm-proxy/deployments/docker/.env`

**Kritische Variablen**:

```bash
# Database
DB_HOST=llm-proxy-postgres
DB_PORT=5432
DB_USER=proxy_user
DB_PASSWORD=<STRONG_PASSWORD>              # ⚠️ SECRET
DB_NAME=llm_proxy
DB_SSLMODE=disable

# Redis
REDIS_HOST=llm-proxy-redis
REDIS_PORT=6379
REDIS_PASSWORD=                            # Optional

# Backend
PORT=8080
OAUTH_JWT_SECRET=<RANDOM_32_CHARS>         # ⚠️ SECRET
ADMIN_API_KEYS=<ADMIN_KEY_1>,<ADMIN_KEY_2> # ⚠️ SECRET

# LLM Provider API Keys
CLAUDE_API_KEY=<YOUR_CLAUDE_KEY>           # ⚠️ SECRET
OPENAI_API_KEY=<YOUR_OPENAI_KEY>           # ⚠️ SECRET

# Monitoring
PROMETHEUS_PORT=9090
GRAFANA_ADMIN_PASSWORD=<GRAFANA_PASS>      # ⚠️ SECRET

# Deployment
VERSION=latest
GITHUB_CONTAINER_REGISTRY=ghcr.io/gerdkrieger
GITLAB_CONTAINER_REGISTRY=registry.gitlab.com/krieger-engineering
```

### Port Bindings (Security)

**⚠️ ALLE Ports sind auf `127.0.0.1` gebunden** (nicht öffentlich):

```yaml
services:
  backend:
    ports:
      - "127.0.0.1:8080:8080"    # Backend
  
  postgres:
    ports:
      - "127.0.0.1:5432:5432"    # PostgreSQL
  
  redis:
    ports:
      - "127.0.0.1:6379:6379"    # Redis
  
  prometheus:
    ports:
      - "127.0.0.1:9090:9090"    # Prometheus
  
  grafana:
    ports:
      - "127.0.0.1:3001:3000"    # Grafana
```

**Öffentlicher Zugriff**: Nur über Nginx Proxy auf Port 8080

### Docker Volumes

```bash
# Persistent volumes
llm-proxy-postgres-data       # PostgreSQL data
llm-proxy-redis-data          # Redis data
llm-proxy-prometheus-data     # Prometheus metrics
llm-proxy-grafana-data        # Grafana dashboards

# Network
llm-proxy-network             # Bridge network
```

---

## Zugänge & Credentials

### Git Repositories

#### GitLab (Primary)
- **URL**: `git@gitlab.com:krieger-engineering/llm-proxy.git`
- **Access**: SSH Key
- **Purpose**: Primary repository, Landing Page registry

#### GitHub (Mirror)
- **URL**: `https://github.com/gerdkrieger/llm-proxy.git`
- **Access**: HTTPS (token)
- **Purpose**: Mirror, Backend/Admin-UI registry

#### Dual-Push Setup
```bash
# Single git push → both remotes
git remote -v
# origin  git@gitlab.com:krieger-engineering/llm-proxy.git (fetch)
# origin  git@gitlab.com:krieger-engineering/llm-proxy.git (push)
# origin  https://github.com/gerdkrieger/llm-proxy.git (push)
```

**Setup**: Siehe `docs/GIT_DUAL_PUSH.md`

### Container Registries

#### GitHub Container Registry
- **URL**: `ghcr.io`
- **User**: `gerdkrieger`
- **Token**: Personal Access Token mit `write:packages` scope
- **Images**: Backend, Admin UI

```bash
docker login ghcr.io -u gerdkrieger
```

#### GitLab Container Registry
- **URL**: `registry.gitlab.com`
- **Project**: `krieger-engineering/llm-proxy`
- **Token**: Personal Access Token oder Deploy Token
- **Images**: Landing Page

```bash
docker login registry.gitlab.com
```

### SSH Server Access

```bash
# Production Server
ssh openweb

# User: <your-user>
# Key: ~/.ssh/id_rsa (or configured key)
```

**SSH Config** (`~/.ssh/config`):
```
Host openweb
    HostName <server-ip>
    User <username>
    IdentityFile ~/.ssh/id_rsa
```

### Database Access

**Local**:
```bash
# Via SSH tunnel
ssh -L 5432:127.0.0.1:5432 openweb

# Connect
psql -h localhost -U proxy_user -d llm_proxy
```

**On Server**:
```bash
docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
```

### Grafana Access

**URL**: `http://localhost:3001` (via SSH tunnel)

```bash
ssh -L 3001:127.0.0.1:3001 openweb
```

**Login**:
- User: `admin`
- Password: `<GRAFANA_ADMIN_PASSWORD>` (from `.env`)

---

## Git-Repository Setup

### Dual-Push Configuration

**Ein `git push` pushed zu BEIDEN Repos** (GitLab + GitHub).

#### Setup Commands

```bash
# 1. Remove GitHub remote if exists
git remote remove github 2>/dev/null || true

# 2. Add GitHub as second push URL to origin
git remote set-url --add --push origin https://github.com/gerdkrieger/llm-proxy.git

# 3. Keep GitLab as fetch and first push
git remote set-url origin git@gitlab.com:krieger-engineering/llm-proxy.git

# 4. Verify
git remote -v
```

**Result**:
```
origin  git@gitlab.com:krieger-engineering/llm-proxy.git (fetch)
origin  git@gitlab.com:krieger-engineering/llm-proxy.git (push)
origin  https://github.com/gerdkrieger/llm-proxy.git (push)
```

### Branch Strategy

- **master**: Production branch (deployed)
- Feature branches: Optional, merge to master
- No staging branch (currently)

### Commit Message Convention

```
<type>: <subject>

<optional body>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code refactoring
- `test`: Tests
- `chore`: Maintenance

**Examples**:
```bash
git commit -m "feat: Add user authentication"
git commit -m "fix: Resolve database connection timeout"
git commit -m "docs: Update deployment guide"
```

---

## Bekannte Issues & Lessons Learned

### 1. ⚠️ Migration MUSS vor Deployment laufen

**Incident: 2026-02-04**
- **Problem**: Backend deployed ohne Migrationen
- **Symptom**: `column client_secret_hash does not exist`
- **Downtime**: 2 Stunden
- **Lösung**: Automatisiertes Migration-System (20.03.2026)

**Lesson**: **NIEMALS** Backend deployen ohne Migrationen!

### 2. ⚠️ Landing Page nur zu GitLab CR

**Problem**: Landing Page wird versehentlich zu GitHub CR gepusht

**Grund**: GitLab-spezifisches Image, Lizenz/Zugriff

**Lösung**:
```bash
# build-and-push.sh prüft automatisch:
# Backend/Admin-UI → GitHub CR
# Landing Page → GitLab CR
```

### 3. ⚠️ Port Bindings auf 127.0.0.1

**Problem**: Datenbank war öffentlich erreichbar

**Lösung**: Alle Ports auf `127.0.0.1` binden
```yaml
ports:
  - "127.0.0.1:5432:5432"  # NOT "5432:5432"
```

### 4. ⚠️ .env File NICHT in Git

**Wichtig**: `.env` enthält Secrets → **NIEMALS** committen!

```bash
# .gitignore
.env
deployments/docker/.env
*.env.local
*.env.production
```

### 5. ⚠️ Backup vor Major Changes

**Lesson**: Immer Backup vor:
- Schema-Änderungen
- Daten-Migration
- Major Deployment

```bash
make backup-db
```

### 6. ⚠️ golang-migrate Dirty State

**Problem**: Migration failed, DB ist "dirty", keine neuen Migrationen möglich

**Symptom**:
```
error: Dirty database version 000014. Fix and force version.
```

**Lösung**:
1. DB-Status manuell prüfen
2. Migration manuell komplettieren/zurückrollen
3. `./scripts/deployment/migrate.sh force 000014`

**Details**: `docs/deployment/DATABASE_MIGRATIONS.md` → "Troubleshooting"

### 7. ⚠️ Docker Login Expiration

**Problem**: Registry login expired, deployment fails

**Lösung**:
```bash
# Re-login local
make login-registries

# Re-login server
ssh openweb
docker login ghcr.io -u gerdkrieger
docker login registry.gitlab.com
```

### 8. ⚠️ Nginx Routing

**Problem**: Backend API nicht erreichbar über Nginx

**Lösung**: Check `nginx.conf` routing:
```nginx
location /api/ {
    proxy_pass http://llm-proxy-backend:8080;
}
```

---

## Troubleshooting

### Deployment fehlgeschlagen

**Check**:
```bash
# 1. Check logs
make logs SERVICE=backend

# 2. Check migrations
make migrate-status

# 3. Check database
ssh openweb
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT 1;"

# 4. Check .env
ssh openweb
cat /opt/llm-proxy/deployments/docker/.env
```

### Container startet nicht

**Check**:
```bash
# 1. Container status
ssh openweb
docker ps -a

# 2. Logs
docker logs llm-proxy-backend

# 3. Health check
curl http://localhost:8080/health
```

### Database Connection Failed

**Check**:
```bash
# 1. PostgreSQL running?
docker ps | grep postgres

# 2. Network exists?
docker network ls | grep llm-proxy

# 3. Credentials correct?
cat /opt/llm-proxy/deployments/docker/.env | grep DB_
```

### Migration Dirty State

**Check**:
```bash
# 1. Check status
make migrate-status
# Output: Version: 000014 (dirty)

# 2. Check what failed
ssh openweb
docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
\dt  # List tables
\d table_name  # Describe table

# 3. Fix manually + force version
./scripts/deployment/migrate.sh force 000014
```

### Cannot Pull Images

**Check**:
```bash
# 1. Registry login
docker login ghcr.io -u gerdkrieger
docker login registry.gitlab.com

# 2. Check image exists
docker manifest inspect ghcr.io/gerdkrieger/llm-proxy-backend:v1.2.0

# 3. Check network
ping ghcr.io
ping registry.gitlab.com
```

### Health Check Failed

**Check**:
```bash
# 1. Backend health endpoint
curl http://localhost:8080/health

# 2. Check logs
docker logs llm-proxy-backend

# 3. Check database connection
docker exec llm-proxy-backend env | grep DB_
```

---

## Monitoring & Logs

### Prometheus Metrics

**URL**: `http://localhost:9090` (via SSH tunnel)

```bash
ssh -L 9090:127.0.0.1:9090 openweb
```

**Wichtige Metrics**:
- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `llm_provider_requests_total` - LLM provider requests
- `database_connections` - DB connection pool

### Grafana Dashboards

**URL**: `http://localhost:3001` (via SSH tunnel)

```bash
ssh -L 3001:127.0.0.1:3001 openweb
```

**Dashboards**:
- LLM-Proxy Overview
- Database Performance
- API Performance
- Error Rates

### Application Logs

```bash
# Backend logs
make logs SERVICE=backend
docker logs -f llm-proxy-backend

# Admin UI logs
make logs SERVICE=admin-ui

# Landing logs
make logs SERVICE=landing

# All logs
docker compose -f /opt/llm-proxy/deployments/docker/docker-compose.registry-deploy.yml logs -f
```

### Log Locations

**Container logs**: Via Docker (ephemeral)

**Persistent logs**: Not configured (future: Loki/ELK)

**Database logs**:
```bash
docker exec llm-proxy-postgres tail -f /var/log/postgresql/postgresql.log
```

---

## Backup & Recovery

### Automatic Backups

**When**: Before every migration (automatic)

**Location**: `/opt/llm-proxy-backups/YYYYMMDD/`

**Format**: `postgres-backup-HHMMSS.sql`

```
/opt/llm-proxy-backups/
├── 20260320/
│   ├── postgres-backup-143000.sql
│   ├── postgres-backup-150000.sql
│   └── postgres-backup-163000.sql
└── 20260321/
    └── postgres-backup-090000.sql
```

### Manual Backup

```bash
# Via Makefile
make backup-db

# Manual
ssh openweb
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > backup.sql
```

### Restore Database

```bash
# 1. Stop backend to prevent writes
ssh openweb
docker stop llm-proxy-backend

# 2. Restore from backup
docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < \
  /opt/llm-proxy-backups/20260320/postgres-backup-143000.sql

# 3. Verify
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy \
  -c "SELECT version FROM schema_migrations;"

# 4. Restart backend
docker start llm-proxy-backend
```

### Full Disaster Recovery

**Scenario**: Kompletter Datenverlust

**Steps**:

1. **Setup fresh server** (like initial setup)
2. **Create volumes + network**
3. **Deploy containers** with old version
4. **Restore database** from last backup
5. **Verify** data integrity
6. **Deploy** current version

**Details**: `docs/deployment/DATABASE_MIGRATIONS.md` → "Recovery Procedures"

---

## Nächste Schritte

### Immediate (bereit für Execution)

1. **Test Migration System**
   ```bash
   make migrate-sync
   make migrate-status
   make migrate-pending
   ```

2. **Test Deployment**
   ```bash
   make release VERSION=test
   ```

3. **Verify Monitoring**
   - Check Grafana dashboards
   - Verify Prometheus targets
   - Test alerts (if configured)

### Short-Term (nächste Woche)

1. **CI/CD Pipeline**
   - Automatische Builds bei git push
   - Automated tests
   - Automated deployments

2. **Alerting**
   - Grafana alerts
   - Email/Slack notifications
   - On-call rotation

3. **Backup Retention Policy**
   - Define retention (z.B. 30 Tage)
   - Automated cleanup script
   - Off-site backup storage

### Medium-Term (nächster Monat)

1. **Staging Environment**
   - Separate staging server
   - Test deployments before production
   - Automated testing

2. **Log Aggregation**
   - Loki oder ELK Stack
   - Centralized logging
   - Log retention policy

3. **Performance Optimization**
   - Database query optimization
   - Redis caching strategy
   - API response time optimization

### Long-Term (nächstes Quartal)

1. **High Availability**
   - Multiple backend instances
   - Load balancing
   - Database replication

2. **Auto-Scaling**
   - Kubernetes migration
   - Horizontal pod autoscaling
   - Resource limits

3. **Advanced Monitoring**
   - Distributed tracing (Jaeger)
   - APM (Application Performance Monitoring)
   - Business metrics dashboards

---

## Wichtige Links & Dokumentation

### Internal Documentation

- **Migration System**: `docs/deployment/DATABASE_MIGRATIONS.md` ⭐
- **Full Deployment Guide**: `docs/deployment/REGISTRY_DEPLOYMENT.md`
- **Deployment Flow**: `docs/deployment/DEPLOYMENT_FLOW.md`
- **Command Cheatsheet**: `docs/deployment/CHEATSHEET.md`
- **Quick Start**: `docs/deployment/QUICKSTART_REGISTRY.md`
- **Git Dual-Push**: `docs/GIT_DUAL_PUSH.md`
- **This Document**: `docs/UEBERGABE_CLAUDE_CODE.md`

### External Resources

- **golang-migrate**: https://github.com/golang-migrate/migrate
- **Gin Framework**: https://gin-gonic.com/
- **Docker Compose**: https://docs.docker.com/compose/
- **Prometheus**: https://prometheus.io/docs/
- **Grafana**: https://grafana.com/docs/

### Support Contacts

- **Developer**: Gerd Krieger (gerd.krieger@gmail.com)
- **Git**: GitLab (primary), GitHub (mirror)
- **Server**: openweb (SSH)

---

## Zusammenfassung: Kritische Punkte

### ✅ DOs

1. ✅ **IMMER** Migrationen vor Deployment (automatisch!)
2. ✅ Backups vor Major Changes
3. ✅ .env Secrets NIEMALS in Git
4. ✅ Ports auf 127.0.0.1 binden
5. ✅ Landing Page nur zu GitLab CR
6. ✅ Dual-Push Git verwenden
7. ✅ Migration-Status prüfen nach Deployment
8. ✅ Health Checks nach Deployment

### ❌ DON'Ts

1. ❌ **NIEMALS** Backend deployen ohne Migrationen
2. ❌ Source-Code auf Production Server
3. ❌ .env File committen
4. ❌ Direkt in Production DB ändern (ohne Migration)
5. ❌ Ports öffentlich binden
6. ❌ Force-Push zu main/master
7. ❌ Skip Pre-Commit Hooks
8. ❌ Migrationen ohne .down.sql

### 🚨 Emergency Contacts & Procedures

**Database Corrupted**:
1. Stop backend: `docker stop llm-proxy-backend`
2. Restore backup: `docker exec -i llm-proxy-postgres psql ... < backup.sql`
3. Verify: `make migrate-status`
4. Restart: `docker start llm-proxy-backend`

**Deployment Failed**:
1. Check logs: `make logs SERVICE=backend`
2. Check migrations: `make migrate-status`
3. Rollback: `make rollback VERSION=previous`

**Migration Dirty State**:
1. Check version: `make migrate-status`
2. Fix manually in DB
3. Force version: `./scripts/deployment/migrate.sh force VERSION`

---

## Changelog: Wichtige Änderungen

### 2026-03-20: Migration System Implementation
- ✅ Automatisiertes Migration-System implementiert
- ✅ Auto-Rollback bei Migration-Failure
- ✅ Umfangreiche Dokumentation erstellt
- ✅ Manual Migration Tools (`migrate.sh`)
- ✅ Makefile erweitert mit Migration-Commands

**Commits**:
- `3cd0fc6` - Add automated database migration system
- `f792296` - Add comprehensive database migration documentation

### 2026-03-08: Registry Deployment Strategy
- ✅ Registry-based deployment implementiert
- ✅ GitHub CR + GitLab CR Integration
- ✅ Server-Deployment-Script erstellt
- ✅ Dual-Push Git Setup
- ✅ Deployment-Dokumentation

### Previous: Initial Development
- Backend API entwickelt (Go/Gin)
- Admin UI entwickelt (React/TypeScript)
- Landing Page entwickelt (React/TypeScript)
- PostgreSQL Schema (13 Migrationen)
- OAuth 2.0 Authentication
- Rate Limiting (Redis)
- Monitoring (Prometheus/Grafana)

---

## Abschluss

**Status**: ✅ Production-ready mit automatisiertem Migration-System

**Nächster Schritt**: Test Deployment mit Migrationen

```bash
# Test Migration System
make migrate-sync
make migrate-status

# Test Full Deployment
make release VERSION=test
```

**Dokumentation**: Vollständig und aktuell (Stand: 20.03.2026)

**Support**: Bei Fragen → gerd.krieger@gmail.com

---

**Ende des Übergabe-Dokuments**

🚀 **Viel Erfolg mit LLM-Proxy!**
