#!/bin/bash
# =============================================================================
# SERVER CLEANUP SCRIPT
# =============================================================================
# Removes unnecessary files from production server
# Keeps ONLY essential config files and Docker volumes
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  SERVER CLEANUP${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Confirm server
read -p "Server hostname (e.g., openweb): " SERVER
if [ -z "$SERVER" ]; then
    echo -e "${RED}Error: Server hostname required${NC}"
    exit 1
fi

echo -e "${YELLOW}This will remove ALL source code from: $SERVER${NC}"
echo -e "${YELLOW}Only essential config files will remain.${NC}"
echo ""

# Show what will be kept
echo -e "${GREEN}Files that will be KEPT:${NC}"
echo "  /opt/llm-proxy/deployments/docker/.env"
echo "  /opt/llm-proxy/deployments/docker/docker-compose.registry-deploy.yml"
echo "  /opt/llm-proxy/deployments/docker/prometheus.yml"
echo "  /opt/llm-proxy/deployments/docker/grafana/"
echo "  /opt/llm-proxy/scripts/deployment/server-deploy.sh"
echo "  /etc/caddy/Caddyfile"
echo "  Docker volumes (postgres_data, redis_data, etc.)"
echo ""

# Show what will be removed
echo -e "${RED}Files that will be REMOVED:${NC}"
echo "  - All source code (.go files, internal/, cmd/, pkg/)"
echo "  - Frontend source (admin-ui/src/, node_modules/)"
echo "  - Landing page source (landing/*.html, *.png)"
echo "  - Git repository (.git/)"
echo "  - Documentation (docs/, README.md)"
echo "  - Build tools (Makefile, Dockerfile)"
echo "  - Old compose files"
echo ""

# Safety check
read -p "Continue? (type 'yes' to confirm): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo -e "${YELLOW}Aborted.${NC}"
    exit 0
fi

echo ""
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Creating backup before cleanup...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

# Create backup
BACKUP_NAME="server-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
echo -e "${YELLOW}Creating backup: $BACKUP_NAME${NC}"

ssh $SERVER "cd /opt && tar -czf /tmp/$BACKUP_NAME llm-proxy/"
scp $SERVER:/tmp/$BACKUP_NAME ./backups/
ssh $SERVER "rm /tmp/$BACKUP_NAME"

echo -e "${GREEN}✓ Backup saved to: ./backups/$BACKUP_NAME${NC}"
echo ""

# Create cleanup script on server
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Removing unnecessary files...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

ssh $SERVER bash << 'ENDSSH'
#!/bin/bash
set -e

cd /opt/llm-proxy

echo "Removing source code..."

# Remove Go source code
rm -rf internal/ cmd/ pkg/ api/ tests/
rm -f go.mod go.sum
rm -f *.go

# Remove admin-ui source
rm -rf admin-ui/src/
rm -rf admin-ui/node_modules/
rm -rf admin-ui/dist/
rm -f admin-ui/package*.json
rm -f admin-ui/vite.config.js
rm -f admin-ui/svelte.config.js
rm -f admin-ui/tailwind.config.js
rm -f admin-ui/postcss.config.js
rm -f admin-ui/jsconfig.json
rm -f admin-ui/index.html
rm -f admin-ui/.dockerignore
rm -f admin-ui/.gitignore
rm -f admin-ui/.vscode/
rm -f admin-ui/README.md
rm -f admin-ui/ADMIN_UI_README.md

# Remove admin-ui Dockerfile (not needed - we pull from registry)
rm -f admin-ui/Dockerfile
rm -f admin-ui/Dockerfile.dev

# Remove landing page source
rm -rf landing/
rm -f docker-compose.landing.yml

# Remove Git repository
rm -rf .git/
rm -f .gitignore
rm -f .gitlab/
rm -f .gitlab-ci.yml
rm -f .git-workflow-quick-ref.txt

# Remove documentation
rm -rf docs/
rm -f README.md
rm -f README_DEPLOYMENT.md
rm -f DEPLOYMENT.md
rm -f OPENWEBUI_*.txt

# Remove build files
rm -f Makefile
rm -f Makefile.registry
rm -f Dockerfile.dev

# Remove old docker-compose files (keep only registry-deploy!)
cd deployments/docker/
rm -f docker-compose.yml
rm -f docker-compose.prod.yml
rm -f docker-compose.build.yml
rm -f Dockerfile
rm -f deploy.sh
rm -f validate-schema.sh
rm -f DEPLOYMENT.md
rm -f README.md

# Remove old deployment scripts
cd ../..
rm -rf scripts/setup/
rm -rf scripts/maintenance/
rm -rf scripts/testing/
rm -f scripts/deployment/build-and-push.sh
rm -f scripts/deployment/setup-dual-push.sh

# Remove misc files
rm -f example-filters.csv
rm -f test-bulk-import.json
rm -f *.sql
rm -f .env.example
rm -f .env.docker.example
rm -f .env.production.example
rm -f .env.local
rm -f .env.development
rm -f .dockerignore
rm -f .air.toml

# Remove configs (keep only what's needed)
cd configs/
rm -f Caddyfile.example
rm -f Caddyfile.correct
rm -f config.example.yaml

# Remove logs directory (logs are in Docker containers)
cd ..
rm -rf logs/

# Remove filter templates (handled by backend container)
rm -rf filter-templates/

# Remove migrations (handled by backend container)
rm -rf migrations/

# Remove binary output directory
rm -rf bin/

# Keep only essential deployments structure
cd deployments/
rm -f docker-compose.openwebui.yml
rm -f docker-compose.production.yml
rm -f .env.backup.*
rm -rf kubernetes/
rm -rf scripts/

echo "✓ Cleanup complete"
ENDSSH

echo -e "${GREEN}✓ Source code removed${NC}"
echo ""

# Show final structure
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Final server structure:${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

ssh $SERVER "cd /opt/llm-proxy && find . -maxdepth 3 -type f ! -path './.claude/*' | sort"

echo ""
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Disk space saved:${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

ssh $SERVER "du -sh /opt/llm-proxy"

echo ""
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ CLEANUP COMPLETE${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Backup location:${NC} ./backups/$BACKUP_NAME"
echo ""
echo -e "${BLUE}Server now has ONLY:${NC}"
echo "  - Configuration files (.env, docker-compose)"
echo "  - Deployment script"
echo "  - Docker volumes (data persists)"
echo ""
echo -e "${YELLOW}To restore from backup:${NC}"
echo "  scp ./backups/$BACKUP_NAME $SERVER:/tmp/"
echo "  ssh $SERVER 'cd /opt && tar -xzf /tmp/$BACKUP_NAME'"
echo ""
ENDSSH
