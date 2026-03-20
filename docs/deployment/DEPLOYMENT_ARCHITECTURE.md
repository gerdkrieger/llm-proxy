# 🏗️ Deployment Architecture

## Overview Diagram

```
┌──────────────────────────────────────────────────────────────────────────┐
│                         LOCAL DEVELOPMENT                                 │
│                                                                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │   Backend    │  │  Admin-UI    │  │   Landing    │                  │
│  │  (Go Code)   │  │  (Svelte)    │  │   (HTML)     │                  │
│  └───────┬──────┘  └───────┬──────┘  └───────┬──────┘                  │
│          │                 │                 │                          │
│          └────────┬────────┴────────┬────────┘                          │
│                   │                 │                                    │
│              ┌────▼─────────────────▼────┐                              │
│              │  make release v1.0.0      │                              │
│              │  - Build Docker images    │                              │
│              │  - Tag with version       │                              │
│              └────┬─────────────────┬────┘                              │
│                   │                 │                                    │
└───────────────────┼─────────────────┼────────────────────────────────────┘
                    │                 │
                    │ Push            │ Push
                    │                 │
         ┌──────────▼─────┐     ┌────▼─────────────────┐
         │ GitHub CR      │     │ GitLab CR            │
         │ (ghcr.io)      │     │ (registry.gitlab)    │
         ├────────────────┤     ├──────────────────────┤
         │ • backend      │     │ • landing (ONLY!)    │
         │ • admin-ui     │     │                      │
         └────────┬───────┘     └─────────┬────────────┘
                  │                       │
                  │ Pull                  │ Pull
                  │                       │
┌─────────────────┼───────────────────────┼────────────────────────────────┐
│                 │  PRODUCTION SERVER    │                                 │
│                 │                       │                                 │
│         ┌───────▼───────────────────────▼────────┐                       │
│         │  docker-compose pull                   │                       │
│         │  docker-compose up -d                  │                       │
│         └───────┬────────────────────────────────┘                       │
│                 │                                                         │
│    ┌────────────┼────────────┬──────────────┬──────────────┐            │
│    │            │            │              │              │            │
│ ┌──▼───┐  ┌────▼─────┐  ┌───▼────┐  ┌─────▼──────┐  ┌───▼──────┐     │
│ │Backend│  │ Admin-UI │  │Landing │  │ PostgreSQL │  │  Redis   │     │
│ │:8080  │  │ :3005    │  │ :8090  │  │ :5433      │  │  :6380   │     │
│ └───────┘  └──────────┘  └────────┘  └────────────┘  └──────────┘     │
│                                                                          │
│    All containers bound to 127.0.0.1 (localhost only)                   │
│                                                                          │
│         ┌────────────────────────────────────┐                          │
│         │  Caddy Reverse Proxy               │                          │
│         ├────────────────────────────────────┤                          │
│         │  scrubgate.tech → Backend/Admin    │                          │
│         │  scrubgate.com  → Landing          │                          │
│         └────────┬───────────────────────────┘                          │
│                  │                                                       │
└──────────────────┼───────────────────────────────────────────────────────┘
                   │
                   ▼
          ┌────────────────┐
          │   INTERNET     │
          │   (HTTPS)      │
          └────────────────┘
```

---

## Component Flow

### 1. Build Phase (Local)

```
Developer → Code Change → Git Commit → make release VERSION=v1.0.0
                                              │
                                              ├─→ Build Backend Image
                                              ├─→ Build Admin-UI Image
                                              └─→ Build Landing Image
```

### 2. Registry Push (Automated)

```
Built Images → Tag with version → Push to registries
                                        │
                      ┌─────────────────┼─────────────────┐
                      │                 │                 │
                  Backend           Admin-UI          Landing
                      │                 │                 │
                      ▼                 ▼                 ▼
              ghcr.io/backend   ghcr.io/admin    gitlab/landing
```

### 3. Server Pull (Automated)

```
Production Server → docker-compose pull → Downloads images from registries
                                               │
                                               ├─→ backend:v1.0.0
                                               ├─→ admin-ui:v1.0.0
                                               └─→ landing:v1.0.0
```

### 4. Deployment (Zero-Downtime)

```
Server Deployment Script:
1. Backup PostgreSQL database
2. Pull new images from registries
3. Stop old containers (gracefully)
4. Start new containers
5. Wait for health checks
6. Verify all containers healthy
7. Cleanup old images
```

---

## Registry Strategy

### Why Two Registries?

| Aspect | GitHub Registry | GitLab Registry |
|--------|----------------|-----------------|
| **Backend** | ✅ Yes | ❌ No |
| **Admin-UI** | ✅ Yes | ❌ No |
| **Landing** | ❌ No | ✅ Yes (ONLY!) |
| **Reason** | Main repo on GitHub | Landing is GitLab exclusive |
| **URL** | ghcr.io/gerdkrieger | registry.gitlab.com/krieger-engineering |

### Image Naming Convention

```
<registry>/<user>/<component>:<version>

Examples:
  ghcr.io/gerdkrieger/llm-proxy-backend:v1.0.0
  ghcr.io/gerdkrieger/llm-proxy-backend:latest
  ghcr.io/gerdkrieger/llm-proxy-admin-ui:v1.0.0
  registry.gitlab.com/krieger-engineering/llm-proxy-landing:v1.0.0
```

---

## Network Architecture

### Production Server Network

