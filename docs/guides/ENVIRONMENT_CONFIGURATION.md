# Environment Configuration Guide

## Overview

This project uses a **clean and professional** environment configuration system that eliminates the common problem of hardcoded URLs and environment-specific builds.

## Key Principle: Runtime Detection

The Admin UI frontend uses `window.location.origin` to automatically detect the API URL at **runtime**, not build-time. This means:

- ✅ **Same Docker image works everywhere** (localhost, staging, production)
- ✅ **No configuration needed** for different environments
- ✅ **No hardcoded URLs** in the codebase
- ✅ **No build-time environment variables** affecting runtime behavior

## How It Works

### Frontend (Admin UI)

**File:** `admin-ui/src/lib/api.js`
```javascript
// Automatically detects the current domain
const API_BASE_URL = window.location.origin;

// Examples:
// - On http://localhost:3005 → API_BASE_URL = "http://localhost:3005"
// - On https://scrubgate.tech → API_BASE_URL = "https://scrubgate.tech"
```

This approach:
- Works in **any environment** without configuration
- Eliminates the need for `VITE_API_BASE_URL` environment variable
- Prevents localhost:8080 hardcoding issues

### Backend (Go)

Backend uses traditional environment variables via `.env` files:

| Environment | File | Location | Usage |
|------------|------|----------|-------|
| **Local Docker Dev** | `.env.local` | Project root | `docker-compose.dev.yml` |
| **Production** | GitLab CI/CD Variable | GitLab | `docker-compose.openwebui.yml` |

## Environment Files Structure

```
llm-proxy/
├── .env.local                    # Local Docker development (gitignored)
├── .env.example                  # Template for .env.local
├── admin-ui/
│   ├── .env.development          # Vite development mode (committed)
│   ├── .env.production           # Vite production mode (committed)
│   └── .env.example              # Documentation (committed)
```

### What Gets Committed?

✅ **Committed (templates & defaults):**
- `.env.example` (project root)
- `admin-ui/.env.example`
- `admin-ui/.env.development`
- `admin-ui/.env.production`

❌ **NOT Committed (contains secrets or environment-specific values):**
- `.env` (any .env without extension)
- `.env.local`
- `.env.*.local`

## Development Modes

### 1. Docker Development (Recommended)

**Start:**
```bash
make dev-docker-up
```

**Environment:**
- Backend: Uses `.env.local` for database passwords, API keys
- Frontend: Uses `window.location.origin` (no .env needed!)
- Database: PostgreSQL on localhost:5433
- Services communicate via Docker network

**Access:**
- Frontend: http://localhost:3005
- Backend: http://localhost:8080
- API calls: Automatically go to http://localhost:3005 → proxied to backend

### 2. Native Development (Advanced)

**Start:**
```bash
# Terminal 1: Start infrastructure
make docker-up

# Terminal 2: Start backend
make dev

# Terminal 3: Start frontend
cd admin-ui && npm run dev
```

**Environment:**
- Backend: Needs `.env` with `DB_HOST=localhost` (not Docker service name)
- Frontend: Runs on :5173, needs to call backend on :8080

**Special Case:** Create `admin-ui/.env.local` (gitignored) if frontend port ≠ backend port:
```bash
# admin-ui/.env.local
VITE_API_BASE_URL=http://localhost:8080
```

This is the **ONLY** case where you need `VITE_API_BASE_URL`!

## Production Deployment

### GitLab CI/CD

**Environment Variables (Set in GitLab):**
```
ENV_FILE        = Complete .env content (File Variable)
ADMIN_API_KEY   = Admin API key
ENVIRONMENT     = production
```

**How It Works:**
1. CI/CD creates `.env` from `ENV_FILE` variable
2. Backend reads `.env` for database, Redis, secrets
3. Admin UI is built with Vite (production mode)
4. Frontend uses `window.location.origin` at runtime
5. Same image deployed → works on any domain!

