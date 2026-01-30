#!/bin/bash

# =============================================================================
# LLM-Proxy Complete Start Script - Backend + Admin UI + Frontend
# =============================================================================

set -e

PROJECT_DIR="/home/krieger/Sites/golang-projekte/llm-proxy"
BACKEND_LOG="/tmp/llm-proxy.log"
ADMIN_UI_LOG="/tmp/admin-ui.log"

cd "$PROJECT_DIR"

echo "========================================"
echo "LLM-Proxy Complete Startup"
echo "========================================"
echo ""

# Farben
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 1. Docker Services prüfen
echo -e "${BLUE}1. Checking Docker services...${NC}"
if ! docker ps | grep -q llm-proxy-postgres; then
    echo -e "${YELLOW}   Docker services not running. Starting...${NC}"
    make docker-up
    echo -e "${YELLOW}   Waiting 10 seconds for services to initialize...${NC}"
    sleep 10
else
    echo -e "${GREEN}   ✓ Docker services already running${NC}"
fi
echo ""

# 2. Backend Build
echo -e "${BLUE}2. Building Backend...${NC}"
if [ ! -f "./bin/llm-proxy" ]; then
    make build
else
    echo -e "${GREEN}   ✓ Binary already exists${NC}"
fi
echo ""

# 3. Backend starten
echo -e "${BLUE}3. Starting Backend Server...${NC}"
if pgrep -f "bin/llm-proxy" > /dev/null; then
    echo -e "${YELLOW}   Backend already running. Stopping old instance...${NC}"
    pkill -9 -f "bin/llm-proxy" || true
    sleep 2
fi

./bin/llm-proxy > "$BACKEND_LOG" 2>&1 &
BACKEND_PID=$!
echo -e "${GREEN}   ✓ Backend started (PID: $BACKEND_PID)${NC}"
echo -e "${GREEN}   ✓ Logs: $BACKEND_LOG${NC}"
echo ""

# 4. Warten bis Backend bereit ist
echo -e "${BLUE}4. Waiting for Backend to be ready...${NC}"
for i in {1..15}; do
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}   ✓ Backend is healthy!${NC}"
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# 5. Admin UI Dependencies prüfen
echo -e "${BLUE}5. Checking Admin UI dependencies...${NC}"
if [ ! -d "admin-ui/node_modules" ]; then
    echo -e "${YELLOW}   Installing Admin UI dependencies...${NC}"
    cd admin-ui
    npm install
    cd ..
    echo -e "${GREEN}   ✓ Dependencies installed${NC}"
else
    echo -e "${GREEN}   ✓ Dependencies already installed${NC}"
fi
echo ""

# 6. Admin UI starten
echo -e "${BLUE}6. Starting Admin UI (Svelte)...${NC}"
if lsof -i :5173 > /dev/null 2>&1; then
    echo -e "${YELLOW}   Admin UI port already in use. Stopping old instance...${NC}"
    fuser -k 5173/tcp || true
    sleep 2
fi

cd admin-ui
npm run dev > "$ADMIN_UI_LOG" 2>&1 &
ADMIN_UI_PID=$!
cd ..
echo -e "${GREEN}   ✓ Admin UI started (PID: $ADMIN_UI_PID)${NC}"
echo -e "${GREEN}   ✓ Logs: $ADMIN_UI_LOG${NC}"
echo ""

# 7. Warten bis Admin UI bereit ist
echo -e "${BLUE}7. Waiting for Admin UI to be ready...${NC}"
for i in {1..15}; do
    if curl -s http://localhost:5173 > /dev/null 2>&1; then
        echo -e "${GREEN}   ✓ Admin UI is ready!${NC}"
        break
    fi
    echo -n "."
    sleep 1
done
echo ""

# 8. Status anzeigen
echo -e "${BLUE}8. System Status:${NC}"
HEALTH=$(curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null || echo "unknown")
if [ "$HEALTH" = "ok" ]; then
    echo -e "${GREEN}   ✓ Backend Health: $HEALTH${NC}"
else
    echo -e "${RED}   ✗ Backend Health: $HEALTH${NC}"
fi

if curl -s http://localhost:5173 > /dev/null 2>&1; then
    echo -e "${GREEN}   ✓ Admin UI: Running${NC}"
else
    echo -e "${RED}   ✗ Admin UI: Not responding${NC}"
fi
echo ""

# 9. Informationen anzeigen
echo "========================================"
echo -e "${GREEN}✓ LLM-Proxy System is running!${NC}"
echo "========================================"
echo ""
echo "🚀 Web Interfaces:"
echo "  • Admin UI (Svelte):  http://localhost:5173"
echo "  • Filter UI (HTML):   file://$PROJECT_DIR/filter-management-advanced.html"
echo ""
echo "🔌 Backend Endpoints:"
echo "  • Health Check:   http://localhost:8080/health"
echo "  • API Base:       http://localhost:8080/v1"
echo "  • Admin API:      http://localhost:8080/admin"
echo "  • Metrics:        http://localhost:8080/metrics"
echo ""
echo "💾 Database Services:"
echo "  • PostgreSQL:     localhost:5433"
echo "  • Redis:          localhost:6380"
echo ""
echo "📊 Monitoring:"
echo "  • Prometheus:     http://localhost:9090"
echo "  • Grafana:        http://localhost:3001 (admin/admin)"
echo ""
echo "🔑 Admin API Key:"
echo "  X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"
echo ""
echo "📝 Logs:"
echo "  • Backend:        tail -f $BACKEND_LOG"
echo "  • Admin UI:       tail -f $ADMIN_UI_LOG"
echo ""
echo "🛠️  Useful Commands:"
echo "  • Stop Backend:   pkill -f bin/llm-proxy"
echo "  • Stop Admin UI:  fuser -k 5173/tcp"
echo "  • Stop All:       pkill -f bin/llm-proxy && fuser -k 5173/tcp"
echo "  • Restart:        ./start-all.sh"
echo "  • Health Check:   make health-check"
echo "  • Run Tests:      ./test-all-filters.sh"
echo ""
echo "📦 Current Status:"
FILTERS=$(curl -s http://localhost:8080/admin/filters/stats -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" 2>/dev/null | jq -r '.total_filters' || echo "?")
echo "  • Active Filters: $FILTERS"
echo ""
echo "========================================"

# Optional: Browser automatisch öffnen
read -p "Open Admin UI in Firefox? [Y/n] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    firefox http://localhost:5173 &
    echo -e "${GREEN}✓ Admin UI opened in Firefox${NC}"
    
    # Auch Filter UI anbieten
    read -p "Also open Filter Management UI? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        firefox "$PROJECT_DIR/filter-management-advanced.html" &
        echo -e "${GREEN}✓ Filter UI opened in Firefox${NC}"
    fi
fi

echo ""
echo "Press Ctrl+C to stop following logs, or close this terminal."
echo "Services will continue running in background."
echo ""
echo "Showing Backend logs (Ctrl+C to stop viewing):"
echo "========================================"

# Logs folgen
tail -f "$BACKEND_LOG"
