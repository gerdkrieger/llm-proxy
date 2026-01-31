# OpenWebUI Integration Guide

**Datum**: 30. Januar 2026  
**Status**: ✅ Ready to Connect

---

## 🎯 Übersicht

Dieser Guide zeigt dir, wie du OpenWebUI mit dem LLM-Proxy verbindest.

### Was wurde vorbereitet:

✅ **OAuth Client erstellt:**
- Client ID: `openwebui-production`
- Client Secret: `openwebui-secret-2026-secure-key-abc123xyz789`
- Scopes: `read write`

✅ **Access Token generiert:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiJvcGVud2VidWktcHJvZHVjdGlvbiIsInNjb3BlIjoicmVhZCB3cml0ZSIsImlzcyI6ImxsbS1wcm94eSIsInN1YiI6Im9wZW53ZWJ1aS1wcm9kdWN0aW9uIiwiZXhwIjoxNzY5NzgzNzQ0LCJuYmYiOjE3Njk3ODAxNDQsImlhdCI6MTc2OTc4MDE0NCwianRpIjoiODBmYzExYjYtODIwNy00MDYyLWFkNGYtMjgwYzQ1ODEyODMzIn0.u0QsujkdgdfH7ibp3JhBJWP7KuHE0Wipzm6CxfYWeGM
```

✅ **Tests erfolgreich:**
- Models Endpoint: ✅ 3 Claude Models verfügbar
- Chat Endpoint: ✅ Claude antwortet korrekt

---

## 📋 OpenWebUI Konfigurieren

### **Methode 1: Via Web UI** (Empfohlen)

#### Schritt 1: OpenWebUI öffnen
```bash
# Im Browser öffnen:
http://localhost:3010
```

#### Schritt 2: Admin Settings öffnen
1. Oben rechts auf dein **Benutzer-Icon** klicken
2. **"Admin Panel"** oder **"Settings"** wählen
3. Navigation zu **"Connections"** oder **"External Connections"**

#### Schritt 3: OpenAI API Konfiguration
Suche nach "OpenAI API" oder "Custom API" Section:

**Eingaben:**
```
API Base URL:  http://host.docker.internal:8080/v1
API Key:       Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiJvcGVud2VidWktcHJvZHVjdGlvbiIsInNjb3BlIjoicmVhZCB3cml0ZSIsImlzcyI6ImxsbS1wcm94eSIsInN1YiI6Im9wZW53ZWJ1aS1wcm9kdWN0aW9uIiwiZXhwIjoxNzY5NzgzNzQ0LCJuYmYiOjE3Njk3ODAxNDQsImlhdCI6MTc2OTc4MDE0NCwianRpIjoiODBmYzExYjYtODIwNy00MDYyLWFkNGYtMjgwYzQ1ODEyODMzIn0.u0QsujkdgdfH7ibp3JhBJWP7KuHE0Wipzm6CxfYWeGM
```

**⚠️ Wichtig:**
- Verwende `host.docker.internal` statt `localhost` (Docker Networking!)
- API Key muss mit "Bearer " beginnen (mit Leerzeichen!)
- `/v1` am Ende der URL nicht vergessen

#### Schritt 4: Verbindung testen
1. Klicke "Test Connection" oder "Save"
2. Es sollte erfolgreich sein
3. Models sollten nun verfügbar sein

#### Schritt 5: Modell auswählen
1. Gehe zurück zum Chat
2. Klicke auf Model-Dropdown (meist oben)
3. Wähle eines der Claude Models:
   - `claude-3-opus-20240229` (Most capable)
   - `claude-3-sonnet-20240229` (Balanced) ⭐ EMPFOHLEN
   - `claude-3-haiku-20240307` (Fast)

#### Schritt 6: Testen!
1. Öffne neuen Chat
2. Schreibe: "Hallo, funktioniert die Verbindung zum LLM-Proxy?"
3. **Erwartung**: Claude sollte antworten!

---

### **Methode 2: Via Docker Umgebungsvariablen**

Falls du OpenWebUI neu starten möchtest:

```bash
# OpenWebUI Container stoppen
docker stop openwebui

