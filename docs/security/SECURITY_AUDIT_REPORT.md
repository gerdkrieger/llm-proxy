# 🔒 Security Audit Report - LLM-Proxy

**Audit Date:** 2026-03-08  
**Repository:** https://github.com/gerdkrieger/llm-proxy  
**Audited Branch:** `develop`, `master`  
**Status:** 🔴 **CRITICAL ISSUES FOUND**

---

## 📊 Executive Summary

| Severity | Count | Status |
|----------|-------|--------|
| 🔴 **CRITICAL** | 3 | **ACTION REQUIRED** |
| 🟠 **HIGH** | 5 | Fix Recommended |
| 🟡 **MEDIUM** | 4 | Improvement Suggested |
| ✅ **GOOD** | 8 | Secure |

---

## 🔴 CRITICAL FINDINGS

### 1. ❌ Token Files Committed to Git History

**Severity:** 🔴 CRITICAL  
**File:** `OPENWEBUI_TOKEN_30DAYS.txt`, `OPENWEBUI_QUICK_CONFIG.txt`

**Issue:**
- Files containing configuration/tokens are tracked by Git
- Present in commits: `df6ed05`, `687af9d`
- Currently in repository (both local and remote)

**Impact:**
- Tokens/configs exposed in Git history
- Accessible to anyone with repo access

**Remediation:**
```bash
# 1. Add to .gitignore
echo "OPENWEBUI_*.txt" >> .gitignore

# 2. Remove from Git tracking (keep local file)
git rm --cached OPENWEBUI_TOKEN_30DAYS.txt OPENWEBUI_QUICK_CONFIG.txt

# 3. Commit changes
git commit -m "security: remove token files from git tracking"

# 4. (Optional) Remove from history if repo is public
git filter-branch --force --index-filter \
  'git rm --cached --ignore-unmatch OPENWEBUI_*.txt' \
  --prune-empty --tag-name-filter cat -- --all
```

---

### 2. ⚠️ Development Credentials in Documentation

**Severity:** 🔴 CRITICAL (if repo goes public)  
**Files:** Multiple `.md` files in `docs/`

**Issue:**
- Hardcoded `admin_dev_key_12345678901234567890123456789012` in 40+ places
- Found in:
  - `docs/guides/ADMIN_API.md`
  - `docs/guides/QUICK_START_FILTERS.md`
  - `docs/RESUME-PROJECT.md`
  - `tests/README.md`
  - And many more...

**Impact:**
- If this is the **actual production key**, it's exposed
- If it's only a dev key, it should be clearly marked as example

**Remediation:**
```bash
# Replace with placeholder
find docs tests -name "*.md" -type f -exec sed -i \
  's/admin_dev_key_12345678901234567890123456789012/YOUR_ADMIN_API_KEY_HERE/g' {} +

# Or add disclaimer
# "⚠️ EXAMPLE ONLY - Replace with your actual key in production!"
```

---

### 3. ⚠️ `.env` File Contains Real Secrets

**Severity:** 🔴 CRITICAL  
**File:** `.env` (root directory)

**Issue:**
- Contains real API keys:
  - `CLAUDE_API_KEY=sk-ant-api03-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`
  - `OPENAI_API_KEY=sk-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX`
  - `OAUTH_JWT_SECRET`, `DB_PASSWORD`, `ADMIN_API_KEYS`

**Impact:**
- ✅ **GOOD:** File is in `.gitignore` and never committed
- ⚠️ **RISK:** Local file could be accidentally shared/copied

**Remediation:**
```bash
# 1. Verify .env is in .gitignore (ALREADY DONE ✅)
grep "^\.env$" .gitignore

# 2. Rotate keys if they were ever exposed:
# - Anthropic: https://console.anthropic.com/settings/keys
# - OpenAI: https://platform.openai.com/api-keys

# 3. Use .env.example template instead:
cp .env .env.local.backup
rm .env
# Only use .env.example with placeholder values
```

---

## 🟠 HIGH SEVERITY FINDINGS

### 4. Docker Compose Default Passwords

**Severity:** 🟠 HIGH  
**Files:** `docker-compose.dev.yml`, `deployments/docker-compose.*.yml`

**Issue:**
```yaml
POSTGRES_PASSWORD: ${DB_PASSWORD:-dev_password_2024}
```
- Fallback password is weak: `dev_password_2024`
- Same password in multiple files

