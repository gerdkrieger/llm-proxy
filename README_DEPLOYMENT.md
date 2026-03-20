# 🚀 Quick Deployment Guide

## One-Command Deployment

```bash
# Build, push to registries, and deploy to production
make release VERSION=v1.0.0
```

That's it! ✅

---

## Documentation

- 📖 **Full Guide:** [docs/deployment/REGISTRY_DEPLOYMENT.md](docs/deployment/REGISTRY_DEPLOYMENT.md)
- 🚀 **Quick Start:** [docs/deployment/QUICKSTART_REGISTRY.md](docs/deployment/QUICKSTART_REGISTRY.md)
- 🏗️ **Architecture:** [docs/deployment/DEPLOYMENT_ARCHITECTURE.md](docs/deployment/DEPLOYMENT_ARCHITECTURE.md)
- 📝 **Cheatsheet:** [docs/deployment/CHEATSHEET.md](docs/deployment/CHEATSHEET.md)

---

## Common Commands

```bash
make help              # Show all commands
make status            # Check production status
make logs SERVICE=X    # View logs
make rollback VERSION=X # Rollback to previous version
make backup-db         # Backup database
```

---

## Registry Strategy

- **Backend + Admin-UI** → GitHub Container Registry (`ghcr.io`)
- **Landing Page** → GitLab Container Registry (ONLY!)

---

## Support

Need help? Check the documentation above or run `make help`.
