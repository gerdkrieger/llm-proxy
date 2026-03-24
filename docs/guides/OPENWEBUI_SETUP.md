# OpenWebUI Integration mit LLM-Proxy

## 📋 Übersicht

Diese Anleitung erklärt, wie Sie **OpenWebUI** mit dem **LLM-Proxy** verbinden, um:

- ✅ Zentrale LLM-Provider-Verwaltung (Claude, OpenAI, etc.)
- ✅ Automatisches Content-Filtering (PII-Schwärzung)
- ✅ Rate Limiting und Caching
- ✅ Request Logging und Monitoring
- ✅ PDF/Bild-Attachment Filterung (OCR-basiert)

---

## 🏗️ Architektur

```
┌─────────────────┐
│   OpenWebUI     │  (Port 3000)
│  chat.aitrail.ch│
└────────┬────────┘
         │ HTTP API
         ↓
┌─────────────────┐
│   LLM-Proxy     │  (Port 8080)
│llmproxy.aitrail │
│      .ch        │
└────────┬────────┘
         │
         ├─────→ Claude (Anthropic)
         ├─────→ OpenAI (GPT-4)
         └─────→ Weitere Provider...
```

**Vorteile:**
- OpenWebUI sendet Requests an LLM-Proxy (nicht direkt an OpenAI/Claude)
- LLM-Proxy filtert sensitive Daten, cached Responses, rate-limited
- Zentrale API-Key-Verwaltung, nur LLM-Proxy kennt Provider-Keys

---

## 🔐 Schritt 1: API-Key vom LLM-Proxy erhalten

### Option A: Statischer API-Key (Empfohlen für OpenWebUI)

Der LLM-Proxy unterstützt **statische API-Keys** im Format `sk-llm-proxy-*` für einfache Client-Authentifizierung.

**Auf dem Server (falls noch nicht vorhanden):**

1. **API-Key aus Konfiguration holen:**

   ```bash
   ssh openweb
   cd /opt/llm-proxy/deployments
   grep "client_api_keys" -A 10 ../configs/config.yaml
   ```

   **Beispiel-Ausgabe:**
   ```yaml
   client_api_keys:
     - key: "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
       name: "OpenWebUI"
       scopes: ["read", "write"]
       enabled: true
   ```

2. **Neuen API-Key generieren (falls keiner existiert):**

   ```bash
   # Generiere sicheren Key
   KEY="sk-llm-proxy-openwebui-$(date +%Y-%m-%d)-$(openssl rand -hex 16)"
   echo "Neuer API Key: $KEY"
   
   # Füge zu config.yaml hinzu
   nano ../configs/config.yaml
   ```

   **In `config.yaml` hinzufügen:**
   ```yaml
   client_api_keys:
     - key: "sk-llm-proxy-openwebui-2026-02-06-[DEIN_RANDOM_STRING]"
       name: "OpenWebUI Production"
       scopes: ["read", "write"]
       enabled: true
   ```

3. **Backend neu starten:**
   ```bash
   cd /opt/llm-proxy/deployments
   docker compose -f docker-compose.openwebui.yml restart backend
   ```

4. **API-Key testen:**
   ```bash
   curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026-02-06-[DEIN_KEY]" \
     https://scrubgate.tech/v1/models
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

### Option B: OAuth Token (Für programmatischen Zugriff)

Falls Sie OAuth verwenden möchten (komplexer, aber flexibler):

1. **OAuth Client erstellen:**
   ```bash
   curl -X POST https://scrubgate.tech/oauth/clients \
     -H "Content-Type: application/json" \
     -d '{
       "name": "OpenWebUI",
       "grant_types": ["client_credentials"],
       "scopes": ["read", "write"]
     }'
   ```

2. **Access Token erhalten:**
   ```bash
   curl -X POST https://scrubgate.tech/oauth/token \
     -d "grant_type=client_credentials" \
     -d "client_id=YOUR_CLIENT_ID" \
     -d "client_secret=YOUR_CLIENT_SECRET"
   ```

---

## 🌐 Schritt 2: OpenWebUI konfigurieren

### 2.1 Admin-Einstellungen öffnen

1. **Öffnen Sie OpenWebUI:** https://chat.aitrail.ch
2. **Melden Sie sich als Admin an**
3. **Klicken Sie auf Ihr Profil-Icon → Admin Panel**
4. **Navigieren Sie zu Settings → Connections**

### 2.2 OpenAI API-Endpunkt konfigurieren

OpenWebUI verwendet die **OpenAI-kompatible API**, daher konfigurieren wir den LLM-Proxy als "OpenAI API":

**Einstellungen:**

| Feld | Wert | Erklärung |
|------|------|-----------|
| **API Base URL** | `https://scrubgate.tech/v1` | LLM-Proxy Endpunkt (nicht OpenAI!) |
| **API Key** | `sk-llm-proxy-openwebui-2026-02-06-[DEIN_KEY]` | Statischer API-Key vom LLM-Proxy |
| **Enable** | ✅ | Aktivieren |

