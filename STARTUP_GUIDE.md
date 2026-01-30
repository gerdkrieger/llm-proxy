# 🚀 LLM-Proxy Startup Guide

Schnelle Anleitung zum Starten und Stoppen aller Services.

---

## ⚡ Quick Start (Alles auf einmal)

```bash
cd ~/Sites/golang-projekte/llm-proxy
./start-all.sh
```

**Das startet automatisch:**
- ✅ Docker Services (PostgreSQL, Redis)
- ✅ Backend Server (Go)
- ✅ Admin UI (Svelte auf Port 5173)
- ✅ Öffnet Browser

---

## 🛑 Alles Stoppen

```bash
./stop-all.sh
```

---

## 📋 Manuelle Methoden

### 1. Backend starten

```bash
# Option A: Development Mode (mit Console Output)
make dev

# Option B: Im Hintergrund
./bin/llm-proxy > /tmp/llm-proxy.log 2>&1 &

# Option C: Mit start-dev.sh (Original)
./start-dev.sh
```

### 2. Admin UI starten

```bash
cd admin-ui
npm run dev
```

Das Admin UI läuft dann auf: **http://localhost:5173**

### 3. Docker Services

```bash
# Starten
make docker-up

# Stoppen
make docker-down

# Status
docker ps
```

---

## 🌐 URLs & Ports

| Service | URL | Beschreibung |
|---------|-----|--------------|
| **Admin UI** | http://localhost:5173 | Svelte Admin Interface |
| **Backend API** | http://localhost:8080 | Go Backend Server |
| **Filter UI** | file:///.../filter-management-advanced.html | HTML Filter Management |
| **PostgreSQL** | localhost:5433 | Datenbank |
| **Redis** | localhost:6380 | Cache |
| **Prometheus** | http://localhost:9090 | Monitoring |
| **Grafana** | http://localhost:3001 | Dashboards (admin/admin) |

---

## 🔍 Status überprüfen

```bash
# Quick Check
curl http://localhost:8080/health

# Alle Services
make health-check

# Detaillierter Status
docker ps
pgrep -f llm-proxy
lsof -i :5173
lsof -i :8080
```

---

## 📝 Logs ansehen

```bash
# Backend Logs
tail -f /tmp/llm-proxy.log

# Admin UI Logs
tail -f /tmp/admin-ui.log

# Docker Logs
make docker-logs
```

---

## 🔑 Credentials

### Admin API Key
```bash
X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012
```

### Database
```bash
Host: localhost:5433
User: proxy_user
Password: dev_password_2024
Database: llm_proxy
```

### Redis
```bash
Host: localhost:6380
Password: (none)
```

---

## 🛠️ Nützliche Befehle

### Development

```bash
# Backend neu builden
make build

# Tests ausführen
make test

# Filter testen
./test-all-filters.sh

# Code formatieren
make fmt

# Linter ausführen
make lint
```

### Services Verwalten

```bash
# Backend neu starten
pkill -f bin/llm-proxy
./bin/llm-proxy > /tmp/llm-proxy.log 2>&1 &

# Admin UI neu starten
fuser -k 5173/tcp
cd admin-ui && npm run dev &

# Docker Services neu starten
make docker-restart
```

---

## 🎯 Features testen

### Content Filter testen

```bash
# Alle Filter anzeigen
curl -s http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" | jq .

# Filter Statistics
curl -s http://localhost:8080/admin/filters/stats \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" | jq .

# Comprehensive Test
./test-all-filters.sh
```

### Bulk Import testen

**Via API:**
```bash
curl -X POST http://localhost:8080/admin/filters/bulk-import \
  -H "Content-Type: application/json" \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  -d @test-bulk-import.json
```

**Via Admin UI:**
1. Öffne http://localhost:5173
2. Navigiere zum Filter Management
3. Lade CSV hoch oder füge Text ein

**Via Filter UI:**
1. Öffne `filter-management-advanced.html`
2. Tab 1: Textarea Import ODER Tab 2: CSV Upload
3. Klicke "Import"

---

## 🐛 Troubleshooting

### Backend startet nicht

```bash
# Port bereits belegt?
lsof -i :8080
pkill -f llm-proxy

# Logs prüfen
tail -100 /tmp/llm-proxy.log
```

### Admin UI startet nicht

```bash
# Port bereits belegt?
lsof -i :5173
fuser -k 5173/tcp

# Dependencies installieren
cd admin-ui
npm install

# Logs prüfen
tail -100 /tmp/admin-ui.log
```

### Docker Services laufen nicht

```bash
# Services starten
make docker-up

# Status prüfen
docker ps

# Logs ansehen
make docker-logs

# Neu starten
make docker-restart
```

### Database Connection Error

```bash
# PostgreSQL Verfügbarkeit prüfen
docker exec llm-proxy-postgres pg_isready -U proxy_user

# Migration ausführen
make migrate-up

# Connection String prüfen
cat .env | grep DB_
```

---

## 📁 Wichtige Dateien

```
llm-proxy/
├── start-all.sh              # Startet alles (Backend + Admin UI)
├── stop-all.sh               # Stoppt alles
├── start-dev.sh              # Original Dev Script
├── test-all-filters.sh       # Filter Tests
├── .env                      # Konfiguration
├── admin-ui/                 # Svelte Admin Interface
│   ├── package.json
│   └── npm run dev           # Admin UI starten
├── filter-management-advanced.html  # HTML Filter UI
├── example-filters.csv       # Beispiel Filter
└── TESTING_REPORT.md         # Test Ergebnisse
```

---

## 🚦 Startup Reihenfolge

Wenn du manuell alles starten willst:

1. **Docker Services** (zuerst!)
   ```bash
   make docker-up
   sleep 10  # Warten bis healthy
   ```

2. **Backend Server**
   ```bash
   make build  # Wenn noch nicht gebaut
   ./bin/llm-proxy > /tmp/llm-proxy.log 2>&1 &
   ```

3. **Admin UI** (nach Backend)
   ```bash
   cd admin-ui
   npm run dev > /tmp/admin-ui.log 2>&1 &
   ```

4. **Browser öffnen**
   ```bash
   firefox http://localhost:5173
   ```

---

## 📚 Weitere Dokumentation

- `README.md` - Haupt-Dokumentation
- `CONTENT_FILTERING.md` - Content Filter API Referenz
- `BULK_IMPORT_GUIDE.md` - Bulk Import Anleitung
- `TESTING_REPORT.md` - Test Ergebnisse
- `admin-ui/README.md` - Admin UI Dokumentation

---

## 💡 Tipps

1. **Immer zuerst Docker Services starten!**
   Backend braucht PostgreSQL und Redis.

2. **Admin UI zeigt "Connection Error"?**
   - Prüfe ob Backend läuft: `curl http://localhost:8080/health`
   - Prüfe `.env` in admin-ui: sollte `VITE_API_BASE_URL=http://localhost:8080` sein

3. **Performance langsam?**
   - Prüfe Cache: `curl http://localhost:8080/admin/cache/stats -H "X-Admin-API-Key: ..."`
   - Prüfe DB Connections: logs ansehen

4. **Nach Code-Änderungen:**
   ```bash
   make build  # Backend neu builden
   pkill -f bin/llm-proxy
   ./start-all.sh
   ```

---

**Viel Erfolg! 🚀**

Bei Problemen: Logs ansehen (`/tmp/llm-proxy.log` und `/tmp/admin-ui.log`)
