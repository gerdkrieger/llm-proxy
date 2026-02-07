#!/bin/bash
set -euo pipefail

##############################################################################
# Auto-Deploy Migrations After Git Commit
##############################################################################
# Purpose: Automatically deploy new migrations after they are committed
# Usage: Called by Git hook or manually
##############################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
MIGRATIONS_DIR="${PROJECT_ROOT}/migrations"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${BLUE}[AUTO-DEPLOY]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[AUTO-DEPLOY]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[AUTO-DEPLOY]${NC} $1"
}

log_error() {
    echo -e "${RED}[AUTO-DEPLOY]${NC} $1"
}

check_for_new_migrations() {
    log_info "Checking for new migrations in last commit..."
    
    # Get list of changed files in last commit
    local changed_files=$(git diff-tree --no-commit-id --name-only -r HEAD 2>/dev/null || echo "")
    
    # Check if any migration files changed
    local migration_files=$(echo "${changed_files}" | grep "^migrations/.*\.up\.sql$" || echo "")
    
    if [ -z "${migration_files}" ]; then
        log_info "No new migration files in last commit"
        return 1
    fi
    
    log_info "Found new migrations:"
    echo "${migration_files}" | sed 's/^/  - /'
    echo "${migration_files}"
}

check_for_filter_changes() {
    log_info "Checking for filter changes in last commit..."
    
    local changed_files=$(git diff-tree --no-commit-id --name-only -r HEAD 2>/dev/null || echo "")
    local filter_files=$(echo "${changed_files}" | grep "^migrations/filters/.*\.sql$" || echo "")
    
    if [ -z "${filter_files}" ]; then
        return 1
    fi
    
    log_info "Found filter changes:"
    echo "${filter_files}" | sed 's/^/  - /'
    echo "${filter_files}"
}

prompt_deployment() {
    local type=$1
    
    echo ""
    log_warn "New ${type} detected in commit!"
    echo ""
    echo "Options:"
    echo "  1) Deploy to production now"
    echo "  2) Deploy later manually"
    echo "  3) Skip (not recommended)"
    echo ""
    read -p "Choose option [1/2/3]: " -n 1 -r
    echo
    
    case $REPLY in
        1)
            return 0
            ;;
        2)
            log_info "Deployment postponed"
            log_info "Deploy manually with: make deploy-migrations"
            return 1
            ;;
        3)
            log_warn "Deployment skipped"
            return 1
            ;;
        *)
            log_error "Invalid option"
            return 1
            ;;
    esac
}

deploy_migrations() {
    log_info "Starting migration deployment..."
    
    "${SCRIPT_DIR}/deploy-migrations.sh" || {
        log_error "Migration deployment failed!"
        return 1
    }
    
    log_success "Migrations deployed successfully"
}

deploy_filters() {
    log_info "Starting filter deployment..."
    
    "${SCRIPT_DIR}/deploy-filters.sh" || {
        log_error "Filter deployment failed!"
        return 1
    }
    
    log_success "Filters deployed successfully"
}

##############################################################################
# Main
##############################################################################

main() {
    cd "${PROJECT_ROOT}"
    
    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        log_error "Not in a git repository"
        exit 1
    fi
    
    local deploy_needed=false
    local deploy_type=""
    
    # Check for migrations
    if migration_files=$(check_for_new_migrations); then
        deploy_needed=true
        deploy_type="migrations"
    fi
    
    # Check for filters
    if filter_files=$(check_for_filter_changes); then
        if [ "${deploy_needed}" = "true" ]; then
            deploy_type="migrations and filters"
        else
            deploy_type="filters"
        fi
        deploy_needed=true
    fi
    
    if [ "${deploy_needed}" = "false" ]; then
        log_info "No migrations or filters to deploy"
        exit 0
    fi
    
    # Prompt user
    if prompt_deployment "${deploy_type}"; then
        if [[ "${deploy_type}" == *"migrations"* ]]; then
            deploy_migrations || exit 1
        fi
        
        if [[ "${deploy_type}" == *"filters"* ]]; then
            deploy_filters || exit 1
        fi
        
        echo ""
        log_success "Auto-deployment completed!"
        echo ""
    fi
}

# Only run if called directly (not sourced)
if [ "${BASH_SOURCE[0]}" = "${0}" ]; then
    main "$@"
fi
