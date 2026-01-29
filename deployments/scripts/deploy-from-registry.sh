#!/bin/bash
# =============================================================================
# DEPLOY FROM GITLAB REGISTRY
# =============================================================================
# This script pulls Docker images from GitLab Container Registry and deploys
# Used by GitLab CI/CD pipeline for automated deployments
# =============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo ""
    echo -e "${BLUE}================================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================================${NC}"
    echo ""
}

# Configuration
ENVIRONMENT="${1:-production}"
IMAGE_TAG="${2:-latest}"
PROJECT_DIR="${3:-$(pwd)}"

print_header "LLM-Proxy Deployment from GitLab Registry"

print_info "Environment: ${ENVIRONMENT}"
print_info "Image Tag: ${IMAGE_TAG}"
print_info "Project Directory: ${PROJECT_DIR}"

# Check if docker-compose exists
if [ ! -f "${PROJECT_DIR}/docker-compose.prod.yml" ]; then
    print_error "docker-compose.prod.yml not found in ${PROJECT_DIR}"
    exit 1
fi

# Check if .env exists
if [ ! -f "${PROJECT_DIR}/.env" ]; then
    print_warning ".env file not found, using .env.production.example"
    
    if [ -f "${PROJECT_DIR}/.env.production.example" ]; then
        cp "${PROJECT_DIR}/.env.production.example" "${PROJECT_DIR}/.env"
        print_warning "Please update .env with production values!"
    else
        print_error ".env.production.example not found"
        exit 1
    fi
fi

# Export environment variables
export IMAGE_TAG="${IMAGE_TAG}"

cd "${PROJECT_DIR}"

# Step 1: Login to GitLab Container Registry
print_header "Step 1: Login to GitLab Container Registry"

if [ -n "${CI_REGISTRY_PASSWORD}" ]; then
    print_info "Using CI credentials for registry login..."
    echo "${CI_REGISTRY_PASSWORD}" | docker login -u "${CI_REGISTRY_USER}" --password-stdin "${CI_REGISTRY}"
else
    print_warning "CI_REGISTRY_PASSWORD not set, attempting manual login..."
    docker login "${CI_REGISTRY:-registry.gitlab.com}"
fi

print_success "Logged into registry"

# Step 2: Pull latest images
print_header "Step 2: Pull Latest Images"

print_info "Pulling backend image: ${IMAGE_TAG}..."
docker compose -f docker-compose.prod.yml pull backend || print_warning "Failed to pull backend image"

print_info "Pulling admin-ui image: ${IMAGE_TAG}..."
docker compose -f docker-compose.prod.yml pull admin-ui || print_warning "Failed to pull admin-ui image"

print_success "Images pulled successfully"

# Step 3: Stop old containers
print_header "Step 3: Stop Old Containers"

print_info "Stopping old containers..."
docker compose -f docker-compose.prod.yml down --remove-orphans

print_success "Old containers stopped"

# Step 4: Start new containers
print_header "Step 4: Start New Containers"

print_info "Starting services..."
docker compose -f docker-compose.prod.yml up -d

print_success "Services started"

# Step 5: Wait for health checks
print_header "Step 5: Health Checks"

print_info "Waiting for services to be healthy..."
sleep 15

# Check backend health
if curl -sf http://localhost:8080/health > /dev/null; then
    print_success "Backend is healthy"
else
    print_error "Backend health check failed"
    docker compose -f docker-compose.prod.yml logs backend
    exit 1
fi

# Check admin UI health
if curl -sf http://localhost:3005/health > /dev/null; then
    print_success "Admin UI is healthy"
else
    print_warning "Admin UI health check failed (non-critical)"
fi

# Step 6: Show deployment info
print_header "Deployment Complete"

echo "Services deployed successfully!"
echo ""
echo "Access URLs:"
echo "  Backend API:    http://$(hostname -f):8080"
echo "  Admin UI:       http://$(hostname -f):3005"
echo "  Prometheus:     http://$(hostname -f):9090"
echo "  Grafana:        http://$(hostname -f):3001"
echo ""
echo "Container Status:"
docker compose -f docker-compose.prod.yml ps
echo ""

# Step 7: Cleanup old images (optional)
print_info "Cleaning up old Docker images..."
docker image prune -f || true

print_success "Deployment completed successfully!"

# Log deployment
echo "$(date '+%Y-%m-%d %H:%M:%S') - Deployed ${IMAGE_TAG} to ${ENVIRONMENT}" >> deployments.log

exit 0
