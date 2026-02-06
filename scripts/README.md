# LLM-Proxy Scripts

Utility scripts for development, deployment, and maintenance.

## 📁 Directory Structure

```
scripts/
├── dev-entrypoint.sh                    # Docker dev container entrypoint
├── maintenance/
│   └── git-update.sh                    # Git workflow automation
└── setup/
    └── create-example-filters.sh        # Create example content filters
```

---

## 🚀 Active Scripts (3 total)

### Development

#### `dev-entrypoint.sh`
Docker container entrypoint for development environment.

**Usage:** Automatically executed by Docker Compose
```bash
# Used in docker-compose.dev.yml
# No manual execution needed
```

**What it does:**
- Waits for PostgreSQL to be ready
- Runs database migrations
- Starts the application

---

### Maintenance

#### `maintenance/git-update.sh`
Optimized Git workflow script for daily development.

**Usage:**
```bash
# Quick Mode (commit only on develop)
./scripts/maintenance/git-update.sh "feat: Add feature"

# Standard Mode (commit + merge master + push) - RECOMMENDED
./scripts/maintenance/git-update.sh -m "fix: Bug fix"

# Release Mode (like Standard + Release Tag)
./scripts/maintenance/git-update.sh -r "feat: Major release"

# With Tests
./scripts/maintenance/git-update.sh --test -m "refactor: Code"

# With Build
./scripts/maintenance/git-update.sh --build -m "feat: Feature"

# Interactive Mode
./scripts/maintenance/git-update.sh
```

**Modes:**
- **Quick:** Commit on develop only
- **Standard:** Commit + merge to master + push both branches (default)
- **Release:** Like Standard + creates version tag

**Options:**
- `--test` - Run tests before commit
- `--build` - Run build before commit
- `--no-push` - Skip automatic push to remote

**See also:** `docs/GIT_WORKFLOW.md`

---

### Setup

#### `setup/create-example-filters.sh`
Creates example content filters for demonstration and testing.

**Usage:**
```bash
# Default (local development)
./scripts/setup/create-example-filters.sh

# Custom URL and API key
BASE_URL=https://llmproxy.aitrail.ch \
ADMIN_KEY=your_api_key \
./scripts/setup/create-example-filters.sh
```

**Creates 8 example filters:**
1. `badword` → `[GEFILTERT]` (word)
2. `damn` → `[*]` (word)
3. `confidential information` → `[VERTRAULICH_ENTFERNT]` (phrase)
4. `Project Phoenix` → `[INTERNES_PROJEKT]` (phrase)
5. Email addresses → `[EMAIL_ENTFERNT]` (regex)
6. Phone numbers → `[TELEFON_ENTFERNT]` (regex)
7. Credit cards → `[KREDITKARTE_ENTFERNT]` (regex)
8. `CompetitorX` → `[KONKURRENT]` (word)

**Environment Variables:**
- `BASE_URL` - API endpoint (default: http://localhost:8080)
- `ADMIN_KEY` - Admin API key (default: local dev key)

---

## 🗑️ Removed Scripts (Cleanup 2026-02-06)

The following scripts were removed as they are no longer needed:

### Development Scripts (replaced by Docker Compose)
- ❌ `start-all.sh` → Use `docker compose -f docker-compose.dev.yml up`
- ❌ `start-dev.sh` → Use `docker compose -f docker-compose.dev.yml up`
- ❌ `status.sh` → Use `docker compose ps`
- ❌ `stop-all.sh` → Use `docker compose down`

### Build Scripts (replaced by deploy.sh & CI/CD)
- ❌ `docker-build.sh` → Use `./deploy.sh` or GitLab CI/CD
- ❌ `install-redaction-tools.sh` → OCR support not actively used

### Maintenance Scripts (historical one-time fixes)
- ❌ `diagnose-live.sh` - Server setup (historical)
- ❌ `fix-live-database.sh` - One-time DB setup
- ❌ `fix-provider-models-*.sh` - Schema migrations (historical)
- ❌ `QUICK-FIX-DOCKER.sh` - Hotfix from Feb 4, 2026
- ❌ `rebuild-admin-ui.sh` - One-time fix
- ❌ `restart-server.sh` - Replaced by Docker
- ❌ `stop-server.sh` - Replaced by Docker

### Setup Scripts (one-time use)
- ❌ `update-caddy-config.sh` - Server setup (historical)

### Testing Scripts (replaced by Admin UI & CI/CD)
- ❌ `test_admin_api.sh` - Use Admin UI or manual curl
- ❌ `test_api.sh` - Use Admin UI or manual curl
- ❌ `test-all-filters.sh` - Use Admin UI
- ❌ `test-content-filters.sh` - Use Admin UI

**Result:** 86% reduction (22 → 3 scripts)

---

## 📚 Related Documentation

- **Git Workflow:** `docs/GIT_WORKFLOW.md`
- **Deployment:** `DEPLOYMENT.md` + `deploy.sh`
- **Development Setup:** `docker-compose.dev.yml`
- **CI/CD Pipeline:** `.gitlab-ci.yml`
- **Docker:** `Dockerfile`, `Dockerfile.dev`

---

## 🛠️ Common Tasks

### Development

```bash
# Start local development environment
docker compose -f docker-compose.dev.yml up -d

# View logs
docker compose -f docker-compose.dev.yml logs -f

# Stop services
docker compose -f docker-compose.dev.yml down

# Create example filters
./scripts/setup/create-example-filters.sh
```

### Git Workflow

```bash
# Daily development (commit + merge + push)
./scripts/maintenance/git-update.sh -m "feat: My feature"

# See all options
./scripts/maintenance/git-update.sh --help
```

### Deployment

```bash
# Deploy to production
./deploy.sh

# Manual deployment steps (if needed)
# See DEPLOYMENT.md for details
```

### Testing

```bash
# Run Go tests
go test ./...

# Test Admin API manually
curl -H "X-Admin-API-Key: your_key" http://localhost:8080/admin/filters

# Use Admin UI
open http://localhost:3005
```

---

## 🔧 Script Maintenance

When adding new scripts:

1. **Consider if it's really needed** - Can Docker Compose do it?
2. **Add to this README** - Document usage and purpose
3. **Make it executable** - `chmod +x script.sh`
4. **Add help text** - Include `--help` flag
5. **Use proper error handling** - `set -e` at minimum

When removing scripts:
1. **Document the replacement** - What should users do instead?
2. **Update this README** - Add to "Removed Scripts" section
3. **Commit with clear message** - Explain why it was removed

---

**Last Updated:** 2026-02-06  
**Maintainer:** LLM-Proxy Team