**No Need For:**
- ❌ VITE_API_BASE_URL in CI/CD
- ❌ Different images for different environments
- ❌ Hardcoded URLs

## Troubleshooting

### Problem: Frontend can't reach backend

**Symptom:** API calls fail with CORS errors or connection refused

**Diagnosis:**
```javascript
// Check in browser console:
console.log(window.location.origin);  // Should match where backend is
```

**Solutions:**

1. **Docker Development:**
   - Frontend should be on same domain as backend
   - Example: Both on http://localhost:3005 (Caddy proxies /api/* to backend)

2. **Native Development (different ports):**
   - Create `admin-ui/.env.local`:
     ```bash
     VITE_API_BASE_URL=http://localhost:8080
     ```
   - Restart Vite dev server: `npm run dev`

### Problem: Getting localhost:8080 in production

**Cause:** Old .env file with hardcoded URL was committed and used at build time

**Solution:**
1. Delete `admin-ui/.env` (should be gitignored)
2. Rebuild Docker image
3. Verify: `docker-compose.dev.yml` and `docker-compose.openwebui.yml` don't set VITE_API_BASE_URL

### Problem: Environment variables not loaded

**Docker Development:**
```bash
# Check if .env.local exists
ls -la .env.local

# Check if container sees it
docker exec llm-proxy-backend-dev env | grep DB_PASSWORD
```

**Native Development:**
```bash
# Check if .env exists
ls -la .env

# Backend should log loaded config on startup
make dev
```

## Best Practices

### DO ✅

- Use `window.location.origin` for runtime URL detection
- Use `.env.local` for local development secrets (gitignored)
- Use GitLab CI/CD Variables for production secrets
- Commit `.env.example` templates
- Document required environment variables

### DON'T ❌

- Hardcode URLs in source code
- Commit `.env` files with secrets or environment-specific values
- Use `VITE_API_BASE_URL` unless absolutely necessary (native dev with different ports)
- Create environment-specific Docker images
- Rely on build-time environment variables for runtime behavior

## Environment Variable Reference

### Backend (.env.local / GitLab CI/CD)

```bash
# Required
DB_HOST=postgres                          # Database host
DB_PASSWORD=dev_password_2024_local       # Database password
REDIS_HOST=redis                          # Redis host
ADMIN_API_KEY=admin_dev_key_12345...      # Admin API key
OAUTH_JWT_SECRET=jwt-secret-min-32-chars  # JWT signing secret

# Optional
CLAUDE_API_KEY=sk-ant-api03-...           # Claude API key
OPENAI_API_KEY=sk-proj-...                # OpenAI API key
LOG_LEVEL=debug                           # Logging level
CACHE_ENABLED=true                        # Enable caching
```

### Frontend (admin-ui/.env.local - ONLY for native dev)

```bash
# ONLY needed if frontend port ≠ backend port
VITE_API_BASE_URL=http://localhost:8080
```

## Migration from Old System

If you have old .env files with hardcoded URLs:

```bash
# 1. Remove problematic files
rm admin-ui/.env

# 2. Update .gitignore (already done)
# admin-ui/.gitignore includes .env

# 3. Rebuild Docker images
docker compose -f docker-compose.dev.yml up --build

# 4. Verify window.location.origin is used
# Check browser console or inspect admin-ui/src/lib/api.js
```

## Summary

| Aspect | Old System (Bad) | New System (Good) |
|--------|------------------|-------------------|
| **API URL Detection** | Build-time (`import.meta.env`) | Runtime (`window.location.origin`) |
| **Configuration** | Hardcoded in .env | Automatic detection |
| **Docker Images** | Different per environment | Same image everywhere |
| **Deployment** | Brittle, error-prone | Robust, zero-config |
| **Development** | Complex setup | Simple, Docker-first |

The new system is **production-ready**, **maintainable**, and **eliminates entire classes of bugs**.
