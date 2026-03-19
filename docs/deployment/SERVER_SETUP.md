# Production Server Setup Guide

## 🎯 Goal: Clean Production Deployment

**Only 2 files should be on production server:**
1. `.env` (production configuration)
2. `docker-compose.yml` (pulls from registry)

**Everything else is managed by Docker!**

---

## 🚨 Current Problem

Server has ALL files from repository (~50 files):
- ❌ Source code (cmd/, internal/, pkg/)
- ❌ Build tools (Makefile, go.mod)
- ❌ Git data (.git/)
- ❌ Documentation (docs/)
- ❌ Secrets in plain text (OPENWEBUI_TOKEN_*.txt)

**This is a security risk and deployment anti-pattern!**

---

## ✅ Correct Setup

### Step 1: Cleanup Current Server

```bash
# On your LOCAL machine:
cd /home/krieger/Sites/golang-projekte/llm-proxy

# Copy cleanup script to server
scp scripts/cleanup-production-server.sh deploy@your-server:/tmp/

# SSH to server
ssh deploy@your-server

# Run cleanup script
chmod +x /tmp/cleanup-production-server.sh
/tmp/cleanup-production-server.sh
```

This will:
- ✅ Backup .env
- ✅ Stop containers
- ✅ Delete all unnecessary files
- ✅ Preserve Docker volumes
- ✅ Create clean directory

---

### Step 2: Setup Docker External Resources (One-Time)

```bash
# On production server:
# Create external network
docker network create llm-proxy-network

# Create external volumes
docker volume create llm-proxy-postgres-data
docker volume create llm-proxy-redis-data

# Verify
docker network ls | grep llm-proxy
docker volume ls | grep llm-proxy
```

---

### Step 3: Create Production docker-compose.yml

```bash
# On production server:
cd /opt/llm-proxy

# Create docker-compose.yml
nano docker-compose.yml
```

**Copy content from:**  
`deployments/docker-compose.production.yml`

Or copy from local machine:

```bash
# On your LOCAL machine:
scp deployments/docker-compose.production.yml deploy@your-server:/opt/llm-proxy/docker-compose.yml
```

---

### Step 4: Create .env File

```bash
# On production server:
cd /opt/llm-proxy

# If backup exists, restore it:
cp /opt/llm-proxy-backup/.env.XXXXXX .env

# Or create new one:
nano .env
```

**Minimal Production .env:**

```bash
# Database
DB_NAME=llm_proxy
DB_USER=proxy_user
DB_PASSWORD=YOUR_SECURE_PASSWORD_HERE

# Redis
REDIS_PORT=6380

# Server
SERVER_PORT=8080
ENVIRONMENT=production

# OAuth
OAUTH_JWT_SECRET=YOUR_RANDOM_64_CHAR_SECRET_HERE
OAUTH_ACCESS_TOKEN_TTL=1h
OAUTH_REFRESH_TOKEN_TTL=720h

# Admin
ADMIN_API_KEYS=YOUR_SECURE_ADMIN_KEY_HERE

# Providers (Optional)
ANTHROPIC_API_KEY=sk-ant-YOUR_KEY_HERE
OPENAI_API_KEY=sk-proj-YOUR_KEY_HERE

# Encryption
ENCRYPTION_KEY=YOUR_32_BYTE_HEX_KEY_HERE
```

---

### Step 5: Login to GitLab Registry

```bash
# On production server:
docker login registry.gitlab.com

# Username: your-gitlab-username
# Password: <gitlab-access-token>  # NOT your GitLab password!
```

**Create GitLab Access Token:**
1. GitLab → Settings → Access Tokens
2. Name: `production-server-pull`
3. Scopes: ✅ `read_registry`
4. Create token
5. Copy token (only shown once!)

---

### Step 6: Pull Images & Start

```bash
# On production server:
cd /opt/llm-proxy

# Pull images from registry
docker-compose pull

# Start services (detached)
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f backend
```

---

## 🔄 Deployment Workflow

### Update Application (New Version)

```bash
# On production server:
cd /opt/llm-proxy

# Pull new images
docker-compose pull

# Restart services (zero-downtime for stateless apps)
docker-compose up -d

# Check health
curl http://localhost:8080/health
docker-compose logs -f backend
```

**That's it!** No code copy, no builds, no complexity!

---

### Rollback to Previous Version

```bash
# Option 1: Use tagged version in docker-compose.yml
nano docker-compose.yml

# Change:
# image: registry.gitlab.com/.../backend:latest
# To:
# image: registry.gitlab.com/.../backend:v1.2.3

# Then:
docker-compose pull
docker-compose up -d

# Option 2: Use previous image ID
docker images | grep backend
docker tag <old-image-id> registry.gitlab.com/.../backend:latest
docker-compose up -d
```

---

## 📁 Final Server Structure

