#!/bin/bash
set -euo pipefail

##############################################################################
# LLM-Proxy Full Deployment Script
##############################################################################
# Purpose: Deploy code updates, migrations, and filters to production
# Usage: ./scripts/deployment/deploy-full.sh [--skip-build] [--skip-migrations] [--skip-filters]
##############################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

# Configuration
SERVER="${LLM_PROXY_SERVER:-openweb}"
DEPLOY_PATH="${LLM_PROXY_DEPLOY_PATH:-/opt/llm-proxy}"
COMPOSE_FILE="deployments/docker-compose.openwebui.yml"

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

print_header() {
    echo ""
    echo "========================================"
    echo "$1"
    echo "========================================"
}

check_git_status() {
    log_info "Checking git status..."
    
    if [ -d "${PROJECT_ROOT}/.git" ]; then
        cd "${PROJECT_ROOT}"
        
        # Check for uncommitted changes
        if ! git diff-index --quiet HEAD --; then
            log_warn "You have uncommitted changes!"
            git status --short
            read -p "Continue anyway? [y/N] " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 0
            fi
        fi
        
        # Show current branch and commit
        local branch=$(git rev-parse --abbrev-ref HEAD)
        local commit=$(git rev-parse --short HEAD)
        log_info "Branch: ${branch}, Commit: ${commit}"
    else
        log_warn "Not a git repository"
    fi
}

sync_code() {
    print_header "Step 1: Syncing Code to Server"
    
    log_info "Syncing code to ${SERVER}:${DEPLOY_PATH}..."
    
    # Use rsync to sync code (excluding certain directories)
    rsync -avz --delete \
        --exclude='.git' \
        --exclude='node_modules' \
        --exclude='dist' \
        --exclude='build' \
        --exclude='.env.local' \
        --exclude='logs/' \
        --exclude='*.log' \
        "${PROJECT_ROOT}/" "${SERVER}:${DEPLOY_PATH}/" || {
        log_error "Failed to sync code"
        exit 1
    }
    
    log_success "Code synced successfully"
}

build_containers() {
    print_header "Step 2: Building Docker Containers"
    
    log_info "Building backend container..."
    ssh ${SERVER} "cd ${DEPLOY_PATH} && docker compose -f ${COMPOSE_FILE} build backend" || {
        log_error "Failed to build backend"
        exit 1
    }
    
    log_info "Building admin-ui container..."
    ssh ${SERVER} "cd ${DEPLOY_PATH} && docker compose -f ${COMPOSE_FILE} build admin-ui" || {
        log_error "Failed to build admin-ui"
        exit 1
    }
    
    log_success "Containers built successfully"
}

run_migrations() {
    print_header "Step 3: Running Database Migrations"
    
    log_info "Running migrations..."
    
    # Call migration script
    bash "${SCRIPT_DIR}/deploy-migrations.sh" --skip-restart || {
        log_error "Migration failed!"
        return 1
    }
    
    log_success "Migrations completed"
}

deploy_filters() {
    print_header "Step 4: Deploying Content Filters"
    
    log_info "Deploying filters..."
    
    # Call filter deployment script
    bash "${SCRIPT_DIR}/deploy-filters.sh" || {
        log_warn "Filter deployment failed (non-critical)"
    }
    
    log_success "Filters deployed"
}

restart_services() {
    print_header "Step 5: Restarting Services"
    
    log_info "Restarting backend..."
    ssh ${SERVER} "cd ${DEPLOY_PATH} && docker compose -f ${COMPOSE_FILE} restart backend" || {
        log_error "Failed to restart backend"
        exit 1
    }
    
    log_info "Restarting admin-ui..."
    ssh ${SERVER} "cd ${DEPLOY_PATH} && docker compose -f ${COMPOSE_FILE} restart admin-ui" || {
        log_warn "Failed to restart admin-ui"
    }
    
    log_info "Waiting for services to start..."
    sleep 15
    
    log_success "Services restarted"
}

