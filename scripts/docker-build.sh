#!/bin/bash

# =============================================================================
# Docker Build Script for LLM-Proxy with OCR Support
# =============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}================================================${NC}"
echo -e "${BLUE}Building LLM-Proxy Docker Image with OCR Support${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_ROOT"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}Error: Docker is not running${NC}"
    exit 1
fi

# Image details
IMAGE_NAME="${IMAGE_NAME:-llm-proxy}"
IMAGE_TAG="${IMAGE_TAG:-latest}"
FULL_IMAGE_NAME="$IMAGE_NAME:$IMAGE_TAG"

echo -e "${YELLOW}Image Name:${NC} $FULL_IMAGE_NAME"
echo -e "${YELLOW}Build Context:${NC} $PROJECT_ROOT"
echo -e "${YELLOW}Dockerfile:${NC} deployments/docker/Dockerfile"
echo ""

# Build the image
echo -e "${GREEN}Step 1: Building Docker image...${NC}"
docker build \
    -t "$FULL_IMAGE_NAME" \
    -f deployments/docker/Dockerfile \
    --build-arg BUILD_DATE="$(date -u +'%Y-%m-%dT%H:%M:%SZ')" \
    --build-arg VCS_REF="$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')" \
    . || {
        echo -e "${RED}Build failed!${NC}"
        exit 1
    }

echo ""
echo -e "${GREEN}Step 2: Verifying OCR tools installation...${NC}"

# Verify Tesseract
if docker run --rm "$FULL_IMAGE_NAME" sh -c "command -v tesseract" > /dev/null 2>&1; then
    TESSERACT_VERSION=$(docker run --rm "$FULL_IMAGE_NAME" tesseract --version 2>&1 | head -1)
    echo -e "${GREEN}✓${NC} Tesseract: $TESSERACT_VERSION"
else
    echo -e "${RED}✗${NC} Tesseract not found"
fi

# Verify ImageMagick
if docker run --rm "$FULL_IMAGE_NAME" sh -c "command -v convert" > /dev/null 2>&1; then
    IMAGEMAGICK_VERSION=$(docker run --rm "$FULL_IMAGE_NAME" convert -version 2>&1 | head -1)
    echo -e "${GREEN}✓${NC} ImageMagick: $IMAGEMAGICK_VERSION"
else
    echo -e "${RED}✗${NC} ImageMagick not found"
fi

# Verify Ghostscript
if docker run --rm "$FULL_IMAGE_NAME" sh -c "command -v gs" > /dev/null 2>&1; then
    GHOSTSCRIPT_VERSION=$(docker run --rm "$FULL_IMAGE_NAME" gs --version 2>&1)
    echo -e "${GREEN}✓${NC} Ghostscript: $GHOSTSCRIPT_VERSION"
else
    echo -e "${RED}✗${NC} Ghostscript not found"
fi

# Verify Poppler
if docker run --rm "$FULL_IMAGE_NAME" sh -c "command -v pdftoppm" > /dev/null 2>&1; then
    POPPLER_VERSION=$(docker run --rm "$FULL_IMAGE_NAME" pdftoppm -v 2>&1 | head -1)
    echo -e "${GREEN}✓${NC} Poppler: $POPPLER_VERSION"
else
    echo -e "${RED}✗${NC} Poppler not found"
fi

echo ""
echo -e "${GREEN}Step 3: Image details${NC}"
docker images "$IMAGE_NAME" --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}" | head -2

echo ""
echo -e "${BLUE}================================================${NC}"
echo -e "${GREEN}✓ Build completed successfully!${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""
echo -e "${YELLOW}Next steps:${NC}"
echo -e "  1. Run standalone:"
echo -e "     ${BLUE}docker run -d --name llm-proxy -p 8080:8080 --env-file .env $FULL_IMAGE_NAME${NC}"
echo ""
echo -e "  2. Run with Docker Compose:"
echo -e "     ${BLUE}cd deployments/docker && docker-compose up -d${NC}"
echo ""
echo -e "  3. Test health:"
echo -e "     ${BLUE}curl http://localhost:8080/health${NC}"
echo ""
echo -e "  4. View logs:"
echo -e "     ${BLUE}docker logs -f llm-proxy${NC}"
echo ""
