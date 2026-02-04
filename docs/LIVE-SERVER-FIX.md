# 🚨 LIVE SERVER FIX - KOMPLETTE ANLEITUNG

## 📊 DIAGNOSE ZUSAMMENFASSUNG

### ✅ Was funktioniert:
- Alle Container sind HEALTHY
- Docker Netzwerk korrekt konfiguriert
- Environment Variables korrekt (DATABASE_HOST=postgres, REDIS_HOST=redis)
- DNS Resolution funktioniert
- Postgres und Redis sind erreichbar

### ❌ Was NICHT funktioniert:
- **KRITISCH:** Datenbank-Tabellen existieren nicht (`ERROR: relation "oauth_clients" does not exist`)
- `/admin/clients` gibt keine Daten zurück (weil DB leer ist)
- Caddy routet `/admin/*` Requests vermutlich nicht zum Backend

---

## 🔧 FIX IN 3 SCHRITTEN

### **SCHRITT 1: Datenbank initialisieren**

Auf dem LIVE-Server (68.183.208.213):

```bash
# Via SSH von diesem Server:
cd /home/krieger/Sites/golang-projekte/llm-proxy
ssh root@68.183.208.213 'bash -s' < fix-live-database.sh
```

**Was das Script macht:**
- Prüft aktuellen Datenbankzustand
- Erstellt alle fehlenden Tabellen (oauth_clients, content_filters, etc.)
- Erstellt Indexes
- Verifiziert Ergebnis

**Erwartete Ausgabe:**
```
CREATE TABLE
CREATE TABLE
...
 client_count 
--------------
            0
(1 row)
```

---

### **SCHRITT 2: Caddy Reverse Proxy konfigurieren**

**Option A: Caddy läuft im gleichen Docker Netzwerk**

```bash
# Auf LIVE-Server:
ssh root@68.183.208.213

# Caddy Config bearbeiten (Pfad kann variieren):
nano /etc/caddy/Caddyfile
# ODER wenn Caddy in Docker:
nano /path/to/caddy/Caddyfile
```

Fügen Sie diese Konfiguration hinzu:

```caddyfile
llmproxy.aitrail.ch {
    # Admin API → Backend
    handle /admin/* {
        reverse_proxy llm-proxy-backend:8080
    }
    
    # LLM API → Backend
    handle /v1/* {
        reverse_proxy llm-proxy-backend:8080
    }
    
    # Health Check → Backend
    handle /health {
        reverse_proxy llm-proxy-backend:8080
    }
    
    # Alles andere → Admin UI
    handle /* {
        reverse_proxy llm-proxy-admin-ui:80
    }
}
```

**WICHTIG:** Wenn Caddy NICHT im Docker Netzwerk ist:
```caddyfile
llmproxy.aitrail.ch {
    handle /admin/* {
        reverse_proxy localhost:8080  # Statt llm-proxy-backend:8080
    }
    
    handle /v1/* {
        reverse_proxy localhost:8080
    }
    
    handle /* {
        reverse_proxy localhost:3005  # Statt llm-proxy-admin-ui:80
    }
}
```

**Caddy neu laden:**
```bash
# System Caddy:
sudo systemctl reload caddy

# Docker Caddy:
docker exec caddy-container caddy reload --config /etc/caddy/Caddyfile
# ODER
docker restart caddy-container
```

---

### **SCHRITT 3: Admin-UI mit korrekter Konfiguration deployen**

Die Admin-UI muss **relative URLs** verwenden (kein absoluter API_BASE_URL).

**Option A: Via GitLab CI/CD (Empfohlen)**

```bash
# Von diesem Server:
cd /home/krieger/Sites/golang-projekte/llm-proxy

# Änderungen committen
git add admin-ui/nginx.conf
git commit -m "fix(admin-ui): update nginx config for Caddy reverse proxy"
git push origin master

# In GitLab UI:
# 1. Gehe zu: https://gitlab.com/krieger-engineering/llm-proxy/-/pipelines
# 2. Klicke auf "Run Pipeline"
# 3. Wähle Branch: master
# 4. Start Pipeline
# 5. Warte bis Build fertig ist
# 6. Manuell "deploy:localhost" Job triggern
```

