# LLM-Proxy Landing Page

Professionelle OnePage-Website für LLM-Proxy in Deutsch und Englisch.

## Features

- ✅ **Zweisprachig**: Deutsch/Englisch mit Sprachumschaltung
- ✅ **Responsive**: Mobile-first Design
- ✅ **Modern**: Gradient-Design, smooth Animations
- ✅ **SEO-optimiert**: Meta-Tags, strukturierte Inhalte
- ✅ **Performance**: Optimierte Assets, Caching-Header
- ✅ **Security**: Security-Header vorkonfiguriert

## Zielgruppe

- IT/KI-Enthusiasten und -Professionals
- KMUs mit Compliance-Anforderungen (DSGVO, HIPAA)
- Enterprise-Kunden mit Security-Fokus

## Deployment-Optionen

### Option 1: Direkt mit Nginx (Production)

```bash
# 1. Installiere Nginx
sudo apt install nginx

# 2. Kopiere Dateien
sudo cp index.html /var/www/html/llm-proxy/
sudo cp nginx.conf /etc/nginx/sites-available/llm-proxy

# 3. Aktiviere Site
sudo ln -s /etc/nginx/sites-available/llm-proxy /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### Option 2: Docker (Schnellstart)

```bash
# Build Image
docker build -t llm-proxy-landing .

# Run Container
docker run -d \
  --name llm-proxy-landing \
  -p 8080:80 \
  --restart unless-stopped \
  llm-proxy-landing:latest

# Verfügbar unter: http://localhost:8080
```

### Option 3: Docker Compose

Erstelle `docker-compose.yml`:

```yaml
version: '3.8'

services:
  landing:
    build: .
    container_name: llm-proxy-landing
    ports:
      - "8080:80"
    restart: unless-stopped
    networks:
      - web

networks:
  web:
    external: true
```

Starten:
```bash
docker-compose up -d
```

### Option 4: Static Hosting (GitHub Pages, Netlify, Vercel)

Die `index.html` ist eine standalone Datei und kann direkt gehosted werden:

1. **GitHub Pages**: Pushe nach `gh-pages` Branch
2. **Netlify**: Drag & Drop der `index.html`
3. **Vercel**: Importiere das Verzeichnis

## SSL/HTTPS mit Caddy

Wenn du bereits Caddy als Reverse Proxy nutzt:

```caddyfile
llmproxy.aitrail.ch {
    reverse_proxy llm-proxy-landing:80
    encode gzip
    
    header {
        X-Frame-Options "SAMEORIGIN"
        X-Content-Type-Options "nosniff"
        X-XSS-Protection "1; mode=block"
        Referrer-Policy "no-referrer-when-downgrade"
    }
}
```

## Anpassungen

### Domain ändern

In `nginx.conf` Zeile 2:
```nginx
server_name llmproxy.aitrail.ch;  # Ändere zu deiner Domain
```

### E-Mail-Adresse ändern

In `index.html` suche nach `info@aitrail.ch` und ersetze mit deiner E-Mail.

### GitHub-Link ändern

In `index.html` suche nach `krieger-engineering/llm-proxy` und ersetze mit deinem Repository.

### Farben anpassen

In `index.html` im `<style>` Block unter `:root`:

```css
:root {
    --primary: #2563eb;        /* Hauptfarbe */
    --secondary: #10b981;      /* Sekundärfarbe */
    --danger: #ef4444;         /* Fehlerfarbe */
    /* ... */
}
```

## Performance-Tipps

### Bilder hinzufügen

Falls du Screenshots oder Logos hinzufügen möchtest:

1. Optimiere Bilder mit [TinyPNG](https://tinypng.com/)
2. Nutze WebP Format für moderne Browser
3. Implementiere Lazy Loading:

```html
<img src="screenshot.webp" loading="lazy" alt="LLM-Proxy Dashboard">
```

### CDN nutzen

Für globale Performance kannst du ein CDN wie Cloudflare vorschalten:

1. Domain zu Cloudflare hinzufügen
2. SSL/TLS auf "Full (strict)" setzen
3. Caching Rules aktivieren
4. Minification aktivieren (HTML, CSS, JS)

## Analytics (Optional)

### Plausible Analytics (Privacy-friendly)

```html
<!-- Im <head> einfügen -->
<script defer data-domain="llmproxy.aitrail.ch" src="https://plausible.io/js/script.js"></script>
```

### Google Analytics (falls gewünscht)

```html
<!-- Im <head> einfügen -->
<script async src="https://www.googletagmanager.com/gtag/js?id=GA_MEASUREMENT_ID"></script>
<script>
  window.dataLayer = window.dataLayer || [];
  function gtag(){dataLayer.push(arguments);}
  gtag('js', new Date());
  gtag('config', 'GA_MEASUREMENT_ID');
</script>
```

## SEO-Optimierung

### Sitemap erstellen

Erstelle `sitemap.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://llmproxy.aitrail.ch/</loc>
    <lastmod>2026-03-02</lastmod>
    <changefreq>weekly</changefreq>
    <priority>1.0</priority>
  </url>
</urlset>
```

### robots.txt erstellen

```txt
User-agent: *
Allow: /

Sitemap: https://llmproxy.aitrail.ch/sitemap.xml
```

## Monitoring

### Uptime Monitoring

Nutze einen Service wie:
- [UptimeRobot](https://uptimerobot.com/) (Free)
- [Pingdom](https://www.pingdom.com/)
- [StatusCake](https://www.statuscake.com/)

### Performance Monitoring

Teste mit:
- [PageSpeed Insights](https://pagespeed.web.dev/)
- [GTmetrix](https://gtmetrix.com/)
- [WebPageTest](https://www.webpagetest.org/)

## Support

Bei Fragen oder Problemen:

- **E-Mail**: info@aitrail.ch
- **GitHub Issues**: https://github.com/krieger-engineering/llm-proxy/issues
- **Website**: https://aitrail.ch

## Lizenz

Die Landing Page steht unter MIT Lizenz und kann frei angepasst werden.

---

**Made with ❤️ by [AItrail / Krieger Engineering](https://aitrail.ch)**
