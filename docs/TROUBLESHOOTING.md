# LLM-Proxy - Troubleshooting Guide

Häufige Probleme und deren Lösungen.

---

## 🔍 Diagnose-Tools

### Schnell-Diagnose
```bash
# Alle Container Status
ssh openweb "docker ps --filter name=llm-proxy"

# Backend Logs (letzte 50 Zeilen)
ssh openweb "docker logs llm-proxy-backend --tail 50"

# Health Check
curl https://llmproxy.aitrail.ch/health
```

### Vollständige Diagnose
```bash
# Diagnose-Skript ausführen
./diagnose-live.sh

# Oder manuell auf LIVE Server:
ssh openweb << 'EOF'
echo "=== Container Status ==="
docker ps --filter name=llm-proxy

echo -e "\n=== Backend Logs ==="
docker logs llm-proxy-backend --tail 20 2>&1 | grep -E "(error|ERROR|fatal|FATAL)"

echo -e "\n=== Database Connection ==="
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "SELECT COUNT(*) FROM content_filters;"

echo -e "\n=== Disk Space ==="
df -h | grep -E "(Filesystem|/dev/)"
EOF
```

---

## 🐛 Häufige Probleme

### Problem 1: Admin-UI zeigt "Failed to fetch"

**Symptome:**
- Browser zeigt "Failed to fetch" in Admin-UI
- Console Error: `Unexpected token '<', "<!doctype "... is not valid JSON`

**Ursachen:**
1. Backend ist nicht erreichbar
2. Admin-UI verwendet falsche API URL
3. Caddy leitet Requests falsch weiter

**Diagnose:**
```bash
# 1. Backend erreichbar?
curl https://llmproxy.aitrail.ch/health
curl https://llmproxy.aitrail.ch/admin/filters \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"

# 2. Admin-UI API URL prüfen
ssh openweb "docker exec llm-proxy-admin-ui cat /usr/share/nginx/html/assets/index-*.js | grep -o 'const Qn=\"[^\"]*\"'"
# Sollte sein: const Qn=""  (LEER für relative URLs)

# 3. Caddy Config prüfen
ssh openweb "cat /etc/caddy/Caddyfile | grep -A 10 llmproxy"
```

**Lösung:**
```bash
# Falls API URL falsch (zeigt localhost):
cd /home/krieger/Sites/golang-projekte/llm-proxy/admin-ui
echo "VITE_API_BASE_URL=" > .env
npm run build
cd ..
docker compose -f docker-compose.openwebui.yml build admin-ui
docker save llm-proxy-admin-ui:latest | ssh openweb "docker load"
ssh openweb "docker restart llm-proxy-admin-ui"
```

---

### Problem 2: Backend startet nicht / crasht

**Symptome:**
- Container Status zeigt "Restarting" oder "Exited"
- `docker ps` zeigt Backend nicht als "Up"

**Diagnose:**
```bash
# Logs anschauen
ssh openweb "docker logs llm-proxy-backend --tail 100"

# Häufige Fehler:
# - "Failed to connect to Redis"
# - "Failed to connect to PostgreSQL"
# - "Failed to load initial filters"
```

**Ursache 1: Redis nicht erreichbar**
```bash
# Redis läuft?
ssh openweb "docker ps | grep redis"

# Wenn nicht:
ssh openweb "docker start llm-proxy-redis"

# Backend neu starten
ssh openweb "docker restart llm-proxy-backend"
```

**Ursache 2: PostgreSQL nicht erreichbar**
```bash
# Postgres läuft?
ssh openweb "docker ps | grep postgres"

# Wenn nicht:
ssh openweb "docker start llm-proxy-postgres"
sleep 5
ssh openweb "docker restart llm-proxy-backend"
```

**Ursache 3: Schema-Mismatch in content_filters**
```bash
# Spaltenreihenfolge prüfen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT column_name, ordinal_position FROM information_schema.columns WHERE table_name = '\''content_filters'\'' ORDER BY ordinal_position;'"

# Falls falsch: Tabelle neu erstellen
scp recreate-content-filters-correct-order.sql openweb:/tmp/
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/recreate-content-filters-correct-order.sql"
ssh openweb "docker restart llm-proxy-backend"
```

**Erwartete Spaltenreihenfolge:**
1. id
2. pattern
3. replacement
4. description
5. filter_type
6. case_sensitive
7. enabled
8. priority
9. created_at
10. updated_at
11. created_by
12. match_count
13. last_matched_at

---

### Problem 3: Container laufen nicht nach Server-Neustart

**Symptome:**
- Nach `ssh openweb "reboot"` sind Container gestoppt
- `docker ps` zeigt keine llm-proxy Container

