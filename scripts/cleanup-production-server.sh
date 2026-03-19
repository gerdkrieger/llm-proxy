#!/bin/bash
# =============================================================================
# Production Server Cleanup Script
# =============================================================================
# Removes ALL unnecessary files from production server
# Keeps ONLY: .env + docker-compose.yml + Docker volumes
# =============================================================================

set -e  # Exit on error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "🧹 Production Server Cleanup"
echo "============================"
echo ""

# Check if running as root or with sudo
if [ "$EUID" -eq 0 ]; then 
  echo -e "${RED}❌ Don't run this script as root!${NC}"
  echo "Run as: ./cleanup-production-server.sh"
  exit 1
fi

# Confirm deployment directory
DEPLOY_DIR="/opt/llm-proxy"
echo -e "${YELLOW}⚠️  This will clean: $DEPLOY_DIR${NC}"
echo ""
echo "Files that will be DELETED:"
echo "  ❌ All source code (cmd/, internal/, pkg/, admin-ui/)"
echo "  ❌ All build tools (Makefile, go.mod, go.sum)"
echo "  ❌ All Git data (.git/, .gitlab/)"
echo "  ❌ All documentation (docs/, README.md)"
echo "  ❌ All test files (tests/)"
echo "  ❌ All secrets (OPENWEBUI_TOKEN_*.txt)"
echo "  ❌ Development configs (.env.local, .env.example)"
echo ""
echo "Files that will be PRESERVED:"
echo "  ✅ .env (backed up first)"
echo "  ✅ Docker volumes (llm-proxy-postgres-data, llm-proxy-redis-data)"
echo ""
read -p "Continue? (yes/no): " confirm

if [ "$confirm" != "yes" ]; then
  echo "Aborted."
  exit 0
fi

echo ""
echo "🔍 Step 1: Backup .env"
echo "---------------------"

# Create backup directory
BACKUP_DIR="/opt/llm-proxy-backup"
if [ ! -d "$BACKUP_DIR" ]; then
  sudo mkdir -p "$BACKUP_DIR"
  sudo chown $USER:$USER "$BACKUP_DIR"
fi

# Backup .env with timestamp
TIMESTAMP=$(date +%Y%m%d-%H%M%S)
if [ -f "$DEPLOY_DIR/.env" ]; then
  cp "$DEPLOY_DIR/.env" "$BACKUP_DIR/.env.$TIMESTAMP"
  echo -e "${GREEN}✅ Backed up: $BACKUP_DIR/.env.$TIMESTAMP${NC}"
else
  echo -e "${YELLOW}⚠️  No .env found to backup${NC}"
fi

# Also backup docker-compose.yml
if [ -f "$DEPLOY_DIR/docker-compose.yml" ]; then
  cp "$DEPLOY_DIR/docker-compose.yml" "$BACKUP_DIR/docker-compose.yml.$TIMESTAMP"
  echo -e "${GREEN}✅ Backed up: $BACKUP_DIR/docker-compose.yml.$TIMESTAMP${NC}"
fi

echo ""
echo "🛑 Step 2: Stop containers"
echo "--------------------------"
cd "$DEPLOY_DIR"
if [ -f "docker-compose.yml" ] || [ -f "docker-compose.dev.yml" ]; then
  docker-compose down || true
  echo -e "${GREEN}✅ Containers stopped${NC}"
else
  echo -e "${YELLOW}⚠️  No docker-compose.yml found${NC}"
fi

echo ""
echo "📦 Step 3: List Docker volumes (will be preserved)"
echo "--------------------------------------------------"
docker volume ls | grep llm-proxy || echo "No llm-proxy volumes found"

echo ""
echo "🗑️  Step 4: Delete unnecessary files"
echo "------------------------------------"

# List files to be deleted
echo "Files marked for deletion:"
cd "$DEPLOY_DIR"
ls -la

echo ""
read -p "Delete these files? (yes/no): " confirm_delete

if [ "$confirm_delete" != "yes" ]; then
  echo "Aborted."
  exit 0
fi

# Delete everything in deployment directory
cd /opt
sudo rm -rf "$DEPLOY_DIR"/*
sudo rm -rf "$DEPLOY_DIR"/.[!.]*  # Delete hidden files

echo -e "${GREEN}✅ Deleted all files${NC}"

echo ""
echo "📁 Step 5: Create clean directory structure"
echo "-------------------------------------------"

sudo mkdir -p "$DEPLOY_DIR"
sudo chown $USER:$USER "$DEPLOY_DIR"

echo -e "${GREEN}✅ Created: $DEPLOY_DIR${NC}"

echo ""
echo "📋 Step 6: Restore essential files"
echo "----------------------------------"

# Restore .env
if [ -f "$BACKUP_DIR/.env.$TIMESTAMP" ]; then
  cp "$BACKUP_DIR/.env.$TIMESTAMP" "$DEPLOY_DIR/.env"
  echo -e "${GREEN}✅ Restored: .env${NC}"
fi

echo ""
echo "✅ Cleanup Complete!"
echo "===================="
echo ""
echo "📋 Next Steps:"
echo ""
echo "1. Create docker-compose.yml on server:"
echo "   nano $DEPLOY_DIR/docker-compose.yml"
echo ""
echo "2. Use this production-ready template:"
echo "   (Pull images from GitLab Registry)"
echo ""
echo "3. Login to GitLab Registry:"
echo "   docker login registry.gitlab.com"
echo ""
echo "4. Pull images:"
echo "   cd $DEPLOY_DIR"
echo "   docker-compose pull"
echo ""
echo "5. Start services:"
echo "   docker-compose up -d"
echo ""
echo "6. Verify:"
echo "   docker-compose ps"
echo "   docker-compose logs -f backend"
echo ""
echo "📦 Docker Volumes (preserved):"
docker volume ls | grep llm-proxy || echo "  (none)"
echo ""
echo "💾 Backups saved to:"
echo "  $BACKUP_DIR/.env.$TIMESTAMP"
echo ""
echo "🎯 Production server is now clean!"
echo "   Only .env + docker-compose.yml should be on server"
echo ""
