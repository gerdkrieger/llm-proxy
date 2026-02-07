#!/bin/bash
set -euo pipefail

##############################################################################
# LLM-Proxy Filter Deployment Script
##############################################################################
# Purpose: Deploy new content filters to production database
# Usage: ./scripts/deployment/deploy-filters.sh [--dry-run]
##############################################################################

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
FILTERS_DIR="${PROJECT_ROOT}/migrations/filters"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Server configuration
SERVER="${LLM_PROXY_SERVER:-openweb}"
DB_USER="${LLM_PROXY_DB_USER:-proxy_user}"
DB_NAME="${LLM_PROXY_DB_NAME:-llm_proxy}"
POSTGRES_CONTAINER="${LLM_PROXY_POSTGRES_CONTAINER:-llm-proxy-postgres}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    
    # Check if filter files exist
    if [ ! -f "${FILTERS_DIR}/enterprise_standard_filters.sql" ]; then
        log_error "Filter SQL file not found: ${FILTERS_DIR}/enterprise_standard_filters.sql"
        exit 1
    fi
    
    # Check SSH access
    if ! ssh -q ${SERVER} exit; then
        log_error "Cannot connect to server: ${SERVER}"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

backup_existing_filters() {
    log_info "Creating backup of existing filters..."
    
    local backup_file="/tmp/filters_backup_${TIMESTAMP}.sql"
    
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} pg_dump -U ${DB_USER} -d ${DB_NAME} -t content_filters > ${backup_file}" || {
        log_error "Failed to create backup"
        exit 1
    }
    
    log_success "Backup created: ${backup_file}"
    echo "${backup_file}"
}

get_current_filter_count() {
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -c 'SELECT COUNT(*) FROM content_filters;'" | tr -d ' '
}

deploy_filters() {
    local dry_run=$1
    
    if [ "$dry_run" = "true" ]; then
        log_warn "DRY RUN MODE - No changes will be made"
    fi
    
    log_info "Current filter count: $(get_current_filter_count)"
    
    # Copy filter file to server
    log_info "Copying filter file to server..."
    scp -q "${FILTERS_DIR}/enterprise_standard_filters.sql" ${SERVER}:/tmp/filters_${TIMESTAMP}.sql || {
        log_error "Failed to copy filter file"
        exit 1
    }
    
    if [ "$dry_run" = "true" ]; then
        log_info "Would execute SQL import on server"
        log_info "Would verify filter count"
        return 0
    fi
    
    # Copy to container and execute
    log_info "Importing filters into database..."
    ssh ${SERVER} "docker cp /tmp/filters_${TIMESTAMP}.sql ${POSTGRES_CONTAINER}:/tmp/ && \
                   docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -f /tmp/filters_${TIMESTAMP}.sql" || {
        log_error "Failed to import filters"
        exit 1
    }
    
    # Verify
    local new_count=$(get_current_filter_count)
    log_success "Import completed. New filter count: ${new_count}"
    
    # Cleanup temp files on server
    log_info "Cleaning up temporary files..."
    ssh ${SERVER} "rm -f /tmp/filters_${TIMESTAMP}.sql" || log_warn "Failed to cleanup temp file"
}

verify_deployment() {
    log_info "Verifying deployment..."
    
    # Check filter count
    local total_count=$(ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -c 'SELECT COUNT(*) FROM content_filters;'" | tr -d ' ')
    local enabled_count=$(ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -c 'SELECT COUNT(*) FROM content_filters WHERE enabled=true;'" | tr -d ' ')
    
    echo ""
    echo "========================================"
    echo "Filter Deployment Verification"
    echo "========================================"
    echo "Total Filters:   ${total_count}"
    echo "Enabled Filters: ${enabled_count}"
    echo "Disabled:        $((total_count - enabled_count))"
    echo "========================================"
    echo ""
    
    if [ "${total_count}" -lt 40 ]; then
        log_warn "Expected at least 40 filters, found ${total_count}"
        return 1
    fi
    
    log_success "Verification passed!"
    return 0
}

show_filter_stats() {
    log_info "Filter statistics by type:"
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -c \"SELECT filter_type, COUNT(*) as count FROM content_filters GROUP BY filter_type ORDER BY count DESC;\""
}

test_filters() {
    log_info "Testing critical filters..."
    
    echo ""
    echo "Testing Email filter..."
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -c \"SELECT COUNT(*) FROM content_filters WHERE description LIKE '%Email%' AND enabled=true;\"" | tr -d ' '
    
    echo "Testing IBAN filter..."
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -c \"SELECT COUNT(*) FROM content_filters WHERE description LIKE '%IBAN%' AND enabled=true;\"" | tr -d ' '
    
    echo "Testing Credit Card filter..."
    ssh ${SERVER} "docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} -t -c \"SELECT COUNT(*) FROM content_filters WHERE description LIKE '%Kreditkarte%' AND enabled=true;\"" | tr -d ' '
    
    log_success "Filter tests completed"
}

##############################################################################
# Main Script
##############################################################################

main() {
    local dry_run=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --dry-run)
                dry_run=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [--dry-run]"
                echo ""
                echo "Options:"
                echo "  --dry-run    Simulate deployment without making changes"
                echo "  -h, --help   Show this help message"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    echo ""
    echo "========================================"
    echo "LLM-Proxy Filter Deployment"
    echo "========================================"
    echo "Server:     ${SERVER}"
    echo "Database:   ${DB_NAME}"
    echo "Container:  ${POSTGRES_CONTAINER}"
    echo "Dry Run:    ${dry_run}"
    echo "========================================"
    echo ""
    
    # Execute deployment steps
    check_prerequisites
    
    if [ "$dry_run" = "false" ]; then
        backup_file=$(backup_existing_filters)
        log_info "Backup saved to: ${backup_file}"
    fi
    
    deploy_filters "${dry_run}"
    
    if [ "$dry_run" = "false" ]; then
        verify_deployment || {
            log_error "Verification failed!"
            log_info "Rollback: ssh ${SERVER} 'docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} < ${backup_file}'"
            exit 1
        }
        
        show_filter_stats
        test_filters
    fi
    
    echo ""
    log_success "Filter deployment completed successfully!"
    echo ""
    
    if [ "$dry_run" = "false" ]; then
        echo "Next steps:"
        echo "1. Test filters via Admin UI: https://llmproxy.aitrail.ch:3005"
        echo "2. Check filter stats: curl -H 'X-Admin-API-Key: ...' https://llmproxy.aitrail.ch/admin/filters/stats"
        echo "3. Monitor logs: ssh ${SERVER} 'docker logs -f llm-proxy-backend | grep filter'"
        echo ""
        echo "Rollback if needed:"
        echo "  ssh ${SERVER} 'docker exec ${POSTGRES_CONTAINER} psql -U ${DB_USER} -d ${DB_NAME} < ${backup_file}'"
    fi
}

# Run main function
main "$@"
