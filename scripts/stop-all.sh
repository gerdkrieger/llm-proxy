#!/bin/bash

# =============================================================================
# LLM-Proxy Stop Script - Stoppt alle Services
# =============================================================================

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "========================================"
echo "Stopping LLM-Proxy Services"
echo "========================================"
echo ""

# Stop Backend
echo -e "${BLUE}1. Stopping Backend Server...${NC}"
if pgrep -f "bin/llm-proxy" > /dev/null; then
    pkill -f "bin/llm-proxy"
    echo -e "${GREEN}   ✓ Backend stopped${NC}"
else
    echo -e "${YELLOW}   ⚠ Backend was not running${NC}"
fi

# Stop Admin UI
echo -e "${BLUE}2. Stopping Admin UI...${NC}"
if lsof -i :5173 > /dev/null 2>&1; then
    fuser -k 5173/tcp 2>/dev/null
    echo -e "${GREEN}   ✓ Admin UI stopped${NC}"
else
    echo -e "${YELLOW}   ⚠ Admin UI was not running${NC}"
fi

# Optional: Docker Services stoppen
echo ""
read -p "Also stop Docker services (PostgreSQL, Redis)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${BLUE}3. Stopping Docker services...${NC}"
    cd deployments/docker
    docker compose down
    cd ../..
    echo -e "${GREEN}   ✓ Docker services stopped${NC}"
else
    echo -e "${YELLOW}   Docker services kept running${NC}"
fi

echo ""
echo "========================================"
echo -e "${GREEN}✓ All services stopped${NC}"
echo "========================================"