**Screenshot-Anleitung:**

```
┌────────────────────────────────────────────┐
│ OpenAI API Configuration                   │
├────────────────────────────────────────────┤
│                                            │
│ Base URL:                                  │
│ ┌────────────────────────────────────────┐ │
│ │https://scrubgate.tech/v1         │ │
│ └────────────────────────────────────────┘ │
│                                            │
│ API Key:                                   │
│ ┌────────────────────────────────────────┐ │
│ │sk-llm-proxy-openwebui-2026-...        │ │
│ └────────────────────────────────────────┘ │
│                                            │
│ [✅] Enable OpenAI API                    │
│                                            │
│           [Save Configuration]             │
└────────────────────────────────────────────┘
```

### 2.3 Alternative: Docker Environment Variables

Falls Sie OpenWebUI als **Docker Container** betreiben:

**docker-compose.yml:**
```yaml
services:
  open-webui:
    image: ghcr.io/open-webui/open-webui:main
    container_name: open-webui
    ports:
      - "3000:8080"
    environment:
      # LLM-Proxy Integration
      - OPENAI_API_BASE_URLS=https://scrubgate.tech/v1
      - OPENAI_API_KEYS=sk-llm-proxy-openwebui-2026-02-06-[DEIN_KEY]
      
      # Optional: Weitere Einstellungen
      - WEBUI_NAME=AI Chat (via LLM-Proxy)
      - ENABLE_OPENAI_API=true
      
    volumes:
      - open-webui:/app/backend/data
    restart: unless-stopped
```

**Container neu starten:**
```bash
docker compose down
docker compose up -d
```

### 2.4 Lokale Installation (Docker Run)

```bash
docker run -d \
  --name open-webui \
  -p 3000:8080 \
  -e OPENAI_API_BASE_URLS=https://scrubgate.tech/v1 \
  -e OPENAI_API_KEYS=sk-llm-proxy-openwebui-2026-02-06-[DEIN_KEY] \
  -v open-webui:/app/backend/data \
  --restart unless-stopped \
  ghcr.io/open-webui/open-webui:main
```

---

## ✅ Schritt 3: Konfiguration testen

### 3.1 Modell-Liste abrufen

**In OpenWebUI:**
1. Öffnen Sie einen neuen Chat
2. Klicken Sie auf das **Modell-Dropdown** oben
3. Sie sollten alle LLM-Proxy-Modelle sehen:
   - `claude-3-haiku-20240307`
   - `claude-3-sonnet-20240229`
   - `gpt-4`
   - `gpt-3.5-turbo`

**Falls keine Modelle erscheinen:**

```bash
# Test API direkt
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-..." \
  https://scrubgate.tech/v1/models | jq .

# Prüfe OpenWebUI Logs
docker logs open-webui --tail 50
```

### 3.2 Test-Chat senden

1. **Wählen Sie ein Modell** (z.B. `claude-3-haiku-20240307`)
2. **Senden Sie eine Nachricht:** `"Hallo, kannst du mir helfen?"`
3. **Prüfen Sie die Antwort**

**Erwartete Ausgabe:**
```
Claude antwortet über LLM-Proxy:
"Natürlich, gerne! Wie kann ich Ihnen helfen?"
```

### 3.3 LLM-Proxy Logs prüfen

**Auf dem Server:**
```bash
ssh openweb
docker logs llm-proxy-backend --tail 50 -f
```

**Erwartete Log-Einträge:**
```json
{
  "level": "info",
  "message": "Chat completion request from client: OpenWebUI",
  "model": "claude-3-haiku-20240307"
}
{
  "level": "info",
  "message": "Routing request to provider: claude",
  "request_id": "..."
}
{
  "level": "info",
  "message": "HTTP request processed",
  "status": 200,
  "duration_ms": 1234
}
```

### 3.4 Content-Filterung testen

**Test mit sensitiven Daten:**

1. **Senden Sie eine Nachricht mit Email:**
   ```
   Meine Email ist john.doe@example.com, bitte kontaktiere mich.
   ```

2. **Prüfen Sie die Logs:**
   ```bash
   docker logs llm-proxy-backend | grep -i "filter\|redact"
   ```

3. **Erwartete Ausgabe:**
   ```json
   {
     "level": "info",
     "message": "Content filtered, 1 pattern matches found"
   }
   ```

4. **Claude erhält:**
   ```
   Meine Email ist [EMAIL_REDACTED], bitte kontaktiere mich.
   ```

---

## 🖼️ Schritt 4: PDF/Bild-Attachments testen

### 4.1 Bild mit PII hochladen

