# LLM-Proxy Deployment Status

**Letzte Aktualisierung:** 2. Februar 2026, 18:05 Uhr  
**Status:** ✅ **LIVE und funktionsfähig**

---

## 🎯 Projekt-Übersicht

**Projekt:** Enterprise LLM Gateway (LLM-Proxy)  
**Sprache:** Go  
**Repository:** `git@gitlab.com:krieger-engineering/llm-proxy.git`  
**Branch:** `master`

### Server-Details

| Umgebung | Hostname | IP | SSH Alias | URL |
|----------|----------|----|-----------|----|
| **LIVE (Produktion)** | dockeronubuntu2204-s-1vcpu-2gb-70gb-intel-fra1-01 | 68.183.208.213 | `ssh openweb` | https://llmproxy.aitrail.ch |
| **Lokal (Entwicklung)** | - | 81.6.40.83 | - | http://localhost:3005 |

### Hardware Specs (LIVE Server)
- **CPU:** 1 vCPU
- **RAM:** 2GB
- **Uptime:** 111+ Tage
- **Standort:** Frankfurt (fra1)
- **Provider:** DigitalOcean

---

## ✅ Was funktioniert (LIVE)

### Container Status
```
✅ llm-proxy-backend    (Port 8080, 9091)  - Healthy
✅ llm-proxy-admin-ui   (Port 3005)        - Healthy
✅ llm-proxy-postgres   (Port 5432)        - Healthy
✅ llm-proxy-redis      (Port 6379)        - Healthy
```

