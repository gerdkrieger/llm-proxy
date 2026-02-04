# 💳 Anthropic API Credits & Error Messages

## 🔴 Typische Fehlermeldungen:

### 1. "Credit balance too low"

```json
{
  "type": "error",
  "error": {
    "type": "insufficient_credits_error",
    "message": "Your credit balance is too low to access the Anthropic API. Please go to Plans & Billing to upgrade or purchase credits."
  }
}
```

**HTTP Status:** 402 Payment Required

**Was bedeutet das:**
- Dein Anthropic Credit-Guthaben ist aufgebraucht
- API-Requests werden blockiert bis Credits nachgekauft werden

---

### 2. "Rate limit exceeded"

```json
{
  "type": "error",
  "error": {
    "type": "rate_limit_error",
    "message": "Rate limit exceeded. Please retry after some time."
  }
}
```

**HTTP Status:** 429 Too Many Requests

**Was bedeutet das:**
- Zu viele Requests in kurzer Zeit
- Unterschiedliche Limits je nach Plan:
  - **Free Tier:** ~5 requests/minute
  - **Build Plan:** ~50 requests/minute  
  - **Scale Plan:** ~1000 requests/minute

---

### 3. "Invalid API key"

```json
{
  "type": "error",
  "error": {
    "type": "authentication_error",
    "message": "invalid x-api-key"
  }
}
```

**HTTP Status:** 401 Unauthorized

**Was bedeutet das:**
- API Key ist ungültig oder abgelaufen
- Oder: Account wurde deaktiviert/gesperrt

---

## 📊 Anthropic Pricing & Credits

### Build Plan (Du hast diesen!):

```
Monatliche Credits: $100 - $500 (je nach Tarif)
Gültigkeit: 30 Tage
Übertragung: Nein (verfallen am Monatsende)

Kosten pro Model:
- Claude 3.5 Sonnet: $3.00 / 1M input tokens, $15.00 / 1M output
- Claude 3.5 Haiku:  $0.80 / 1M input tokens, $4.00 / 1M output  
- Claude 3 Opus:     $15.00 / 1M input tokens, $75.00 / 1M output
- Claude 3 Haiku:    $0.25 / 1M input tokens, $1.25 / 1M output
```

### Wie viel kosten typische Requests?

```
Kurze Frage (100 Tokens In, 200 Tokens Out):
- Haiku: $0.0001  (1/100 Cent)
- Sonnet: $0.003  (3/10 Cent)
- Opus: $0.015    (1.5 Cent)

Lange Konversation (1000 Tokens In, 2000 Tokens Out):
- Haiku: $0.001   (1/10 Cent)
- Sonnet: $0.033  (3.3 Cent)
- Opus: $0.165    (16.5 Cent)

Dokument-Analyse mit Vision (5000 Tokens):
- Haiku: $0.004   (4/10 Cent)
- Sonnet: $0.015  (1.5 Cent)
```

**Mit $100 Credits kannst du etwa:**
- 100,000 Haiku-Requests (kurz)
- 30,000 Sonnet-Requests (kurz)
- 6,000 Opus-Requests (kurz)

---

## 🔍 Credits überprüfen

### Via Anthropic Console:

```
1. Gehe zu: https://console.anthropic.com
2. Login mit deinem Account
3. Navigiere zu: Settings → Billing
4. Siehe:
   - Current balance: $XX.XX
   - Monthly limit: $XXX.XX
   - Usage this month: $XX.XX
```

### Via API (Programmatisch):

```bash
# Anthropic bietet KEINEN direkten API-Endpoint für Credit-Balance
# Du musst die Console verwenden oder Usage tracking implementieren
```

### In unserem LLM-Proxy:

```bash
# Wir tracken Usage in der DB:
PGPASSWORD='dev_password_2024' psql -h localhost -p 5433 -U proxy_user -d llm_proxy -c "
  SELECT 
    model,
    COUNT(*) as requests,
    SUM(input_tokens) as total_input_tokens,
    SUM(output_tokens) as total_output_tokens,
    SUM(cost) as total_cost_usd
  FROM request_logs
  WHERE created_at > NOW() - INTERVAL '30 days'
  GROUP BY model
  ORDER BY total_cost_usd DESC;
"
```

---

## 🚨 Wie LLM-Proxy mit Credit-Fehlern umgeht

### Aktuelles Verhalten:

```go
// In chat_handler.go:
claudeResp, err := h.providerManager.CreateMessage(ctx, claudeReq)
if err != nil {
    // Check if it's a Claude API error
    if apiErr, ok := err.(*claude.APIError); ok {
        statusCode = h.mapClaudeStatusCode(apiErr.StatusCode)
        errorType = apiErr.Type
    }
    
    // Error wird an Client weitergeleitet
    h.respondError(w, statusCode, errorType, "Provider error: "+err.Error())
    return
}
```

