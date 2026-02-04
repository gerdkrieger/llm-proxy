#!/bin/bash
# =============================================================================
# QUICK START SCRIPT - DEVELOPMENT MODE
# =============================================================================
# Startet alle benötigten Services für lokale Entwicklung
# =============================================================================

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   LLM-Proxy Development Server${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Check if in correct directory
if [ ! -f "go.mod" ]; then
    echo -e "${YELLOW}⚠️  Nicht im Projekt-Verzeichnis!${NC}"
    echo "Bitte führe das Script aus:"
    echo "  cd /home/krieger/Sites/golang-projekte/llm-proxy"
    echo "  ./start-dev.sh"
    exit 1
fi

# Step 1: Start Docker Services
echo -e "${GREEN}1. Starte Docker Services (PostgreSQL, Redis)...${NC}"
cd deployments/docker
docker compose up -d postgres redis

echo -e "${GREEN}   Warte auf Services...${NC}"
sleep 5

# Check if services are healthy
if docker compose ps postgres | grep -q "healthy"; then
    echo -e "${GREEN}   ✓ PostgreSQL gestartet${NC}"
else
    echo -e "${YELLOW}   ⚠️  PostgreSQL noch nicht ready${NC}"
fi

if docker compose ps redis | grep -q "healthy"; then
    echo -e "${GREEN}   ✓ Redis gestartet${NC}"
else
    echo -e "${YELLOW}   ⚠️  Redis noch nicht ready${NC}"
fi

cd ../..

# Step 2: Check .env file
echo ""
echo -e "${GREEN}2. Prüfe Konfiguration...${NC}"
if [ ! -f ".env" ]; then
    echo -e "${YELLOW}   ⚠️  .env Datei nicht gefunden, erstelle aus Template...${NC}"
    cp .env.production.example .env
    echo -e "${YELLOW}   ⚠️  Bitte .env Datei mit echten Werten aktualisieren!${NC}"
fi

# Step 3: Build and start backend
echo ""
echo -e "${GREEN}3. Starte Backend Server...${NC}"
echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}   Server startet jetzt...${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "${GREEN}Backend API:    ${NC}http://localhost:8080"
echo -e "${GREEN}Health Check:   ${NC}http://localhost:8080/health"
echo -e "${GREEN}Metrics:        ${NC}http://localhost:8080/metrics"
echo -e "${GREEN}Prometheus:     ${NC}http://localhost:9090"
echo -e "${GREEN}Grafana:        ${NC}http://localhost:3001"
echo ""
echo -e "${YELLOW}Admin UI starten (in neuem Terminal):${NC}"
echo -e "  cd admin-ui && npm run dev"
echo ""
echo -e "${YELLOW}Drücke Ctrl+C zum Beenden${NC}"
echo ""

# Start server
go run cmd/server/main.go