**Impact:**
- If env var not set, weak default is used
- Predictable password in development

**Remediation:**
```yaml
# Option 1: Remove default, require env var
POSTGRES_PASSWORD: ${DB_PASSWORD:?ERROR: DB_PASSWORD not set}

# Option 2: Generate random default
POSTGRES_PASSWORD: ${DB_PASSWORD:-$(openssl rand -hex 16)}
```

---

### 5. Hardcoded Test Credentials

**Severity:** 🟠 HIGH  
**Files:** `tests/load/*.js`, `tests/integration/api_test.go`

**Issue:**
```javascript
const CLIENT_SECRET = __ENV.CLIENT_SECRET || 'test_secret_123456';
const ADMIN_API_KEY = __ENV.ADMIN_API_KEY || 'admin_dev_key_12345...';
```

**Impact:**
- Test files work without configuration
- Could leak into production if copied

**Remediation:**
```javascript
// Fail if env not set (safer)
const CLIENT_SECRET = __ENV.CLIENT_SECRET;
if (!CLIENT_SECRET) {
  throw new Error('CLIENT_SECRET env var required');
}
```

---

### 6. No Secret Scanning in CI/CD

**Severity:** 🟠 HIGH  
**File:** `.gitlab-ci.yml.disabled`

**Issue:**
- No automated secret detection
- No pre-commit hooks for secret scanning

**Remediation:**
```yaml
# Add to .gitlab-ci.yml or GitHub Actions:
secret-scan:
  stage: security
  script:
    - |
      # Check for common secret patterns
      if git diff HEAD~1 | grep -E "sk-[a-zA-Z0-9]{20,}|api[_-]?key.*=.*['\"][^'\"]{20,}"; then
        echo "❌ Potential secret detected!"
        exit 1
      fi
```

---

## 🟡 MEDIUM SEVERITY FINDINGS

### 7. SQL Files with Test Data

**Severity:** 🟡 MEDIUM  
**Files:** `seed-filters-live.sql`, `migrate-content-filters-schema.sql`

**Issue:**
- Root directory contains SQL files with test/production data
- Should be in `migrations/` directory

**Remediation:**
```bash
mv seed-filters-live.sql migrations/seeds/
mv migrate-content-filters-schema.sql migrations/archive/
mv recreate-content-filters-correct-order.sql migrations/archive/
```

---

### 8. No HTTPS Enforcement in Code

**Severity:** 🟡 MEDIUM

**Issue:**
- No code-level enforcement of HTTPS for API endpoints
- Relies on Caddy/reverse proxy

**Remediation:**
```go
// In router.go
if cfg.Environment == "production" && r.URL.Scheme != "https" {
    http.Redirect(w, r, "https://"+r.Host+r.URL.Path, http.StatusMovedPermanently)
}
```

---

### 9. Missing Security Headers

**Severity:** 🟡 MEDIUM

**Issue:**
- No security headers middleware
- Missing: CSP, X-Frame-Options, HSTS, etc.

**Remediation:**
```go
// Add security headers middleware
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        next.ServeHTTP(w, r)
    })
}
```

---

### 10. Encryption Key in Plain Text

**Severity:** 🟡 MEDIUM  
**File:** `.env`

**Issue:**
```bash
ENCRYPTION_KEY=e5a6c2183015ccd4f3d9c832eccb01c1a932f8f082f31fc0f3b72fe0c60f0de9
```
- Encryption key stored in `.env` (not in a secret manager)

**Impact:**
- ✅ Good: Used for encrypting DB secrets
- ⚠️ Risk: If `.env` is compromised, all encrypted data is compromised

**Remediation:**
```bash
# Option 1: Use secret manager (production)
# AWS: AWS Secrets Manager
# GCP: Google Secret Manager
# Azure: Azure Key Vault

# Option 2: Kubernetes secret
kubectl create secret generic llm-proxy-encryption-key \
  --from-literal=ENCRYPTION_KEY=$(openssl rand -hex 32)
```

---

## ✅ GOOD PRACTICES FOUND

### 1. ✅ `.env` in `.gitignore`
- `.env` properly excluded from Git
- Never committed to repository

