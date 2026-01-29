# =============================================================================
# ADMIN-UI DOCKERFILE - MULTI-STAGE BUILD
# =============================================================================
# Stage 1: Build Svelte app
# Stage 2: Serve with Nginx
# =============================================================================

# -----------------------------------------------------------------------------
# STAGE 1: Build
# -----------------------------------------------------------------------------
FROM node:22-alpine AS builder

# Set working directory
WORKDIR /build

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci --only=production=false

# Copy source code
COPY . .

# Build the application
RUN npm run build

# -----------------------------------------------------------------------------
# STAGE 2: Runtime with Nginx
# -----------------------------------------------------------------------------
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built files from builder
COPY --from=builder /build/dist /usr/share/nginx/html

# Add healthcheck script
RUN apk add --no-cache curl

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:80/ || exit 1

# Expose port
EXPOSE 80

# Labels
LABEL maintainer="LLM-Proxy Team"
LABEL description="Admin UI for LLM-Proxy"
LABEL version="1.0.0"

# Run nginx
CMD ["nginx", "-g", "daemon off;"]

# =============================================================================
# BUILD INSTRUCTIONS
# =============================================================================
# Build image:
#   docker build -t llm-proxy-admin-ui:latest -f Dockerfile .
#
# Run container:
#   docker run -d \
#     --name llm-proxy-admin-ui \
#     -p 3005:80 \
#     llm-proxy-admin-ui:latest
#
# With environment variable:
#   docker run -d \
#     --name llm-proxy-admin-ui \
#     -p 3005:80 \
#     -e VITE_API_BASE_URL=http://backend:8080 \
#     llm-proxy-admin-ui:latest
# =============================================================================
