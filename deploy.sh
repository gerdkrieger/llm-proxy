#!/bin/bash
# =============================================================================
# LLM-PROXY DEPLOYMENT SCRIPT
# =============================================================================
# Automatisches Deployment zu Production Server
# Baut Docker Images lokal und überträgt sie via SSH
# =============================================================================

set -euo pipefail

# Farben für Output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Server Configuration
SERVER_HOST="openweb"
SERVER_USER="root"
SERVER_PATH="/opt/llm-proxy/deployments"

# Image Names
BACKEND_IMAGE="llm-proxy-backend:latest"
ADMIN_UI_IMAGE="llm-proxy-admin-ui:latest"

# =============================================================================
# Helper Functions
# =============================================================================

log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    # Check if docker is available
    if ! command -v docker &> /dev/null; then
        log_error "Docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if ssh is available
    if ! command -v ssh &> /dev/null; then
        log_error "SSH is not installed or not in PATH"
        exit 1
    fi
    
    # Check if we can connect to server
    if ! ssh -o ConnectTimeout=5 "${SERVER_HOST}" "echo 'SSH connection successful'" &> /dev/null; then
        log_error "Cannot connect to server ${SERVER_HOST}"
        log_info "Make sure SSH key is configured and server is reachable"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

build_backend() {
    log_info "Building backend Docker image..."
    
    docker build \
        -t "${BACKEND_IMAGE}" \
        -f deployments/docker/Dockerfile \
        . || {
            log_error "Backend build failed"
            exit 1
        }
    
    log_success "Backend image built: ${BACKEND_IMAGE}"
}

build_admin_ui() {
    log_info "Building admin-ui Docker image..."
    
    cd admin-ui
    docker build \
        -t "${ADMIN_UI_IMAGE}" \
        -f Dockerfile \
        . || {
            log_error "Admin-UI build failed"
            exit 1
        }
    cd ..
    
    log_success "Admin-UI image built: ${ADMIN_UI_IMAGE}"
}

transfer_images() {
    log_info "Transferring images to production server..."
    
    # Transfer backend
    log_info "Transferring backend image (this may take a while)..."
    docker save "${BACKEND_IMAGE}" | ssh "${SERVER_HOST}" "docker load" || {
        log_error "Failed to transfer backend image"
        exit 1
    }
    log_success "Backend image transferred"
    
    # Transfer admin-ui
    log_info "Transferring admin-ui image..."
    docker save "${ADMIN_UI_IMAGE}" | ssh "${SERVER_HOST}" "docker load" || {
        log_error "Failed to transfer admin-ui image"
        exit 1
    }
    log_success "Admin-UI image transferred"
}

deploy_to_production() {
    log_info "Deploying to production..."
    
    ssh "${SERVER_HOST}" << 'ENDSSH' || {
        log_error "Deployment failed"
        exit 1
    }
        set -e
        cd /opt/llm-proxy/deployments
        
        echo "🛑 Stopping services..."
        docker compose -f docker-compose.openwebui.yml stop backend admin-ui
        
        echo "🔄 Recreating services with new images..."
        docker compose -f docker-compose.openwebui.yml up -d --force-recreate backend admin-ui
        
        echo "⏳ Waiting for services to start (30s)..."
        sleep 30
        
        echo "✅ Deployment complete"
ENDSSH
    
    log_success "Services deployed and restarted"
}

health_check() {
    log_info "Running health checks..."
    
    # Check container status
    log_info "Checking container status..."
    ssh "${SERVER_HOST}" "docker ps --filter 'name=llm-proxy-(backend|admin-ui)' --format 'table {{.Names}}\t{{.Status}}'" || {
        log_warning "Could not get container status"
    }
    
    # Wait a bit for health checks
    sleep 10
    
    # Check backend health endpoint
    log_info "Checking backend health endpoint..."
    HEALTH_STATUS=$(ssh "${SERVER_HOST}" "curl -s http://localhost:8080/health | grep -o '\"status\":\"ok\"' || echo 'FAILED'")
    
    if [[ "$HEALTH_STATUS" == *"ok"* ]]; then
        log_success "Backend is healthy"
    else
        log_error "Backend health check failed"
        log_warning "Check logs: ssh ${SERVER_HOST} 'docker logs llm-proxy-backend --tail 50'"
        exit 1
    fi
    
    # Check if admin-ui is responding
    log_info "Checking admin-ui..."
    ADMIN_STATUS=$(ssh "${SERVER_HOST}" "curl -s -o /dev/null -w '%{http_code}' http://localhost:3005/")
    
    if [[ "$ADMIN_STATUS" == "200" ]]; then
        log_success "Admin-UI is responding"
    else
        log_warning "Admin-UI returned status code: ${ADMIN_STATUS}"
    fi
}

show_deployment_info() {
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_success "🚀 DEPLOYMENT SUCCESSFUL"
    log_success "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    log_info "Production URLs:"
    echo "  🌐 Backend API:    https://llmproxy.aitrail.ch"
    echo "  🌐 Admin UI:       https://llmproxy.aitrail.ch:3005"
    echo "  🌐 Health Check:   https://llmproxy.aitrail.ch/health"
    echo "  🌐 OpenWebUI:      https://chat.aitrail.ch"
    echo ""
    log_info "Useful commands:"
    echo "  📊 View logs:      ssh ${SERVER_HOST} 'docker logs llm-proxy-backend -f'"
    echo "  📊 Check status:   ssh ${SERVER_HOST} 'docker ps'"
    echo "  🔄 Restart:        ssh ${SERVER_HOST} 'cd ${SERVER_PATH} && docker compose -f docker-compose.openwebui.yml restart backend admin-ui'"
    echo ""
}

# =============================================================================
# Main Deployment Flow
# =============================================================================

main() {
    echo ""
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log_info "🚀 LLM-PROXY DEPLOYMENT SCRIPT"
    log_info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # Step 1: Prerequisites
    check_prerequisites
    echo ""
    
    # Step 2: Build Images
    log_info "📦 Step 1/4: Building Docker images..."
    build_backend
    build_admin_ui
    echo ""
    
    # Step 3: Transfer Images
    log_info "🚚 Step 2/4: Transferring images to server..."
    transfer_images
    echo ""
    
    # Step 4: Deploy
    log_info "🚀 Step 3/4: Deploying to production..."
    deploy_to_production
    echo ""
    
    # Step 5: Health Check
    log_info "🏥 Step 4/4: Running health checks..."
    health_check
    echo ""
    
    # Show final info
    show_deployment_info
}

# =============================================================================
# Script Entry Point
# =============================================================================

# Trap errors
trap 'log_error "Deployment failed at line $LINENO"' ERR

# Run main function
main "$@"