**Was passiert:**
1. ✅ Error wird geloggt
2. ✅ HTTP Status wird korrekt weitergeleitet (402, 429, etc.)
3. ✅ Request-Log in DB erstellt (mit Error)
4. ⚠️ **ABER:** Kein automatischer Fallback zu anderem Provider

---

## 💡 Empfohlene Verbesserungen

### 1. Auto-Fallback bei Credit-Fehler

```go
// Pseudo-Code für Verbesserung:
func (h *ChatHandler) HandleChatCompletion(...) {
    // Versuche primären Provider
    resp, err := h.providerManager.CreateMessage(ctx, req)
    
    if err != nil && isCreditError(err) {
        h.logger.Warn("Primary provider out of credits, trying fallback...")
        
        // Fallback zu günstigerem Model oder anderem Provider
        resp, err = h.providerManager.CreateMessageWithFallback(ctx, req)
    }
    
    // ...
}
```

### 2. Credit-Warning in Admin UI

```javascript
// Dashboard zeigt Warning bei hoher Usage:
if (monthlyUsage > 0.8 * monthlyLimit) {
    showAlert("⚠️ 80% of monthly credits used!");
}
```

### 3. Rate Limiting

```go
// In middleware: Limit Requests pro Client
if requestsThisMinute > 50 {
    return http.StatusTooManyRequests
}
```

---

## 📈 Best Practices

### 1. Model-Auswahl optimieren:

```
Aufgabe → Empfohlenes Model:
- Einfache Fragen → Haiku (günstig, schnell)
- Komplexe Reasoning → Sonnet (beste Balance)
- Kritische Tasks → Opus (teuer, aber beste Qualität)
```

**Im LLM-Proxy:**
- Admin UI → Manage Models
- Deaktiviere teure Models (Opus) für Standard-User
- Aktiviere nur Haiku + Sonnet

### 2. Caching nutzen:

```
Im LLM-Proxy bereits implementiert:
- Identical Requests werden gecached
- Spart API-Calls und Credits
```

Check Cache-Hit-Rate:
```bash
curl -H "Authorization: Bearer admin_..." \
  http://localhost:8080/admin/cache/stats
```

### 3. Request-Limits setzen:

```yaml
# In OpenWebUI Settings:
Max Tokens: 2000 (statt 4000)
→ Halbiert Output-Kosten
```

### 4. Usage Monitoring:

```sql
-- Tägliche Costs anzeigen:
SELECT 
  DATE(created_at) as date,
  SUM(cost) as daily_cost_usd,
  COUNT(*) as requests
FROM request_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

---

## 🔧 Quick Fixes

### "Credits aufgebraucht - Was jetzt?"

**Option A: Credits nachkaufen**
```
1. https://console.anthropic.com
2. Settings → Billing
3. Purchase Credits → $10, $50, $100
4. Credits sind sofort verfügbar
```

**Option B: Model downgraden**
```
1. Admin UI → Providers → Manage Models
2. Deaktiviere: Claude 3 Opus, Claude 3.5 Sonnet
3. Aktiviere nur: Claude 3 Haiku
4. OpenWebUI verwendet automatisch günstigeres Model
```

**Option C: Anderen Provider nutzen**
```
1. Aktiviere OpenAI Provider (falls API Key vorhanden)
2. Admin UI → Providers → OpenAI → Enable
3. Models werden automatisch verfügbar
```

---

## 📞 Support

**Anthropic Status Page:**
https://status.anthropic.com

**Billing Support:**
- Email: support@anthropic.com
- Console: Help → Contact Support

**LLM-Proxy Logs:**
```bash
# Error-Logs anzeigen:
tail -f /tmp/llm-proxy-sync.log | grep -i "error\|credit\|rate"

# Request-Logs mit Errors:
PGPASSWORD='dev_password_2024' psql -h localhost -p 5433 -U proxy_user -d llm_proxy -c "
  SELECT request_id, model, error_message, created_at
  FROM request_logs
  WHERE error_message IS NOT NULL
  ORDER BY created_at DESC
  LIMIT 20;
"
```

---

## ✅ Zusammenfassung

**Dein Account Status:**
- ✅ Build Plan aktiv
- ✅ Credits vorhanden ($XX.XX)
- ✅ API Key funktioniert

**Empfohlene Actions:**
1. Models im Admin UI optimieren (Haiku für Standard)
2. Usage Monitoring aktivieren
3. Cache-Hit-Rate überwachen
4. Bei 80% Credits → Warning einrichten

**Bei Credit-Fehlern:**
1. Prüfe Balance in Anthropic Console
2. Kaufe Credits nach ODER
3. Downgrade zu Haiku ODER
4. Aktiviere Fallback-Provider

Viel Erfolg! 🚀
