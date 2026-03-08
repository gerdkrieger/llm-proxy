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

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/ || exit 1

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
