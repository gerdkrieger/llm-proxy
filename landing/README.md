# LLM-Proxy Landing Page

Static marketing landing page for scrubgate.com

## Overview

This is a simple NGINX-based landing page that serves static HTML/CSS content for the LLM-Proxy product marketing site.

## Structure

```
landing/
├── Dockerfile              # NGINX-based container
├── README.md              # This file
├── index.html             # Main landing page
├── impressum.html         # Legal notice (German)
├── datenschutz.html       # Privacy policy (German)
├── robots.txt             # SEO robots file
├── sitemap.xml            # SEO sitemap
├── site.webmanifest       # PWA manifest
├── 50x.html               # NGINX error page
└── *.png, *.ico, *.svg    # Favicon and app icons
```

## Build & Run

### Local Development

```bash
# Build the image
docker build -t llm-proxy-landing:latest .

# Run locally
docker run -p 8090:80 llm-proxy-landing:latest

# Access at http://localhost:8090
```

### Production Deployment

```bash
# Using docker-compose
docker compose -f docker-compose.landing.yml up -d --build

# Or as part of the main stack
docker compose -f deployments/docker/docker-compose.prod.yml up -d
```

## Deployment Details

- **Domain:** scrubgate.com
- **Port:** 8090 (localhost-only)
- **Reverse Proxy:** Caddy (configured in `/etc/caddy/Caddyfile`)
- **Health Check:** `/health` endpoint
- **Network:** llm-proxy-network (external)

## Caddyfile Configuration

The landing page is served via Caddy reverse proxy:

```caddyfile
scrubgate.com {
    reverse_proxy 127.0.0.1:8090
    encode gzip

    header {
        X-Frame-Options "SAMEORIGIN"
        X-Content-Type-Options "nosniff"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "no-referrer-when-downgrade"
    }
}
```

## SEO Optimization

The landing page includes:
- ✅ Meta tags (Open Graph, Twitter Card)
- ✅ Schema.org JSON-LD
- ✅ Sitemap.xml
- ✅ Robots.txt
- ✅ Canonical URLs
- ✅ PWA manifest
- ✅ Favicons (all sizes)

## Security

- ✅ Port bound to `127.0.0.1` (not publicly exposed)
- ✅ Served via Caddy with automatic HTTPS
- ✅ Security headers configured
- ✅ No sensitive data in static files

## Maintenance

### Update Content

1. Edit HTML files in `landing/`
2. Rebuild the Docker image
3. Restart the container

```bash
docker compose -f docker-compose.landing.yml up -d --build --force-recreate
```

### Health Check

Check if the landing page is healthy:

```bash
curl http://localhost:8090/health
# Should return: healthy
```

## Notes

- The landing page is completely static (no backend)
- All content is German-language focused
- Optimized for SEO and performance
- Responsive design for mobile/tablet/desktop
