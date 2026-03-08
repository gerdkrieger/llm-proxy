# 🚀 LLM-Proxy Projekt Wiederaufnahme

**Erstellt:** 2. Februar 2026  
**Projekt-Status:** ✅ **LIVE und funktionsfähig**

Diese Datei dient als Einstiegspunkt zur Wiederaufnahme des Projekts ohne den kompletten Code neu analysieren zu müssen.

---

## 📚 Dokumentations-Übersicht

Das Projekt hat folgende Dokumentationsdateien:

| Datei | Zweck | Wann zu lesen |
|-------|-------|---------------|
| **DEPLOYMENT-STATUS.md** | Aktueller Status, was funktioniert, Konfiguration | ZUERST lesen |
| **LIVE-SERVER-COMMANDS.md** | Alle wichtigen Befehle für LIVE Server | Als Referenz |
| **NEXT-STEPS.md** | Was als nächstes zu tun ist | Für Weiterentwicklung |
| **TROUBLESHOOTING.md** | Problemlösungen | Bei Problemen |
| **RESUME-PROJECT.md** | Diese Datei - Einstiegspunkt | Start |

---

## ⚡ Schnellstart

### 1. Projekt-Überblick verschaffen (5 Minuten)

```bash
# In Projekt-Verzeichnis wechseln
cd /home/krieger/Sites/golang-projekte/llm-proxy

# Deployment Status lesen
cat DEPLOYMENT-STATUS.md | less

# Git Status prüfen
git status
git log --oneline -10
```

### 2. LIVE System Status prüfen (2 Minuten)

```bash
# Container Status
ssh openweb "docker ps --filter name=llm-proxy"

# Health Check
curl https://llmproxy.aitrail.ch/health

# Admin-UI öffnen
# Browser: https://llmproxy.aitrail.ch
# Login mit: YOUR_ADMIN_API_KEY_HERE
```

### 3. Lokale Entwicklungsumgebung (5 Minuten)

```bash
# Admin-UI .env für lokal konfigurieren
echo "VITE_API_BASE_URL=http://localhost:8080" > admin-ui/.env

# Container starten
docker compose -f docker-compose.openwebui.yml up -d

# Admin-UI öffnen
# Browser: http://localhost:3005
```

---

## 🎯 Typische Aufgaben

### Neue Features entwickeln

```bash
# 1. Branch erstellen
git checkout -b feature/neue-feature

# 2. Code ändern
# ... Development ...

# 3. Lokal testen
docker compose -f docker-compose.openwebui.yml up -d
# Browser öffnen: http://localhost:3005

# 4. Commit & Push
git add .
git commit -m "feat: beschreibung der änderung"
git push origin feature/neue-feature

# 5. Merge Request auf GitLab erstellen
```

### Bug auf LIVE fixen

```bash
# 1. Problem analysieren
ssh openweb "docker logs llm-proxy-backend --tail 100"

# 2. Lokalen Code fixen
# ... Code ändern ...

# 3. Lokal testen
docker compose -f docker-compose.openwebui.yml restart backend

# 4. Image bauen
docker compose -f docker-compose.openwebui.yml build backend

# 5. Auf LIVE deployen
docker save llm-proxy-backend:latest | ssh openweb "docker load"
ssh openweb "docker restart llm-proxy-backend"

# 6. Verifizieren
curl https://llmproxy.aitrail.ch/health

# 7. Commit & Push
git add .
git commit -m "fix: beschreibung des bugs"
git push origin master
```

### Datenbank-Migration durchführen

```bash
# 1. Migration SQL-Datei erstellen
# migrations/001_add_column.sql

# 2. Lokal testen
docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < migrations/001_add_column.sql

# 3. Auf LIVE ausführen
scp migrations/001_add_column.sql openweb:/tmp/
ssh openweb "docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy < /tmp/001_add_column.sql"

# 4. Backend neu starten
ssh openweb "docker restart llm-proxy-backend"
```

### Content Filter hinzufügen

```bash
# Option 1: Via Admin-UI (empfohlen)
# 1. https://llmproxy.aitrail.ch öffnen
# 2. Login mit Admin Key
# 3. "Filters" Tab → "New Filter"
# 4. Pattern, Type, Replacement konfigurieren

# Option 2: Via SQL
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy" << 'EOF'
INSERT INTO content_filters (
    pattern, filter_type, replacement, 
    priority, description, enabled
) VALUES (
    'geheimes-projekt', 'word', '[PROJEKT]',
    100, 'Internes Projekt', true
);
EOF

# Backend neu starten um Filter zu laden
ssh openweb "docker restart llm-proxy-backend"
```

---

## 🔑 Wichtige Informationen

### Zugangsdaten

**LIVE Server:**
- **SSH:** `ssh openweb` (root@68.183.208.213)
- **Admin-UI:** https://llmproxy.aitrail.ch
- **Admin API Key:** `YOUR_ADMIN_API_KEY_HERE`

**Datenbank:**
- **Host:** llm-proxy-postgres (Docker) / localhost:5432 (extern)
- **User:** proxy_user
- **Database:** llm_proxy

### Container auf LIVE

```
llm-proxy-backend    → Port 8080 (API), 9091 (Metrics)
llm-proxy-admin-ui   → Port 3005
llm-proxy-postgres   → Port 5432
llm-proxy-redis      → Port 6379
```

### Wichtige URLs

```
https://llmproxy.aitrail.ch           → Admin-UI
https://llmproxy.aitrail.ch/health    → Backend Health
https://llmproxy.aitrail.ch/admin/*   → Admin API
https://llmproxy.aitrail.ch/v1/*      → LLM Proxy API
http://68.183.208.213:9091/metrics    → Prometheus Metrics
```

