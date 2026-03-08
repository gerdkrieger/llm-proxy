# 🚀 Quick Start - Content Filters

## Schritt 1: Server starten

```bash
cd /home/krieger/Sites/golang-projekte/llm-proxy

# Server im Hintergrund starten
./bin/llm-proxy &

# Oder im Vordergrund mit Logs:
./bin/llm-proxy
```

**Warte 5 Sekunden**, bis der Server vollständig gestartet ist.

---

## Schritt 2: Filter erstellen

### Option A: Automatisches Setup-Script

```bash
# Beispiel-Filter automatisch erstellen
./create-example-filters.sh
```

Dieses Script erstellt:
- 2x Schimpfwort-Filter (badword, damn)
- 2x Vertraulichkeits-Filter (confidential information, Project Phoenix)
- 3x PII-Filter (Email, Telefon, Kreditkarte)
- 1x Konkurrenz-Filter (CompetitorX)

### Option B: Einzelnen Filter manuell erstellen

```bash
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "badword",
    "replacement": "[GEFILTERT]",
    "description": "Filtert Schimpfwörter",
    "filter_type": "word",
    "case_sensitive": false,
    "enabled": true,
    "priority": 100
  }'
```

### Option C: Web Interface verwenden

```bash
# HTML-Datei im Browser öffnen
firefox filter-management.html
# oder
google-chrome filter-management.html
```

---

## Schritt 3: Filter anzeigen

```bash
# Alle Filter auflisten
curl -s http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" | jq '.'

# Nur die wichtigsten Infos
curl -s http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  | jq '.filters[] | {id, pattern, replacement, filter_type, enabled, priority}'
```

---

## Schritt 4: Filter testen

```bash
# Ad-hoc Test eines Filters
curl -X POST http://localhost:8080/admin/filters/test \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "This is a badword test",
    "pattern": "badword",
    "replacement": "[FILTERED]",
    "filter_type": "word"
  }' | jq '.'
```

**Erwartete Ausgabe:**
```json
{
  "original_text": "This is a badword test",
  "filtered_text": "This is a [FILTERED] test",
  "matches": [
    {
      "filter_id": 0,
      "pattern": "badword",
      "replacement": "[FILTERED]",
      "match_count": 1
    }
  ],
  "has_matches": true
}
```

---

## Schritt 5: Filter in Chat-Request testen

```bash
# 1. OAuth Token holen
TOKEN_RESPONSE=$(curl -s -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "test_client",
    "client_secret": "test_secret_123456",
    "scope": "read write"
  }')

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
echo "Access Token: $ACCESS_TOKEN"

# 2. Chat-Request mit gefiltertem Content senden
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "messages": [
      {
        "role": "user",
        "content": "This message contains a badword and confidential information"
      }
    ],
    "max_tokens": 50
  }' | jq '.'
```

**Was passiert:**
- Der User-Prompt wird **automatisch gefiltert**
- "badword" → "[GEFILTERT]"
- "confidential information" → "[VERTRAULICH_ENTFERNT]"
- Der gefilterte Text wird an Claude gesendet
- Im Server-Log siehst du: "Applied 2 content filters to request..."

---

## Schritt 6: Statistiken anzeigen

```bash
# Filter-Statistiken
curl -s http://localhost:8080/admin/filters/stats \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" | jq '.'
```

**Ausgabe:**
```json
{
  "total_filters": 8,
  "enabled_filters": 8,
  "cached_filters": 8,
  "total_matches": 2,
  "by_type": {
    "word": 2,
    "phrase": 2,
    "regex": 3
  },
  "cache_age_seconds": 45,
  "last_match": "2026-01-30T15:30:00Z"
}
```

---

## Verwaltung: Filter ändern/löschen

### Filter deaktivieren (ohne zu löschen)

```bash
# Filter mit ID 1 deaktivieren
curl -X PUT http://localhost:8080/admin/filters/1 \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{"enabled": false}'
```

### Filter aktualisieren

```bash
# Priorität ändern
curl -X PUT http://localhost:8080/admin/filters/1 \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "priority": 150,
    "description": "Aktualisierte Beschreibung"
  }'
```

### Filter löschen

```bash
# Filter mit ID 1 löschen
curl -X DELETE http://localhost:8080/admin/filters/1 \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE"
```

### Cache aktualisieren

```bash
# Nach mehreren Änderungen Cache manuell neu laden
curl -X POST http://localhost:8080/admin/filters/refresh \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE"
```

---

## Direkter Datenbankzugriff (Optional)

Falls du Filter direkt in der Datenbank sehen willst:

```bash
# Alle Filter in Datenbank anzeigen
PGPASSWORD=dev_password_2024 psql -h localhost -p 5433 -U proxy_user -d llm_proxy \
  -c "SELECT id, pattern, replacement, filter_type, enabled, priority, match_count FROM content_filters ORDER BY priority DESC;"

# Filter hinzufügen (direkt in DB - nicht empfohlen)
PGPASSWORD=dev_password_2024 psql -h localhost -p 5433 -U proxy_user -d llm_proxy \
  -c "INSERT INTO content_filters (pattern, replacement, filter_type, enabled, priority) VALUES ('test', '[FILTERED]', 'word', true, 100);"
```

⚠️ **Wichtig:** Nach direkten DB-Änderungen unbedingt Cache aktualisieren:
```bash
curl -X POST http://localhost:8080/admin/filters/refresh \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE"
```

---

## Troubleshooting

### Server läuft nicht
```bash
# Prüfen ob Server läuft
ps aux | grep llm-proxy

# Server starten
cd /home/krieger/Sites/golang-projekte/llm-proxy
./bin/llm-proxy
```

### Filter werden nicht angewendet
```bash
# 1. Cache aktualisieren
curl -X POST http://localhost:8080/admin/filters/refresh \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE"

# 2. Prüfen ob Filter enabled sind
curl -s http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  | jq '.filters[] | select(.enabled == false)'

# 3. Server-Logs prüfen
tail -f /path/to/log
```

### Admin API Key falsch
Der Standard-Dev-Key ist:
```
YOUR_ADMIN_API_KEY_HERE
```

Dieser ist in der `.env` Datei konfiguriert als `ADMIN_API_KEY`.

---

## Beispiel-Workflows

### Workflow 1: Schimpfwörter filtern

```bash
# 1. Filter erstellen
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{"pattern": "badword", "replacement": "[FILTERED]", "filter_type": "word", "enabled": true, "priority": 100}'

# 2. Testen
curl -X POST http://localhost:8080/admin/filters/test \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{"text": "This is a badword", "pattern": "badword", "replacement": "[FILTERED]", "filter_type": "word"}'

# 3. In Chat verwenden
# (siehe Schritt 5 oben)
```

### Workflow 2: Email-Adressen filtern

```bash
# 1. Regex-Filter für Emails erstellen
curl -X POST http://localhost:8080/admin/filters \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "pattern": "\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b",
    "replacement": "[EMAIL]",
    "filter_type": "regex",
    "enabled": true,
    "priority": 90
  }'

# 2. Testen
curl -X POST http://localhost:8080/admin/filters/test \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Contact me at john.doe@example.com",
    "pattern": "\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b",
    "replacement": "[EMAIL]",
    "filter_type": "regex"
  }'
```

---

## Weitere Ressourcen

- **Vollständige Dokumentation:** `CONTENT_FILTERING.md`
- **Test-Script:** `# Use Admin UI at http://localhost:3005`
- **Setup-Script:** `./create-example-filters.sh`
- **Web Interface:** `filter-management.html`

---

**Happy Filtering! 🛡️**
