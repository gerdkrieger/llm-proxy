# LLM-Proxy LIVE Server - Wichtige Befehle

**Server:** 68.183.208.213 (dockeronubuntu2204-s-1vcpu-2gb-70gb-intel-fra1-01)  
**SSH:** `ssh openweb`  
**User:** root

---

## 📋 Schnellübersicht - Container Management

### Container Status prüfen
```bash
ssh openweb "docker ps"
ssh openweb "docker ps -a"  # Auch gestoppte Container
```

### Alle LLM-Proxy Container anzeigen
```bash
ssh openweb "docker ps --filter name=llm-proxy --format 'table {{.Names}}\t{{.Status}}\t{{.Ports}}'"
```

### Container starten
```bash
ssh openweb "docker start llm-proxy-backend"
ssh openweb "docker start llm-proxy-admin-ui"
ssh openweb "docker start llm-proxy-postgres"
ssh openweb "docker start llm-proxy-redis"

# Alle auf einmal
ssh openweb "docker start llm-proxy-backend llm-proxy-admin-ui llm-proxy-postgres llm-proxy-redis"
```

### Container stoppen
```bash
ssh openweb "docker stop llm-proxy-backend"
ssh openweb "docker stop llm-proxy-admin-ui"
ssh openweb "docker stop llm-proxy-postgres"
ssh openweb "docker stop llm-proxy-redis"

# Alle auf einmal
ssh openweb "docker stop llm-proxy-backend llm-proxy-admin-ui llm-proxy-postgres llm-proxy-redis"
```

### Container neu starten
```bash
ssh openweb "docker restart llm-proxy-backend"
ssh openweb "docker restart llm-proxy-admin-ui"
```

---

## 📝 Logs anzeigen

### Backend Logs
```bash
# Letzte 50 Zeilen
ssh openweb "docker logs llm-proxy-backend --tail 50"

# Logs live verfolgen
ssh openweb "docker logs llm-proxy-backend -f"

# Nur Fehler
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep -i error"

# Nach bestimmtem Pattern suchen
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep -i 'filter'"
```

### Admin-UI Logs
```bash
ssh openweb "docker logs llm-proxy-admin-ui --tail 50"
```

### Postgres Logs
```bash
ssh openweb "docker logs llm-proxy-postgres --tail 50"
```

### Redis Logs
```bash
ssh openweb "docker logs llm-proxy-redis --tail 50"
```

---

## 🗄️ Datenbank-Befehle

### Datenbank-Verbindung herstellen
```bash
ssh openweb "docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy"
```

### Einzelne SQL-Befehle ausführen
```bash
# Alle Tabellen auflisten
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c '\dt'"

# Content Filters anzeigen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT COUNT(*) FROM content_filters;'"

# Filter Details
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT id, pattern, filter_type, enabled FROM content_filters ORDER BY priority DESC;'"

# OAuth Clients anzeigen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT client_id, name, enabled FROM oauth_clients;'"

# Provider Modelle zählen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT provider_id, COUNT(*) FROM provider_models GROUP BY provider_id;'"
```

### Datenbank Backup erstellen
```bash
# Backup erstellen
ssh openweb "docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > /tmp/llm_proxy_backup_$(date +%Y%m%d_%H%M%S).sql"

# Backup herunterladen
scp openweb:/tmp/llm_proxy_backup_*.sql ~/backups/

# Backup einspielen (VORSICHT!)
cat backup.sql | ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy"
```

### SQL Skript ausführen
```bash
# Lokales SQL-File auf Server hochladen und ausführen
scp my-script.sql openweb:/tmp/
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/my-script.sql"
```

---

## 🔄 Caddy Reverse Proxy

### Caddy Status prüfen
```bash
ssh openweb "systemctl status caddy"
```

### Caddy neu laden (nach Config-Änderungen)
```bash
ssh openweb "systemctl reload caddy.service"
```

### Caddy komplett neu starten
```bash
ssh openweb "systemctl restart caddy.service"
```

### Caddy Konfiguration prüfen
```bash
ssh openweb "cat /etc/caddy/Caddyfile"

# Syntax Check
ssh openweb "caddy validate --config /etc/caddy/Caddyfile"
```

### Caddy Logs
```bash
ssh openweb "journalctl -u caddy -f"
ssh openweb "journalctl -u caddy --since '10 minutes ago'"
```