**Ursache:**
Container wurden manuell erstellt und haben zwar `restart: unless-stopped`, aber Docker Service war beim Erstellen evtl. nicht korrekt konfiguriert.

**Lösung:**
```bash
# Alle Container starten (in korrekter Reihenfolge)
ssh openweb "docker start llm-proxy-postgres"
sleep 5
ssh openweb "docker start llm-proxy-redis"
sleep 3
ssh openweb "docker start llm-proxy-backend"
sleep 5
ssh openweb "docker start llm-proxy-admin-ui"

# Verifizieren
ssh openweb "docker ps | grep llm-proxy"
```

**Permanente Lösung:**
Systemd Service erstellen der Container beim Boot startet:
```bash
ssh openweb "cat > /etc/systemd/system/llm-proxy.service" << 'EOF'
[Unit]
Description=LLM-Proxy Containers
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/bin/docker start llm-proxy-postgres llm-proxy-redis
ExecStart=/bin/sleep 5
ExecStart=/usr/bin/docker start llm-proxy-backend
ExecStart=/bin/sleep 3
ExecStart=/usr/bin/docker start llm-proxy-admin-ui

[Install]
WantedBy=multi-user.target
EOF

ssh openweb "systemctl daemon-reload && systemctl enable llm-proxy.service"
```

---

### Problem 4: Datenbank "relation does not exist"

**Symptome:**
- Backend Logs: `ERROR: relation "oauth_clients" does not exist`
- API Requests schlagen fehl mit 500 Internal Server Error

**Ursache:**
Datenbank-Tabellen wurden nicht initialisiert.

**Lösung:**
```bash
# Datenbank initialisieren
scp fix-live-database.sh openweb:/tmp/
ssh openweb "chmod +x /tmp/fix-live-database.sh && /tmp/fix-live-database.sh"

# Filter importieren
scp seed-filters-live.sql openweb:/tmp/
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/seed-filters-live.sql"

# Backend neu starten
ssh openweb "docker restart llm-proxy-backend"
```

---

### Problem 5: SSL/TLS Fehler - "Connection not secure"

**Symptome:**
- Browser zeigt Zertifikatsfehler
- `curl` gibt SSL-Fehler

**Diagnose:**
```bash
# Zertifikat prüfen
openssl s_client -connect llmproxy.aitrail.ch:443 -servername llmproxy.aitrail.ch </dev/null 2>/dev/null | openssl x509 -noout -dates

# Caddy Status
ssh openweb "systemctl status caddy"

# Caddy Logs
ssh openweb "journalctl -u caddy --since '10 minutes ago'"
```

**Lösung:**
```bash
# Caddy neu starten (erneuert Zertifikat automatisch)
ssh openweb "systemctl restart caddy"

# Falls DNS-Problem:
# - DNS-Eintrag für llmproxy.aitrail.ch auf 68.183.208.213 prüfen
# - Warten bis DNS propagiert ist (bis zu 48h)
```

---

### Problem 6: Content Filters funktionieren nicht

**Symptome:**
- Filter werden nicht angewendet
- Text wird nicht gefiltert obwohl Filter existiert

**Diagnose:**
```bash
# Filter in DB vorhanden?
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT COUNT(*) FROM content_filters WHERE enabled = true;'"

# Backend hat Filter geladen?
ssh openweb "docker logs llm-proxy-backend 2>&1 | grep -i 'filter'"
```

**Häufige Ursachen:**

**1. Filter ist disabled:**
```sql
-- Filter aktivieren
UPDATE content_filters SET enabled = true WHERE id = 1;
```

**2. Priorität ist zu niedrig:**
```sql
-- Höhere Priorität setzen (100 = höchste)
UPDATE content_filters SET priority = 100 WHERE id = 1;
```

**3. Regex-Pattern fehlerhaft:**
```sql
-- Pattern testen in DB
SELECT 'test@example.com' ~ '\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b';
-- Sollte 't' (true) zurückgeben
```

**4. Backend muss neu gestartet werden:**
```bash
# Backend lädt Filter beim Start
ssh openweb "docker restart llm-proxy-backend"
```

---

### Problem 7: SSH Connection Refused

**Symptome:**
- `ssh openweb` gibt "Connection refused"
- Sporadisch nicht erreichbar

**Ursache:**
- SSH Rate Limiting
- Firewall Regeln
- Server überlastet

**Lösung:**
```bash
# 1-2 Minuten warten
sleep 60

# Erneut versuchen
ssh openweb

# Falls immer noch nicht: Direkt via IP
ssh root@68.183.208.213

# Server-Status über DigitalOcean Console prüfen
```

---

### Problem 8: Disk Space voll

**Symptome:**
- Container können nicht starten
- `docker logs` gibt "no space left on device"

**Diagnose:**
```bash
ssh openweb "df -h"
ssh openweb "docker system df"
```

