# GitLab CI/CD Pipeline Documentation

## 📋 Table of Contents

- [Overview](#overview)
- [Pipeline Stages](#pipeline-stages)
- [Setup Instructions](#setup-instructions)
- [CI/CD Variables](#cicd-variables)
- [Pipeline Workflows](#pipeline-workflows)
- [Deployment Process](#deployment-process)
- [Troubleshooting](#troubleshooting)
- [Best Practices](#best-practices)

---

## Overview

The LLM-Proxy project uses GitLab CI/CD for automated testing, building, and deployment. The pipeline includes:

- **Automated Testing** - Unit and integration tests
- **Code Quality** - Linting, formatting, security scans
- **Docker Images** - Automated builds and push to GitLab Container Registry
- **Multi-Environment Deployment** - Development, Staging, Production
- **Security Scanning** - Vulnerability checks and secret detection

### Pipeline File

All CI/CD configuration is in `.gitlab-ci.yml` at the project root.

---

## Pipeline Stages

### 1. **Lint** 🔍

Checks code quality and formatting.

| Job | Description | Runs On |
|-----|-------------|---------|
| `lint:backend` | Go linter (golangci-lint) | MRs, main, develop |
| `lint:admin-ui` | JavaScript/Svelte linter | MRs, main, develop |
| `format:check` | Go code formatting check | MRs, main |

**Duration:** ~2-3 minutes

### 2. **Test** 🧪

Runs automated tests.

| Job | Description | Runs On |
|-----|-------------|---------|
| `test:unit` | Go unit tests with coverage | MRs, main, develop |
| `test:integration` | Integration tests | MRs, main, develop |
| `test:admin-ui` | Admin UI tests | MRs, main, develop |

**Duration:** ~5-8 minutes  
**Services:** PostgreSQL 14, Redis 7

### 3. **Security** 🔒

Security and vulnerability scanning.

| Job | Description | Runs On |
|-----|-------------|---------|
| `security:go-dependencies` | Check for known vulnerabilities | MRs, main, develop |
| `security:npm-audit` | npm security audit | MRs, main, develop |
| `security:secrets-scan` | Scan for hardcoded secrets | MRs, main |

**Duration:** ~2-4 minutes

### 4. **Build** 🏗️

Compiles application binaries.

| Job | Description | Runs On | Artifacts |
|-----|-------------|---------|-----------|
| `build:backend` | Build Go binary | MRs, main, develop, tags | `bin/llm-proxy` |
| `build:admin-ui` | Build Svelte app | MRs, main, develop, tags | `admin-ui/dist/` |

**Duration:** ~3-5 minutes  
**Artifacts Retention:** 1 day

### 5. **Docker** 🐳

Builds and pushes Docker images to GitLab Container Registry.

| Job | Description | Runs On | Images |
|-----|-------------|---------|--------|
| `docker:backend` | Build backend Docker image | main, develop, tags | `registry.gitlab.com/.../backend` |
| `docker:admin-ui` | Build Admin UI Docker image | main, develop, tags | `registry.gitlab.com/.../admin-ui` |

**Duration:** ~5-10 minutes  
**Image Tags:**
- `<commit-sha>` - Always created
- `<branch-name>` - Branch-specific (e.g., `main`, `develop`)
- `latest` - Only for `main` branch
- `<tag>` - For Git tags (e.g., `v1.0.0`)

### 6. **Deploy** 🚀

Deploys to various environments.

| Job | Description | Runs On | When |
|-----|-------------|---------|------|
| `deploy:development` | Deploy to dev environment | develop | Manual |
| `deploy:staging` | Deploy to staging | main | Manual |
| `deploy:production` | Deploy to production | tags | Manual |

**Duration:** ~3-5 minutes per deployment

---

## Setup Instructions

### Step 1: Enable GitLab Container Registry

1. Go to your project on GitLab.com
2. Navigate to **Settings → General → Visibility**
3. Expand **Container Registry** section
4. Ensure **Container Registry** is enabled

### Step 2: Configure CI/CD Variables

All required variables must be set in **Settings → CI/CD → Variables**.

#### Quick Setup (Minimum Required)

```bash
# For deployments, you need:
SSH_PRIVATE_KEY          # SSH key for deployment access
DEV_HOST                 # Development server (e.g., dev.example.com)
DEV_USER                 # SSH username (e.g., deploy)
DEV_PATH                 # Project path (e.g., /opt/llm-proxy)
DEV_URL                  # Dev URL (e.g., https://dev.example.com)
```

See [`.gitlab/ci-variables.md`](.gitlab/ci-variables.md) for complete variable list.

### Step 3: Set Up Deployment Servers

On each deployment server (dev, staging, production):

#### 3.1. Install Prerequisites

```bash
# Install Docker and Docker Compose
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
```

#### 3.2. Create Deployment User

```bash
# Create deploy user
sudo useradd -m -s /bin/bash deploy
sudo usermod -aG docker deploy

# Add GitLab CI SSH key to authorized_keys
sudo su - deploy
mkdir -p ~/.ssh
chmod 700 ~/.ssh
# Paste your public key into ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

#### 3.3. Create Project Directory

```bash
# Create project directory
sudo mkdir -p /opt/llm-proxy
sudo chown deploy:deploy /opt/llm-proxy

# Clone repository (or copy files)
cd /opt/llm-proxy
git clone <your-repo-url> .

# Create .env file
cp .env.production.example .env
# Edit .env with production values
nano .env
```

### Step 4: Generate SSH Key for CI/CD

```bash
# On your local machine
ssh-keygen -t ed25519 -C "gitlab-ci-llm-proxy" -f ~/.ssh/gitlab_ci_llm_proxy -N ""

# This creates:
# - Private key: ~/.ssh/gitlab_ci_llm_proxy
# - Public key: ~/.ssh/gitlab_ci_llm_proxy.pub

# Copy public key to all deployment servers
ssh-copy-id -i ~/.ssh/gitlab_ci_llm_proxy.pub deploy@dev.example.com
ssh-copy-id -i ~/.ssh/gitlab_ci_llm_proxy.pub deploy@staging.example.com
ssh-copy-id -i ~/.ssh/gitlab_ci_llm_proxy.pub deploy@prod.example.com

# Add private key to GitLab CI/CD Variables
cat ~/.ssh/gitlab_ci_llm_proxy
# Copy entire output and add as SSH_PRIVATE_KEY variable in GitLab
```

### Step 5: Test the Pipeline

```bash
# Create a test commit
git checkout -b test-cicd
echo "# Testing CI/CD" >> README.md
git add README.md
git commit -m "test: CI/CD pipeline"
git push origin test-cicd

# Create merge request on GitLab
# Pipeline should start automatically
```

---

## CI/CD Variables

### Auto-Provided by GitLab

These are automatically available (no setup needed):

| Variable | Description |
|----------|-------------|
| `CI_REGISTRY` | GitLab Container Registry URL |
| `CI_REGISTRY_USER` | Registry username |
| `CI_REGISTRY_PASSWORD` | Registry password (token) |
| `CI_REGISTRY_IMAGE` | Full image path (e.g., `registry.gitlab.com/user/project`) |
| `CI_COMMIT_SHORT_SHA` | Short commit SHA (8 chars) |
| `CI_COMMIT_REF_SLUG` | Branch/tag name (URL-safe) |
| `CI_COMMIT_TAG` | Git tag (if tagged) |

### Required Variables

Configure these in **Settings → CI/CD → Variables**:

#### Development Environment
```
SSH_PRIVATE_KEY     # SSH private key for deployments (masked)
DEV_HOST            # dev.example.com
DEV_USER            # deploy
DEV_PATH            # /opt/llm-proxy
DEV_URL             # https://dev.example.com
```

#### Staging Environment
```
STAGING_HOST        # staging.example.com
STAGING_USER        # deploy
STAGING_PATH        # /opt/llm-proxy
STAGING_URL         # https://staging.example.com
```

#### Production Environment
```
PROD_HOST           # prod.example.com (protected)
PROD_USER           # deploy (protected)
PROD_PATH           # /opt/llm-proxy (protected)
PROD_URL            # https://api.example.com (protected)
```

**Protection:** Mark production variables as "Protected" to restrict access.

See complete list: [`.gitlab/ci-variables.md`](.gitlab/ci-variables.md)

---

## Pipeline Workflows

### Workflow 1: Feature Development (Merge Request)

```
1. Create feature branch
   git checkout -b feature/new-feature

2. Make changes, commit
   git add .
   git commit -m "feat: add new feature"

3. Push to GitLab
   git push origin feature/new-feature

4. Create Merge Request
   → Pipeline starts automatically

5. Pipeline runs:
   ✓ Lint (backend + admin-ui)
   ✓ Format check
   ✓ Unit tests
   ✓ Integration tests
   ✓ Security scans
   ✓ Build (backend + admin-ui)
   ✗ Docker (skipped - only on main/develop)
   ✗ Deploy (skipped - only on main/develop/tags)

6. Review pipeline results in MR

7. Merge when all checks pass
```

### Workflow 2: Development Deployment

```
1. Merge to develop branch
   → Pipeline starts automatically

2. Pipeline runs all stages:
   ✓ Lint
   ✓ Test
   ✓ Security
   ✓ Build
   ✓ Docker (builds and pushes images with 'develop' tag)
   ⏸️ Deploy (manual - requires approval)

3. Manually trigger deploy:development
   - Go to CI/CD → Pipelines
   - Find your pipeline
   - Click ▶️ on "deploy:development" job

4. Deployment script:
   - SSH to dev server
   - Pull latest Docker images
   - Restart services
   - Run health checks

5. Verify deployment at DEV_URL
```

### Workflow 3: Production Release

```
1. Merge to main branch
   → Staging pipeline starts

2. Pipeline builds images with 'main' and 'latest' tags

3. Manually deploy to staging:
   - Trigger "deploy:staging" job
   - Test thoroughly in staging environment

4. Create release tag when ready:
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0

5. Tag pipeline starts:
   - Builds images with 'v1.0.0' tag
   - Creates GitLab Release
   - deploy:production job available

6. Manually trigger production deployment:
   - Go to CI/CD → Pipelines
   - Find tag pipeline
   - Click ▶️ on "deploy:production"

7. Verify production deployment

8. Monitor with Grafana/Prometheus
```

---

## Deployment Process

### How Deployment Works

The deployment jobs use SSH to connect to target servers and execute:

1. **Login to GitLab Container Registry**
   ```bash
   docker login registry.gitlab.com
   ```

2. **Pull Latest Images**
   ```bash
   docker compose -f docker-compose.prod.yml pull
   ```

3. **Restart Services**
   ```bash
   docker compose -f docker-compose.prod.yml up -d
   ```

4. **Health Checks**
   ```bash
   curl -f http://localhost:8080/health
   ```

### Manual Deployment (Without CI/CD)

On the deployment server:

```bash
# 1. Login to registry
docker login registry.gitlab.com

# 2. Set environment variables
export CI_REGISTRY_IMAGE="registry.gitlab.com/youruser/llm-proxy"
export IMAGE_TAG="v1.0.0"  # or "latest", "main", etc.

# 3. Pull and start
cd /opt/llm-proxy
docker compose -f deployments/docker/docker-compose.registry.yml pull
docker compose -f deployments/docker/docker-compose.registry.yml up -d

# 4. Verify
curl http://localhost:8080/health
docker compose -f deployments/docker/docker-compose.registry.yml ps
```

### Using Deployment Script

```bash
cd /opt/llm-proxy
./deployments/scripts/deploy-from-registry.sh production v1.0.0
```

---

## Troubleshooting

### Issue: Pipeline Fails at Docker Build

**Symptoms:**
- `docker:backend` or `docker:admin-ui` job fails
- Error: "Cannot connect to Docker daemon"

**Solution:**
```bash
# Check if Container Registry is enabled
# Settings → General → Visibility → Container Registry

# Verify GitLab Runner has Docker-in-Docker (dind) capability
# Contact GitLab admin if using shared runners
```

### Issue: Tests Fail with Database Connection Error

**Symptoms:**
- `test:unit` or `test:integration` fails
- Error: "dial tcp: lookup postgres: no such host"

**Solution:**
```yaml
# In .gitlab-ci.yml, ensure services are configured:
services:
  - postgres:14-alpine
  - redis:7-alpine

# And DB_HOST uses service name:
variables:
  DB_HOST: postgres  # NOT localhost!
  REDIS_HOST: redis  # NOT localhost!
```

### Issue: Deployment Fails with SSH Error

**Symptoms:**
- `deploy:*` job fails
- Error: "Permission denied (publickey)"

**Solution:**
```bash
# 1. Verify SSH_PRIVATE_KEY variable format in GitLab:
# It should include BEGIN/END lines:
-----BEGIN OPENSSH PRIVATE KEY-----
...key content...
-----END OPENSSH PRIVATE KEY-----

# 2. Check public key is in authorized_keys on server:
ssh deploy@dev.example.com
cat ~/.ssh/authorized_keys  # Should contain your public key

# 3. Test SSH connection manually:
ssh -i ~/.ssh/gitlab_ci_llm_proxy deploy@dev.example.com
```

### Issue: Docker Images Not Found During Deployment

**Symptoms:**
- Deployment succeeds but services don't start
- Error: "Error response from daemon: manifest not found"

**Solution:**
```bash
# 1. Verify images were built and pushed:
# Go to Packages & Registries → Container Registry

# 2. Check image tag exists:
docker manifest inspect registry.gitlab.com/youruser/llm-proxy/backend:v1.0.0

# 3. Ensure deployment uses correct IMAGE_TAG:
export IMAGE_TAG="v1.0.0"  # Must match what was built

# 4. Login to registry on deployment server:
docker login registry.gitlab.com
```

### Issue: Health Check Fails After Deployment

**Symptoms:**
- Deployment script reports "Backend health check failed"
- Service is running but not responding

**Solution:**
```bash
# 1. Check container logs:
docker compose -f docker-compose.prod.yml logs backend

# 2. Check if service is actually running:
docker compose -f docker-compose.prod.yml ps

# 3. Check environment variables:
docker compose -f docker-compose.prod.yml exec backend env | grep -E "DB|REDIS|CLAUDE"

# 4. Check database connection:
docker compose -f docker-compose.prod.yml exec backend wget -q --tries=1 -O- http://localhost:8080/health

# 5. Verify .env file exists and is correct:
cat /opt/llm-proxy/.env
```

### Issue: Merge Request Pipeline Hangs

**Symptoms:**
- Pipeline shows "pending" forever
- Jobs don't start

**Solution:**
```bash
# Check GitLab Runner availability:
# Settings → CI/CD → Runners
# Ensure at least one runner is active and available

# If using shared runners:
# Settings → CI/CD → Runners → Enable shared runners

# Check runner tags match job requirements
# Most jobs don't require specific tags
```

---

## Best Practices

### 1. Branch Strategy

```
main (production)
 ├─ develop (staging)
 │   ├─ feature/feature-1
 │   ├─ feature/feature-2
 │   └─ bugfix/fix-1
 └─ hotfix/critical-fix
```

- **Feature branches** → Merge to `develop`
- **Develop** → Tested, merged to `main`
- **Main** → Tagged for production releases
- **Hotfixes** → Branch from `main`, merge to both `main` and `develop`

### 2. Commit Messages

Follow conventional commits:

```
feat: add new feature
fix: resolve bug in authentication
docs: update API documentation
test: add unit tests for OAuth service
refactor: improve cache implementation
chore: update dependencies
ci: update GitLab CI configuration
```

### 3. Versioning

Use semantic versioning for releases:

```
v1.0.0 - Major.Minor.Patch
 │ │ │
 │ │ └─ Patch: Bug fixes
 │ └─── Minor: New features (backward compatible)
 └───── Major: Breaking changes
```

Example:
```bash
# Create release tag
git tag -a v1.2.3 -m "Release v1.2.3: Add streaming support"
git push origin v1.2.3
```

### 4. Environment Management

| Environment | Branch/Tag | Purpose | Deployment |
|-------------|------------|---------|------------|
| Development | `develop` | Active development | Manual (frequent) |
| Staging | `main` | Pre-production testing | Manual (before release) |
| Production | `tags` | Live users | Manual (on release) |

### 5. Security

- **Never commit secrets** - Use CI/CD variables
- **Rotate credentials regularly** - Every 90 days
- **Use protected branches** - Require approvals for `main`
- **Enable branch protection** - Prevent force push to `main`
- **Review security scan results** - Address vulnerabilities
- **Audit deployments** - Monitor deployment logs

### 6. Testing

- **Write tests first** - TDD approach
- **Maintain coverage** - Aim for >80%
- **Run tests locally** - Before pushing
- **Fix failing tests** - Don't merge broken builds
- **Test deployments** - Always test in staging first

### 7. Monitoring

After deployment:

- **Check logs** - `docker compose logs -f`
- **Monitor metrics** - Grafana dashboards
- **Verify health** - `/health` endpoint
- **Test functionality** - Smoke tests
- **Watch alerts** - Prometheus alerts

---

## CI/CD Metrics

Track these metrics to improve your pipeline:

- **Pipeline Success Rate** - Target: >95%
- **Average Pipeline Duration** - Target: <15 minutes
- **Deployment Frequency** - Goal: Multiple per week
- **Mean Time to Recovery** - Goal: <1 hour
- **Test Coverage** - Target: >80%

View in GitLab: **Analytics → CI/CD Analytics**

---

## Additional Resources

- **GitLab CI/CD Documentation:** https://docs.gitlab.com/ee/ci/
- **Docker Compose Documentation:** https://docs.docker.com/compose/
- **Container Registry:** https://docs.gitlab.com/ee/user/packages/container_registry/
- **GitLab Environments:** https://docs.gitlab.com/ee/ci/environments/

---

## Support

For CI/CD issues:

1. Check [Troubleshooting](#troubleshooting) section
2. Review pipeline logs in GitLab
3. Check [`.gitlab/ci-variables.md`](.gitlab/ci-variables.md) for variable setup
4. Review [project issues](https://gitlab.com/youruser/llm-proxy/-/issues)

---

**Last Updated:** January 29, 2026  
**Pipeline Version:** 1.0.0
