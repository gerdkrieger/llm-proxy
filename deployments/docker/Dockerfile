# =============================================================================
# LLM-PROXY DOCKERFILE - MULTI-STAGE BUILD
# =============================================================================
# Optimized for production: minimal image size, security best practices
# =============================================================================

# -----------------------------------------------------------------------------
# STAGE 1: Build
# -----------------------------------------------------------------------------
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies (cached layer)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a \
    -o llm-proxy \
    ./cmd/server/main.go

# Verify binary
RUN chmod +x llm-proxy && ./llm-proxy --version || true

# -----------------------------------------------------------------------------
# STAGE 2: Runtime
# -----------------------------------------------------------------------------
FROM alpine:latest

# Install minimal runtime dependencies
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    wget

# Create non-root user
RUN addgroup -g 1001 -S llmproxy && \
    adduser -u 1001 -S llmproxy -G llmproxy

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/llm-proxy /app/llm-proxy

# Change ownership
RUN chown -R llmproxy:llmproxy /app

# Switch to non-root user
USER llmproxy

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD wget -q --tries=1 --spider http://localhost:8080/health || exit 1

# Environment
ENV ENVIRONMENT=production

# Run the application
ENTRYPOINT ["/app/llm-proxy"]

# =============================================================================
# BUILD INSTRUCTIONS
# =============================================================================
# Build image:
#   docker build -t llm-proxy:latest -f deployments/docker/Dockerfile .
#
# Run container:
#   docker run -d \
#     --name llm-proxy \
#     -p 8080:8080 \
#     -p 9090:9090 \
#     --env-file .env \
#     llm-proxy:latest
#
# Build with Docker Compose:
#   cd deployments/docker
#   docker-compose build
#   docker-compose up -d
#
# Multi-platform build:
#   docker buildx build \
#     --platform linux/amd64,linux/arm64 \
#     -t llm-proxy:latest \
#     -f deployments/docker/Dockerfile .
