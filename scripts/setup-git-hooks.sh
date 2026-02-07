#!/bin/bash
# =============================================================================
# SETUP GIT HOOKS - Install and configure Git hooks for LLM-Proxy
# =============================================================================
# This script installs Git hooks that automate deployment workflows:
# - post-commit: Auto-detect and deploy database migrations
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Get script directory and repo root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo ""
echo -e "${BOLD}${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${BLUE}  LLM-PROXY GIT HOOKS SETUP${NC}"
echo -e "${BOLD}${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Change to repo root
cd "$REPO_ROOT"

# Verify we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${RED}Error: Not in a Git repository root${NC}"
    exit 1
fi

# Create hooks directory if it doesn't exist
HOOKS_DIR=".git/hooks"
mkdir -p "$HOOKS_DIR"

# -----------------------------------------------------------------------------
# Install post-commit hook
# -----------------------------------------------------------------------------

echo -e "${YELLOW}Installing Git hooks...${NC}"
echo ""

POST_COMMIT_HOOK="$HOOKS_DIR/post-commit"

if [ -f "$POST_COMMIT_HOOK" ]; then
    echo -e "${YELLOW}⚠ post-commit hook already exists${NC}"
    echo -e "${YELLOW}  Location: $POST_COMMIT_HOOK${NC}"
    echo ""
    read -p "Overwrite existing hook? [y/N] " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${YELLOW}Skipping post-commit hook installation${NC}"
        exit 0
    fi
fi

# Create post-commit hook
cat > "$POST_COMMIT_HOOK" << 'EOFHOOK'
#!/bin/bash
# =============================================================================
# GIT POST-COMMIT HOOK - Auto-detect and deploy database migrations
# =============================================================================
# This hook automatically detects new migration files or filter changes
# in commits and prompts the user to deploy them to production.
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get the root directory of the repo
REPO_ROOT=$(git rev-parse --show-toplevel)

# Change to repo root
cd "$REPO_ROOT"

# Check if auto-deploy script exists
if [ ! -x "$REPO_ROOT/scripts/deployment/auto-deploy-migrations.sh" ]; then
    echo -e "${YELLOW}Warning: Auto-deploy script not found or not executable${NC}"
    exit 0
fi

# Get the files changed in the last commit
CHANGED_FILES=$(git diff-tree --no-commit-id --name-only -r HEAD)

# Check for migration files
MIGRATION_FILES=$(echo "$CHANGED_FILES" | grep -E "^migrations/[0-9]+.*\.(up|down)\.sql$" || true)

# Check for filter files
FILTER_FILES=$(echo "$CHANGED_FILES" | grep -E "^migrations/filters/.*\.(sql|csv)$" || true)

# If no migrations or filters were changed, exit silently
if [ -z "$MIGRATION_FILES" ] && [ -z "$FILTER_FILES" ]; then
    exit 0
fi

# Show what was detected
echo ""
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  DATABASE CHANGES DETECTED IN COMMIT${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

if [ -n "$MIGRATION_FILES" ]; then
    echo -e "${YELLOW}Migration files:${NC}"
    echo "$MIGRATION_FILES" | while read -r file; do
        echo "  • $file"
    done
    echo ""
fi

if [ -n "$FILTER_FILES" ]; then
    echo -e "${YELLOW}Filter files:${NC}"
    echo "$FILTER_FILES" | while read -r file; do
        echo "  • $file"
    done
    echo ""
fi

# Call the auto-deploy script
exec "$REPO_ROOT/scripts/deployment/auto-deploy-migrations.sh"
EOFHOOK

# Make hook executable
chmod +x "$POST_COMMIT_HOOK"

echo -e "${GREEN}✓ post-commit hook installed${NC}"
echo -e "  Location: $POST_COMMIT_HOOK"
echo ""

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------

echo -e "${BOLD}${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${GREEN}  GIT HOOKS SETUP COMPLETE${NC}"
echo -e "${BOLD}${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo ""

echo -e "${BOLD}Installed hooks:${NC}"
echo -e "  ${GREEN}✓${NC} post-commit - Auto-deploy migrations after commit"
echo ""

echo -e "${BOLD}What happens now:${NC}"
echo -e "  1. When you commit a migration file, the hook detects it"
echo -e "  2. You're prompted: 'Deploy to production? [y/N]'"
echo -e "  3. If you choose 'yes', the migration is deployed automatically"
echo -e "  4. Database is backed up before deployment"
echo -e "  5. Backend is restarted after successful deployment"
echo ""

echo -e "${BOLD}Trigger files:${NC}"
echo -e "  • ${YELLOW}migrations/*.up.sql${NC}     - Database migrations"
echo -e "  • ${YELLOW}migrations/*.down.sql${NC}   - Migration rollbacks"
echo -e "  • ${YELLOW}migrations/filters/*.sql${NC} - Content filters"
echo -e "  • ${YELLOW}migrations/filters/*.csv${NC} - Filter bulk imports"
echo ""

echo -e "${BOLD}Manual deployment:${NC}"
echo -e "  ${BLUE}make deploy-migrations${NC}   - Deploy pending migrations"
echo -e "  ${BLUE}make deploy-filters${NC}      - Deploy content filters"
echo -e "  ${BLUE}make deploy-full${NC}         - Full deployment"
echo ""

echo -e "${BOLD}Disable auto-deploy:${NC}"
echo -e "  To temporarily skip: Answer 'N' when prompted"
echo -e "  To disable: ${BLUE}rm .git/hooks/post-commit${NC}"
echo ""

echo -e "${GREEN}Git hooks are now active and ready to use!${NC}"
echo ""
