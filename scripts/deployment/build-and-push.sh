#!/bin/bash
# =============================================================================
# BUILD AND PUSH DOCKER IMAGES TO REGISTRIES
# =============================================================================
# This script builds all Docker images and pushes them to the correct registries:
# - LLM-Proxy components → GitHub Container Registry (ghcr.io)
# - Landing Page → GitLab Container Registry (ONLY!)
# =============================================================================

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REGISTRY="ghcr.io/gerdkrieger"
GITLAB_REGISTRY="registry.gitlab.com/krieger-engineering"
VERSION="${1:-latest}"  # Default to 'latest' if no version specified

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  LLM-PROXY DOCKER BUILD & PUSH${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Version: ${VERSION}${NC}"
echo ""

# Check if logged in to registries
echo -e "${YELLOW}Checking registry authentication...${NC}"

# GitHub
if ! docker info 2>&1 | grep -q "ghcr.io"; then
    echo -e "${YELLOW}Please login to GitHub Container Registry:${NC}"
    echo "docker login ghcr.io -u gerdkrieger"
    exit 1
fi

# GitLab
if ! docker info 2>&1 | grep -q "registry.gitlab.com"; then
    echo -e "${YELLOW}Please login to GitLab Container Registry:${NC}"
    echo "docker login registry.gitlab.com -u <username>"
    exit 1
fi

echo -e "${GREEN}✓ Registry authentication OK${NC}"
echo ""

# =============================================================================
# BUILD & PUSH BACKEND
# =============================================================================
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Building Backend Image...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

docker build \
    -f deployments/docker/Dockerfile \
    -t ${GITHUB_REGISTRY}/llm-proxy-backend:${VERSION} \
    -t ${GITHUB_REGISTRY}/llm-proxy-backend:latest \
    .

echo -e "${GREEN}✓ Backend built${NC}"
echo -e "${YELLOW}Pushing to GitHub Registry...${NC}"

docker push ${GITHUB_REGISTRY}/llm-proxy-backend:${VERSION}
docker push ${GITHUB_REGISTRY}/llm-proxy-backend:latest

echo -e "${GREEN}✓ Backend pushed${NC}"
echo ""

# =============================================================================
# BUILD & PUSH ADMIN-UI
# =============================================================================
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Building Admin-UI Image...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

docker build \
    -f admin-ui/Dockerfile \
    -t ${GITHUB_REGISTRY}/llm-proxy-admin-ui:${VERSION} \
    -t ${GITHUB_REGISTRY}/llm-proxy-admin-ui:latest \
    ./admin-ui

echo -e "${GREEN}✓ Admin-UI built${NC}"
echo -e "${YELLOW}Pushing to GitHub Registry...${NC}"

docker push ${GITHUB_REGISTRY}/llm-proxy-admin-ui:${VERSION}
docker push ${GITHUB_REGISTRY}/llm-proxy-admin-ui:latest

echo -e "${GREEN}✓ Admin-UI pushed${NC}"
echo ""

# =============================================================================
# BUILD & PUSH LANDING PAGE (GITLAB ONLY!)
# =============================================================================
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Building Landing Page Image...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${RED}⚠ Landing Page goes ONLY to GitLab Registry!${NC}"

docker build \
    -f landing/Dockerfile \
    -t ${GITLAB_REGISTRY}/llm-proxy-landing:${VERSION} \
    -t ${GITLAB_REGISTRY}/llm-proxy-landing:latest \
    ./landing

echo -e "${GREEN}✓ Landing Page built${NC}"
echo -e "${YELLOW}Pushing to GitLab Registry...${NC}"

docker push ${GITLAB_REGISTRY}/llm-proxy-landing:${VERSION}
docker push ${GITLAB_REGISTRY}/llm-proxy-landing:latest

echo -e "${GREEN}✓ Landing Page pushed${NC}"
echo ""

# =============================================================================
# SUMMARY
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ ALL IMAGES BUILT AND PUSHED${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${YELLOW}GitHub Registry (ghcr.io):${NC}"
echo "  - ${GITHUB_REGISTRY}/llm-proxy-backend:${VERSION}"
echo "  - ${GITHUB_REGISTRY}/llm-proxy-admin-ui:${VERSION}"
echo ""
echo -e "${YELLOW}GitLab Registry:${NC}"
echo "  - ${GITLAB_REGISTRY}/llm-proxy-landing:${VERSION}"
echo ""
echo -e "${GREEN}Ready to deploy on production server!${NC}"
echo ""
echo -e "${BLUE}Deploy with:${NC}"
echo "  ssh openweb 'cd /opt/llm-proxy && ./scripts/deploy.sh ${VERSION}'"
