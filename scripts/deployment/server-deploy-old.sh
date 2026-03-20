#!/bin/bash
# =============================================================================
# SERVER-SIDE DEPLOYMENT SCRIPT
# =============================================================================
# This script runs ON THE PRODUCTION SERVER to deploy updates
# It pulls pre-built images from registries and restarts containers
# =============================================================================

set -e  # Exit on any error

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
VERSION="${1:-latest}"
COMPOSE_FILE="/opt/llm-proxy/deployments/docker/docker-compose.registry-deploy.yml"
BACKUP_DIR="/opt/llm-proxy-backups"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  LLM-PROXY DEPLOYMENT${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Version: ${VERSION}${NC}"
echo -e "${GREEN}Time: $(date)${NC}"
echo ""

# =============================================================================
# PRE-DEPLOYMENT CHECKS
# =============================================================================
echo -e "${YELLOW}Running pre-deployment checks...${NC}"

# Check if compose file exists
if [ ! -f "$COMPOSE_FILE" ]; then
    echo -e "${RED}✗ Compose file not found: $COMPOSE_FILE${NC}"
    exit 1
fi

# Check if .env exists
if [ ! -f "/opt/llm-proxy/deployments/docker/.env" ]; then
    echo -e "${RED}✗ .env file not found${NC}"
    exit 1
fi

# Check if network exists
if ! docker network ls | grep -q "llm-proxy-network"; then
    echo -e "${YELLOW}Creating network llm-proxy-network...${NC}"
    docker network create llm-proxy-network
fi

# Check if volumes exist
for vol in llm-proxy-postgres-data llm-proxy-redis-data llm-proxy-prometheus-data llm-proxy-grafana-data; do
    if ! docker volume ls | grep -q "$vol"; then
        echo -e "${YELLOW}Creating volume $vol...${NC}"
        docker volume create $vol
    fi
done

echo -e "${GREEN}✓ Pre-deployment checks passed${NC}"
echo ""

# =============================================================================
# BACKUP CURRENT STATE
# =============================================================================
echo -e "${YELLOW}Creating backup...${NC}"

# Create backup directory
mkdir -p "$BACKUP_DIR/$(date +%Y%m%d)"

# Backup PostgreSQL
echo -e "${YELLOW}Backing up PostgreSQL...${NC}"
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > \
    "$BACKUP_DIR/$(date +%Y%m%d)/postgres-backup-$(date +%H%M%S).sql" 2>/dev/null || \
    echo -e "${YELLOW}⚠ PostgreSQL backup skipped (container not running)${NC}"

# Save current image versions
docker-compose -f "$COMPOSE_FILE" images > \
    "$BACKUP_DIR/$(date +%Y%m%d)/images-before-$(date +%H%M%S).txt"

echo -e "${GREEN}✓ Backup created${NC}"
echo ""

# =============================================================================
# PULL NEW IMAGES
# =============================================================================
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Pulling images from registries...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

# Login check
echo -e "${YELLOW}Checking registry authentication...${NC}"

# GitHub Container Registry
if ! docker login ghcr.io -u gerdkrieger --password-stdin < /dev/null 2>/dev/null; then
    echo -e "${RED}✗ Not logged in to GitHub Container Registry${NC}"
    echo -e "${YELLOW}Please run: docker login ghcr.io -u gerdkrieger${NC}"
    exit 1
fi

# GitLab Container Registry
if ! docker info 2>&1 | grep -q "registry.gitlab.com"; then
    echo -e "${RED}✗ Not logged in to GitLab Container Registry${NC}"
    echo -e "${YELLOW}Please run: docker login registry.gitlab.com${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Registry authentication OK${NC}"
echo ""

# Pull images
export VERSION="$VERSION"
docker-compose -f "$COMPOSE_FILE" pull

echo -e "${GREEN}✓ Images pulled${NC}"
echo ""

# =============================================================================
# HEALTH CHECK BEFORE DEPLOYMENT
# =============================================================================
echo -e "${YELLOW}Checking current container health...${NC}"

# Save health status
docker ps --filter "name=llm-proxy" --format "table {{.Names}}\t{{.Status}}" > \
    "$BACKUP_DIR/$(date +%Y%m%d)/health-before-$(date +%H%M%S).txt"

echo -e "${GREEN}✓ Health status saved${NC}"
echo ""

# =============================================================================
# DEPLOY
# =============================================================================
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Deploying containers...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

# Deploy with zero-downtime (depends_on ensures proper startup order)
docker-compose -f "$COMPOSE_FILE" up -d --remove-orphans

echo -e "${GREEN}✓ Containers deployed${NC}"
echo ""

# =============================================================================
# POST-DEPLOYMENT HEALTH CHECKS
# =============================================================================
echo -e "${YELLOW}Waiting for containers to become healthy...${NC}"

# Wait for health checks (max 2 minutes)
TIMEOUT=120
ELAPSED=0
INTERVAL=5

while [ $ELAPSED -lt $TIMEOUT ]; do
    UNHEALTHY=$(docker ps --filter "name=llm-proxy" --format "{{.Names}} {{.Status}}" | grep -v "healthy" | wc -l)
    
    if [ "$UNHEALTHY" -eq 0 ]; then
        echo -e "${GREEN}✓ All containers are healthy${NC}"
        break
    fi
    
    echo -e "${YELLOW}Waiting... ($ELAPSED/$TIMEOUT seconds)${NC}"
    sleep $INTERVAL
    ELAPSED=$((ELAPSED + INTERVAL))
done

# Final health check
echo ""
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}DEPLOYMENT STATUS${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"
docker ps --filter "name=llm-proxy" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

# Check for unhealthy containers
UNHEALTHY=$(docker ps --filter "name=llm-proxy" --format "{{.Names}} {{.Status}}" | grep -v "healthy" | grep -v "postgres\|redis\|prometheus\|grafana" || true)

if [ -n "$UNHEALTHY" ]; then
    echo ""
    echo -e "${RED}⚠ WARNING: Some containers are not healthy:${NC}"
    echo "$UNHEALTHY"
    echo ""
    echo -e "${YELLOW}Check logs with:${NC}"
    echo "  docker-compose -f $COMPOSE_FILE logs -f [service-name]"
    echo ""
    echo -e "${YELLOW}Rollback with:${NC}"
    echo "  ./scripts/deployment/server-deploy.sh [previous-version]"
    exit 1
fi

# =============================================================================
# CLEANUP
# =============================================================================
echo ""
echo -e "${YELLOW}Cleaning up old images...${NC}"
docker image prune -f > /dev/null 2>&1

# Keep last 7 days of backups
find "$BACKUP_DIR" -type d -mtime +7 -exec rm -rf {} + 2>/dev/null || true

echo -e "${GREEN}✓ Cleanup complete${NC}"
echo ""

# =============================================================================
# SUCCESS
# =============================================================================
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}✓ DEPLOYMENT SUCCESSFUL${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Version: ${VERSION}${NC}"
echo -e "${GREEN}Time: $(date)${NC}"
echo ""
echo -e "${BLUE}Services:${NC}"
echo "  Backend:    http://localhost:8080"
echo "  Admin UI:   http://localhost:3005"
echo "  Landing:    http://localhost:8090"
echo "  Prometheus: http://localhost:9092"
echo "  Grafana:    http://localhost:3002"
echo ""
echo -e "${BLUE}Public URLs (via Caddy):${NC}"
echo "  https://scrubgate.tech     (Backend API + Admin UI)"
echo "  https://scrubgate.com      (Landing Page)"
echo ""