# Neu starten mit Umgebungsvariablen
docker run -d \
  --name openwebui \
  -p 3010:8080 \
  -e OPENAI_API_BASE_URL=http://host.docker.internal:8080/v1 \
  -e OPENAI_API_KEY="Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiJvcGVud2VidWktcHJvZHVjdGlvbiIsInNjb3BlIjoicmVhZCB3cml0ZSIsImlzcyI6ImxsbS1wcm94eSIsInN1YiI6Im9wZW53ZWJ1aS1wcm9kdWN0aW9uIiwiZXhwIjoxNzY5NzgzNzQ0LCJuYmYiOjE3Njk3ODAxNDQsImlhdCI6MTc2OTc4MDE0NCwianRpIjoiODBmYzExYjYtODIwNy00MDYyLWFkNGYtMjgwYzQ1ODEyODMzIn0.u0QsujkdgdfH7ibp3JhBJWP7KuHE0Wipzm6CxfYWeGM" \
  -v open-webui:/app/backend/data \
  ghcr.io/open-webui/open-webui:main
```

---

## 🔧 Troubleshooting

### Problem 1: "Connection Refused" oder "Cannot connect"

**Ursache**: Docker kann `localhost` nicht erreichen

**Lösung 1** - host.docker.internal verwenden:
```
API Base URL: http://host.docker.internal:8080/v1
```

**Lösung 2** - Host-Netzwerk prüfen:
```bash
# Teste von OpenWebUI Container aus
docker exec openwebui curl http://host.docker.internal:8080/health
```

**Lösung 3** - Proxy IP direkt verwenden:
```bash
# Finde deine lokale IP
ip addr show | grep "inet 192"

# Verwende die IP:
API Base URL: http://192.168.x.x:8080/v1
```

---

### Problem 2: "Unauthorized" oder 401 Error

**Ursache**: Token fehlt oder ist falsch

**Lösung**:
- Prüfe dass API Key mit "Bearer " beginnt (mit Leerzeichen!)
- Korrekter Format:
  ```
  Bearer eyJhbGciOi...
  ```
- NICHT:
  ```
  eyJhbGciOi...  (fehlt "Bearer ")
  Bearer eyJhbGciOi...  (zu viele Spaces)
  ```

---

### Problem 3: Token expired (nach 1 Stunde)

**Ursache**: Access Token läuft nach 1 Stunde ab

**Lösung - Neuen Token generieren:**
```bash
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "openwebui-production",
    "client_secret": "openwebui-secret-2026-secure-key-abc123xyz789",
    "scope": "read write"
  }'

# Neuen Token in OpenWebUI Settings eintragen
```

**Alternative - Refresh Token verwenden:**
```bash
# Refresh Token (gültig für 30 Tage):
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiJvcGVud2VidWktcHJvZHVjdGlvbiIsInNjb3BlIjoicmVmcmVzaCIsImlzcyI6ImxsbS1wcm94eSIsInN1YiI6Im9wZW53ZWJ1aS1wcm9kdWN0aW9uIiwiZXhwIjoxNzcyMzcyMTQ0LCJuYmYiOjE3Njk3ODAxNDQsImlhdCI6MTc2OTc4MDE0NCwianRpIjoiODM1ZjVkYzEtNDkwYi00ODllLTg2ZDYtNjc2NDM2OTg0MDJhIn0.iEayoAz8FCGfXoNqn8G8-7nmOuq8yigWWMz5nWFBQjY

# Neuen Access Token holen:
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "refresh_token",
    "refresh_token": "REFRESH_TOKEN_HIER"
  }'
```

---

### Problem 4: "No models found"

**Ursache**: Models werden nicht geladen

**Lösung**:
```bash
# 1. Teste Models Endpoint direkt:
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer YOUR_TOKEN"

# 2. Sollte 3 Claude Models zeigen:
# - claude-3-opus-20240229
# - claude-3-sonnet-20240229
# - claude-3-haiku-20240307

# 3. Falls leer: Claude API Key prüfen in config.yaml
```

---

### Problem 5: "Empty response" oder keine Antwort

**Ursache**: Content Filters könnten zu aggressiv sein

**Lösung**:
```bash
# 1. Prüfe Filter im Admin UI:
http://localhost:5173 → Filters

# 2. Deaktiviere temporär alle Filter zum Testen

# 3. Teste Chat direkt:
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [{"role": "user", "content": "Test"}],
    "max_tokens": 100
  }'
```

---

## 📊 Verification Checklist

Nach der Konfiguration, prüfe:

- [ ] OpenWebUI öffnet sich: `http://localhost:3010`
- [ ] Settings zeigen LLM-Proxy als Verbindung
- [ ] Models sind sichtbar (3 Claude Models)
- [ ] Chat öffnet sich
- [ ] Model kann ausgewählt werden
- [ ] Nachricht kann gesendet werden
- [ ] Claude antwortet
- [ ] Content Filtering funktioniert (teste mit "password" → sollte zu `[***PASSWORD***]` werden)