```bash
/opt/llm-proxy/
├── .env                    # ✅ Production config (ONLY config file!)
└── docker-compose.yml      # ✅ Production compose (registry-based)

/var/lib/docker/volumes/
├── llm-proxy-postgres-data/   # Persistent database data
└── llm-proxy-redis-data/      # Persistent cache data
```

**That's literally all!** No source code, no build tools, nothing else!

---

## 🔒 Security Verification

### Check Port Bindings

```bash
# Should show 127.0.0.1 only:
docker ps --format "table {{.Names}}\t{{.Ports}}"

# Expected output:
# llm-proxy-postgres    127.0.0.1:5433->5432/tcp ✅
# llm-proxy-redis       127.0.0.1:6380->6379/tcp ✅
# llm-proxy-backend     127.0.0.1:8080->8080/tcp ✅
# llm-proxy-admin-ui    127.0.0.1:3005->80/tcp ✅
```

### Check Listening Ports

```bash
# Should show 127.0.0.1 only:
sudo netstat -tuln | grep LISTEN

# Expected (all localhost):
# tcp  0  0  127.0.0.1:5433   0.0.0.0:*  LISTEN ✅
# tcp  0  0  127.0.0.1:6380   0.0.0.0:*  LISTEN ✅
# tcp  0  0  127.0.0.1:8080   0.0.0.0:*  LISTEN ✅
```

### Check File Count

```bash
# Should show ONLY 2 files:
cd /opt/llm-proxy
ls -la

# Expected:
# .env               ✅
# docker-compose.yml ✅
# (nothing else!)
```

---

## 🌐 Reverse Proxy Setup (Required for External Access)

### Option 1: Caddy (Recommended)

```bash
# Install Caddy
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# Create Caddyfile
sudo nano /etc/caddy/Caddyfile
```

**Caddyfile:**

```
llmproxy.example.com {
  reverse_proxy localhost:8080
  tls your-email@example.com
}

admin.llmproxy.example.com {
  reverse_proxy localhost:3005
  tls your-email@example.com
}
```

```bash
# Restart Caddy
sudo systemctl restart caddy
sudo systemctl status caddy
```

---

### Option 2: Nginx + Certbot

```bash
# Install Nginx + Certbot
sudo apt install -y nginx certbot python3-certbot-nginx

# Create config
sudo nano /etc/nginx/sites-available/llm-proxy
```

**Nginx Config:**

```nginx
server {
    listen 80;
    server_name llmproxy.example.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 80;
    server_name admin.llmproxy.example.com;
    
    location / {
        proxy_pass http://localhost:3005;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

```bash
# Enable config
sudo ln -s /etc/nginx/sites-available/llm-proxy /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx

# Get SSL certificates
sudo certbot --nginx -d llmproxy.example.com -d admin.llmproxy.example.com
```

---

## 🔍 Monitoring & Troubleshooting

### View Logs

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend

# Last 100 lines
docker-compose logs --tail=100 backend
```

### Check Health

```bash
# Backend health
curl http://localhost:8080/health

# Check if services are running
docker-compose ps

# Check resources
docker stats

# Check disk usage
docker system df
```

### Common Issues

**Issue: "Cannot connect to Docker daemon"**
```bash
# Add user to docker group
sudo usermod -aG docker $USER
# Logout and login again
```

**Issue: "Network not found"**
```bash
# Create external network
docker network create llm-proxy-network
```

**Issue: "Volume not found"**
```bash
# Create external volumes
docker volume create llm-proxy-postgres-data
docker volume create llm-proxy-redis-data
```

**Issue: "Images cannot be pulled"**
```bash
# Login to registry again
docker login registry.gitlab.com
```

---

## 📊 Before vs After Comparison

| Aspect | Before ❌ | After ✅ |
|--------|-----------|----------|
| **Files on Server** | ~50 files | 2 files |
| **Source Code** | ✅ Yes | ❌ No |
| **Secrets** | Plain text | Env only |
| **Build Tools** | ✅ Yes | ❌ No |
| **Git Data** | ✅ Yes | ❌ No |
| **Deployment** | rsync/scp | docker pull |
| **Update Time** | Minutes | Seconds |
| **Rollback** | Difficult | Easy |
| **Security** | 🔴 Low | 🟢 High |
| **Maintenance** | 🔴 Complex | 🟢 Simple |

---

## ✅ Checklist

Before going live, verify:

- [ ] Only 2 files on server: `.env` + `docker-compose.yml`
- [ ] All ports bound to `127.0.0.1`
- [ ] External volumes created
- [ ] External network created
- [ ] GitLab registry login works
- [ ] Images can be pulled
- [ ] Services start successfully
- [ ] Health checks pass
- [ ] Reverse proxy configured
- [ ] SSL certificates installed
- [ ] Firewall configured (ufw)
- [ ] Backups configured

---

## 📞 Support

- Documentation: `/docs/deployment/`
- Scripts: `/scripts/`
- Deployment files: `/deployments/`
- GitLab Registry: `registry.gitlab.com/krieger-engineering/llm-proxy`

---

**Next Step:** Run cleanup script and deploy cleanly!
