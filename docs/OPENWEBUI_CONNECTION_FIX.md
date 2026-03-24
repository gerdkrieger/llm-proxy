# OpenWebUI Connection Fix

## 🔍 Problem-Analyse

### Log-Analyse vom Production Server

**Datum:** 2026-02-06  
**Zeit:** 10:45 - 10:53 Uhr

#### ❌ Gefundenes Problem:

```json
{
  "level": "warn",
  "message": "Token validation failed: invalid token: failed to parse token: token is malformed: token contains an invalid number of segments",
  "method": "GET",
  "path": "/v1/models",
  "remote_addr": "172.18.0.2",     // <-- Docker network IP (OpenWebUI)
  "status": 401,                    // <-- Unauthorized
  "user_agent": "Python/3.11 aiohttp/3.13.2"
}
```

**Details:**
- OpenWebUI versucht Verbindungen herzustellen ✅
- Sendet Requests an `/v1/models` (Model-Liste abrufen) ✅
- **ABER:** Authentifizierung schlägt fehl ❌
- Status 401 (Unauthorized) bei JEDEM Request

#### Token-Problem:

```
Token: "dmin_dev_key_1234567..."  // <-- FALSCH! (truncated)
```

Der Token ist **unvollständig/falsch**:
- Sollte beginnen mit: `sk-llm-proxy-` (statischer API Key)
- Aktuell: `dmin_dev_key_...` (sieht aus wie Admin-Key, falsch!)

---

## 🔧 Lösung

### Schritt 1: Korrekten API-Key in OpenWebUI eintragen

**OpenWebUI muss einen STATIC API KEY verwenden!**

#### Option A: Bestehenden Key aus Config holen

```bash
# Auf dem Server
ssh openweb
cd /opt/llm-proxy/deployments
grep -A 10 "client_api_keys" ../configs/config.yaml
```

**Erwartete Ausgabe:**
```yaml
client_api_keys:
  - key: "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
    name: "OpenWebUI"
    scopes: ["read", "write"]
    enabled: true
```

➡️ **Verwenden Sie diesen Key:** `sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789`

#### Option B: Neuen Key generieren

```bash
# Generiere neuen sicheren Key
KEY="sk-llm-proxy-openwebui-$(date +%Y-%m-%d)-$(openssl rand -hex 16)"
echo "Neuer Key: $KEY"

# Füge zu config.yaml hinzu
nano /opt/llm-proxy/configs/config.yaml
```

**In `config.yaml` einfügen:**
```yaml
client_api_keys:
  - key: "sk-llm-proxy-openwebui-2026-02-06-[IHR_RANDOM_STRING]"
    name: "OpenWebUI Production"
    scopes: ["read", "write"]
    enabled: true
```

**Backend neu starten:**
```bash
cd /opt/llm-proxy/deployments
docker compose -f docker-compose.openwebui.yml restart backend
```

### Schritt 2: OpenWebUI konfigurieren

#### In OpenWebUI Admin Panel:

1. Öffnen: https://chat.aitrail.ch
2. Login als Admin
3. **Settings → Connections → OpenAI API**
4. Eintragen:

| Feld | Wert | Wichtig! |
|------|------|----------|
| **Base URL** | `https://scrubgate.tech/v1` | OHNE Trailing Slash! |
| **API Key** | `sk-llm-proxy-openwebui-2026-...` | ✅ Muss mit `sk-llm-proxy-` beginnen! |
| **Enable** | ✅ Aktiviert | |

5. **Save Configuration**

#### Alternative: Docker Environment Variables

Falls OpenWebUI als Container läuft:

```bash
# docker-compose.yml bearbeiten
nano docker-compose.yml
```

```yaml
services:
  open-webui:
    environment:
      # LLM-Proxy Integration
      - OPENAI_API_BASE_URLS=https://scrubgate.tech/v1
      - OPENAI_API_KEYS=sk-llm-proxy-openwebui-2026-02-06-[IHR_KEY]
```

**Container neu starten:**
```bash
docker compose restart open-webui
```

### Schritt 3: Verbindung testen

#### Test 1: Models abrufen

```bash
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026-..." \
  https://scrubgate.tech/v1/models | jq .
```

**Erwartete Ausgabe:**
```json
{
  "object": "list",
  "data": [
    {"id": "claude-3-haiku-20240307", ...},
    {"id": "gpt-4", ...}
  ]
}
```

#### Test 2: Chat-Completion senden

In OpenWebUI:
1. Neuen Chat öffnen
2. Model auswählen (sollte jetzt erscheinen!)
3. Nachricht senden: `"Hallo!"`
4. Antwort sollte kommen ✅

#### Test 3: Logs prüfen

```bash
ssh openweb
docker logs llm-proxy-backend --tail 20
```

