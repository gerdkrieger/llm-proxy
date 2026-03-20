#!/bin/bash
# =============================================================================
# MANUAL DATABASE MIGRATION TOOL
# =============================================================================
# Run database migrations manually (for development or emergency fixes)
# =============================================================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Configuration
COMMAND="${1:-help}"
VERSION_ARG="${2}"
SERVER="${SERVER:-openweb}"
MIGRATIONS_LOCAL="/home/krieger/Sites/golang-projekte/llm-proxy/migrations"
MIGRATIONS_SERVER="/opt/llm-proxy/migrations"

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

show_help() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}  DATABASE MIGRATION TOOL${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    echo -e "${GREEN}Usage:${NC}"
    echo "  $0 <command> [args]"
    echo ""
    echo -e "${GREEN}Commands:${NC}"
    echo "  ${YELLOW}status${NC}               Show current migration version"
    echo "  ${YELLOW}pending${NC}              List pending migrations"
    echo "  ${YELLOW}up${NC}                   Apply all pending migrations"
    echo "  ${YELLOW}up [N]${NC}               Apply next N migrations"
    echo "  ${YELLOW}down${NC}                 Rollback last migration"
    echo "  ${YELLOW}down [N]${NC}             Rollback last N migrations"
    echo "  ${YELLOW}force [VERSION]${NC}      Force set version (DANGEROUS!)"
    echo "  ${YELLOW}create [NAME]${NC}        Create new migration files"
    echo "  ${YELLOW}sync${NC}                 Sync migrations to server"
    echo ""
    echo -e "${GREEN}Examples:${NC}"
    echo "  $0 status                    # Show current version"
    echo "  $0 pending                   # List pending migrations"
    echo "  $0 up                        # Apply all migrations"
    echo "  $0 up 1                      # Apply next 1 migration"
    echo "  $0 down                      # Rollback last migration"
    echo "  $0 create add_user_roles     # Create new migration"
    echo "  $0 sync                      # Copy migrations to server"
    echo ""
    echo -e "${GREEN}Environment:${NC}"
    echo "  SERVER=openweb               # SSH host (default: openweb)"
    echo ""
}

run_migration_on_server() {
    local cmd="$1"
    local args="$2"
    
    ssh $SERVER "docker run --rm \
        --network llm-proxy-network \
        -v $MIGRATIONS_SERVER:/migrations \
        migrate/migrate:v4.17.0 \
        -path=/migrations \
        -database=\"postgres://\$DB_USER:\$DB_PASSWORD@llm-proxy-postgres:5432/\$DB_NAME?sslmode=disable\" \
        $cmd $args"
}

# =============================================================================
# COMMANDS
# =============================================================================

