# Stage 1: Build Go backend
FROM golang:1.21-alpine AS go-builder

WORKDIR /build
COPY form-backend/go.mod form-backend/go.sum* ./
RUN go mod download || true
COPY form-backend/main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o form-backend .

# Stage 2: Runtime with Nginx + Go backend
FROM nginx:alpine

# Copy custom nginx config
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy landing page HTML files
COPY index.html /usr/share/nginx/html/
COPY impressum.html /usr/share/nginx/html/
COPY datenschutz.html /usr/share/nginx/html/

# Copy SEO files
COPY robots.txt /usr/share/nginx/html/
COPY sitemap.xml /usr/share/nginx/html/

# Copy favicons and manifest
COPY favicon.ico /usr/share/nginx/html/
COPY favicon.svg /usr/share/nginx/html/
COPY favicon-16x16.png /usr/share/nginx/html/
COPY favicon-32x32.png /usr/share/nginx/html/
COPY apple-touch-icon.png /usr/share/nginx/html/
COPY android-chrome-192x192.png /usr/share/nginx/html/
COPY android-chrome-512x512.png /usr/share/nginx/html/
COPY site.webmanifest /usr/share/nginx/html/

# Copy Go backend binary
COPY --from=go-builder /build/form-backend /usr/local/bin/form-backend

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Environment variables for email configuration
ENV TO_EMAIL=service@aitrail.ch
ENV SMTP_HOST=
ENV SMTP_PORT=
ENV SMTP_USER=
ENV SMTP_PASS=

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Expose port
EXPOSE 80

# Start both services via entrypoint
ENTRYPOINT ["/entrypoint.sh"]
