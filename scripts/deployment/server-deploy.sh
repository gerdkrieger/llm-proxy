#!/bin/bash
# =============================================================================
# SERVER-SIDE DEPLOYMENT SCRIPT WITH DATABASE MIGRATIONS
# =============================================================================
# This script runs ON THE PRODUCTION SERVER to deploy updates
# Includes automatic database migration execution
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
MIGRATIONS_DIR="/opt/llm-proxy/migrations"
MIGRATION_BACKUP=""

# Load .env for database credentials
if [ -f "/opt/llm-proxy/deployments/docker/.env" ]; then
    export $(grep -v '^#' /opt/llm-proxy/deployments/docker/.env | xargs)
fi

# Database connection string
DB_USER="${DB_USER:-proxy_user}"
DB_NAME="${DB_NAME:-llm_proxy}"
DB_HOST="llm-proxy-postgres"
DB_PORT="5432"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  LLM-PROXY DEPLOYMENT WITH MIGRATIONS${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Version: ${VERSION}${NC}"
echo -e "${GREEN}Time: $(date)${NC}"
echo ""

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

# Run SQL query in database
run_sql() {
    local query="$1"
    docker exec llm-proxy-postgres psql -U "$DB_USER" -d "$DB_NAME" -t -A -c "$query" 2>/dev/null
}

# Get current migration version
get_migration_version() {
    local version=$(run_sql "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null || echo "0")
    echo "$version" | tr -d ' '
}

# Check if migration table exists
migration_table_exists() {
    local exists=$(run_sql "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'schema_migrations');" 2>/dev/null || echo "f")
    [ "$exists" = "t" ]
}

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

# Check if PostgreSQL is running
if ! docker ps | grep -q "llm-proxy-postgres"; then
    echo -e "${RED}✗ PostgreSQL container not running${NC}"
    exit 1
fi

# Check database connectivity
if ! run_sql "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${RED}✗ Cannot connect to database${NC}"
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
MIGRATION_BACKUP="$BACKUP_DIR/$(date +%Y%m%d)/postgres-backup-$(date +%H%M%S).sql"
docker exec llm-proxy-postgres pg_dump -U "$DB_USER" "$DB_NAME" > "$MIGRATION_BACKUP" 2>/dev/null

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database backup created: $MIGRATION_BACKUP${NC}"
    echo -e "${GREEN}  Size: $(du -h "$MIGRATION_BACKUP" | cut -f1)${NC}"
else
    echo -e "${RED}✗ Database backup failed!${NC}"
    exit 1
fi

# Save current schema version
CURRENT_VERSION=$(get_migration_version)
echo "$CURRENT_VERSION" > "$BACKUP_DIR/$(date +%Y%m%d)/schema-version-$(date +%H%M%S).txt"
echo -e "${GREEN}✓ Current schema version: $CURRENT_VERSION${NC}"

# Save current image versions
docker-compose -f "$COMPOSE_FILE" images > \
    "$BACKUP_DIR/$(date +%Y%m%d)/images-before-$(date +%H%M%S).txt" 2>/dev/null

echo ""

# =============================================================================
# DATABASE MIGRATIONS
# =============================================================================
echo -e "${BLUE}──────────────────────────────────────${NC}"
echo -e "${BLUE}Running Database Migrations...${NC}"
echo -e "${BLUE}──────────────────────────────────────${NC}"

# Check if migrations directory exists
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${YELLOW}⚠ No migrations directory found at: $MIGRATIONS_DIR${NC}"
    echo -e "${YELLOW}⚠ Skipping migrations step${NC}"
    echo ""
else
    # Check if there are any migration files
    if [ -z "$(ls -A $MIGRATIONS_DIR/*.sql 2>/dev/null)" ]; then
        echo -e "${YELLOW}⚠ No migration files found${NC}"
        echo -e "${YELLOW}⚠ Skipping migrations step${NC}"
        echo ""
    else
        echo -e "${YELLOW}Current schema version: $CURRENT_VERSION${NC}"
        
        # Count migration files
        MIGRATION_COUNT=$(ls -1 $MIGRATIONS_DIR/*.up.sql 2>/dev/null | wc -l)
        echo -e "${YELLOW}Found $MIGRATION_COUNT migration files${NC}"
        echo ""
        
        # Run migrations using golang-migrate
        echo -e "${YELLOW}Applying migrations...${NC}"
        
        docker run --rm \
            --network llm-proxy-network \
            -v "$MIGRATIONS_DIR:/migrations" \
            migrate/migrate:v4.17.0 \
            -path=/migrations \
            -database="postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" \
            up
        
        MIGRATION_EXIT_CODE=$?
        
        if [ $MIGRATION_EXIT_CODE -eq 0 ]; then
            NEW_VERSION=$(get_migration_version)
            echo -e "${GREEN}✓ Migrations completed successfully${NC}"
            echo -e "${GREEN}  Old version: $CURRENT_VERSION${NC}"
            echo -e "${GREEN}  New version: $NEW_VERSION${NC}"
            
            # Check if any migrations were actually applied
            if [ "$NEW_VERSION" = "$CURRENT_VERSION" ]; then
                echo -e "${YELLOW}  (No new migrations to apply)${NC}"
            else
                echo -e "${GREEN}  Applied $(($NEW_VERSION - $CURRENT_VERSION)) migration(s)${NC}"
            fi
        else
            echo -e "${RED}✗ Migrations failed!${NC}"
            echo -e "${RED}Exit code: $MIGRATION_EXIT_CODE${NC}"
            echo ""
            echo -e "${YELLOW}Checking migration status...${NC}"
            
            # Get dirty flag
            DIRTY=$(run_sql "SELECT dirty FROM schema_migrations ORDER BY version DESC LIMIT 1;" 2>/dev/null)
            
            if [ "$DIRTY" = "t" ]; then
                echo -e "${RED}⚠ Database is in DIRTY state!${NC}"
                echo -e "${RED}⚠ Migration was partially applied and failed${NC}"
                echo ""
                echo -e "${YELLOW}Manual intervention required:${NC}"
                echo -e "${YELLOW}1. Check error logs above${NC}"
                echo -e "${YELLOW}2. Fix the migration SQL${NC}"
                echo -e "${YELLOW}3. Manually force version or rollback${NC}"
                echo ""
                echo -e "${YELLOW}To force version (if manually fixed):${NC}"
                echo -e "  docker run --rm --network llm-proxy-network -v $MIGRATIONS_DIR:/migrations migrate/migrate:v4 -path=/migrations -database=\"postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable\" force VERSION"
                echo ""
                echo -e "${YELLOW}To rollback:${NC}"
                echo -e "  docker run --rm --network llm-proxy-network -v $MIGRATIONS_DIR:/migrations migrate/migrate:v4 -path=/migrations -database=\"postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable\" down 1"
            fi
            
            echo ""
            echo -e "${BLUE}──────────────────────────────────────${NC}"
            echo -e "${BLUE}ROLLBACK: Restoring database backup${NC}"
            echo -e "${BLUE}──────────────────────────────────────${NC}"
            
            # Restore from backup
            if [ -f "$MIGRATION_BACKUP" ]; then
                echo -e "${YELLOW}Restoring from: $MIGRATION_BACKUP${NC}"
                docker exec -i llm-proxy-postgres psql -U "$DB_USER" "$DB_NAME" < "$MIGRATION_BACKUP"
                
                if [ $? -eq 0 ]; then
                    echo -e "${GREEN}✓ Database restored from backup${NC}"
                    RESTORED_VERSION=$(get_migration_version)
                    echo -e "${GREEN}✓ Schema version restored: $RESTORED_VERSION${NC}"
                else
                    echo -e "${RED}✗ Database restore failed!${NC}"
                fi
            else
                echo -e "${RED}✗ Backup file not found: $MIGRATION_BACKUP${NC}"
            fi
            
            echo ""
            echo -e "${RED}========================================${NC}"
            echo -e "${RED}  DEPLOYMENT ABORTED${NC}"
            echo -e "${RED}========================================${NC}"
            echo -e "${RED}Reason: Database migration failed${NC}"
            echo -e "${YELLOW}Fix the migration and try again.${NC}"
            echo ""
            exit 1
        fi
    fi
fi

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
echo -e "${BLUE}Database Migration Summary:${NC}"
echo -e "  Previous schema version: $CURRENT_VERSION"
echo -e "  Current schema version:  $(get_migration_version)"
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
