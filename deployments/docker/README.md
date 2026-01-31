# LLM-Proxy Docker Deployment mit OCR-Support

## 🎯 Übersicht

Dieses Docker-Setup enthält alle notwendigen Tools für OCR-basierte Dokumenten-Redaktion:

- **Tesseract OCR** (Englisch + Deutsch)
- **ImageMagick** (Bildmanipulation)
- **Ghostscript** (PDF-Verarbeitung)
- **Poppler Utils** (PDF zu Bild Konvertierung)

## 📋 Voraussetzungen

```bash
# Docker & Docker Compose installiert
docker --version
docker-compose --version

# Mindestens Docker 20.10+ und Docker Compose 1.29+
```

## 🚀 Quick Start

### Option 1: Mit Build-Script (Empfohlen)

```bash
# Zum Projekt-Root wechseln
cd ~/Sites/golang-projekte/llm-proxy

# Docker Image bauen mit OCR-Tools
./scripts/docker-build.sh

# Container starten
docker run -d \
  --name llm-proxy \
  -p 8080:8080 \
  -p 9090:9090 \
  --env-file .env \
  llm-proxy:latest

# Logs anschauen
docker logs -f llm-proxy
```

### Option 2: Manuell mit Docker Compose

```bash
cd ~/Sites/golang-projekte/llm-proxy/deployments/docker

# Services starten (Postgres, Redis, Prometheus, Grafana)
docker-compose up -d postgres redis prometheus grafana

# LLM-Proxy Service aktivieren (in docker-compose.yml auskommentieren)
# Dann:
docker-compose up -d llm-proxy
```

### Option 3: Standalone Docker Build

```bash
cd ~/Sites/golang-projekte/llm-proxy

# Image bauen
docker build -t llm-proxy:latest -f deployments/docker/Dockerfile .

# Überprüfen, ob OCR-Tools installiert sind
docker run --rm llm-proxy:latest tesseract --version
docker run --rm llm-proxy:latest convert --version
docker run --rm llm-proxy:latest gs --version
docker run --rm llm-proxy:latest pdftoppm -v
```

## 🔧 Konfiguration

### Umgebungsvariablen (.env)

Erstelle eine `.env` Datei im Projekt-Root:

```bash
# Database
DB_HOST=postgres
DB_PORT=5432
DB_NAME=llm_proxy
DB_USER=proxy_user
DB_PASSWORD=dev_password_2024

# Redis
REDIS_HOST=redis
REDIS_PORT=6379

# Server
SERVER_PORT=8080
ENVIRONMENT=production

# API Keys
CLAUDE_API_KEY=sk-ant-api03-...
OPENAI_API_KEY=sk-proj-...

# Admin Key
ADMIN_API_KEY=admin_dev_key_12345678901234567890123456789012
```

## 🧪 Testen der OCR-Funktionalität

### Im laufenden Container

```bash
# In den Container einsteigen
docker exec -it llm-proxy sh

# Tesseract testen
tesseract --version
tesseract --list-langs

# ImageMagick testen
convert --version

# Ghostscript testen
gs --version

# Temp-Verzeichnis überprüfen
ls -la /tmp/llm-proxy-redaction/
```

### End-to-End Test mit API

```bash
# Test-Bild mit PII erstellen
convert -size 800x400 xc:white \
  -pointsize 18 -draw "text 50,50 'Email: test@example.com'" \
  -pointsize 18 -draw "text 50,80 'Credit Card: 4532-1234-5678-9010'" \
  /tmp/test_document.png

# Base64 enkodieren
BASE64_IMAGE=$(base64 -w 0 /tmp/test_document.png)

# An Container-API senden
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-llm-proxy-..." \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"claude-3-haiku-20240307\",
    \"messages\": [{
      \"role\": \"user\",
      \"content\": [
        {\"type\": \"text\", \"text\": \"Was siehst du in diesem Dokument?\"},
        {\"type\": \"image_url\", \"image_url\": {\"url\": \"data:image/png;base64,$BASE64_IMAGE\"}}
      ]
    }],
    \"max_tokens\": 500
  }"
```

**Erwartetes Ergebnis:** Email und Kreditkarte werden als `[redacted]` angezeigt.

## 📊 Monitoring

### Container-Status

```bash
# Alle Container anzeigen
docker-compose ps

# Logs für spezifischen Service
docker-compose logs -f llm-proxy
docker-compose logs -f postgres
docker-compose logs -f redis

# Ressourcen-Nutzung
docker stats llm-proxy
```

### Health Checks

```bash
# LLM-Proxy Health
curl http://localhost:8080/health

# Prometheus Metrics
curl http://localhost:8080/metrics

# Grafana Dashboard
open http://localhost:3001  # admin/admin
```