---

## 🎉 Success Indicators

**Du weißt dass es funktioniert wenn:**

1. ✅ OpenWebUI zeigt Claude Models in der Auswahl
2. ✅ Du kannst ein Claude Model auswählen
3. ✅ Chat startet ohne Fehler
4. ✅ Claude antwortet auf deine Nachrichten
5. ✅ Im LLM-Proxy Terminal siehst du Request Logs
6. ✅ Im Admin UI (Stats) siehst du Request Counts

---

## 📝 Credentials Referenz

**Für Copy & Paste:**

```bash
# OAuth Client
Client ID:     openwebui-production
Client Secret: openwebui-secret-2026-secure-key-abc123xyz789

# API Configuration
API Base URL:  http://host.docker.internal:8080/v1

# Access Token (expires in 1h)
Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiJvcGVud2VidWktcHJvZHVjdGlvbiIsInNjb3BlIjoicmVhZCB3cml0ZSIsImlzcyI6ImxsbS1wcm94eSIsInN1YiI6Im9wZW53ZWJ1aS1wcm9kdWN0aW9uIiwiZXhwIjoxNzY5NzgzNzQ0LCJuYmYiOjE3Njk3ODAxNDQsImlhdCI6MTc2OTc4MDE0NCwianRpIjoiODBmYzExYjYtODIwNy00MDYyLWFkNGYtMjgwYzQ1ODEyODMzIn0.u0QsujkdgdfH7ibp3JhBJWP7KuHE0Wipzm6CxfYWeGM

# Refresh Token (expires in 30d)
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjbGllbnRfaWQiOiJvcGVud2VidWktcHJvZHVjdGlvbiIsInNjb3BlIjoicmVmcmVzaCIsImlzcyI6ImxsbS1wcm94eSIsInN1YiI6Im9wZW53ZWJ1aS1wcm9kdWN0aW9uIiwiZXhwIjoxNzcyMzcyMTQ0LCJuYmYiOjE3Njk3ODAxNDQsImlhdCI6MTc2OTc4MDE0NCwianRpIjoiODM1ZjVkYzEtNDkwYi00ODllLTg2ZDYtNjc2NDM2OTg0MDJhIn0.iEayoAz8FCGfXoNqn8G8-7nmOuq8yigWWMz5nWFBQjY

# Available Models
- claude-3-opus-20240229
- claude-3-sonnet-20240229 ⭐ (Empfohlen)
- claude-3-haiku-20240307
```

---

## 🔄 Token Renewal Script

Für automatische Token-Erneuerung (optional):

```bash
#!/bin/bash
# renew-token.sh

CLIENT_ID="openwebui-production"
CLIENT_SECRET="openwebui-secret-2026-secure-key-abc123xyz789"
PROXY_URL="http://localhost:8080"

# Get new token
TOKEN_RESPONSE=$(curl -s -X POST $PROXY_URL/oauth/token \
  -H "Content-Type: application/json" \
  -d "{
    \"grant_type\": \"client_credentials\",
    \"client_id\": \"$CLIENT_ID\",
    \"client_secret\": \"$CLIENT_SECRET\",
    \"scope\": \"read write\"
  }")

# Extract access token
ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')

echo "New Access Token:"
echo "Bearer $ACCESS_TOKEN"
echo ""
echo "Copy this to OpenWebUI Settings!"
```

**Usage:**
```bash
chmod +x renew-token.sh
./renew-token.sh
```

---

## 📚 Weitere Ressourcen

- **LLM-Proxy Admin UI**: http://localhost:5173
- **OpenWebUI**: http://localhost:3010
- **API Health Check**: http://localhost:8080/health
- **API Models**: http://localhost:8080/v1/models

---

## 🆘 Support

Bei Problemen:

1. **Logs prüfen:**
   ```bash
   # LLM-Proxy Logs
   tail -f logs/llm-proxy.log
   
   # OpenWebUI Logs
   docker logs openwebui -f
   ```

2. **Health Check:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Token testen:**
   ```bash
   curl http://localhost:8080/v1/models \
     -H "Authorization: Bearer YOUR_TOKEN"
   ```

4. **Admin UI öffnen:**
   - Providers: http://localhost:5173 (🤖 Providers)
   - Clients: http://localhost:5173 (👥 Clients)
   - Stats: http://localhost:5173 (📈 Statistics)

---

**Status**: ✅ Ready to Connect  
**Letzte Aktualisierung**: 30. Januar 2026  
**Version**: 1.0
