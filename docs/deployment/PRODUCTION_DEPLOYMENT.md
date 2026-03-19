# Production Deployment Guide

## 🎯 Deployment-Strategie: Registry-Based

### ✅ Was auf dem Server sein sollte:

```
/opt/llm-proxy/
├── .env                    # Production configuration (ONLY config file!)
└── docker-compose.yml      # Production compose (pulls from registry)
```

**Das war's!** Kein Source-Code, keine Build-Tools, keine Secrets!

---

## 🚨 AKTUELLES PROBLEM

Der Server hat aktuell **ALLE Dateien** vom Repository:

```bash
# ❌ SHOULD NOT BE ON SERVER:
- cmd/, internal/, pkg/      # Source-Code
- admin-ui/                  # Source-Code
- go.mod, go.sum            # Build-Dependencies
- Makefile, .air.toml       # Build-Tools
- .git/, .gitlab/           # Git-Data
- tests/                    # Test-Files
- docs/, README.md          # Documentation
- .env.local, .env.example  # Development-Configs
- Dockerfile.dev            # Development-Dockerfile
- OPENWEBUI_TOKEN_*.txt     # 🔴 SECRETS IN PLAIN TEXT!
```

---

## 📋 Deployment-Schritte

### 1. Server aufräumen

```bash
# SSH auf Server
ssh deploy@your-server

# Backup existing .env
cd /opt/llm-proxy
cp .env /opt/llm-proxy-backup/.env

# Stop containers
docker-compose down

# ALLES löschen außer Docker Volumes
cd /opt
sudo rm -rf llm-proxy/

# Neu anlegen (nur was nötig ist)
sudo mkdir -p /opt/llm-proxy
sudo chown deploy:deploy /opt/llm-proxy
cd /opt/llm-proxy

# .env wiederherstellen
cp /opt/llm-proxy-backup/.env .env
```

---

### 2. Production docker-compose.yml erstellen

Erstelle `/opt/llm-proxy/docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:14-alpine
    container_name: llm-proxy-postgres
    restart: unless-stopped
    environment:
      POSTGRES_DB: ${DB_NAME:-llm_proxy}
      POSTGRES_USER: ${DB_USER:-proxy_user}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    ports:
      - "127.0.0.1:5433:5432"
    volumes:
      - llm-proxy-postgres-data:/var/lib/postgresql/data
    networks:
      - llm-proxy-network

  redis:
    image: redis:7-alpine
    container_name: llm-proxy-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --maxmemory 512mb
    ports:
      - "127.0.0.1:6380:6379"
    volumes:
      - llm-proxy-redis-data:/data
    networks:
      - llm-proxy-network

  backend:
    image: registry.gitlab.com/krieger-engineering/llm-proxy/backend:latest
    container_name: llm-proxy-backend
    restart: unless-stopped
    env_file: .env
    environment:
      DATABASE_HOST: postgres
      DATABASE_PORT: 5432
      REDIS_HOST: redis
      REDIS_PORT: 6379
      ENVIRONMENT: production
    ports:
      - "127.0.0.1:8080:8080"
    depends_on:
      - postgres
      - redis
    networks:
      - llm-proxy-network

  admin-ui:
    image: registry.gitlab.com/krieger-engineering/llm-proxy/admin-ui:latest
    container_name: llm-proxy-admin-ui
    restart: unless-stopped
    ports:
      - "127.0.0.1:3005:80"
    depends_on:
      - backend
    networks:
      - llm-proxy-network

volumes:
  llm-proxy-postgres-data:
    external: true
  llm-proxy-redis-data:
    external: true

networks:
  llm-proxy-network:
    external: true
```

---

### 3. Docker Volumes erstellen (einmalig)

```bash
# Create external volumes (persist data across deployments)
docker volume create llm-proxy-postgres-data
docker volume create llm-proxy-redis-data

# Create network
docker network create llm-proxy-network
```

---

### 4. GitLab Registry Login

```bash
# Login to GitLab Container Registry
docker login registry.gitlab.com

# Username: your-gitlab-username
# Password: your-gitlab-access-token (NOT your password!)
```

**GitLab Access Token erstellen:**
1. GitLab → Settings → Access Tokens
2. Name: `production-server-pull`
3. Scopes: `read_registry`
4. Create token
5. Token kopieren (nur einmal sichtbar!)

---

### 5. Images pullen und starten

```bash
cd /opt/llm-proxy

# Pull latest images from registry
docker-compose pull

# Start services
docker-compose up -d

# Check logs
docker-compose logs -f backend
```

---

## 🔄 Updates deployen

```bash
cd /opt/llm-proxy

# Pull new images
docker-compose pull

# Restart with new images
docker-compose up -d

# Check health
docker-compose ps
curl http://localhost:8080/health
```

**Das war's!** Kein Source-Code-Transfer, keine Builds auf dem Server!

---

## 🏗️ CI/CD Pipeline (GitLab)

Die `.gitlab-ci.yml` sollte:

1. **Build Stage**: Images bauen
2. **Push Stage**: Images zu Registry pushen
3. **Deploy Stage**: SSH auf Server → `docker-compose pull && up -d`

Beispiel `.gitlab-ci.yml`:

```yaml
stages:
  - build
  - push
  - deploy

build:
  stage: build
  script:
    - docker build -t $CI_REGISTRY_IMAGE/backend:$CI_COMMIT_SHORT_SHA .
    - docker build -t $CI_REGISTRY_IMAGE/admin-ui:$CI_COMMIT_SHORT_SHA ./admin-ui

push:
  stage: push
  script:
    - docker push $CI_REGISTRY_IMAGE/backend:$CI_COMMIT_SHORT_SHA
    - docker push $CI_REGISTRY_IMAGE/admin-ui:$CI_COMMIT_SHORT_SHA
    - docker tag $CI_REGISTRY_IMAGE/backend:$CI_COMMIT_SHORT_SHA $CI_REGISTRY_IMAGE/backend:latest
    - docker push $CI_REGISTRY_IMAGE/backend:latest

deploy:
  stage: deploy
  only:
    - master
  script:
    - ssh deploy@$PRODUCTION_SERVER "cd /opt/llm-proxy && docker-compose pull && docker-compose up -d"
```

---

## 🔒 Security Benefits

### ✅ Mit Registry-Deployment:

- ✅ **Kein Source-Code** auf Server
- ✅ **Keine Secrets** im Dateisystem
- ✅ **Atomic Updates** (pull → restart)
- ✅ **Rollback einfach** (alte Image-Version)
- ✅ **Kleinere Attack-Surface**
- ✅ **Audit-Trail** (welches Image läuft?)

### ❌ Mit Current-Deployment:

- ❌ Source-Code exposed
- ❌ Secrets in plain text
- ❌ Build-Tools auf Server
- ❌ Git-History auf Server
- ❌ Größere Attack-Surface

---

## 📝 Server-Cleanup-Script

Speichere als `/opt/cleanup-server.sh`:

```bash
#!/bin/bash
set -e

echo "🧹 Cleaning up production server..."

# Backup .env
mkdir -p /opt/llm-proxy-backup
cp /opt/llm-proxy/.env /opt/llm-proxy-backup/.env.$(date +%Y%m%d-%H%M%S)

# Stop containers
cd /opt/llm-proxy
docker-compose down

# List volumes (DON'T delete!)
echo "📦 Existing volumes (will be preserved):"
docker volume ls | grep llm-proxy

# Delete everything except volumes
cd /opt
rm -rf llm-proxy/

# Create clean directory
mkdir -p /opt/llm-proxy
chown deploy:deploy /opt/llm-proxy

echo "✅ Server cleaned!"
echo "📋 Next steps:"
echo "1. Copy .env from backup"
echo "2. Create docker-compose.yml"
echo "3. docker-compose pull && docker-compose up -d"
```

---

## ⚠️ Was NIEMALS auf Server sein sollte:

```bash
# 🔴 NEVER ON PRODUCTION:
cmd/                       # Source-Code
internal/                  # Source-Code
pkg/                       # Source-Code
admin-ui/src/             # Source-Code
tests/                    # Test-Files
docs/                     # Documentation
.git/                     # Git-Data
.gitlab/                  # GitLab-Data
go.mod, go.sum           # Dependencies
Makefile                 # Build-Tool
.air.toml                # Development-Tool
Dockerfile.dev           # Development-Dockerfile
.env.local               # Development-Config
.env.example             # Template
README.md                # Documentation
deploy.sh                # Deployment-Script
*.txt (tokens, configs)  # 🔴 SECRETS!
```

---

## ✅ Was auf Server sein sollte:

```bash
# ✅ PRODUCTION ONLY:
/opt/llm-proxy/
├── .env                    # Production config (from .env.production)
└── docker-compose.yml      # Production compose (registry-based)

# Docker manages:
/var/lib/docker/volumes/
├── llm-proxy-postgres-data/   # Persistent data
└── llm-proxy-redis-data/      # Persistent data
```

---

## 🎯 Zusammenfassung

| Aspekt | Aktuell ❌ | Richtig ✅ |
|--------|-----------|-----------|
| **Dateien auf Server** | ~50 Files | 2 Files |
| **Source-Code** | ✅ Ja | ❌ Nein |
| **Secrets** | ✅ Plain text | ❌ Env only |
| **Build-Tools** | ✅ Ja | ❌ Nein |
| **Deployment** | rsync/scp | docker pull |
| **Updates** | Full copy | Image pull |
| **Rollback** | Difficult | Easy |
| **Security** | 🔴 Low | 🟢 High |

---

## 📞 Support

Bei Fragen:
1. Dokumentation: `docs/deployment/`
2. GitLab CI/CD: `.gitlab-ci.yml`
3. Docker Registry: `registry.gitlab.com/krieger-engineering/llm-proxy`

---

**Nächster Schritt**: Server aufräumen und auf Registry-Deployment umstellen!