1. **Erstellen Sie ein Test-Bild mit Text:**
   ```bash
   convert -size 800x600 xc:white \
     -font DejaVu-Sans -pointsize 24 -fill black \
     -annotate +50+100 'Email: test@example.com' \
     -annotate +50+150 'Phone: +49 151 12345678' \
     test-pii.png
   ```

2. **In OpenWebUI:**
   - Öffnen Sie einen Chat mit Claude
   - Klicken Sie auf das **Attachment-Icon** 📎
   - Laden Sie `test-pii.png` hoch
   - Senden Sie: `"Was steht in diesem Bild?"`

3. **Erwartetes Ergebnis:**
   - LLM-Proxy führt **OCR** durch
   - Erkennt Email und Telefonnummer
   - Schwärzt sie mit schwarzen Balken
   - Claude erhält das geschwärzte Bild
   - Antwort: `"Email: [redacted], Phone: [redacted]"`

### 4.2 Logs prüfen

```bash
docker logs llm-proxy-backend | grep -i "ocr\|redact\|attachment"
```

**Erwartete Ausgabe:**
```json
{
  "level": "info",
  "message": "Starting redaction for file: test-pii.png (type: image)"
}
{
  "level": "info",
  "message": "OCR extracted 13 words from document"
}
{
  "level": "info",
  "message": "Found 2 PII matches, proceeding with redaction"
}
{
  "level": "info",
  "message": "Successfully redacted 2 locations in document"
}
{
  "level": "info",
  "message": "Image redacted, replacing with redacted version"
}
```

---

## 🔧 Troubleshooting

### Problem: "Unauthorized" oder 401-Fehler

**Symptom:**
```json
{"error": "invalid or expired token"}
```

**Lösung:**
1. **API-Key prüfen:**
   ```bash
   curl -H "Authorization: Bearer DEIN_KEY" \
     https://scrubgate.tech/v1/models
   ```

2. **Key Format prüfen:**
   - Muss mit `sk-llm-proxy-` beginnen
   - Kein Prefix wie `Bearer` im Key selbst

3. **Config prüfen:**
   ```bash
   ssh openweb
   cd /opt/llm-proxy/deployments
   grep -A 5 "client_api_keys" ../configs/config.yaml
   ```

4. **Backend neu starten:**
   ```bash
   docker compose -f docker-compose.openwebui.yml restart backend
   ```

### Problem: Keine Modelle in OpenWebUI

**Symptom:**
- Modell-Dropdown ist leer
- Fehler: "Failed to fetch models"

**Lösung:**
1. **Base URL prüfen:**
   - Muss `/v1` am Ende haben
   - ✅ `https://scrubgate.tech/v1`
   - ❌ `https://scrubgate.tech` (ohne /v1)

2. **Netzwerk-Verbindung testen:**
   ```bash
   curl https://scrubgate.tech/v1/models
   ```

3. **OpenWebUI Cache leeren:**
   - Logout aus OpenWebUI
   - Browser-Cache leeren
   - Neu einloggen

### Problem: Content-Filter funktioniert nicht

**Symptom:**
- Sensitive Daten werden nicht gefiltert
- Keine "redacted" Platzhalter

**Lösung:**
1. **Filter prüfen:**
   ```bash
   curl -H "X-Admin-API-Key: admin_dev_key_..." \
     https://scrubgate.tech/admin/filters
   ```

2. **Filter aktivieren:**
   - Öffnen Sie Admin UI: https://scrubgate.tech:3005
   - Navigieren zu **Content Filters**
   - Stellen Sie sicher, dass Filter **enabled** sind

3. **Logs prüfen:**
   ```bash
   docker logs llm-proxy-backend | grep -i filter
   ```

### Problem: OCR funktioniert nicht

**Symptom:**
- Bilder werden nicht analysiert
- Keine OCR-Logs

**Lösung:**
1. **Dependencies prüfen:**
   ```bash
   ssh openweb
   which tesseract && which pdftoppm && which gs
   ```

2. **Falls fehlen, installieren:**
   ```bash
   sudo apt update
   sudo apt install -y tesseract-ocr tesseract-ocr-eng tesseract-ocr-deu \
     poppler-utils ghostscript imagemagick
   ```

3. **Backend neu starten:**
   ```bash
   cd /opt/llm-proxy/deployments
   docker compose -f docker-compose.openwebui.yml restart backend
   ```

### Problem: Zu viele 429-Fehler (Rate Limiting)

**Symptom:**
```json
{"error": "rate limit exceeded"}
```

**Lösung:**
1. **Rate Limits erhöhen:**
   ```bash
   nano ../configs/config.yaml
   ```

   **Ändern:**
   ```yaml
   rate_limiting:
     enabled: true
     default_rpm: 1000  # Erhöhen auf 2000
     default_rpd: 50000 # Erhöhen auf 100000
   ```

