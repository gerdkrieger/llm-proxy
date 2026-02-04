# LLM-Proxy Production Deployment Guide

## 📋 Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Configuration](#configuration)
- [Deployment Methods](#deployment-methods)
- [Service Architecture](#service-architecture)
- [Monitoring & Logging](#monitoring--logging)
- [Backup & Recovery](#backup--recovery)
- [Troubleshooting](#troubleshooting)
- [Security Best Practices](#security-best-practices)

---

## Overview

This guide covers deploying the LLM-Proxy system in production using Docker Compose. The deployment includes:

- **Backend API** (Go application)
- **Admin UI** (Svelte + Nginx)
- **PostgreSQL** (Database)
- **Redis** (Caching)
- **Prometheus** (Metrics)
- **Grafana** (Dashboards)

---

## Prerequisites

### Required Software

- **Docker** 20.10+ ([Install Docker](https://docs.docker.com/get-docker/))
- **Docker Compose** 2.0+ (comes with Docker Desktop)
- **Git** (for cloning the repository)
- **Make** (optional, for convenience commands)

### System Requirements

**Minimum:**
- CPU: 2 cores
- RAM: 4GB
- Disk: 10GB SSD

**Recommended:**
- CPU: 4+ cores
- RAM: 8GB+
- Disk: 20GB+ SSD

### Network Requirements

**Ports to expose:**
- `8080` - Backend API
- `3005` - Admin UI
- `9090` - Prometheus
- `3001` - Grafana
- `5433` - PostgreSQL (optional, for external access)
- `6380` - Redis (optional, for external access)

---

## Quick Start

### 1. Clone Repository

```bash
git clone <repository-url>
cd llm-proxy
```

### 2. Configure Environment

```bash
# Copy example configuration
cp .env.production.example .env

# Edit configuration (IMPORTANT!)
nano .env
```

**Required changes in `.env`:**
- `DB_PASSWORD` - Strong PostgreSQL password
- `OAUTH_JWT_SECRET` - Strong JWT secret (32+ chars)
- `ADMIN_API_KEY` - Strong API key (32+ chars)
- `CLAUDE_API_KEY` - Your Claude API key
- `GRAFANA_ADMIN_PASSWORD` - Strong Grafana password
- `CORS_ALLOWED_ORIGINS` - Your domain(s)

**Generate strong secrets:**
```bash
# JWT Secret (32+ chars)
openssl rand -base64 32

# Admin API Key (64 hex chars)
openssl rand -hex 32

# Database Password
openssl rand -base64 24

# Grafana Password
openssl rand -base64 16
```

### 3. Deploy with Script (Recommended)

```bash
cd deployments/docker
chmod +x deploy.sh
./deploy.sh deploy
```

**Or using Makefile:**
```bash
make deploy-prod
```

### 4. Verify Deployment

```bash
# Check service health
./deploy.sh health

# Or with Make
make deploy-health
```

**Access the services:**
- Backend API: http://localhost:8080
- Admin UI: http://localhost:3005
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3001

---

## Configuration

### Environment Variables

Complete list of configuration options in `.env`:

#### Application Settings
```bash
ENVIRONMENT=production
SERVER_PORT=8080
METRICS_PORT=9091
ADMIN_UI_PORT=3005
```

#### Database (PostgreSQL)
```bash
DB_HOST=postgres
DB_PORT=5432
DB_NAME=llm_proxy
DB_USER=proxy_user
DB_PASSWORD=CHANGE_ME
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m
```

#### Redis
```bash
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

#### OAuth / JWT
```bash
OAUTH_JWT_SECRET=CHANGE_ME_32_CHARS_MIN
OAUTH_ACCESS_TOKEN_EXPIRY=3600
OAUTH_REFRESH_TOKEN_EXPIRY=604800
```

#### Admin API
```bash
ADMIN_API_KEY=CHANGE_ME_STRONG_KEY
```

#### Claude API
```bash
CLAUDE_API_KEY=sk-ant-api03-your-key
CLAUDE_MAX_RETRIES=3
CLAUDE_RETRY_DELAY=1000
CLAUDE_TIMEOUT=60000
```

#### Caching
```bash
CACHE_ENABLED=true
CACHE_DEFAULT_TTL=3600
CACHE_MAX_SIZE=1000
```

#### Logging
```bash
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout
```

#### CORS
```bash
CORS_ALLOWED_ORIGINS=http://localhost:3005,https://your-domain.com
CORS_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Authorization,Content-Type,X-Requested-With
CORS_MAX_AGE=3600
```

---

## Deployment Methods

### Method 1: Deployment Script (Recommended)

The `deploy.sh` script provides an interactive menu for all deployment operations.

```bash
cd deployments/docker
./deploy.sh
```

**Available commands:**
```bash
./deploy.sh deploy     # Full deployment
./deploy.sh update     # Update deployment
./deploy.sh start      # Start services
./deploy.sh stop       # Stop services
./deploy.sh restart    # Restart services
./deploy.sh status     # Show status
./deploy.sh logs       # Show logs
./deploy.sh health     # Health check
./deploy.sh backup     # Backup data
./deploy.sh clean      # Clean up
```

### Method 2: Makefile Commands

```bash
make deploy-prod       # Full deployment
make deploy-update     # Update deployment
make deploy-start      # Start services
make deploy-stop       # Stop services
make deploy-status     # Show status
make deploy-logs       # Show logs
make deploy-health     # Health check
make deploy-backup     # Backup data
make deploy-clean      # Clean up
```

### Method 3: Docker Compose Directly

```bash
cd deployments/docker

# Build and start
docker compose -f docker-compose.prod.yml up -d --build

# Stop
docker compose -f docker-compose.prod.yml down

# View logs
docker compose -f docker-compose.prod.yml logs -f

# Show status
docker compose -f docker-compose.prod.yml ps
```

---

## Service Architecture

### Container Overview

| Service | Container Name | Ports | Description |
|---------|---------------|-------|-------------|
| Backend | llm-proxy-backend | 8080, 9091 | Go API server |
| Admin UI | llm-proxy-admin-ui | 3005 | Svelte + Nginx |
| PostgreSQL | llm-proxy-postgres | 5433 | Database |
| Redis | llm-proxy-redis | 6380 | Cache |
| Prometheus | llm-proxy-prometheus | 9090 | Metrics |
| Grafana | llm-proxy-grafana | 3001 | Dashboards |

### Network Architecture

```
┌─────────────────────────────────────────────────┐
│  llm-proxy-network (Bridge Network)             │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐      │
│  │ Backend  │  │ Admin UI │  │Prometheus│      │
│  │  :8080   │  │  :3005   │  │  :9090   │      │
│  └────┬─────┘  └─────────┘  └────┬─────┘      │
│       │                            │            │
│  ┌────┴─────┐  ┌──────────┐  ┌────┴─────┐      │
│  │PostgreSQL│  │  Redis   │  │ Grafana  │      │
│  │  :5433   │  │  :6380   │  │  :3001   │      │
│  └──────────┘  └──────────┘  └──────────┘      │
└─────────────────────────────────────────────────┘
```

### Volume Mounts

Persistent data is stored in Docker volumes:

- `llm-proxy-postgres-data` - PostgreSQL database
- `llm-proxy-redis-data` - Redis cache
- `llm-proxy-prometheus-data` - Prometheus metrics
- `llm-proxy-grafana-data` - Grafana dashboards

**View volumes:**
```bash
docker volume ls | grep llm-proxy
```

**Inspect volume:**
```bash
docker volume inspect llm-proxy-postgres-data
```

---

## Monitoring & Logging

### Prometheus Metrics

**Access:** http://localhost:9090

**Available metrics:**
- Request count and duration
- Token usage (input/output)
- Cost tracking
- Cache hit/miss ratio
- Error rates
- Provider health status

**Example queries:**
```promql
# Request rate
rate(http_requests_total[5m])

# Cache hit rate
rate(cache_hits_total[5m]) / rate(cache_requests_total[5m])

# Average response time
rate(http_request_duration_seconds_sum[5m]) / rate(http_request_duration_seconds_count[5m])
```

### Grafana Dashboards

**Access:** http://localhost:3001

**Default credentials:**
- Username: `admin`
- Password: (set in `.env` as `GRAFANA_ADMIN_PASSWORD`)

**Pre-configured dashboards:**
- System Overview
- API Performance
- Cache Statistics
- Provider Health
- Cost Tracking

### Application Logs

**View logs:**
```bash
# All services
docker compose -f docker-compose.prod.yml logs -f

# Specific service
docker compose -f docker-compose.prod.yml logs -f backend

# Last 100 lines
docker compose -f docker-compose.prod.yml logs --tail=100 backend
```

**Log format:**
- Development: Console (human-readable)
- Production: JSON (for log aggregation)

**Log levels:**
- `trace` - Very detailed debugging
- `debug` - Debugging information
- `info` - General information (default)
- `warn` - Warning messages
- `error` - Error messages
- `fatal` - Fatal errors (application crashes)

---

## Backup & Recovery

### Automated Backup

**Run backup:**
```bash
./deploy.sh backup
# Or
make deploy-backup
```

**Backup location:**
```
backups/YYYYMMDD_HHMMSS/
├── postgres_backup.sql
└── redis_backup.rdb
```

### Manual Backup

**PostgreSQL:**
```bash
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > backup.sql
```

**Redis:**
```bash
docker exec llm-proxy-redis redis-cli SAVE
docker cp llm-proxy-redis:/data/dump.rdb backup.rdb
```

### Restore from Backup

**PostgreSQL:**
```bash
# Stop backend to prevent connections
docker compose -f docker-compose.prod.yml stop backend

# Restore database
docker exec -i llm-proxy-postgres psql -U proxy_user llm_proxy < backup.sql

# Start backend
docker compose -f docker-compose.prod.yml start backend
```

**Redis:**
```bash
# Stop Redis
docker compose -f docker-compose.prod.yml stop redis

# Copy backup
docker cp backup.rdb llm-proxy-redis:/data/dump.rdb

# Start Redis
docker compose -f docker-compose.prod.yml start redis
```

### Backup Strategy

**Recommended schedule:**
- Daily automated backups
- Weekly full backups to external storage
- Monthly archival backups

**Retention policy:**
- Daily: Keep 7 days
- Weekly: Keep 4 weeks
- Monthly: Keep 12 months

---

## Troubleshooting

### Common Issues

#### Services won't start

**Check Docker status:**
```bash
docker ps -a
```

**Check logs:**
```bash
docker compose -f docker-compose.prod.yml logs backend
```

**Check environment variables:**
```bash
docker compose -f docker-compose.prod.yml config
```

#### Database connection errors

**Verify PostgreSQL is healthy:**
```bash
docker exec llm-proxy-postgres pg_isready -U proxy_user
```

**Check connection from backend:**
```bash
docker exec llm-proxy-backend psql -h postgres -U proxy_user -d llm_proxy
```

**Reset database:**
```bash
docker compose -f docker-compose.prod.yml restart postgres
```

#### Cache not working

**Check Redis status:**
```bash
docker exec llm-proxy-redis redis-cli ping
```

**Clear cache:**
```bash
docker exec llm-proxy-redis redis-cli FLUSHALL
```

#### High memory usage

**Check resource usage:**
```bash
docker stats
```

**Adjust resource limits in `docker-compose.prod.yml`:**
```yaml
deploy:
  resources:
    limits:
      memory: 2G  # Increase if needed
```

#### Backend API not responding

**Check health endpoint:**
```bash
curl http://localhost:8080/health
```

**Restart backend:**
```bash
docker compose -f docker-compose.prod.yml restart backend
```

**Check environment variables:**
```bash
docker exec llm-proxy-backend env | grep -E "DB|REDIS|CLAUDE"
```

### Performance Optimization

#### Increase cache size
```bash
# In .env
CACHE_MAX_SIZE=5000
CACHE_DEFAULT_TTL=7200
```

#### Scale backend horizontally
```bash
docker compose -f docker-compose.prod.yml up -d --scale backend=3
```

#### Optimize database connections
```bash
# In .env
DB_MAX_OPEN_CONNS=50
DB_MAX_IDLE_CONNS=10
```

---

## Security Best Practices

### 1. Secrets Management

**Never commit secrets to Git:**
```bash
# Add to .gitignore
.env
.env.production
*.key
*.pem
```

**Use strong secrets:**
- Minimum 32 characters for JWT secrets
- Use cryptographically secure random generation
- Rotate secrets regularly (quarterly)

**Environment-based secrets:**
```bash
# For cloud deployments, use secret managers
# AWS: AWS Secrets Manager
# GCP: Google Secret Manager
# Azure: Azure Key Vault
```

### 2. Network Security

**Use reverse proxy (Nginx/Traefik):**
```nginx
# Nginx config example
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

**Configure firewall:**
```bash
# Allow only necessary ports
ufw allow 443/tcp  # HTTPS
ufw allow 22/tcp   # SSH
ufw enable
```

**Update CORS settings:**
```bash
# In .env - restrict to your domain
CORS_ALLOWED_ORIGINS=https://yourdomain.com
```

### 3. Container Security

**Run as non-root:**
```yaml
# Already configured in Dockerfile
USER llmproxy
```

**Scan images for vulnerabilities:**
```bash
docker scan llm-proxy-backend:latest
```

**Keep images updated:**
```bash
docker compose -f docker-compose.prod.yml pull
docker compose -f docker-compose.prod.yml up -d
```

### 4. Database Security

**Use strong passwords:**
```bash
DB_PASSWORD=$(openssl rand -base64 24)
```

**Restrict network access:**
```yaml
# In docker-compose.prod.yml - remove public port mapping
# ports:
#   - "5433:5432"  # Comment this out
```

**Regular backups:**
```bash
# Automate with cron
0 2 * * * /path/to/llm-proxy/deployments/docker/deploy.sh backup
```

### 5. API Security

**Rate limiting:**
```bash
# In .env
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

**API key rotation:**
```bash
# Generate new admin API key
openssl rand -hex 32

# Update .env and restart
docker compose -f docker-compose.prod.yml restart backend
```

**Monitor for abuse:**
- Check Grafana dashboards regularly
- Set up alerts for unusual activity
- Review access logs weekly

### 6. SSL/TLS

**Use Let's Encrypt for free SSL:**
```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Generate certificate
sudo certbot --nginx -d api.yourdomain.com
```

**Update admin UI to use HTTPS:**
```bash
# In .env
VITE_API_BASE_URL=https://api.yourdomain.com
```

---

## Production Checklist

Before going live, verify:

- [ ] All secrets changed from defaults
- [ ] SSL/TLS certificates configured
- [ ] Firewall rules configured
- [ ] Backup automation set up
- [ ] Monitoring dashboards configured
- [ ] Alert rules set up in Grafana
- [ ] Log aggregation configured
- [ ] Domain DNS configured
- [ ] CORS origins restricted to your domain
- [ ] Rate limiting enabled
- [ ] Resource limits appropriate for load
- [ ] Health checks passing
- [ ] Load testing completed
- [ ] Documentation updated

---

## Support & Updates

### Updating the System

**Update to latest version:**
```bash
# Pull latest code
git pull origin main

# Rebuild and restart
./deploy.sh update
```

**Apply database migrations:**
```bash
make migrate-up
```

**Rollback if needed:**
```bash
docker compose -f docker-compose.prod.yml down
docker compose -f docker-compose.prod.yml up -d
```

### Getting Help

- **Documentation:** See README.md, TESTING.md, ADMIN_API.md
- **Logs:** Check `docker compose logs` for errors
- **Health checks:** Use `./deploy.sh health`
- **GitHub Issues:** Report bugs and feature requests

---

## License

[Add your license information here]

---

**Last Updated:** January 29, 2026
**Version:** 1.0.0