## 🛠️ Troubleshooting

### OCR-Tools fehlen im Container

```bash
# Container neu bauen mit --no-cache
docker build --no-cache -t llm-proxy:latest -f deployments/docker/Dockerfile .

# Überprüfen, ob Tools installiert sind
docker run --rm llm-proxy:latest sh -c "apk list | grep tesseract"
docker run --rm llm-proxy:latest sh -c "apk list | grep imagemagick"
```

### Temp-Verzeichnis Fehler

```bash
# Überprüfen, ob Verzeichnis existiert
docker exec llm-proxy ls -la /tmp/llm-proxy-redaction/

# Manuell erstellen, falls nötig
docker exec llm-proxy mkdir -p /tmp/llm-proxy-redaction
docker exec llm-proxy chmod 1777 /tmp/llm-proxy-redaction
```

### Redaction funktioniert nicht

```bash
# Container-Logs überprüfen
docker logs llm-proxy | grep -i "redact"

# Database-Logs überprüfen
docker exec llm-proxy sh -c "psql -U proxy_user -d llm_proxy -c 'SELECT * FROM filter_matches WHERE filter_id IS NULL LIMIT 10;'"
```

### Container startet nicht

```bash
# Detaillierte Fehler anzeigen
docker logs llm-proxy

# Container im interaktiven Modus starten
docker run -it --rm --env-file .env llm-proxy:latest sh

# Konfiguration überprüfen
docker exec llm-proxy cat /app/configs/config.yaml
```

## 🔄 Updates & Maintenance

### Image neu bauen

```bash
cd ~/Sites/golang-projekte/llm-proxy

# Alte Container stoppen
docker-compose down

# Image neu bauen
./scripts/docker-build.sh

# Services neu starten
docker-compose up -d
```

### Datenbank-Migrationen

```bash
# Migrations im Container ausführen
docker exec llm-proxy ./llm-proxy migrate up

# Rollback
docker exec llm-proxy ./llm-proxy migrate down 1
```

### Volumes bereinigen

```bash
# WARNUNG: Löscht alle Daten!
docker-compose down -v

# Nur spezifisches Volume löschen
docker volume rm llm-proxy-postgres-data
docker volume rm llm-proxy-redis-data
```

## 📁 Verzeichnisstruktur im Container

```
/app/
├── llm-proxy          # Haupt-Binary
└── configs/           # Konfigurationsdateien

/tmp/
└── llm-proxy-redaction/  # Temporäre Dateien für OCR/Redaction
```

## 🔐 Sicherheit

### Production Checklist

- [ ] Nicht-Root User verwendet (llmproxy:llmproxy)
- [ ] Secrets über Environment Variables, nicht im Image
- [ ] Health Checks konfiguriert
- [ ] Resource Limits gesetzt
- [ ] Logging konfiguriert
- [ ] OCR Temp-Verzeichnis mit korrekten Permissions (1777)
- [ ] ImageMagick Policy angepasst (falls PDF-Konvertierung Probleme macht)

### ImageMagick Policy (Falls PDF-Probleme)

```bash
# Im Container Policy überprüfen
docker exec llm-proxy cat /etc/ImageMagick-7/policy.xml

# Falls PDF-Konvertierung blockiert ist, Policy anpassen:
# (Normalerweise nicht nötig in Alpine)
```

## 📈 Performance

### Typische Redaction-Zeiten

- **Kleines Bild (800x600):** ~500ms
- **Großes Bild (1920x1080):** ~1-2s
- **PDF (1 Seite):** ~2-3s
- **PDF (Multi-Page):** ~2s pro Seite

### Optimierungen

```yaml
# In docker-compose.yml:
deploy:
  resources:
    limits:
      cpus: '2.0'      # Mehr CPUs für OCR
      memory: 2G        # Mehr RAM für große Dokumente
```

## 🆘 Support

Bei Problemen:

1. Logs überprüfen: `docker logs llm-proxy`
2. Health-Status: `curl http://localhost:8080/health`
3. OCR-Tools testen (siehe oben)
4. Database-Logs: `docker-compose logs postgres`

## 📝 Nächste Schritte

1. **Production Deployment:**
   ```bash
   # Mit HTTPS Reverse Proxy (Traefik/Nginx)
   # Secrets Management (Docker Secrets)
   # Backup-Strategie für Volumes
   ```

2. **Monitoring Setup:**
   ```bash
   # Prometheus + Grafana bereits included
   # Custom Dashboards für OCR Metrics
   ```

3. **Scaling:**
   ```bash
   # Multi-Container Setup mit Load Balancer
   docker-compose up --scale llm-proxy=3
   ```
