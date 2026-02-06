# LLM-Proxy Documentation

Welcome to the LLM-Proxy documentation. This directory contains comprehensive guides and references for using, deploying, and maintaining the LLM-Proxy system.

## 📚 Documentation Structure

### Getting Started

- **[RESUME-PROJECT.md](RESUME-PROJECT.md)** - Start here when resuming work on the project
- **[TESTING.md](TESTING.md)** - Testing guide and best practices
- **[MAINTENANCE.md](MAINTENANCE.md)** - Routine maintenance tasks

### Core Documentation

- **[GIT_WORKFLOW.md](GIT_WORKFLOW.md)** - Git workflow and branching strategy
- **[CI_CD_REGISTRY_SETUP.md](CI_CD_REGISTRY_SETUP.md)** - CI/CD pipeline and GitLab Container Registry setup

### Deployment

- **[deployment/CICD.md](deployment/CICD.md)** - CI/CD pipeline configuration and usage

### User Guides

All user-facing guides are in the `guides/` directory:

- **[guides/ADMIN_API.md](guides/ADMIN_API.md)** - Admin API reference
- **[guides/CONTENT_FILTERING.md](guides/CONTENT_FILTERING.md)** - Content filtering system
- **[guides/FILTER_MANAGEMENT_GUIDE.md](guides/FILTER_MANAGEMENT_GUIDE.md)** - Managing content filters
- **[guides/QUICK_START_FILTERS.md](guides/QUICK_START_FILTERS.md)** - Quick start guide for filters
- **[guides/BULK_IMPORT_GUIDE.md](guides/BULK_IMPORT_GUIDE.md)** - Bulk importing filters
- **[guides/OPENWEBUI_INTEGRATION_GUIDE.md](guides/OPENWEBUI_INTEGRATION_GUIDE.md)** - Open WebUI integration
- **[guides/MODEL_MANAGEMENT_MVP.md](guides/MODEL_MANAGEMENT_MVP.md)** - Managing LLM models
- **[guides/MODEL_SYNC_GUIDE.md](guides/MODEL_SYNC_GUIDE.md)** - Syncing provider models
- **[guides/ANTHROPIC_CREDITS_GUIDE.md](guides/ANTHROPIC_CREDITS_GUIDE.md)** - Managing Anthropic API credits
- **[guides/ENVIRONMENT_CONFIGURATION.md](guides/ENVIRONMENT_CONFIGURATION.md)** - Environment setup and configuration

## 🚀 Quick Links

### Development

```bash
# Start local development
docker compose -f docker-compose.dev.yml up -d

# Check status
docker compose ps

# View logs
docker compose logs -f
```

**Services:**
- Backend API: http://localhost:8080
- Admin UI: http://localhost:3005
- PostgreSQL: localhost:5433
- Redis: localhost:6380

### Git Workflow

```bash
# Daily development (commit + merge + push)
./scripts/maintenance/git-update.sh -m "feat: My feature"

# See workflow details
cat docs/GIT_WORKFLOW.md
```

### Deployment

```bash
# Deploy to production
./deploy.sh

# See deployment guide
cat DEPLOYMENT.md
```

## 📖 Additional Resources

### Root Level Documentation

- **[../README.md](../README.md)** - Project overview and quick start
- **[../DEPLOYMENT.md](../DEPLOYMENT.md)** - Deployment guide
- **[../scripts/README.md](../scripts/README.md)** - Scripts documentation

### Specialized Documentation

- **[../migrations/README.md](../migrations/README.md)** - Database migrations
- **[../migrations/DEPLOYMENT_CHECKLIST.md](../migrations/DEPLOYMENT_CHECKLIST.md)** - Pre-deployment checklist
- **[../filter-templates/README.md](../filter-templates/README.md)** - Filter templates
- **[../admin-ui/README.md](../admin-ui/README.md)** - Admin UI documentation

## 🗂️ Documentation Cleanup (2026-02-06)

The following documentation was removed as it was outdated or superseded:

### Removed Documents

- **Session Docs** (historical snapshots from Feb 4, 2026)
  - `sessions/SESSION-SUMMARY-2026-02-04.md`
  - `sessions/SESSION-CONTINUATION-2026-02-04.md`
  - `sessions/SESSION-NEXT-STEPS.md`

- **Fix/Status Docs** (one-time fixes, completed)
  - `CODE-CLEANUP-SUMMARY.md`
  - `LIVE-SERVER-FIX.md`
  - `LIVE-SERVER-COMMANDS.md` (replaced by Docker Compose)
  - `deployment/DEPLOYMENT-STATUS.md`
  - `deployment/DOCKER-DEPLOYMENT-FIX.md`

- **Duplicates** (superseded by newer versions)
  - `deployment/GIT_WORKFLOW.md` → moved to `GIT_WORKFLOW.md`
  - `deployment/DEPLOYMENT.md` → moved to root `DEPLOYMENT.md`

- **Outdated**
  - `TESTING_REPORT.md` (old test report)
  - `NEXT-STEPS.md` (completed TODOs)
  - `TROUBLESHOOTING.md` (referenced deleted scripts)

### Result

- **Before:** 30 documentation files
- **After:** 17 documentation files  
- **Reduction:** 43% fewer files to maintain

All important information from removed documents has been integrated into current documentation.

## 🛠️ Maintaining Documentation

### When Adding New Documentation

1. **Choose the right location:**
   - Root level: Project-wide documentation (README, DEPLOYMENT)
   - `docs/guides/`: User guides and how-tos
   - `docs/deployment/`: Deployment-specific docs
   - Component directories: Component-specific docs (admin-ui/, migrations/, etc.)

2. **Update this README** - Add the new document to the appropriate section

3. **Use clear naming:**
   - `GUIDE.md` for step-by-step instructions
   - `API.md` for API references
   - `TROUBLESHOOTING.md` for common issues
   - `README.md` for directory overviews

4. **Keep it current:**
   - Review annually
   - Update when features change
   - Remove when superseded

### When Removing Documentation

1. **Verify it's truly obsolete** - Check for unique information
2. **Move important info** - Integrate into current docs if needed
3. **Update this README** - Add to "Removed Documents" section
4. **Update links** - Fix broken references in other docs

## 📝 Contributing

When updating documentation:

1. **Keep it concise** - Focus on what users need to know
2. **Use examples** - Show real commands and code
3. **Update dates** - Add "Last Updated" dates for time-sensitive docs
4. **Test commands** - Verify all commands actually work
5. **Link related docs** - Help users discover relevant information

---

**Last Updated:** 2026-02-06  
**Maintainer:** LLM-Proxy Team
