# OpenWebUI Verbindung Richtig Konfigurieren

**Datum:** 2026-02-07  
**Problem:** Falscher Header (`X-Admin-API-Key` statt `Authorization: Bearer`)  
**Lösung:** Richtige Konfiguration in OpenWebUI

---

## Zugriff auf OpenWebUI Admin Panel

1. **Öffnen Sie:** `https://chat.aitrail.ch`
2. **Login:** Mit Ihrem Admin Account
3. **Navigation:** Klicken Sie auf Ihr Profil-Icon (oben rechts)
4. **Wählen Sie:** "Admin Panel" oder "Settings"

---

## Variante 1: Connections/Model Providers (Häufigste UI)

### Schritt 1: Settings finden

Nach dem Login:
```
[Profil Icon] → Admin Panel → Settings → Connections
```

Oder direkt:
```
[Profil Icon] → Settings → Connections
```

### Schritt 2: OpenAI Provider konfigurieren

Sie sollten eine Sektion sehen wie:

```
┌─────────────────────────────────────────┐
│ OpenAI                                  │
├─────────────────────────────────────────┤
│ API Base URL                            │
│ [https://llmproxy.aitrail.ch/v1     ]  │
│                                         │
│ API Key                                 │
│ [*********************************** ]  │
│                                         │
│ Custom Headers (Optional)               │
│ [                                    ]  │
└─────────────────────────────────────────┘
```

### Schritt 3: Korrekte Konfiguration

**Füllen Sie aus:**

1. **API Base URL:**
   ```
   https://llmproxy.aitrail.ch/v1
   ```

2. **API Key:**
   ```
   sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
   ```

3. **Custom Headers:**
   ```
   (LEER LASSEN!)
   ```

   **Wichtig:** Wenn Sie aktuell so etwas haben:
   ```json
   {
     "X-Admin-API-Key": "sk-llm-proxy-..."
   }
   ```
   
   **LÖSCHEN SIE DAS!** Lassen Sie das Feld leer.

### Schritt 4: Speichern und Testen

1. Klicken Sie **"Save"** oder **"Update"**
2. Klicken Sie **"Test Connection"** (falls verfügbar)
3. Sollte zeigen: ✅ Connection successful

---

## Variante 2: Model Configuration (Alternative UI)

Manche OpenWebUI Versionen haben ein anderes Interface:

### Navigation:
```
Admin Panel → Models → Add Model
```

### Konfiguration:

```
┌─────────────────────────────────────────┐
│ Add OpenAI Model                        │
├─────────────────────────────────────────┤
│ Model Name: GPT-4                       │
│ Model ID: gpt-4                         │
│                                         │
│ Provider: OpenAI Compatible             │
│                                         │
│ Base URL:                               │
│ [https://llmproxy.aitrail.ch/v1     ]  │
│                                         │
│ API Key:                                │
│ [sk-llm-proxy-openwebui-2026...     ]  │
│                                         │
│ Headers (JSON):                         │
│ [                                    ]  │
│   ← LEER LASSEN!                        │
└─────────────────────────────────────────┘
```

---

## Variante 3: Environment Variables (Fortgeschritten)

Falls die Web-UI keine Verbindungen zulässt, können Sie es via Environment Variables konfigurieren:

### Docker Compose Methode:

```bash
ssh openweb
cd /opt/open-webui  # Oder wo immer OpenWebUI installiert ist

# Backup erstellen
cp docker-compose.yml docker-compose.yml.backup

# Editieren
nano docker-compose.yml
```

**Fügen Sie hinzu:**