case "$COMMAND" in
    status)
        echo -e "${BLUE}Checking migration status on: $SERVER${NC}"
        echo ""
        run_migration_on_server "version" ""
        ;;
    
    pending)
        echo -e "${BLUE}Checking pending migrations on: $SERVER${NC}"
        echo ""
        # Get current version
        CURRENT=$(ssh $SERVER "docker run --rm \
            --network llm-proxy-network \
            -v $MIGRATIONS_SERVER:/migrations \
            migrate/migrate:v4.17.0 \
            -path=/migrations \
            -database=\"postgres://\$DB_USER:\$DB_PASSWORD@llm-proxy-postgres:5432/\$DB_NAME?sslmode=disable\" \
            version 2>&1 | grep -oP '(?<=version )\d+' || echo 0")
        
        echo -e "${GREEN}Current version: $CURRENT${NC}"
        echo ""
        echo -e "${YELLOW}Pending migrations:${NC}"
        
        # List migration files higher than current version
        ssh $SERVER "cd $MIGRATIONS_SERVER && ls -1 *.up.sql | while read file; do
            version=\$(echo \$file | grep -oP '^\d+')
            if [ \"\$version\" -gt \"$CURRENT\" ]; then
                echo \"  - \$file (version \$version)\"
            fi
        done"
        ;;
    
    up)
        if [ -z "$VERSION_ARG" ]; then
            echo -e "${BLUE}Applying all pending migrations on: $SERVER${NC}"
            echo ""
            run_migration_on_server "up" ""
        else
            echo -e "${BLUE}Applying next $VERSION_ARG migration(s) on: $SERVER${NC}"
            echo ""
            run_migration_on_server "up" "$VERSION_ARG"
        fi
        ;;
    
    down)
        if [ -z "$VERSION_ARG" ]; then
            echo -e "${YELLOW}Rolling back last migration on: $SERVER${NC}"
            echo ""
            read -p "Are you sure? (type 'yes' to confirm): " CONFIRM
            if [ "$CONFIRM" != "yes" ]; then
                echo "Aborted."
                exit 0
            fi
            run_migration_on_server "down" "1"
        else
            echo -e "${YELLOW}Rolling back last $VERSION_ARG migration(s) on: $SERVER${NC}"
            echo ""
            read -p "Are you sure? (type 'yes' to confirm): " CONFIRM
            if [ "$CONFIRM" != "yes" ]; then
                echo "Aborted."
                exit 0
            fi
            run_migration_on_server "down" "$VERSION_ARG"
        fi
        ;;
    
    force)
        if [ -z "$VERSION_ARG" ]; then
            echo -e "${RED}Error: Version argument required${NC}"
            echo "Usage: $0 force VERSION"
            exit 1
        fi
        
        echo -e "${RED}⚠ WARNING: Force setting version to $VERSION_ARG${NC}"
        echo -e "${RED}This should only be used after manually fixing a dirty migration!${NC}"
        echo ""
        read -p "Are you sure? (type 'yes' to confirm): " CONFIRM
        if [ "$CONFIRM" != "yes" ]; then
            echo "Aborted."
            exit 0
        fi
        
        run_migration_on_server "force" "$VERSION_ARG"
        ;;
    
    create)
        if [ -z "$VERSION_ARG" ]; then
            echo -e "${RED}Error: Migration name required${NC}"
            echo "Usage: $0 create MIGRATION_NAME"
            exit 1
        fi
        
        # Find next version number
        LAST_VERSION=$(ls -1 $MIGRATIONS_LOCAL/*.up.sql 2>/dev/null | tail -1 | grep -oP '^\d+' || echo 0)
        NEXT_VERSION=$(printf "%06d" $((LAST_VERSION + 1)))
        
        UP_FILE="$MIGRATIONS_LOCAL/${NEXT_VERSION}_${VERSION_ARG}.up.sql"
        DOWN_FILE="$MIGRATIONS_LOCAL/${NEXT_VERSION}_${VERSION_ARG}.down.sql"
        
        # Create up migration
        cat > "$UP_FILE" <<EOF
-- Migration: ${VERSION_ARG}
-- Created: $(date)
-- Description: TODO: Add description

BEGIN;

-- TODO: Add your migration SQL here
-- Example:
-- ALTER TABLE users ADD COLUMN email VARCHAR(255);

COMMIT;
EOF
        
        # Create down migration
        cat > "$DOWN_FILE" <<EOF
-- Rollback: ${VERSION_ARG}
-- Created: $(date)

BEGIN;

-- TODO: Add rollback SQL here
-- Example:
-- ALTER TABLE users DROP COLUMN email;

COMMIT;
EOF
        
        echo -e "${GREEN}✓ Created migration files:${NC}"
        echo "  $UP_FILE"
        echo "  $DOWN_FILE"
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo "  1. Edit the SQL files"
        echo "  2. Test locally"
        echo "  3. Commit to git"
        echo "  4. Sync to server: $0 sync"
        ;;
    
    sync)
        echo -e "${BLUE}Syncing migrations to server: $SERVER${NC}"
        echo ""
        
        # Create migrations directory on server if not exists
        ssh $SERVER "mkdir -p $MIGRATIONS_SERVER"
        
        # Rsync migrations
        rsync -avz --progress \
            $MIGRATIONS_LOCAL/*.sql \
            $SERVER:$MIGRATIONS_SERVER/
        
        echo ""
        echo -e "${GREEN}✓ Migrations synced${NC}"
        ;;
    
    help|*)
        show_help
        ;;
esac