**Erwartete Ausgabe:**
```json
{
  "level": "info",
  "message": "Detected static API key format, validating: sk-llm-proxy-openweb...",
}
{
  "level": "info",
  "message": "Static API key authenticated successfully: OpenWebUI",
}
{
  "level": "info",
  "message": "Chat completion request from client: OpenWebUI",
  "status": 200
}
```

**✅ Keine 401-Fehler mehr!**

---

## 🔴 Live Monitor - Neue Admin UI Komponente

### Was wurde erstellt:

Eine **visuelle Monitoring-Komponente** in der Admin UI die **in Echtzeit** zeigt:

#### Features:

1. **📊 Recent Activity Stats**
   - Requests letzte Minute
   - Requests letzte Stunde
   - Error Rate
   - Gefilterte Requests

2. **✅ Client Connection Status**
   - **OpenWebUI Status:**
     - ✅ Connected (grün)
     - 🔐 Authentication Failed (gelb)
     - ❌ Error (rot)
     - ⚫ Disconnected (grau)
   - Last Seen (z.B. "30s ago")
   - Total Requests
   - Failed Requests
   
   - **Other Clients Status:**
     - Gleiche Metriken für andere Clients

3. **📋 Recent Requests Table**
   - Letzte 50 Requests
   - Zeit, Status, Method, Path
   - Client-Erkennung (OpenWebUI vs Others)
   - Duration

4. **🔄 Auto-Refresh**
   - Automatisches Update alle 5 Sekunden
   - An/Aus-Schalter
   - Manueller Refresh-Button

### Zugriff:

```
URL: https://scrubgate.tech:3005
Login: YOUR_ADMIN_API_KEY_HERE

Menü: 🔴 Live Monitor (grüner Button)
```

### Screenshots/Beschreibung:

```
┌─────────────────────────────────────────────────────────────┐
│ Live Monitor                  [🔄 Auto-Refresh ON] [🔄 Refresh Now] │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐                │
│ │  15   │  │  142  │  │  5.2% │  │   38  │                │
│ │ Last  │  │ Last  │  │ Error │  │Filtered│               │
│ │Minute │  │ Hour  │  │ Rate  │  │ Requests│              │
│ └───────┘  └───────┘  └───────┘  └───────┘                │
│                                                             │
│ Client Connection Status                                    │
│ ┌─────────────────────────────────────────────────────────┐ │
│ │ ✅ OpenWebUI         Connected                          │ │
│ │    Last Seen: 30s ago                                   │ │
│ │    Total Requests: 142                                  │ │
│ │    Failed Requests: 0                                   │ │
│ └─────────────────────────────────────────────────────────┘ │
│                                                             │
│ Recent Requests (Last 50)                                   │
│ ┌───────────────────────────────────────────────────────┐  │
│ │Time    Status  Method  Path              Client       │  │
│ ├───────────────────────────────────────────────────────┤  │
│ │30s ago  200   POST   /v1/chat/completions  OpenWebUI │  │
│ │1m ago   200   GET    /v1/models           OpenWebUI  │  │
│ │2m ago   200   POST   /v1/chat/completions  OpenWebUI │  │
│ └───────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Wie es OpenWebUI erkennt:

Die Komponente identifiziert OpenWebUI durch:
- IP-Adresse: `172.18.0.2` (Docker network)
- User-Agent: `Python/3.11 aiohttp/3.13.2`

**Auth Failed Detection:**
- Status 401 → Zeigt "🔐 Authentication Failed"
- Plus Hinweis: "API Key authentication failing. Check OpenWebUI configuration."

---

## ✅ Checkliste für erfolgreiche Verbindung

### Vor der Konfiguration:

- [ ] OpenWebUI läuft und ist erreichbar
- [ ] LLM-Proxy Backend läuft (Port 8080)
- [ ] Admin UI läuft (Port 3005)

### Konfiguration:

- [ ] Statischer API-Key erstellt (`sk-llm-proxy-...`)
- [ ] Key in `config.yaml` hinzugefügt
- [ ] Backend neu gestartet
- [ ] Key in OpenWebUI eingetragen (Settings → Connections)
- [ ] Base URL korrekt: `https://scrubgate.tech/v1`

### Verifizierung:

- [ ] Models erscheinen in OpenWebUI Dropdown
- [ ] Test-Chat funktioniert
- [ ] Logs zeigen `status: 200` (keine 401 mehr)
- [ ] Live Monitor zeigt "✅ Connected"

---

## 🚨 Troubleshooting

### Problem 1: Immer noch 401-Fehler

**Symptom:**
```json
{"status": 401, "message": "invalid or expired token"}
```

**Lösung:**
1. API-Key Format prüfen:
   - ✅ MUSS beginnen mit: `sk-llm-proxy-`
   - ❌ NICHT: `admin_dev_key_...`
   - ❌ NICHT: `sk-ant-...` (Claude Key)
   - ❌ NICHT: `sk-proj-...` (OpenAI Key)