---

## 🔍 API Tests

### Health Check
```bash
curl https://llmproxy.aitrail.ch/health
# Erwartete Antwort: {"status":"ok","timestamp":"..."}
```

### Admin Endpoints (benötigen API Key)
```bash
# Providers auflisten
curl -s https://llmproxy.aitrail.ch/admin/providers \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" | jq

# Content Filters auflisten
curl -s https://llmproxy.aitrail.ch/admin/filters \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" | jq

# OAuth Clients auflisten
curl -s https://llmproxy.aitrail.ch/admin/clients \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" | jq

# Stats abrufen
curl -s https://llmproxy.aitrail.ch/admin/stats/usage \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" | jq
```

### Metrics
```bash
curl -s http://68.183.208.213:9091/metrics | head -50
```

---

## 🚀 Deployment / Updates

### Neues Backend Image deployen
```bash
# 1. Lokal builden
cd /home/krieger/Sites/golang-projekte/llm-proxy
docker compose -f docker-compose.openwebui.yml build backend

# 2. Image auf Server übertragen
docker save llm-proxy-backend:latest | ssh openweb "docker load"

# 3. Container neu starten
ssh openweb "docker stop llm-proxy-backend && docker rm llm-proxy-backend"
ssh openweb "docker run -d --name llm-proxy-backend --network llm-proxy-network --restart unless-stopped -p 8080:8080 -p 9091:9090 --env-file /path/to/.env llm-proxy-backend:latest"

# ODER mit docker-compose (wenn docker-compose.yml auf Server vorhanden)
ssh openweb "cd /path/to/llm-proxy && docker-compose up -d backend"
```

### Neues Admin-UI Image deployen
```bash
# 1. Lokal mit leerem VITE_API_BASE_URL builden
cd /home/krieger/Sites/golang-projekte/llm-proxy/admin-ui
echo "VITE_API_BASE_URL=" > .env
npm run build

# 2. Docker Image builden
cd ..
docker compose -f docker-compose.openwebui.yml build admin-ui

# 3. Image auf Server übertragen
docker save llm-proxy-admin-ui:latest | ssh openweb "docker load"

# 4. Container neu starten
ssh openweb "docker stop llm-proxy-admin-ui && docker rm llm-proxy-admin-ui"
ssh openweb "docker run -d --name llm-proxy-admin-ui --network llm-proxy-network --restart unless-stopped -p 3005:80 --health-cmd='curl -f http://localhost:80/ || exit 1' --health-interval=30s llm-proxy-admin-ui:latest"
```

### Via GitLab CI/CD
```bash
# Auf GitLab Pipeline auslösen (wenn konfiguriert)
git push origin master
# Pipeline läuft automatisch
```

---

## 🐛 Troubleshooting

### Container läuft nicht / crasht
```bash
# Status prüfen
ssh openweb "docker ps -a | grep llm-proxy-backend"

# Logs anschauen
ssh openweb "docker logs llm-proxy-backend --tail 100"

# Container interaktiv starten (Debug)
ssh openweb "docker run -it --rm llm-proxy-backend:latest /bin/sh"
```

### Datenbank-Verbindung fehlgeschlagen
```bash
# Postgres erreichbar?
ssh openweb "docker exec llm-proxy-backend ping -c 3 postgres"

# Postgres läuft?
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT 1;'"

# Backend ENV prüfen
ssh openweb "docker exec llm-proxy-backend env | grep DATABASE"
```

### Redis-Verbindung fehlgeschlagen
```bash
# Redis erreichbar?
ssh openweb "docker exec llm-proxy-backend ping -c 3 redis"

# Redis läuft?
ssh openweb "docker exec llm-proxy-redis redis-cli ping"
# Erwartete Antwort: PONG
```

### Admin-UI zeigt "Failed to fetch"
```bash
# 1. Backend erreichbar?
curl http://localhost:8080/health

# 2. Admin-UI Container läuft?
ssh openweb "docker ps | grep admin-ui"

# 3. JavaScript API URL prüfen
ssh openweb "docker exec llm-proxy-admin-ui cat /usr/share/nginx/html/assets/index-*.js | grep -o 'const Qn=\"[^\"]*\"'"
# Sollte leer sein: const Qn=""

# 4. Nginx Config prüfen
ssh openweb "docker exec llm-proxy-admin-ui cat /etc/nginx/conf.d/default.conf"
```

