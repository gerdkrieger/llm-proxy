#!/bin/bash
set -euo pipefail

##############################################################################
# LLM-Proxy Database Migration Deployment Script
##############################################################################
# Purpose: Deploy database migrations to production
# Usage: ./scripts/deployment/deploy-migrations.sh [migration_number]
##############################################################################

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MIGRATIONS_DIR="${PROJECT_ROOT}/migrations"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Server configuration
SERVER="${LLM_PROXY_SERVER:-openweb}"
DB_USER="${LLM_PROXY_DB_USER:-proxy_user}"
DB_NAME="${LLM_PROXY_DB_NAME:-llm_proxy}"
POSTGRES_CONTAINER="${LLM_PROXY_POSTGRES_CONTAINER:-llm-proxy-postgres}"
BACKEND_CONTAINER="${LLM_PROXY_BACKEND_CONTAINER:-llm-proxy-backend}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

##############################################################################
# Functions
##############################################################################

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check SSH access
    if ! ssh -q ${SERVER} exit; then
        log_error "Cannot connect to server: ${SERVER}"
        exit 1
    fi
    
    # Check if migrations directory exists
    if [ ! -d "${MIGRATIONS_DIR}" ]; then
        log_error "Migrations directory not found: ${MIGRATIONS_DIR}"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

backup_database() {
    log_info "Creating full database backup..."
    
    local backup_file="/tmp/llm_proxy_backup_${TIMESTAMP}.sql"
    
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} pg_dump -U ${DB_USER} ${DB_NAME} > ${backup_file}" || {
        log_error "Failed to create backup"
        exit 1
    }
    
    # Also create a compressed backup
    ssh ${SERVER} "gzip -c ${backup_file} > ${backup_file}.gz"
    
    log_success "Backup created: ${backup_file}.gz"
    echo "${backup_file}"
}

get_current_migration_version() {
    # Try to get from schema_migrations table (if exists)
    local version=$(ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -c \"SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1;\" 2>/dev/null" | tr -d ' ' || echo "unknown")
    echo "${version}"
}