2. Key in config.yaml prüfen:
   ```bash
   grep -A 5 "client_api_keys" /opt/llm-proxy/configs/config.yaml
   ```

3. Backend-Logs prüfen:
   ```bash
   docker logs llm-proxy-backend | grep -i "api key"
   ```

### Problem 2: Models erscheinen nicht in OpenWebUI

**Symptom:**
- Model-Dropdown ist leer
- Fehler: "Failed to fetch models"

**Lösung:**
1. Base URL prüfen:
   - ✅ `https://scrubgate.tech/v1` (mit /v1!)
   - ❌ `https://scrubgate.tech` (ohne /v1)

2. Netzwerk-Verbindung testen:
   ```bash
   curl https://scrubgate.tech/v1/models
   ```

3. OpenWebUI Cache leeren:
   - Logout
   - Browser-Cache leeren
   - Neu einloggen

### Problem 3: Live Monitor zeigt "Disconnected"

**Symptom:**
- Status: ⚫ Disconnected
- Last Seen: Never

**Bedeutung:**
- Keine Requests in letzten 5 Minuten
- OpenWebUI sendet KEINE Requests

**Lösung:**
1. OpenWebUI erreichbar prüfen:
   ```bash
   docker ps | grep open-webui
   ```

2. OpenWebUI Logs prüfen:
   ```bash
   docker logs open-webui --tail 50
   ```

3. OpenWebUI Config prüfen:
   - Settings → Connections
   - API Key vorhanden?
   - Base URL korrekt?

### Problem 4: Live Monitor zeigt "🔐 Authentication Failed"

**Symptom:**
- Status: 🔐 Authentication Failed (gelb)
- Failed Requests > 0
- Hinweis: "API Key authentication failing"

**Lösung:**
- **Das ist GENAU Ihr aktuelles Problem!**
- API-Key in OpenWebUI ist falsch
- Siehe "Schritt 1: Korrekten API-Key eintragen" oben

---

## 📊 Monitoring nach dem Fix

### Mit Live Monitor:

1. **Öffnen:** https://scrubgate.tech:3005
2. **Navigieren:** 🔴 Live Monitor
3. **Prüfen:**
   - ✅ OpenWebUI Status: "Connected" (grün)
   - ✅ Failed Requests: 0
   - ✅ Recent Requests: Status 200

### Mit Logs:

```bash
# Real-Time Monitoring
ssh openweb
docker logs llm-proxy-backend -f | grep -i "openwebui\|172.18.0.2"
```

**Erfolgreiche Logs:**
```json
{"message": "Static API key authenticated successfully: OpenWebUI"}
{"message": "Chat completion request from client: OpenWebUI", "status": 200}
{"message": "Routing request to provider: claude"}
```

### Mit Scripts:

```bash
# Quick Check
./scripts/maintenance/quick-check.sh

# Live Monitor
./scripts/maintenance/monitor-requests.sh openweb
```

---

## 📚 Weitere Dokumentation

- **OpenWebUI Setup:** [docs/guides/OPENWEBUI_SETUP.md](guides/OPENWEBUI_SETUP.md)
- **Complete Monitoring:** [docs/guides/COMPLETE_MONITORING_GUIDE.md](guides/COMPLETE_MONITORING_GUIDE.md)
- **PDF Filtering:** [docs/guides/PDF_ATTACHMENT_FILTERING.md](guides/PDF_ATTACHMENT_FILTERING.md)

---

## 🎯 Zusammenfassung

**Problem:** OpenWebUI kann nicht mit LLM-Proxy kommunizieren
- ❌ Status 401 (Unauthorized)
- ❌ Falscher/unvollständiger API-Key
- ❌ Token beginnt NICHT mit `sk-llm-proxy-`

**Lösung:** Korrekten Static API Key verwenden
- ✅ Key muss Format haben: `sk-llm-proxy-...`
- ✅ In OpenWebUI Settings → Connections eintragen
- ✅ Base URL: `https://scrubgate.tech/v1`

**Verifizierung:** Live Monitor in Admin UI
- ✅ Zeigt Connection Status in Echtzeit
- ✅ Erkennt Authentication Failures
- ✅ Auto-Refresh alle 5 Sekunden

**Deployment:** Admin UI neu deployen
```bash
cd /opt/llm-proxy/deployments
docker compose -f docker-compose.openwebui.yml build admin-ui
docker compose -f docker-compose.openwebui.yml up -d admin-ui
```

---

**Last Updated:** 2026-02-06  
**Status:** ❌ OpenWebUI nicht korrekt konfiguriert  
**Next Step:** API-Key in OpenWebUI eintragen → Sofort ✅ Connected
