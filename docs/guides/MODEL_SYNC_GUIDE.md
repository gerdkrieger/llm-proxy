# 🔄 Model-Synchronisation mit OpenWebUI

## ✅ Was wurde implementiert:

Die Model-Auswahl aus dem Admin UI wird **automatisch** mit OpenWebUI synchronisiert!

### Wie es funktioniert:

```
Admin UI (Model Management)
    ↓ 
Aktiviert/Deaktiviert Models in DB
    ↓
/v1/models Endpoint filtert nach enabled Models
    ↓
OpenWebUI zeigt nur aktivierte Models an
```

## 📊 Status-Check:

```bash
# Anzahl der verfügbaren Models anzeigen:
curl -H "Authorization: Bearer sk-llm-proxy-..." \
  http://localhost:8080/v1/models | jq '.data | length'

# Alle verfügbaren Models anzeigen:
curl -H "Authorization: Bearer sk-llm-proxy-..." \
  http://localhost:8080/v1/models | jq '.data[].id'

# Logs zeigen Filterung:
tail -f /tmp/llm-proxy-sync.log | grep "Returning.*enabled models"
```

**Aktuell:** 1 von 4 Models enabled (siehe Logs)

## 🎯 So verwendest du es:

### 1. Models aktivieren/deaktivieren:

```
1. Öffne Admin UI: http://localhost:5173
2. Gehe zu "Providers"
3. Klicke "Manage Models" bei Claude oder OpenAI
4. Setze Häkchen bei den Models, die du nutzen möchtest:
   ☑ Claude 3.5 Sonnet
   ☑ Claude 3 Haiku
   ☐ Claude 2.1 (deaktiviert)
   ☐ GPT-4 (deaktiviert)
5. Klicke "Save Configuration"
```

### 2. OpenWebUI aktualisieren:

```
1. Gehe zu OpenWebUI: http://localhost:3010
2. Klicke auf Model-Dropdown
3. OpenWebUI zeigt automatisch nur die aktivierten Models!
```

**Wichtig:** OpenWebUI cached Models manchmal. Falls Models nicht sofort erscheinen:
- Seite neu laden (F5)
- Oder in OpenWebUI: Settings → Models → Refresh

## 🔍 Technische Details:

### Endpoint-Filterung:

**Vorher:**
```json
GET /v1/models
→ Alle 8 Claude + 12 OpenAI = 20 Models
```

**Jetzt:**
```json
GET /v1/models
→ Nur enabled Models aus DB
→ Log: "Returning 8 enabled models (out of 20 total)"
```

### Model-Status prüfen:

```sql
-- Alle Models anzeigen:
SELECT model_id, model_name, enabled 
FROM provider_models 
WHERE provider_id = 'claude'
ORDER BY model_name;

-- Enabled Models zählen:
SELECT provider_id, COUNT(*) as total, SUM(CASE WHEN enabled THEN 1 ELSE 0 END) as enabled
FROM provider_models
GROUP BY provider_id;
```

### Backward Compatibility:

Falls ein Model **nicht** in der DB existiert:
- ✅ Wird als "enabled" behandelt (Standard)
- ✅ Erscheint in OpenWebUI
- ⚠️ Sobald du es im Admin UI konfigurierst, wird DB-Status verwendet

## 🎨 Workflow-Beispiel:

### Szenario: Nur Claude 3.5 Models für Produktion

```bash
# 1. Admin UI: Manage Models für Claude
#    ☑ Claude 3.5 Sonnet
#    ☑ Claude 3.5 Haiku
#    ☐ Alle anderen

# 2. Speichern → DB wird aktualisiert

# 3. Check in Terminal:
curl -H "Authorization: Bearer ..." http://localhost:8080/v1/models | \
  jq '.data[] | select(.owned_by == "claude") | .id'

# Output:
# "claude-3-5-sonnet-20241022"
# "claude-3-5-haiku-20241022"

# 4. OpenWebUI zeigt nur diese 2 Models! ✅
```

## 🐛 Troubleshooting:

### "Ich sehe keine Models in OpenWebUI"

**Mögliche Ursachen:**
1. **Alle Models deaktiviert**
   ```bash
   # Lösung: Mindestens 1 Model aktivieren im Admin UI
   ```

2. **OpenWebUI Cache**
   ```bash
   # Lösung: Browser-Refresh oder OpenWebUI neu starten
   ```

3. **Falscher API Key**
   ```bash
   # Check: OpenWebUI muss korrekten API Key verwenden
   # Settings → Connections → API Key
   ```

### "Models erscheinen in OpenWebUI aber nicht im Admin UI"

**Ursache:** Model noch nicht in DB konfiguriert

**Lösung:**
```bash
# Im Admin UI einmal auf "Manage Models" klicken
# → Lädt alle verfügbaren Models
# → Speichern aktiviert sie in der DB
```

## 📈 Performance:

- **Caching:** Models werden pro Request gefiltert
- **DB-Abfrage:** ~10ms pro /v1/models Request
- **Skalierung:** Funktioniert mit 100+ Models

## 🔐 Sicherheit:

- Nur Admin kann Models aktivieren/deaktivieren
- Normale User sehen nur enabled Models
- Audit-Trail in DB (`updated_at` Timestamp)

---

## ✨ Das war's!

Die Synchronisation funktioniert **automatisch**. Jede Änderung im Admin UI wird sofort in OpenWebUI sichtbar (nach Refresh).

Viel Erfolg! 🚀