list_pending_migrations() {
    log_info "Scanning for migration files..."
    
    local current_version=$(get_current_migration_version)
    log_info "Current database version: ${current_version}"
    
    # List all .up.sql files
    ls -1 "${MIGRATIONS_DIR}"/*.up.sql 2>/dev/null | sort || {
        log_warn "No migration files found"
        return 0
    }
}

deploy_migration() {
    local migration_file=$1
    local migration_name=$(basename "${migration_file}" .up.sql)
    
    log_info "Deploying migration: ${migration_name}"
    
    # Copy migration file to server
    scp -q "${migration_file}" ${SERVER}:/tmp/${migration_name}.sql || {
        log_error "Failed to copy migration file"
        exit 1
    }
    
    # Execute migration
    log_info "Executing migration..."
    ssh ${SERVER} "docker cp /tmp/${migration_name}.sql ${POSTGRES_CONTAINER}:/tmp/ && \
                   docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -f /tmp/${migration_name}.sql" || {
        log_error "Migration failed!"
        return 1
    }
    
    log_success "Migration ${migration_name} completed"
    
    # Cleanup
    ssh ${SERVER} "rm -f /tmp/${migration_name}.sql"
}

deploy_all_pending() {
    log_info "Deploying all pending migrations..."
    
    local migrations=($(ls -1 "${MIGRATIONS_DIR}"/*.up.sql 2>/dev/null | sort))
    
    if [ ${#migrations[@]} -eq 0 ]; then
        log_warn "No migrations found"
        return 0
    fi
    
    log_info "Found ${#migrations[@]} migration files"
    
    for migration in "${migrations[@]}"; do
        deploy_migration "${migration}" || {
            log_error "Failed to deploy migration: $(basename ${migration})"
            return 1
        }
    done
    
    log_success "All migrations deployed"
}

restart_backend() {
    log_info "Restarting backend to apply changes..."
    
    ssh ${SERVER} "cd /opt/llm-proxy && docker compose -f deployments/docker-compose.openwebui.yml restart ${BACKEND_CONTAINER}" || {
        log_error "Failed to restart backend"
        return 1
    }
    
    log_info "Waiting for backend to start..."
    sleep 10
    
    # Check if backend is healthy
    local health=$(ssh ${SERVER} "docker inspect --format='{{.State.Health.Status}}' ${BACKEND_CONTAINER} 2>/dev/null" || echo "unknown")
    
    if [ "${health}" = "healthy" ] || [ "${health}" = "unknown" ]; then
        log_success "Backend restarted successfully"
    else
        log_warn "Backend health status: ${health}"
    fi
}

verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check database connection
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -c 'SELECT 1;'" > /dev/null || {
        log_error "Database connection failed!"
        return 1
    }
    
    # Check backend logs for errors
    log_info "Checking backend logs for errors..."
    local errors=$(ssh ${SERVER} "docker logs --tail=50 ${BACKEND_CONTAINER} 2>&1 | grep -i 'error\|fatal' | wc -l")
    
    if [ "${errors}" -gt 0 ]; then
        log_warn "Found ${errors} error/fatal messages in backend logs"
        log_info "View logs: ssh ${SERVER} 'docker logs --tail=100 ${BACKEND_CONTAINER}'"
    else
        log_success "No errors in backend logs"
    fi
    
    log_success "Verification completed"
}

show_migration_status() {
    echo ""
    echo "========================================"
    echo "Migration Status"
    echo "========================================"
    
    # Show table counts
    log_info "Table row counts:"
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -c \"
        SELECT 
            schemaname,
            tablename,
            pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
        FROM pg_tables
        WHERE schemaname = 'public'
        ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC
        LIMIT 10;
    \""
    
    echo "========================================"
}

##############################################################################
# Main Script
##############################################################################

main() {
    local migration_number=""
    local skip_backup=false
    local skip_restart=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-backup)
                skip_backup=true
                shift
                ;;
            --skip-restart)
                skip_restart=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [migration_number] [options]"
                echo ""
                echo "Arguments:"
                echo "  migration_number  Specific migration to deploy (e.g., 000007)"
                echo "                    If omitted, deploys all pending migrations"
                echo ""
                echo "Options:"
                echo "  --skip-backup     Skip database backup (NOT recommended)"
                echo "  --skip-restart    Skip backend restart"
                echo "  -h, --help        Show this help message"
                echo ""
                echo "Examples:"
                echo "  $0                    # Deploy all pending migrations"
                echo "  $0 000007             # Deploy specific migration"
                echo "  $0 --skip-restart     # Deploy without restart"
                exit 0
                ;;
            *)
                migration_number=$1
                shift
                ;;
        esac
    done
    
    echo ""
    echo "========================================"
    echo "LLM-Proxy Database Migration Deployment"
    echo "========================================"
    echo "Server:        ${SERVER}"
    echo "Database:      ${DB_NAME}"
    echo "Migration:     ${migration_number:-all pending}"
    echo "Skip Backup:   ${skip_backup}"
    echo "Skip Restart:  ${skip_restart}"
    echo "========================================"
    echo ""
    
    # Confirm deployment
    if [ "${skip_backup}" = "false" ]; then
        read -p "This will modify the production database. Continue? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_warn "Deployment cancelled"
            exit 0
        fi
    fi
    
    # Execute deployment steps
    check_prerequisites
    
    local backup_file=""
    if [ "${skip_backup}" = "false" ]; then
        backup_file=$(backup_database)
    fi
    
    if [ -n "${migration_number}" ]; then
        # Deploy specific migration
        local migration_file="${MIGRATIONS_DIR}/${migration_number}_*.up.sql"
        if [ -f ${migration_file} ]; then
            deploy_migration "${migration_file}" || {
                log_error "Migration deployment failed!"
                if [ -n "${backup_file}" ]; then
                    log_info "Rollback: ssh ${SERVER} 'docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} < ${backup_file}'"
                fi
                exit 1
            }
        else
            log_error "Migration file not found: ${migration_file}"
            exit 1
        fi
    else
        # Deploy all pending migrations
        deploy_all_pending || {
            log_error "Migration deployment failed!"
            if [ -n "${backup_file}" ]; then
                log_info "Rollback: ssh ${SERVER} 'docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} < ${backup_file}'"
            fi
            exit 1
        }
    fi
    
    if [ "${skip_restart}" = "false" ]; then
        restart_backend || log_warn "Backend restart failed, but migration succeeded"
    fi
    
    verify_deployment
    show_migration_status
    
    echo ""
    log_success "Deployment completed successfully!"
    echo ""
    
    if [ -n "${backup_file}" ]; then
        echo "Backup location: ${SERVER}:${backup_file}.gz"
        echo ""
        echo "Rollback if needed:"
        echo "  ssh ${SERVER} 'gunzip ${backup_file}.gz && docker exec -i ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} < ${backup_file}'"
    fi
    
    echo ""
    echo "Next steps:"
    echo "1. Check backend logs: ssh ${SERVER} 'docker logs -f ${BACKEND_CONTAINER}'"
    echo "2. Test API endpoints: curl https://llmproxy.aitrail.ch/health"
    echo "3. Monitor for errors in Live Monitor: https://llmproxy.aitrail.ch:3005"
}

# Run main function
main "$@"