```yaml
services:
  open-webui:
    image: ghcr.io/open-webui/open-webui:main
    environment:
      # Ihre bestehenden Variablen...
      
      # LLM-Proxy Konfiguration
      - OPENAI_API_BASE_URLS=https://llmproxy.aitrail.ch/v1
      - OPENAI_API_KEYS=sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

**Wichtig:** Kein `X-Admin-API-Key` als Environment Variable!

**Neustart:**
```bash
docker compose down
docker compose up -d
```

---

## Was Sie aktuell wahrscheinlich haben (FALSCH)

### Szenario A: Custom Header mit falschem Namen

```json
Custom Headers:
{
  "X-Admin-API-Key": "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
}
```

**Problem:** 
- `X-Admin-API-Key` wird nur vom AdminMiddleware gelesen
- AdminMiddleware ist nur auf `/admin/*` routes aktiv
- OpenWebUI braucht `/v1/*` routes
- APIKeyMiddleware auf `/v1/*` liest nur `Authorization` header

**Ergebnis:** 401 Unauthorized

---

### Szenario B: API Key Feld leer + Custom Header

```
API Key: [leer]

Custom Headers:
{
  "X-Admin-API-Key": "sk-llm-proxy-..."
}
```

**Problem:** Gleich wie Szenario A

**Ergebnis:** 401 Unauthorized

---

## Was Sie haben sollten (RICHTIG)

### Methode 1: Nur API Key Feld (Empfohlen) ⭐

```
Base URL: https://llmproxy.aitrail.ch/v1
API Key: sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
Custom Headers: [leer]
```

**Was passiert:**
- OpenWebUI sendet automatisch: `Authorization: Bearer sk-llm-proxy-...`
- APIKeyMiddleware erkennt den Header
- Validiert den Key
- Erlaubt Zugriff

**Ergebnis:** ✅ 200 OK

---

### Methode 2: Custom Header mit richtigem Namen (Alternative)

Falls Sie unbedingt Custom Headers nutzen wollen:

```
Base URL: https://llmproxy.aitrail.ch/v1
API Key: [leer]

Custom Headers:
{
  "Authorization": "Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
}
```

**Wichtig:** Das Format muss EXAKT so sein:
- Key: `"Authorization"` (nicht `X-Admin-API-Key`)
- Value: `"Bearer sk-llm-proxy-..."` (mit `Bearer ` Prefix!)

**Ergebnis:** ✅ 200 OK

---

## Nach der Konfiguration: Testen

### Test 1: Verbindung testen in OpenWebUI

1. Nach dem Speichern sollte es einen **"Test Connection"** Button geben
2. Klicken Sie darauf
3. Sollte zeigen: ✅ "Connection successful" oder ähnlich

### Test 2: Modelle auflisten

1. Gehen Sie zu einem Chat
2. Klicken Sie auf Model Selector
3. Sollten sehen:
   ```
   - gpt-4
   - gpt-3.5-turbo
   - claude-3-opus
   - claude-3-sonnet
   - ...
   ```

Falls Sie KEINE Modelle sehen → Verbindung fehlgeschlagen

### Test 3: Chat Message senden

1. Wählen Sie ein Model (z.B. gpt-3.5-turbo)
2. Senden Sie: "Hello, please respond"
3. Sollte Antwort bekommen

Falls Fehler kommt → Überprüfen Sie Logs

### Test 4: Logs überprüfen

```bash
# LLM-Proxy Backend Logs
ssh openweb "docker logs -f llm-proxy-backend 2>&1 | grep '172.18.0.2'"
```

**Sollte zeigen:**
```json
{
  "level": "info",
  "message": "API request authenticated",
  "key_name": "OpenWebUI",
  "endpoint": "/v1/models",
  "status": 200
}
```

**Nicht mehr:**
```json
{
  "level": "warn",
  "message": "Token validation failed",
  "status": 401
}
```

### Test 5: Live Monitor

1. Öffnen Sie: `https://llmproxy.aitrail.ch:3005`
2. Login: `YOUR_ADMIN_API_KEY_HERE`
3. Klicken Sie: **🔴 Live Monitor**
4. Sollte zeigen:
   ```
   ✅ Connected
   Client: OpenWebUI
   IP: 172.18.0.2
   Last Request: 2s ago
   Status: 200
   ```

---

## Troubleshooting

### Problem: "Kann Custom Headers Feld nicht finden"

**Lösung:** Nutzen Sie einfach das normale "API Key" Feld!

Das ist sogar besser, weil OpenWebUI dann automatisch den richtigen Header (`Authorization: Bearer`) setzt.

---

### Problem: "Custom Headers Feld akzeptiert kein JSON"

Manche OpenWebUI Versionen haben verschiedene Formate:

**Format A: JSON**
```json
{
  "Authorization": "Bearer sk-llm-proxy-..."
}
```

**Format B: Key-Value Pairs**
```
Header Name: Authorization
Header Value: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

**Format C: Einzel-Line**
```
Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789
```

Probieren Sie das Format, das Ihre OpenWebUI Version verwendet.

---

### Problem: "Test Connection" zeigt Erfolg, aber Chat funktioniert nicht

**Mögliche Ursachen:**

1. **Falsches Model gewählt:**
   - Überprüfen Sie, dass das Model im LLM-Proxy konfiguriert ist
   - Schauen Sie in `/opt/llm-proxy/configs/config.yaml`

2. **Scopes fehlen:**
   - Client Key braucht `["read", "write"]` scopes
   - `read` für `/v1/models`
   - `write` für `/v1/chat/completions`

3. **Filter blockiert Request:**
   - Überprüfen Sie Admin UI → Filters → Blocked Requests
   - Vielleicht enthält Ihre Nachricht PII?

---

### Problem: "401 Unauthorized" immer noch

**Debug Schritt 1: Manueller Test**

```bash
# Testen Sie direkt mit curl
curl -H "Authorization: Bearer sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789" \
  https://llmproxy.aitrail.ch/v1/models

# Sollte Liste von Modellen zeigen
```

Falls das funktioniert → OpenWebUI ist falsch konfiguriert

Falls das NICHT funktioniert → Backend Problem

**Debug Schritt 2: Überprüfen Sie die Config**

```bash
ssh openweb "grep -A 5 'sk-llm-proxy-openwebui' /opt/llm-proxy/configs/config.yaml"
```

Sollte zeigen:
```yaml
- key: "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
  name: "OpenWebUI"
  scopes: ["read", "write"]
  enabled: true  # ← Muss true sein!
```

**Debug Schritt 3: Backend neustarten**

```bash
ssh openweb "cd /opt/llm-proxy && docker compose -f docker-compose.openwebui.yml restart backend"

# Warten Sie 10 Sekunden
sleep 10

# Dann erneut testen
```

---

## Zusammenfassung

### Was Sie ändern müssen:

**ALT (aktuell - falsch):**
```
Custom Headers:
{
  "X-Admin-API-Key": "sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789"
}
```

**NEU (korrekt):**
```
API Key Field:
sk-llm-proxy-openwebui-2026-01-30-secure-key-abc123xyz789

Custom Headers:
(leer)
```

### Warum:

- `X-Admin-API-Key` ist für `/admin/*` routes
- OpenWebUI braucht `/v1/*` routes
- `/v1/*` routes erwarten `Authorization: Bearer` header
- OpenWebUI's "API Key" Feld fügt automatisch `Authorization: Bearer` hinzu

### Das Ergebnis:

- ✅ OpenWebUI kann Modelle auflisten
- ✅ OpenWebUI kann Chat Messages senden
- ✅ Live Monitor zeigt grünen Status
- ✅ Logs zeigen 200 Status Codes
- ✅ Keine 401 Fehler mehr

---

**Nächster Schritt:** Melden Sie sich bei OpenWebUI an und ändern Sie die Konfiguration wie oben beschrieben!