verify_deployment() {
    print_header "Step 6: Verifying Deployment"
    
    # Check container status
    log_info "Checking container status..."
    ssh ${SERVER} "cd ${DEPLOY_PATH} && docker compose -f ${COMPOSE_FILE} ps" || {
        log_warn "Could not get container status"
    }
    
    # Check backend health
    log_info "Checking backend health..."
    local health_status=$(curl -s -o /dev/null -w "%{http_code}" https://llmproxy.aitrail.ch/health || echo "000")
    
    if [ "${health_status}" = "200" ]; then
        log_success "Backend is healthy (${health_status})"
    else
        log_error "Backend health check failed (${health_status})"
        return 1
    fi
    
    # Check admin-ui health
    log_info "Checking admin-ui..."
    local ui_status=$(curl -s -o /dev/null -w "%{http_code}" https://llmproxy.aitrail.ch:3005 || echo "000")
    
    if [ "${ui_status}" = "200" ]; then
        log_success "Admin UI is healthy (${ui_status})"
    else
        log_warn "Admin UI check returned (${ui_status})"
    fi
    
    # Check for errors in logs
    log_info "Checking logs for errors..."
    local errors=$(ssh ${SERVER} "docker logs --tail=50 llm-proxy-backend 2>&1 | grep -i 'error\|fatal' | wc -l")
    
    if [ "${errors}" -gt 0 ]; then
        log_warn "Found ${errors} error messages in backend logs"
        log_info "View logs: ssh ${SERVER} 'docker logs --tail=100 llm-proxy-backend'"
    else
        log_success "No errors in backend logs"
    fi
    
    log_success "Deployment verified"
}

show_deployment_summary() {
    print_header "Deployment Summary"
    
    echo "Server:          ${SERVER}"
    echo "Deploy Path:     ${DEPLOY_PATH}"
    echo "Backend URL:     https://llmproxy.aitrail.ch"
    echo "Admin UI URL:    https://llmproxy.aitrail.ch:3005"
    echo ""
    
    # Show container status
    log_info "Container Status:"
    ssh ${SERVER} "cd ${DEPLOY_PATH} && docker compose -f ${COMPOSE_FILE} ps --format 'table {{.Name}}\t{{.Status}}'" | grep -E 'llm-proxy|open-webui' || true
    
    echo ""
    log_info "Database Info:"
    ssh ${SERVER} "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \"
        SELECT 
            'Content Filters' as table_name,
            COUNT(*) as row_count,
            pg_size_pretty(pg_total_relation_size('content_filters')) as size
        FROM content_filters
        UNION ALL
        SELECT 
            'Request Logs' as table_name,
            COUNT(*) as row_count,
            pg_size_pretty(pg_total_relation_size('request_logs')) as size
        FROM request_logs;
    \" 2>/dev/null" || log_warn "Could not fetch database info"
}

##############################################################################
# Main Script
##############################################################################

main() {
    local skip_build=false
    local skip_migrations=false
    local skip_filters=false
    local skip_verify=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --skip-build)
                skip_build=true
                shift
                ;;
            --skip-migrations)
                skip_migrations=true
                shift
                ;;
            --skip-filters)
                skip_filters=true
                shift
                ;;
            --skip-verify)
                skip_verify=true
                shift
                ;;
            -h|--help)
                echo "Usage: $0 [options]"
                echo ""
                echo "Full deployment: code + migrations + filters + restart"
                echo ""
                echo "Options:"
                echo "  --skip-build        Skip Docker build step"
                echo "  --skip-migrations   Skip database migrations"
                echo "  --skip-filters      Skip filter deployment"
                echo "  --skip-verify       Skip deployment verification"
                echo "  -h, --help          Show this help message"
                echo ""
                echo "Environment variables:"
                echo "  LLM_PROXY_SERVER        Server hostname (default: openweb)"
                echo "  LLM_PROXY_DEPLOY_PATH   Deployment path (default: /opt/llm-proxy)"
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                exit 1
                ;;
        esac
    done
    
    print_header "LLM-Proxy Full Deployment"
    
    echo "Server:          ${SERVER}"
    echo "Deploy Path:     ${DEPLOY_PATH}"
    echo "Skip Build:      ${skip_build}"
    echo "Skip Migrations: ${skip_migrations}"
    echo "Skip Filters:    ${skip_filters}"
    echo ""
    
    # Confirm deployment
    read -p "Deploy to production? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_warn "Deployment cancelled"
        exit 0
    fi
    
    # Execute deployment
    local start_time=$(date +%s)
    
    check_git_status
    sync_code
    
    if [ "${skip_build}" = "false" ]; then
        build_containers
    fi
    
    if [ "${skip_migrations}" = "false" ]; then
        run_migrations || {
            log_error "Migrations failed - stopping deployment"
            exit 1
        }
    fi
    
    if [ "${skip_filters}" = "false" ]; then
        deploy_filters || log_warn "Filter deployment failed (continuing)"
    fi
    
    restart_services
    
    if [ "${skip_verify}" = "false" ]; then
        verify_deployment || {
            log_error "Verification failed!"
            exit 1
        }
    fi
    
    show_deployment_summary
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    echo ""
    log_success "Full deployment completed in ${duration} seconds!"
    echo ""
    echo "Next steps:"
    echo "1. Test API:         curl https://llmproxy.aitrail.ch/v1/models"
    echo "2. Check Admin UI:   https://llmproxy.aitrail.ch:3005"
    echo "3. Monitor logs:     ssh ${SERVER} 'docker logs -f llm-proxy-backend'"
    echo "4. Live Monitor:     https://llmproxy.aitrail.ch:3005 → Live Monitor"
    echo ""
}

# Run main function
main "$@"
