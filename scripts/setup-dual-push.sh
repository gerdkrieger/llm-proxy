#!/bin/bash
# =============================================================================
# SETUP DUAL PUSH FOR GIT REMOTE
# =============================================================================
# Configures git to push to both GitLab and GitHub with a single 'git push'
# =============================================================================

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  GIT DUAL-PUSH SETUP${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if we're in a git repo
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo -e "${YELLOW}Error: Not in a git repository${NC}"
    exit 1
fi

# Get current remotes
echo -e "${BLUE}Current remotes:${NC}"
git remote -v
echo ""

# Prompt for GitLab URL (SSH)
read -p "GitLab URL (SSH): " GITLAB_URL
if [ -z "$GITLAB_URL" ]; then
    echo -e "${YELLOW}Error: GitLab URL required${NC}"
    exit 1
fi

# Prompt for GitHub URL (HTTPS or SSH)
read -p "GitHub URL (HTTPS or SSH): " GITHUB_URL
if [ -z "$GITHUB_URL" ]; then
    echo -e "${YELLOW}Error: GitHub URL required${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}Configuring dual-push...${NC}"

# Set origin fetch to GitLab
git remote set-url origin "$GITLAB_URL"

# Clear existing push URLs
git remote set-url --delete --push origin ".*" 2>/dev/null || true

# Add both push URLs
git remote set-url --add --push origin "$GITLAB_URL"
git remote set-url --add --push origin "$GITHUB_URL"

# Remove old github remote if it exists
git remote remove github 2>/dev/null || true

echo -e "${GREEN}✓ Dual-push configured!${NC}"
echo ""

# Show new configuration
echo -e "${BLUE}New remote configuration:${NC}"
git remote -v
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ SETUP COMPLETE${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""
echo -e "${BLUE}Usage:${NC}"
echo "  git push    # Pushes to BOTH GitLab and GitHub"
echo ""
echo -e "${BLUE}Test it:${NC}"
echo "  git commit --allow-empty -m 'Test dual push'"
echo "  git push"
echo ""