2. **Client-spezifische Limits:**
   ```yaml
   client_api_keys:
     - key: "sk-llm-proxy-openwebui-..."
       name: "OpenWebUI"
       scopes: ["read", "write"]
       enabled: true
       rate_limit_rpm: 2000  # Custom Limit
   ```

3. **Backend neu starten**

---

## 📊 Monitoring & Logs

### LLM-Proxy Admin UI

**URL:** https://scrubgate.tech:3005

**Features:**
- ✅ Content Filter Management
- ✅ Request Logs (letzte 1000 Requests)
- ✅ Filter Match Statistics
- ✅ System Health

**Login:**
- API Key: `YOUR_ADMIN_API_KEY_HERE`

### Request Logs via API

```bash
curl -H "X-Admin-API-Key: admin_dev_key_..." \
  https://scrubgate.tech/admin/logs | jq .
```

**Beispiel-Ausgabe:**
```json
{
  "logs": [
    {
      "request_id": "abc123",
      "client_id": "OpenWebUI",
      "model": "claude-3-haiku-20240307",
      "status": 200,
      "duration_ms": 1234,
      "filter_matches": 2,
      "created_at": "2026-02-06T10:30:00Z"
    }
  ]
}
```

### Prometheus Metrics

**URL:** http://scrubgate.tech:9091/metrics

**Wichtige Metriken:**
```
# Request Count
llm_proxy_requests_total{client="OpenWebUI",model="claude-3-haiku"}

# Request Duration
llm_proxy_request_duration_seconds_bucket{client="OpenWebUI"}

# Filter Matches
llm_proxy_content_filtered_total{filter_type="email"}

# Provider Errors
llm_proxy_provider_errors_total{provider="claude"}
```

---

## 🔒 Sicherheit & Best Practices

### API-Key Sicherheit

1. **Niemals API-Keys in Git committen**
2. **Verwenden Sie starke, zufällige Keys:**
   ```bash
   openssl rand -hex 32
   ```

3. **Rotation:**
   - Alte Keys regelmäßig deaktivieren
   - Neue Keys generieren
   - In OpenWebUI aktualisieren

4. **Scopes begrenzen:**
   ```yaml
   client_api_keys:
     - key: "sk-llm-proxy-readonly-..."
       name: "ReadOnly Client"
       scopes: ["read"]  # Nur lesen, kein write
       enabled: true
   ```

### Netzwerk-Sicherheit

1. **HTTPS verwenden:**
   - ✅ `https://scrubgate.tech`
   - ❌ `http://scrubgate.tech`

2. **Firewall-Regeln:**
   ```bash
   # Nur bestimmte IPs erlauben
   ufw allow from 192.168.1.0/24 to any port 8080
   ```

3. **Rate Limiting aktivieren:**
   ```yaml
   rate_limiting:
     enabled: true
   ```

### Content-Filter Best Practices

1. **Alle PII-Filter aktivieren:**
   - Email-Adressen
   - Telefonnummern
   - Kreditkarten
   - IBANs
   - Sozialversicherungsnummern

2. **Custom Filter für branchenspezifische Daten:**
   ```bash
   curl -X POST https://scrubgate.tech/admin/filters \
     -H "X-Admin-API-Key: ..." \
     -d '{
       "pattern": "\\b\\d{3}-\\d{3}-\\d{4}\\b",
       "replacement": "[CUSTOMER_ID]",
       "filter_type": "regex",
       "description": "Customer IDs"
     }'
   ```

3. **Monitoring:**
   - Prüfen Sie regelmäßig Filter-Matches
   - Passen Sie Patterns an, wenn falsche Positives auftreten

---

## 📚 Weiterführende Dokumentation

- **[PDF & Attachment Filtering Guide](PDF_ATTACHMENT_FILTERING.md)**
- **[Content Filtering](CONTENT_FILTERING.md)**
- **[Admin API Reference](ADMIN_API.md)**
- **[Monitoring & Logging](MONITORING_AND_LOGGING.md)**

---

## 🎯 Zusammenfassung

**OpenWebUI mit LLM-Proxy in 3 Schritten:**

1. ✅ **API-Key vom LLM-Proxy holen** (`sk-llm-proxy-...`)
2. ✅ **In OpenWebUI konfigurieren:**
   - Base URL: `https://scrubgate.tech/v1`
   - API Key: `sk-llm-proxy-openwebui-...`
3. ✅ **Testen:**
   - Modelle laden
   - Chat senden
   - Content-Filterung prüfen

**Das war's!** 🎉

OpenWebUI kommuniziert jetzt über LLM-Proxy mit Claude/OpenAI, und alle Requests werden automatisch gefiltert und geloggt.

---

**Last Updated:** 2026-02-06  
**Status:** ✅ Produktionsfertig  
**OpenWebUI Version:** v0.1.x+