### 2. ✅ Example Files Provided
- `.env.example`, `.env.docker.example`, `.env.production.example`
- Good documentation for setup

### 3. ✅ Secrets Encrypted in Database
- Provider API keys encrypted with AES-256-GCM
- Migration `000012_add_provider_api_keys.up.sql` properly structured

### 4. ✅ Password Hashing
- OAuth client secrets hashed (bcrypt)
- Migration `001_add_hash_and_stats_columns.sql` adds `client_secret_hash`

### 5. ✅ No Hardcoded Secrets in Go Code
- All secrets loaded from environment variables
- `viper.GetString()`, `os.Getenv()` used correctly

### 6. ✅ Docker Secrets Support
- Docker Compose supports env vars
- No secrets in Dockerfiles

### 7. ✅ JWT Token Expiration
- Access tokens: 1h
- Refresh tokens: 720h (30 days)
- Configurable via env

### 8. ✅ CORS Configured
- Proper CORS headers in router
- Whitelist of allowed origins

---

## 📋 Recommended Actions (Priority Order)

### 🔴 IMMEDIATE (Within 24h)

1. **Remove token files from Git:**
   ```bash
   git rm --cached OPENWEBUI_TOKEN_30DAYS.txt OPENWEBUI_QUICK_CONFIG.txt
   echo "OPENWEBUI_*.txt" >> .gitignore
   git commit -m "security: remove token files from git tracking"
   git push origin develop
   ```

2. **Verify API keys not exposed:**
   - Check if repo is public on GitHub
   - If yes, rotate ALL keys immediately:
     - Anthropic: https://console.anthropic.com/settings/keys
     - OpenAI: https://platform.openai.com/api-keys

3. **Replace hardcoded dev keys in docs:**
   ```bash
   find docs tests -name "*.md" -exec sed -i \
     's/admin_dev_key_12345678901234567890123456789012/YOUR_ADMIN_API_KEY_HERE/g' {} +
   ```

---

### 🟠 HIGH PRIORITY (Within 1 Week)

4. **Add secret scanning to CI/CD**
5. **Implement security headers middleware**
6. **Move SQL files to proper directories**
7. **Add pre-commit hooks for secret detection**

---

### 🟡 MEDIUM PRIORITY (Within 1 Month)

8. **Migrate to secret manager (production)**
9. **Enforce HTTPS in code**
10. **Add security audit to README**
11. **Create SECURITY.md file**

---

## 🔍 Best Practices for Public GitHub Repo

### ✅ DO:
- ✅ Use `.env.example` with placeholder values
- ✅ Add comprehensive `.gitignore`
- ✅ Encrypt secrets in database
- ✅ Use environment variables for all secrets
- ✅ Document security setup in README
- ✅ Add `SECURITY.md` with vulnerability reporting

### ❌ DON'T:
- ❌ Commit `.env` files with real secrets
- ❌ Hardcode API keys in code
- ❌ Include production configs in repo
- ❌ Use weak default passwords
- ❌ Commit database dumps with real data
- ❌ Store encryption keys in plain text files

---

## 📄 Files to Create

### 1. `SECURITY.md`
```markdown
# Security Policy

## Reporting a Vulnerability

Email: security@yourdomain.com
Response Time: 48 hours

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Security Features

- API key encryption (AES-256-GCM)
- Password hashing (bcrypt)
- JWT tokens (HS256)
- CORS protection
- Rate limiting
```

### 2. `.github/dependabot.yml`
```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "npm"
    directory: "/admin-ui"
    schedule:
      interval: "weekly"
```

### 3. `.pre-commit-config.yaml`
```yaml
repos:
  - repo: https://github.com/gitleaks/gitleaks
    rev: v8.18.0
    hooks:
      - id: gitleaks
```

---

## 🎯 Compliance Checklist

- [ ] All secrets removed from Git history
- [ ] `.gitignore` includes all secret files
- [ ] Example files use placeholder values
- [ ] Production secrets in secret manager
- [ ] Security headers implemented
- [ ] Secret scanning in CI/CD
- [ ] SECURITY.md created
- [ ] Pre-commit hooks configured
- [ ] Documentation updated
- [ ] Team trained on security practices

---

**Report Generated:** 2026-03-08  
**Auditor:** Automated Security Scan  
**Next Audit:** Recommended within 3 months