---

## 🚨 Bekannte Probleme

1. **OpenAI API Key ungültig** (LIVE)
   - Status: ⚠️ Provider zeigt "unhealthy"
   - Impact: Minimal (Claude funktioniert)
   - Fix: Siehe NEXT-STEPS.md → Punkt 1

2. **Admin API Key ist Dev-Key**
   - Status: ⚠️ Nicht für Produktion geeignet
   - Impact: Sicherheitsrisiko
   - Fix: Siehe NEXT-STEPS.md → Punkt 2

3. **SSH sporadisch "Connection refused"**
   - Status: ⚠️ Temporär
   - Impact: Minimal (1-2 Minuten warten)
   - Workaround: Siehe TROUBLESHOOTING.md → Problem 7

---

## 📝 Letzte Änderungen

**Commit:** `784cfaf` (2. Februar 2026)
```
fix(admin-ui): Use relative URLs for production with Caddy reverse proxy

- Modified api.js to handle empty VITE_API_BASE_URL
- Added .env.example with documentation
```

**Was wurde gefixt:**
- ✅ Admin-UI zeigt "Failed to fetch" → GELÖST
- ✅ Content Filters werden nicht geladen → GELÖST
- ✅ Datenbank Schema-Mismatch → GELÖST
- ✅ Container laufen nicht → GELÖST

---

## 📖 Architektur-Überblick

```
Internet
  ↓ HTTPS (443)
Caddy (Host-Prozess)
  ↓
  ├→ Admin-UI Container (Port 3005) [SPA: Svelte]
  └→ Backend Container (Port 8080)  [Go: Gin Framework]
       ↓
       ├→ Postgres Container (Port 5432) [15 Tabellen]
       └→ Redis Container (Port 6379)    [Caching]
```

**Technologie-Stack:**
- **Backend:** Go 1.21+, Gin, pgx (PostgreSQL Driver)
- **Frontend:** Svelte, Vite, Tailwind CSS
- **Datenbank:** PostgreSQL 14
- **Cache:** Redis 7
- **Reverse Proxy:** Caddy 2 (nativ)
- **Container:** Docker, Docker Compose
- **CI/CD:** GitLab CI/CD

---

## 🎓 Entwickler-Onboarding

### Neue Entwickler - Erste Schritte

1. **Repository klonen**
   ```bash
   git clone git@gitlab.com:krieger-engineering/llm-proxy.git
   cd llm-proxy
   ```

2. **Dokumentation lesen** (in dieser Reihenfolge)
   - RESUME-PROJECT.md (diese Datei)
   - DEPLOYMENT-STATUS.md
   - LIVE-SERVER-COMMANDS.md

3. **Lokale Entwicklungsumgebung aufsetzen**
   ```bash
   # Admin-UI Environment
   echo "VITE_API_BASE_URL=http://localhost:8080" > admin-ui/.env
   
   # Container starten
   docker compose -f docker-compose.openwebui.yml up -d
   
   # Logs verfolgen
   docker compose -f docker-compose.openwebui.yml logs -f backend
   ```

4. **LIVE System erkunden** (READ-ONLY)
   ```bash
   # Container Status
   ssh openweb "docker ps"
   
   # Logs anschauen
   ssh openweb "docker logs llm-proxy-backend --tail 50"
   
   # Datenbank erkunden
   ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c '\dt'"
   ```

5. **Erste Änderung machen**
   - Kleine Änderung in Admin-UI (z.B. Text ändern)
   - Lokal testen
   - Branch erstellen, committen, MR erstellen

### Code-Struktur verstehen

```
llm-proxy/
├── cmd/                      # Main Applications
│   └── server/              # Backend Hauptprogramm
├── internal/                # Private Go Code
│   ├── application/         # Business Logic
│   ├── domain/              # Domain Models
│   ├── infrastructure/      # External Services (DB, Redis, APIs)
│   └── interfaces/          # HTTP Handlers (API)
├── admin-ui/                # Admin Frontend (Svelte)
│   ├── src/                 # Svelte Components
│   │   ├── components/      # UI Components
│   │   └── lib/             # API Client, Utils
│   └── nginx.conf           # Nginx Config für Container
├── deployments/             # Deployment Configs
├── migrations/              # SQL Migrationen
└── docker-compose.*.yml     # Docker Compose Configs
```

---

## 🆘 Hilfe & Support

### Bei Problemen

1. **Logs prüfen:** Siehe LIVE-SERVER-COMMANDS.md
2. **Troubleshooting:** Siehe TROUBLESHOOTING.md
3. **GitLab Issues:** Neues Issue im Repository anlegen
4. **Dokumentation:** README.md, LIVE-SERVER-FIX.md

### Wichtige Kommandos auswendig lernen

```bash
# Container Status
ssh openweb "docker ps | grep llm-proxy"

# Backend Logs
ssh openweb "docker logs llm-proxy-backend --tail 50"

# Health Check
curl https://llmproxy.aitrail.ch/health

# Admin-UI neu starten
ssh openweb "docker restart llm-proxy-admin-ui"

# Datenbank Query
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c 'SELECT COUNT(*) FROM content_filters;'"
```

---

## ✅ Projekt erfolgreich wiederaufgenommen!

Du bist jetzt bereit weiterzuarbeiten! 🎉

**Nächste Schritte:**
1. DEPLOYMENT-STATUS.md vollständig lesen
2. NEXT-STEPS.md durchgehen
3. Erste Aufgabe aus NEXT-STEPS.md wählen
4. Los geht's!

Bei Fragen: Alle Informationen sind in den Dokumentations-Dateien. 📚

---

**Viel Erfolg! 🚀**