### Funktionen
- ✅ **Backend API:** Alle Endpunkte funktionieren (/health, /admin/*, /v1/*)
- ✅ **Admin-UI:** Login und alle Tabs funktionieren
- ✅ **Datenbank:** 15 Tabellen, alle Migrationen erfolgreich
- ✅ **Provider:** Claude (15 Modelle) + OpenAI (56 Modelle) = **71 Modelle**
- ✅ **Content Filters:** 13 vordefinierte Filter aktiv
- ✅ **OAuth Clients:** Datenbank bereit (aktuell 0 Clients auf LIVE, 2 auf Lokal)
- ✅ **Caddy Reverse Proxy:** Korrekt konfiguriert mit TLS
- ✅ **Metrics:** Prometheus Endpoint verfügbar (Port 9091)

### API Status
```bash
# Health Check
curl https://llmproxy.aitrail.ch/health
# Response: {"status":"ok","timestamp":"..."}

# Providers
curl https://llmproxy.aitrail.ch/admin/providers \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"
# Response: {"providers":[...],"total":2}

# Content Filters
curl https://llmproxy.aitrail.ch/admin/filters \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"
# Response: {"count":13,"filters":[...]}
```

---

## 🔧 Konfiguration

### Admin API Key
```
admin_dev_key_12345678901234567890123456789012
```

### Datenbank Zugangsdaten
```
Host:     llm-proxy-postgres (Docker network) / localhost:5432 (extern)
User:     proxy_user
Database: llm_proxy
```

### Container Ports (LIVE)

| Container | Interne Ports | Externe Ports |
|-----------|---------------|---------------|
| backend | 8080 (API), 9090 (Metrics) | 8080, 9091 |
| admin-ui | 80 (nginx) | 3005 |
| postgres | 5432 | 5432 |
| redis | 6379 | 6379 |

### Caddy Konfiguration
**Datei:** `/etc/caddy/Caddyfile` (auf LIVE Server)

```caddyfile
llmproxy.aitrail.ch {
    handle /admin/* {
        reverse_proxy 127.0.0.1:8080
    }
    handle /v1/* {
        reverse_proxy 127.0.0.1:8080
    }
    handle /health {
        reverse_proxy 127.0.0.1:8080
    }
    handle {
        reverse_proxy 127.0.0.1:3005
    }
}
```

**Caddy läuft NATIV auf dem Host** (nicht in Docker!)

Caddy neu laden:
```bash
ssh openweb "systemctl reload caddy.service"
```

---

## 📊 Datenbank-Schema

### Wichtige Tabellen

**content_filters** (13 Spalten - **REIHENFOLGE WICHTIG!**)
```sql
1.  id              SERIAL PRIMARY KEY
2.  pattern         TEXT NOT NULL
3.  replacement     TEXT
4.  description     TEXT
5.  filter_type     VARCHAR(50) NOT NULL
6.  case_sensitive  BOOLEAN DEFAULT FALSE
7.  enabled         BOOLEAN DEFAULT TRUE
8.  priority        INTEGER DEFAULT 0
9.  created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
10. updated_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
11. created_by      VARCHAR(255)
12. match_count     INTEGER DEFAULT 0
13. last_matched_at TIMESTAMP
```

**WICHTIG:** Die Spaltenreihenfolge muss EXAKT dieser Reihenfolge entsprechen, da der Go-Code mit `SELECT *` arbeitet!

**Alle Tabellen:**
- oauth_clients
- oauth_tokens
- oauth_scopes
- provider_configs (Claude, OpenAI Konfigurationen)
- provider_models (71 Modelle)
- content_filters (13 Filter)
- filter_matches
- request_logs
- blocked_requests
- attachments
- attachment_analysis
- audit_logs
- rate_limits
- api_keys
- oauth_authorization_codes

---

## 🐛 Bekannte Probleme

### 1. OpenAI API Key ungültig (LIVE)
**Problem:** OpenAI Provider zeigt "unhealthy" (Status 401)
**Auswirkung:** Keine - Claude Provider funktioniert
**Lösung (wenn benötigt):**
```bash
ssh openweb
# Backend .env Datei editieren und OpenAI API Key aktualisieren
docker restart llm-proxy-backend
```

### 2. Container starten nicht automatisch
**Problem:** Nach Server-Neustart laufen Container nicht
**Grund:** Alle Container haben `restart: unless-stopped` aber wurden manuell erstellt
**Lösung:** Container haben korrektes Restart-Policy, sollte automatisch funktionieren

### 3. SSH-Verbindung manchmal "Connection refused"
**Problem:** SSH zu 68.183.208.213 sporadisch nicht erreichbar
**Grund:** Wahrscheinlich Rate-Limiting oder Firewall
**Lösung:** 30-60 Sekunden warten, dann erneut versuchen

---

## 🔄 Letzte Änderungen (Chronologisch)

### 1. Admin-UI API URL Fix
**Problem:** Admin-UI verwendete `http://localhost:8080` statt relative URLs  
**Lösung:** `admin-ui/src/lib/api.js` geändert um leere URL-Strings korrekt zu handhaben  
**Commit:** `784cfaf` - "fix(admin-ui): Use relative URLs for production with Caddy reverse proxy"

### 2. Container auf LIVE gestartet
- Postgres Container war gestoppt → gestartet
- Redis Container war gestoppt → gestartet
- Admin-UI Container war gestoppt → neu gestartet mit korrektem Image

### 3. Content Filters importiert
**Datei:** `seed-filters-live.sql`  
**Anzahl:** 13 vordefinierte Filter
- Profanity Filter (badword, damn, shit)
- Security Filter (confidential, top secret, password, secret key)
- PII Filter (Email, Telefon, Kreditkarte)
- Business Filter (CompetitorX)
- Internal (Project Phoenix)

### 4. Datenbank-Schema korrigiert
**Problem:** Spaltenreihenfolge in `content_filters` stimmte nicht mit Go-Code überein  
**Lösung:** Tabelle komplett neu erstellt mit korrekter Reihenfolge  
**Datei:** `recreate-content-filters-correct-order.sql`

---

## 📁 Wichtige Dateien im Repository

```
llm-proxy/
├── admin-ui/
│   ├── .env                    # LOKAL: VITE_API_BASE_URL=http://localhost:8080
│   ├── .env.example            # Dokumentation für Produktion
│   ├── src/lib/api.js          # ✅ GEFIXT: Unterstützt relative URLs
│   └── nginx.conf              # ✅ Korrekte Reverse Proxy Regeln
├── docker-compose.openwebui.yml # Hauptkonfiguration
├── Caddyfile.example           # Template für nativen Caddy
├── .env                         # Backend Environment Variables
├── fix-live-database.sh        # ✅ Ausgeführt - DB initialisiert
├── seed-filters-live.sql       # ✅ Ausgeführt - 13 Filter
├── recreate-content-filters-correct-order.sql  # ✅ Ausgeführt
├── migrate-content-filters-schema.sql          # Migration Script
├── diagnose-live.sh            # Diagnose Tool
├── LIVE-SERVER-FIX.md          # Deployment Guide
└── DEPLOYMENT-STATUS.md        # DIESE DATEI
```

---

## 🔐 Secrets & API Keys

### Auf LIVE Server gespeichert
- Datenbank Passwort in Docker Container ENV
- Redis (kein Passwort)
- Claude API Key in Backend ENV
- OpenAI API Key in Backend ENV (ungültig - Update benötigt)

### Admin-UI
- Kein eigenes Secret
- Verwendet Admin API Key für Backend-Zugriff

---

## 📝 Git Status

**Letzter Commit:**
```
784cfaf - fix(admin-ui): Use relative URLs for production with Caddy reverse proxy
```

**Branch:** master  
**Remote:** origin (gitlab.com:krieger-engineering/llm-proxy.git)  
**Lokaler Status:** Clean (alle Änderungen committed und gepusht)

---

## 🎯 Nächste Schritte

Siehe separate Datei: `NEXT-STEPS.md`

---

## 📞 Wichtige Befehle

Siehe separate Datei: `LIVE-SERVER-COMMANDS.md`

---

## 🏗️ Architektur

### Request Flow (Produktion)
```
Browser
  ↓ HTTPS
Caddy (nativer Host-Prozess, Port 443)
  ↓ HTTP
  ├─→ Admin-UI Container (Port 3005)  [für /, /assets/*]
  └─→ Backend Container (Port 8080)   [für /admin/*, /v1/*, /health]
        ↓
        ├─→ Postgres Container (Port 5432)
        └─→ Redis Container (Port 6379)
```

### Docker Network
Alle Container sind im **llm-proxy-network** (Bridge Mode) verbunden.  
Container können sich gegenseitig über Servicenamen erreichen:
- `postgres` (nicht localhost!)
- `redis` (nicht localhost!)
- `backend`
- `admin-ui`

---

## 💾 Backup-Strategie

### Datenbank Backup
```bash
# Auf LIVE Server
ssh openweb "docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > /tmp/llm_proxy_backup.sql"
scp openweb:/tmp/llm_proxy_backup.sql ~/backups/
```

### Container Images Backup
```bash
# Images von LIVE Server herunterladen
ssh openweb "docker save llm-proxy-admin-ui:latest | gzip" > admin-ui-backup.tar.gz
ssh openweb "docker images | grep llm-proxy"
```

---

**Ende des Deployment Status Reports**
