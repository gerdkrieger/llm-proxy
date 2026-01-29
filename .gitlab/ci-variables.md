# GitLab CI/CD Variables Configuration

This document lists all required and optional CI/CD variables for the LLM-Proxy pipeline.

## Required Variables

These variables must be configured in GitLab **Settings → CI/CD → Variables**:

### Container Registry Access
These are automatically provided by GitLab:
- `CI_REGISTRY` - GitLab Container Registry URL (automatic)
- `CI_REGISTRY_USER` - Registry username (automatic)
- `CI_REGISTRY_PASSWORD` - Registry password (automatic)
- `CI_REGISTRY_IMAGE` - Full image path (automatic)

### Deployment - Development Environment

Configure these for `deploy:development` job:

| Variable | Description | Example | Protected | Masked |
|----------|-------------|---------|-----------|--------|
| `DEV_HOST` | Development server hostname/IP | `dev.example.com` | No | No |
| `DEV_USER` | SSH username for dev server | `deploy` | No | No |
| `DEV_PATH` | Path to project on dev server | `/opt/llm-proxy` | No | No |
| `DEV_URL` | Development environment URL | `https://dev.example.com` | No | No |
| `SSH_PRIVATE_KEY` | SSH private key for deployment | `-----BEGIN OPENSSH PRIVATE KEY-----...` | No | Yes |

### Deployment - Staging Environment

Configure these for `deploy:staging` job:

| Variable | Description | Example | Protected | Masked |
|----------|-------------|---------|-----------|--------|
| `STAGING_HOST` | Staging server hostname/IP | `staging.example.com` | No | No |
| `STAGING_USER` | SSH username for staging server | `deploy` | No | No |
| `STAGING_PATH` | Path to project on staging server | `/opt/llm-proxy` | No | No |
| `STAGING_URL` | Staging environment URL | `https://staging.example.com` | No | No |

### Deployment - Production Environment

Configure these for `deploy:production` job:

| Variable | Description | Example | Protected | Masked |
|----------|-------------|---------|-----------|--------|
| `PROD_HOST` | Production server hostname/IP | `prod.example.com` | **Yes** | No |
| `PROD_USER` | SSH username for production server | `deploy` | **Yes** | No |
| `PROD_PATH` | Path to project on production server | `/opt/llm-proxy` | **Yes** | No |
| `PROD_URL` | Production environment URL | `https://api.example.com` | **Yes** | No |

## Optional Variables

### Test Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GO_VERSION` | Go version for builds | `1.21` |
| `POSTGRES_VERSION` | PostgreSQL version for tests | `14` |
| `REDIS_VERSION` | Redis version for tests | `7` |

### Build Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `CGO_ENABLED` | Enable/disable CGO | `0` |
| `GOOS` | Target OS | `linux` |
| `GOARCH` | Target architecture | `amd64` |

## How to Configure Variables in GitLab

### Method 1: GitLab Web UI

1. Navigate to your project on GitLab.com
2. Go to **Settings → CI/CD**
3. Expand **Variables** section
4. Click **Add variable**
5. Fill in:
   - **Key**: Variable name (e.g., `PROD_HOST`)
   - **Value**: Variable value
   - **Type**: Variable or File
   - **Environment scope**: All (default) or specific environment
   - **Protect variable**: Check for production variables
   - **Mask variable**: Check for sensitive values
6. Click **Add variable**

### Method 2: GitLab CLI

```bash
# Install glab CLI
# https://gitlab.com/gitlab-org/cli

# Add a variable
glab variable set PROD_HOST "prod.example.com" --scope=prod

# Add a protected, masked variable
glab variable set SSH_PRIVATE_KEY "$(cat ~/.ssh/id_rsa)" --protected --masked
```

## SSH Key Setup for Deployment

### 1. Generate SSH Key Pair

```bash
# Generate a new SSH key pair (no passphrase for CI/CD)
ssh-keygen -t ed25519 -C "gitlab-ci-deploy" -f ~/.ssh/gitlab_ci_deploy -N ""

# This creates:
# - Private key: ~/.ssh/gitlab_ci_deploy
# - Public key: ~/.ssh/gitlab_ci_deploy.pub
```

### 2. Add Public Key to Deployment Servers

```bash
# Copy public key to development server
ssh-copy-id -i ~/.ssh/gitlab_ci_deploy.pub deploy@dev.example.com

# Copy public key to staging server
ssh-copy-id -i ~/.ssh/gitlab_ci_deploy.pub deploy@staging.example.com

# Copy public key to production server
ssh-copy-id -i ~/.ssh/gitlab_ci_deploy.pub deploy@prod.example.com
```

### 3. Add Private Key to GitLab

```bash
# Display private key
cat ~/.ssh/gitlab_ci_deploy

# Copy the entire output including:
# -----BEGIN OPENSSH PRIVATE KEY-----
# ...key content...
# -----END OPENSSH PRIVATE KEY-----
```

Then add to GitLab:
- **Key**: `SSH_PRIVATE_KEY`
- **Value**: Paste the entire private key
- **Type**: Variable
- **Mask variable**: ✅ Check

## Environment-Specific Variables

You can scope variables to specific environments:

### Development Scope
```
Environment scope: develop
Variables:
- DEPLOY_ENV=development
- LOG_LEVEL=debug
```

### Staging Scope
```
Environment scope: main
Variables:
- DEPLOY_ENV=staging
- LOG_LEVEL=info
```

### Production Scope
```
Environment scope: tags
Variables:
- DEPLOY_ENV=production
- LOG_LEVEL=warn
```

## Variable Priority

GitLab resolves variables in this order (highest to lowest):
1. Pipeline-specific variables
2. Project variables
3. Group variables
4. Instance variables
5. Predefined variables

## Testing Variables Locally

You can test the pipeline locally using GitLab Runner:

```bash
# Install gitlab-runner
# https://docs.gitlab.com/runner/install/

# Register runner
gitlab-runner register

# Run a specific job locally
gitlab-runner exec docker lint:backend

# Run with custom variables
gitlab-runner exec docker \
  --env "DEV_HOST=localhost" \
  --env "DEV_USER=deploy" \
  deploy:development
```

## Security Best Practices

1. **Never commit secrets** to the repository
2. **Use protected variables** for production
3. **Mask sensitive values** (SSH keys, passwords, tokens)
4. **Rotate credentials regularly** (every 90 days)
5. **Use environment scopes** to limit variable access
6. **Enable "Protected" flag** for production variables
7. **Audit variable access** regularly in GitLab audit logs

## Common Issues

### Issue: SSH Connection Fails

**Solution:**
- Verify SSH key is correctly formatted (include BEGIN/END lines)
- Check SSH_PRIVATE_KEY variable is masked but not protected initially
- Ensure public key is in `~/.ssh/authorized_keys` on target server
- Verify DEPLOY_HOST is correct and accessible

### Issue: Docker Login Fails

**Solution:**
- Verify CI_REGISTRY_PASSWORD is set correctly
- Check Container Registry is enabled: Settings → General → Visibility → Container Registry
- Ensure Docker is running on GitLab Runner

### Issue: Test Database Connection Fails

**Solution:**
- Verify service names match (postgres, redis)
- Check DB_HOST uses service name, not localhost
- Ensure ports are default (5432 for PostgreSQL, 6379 for Redis)

## References

- [GitLab CI/CD Variables Documentation](https://docs.gitlab.com/ee/ci/variables/)
- [GitLab Container Registry](https://docs.gitlab.com/ee/user/packages/container_registry/)
- [GitLab CI/CD SSH Keys](https://docs.gitlab.com/ee/ci/ssh_keys/)
