#!/bin/bash
# =============================================================================
# LLM-PROXY LOCAL SETUP VERIFICATION SCRIPT
# =============================================================================
# Run this script to verify your local development environment
# Usage: ./verify-local-setup.sh
# =============================================================================

set -e

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║  🔍 LLM-PROXY LOCAL SETUP VERIFICATION                        ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check function
check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
    else
        echo -e "${RED}✗${NC} $1"
        exit 1
    fi
}

check_exists() {
    if [ -e "$1" ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2 (missing: $1)"
        return 1
    fi
}

echo "┌─────────────────────────────────────────────────────────────┐"
echo "│  Development Tools                                          │"
echo "└─────────────────────────────────────────────────────────────┘"

# Check Go
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    echo -e "${GREEN}✓${NC} Go installed: $GO_VERSION"
else
    echo -e "${RED}✗${NC} Go not found"
fi

# Check Node.js
if command -v node &> /dev/null; then
    NODE_VERSION=$(node --version)
    echo -e "${GREEN}✓${NC} Node.js installed: $NODE_VERSION"
else
    echo -e "${RED}✗${NC} Node.js not found"
fi

# Check npm
if command -v npm &> /dev/null; then
    NPM_VERSION=$(npm --version)
    echo -e "${GREEN}✓${NC} npm installed: v$NPM_VERSION"
else
    echo -e "${RED}✗${NC} npm not found"
fi

# Check Docker
if command -v docker &> /dev/null; then
    DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
    echo -e "${GREEN}✓${NC} Docker installed: $DOCKER_VERSION"
else
    echo -e "${YELLOW}⚠${NC} Docker not found (optional for development)"
fi

echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│  Project Files                                              │"
echo "└─────────────────────────────────────────────────────────────┘"

check_exists ".env" ".env file exists"
check_exists "configs/config.example.yaml" "config.example.yaml exists"
check_exists "docker-compose.dev.yml" "docker-compose.dev.yml exists"
check_exists "admin-ui/package.json" "admin-ui/package.json exists"
check_exists "go.mod" "go.mod exists"

echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│  Go Backend                                                 │"
echo "└─────────────────────────────────────────────────────────────┘"

# Check Go dependencies
echo -n "Checking Go dependencies... "
go mod download &> /dev/null
check "Dependencies downloaded"

# Try to build
echo -n "Building Go binary... "
go build -o bin/llm-proxy ./cmd/server &> /dev/null
check "Build successful"

if [ -f "bin/llm-proxy" ]; then
    SIZE=$(du -h bin/llm-proxy | cut -f1)
    echo -e "${GREEN}✓${NC} Binary created: bin/llm-proxy ($SIZE)"
fi

echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│  Admin-UI Frontend                                          │"
echo "└─────────────────────────────────────────────────────────────┘"

if [ -d "admin-ui/node_modules" ]; then
    echo -e "${GREEN}✓${NC} node_modules exists"
else
    echo -e "${YELLOW}⚠${NC} node_modules missing (run: cd admin-ui && npm install)"
fi

check_exists "admin-ui/src/App.svelte" "App.svelte exists"
check_exists "admin-ui/vite.config.js" "Vite config exists"

echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│  Configuration Check                                        │"
echo "└─────────────────────────────────────────────────────────────┘"

# Check .env variables
REQUIRED_VARS=("DB_HOST" "DB_PORT" "DB_NAME" "DB_USER" "OAUTH_JWT_SECRET" "ADMIN_API_KEYS")
MISSING=0

for VAR in "${REQUIRED_VARS[@]}"; do
    if grep -q "^$VAR=" .env 2>/dev/null; then
        echo -e "${GREEN}✓${NC} $VAR configured in .env"
    else
        echo -e "${RED}✗${NC} $VAR missing in .env"
        MISSING=1
    fi
done

if [ $MISSING -eq 1 ]; then
    echo -e "${YELLOW}⚠${NC} Some required environment variables are missing"
fi

echo ""
echo "┌─────────────────────────────────────────────────────────────┐"
echo "│  Docker Setup                                               │"
echo "└─────────────────────────────────────────────────────────────┘"

if command -v docker &> /dev/null; then
    if docker compose -f docker-compose.dev.yml config --services &> /dev/null; then
        SERVICES=$(docker compose -f docker-compose.dev.yml config --services | wc -l)
        echo -e "${GREEN}✓${NC} Docker Compose config valid ($SERVICES services)"
        docker compose -f docker-compose.dev.yml config --services | while read service; do
            echo "  • $service"
        done
    else
        echo -e "${RED}✗${NC} Docker Compose config invalid"
    fi
else
    echo -e "${YELLOW}⚠${NC} Docker not installed (skipping)"
fi

echo ""
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║  ✅ VERIFICATION COMPLETE                                     ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""
echo "Next steps:"
echo "  1. Start development environment:"
echo "     docker compose -f docker-compose.dev.yml up -d"
echo ""
echo "  2. Access services:"
echo "     • Backend:  http://localhost:8080"
echo "     • Admin-UI: http://localhost:3005"
echo ""
echo "  3. View logs:"
echo "     docker compose -f docker-compose.dev.yml logs -f"
echo ""
