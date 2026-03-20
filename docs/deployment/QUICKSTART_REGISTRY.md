# 🚀 Quick Start: Registry-Based Deployment

## TL;DR

```bash
# 1. Login to registries (one-time)
make login-registries

# 2. Build, push, and deploy
make release VERSION=v1.0.0

# Done! ✅
```

---

## Prerequisites

- Docker installed locally
- SSH access to production server (`openweb`)
- GitHub account (gerdkrieger)
- GitLab account (krieger-engineering)

---

## First-Time Setup

### 1. Login to Container Registries

```bash
# On your local machine
make login-registries
```

This will prompt for:
- GitHub Personal Access Token (needs `write:packages` scope)
- GitLab Personal Access Token or Deploy Token

### 2. Setup Production Server

```bash
# SSH to server
ssh openweb

# Create directories
mkdir -p /opt/llm-proxy/{scripts/deployment,deployments/docker}
mkdir -p /opt/llm-proxy-backups

# Create Docker resources
docker network create llm-proxy-network
docker volume create llm-proxy-postgres-data
docker volume create llm-proxy-redis-data
docker volume create llm-proxy-prometheus-data
docker volume create llm-proxy-grafana-data

# Login to registries (on server)
docker login ghcr.io -u gerdkrieger
docker login registry.gitlab.com

# Exit server
exit
```

### 3. Create .env File on Server

```bash
# Copy .env template to server
scp deployments/docker/.env.example openweb:/opt/llm-proxy/deployments/docker/.env

# Edit .env on server
ssh openweb "nano /opt/llm-proxy/deployments/docker/.env"
```

**Required .env variables:**
```bash
# Database
DB_PASSWORD=<STRONG_PASSWORD_HERE>
DB_NAME=llm_proxy
DB_USER=proxy_user

# Backend
OAUTH_JWT_SECRET=<RANDOM_SECRET_32_CHARS>
ADMIN_API_KEYS=<YOUR_ADMIN_KEY>

# API Keys
CLAUDE_API_KEY=<YOUR_CLAUDE_KEY>
OPENAI_API_KEY=<YOUR_OPENAI_KEY>

# Optional
GRAFANA_ADMIN_PASSWORD=<GRAFANA_PASSWORD>
```

---

## Daily Workflow

### Deploy New Version

```bash
# 1. Make code changes
git add .
git commit -m "Feature: Add new functionality"
git push

# 2. Build, push, and deploy in one command
make release VERSION=v1.2.0
```

That's it! The command will:
- ✅ Build all Docker images
- ✅ Push to GitHub/GitLab registries
- ✅ Copy deployment files to server
- ✅ Pull images on server
- ✅ Backup database
- ✅ Deploy with zero downtime
- ✅ Health check all containers

---

## Quick Operations

### Check Production Status

```bash
make status
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

### Rollback

```bash
# Rollback to previous version
make rollback VERSION=v1.1.0
```

### Backup Database

```bash
make backup-db
```

---

## Troubleshooting

### "Not logged in to registry"

```bash
make login-registries
```

### "Container unhealthy after deployment"

```bash
# Check logs
make logs SERVICE=backend

# Rollback
make rollback VERSION=<previous-version>
```

### "Cannot connect to Docker daemon"

```bash
# Start Docker
sudo systemctl start docker

# Check status
docker info
```

---

## File Locations

### Local Machine
```
llm-proxy/
├── scripts/deployment/build-and-push.sh  # Build & push script
├── deployments/docker/
│   └── docker-compose.registry-deploy.yml # Registry-based compose
├── Makefile.registry                      # Make commands
└── docs/deployment/REGISTRY_DEPLOYMENT.md # Full documentation
```

### Production Server
```
/opt/llm-proxy/
├── deployments/docker/
│   ├── .env                               # Configuration (SECRET!)
│   ├── docker-compose.registry-deploy.yml # Deployment config
│   └── prometheus.yml                     # Metrics config
└── scripts/deployment/
    └── server-deploy.sh                   # Server-side deployment
```

---

## Next Steps

1. **Read full documentation:** `docs/deployment/REGISTRY_DEPLOYMENT.md`
2. **Setup CI/CD (optional):** Automate builds on git push
3. **Setup monitoring:** Grafana dashboards
4. **Setup alerts:** Notify on unhealthy containers

---

## Support

- Full docs: `docs/deployment/REGISTRY_DEPLOYMENT.md`
- Commands: `make help`
- Logs: `make logs SERVICE=<name>`
- Status: `make status`

---

**Happy Deploying! 🚀**
