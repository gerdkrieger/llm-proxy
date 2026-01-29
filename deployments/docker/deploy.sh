#!/bin/bash
# =============================================================================
# LLM-PROXY DEPLOYMENT SCRIPT
# =============================================================================
# This script helps deploy the LLM-Proxy stack in production
# =============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

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

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    
    print_success "All prerequisites met"
}

# Check environment file
check_env() {
    print_info "Checking environment configuration..."
    
    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        print_warning ".env file not found"
        
        if [ -f "$PROJECT_ROOT/.env.production.example" ]; then
            print_info "Creating .env from .env.production.example..."
            cp "$PROJECT_ROOT/.env.production.example" "$PROJECT_ROOT/.env"
            print_warning "Please update .env with your production values before deploying!"
            exit 1
        else
            print_error "No .env.production.example found"
            exit 1
        fi
    fi
    
    # Check for placeholder values
    if grep -q "CHANGE_ME" "$PROJECT_ROOT/.env"; then
        print_error "Found CHANGE_ME placeholders in .env file"
        print_error "Please update all placeholder values before deploying"
        exit 1
    fi
    
    print_success "Environment configuration OK"
}

# Build images
build_images() {
    print_header "Building Docker Images"
    
    cd "$SCRIPT_DIR"
    
    print_info "Building backend image..."
    docker compose -f docker-compose.prod.yml build backend
    
    print_info "Building admin-ui image..."
    docker compose -f docker-compose.prod.yml build admin-ui
    
    print_success "All images built successfully"
}

# Start services
start_services() {
    print_header "Starting Services"
    
    cd "$SCRIPT_DIR"
    
    print_info "Starting all services..."
    docker compose -f docker-compose.prod.yml up -d
    
    print_success "All services started"
}

# Stop services
stop_services() {
    print_header "Stopping Services"
    
    cd "$SCRIPT_DIR"
    
    print_info "Stopping all services..."
    docker compose -f docker-compose.prod.yml down
    
    print_success "All services stopped"
}

# Show status
show_status() {
    print_header "Service Status"
    
    cd "$SCRIPT_DIR"
    docker compose -f docker-compose.prod.yml ps
}

# Show logs
show_logs() {
    print_header "Service Logs"
    
    cd "$SCRIPT_DIR"
    
    if [ -n "$1" ]; then
        docker compose -f docker-compose.prod.yml logs -f "$1"
    else
        docker compose -f docker-compose.prod.yml logs -f
    fi
}

# Health check
health_check() {
    print_header "Health Check"
    
    print_info "Waiting for services to be healthy..."
    sleep 10
    
    # Check backend
    if curl -sf http://localhost:8080/health > /dev/null; then
        print_success "Backend is healthy"
    else
        print_error "Backend is not responding"
    fi
    
    # Check admin UI
    if curl -sf http://localhost:3005/health > /dev/null; then
        print_success "Admin UI is healthy"
    else
        print_error "Admin UI is not responding"
    fi
    
    # Check Prometheus
    if curl -sf http://localhost:9090/-/healthy > /dev/null; then
        print_success "Prometheus is healthy"
    else
        print_error "Prometheus is not responding"
    fi
    
    # Check Grafana
    if curl -sf http://localhost:3001/api/health > /dev/null; then
        print_success "Grafana is healthy"
    else
        print_error "Grafana is not responding"
    fi
}

# Show access URLs
show_urls() {
    print_header "Access URLs"
    
    echo "Backend API:    http://localhost:8080"
    echo "Admin UI:       http://localhost:3005"
    echo "Prometheus:     http://localhost:9090"
    echo "Grafana:        http://localhost:3001"
    echo ""
    echo "API Health:     http://localhost:8080/health"
    echo "API Metrics:    http://localhost:9091/metrics"
    echo ""
}

# Full deployment
full_deploy() {
    print_header "LLM-Proxy Full Deployment"
    
    check_prerequisites
    check_env
    build_images
    start_services
    health_check
    show_urls
    
    print_success "Deployment complete!"
}

# Update deployment (rebuild and restart)
update_deploy() {
    print_header "Updating Deployment"
    
    check_prerequisites
    check_env
    
    print_info "Stopping services..."
    stop_services
    
    build_images
    start_services
    health_check
    
    print_success "Update complete!"
}

# Backup data
backup_data() {
    print_header "Backing Up Data"
    
    BACKUP_DIR="$PROJECT_ROOT/backups/$(date +%Y%m%d_%H%M%S)"
    mkdir -p "$BACKUP_DIR"
    
    print_info "Creating backup in $BACKUP_DIR..."
    
    # Backup PostgreSQL
    print_info "Backing up PostgreSQL..."
    docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > "$BACKUP_DIR/postgres_backup.sql"
    
    # Backup Redis
    print_info "Backing up Redis..."
    docker exec llm-proxy-redis redis-cli SAVE
    docker cp llm-proxy-redis:/data/dump.rdb "$BACKUP_DIR/redis_backup.rdb"
    
    print_success "Backup saved to $BACKUP_DIR"
}

# Clean up (remove all containers and volumes)
clean_up() {
    print_header "Clean Up"
    
    print_warning "This will remove all containers and volumes!"
    read -p "Are you sure? (yes/no): " confirm
    
    if [ "$confirm" = "yes" ]; then
        cd "$SCRIPT_DIR"
        docker compose -f docker-compose.prod.yml down -v
        print_success "Clean up complete"
    else
        print_info "Clean up cancelled"
    fi
}

# Main menu
show_menu() {
    echo ""
    echo "==================================================="
    echo "       LLM-Proxy Deployment Management"
    echo "==================================================="
    echo ""
    echo "1. Full Deploy (build + start)"
    echo "2. Update Deploy (rebuild + restart)"
    echo "3. Start Services"
    echo "4. Stop Services"
    echo "5. Restart Services"
    echo "6. Show Status"
    echo "7. Show Logs"
    echo "8. Health Check"
    echo "9. Show URLs"
    echo "10. Backup Data"
    echo "11. Clean Up (remove all)"
    echo "12. Exit"
    echo ""
}

# Main script
main() {
    if [ $# -eq 0 ]; then
        # Interactive mode
        while true; do
            show_menu
            read -p "Select option: " choice
            
            case $choice in
                1) full_deploy ;;
                2) update_deploy ;;
                3) start_services ;;
                4) stop_services ;;
                5) stop_services && start_services ;;
                6) show_status ;;
                7) show_logs ;;
                8) health_check ;;
                9) show_urls ;;
                10) backup_data ;;
                11) clean_up ;;
                12) exit 0 ;;
                *) print_error "Invalid option" ;;
            esac
            
            echo ""
            read -p "Press Enter to continue..."
        done
    else
        # Command line mode
        case "$1" in
            deploy) full_deploy ;;
            update) update_deploy ;;
            start) start_services ;;
            stop) stop_services ;;
            restart) stop_services && start_services ;;
            status) show_status ;;
            logs) shift; show_logs "$@" ;;
            health) health_check ;;
            urls) show_urls ;;
            backup) backup_data ;;
            clean) clean_up ;;
            *)
                echo "Usage: $0 {deploy|update|start|stop|restart|status|logs|health|urls|backup|clean}"
                exit 1
                ;;
        esac
    fi
}

# Run main
main "$@"