### Caddy antwortet nicht
```bash
# Caddy läuft?
ssh openweb "systemctl status caddy"

# Port 443 offen?
ssh openweb "netstat -tulpn | grep :443"

# Zertifikat OK?
ssh openweb "curl -I https://llmproxy.aitrail.ch"

# Caddy neu starten
ssh openweb "systemctl restart caddy"
```

### Spaltenreihenfolge in content_filters kaputt
```bash
# Schema prüfen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT column_name, ordinal_position FROM information_schema.columns WHERE table_name = '\''content_filters'\'' ORDER BY ordinal_position;'"

# Falls Reihenfolge falsch: recreate-content-filters-correct-order.sql ausführen
scp recreate-content-filters-correct-order.sql openweb:/tmp/
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/recreate-content-filters-correct-order.sql"
ssh openweb "docker restart llm-proxy-backend"
```

---

## 🔐 Secrets / Sensitive Daten

### Admin API Key anzeigen
```bash
ssh openweb "docker exec llm-proxy-backend env | grep ADMIN_API_KEY"
```

### Datenbank Passwort anzeigen
```bash
ssh openweb "docker exec llm-proxy-backend env | grep DATABASE_PASSWORD"
```

### Claude API Key prüfen
```bash
ssh openweb "docker exec llm-proxy-backend env | grep ANTHROPIC_API_KEY"
```

---

## 📊 Monitoring & Statistiken

### Disk Space prüfen
```bash
ssh openweb "df -h"
```

### Docker Disk Usage
```bash
ssh openweb "docker system df"
```

### Container Ressourcen
```bash
ssh openweb "docker stats --no-stream"
```

### Aktive Verbindungen
```bash
ssh openweb "netstat -an | grep -E ':(8080|3005|5432|6379|443)' | grep ESTABLISHED | wc -l"
```

### Letzte Request Logs (aus Backend)
```bash
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep 'HTTP request processed' | tail -20"
```

---

## 🧹 Cleanup

### Alte/gestoppte Container entfernen
```bash
ssh openweb "docker ps -a | grep Exit | awk '{print \$1}' | xargs docker rm"
```

### Unbenutzte Images löschen
```bash
ssh openweb "docker image prune -a"
```

### Komplettes Docker Cleanup
```bash
ssh openweb "docker system prune -a --volumes"
# VORSICHT: Löscht ALLES unbenutzte inkl. Volumes!
```

---

## 🔄 Vollständiger Neustart aller Services

```bash
# Alle Container stoppen
ssh openweb "docker stop llm-proxy-backend llm-proxy-admin-ui llm-proxy-postgres llm-proxy-redis"

# Warten 5 Sekunden
sleep 5

# In korrekter Reihenfolge starten
ssh openweb "docker start llm-proxy-postgres"
sleep 5
ssh openweb "docker start llm-proxy-redis"
sleep 3
ssh openweb "docker start llm-proxy-backend"
sleep 5
ssh openweb "docker start llm-proxy-admin-ui"

# Status prüfen
ssh openweb "docker ps | grep llm-proxy"
```

---

## 📦 Komplett-Neuinstallation (Nuclear Option)

```bash
# VORSICHT! Nur wenn wirklich alles neu aufgesetzt werden soll!

# 1. Backup erstellen!
ssh openweb "docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > /tmp/backup_before_reinstall.sql"
scp openweb:/tmp/backup_before_reinstall.sql ~/backups/

# 2. Alle Container entfernen
ssh openweb "docker stop llm-proxy-backend llm-proxy-admin-ui llm-proxy-postgres llm-proxy-redis"
ssh openweb "docker rm llm-proxy-backend llm-proxy-admin-ui llm-proxy-postgres llm-proxy-redis"

# 3. Network entfernen
ssh openweb "docker network rm llm-proxy-network"

# 4. Volumes entfernen (LÖSCHT ALLE DATEN!)
ssh openweb "docker volume rm llm-proxy_postgres-data llm-proxy_redis-data"

# 5. Neu aufsetzen
# ... docker-compose hochladen und ausführen ...
```

---

**Ende der Befehlsreferenz**