```
┌─────────────────────────────────────────────────────────┐
│ Host: openweb (DigitalOcean Droplet)                    │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │ Docker Network: llm-proxy-network              │    │
│  │                                                 │    │
│  │  ┌──────────┐  ┌──────────┐  ┌─────────────┐ │    │
│  │  │ Backend  │  │ Admin-UI │  │  Landing    │ │    │
│  │  │ 172.21.  │  │ 172.21.  │  │  172.21.    │ │    │
│  │  │ 0.2:8080 │  │ 0.3:80   │  │  0.7:80     │ │    │
│  │  └─────┬────┘  └─────┬────┘  └──────┬──────┘ │    │
│  │        │             │              │         │    │
│  │        │      ┌──────┴──────┐       │         │    │
│  │        └──────┤ PostgreSQL  ├───────┘         │    │
│  │               │ 172.21.0.4  │                 │    │
│  │               └──────┬──────┘                 │    │
│  │                      │                        │    │
│  │               ┌──────▼──────┐                 │    │
│  │               │   Redis     │                 │    │
│  │               │ 172.21.0.5  │                 │    │
│  │               └─────────────┘                 │    │
│  └────────────────────────────────────────────────┘    │
│                                                          │
│  ┌────────────────────────────────────────────────┐    │
│  │ Host Network (127.0.0.1)                       │    │
│  │                                                 │    │
│  │  :8080 → Backend                               │    │
│  │  :3005 → Admin-UI                              │    │
│  │  :8090 → Landing                               │    │
│  │  :5433 → PostgreSQL (localhost only!)          │    │
│  │  :6380 → Redis (localhost only!)               │    │
│  └────────────────────────────────────────────────┘    │
│                      │                                   │
│               ┌──────▼──────┐                           │
│               │    Caddy    │                           │
│               │  :80 :443   │                           │
│               └──────┬──────┘                           │
└──────────────────────┼──────────────────────────────────┘
                       │
              ┌────────▼────────┐
              │  Public Internet │
              │  (HTTPS Only)    │
              └──────────────────┘
```

### Port Mapping Summary

| Service | Container Port | Host Port | Public Access |
|---------|---------------|-----------|---------------|
| Backend | 8080 | 127.0.0.1:8080 | Via Caddy (scrubgate.tech) |
| Admin-UI | 80 | 127.0.0.1:3005 | Via Caddy (scrubgate.tech) |
| Landing | 80 | 127.0.0.1:8090 | Via Caddy (scrubgate.com) |
| PostgreSQL | 5432 | 127.0.0.1:5433 | ❌ NEVER PUBLIC |
| Redis | 6379 | 127.0.0.1:6380 | ❌ NEVER PUBLIC |

---

## Security Layers

```
┌─────────────────────────────────────────┐
│ Layer 1: DigitalOcean Firewall          │  ← Hoster-Level
├─────────────────────────────────────────┤
│ Layer 2: UFW (Ubuntu Firewall)          │  ← OS-Level
├─────────────────────────────────────────┤
│ Layer 3: 127.0.0.1 Port Bindings        │  ← Docker-Level
├─────────────────────────────────────────┤
│ Layer 4: Caddy Reverse Proxy            │  ← Application-Level
│         - HTTPS only                     │
│         - Security headers               │
│         - Rate limiting                  │
└─────────────────────────────────────────┘
```

---

## Deployment States

### Version Lifecycle

```
Development → Build → Registry → Production
                │
                ├─→ v1.0.0 (latest)    ← Currently deployed
                ├─→ v1.0.1             ← Hotfix ready
                ├─→ v1.1.0-dev         ← Testing
                └─→ v2.0.0-beta        ← Future release
```

### Rollback Process

```
Production (v1.0.1) → Problem Detected
                          │
                          ├─→ Health check fails
                          ├─→ Manual intervention
                          │
                          ▼
                    Rollback to v1.0.0
                          │
                          ├─→ Pull old image
                          ├─→ Restore DB backup
                          ├─→ Deploy v1.0.0
                          │
                          ▼
                    ✅ Service restored
```

---

## Data Flow

### Request Lifecycle

```
User → HTTPS → Caddy → Backend Container → PostgreSQL
                │                          └─→ Redis
                │
                └──→ Admin-UI Container
                └──→ Landing Container
```

### Monitoring Flow

```
All Containers → Metrics → Prometheus → Grafana → Dashboard
                                           │
                                           └─→ Alerts
```

---

## File System Layout

### Production Server

```
/opt/llm-proxy/
├── deployments/docker/
│   ├── .env                          # SECRET! Contains passwords
│   ├── docker-compose.registry-deploy.yml
│   ├── prometheus.yml
│   └── grafana/
│       ├── provisioning/
│       └── dashboards/
├── scripts/deployment/
│   └── server-deploy.sh              # Deployment automation
└── /volumes/ (Docker volumes)
    ├── postgres_data/                # Persistent DB data
    ├── redis_data/                   # Persistent cache
    ├── prometheus_data/              # Metrics history
    └── grafana_data/                 # Dashboards

/opt/llm-proxy-backups/
└── 20260320/
    ├── postgres-backup-120530.sql    # Daily backups
    ├── images-before-120530.txt      # Image versions
    └── health-before-120530.txt      # Container health
```

---

## Summary

### ✅ Benefits of This Architecture

- **Security:** Multiple firewall layers, localhost-only bindings
- **Scalability:** Easy to add containers/replicas
- **Reliability:** Automated backups, health checks, rollbacks
- **Simplicity:** No source code on server, just configs
- **Speed:** Pull pre-built images, no compilation needed
- **Consistency:** Same images in dev/staging/prod

### 🔒 Security Features

- All database/cache ports on 127.0.0.1
- DigitalOcean + UFW firewall layers
- HTTPS only via Caddy
- No source code on production server
- Secrets in .env (never committed)
- Regular automated backups

### 🚀 Deployment Features

- Zero-downtime deployments
- Automated database backups
- Health checks before/after deployment
- Easy rollbacks to any version
- Version-tagged images
- Automated cleanup of old images

---

**Last Updated:** 2026-03-20  
**Architecture Version:** 2.0