**Option B: Manuell auf LIVE-Server**

```bash
ssh root@68.183.208.213

# Zum Projektverzeichnis
cd /path/to/llm-proxy  # Wo auch immer das Repo liegt

# Git pull
git pull origin master

# Admin-UI neu bauen (mit leerem API_BASE_URL für relative Pfade)
cd admin-ui
VITE_API_BASE_URL="" npm run build

# Docker Image neu bauen
cd ..
docker-compose -f docker-compose.openwebui.yml build admin-ui

# Container neu starten
docker-compose -f docker-compose.openwebui.yml up -d admin-ui

# Backend auch neu starten (damit Datenbank-Verbindung initialisiert wird)
docker restart llm-proxy-backend
```

---

## ✅ VERIFIKATION

### 1. Datenbank prüfen
```bash
ssh root@68.183.208.213
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "\dt"
```
**Erwartung:** Liste von 15 Tabellen

### 2. Backend API prüfen
```bash
# Von LIVE-Server:
curl http://localhost:8080/health
curl http://localhost:8080/admin/clients -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012"
```
**Erwartung:** 
```json
{"status":"ok","timestamp":"..."}
{"clients":[],"total":0}
```

### 3. Caddy Routing prüfen
```bash
# Von LIVE-Server:
curl -I http://localhost:3005/admin/clients

# Von außen:
curl -I https://llmproxy.aitrail.ch/admin/clients
```
**Erwartung:** Status 200 oder 401 (nicht 404!)

### 4. Admin-UI testen
```bash
# Browser öffnen:
https://llmproxy.aitrail.ch

# Login mit Admin API Key:
admin_dev_key_12345678901234567890123456789012

# Navigiere zu "Clients" Tab
# Sollte leere Liste zeigen (nicht "Failed to fetch")

# Navigiere zu "Filters" Tab  
# Sollte leere Liste zeigen (nicht "Failed to fetch")
```

---

## 🎯 TEST-CLIENTS ERSTELLEN

Nachdem alles läuft, erstellen Sie Test-Clients:

```bash
# Von LIVE-Server:
curl -X POST http://localhost:8080/admin/clients \
  -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "openwebui-production",
    "client_secret": "secret_12345",
    "name": "OpenWebUI Production Client",
    "enabled": true
  }'
```

Oder über Admin-UI:
1. Gehe zu https://llmproxy.aitrail.ch
2. Login
3. Klicke "Clients" Tab
4. Klicke "Add Client"
5. Fülle Formular aus
6. Klicke "Save"

---

## 🔍 TROUBLESHOOTING

### Problem: "Failed to fetch" im Admin-UI
**Ursache:** Caddy routet Requests nicht zum Backend  
**Lösung:** Caddy Config prüfen (siehe Schritt 2)

### Problem: "relation does not exist"
**Ursache:** Datenbank-Schema nicht erstellt  
**Lösung:** Schritt 1 wiederholen

### Problem: Admin-UI zeigt weiße Seite
**Ursache:** JavaScript Build-Fehler oder falscher API_BASE_URL  
**Lösung:** Browser Console öffnen (F12) und Fehler prüfen

### Problem: CORS Errors
**Ursache:** Backend und Frontend haben unterschiedliche Origins  
**Lösung:** Caddy muss beide auf gleicher Domain hosten (bereits konfiguriert)

---

## 📝 FINALE CHECKLISTE

- [ ] Schritt 1: Datenbank initialisiert (`fix-live-database.sh`)
- [ ] Schritt 2: Caddy konfiguriert und neu geladen
- [ ] Schritt 3: Admin-UI neu deployed (via GitLab oder manuell)
- [ ] Verifikation 1: Datenbank-Tabellen existieren
- [ ] Verifikation 2: Backend API antwortet
- [ ] Verifikation 3: Caddy routet korrekt
- [ ] Verifikation 4: Admin-UI lädt ohne Fehler
- [ ] Test-Client erstellt
- [ ] OpenWebUI kann LLM-Proxy erreichen

---

**Nach Abschluss sollte alles funktionieren!** ✅