**Lösung:**
```bash
# Docker Cleanup
ssh openweb "docker system prune -a --volumes"
# VORSICHT: Löscht ALLE unbenutzten Images/Volumes!

# Alte Logs löschen
ssh openweb "journalctl --vacuum-time=7d"

# Temp-Files löschen
ssh openweb "rm -rf /tmp/*"
```

---

### Problem 9: "number of field descriptions must equal number of destinations"

**Symptome:**
- Backend Logs: `number of field descriptions must equal number of destinations, got X and Y`
- Filter API gibt 500 Error

**Ursache:**
Spaltenanzahl in Datenbank stimmt nicht mit Go-Struct überein.

**Diagnose:**
```bash
# Spaltenanzahl prüfen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT COUNT(*) FROM information_schema.columns WHERE table_name = '\''content_filters'\'';'"
# MUSS 13 sein!

# Spaltenreihenfolge prüfen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT column_name, ordinal_position FROM information_schema.columns WHERE table_name = '\''content_filters'\'' ORDER BY ordinal_position;'"
```

**Lösung:**
```bash
# Tabelle mit korrekter Struktur neu erstellen
scp recreate-content-filters-correct-order.sql openweb:/tmp/
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/recreate-content-filters-correct-order.sql"
ssh openweb "docker restart llm-proxy-backend"
```

---

### Problem 10: OpenAI Provider unhealthy

**Symptome:**
- Provider-Status zeigt "unhealthy"
- Backend Logs: `"code": "invalid_api_key"`

**Ursache:**
OpenAI API Key ist ungültig oder abgelaufen.

**Lösung:**
```bash
# 1. Gültigen API Key von platform.openai.com holen

# 2. Auf LIVE Server aktualisieren
ssh openweb
# ENV-Datei editieren oder Container mit neuem Key neu erstellen

# 3. Backend neu starten
docker restart llm-proxy-backend

# 4. Verifizieren
curl https://llmproxy.aitrail.ch/admin/providers \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  | jq '.providers[] | select(.id=="openai") | .status'
# Sollte "healthy" sein
```

---

## 🔧 Nützliche Debug-Befehle

### Container Logs live verfolgen
```bash
ssh openweb "docker logs llm-proxy-backend -f"
```

### In Container einsteigen
```bash
ssh openweb "docker exec -it llm-proxy-backend /bin/sh"
```

### Netzwerk-Verbindungen prüfen
```bash
# Backend kann Postgres erreichen?
ssh openweb "docker exec llm-proxy-backend ping -c 3 postgres"

# Backend kann Redis erreichen?
ssh openweb "docker exec llm-proxy-backend ping -c 3 redis"
```

### HTTP-Requests debuggen
```bash
# Mit Headern
curl -v https://llmproxy.aitrail.ch/health

# Mit Admin Key
curl -v https://llmproxy.aitrail.ch/admin/filters \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"

# Response Time messen
time curl -s https://llmproxy.aitrail.ch/health
```

### Datenbank-Queries debuggen
```bash
# Query mit Execution Plan
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'EXPLAIN ANALYZE SELECT * FROM content_filters ORDER BY priority DESC;'"

# Locks prüfen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT * FROM pg_locks;'"
```

---

## 📞 Wenn nichts hilft

### Complete System Restart
```bash
# 1. Alle Container stoppen
ssh openweb "docker stop llm-proxy-backend llm-proxy-admin-ui llm-proxy-postgres llm-proxy-redis"

# 2. Caddy neu starten
ssh openweb "systemctl restart caddy"

# 3. Container in korrekter Reihenfolge starten
ssh openweb "docker start llm-proxy-postgres"
sleep 10
ssh openweb "docker start llm-proxy-redis"
sleep 5
ssh openweb "docker start llm-proxy-backend"
sleep 5
ssh openweb "docker start llm-proxy-admin-ui"

# 4. Warten und testen
sleep 15
curl https://llmproxy.aitrail.ch/health
```

### Backup wiederherstellen
```bash
# Falls Datenbank kaputt: Von Backup wiederherstellen
scp ~/backups/llm_proxy_backup.sql openweb:/tmp/
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/llm_proxy_backup.sql"
ssh openweb "docker restart llm-proxy-backend"
```

### Logs an Support senden
```bash
# Alle relevanten Logs sammeln
ssh openweb "docker logs llm-proxy-backend --tail 200 > /tmp/backend.log"
ssh openweb "docker logs llm-proxy-admin-ui --tail 200 > /tmp/admin-ui.log"
ssh openweb "journalctl -u caddy --since '1 hour ago' > /tmp/caddy.log"
scp openweb:/tmp/*.log ~/llm-proxy-logs/
```

---

**Ende des Troubleshooting Guides**
