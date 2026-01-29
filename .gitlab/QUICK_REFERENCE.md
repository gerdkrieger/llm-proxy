# GitLab CI/CD Quick Reference

## 🚀 Quick Commands

### Local Development
```bash
# Run tests
make test

# Run linter
make lint

# Build locally
make build

# Format code
make fmt
```

### Git Workflow
```bash
# Create feature branch
git checkout -b feature/my-feature

# Commit with conventional format
git commit -m "feat: add new feature"

# Push and create MR
git push origin feature/my-feature
```

### Deployment
```bash
# On deployment server
cd /opt/llm-proxy
./deployments/scripts/deploy-from-registry.sh production v1.0.0
```

---

## 📊 Pipeline Stages

| Stage | Duration | Runs On |
|-------|----------|---------|
| Lint | ~2-3 min | MR, main, develop |
| Test | ~5-8 min | MR, main, develop |
| Security | ~2-4 min | MR, main, develop |
| Build | ~3-5 min | MR, main, develop, tags |
| Docker | ~5-10 min | main, develop, tags |
| Deploy | ~3-5 min | Manual trigger |

**Total Duration:** ~20-35 minutes (without deployment)

---

## 🔑 Required CI/CD Variables

### Minimum Setup (Development)
```
SSH_PRIVATE_KEY    # SSH key for deployments
DEV_HOST           # dev.example.com
DEV_USER           # deploy
DEV_PATH           # /opt/llm-proxy
DEV_URL            # https://dev.example.com
```

### Production (Additional)
```
PROD_HOST          # prod.example.com (protected)
PROD_USER          # deploy (protected)
PROD_PATH          # /opt/llm-proxy (protected)
PROD_URL           # https://api.example.com (protected)
```

**Set in:** Settings → CI/CD → Variables

---

## 🏷️ Image Tags

| Tag | Description | Example |
|-----|-------------|---------|
| `<sha>` | Commit SHA | `abc12345` |
| `<branch>` | Branch name | `main`, `develop` |
| `latest` | Latest main | `latest` |
| `<version>` | Git tag | `v1.0.0` |

**Registry:** `registry.gitlab.com/<namespace>/<project>/<service>:<tag>`

---

## 🔄 Common Workflows

### Deploy to Development
```
1. Push to develop branch
2. Wait for pipeline to complete
3. Go to CI/CD → Pipelines
4. Click ▶️ on "deploy:development"
5. Verify at DEV_URL
```

### Create Production Release
```
1. Merge to main
2. Test in staging
3. Create tag: git tag -a v1.0.0 -m "Release v1.0.0"
4. Push tag: git push origin v1.0.0
5. Wait for tag pipeline
6. Trigger "deploy:production"
7. Verify at PROD_URL
```

---

## 🐛 Quick Troubleshooting

### Pipeline Stuck?
```bash
# Check runner availability
Settings → CI/CD → Runners
Enable "shared runners" if needed
```

### Tests Failing?
```bash
# Run locally first
make test

# Check service connections
DB_HOST=postgres  # NOT localhost in CI!
```

### Deployment Fails?
```bash
# Verify SSH key in GitLab
# Check public key on server:
cat ~/.ssh/authorized_keys

# Test SSH manually:
ssh deploy@dev.example.com
```

### Images Not Found?
```bash
# Check registry
Packages & Registries → Container Registry

# Verify image exists
docker manifest inspect registry.gitlab.com/.../backend:v1.0.0

# Login on deployment server
docker login registry.gitlab.com
```

---

## 📚 Documentation

- **Full CI/CD Guide:** [`CICD.md`](../CICD.md)
- **Variable Setup:** [`.gitlab/ci-variables.md`](ci-variables.md)
- **Deployment Guide:** [`DEPLOYMENT.md`](../DEPLOYMENT.md)
- **Pipeline Config:** [`.gitlab-ci.yml`](../.gitlab-ci.yml)

---

## 🔗 Quick Links

- **Pipelines:** https://gitlab.com/[namespace]/llm-proxy/-/pipelines
- **Container Registry:** https://gitlab.com/[namespace]/llm-proxy/container_registry
- **Environments:** https://gitlab.com/[namespace]/llm-proxy/-/environments
- **CI/CD Settings:** https://gitlab.com/[namespace]/llm-proxy/-/settings/ci_cd

---

## 💡 Tips

1. **Always test locally before pushing**
2. **Use conventional commit messages**
3. **Review pipeline logs for failures**
4. **Deploy to staging before production**
5. **Monitor metrics after deployment**
6. **Keep .env files updated on servers**
7. **Rotate SSH keys every 90 days**

---

**Need Help?** See [CICD.md](../CICD.md) for detailed documentation.
