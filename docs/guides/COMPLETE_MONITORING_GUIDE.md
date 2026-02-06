# Vollständiges Request & Filter Monitoring

## 📋 Übersicht

Dieses Dokument zeigt Ihnen, wie Sie **vollständige Kontrolle** über alle Anfragen und Filterungen im LLM-Proxy haben.

**Was Sie überwachen können:**
- ✅ Alle Requests (erfolgreich & fehlerhaft)
- ✅ Alle gefilterten Inhalte (PII-Matches)
- ✅ Performance-Metriken (Duration, Tokens)
- ✅ Error Rates & Provider-Status
- ✅ Real-Time Activity

---

## 🎯 Schnellstart: 3 Einfache Monitoring-Scripts

### 1. Quick Check - Systemstatus & Übersicht

```bash
./scripts/maintenance/quick-check.sh
```

**Zeigt:**
- ✓ Container Status (Backend, DB, Redis)
- ✓ Letzte 1 Stunde: Requests, Erfolg, Fehler
- ✓ Filterung: Wieviele Matches, welche Pattern
- ✓ Letzte 5 Requests
- ✓ Systemresourcen (CPU, Memory, DB-Size)

**Ausgabe-Beispiel:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  LLM-PROXY QUICK CHECK
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. CONTAINER STATUS
✓ Backend       - Running (Up 2 hours)
✓ PostgreSQL    - Running (Up 2 hours)
✓ Redis         - Running (Up 2 hours)

2. LETZTE 1 STUNDE
Total Requests:      142
Erfolgreiche:        135
Fehler:              7
Ø Duration:          1234 ms

3. FILTERUNG (LETZTE 1 STUNDE)
✓ Filterung aktiv
  Filter-Matches:       45
  Gefilterte Requests:  38
  
  Top gefilterte Pattern:
    - EMAIL: 23
    - PHONE: 15
    - CREDIT_CARD: 7
```

---

### 2. Filter Report - Detaillierte Analyse

```bash
./scripts/maintenance/filter-report.sh [server] [days]

# Beispiel: Letzte 7 Tage
./scripts/maintenance/filter-report.sh openweb 7

# Beispiel: Letzte 30 Tage
./scripts/maintenance/filter-report.sh openweb 30
```

**Zeigt:**
1. **Zusammenfassung:** Total Matches, Requests, Patterns
2. **Nach Typ:** EMAIL, PHONE, CREDIT_CARD, etc.
3. **Tägliche Übersicht:** Matches pro Tag
4. **Nach Provider:** Claude vs OpenAI
5. **Letzte 20 Matches:** Detailliert mit Timestamps
6. **Statistik:** Requests mit vs ohne Filterung

**Ausgabe-Beispiel:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  LLM-PROXY FILTER REPORT
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. ZUSAMMENFASSUNG
✓ Total Filter-Matches:     245
✓ Betroffene Requests:      198
✓ Verschiedene Pattern:     5

2. FILTER-MATCHES NACH TYP
 PII-Typ      | Anzahl | Requests | Letzter Match
--------------+--------+----------+------------------
 EMAIL        | 123    | 98       | 2026-02-06 11:30
 PHONE        | 78     | 62       | 2026-02-06 11:25
 CREDIT_CARD  | 34     | 28       | 2026-02-06 10:45
 IBAN         | 8      | 7        | 2026-02-06 09:12
 SSN          | 2      | 2        | 2026-02-05 15:30

3. TÄGLICHE ÜBERSICHT
 Datum      | Matches | Requests | Patterns
------------+---------+----------+---------
 2026-02-06 | 45      | 38       | 4
 2026-02-05 | 67      | 52       | 5
 2026-02-04 | 89      | 71       | 4
 ...

6. REQUESTS MIT VS OHNE FILTERUNG
Total Requests:          542
Gefilterte Requests:     198 (36.5%)
Ungefilterte Requests:   344
```

**Das zeigt Ihnen genau:**
- Wieviel % der Requests gefiltert werden
- Welche PII-Typen am häufigsten vorkommen
- Ob die Filterung funktioniert

---

### 3. Live Monitor - Real-Time Requests

```bash
./scripts/maintenance/monitor-requests.sh [server]

# Beispiel
./scripts/maintenance/monitor-requests.sh openweb
```

**Zeigt:** Live-Stream aller Requests mit farbiger Ausgabe

**Ausgabe-Beispiel:**
```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  LLM-PROXY REQUEST MONITOR
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[11:30:15] ✅ SUCCESS
  ├─ Request: 74aaba0e10c4/718J5lA6Ok-000013
  ├─ Method: POST /v1/chat/completions
  ├─ Model: claude-3-haiku-20240307
  ├─ Status: 200
  ├─ Duration: 1896ms
  └─ HTTP request processed

[11:30:15] 🔒 FILTERED
  ├─ Request: 74aaba0e10c4/718J5lA6Ok-000013
  └─ Found 3 PII matches, proceeding with redaction

[11:30:15] 🔒 FILTERED
  ├─ Request: 74aaba0e10c4/718J5lA6Ok-000013
  └─ Successfully redacted 3 locations in document
```

**Drücken Sie Ctrl+C zum Beenden**

---

## 📊 SQL-Queries für detaillierte Analyse

### Query 1: Alle Requests der letzten 24 Stunden

```bash
ssh openweb
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy
```

```sql
SELECT 
  request_id,
  method,
  model,
  provider,
  status_code,
  duration_ms,
  total_tokens,
  ip_address,
  TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') as time
FROM request_logs
WHERE created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC
LIMIT 50;
```

### Query 2: Nur erfolgreiche Requests (Status 200)

```sql
SELECT 
  request_id,
  model,
  status_code,
  total_tokens,
  duration_ms,
  TO_CHAR(created_at, 'HH24:MI:SS') as time
FROM request_logs
WHERE status_code = 200
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC;
```

### Query 3: **ALLE GEFILTERTEN INHALTE (PII-Matches)**

```sql
SELECT 
  fm.request_id,
  fm.pattern,
  fm.replacement,
  fm.match_count,
  fm.matched_text,
  fm.model,
  fm.provider,
  TO_CHAR(fm.created_at, 'YYYY-MM-DD HH24:MI:SS') as time
FROM filter_matches fm
ORDER BY fm.created_at DESC
LIMIT 100;
```

**Das ist die wichtigste Query!** Hier sehen Sie **jeden einzelnen Fall** wo PII gefiltert wurde.

### Query 4: Zusammenfassung nach PII-Typ

```sql
SELECT 
  pattern as pii_type,
  COUNT(*) as total_matches,
  COUNT(DISTINCT request_id) as unique_requests,
  TO_CHAR(MAX(created_at), 'YYYY-MM-DD HH24:MI') as last_match
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY pattern
ORDER BY COUNT(*) DESC;
```

### Query 5: Tägliche Statistik

```sql
SELECT 
  DATE(created_at) as date,
  COUNT(*) as total_matches,
  COUNT(DISTINCT request_id) as filtered_requests,
  COUNT(DISTINCT pattern) as unique_patterns
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '30 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
```

### Query 6: Requests mit vs ohne Filterung

```sql
-- Total Requests
SELECT COUNT(*) as total_requests
FROM request_logs
WHERE created_at > NOW() - INTERVAL '24 hours';

-- Gefilterte Requests
SELECT COUNT(DISTINCT request_id) as filtered_requests
FROM filter_matches
WHERE created_at > NOW() - INTERVAL '24 hours';

-- Kombiniert (mit Prozent)
SELECT 
  (SELECT COUNT(*) FROM request_logs WHERE created_at > NOW() - INTERVAL '24 hours') as total,
  COUNT(DISTINCT fm.request_id) as filtered,
  ROUND(
    COUNT(DISTINCT fm.request_id)::numeric * 100 / 
    (SELECT COUNT(*) FROM request_logs WHERE created_at > NOW() - INTERVAL '24 hours')
  , 1) as percent_filtered
FROM filter_matches fm
WHERE fm.created_at > NOW() - INTERVAL '24 hours';
```

### Query 7: Langsame Requests (Performance-Analyse)

```sql
SELECT 
  request_id,
  model,
  duration_ms,
  total_tokens,
  TO_CHAR(created_at, 'HH24:MI:SS') as time
FROM request_logs
WHERE duration_ms > 5000  -- Länger als 5 Sekunden
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY duration_ms DESC
LIMIT 20;
```

### Query 8: Fehlerhafte Requests (Error-Analyse)

```sql
SELECT 
  request_id,
  method,
  path,
  status_code,
  error_message,
  TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI:SS') as time
FROM request_logs
WHERE status_code >= 400
  AND created_at > NOW() - INTERVAL '24 hours'
ORDER BY created_at DESC
LIMIT 50;
```

### Query 9: Top Models nach Usage

```sql
SELECT 
  model,
  COUNT(*) as request_count,
  SUM(total_tokens) as total_tokens,
  AVG(duration_ms)::integer as avg_duration_ms
FROM request_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY model
ORDER BY request_count DESC;
```

### Query 10: IP-Adressen mit den meisten Requests

```sql
SELECT 
  ip_address,
  COUNT(*) as request_count,
  COUNT(DISTINCT DATE(created_at)) as active_days,
  MAX(created_at) as last_request
FROM request_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY ip_address
ORDER BY request_count DESC
LIMIT 20;
```

---

## 🌐 Admin UI - Grafische Oberfläche

### URL öffnen

```
https://llmproxy.aitrail.ch:3005
```

### Login

```
API Key: admin_dev_key_12345678901234567890123456789012
```

### Features

1. **Request Logs**
   - Alle Requests in Tabellenform
   - Filter nach Status, Model, Client
   - Export als CSV

2. **Content Filters**
   - Alle Filter anzeigen/bearbeiten
   - Filter aktivieren/deaktivieren
   - Neue Filter erstellen

3. **Statistics**
   - Filter-Matches pro Tag
   - Top gefilterte Pattern
   - Erfolgsrate

4. **System Health**
   - Container Status
   - Database Size
   - Request Rate

---

## 🔌 Admin API - Programmatischer Zugriff

### Basis-URL

```
https://llmproxy.aitrail.ch/admin
```

### Authentication

```bash
# Header
X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012
```

### Endpoints

#### 1. Request Logs abrufen

```bash
curl -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  "https://llmproxy.aitrail.ch/admin/logs?limit=100&offset=0" | jq .
```

**Parameter:**
- `limit` - Anzahl Einträge (default: 100)
- `offset` - Offset für Pagination
- `client_id` - Filter nach Client
- `status` - Filter nach Status (200, 400, 500)
- `since` - ISO 8601 Timestamp (z.B. `2026-02-06T00:00:00Z`)

**Response:**
```json
{
  "logs": [
    {
      "request_id": "74aaba0e10c4/718J5lA6Ok-000013",
      "client_id": "openwebui-client-id",
      "method": "POST",
      "path": "/v1/chat/completions",
      "model": "claude-3-haiku-20240307",
      "provider": "claude",
      "status_code": 200,
      "duration_ms": 1896,
      "prompt_tokens": 660,
      "completion_tokens": 49,
      "total_tokens": 709,
      "cached": false,
      "ip_address": "172.31.0.1",
      "created_at": "2026-02-06T10:33:16Z"
    }
  ],
  "total": 142,
  "limit": 100,
  "offset": 0
}
```

#### 2. Filter Statistics

```bash
curl -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  "https://llmproxy.aitrail.ch/admin/filters/stats" | jq .
```

**Response:**
```json
{
  "stats": [
    {
      "filter_id": 3,
      "filter_name": "Email Filter",
      "pattern": "email",
      "match_count": 123,
      "last_match": "2026-02-06T11:30:15Z"
    },
    {
      "filter_id": 4,
      "filter_name": "Phone Filter",
      "pattern": "phone",
      "match_count": 78,
      "last_match": "2026-02-06T11:25:00Z"
    }
  ]
}
```

#### 3. Filter Matches abrufen

```bash
curl -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  "https://llmproxy.aitrail.ch/admin/filter-matches?limit=50" | jq .
```

**Response:**
```json
{
  "matches": [
    {
      "id": "uuid-here",
      "request_id": "74aaba0e10c4/718J5lA6Ok-000013",
      "filter_id": 3,
      "pattern": "EMAIL",
      "replacement": "[EMAIL]",
      "match_count": 1,
      "model": "claude-3-haiku-20240307",
      "provider": "claude",
      "created_at": "2026-02-06T10:33:15Z"
    }
  ]
}
```

#### 4. System Health

```bash
curl -H "X-Admin-API-Key: admin_dev_key_12345678901234567890123456789012" \
  "https://llmproxy.aitrail.ch/admin/health" | jq .
```

**Response:**
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "uptime": "2h 15m",
  "version": "1.0.0"
}
```

---

## 📈 Prometheus Metrics

### Metrics Endpoint

```bash
curl http://llmproxy.aitrail.ch:9091/metrics
```

### Wichtige Metriken

#### Request Metrics

```prometheus
# Total Requests
llm_proxy_requests_total{client="OpenWebUI",model="claude-3-haiku",status="200"} 142

# Request Duration (Histogram)
llm_proxy_request_duration_seconds_bucket{client="OpenWebUI",le="1.0"} 98
llm_proxy_request_duration_seconds_bucket{client="OpenWebUI",le="2.0"} 135
llm_proxy_request_duration_seconds_bucket{client="OpenWebUI",le="5.0"} 140
llm_proxy_request_duration_seconds_sum{client="OpenWebUI"} 175.234
llm_proxy_request_duration_seconds_count{client="OpenWebUI"} 142
```

#### Filter Metrics

```prometheus
# Content Filtered Total
llm_proxy_content_filtered_total{filter_type="email"} 123
llm_proxy_content_filtered_total{filter_type="phone"} 78
llm_proxy_content_filtered_total{filter_type="credit_card"} 34

# Filter Match Rate
llm_proxy_filter_match_rate{filter_id="3"} 0.365
```

#### Provider Metrics

```prometheus
# Provider Requests
llm_proxy_provider_requests_total{provider="claude",status="success"} 135
llm_proxy_provider_requests_total{provider="claude",status="error"} 7

# Provider Errors
llm_proxy_provider_errors_total{provider="claude",error_type="rate_limit"} 3
llm_proxy_provider_errors_total{provider="claude",error_type="timeout"} 2
```

### Grafana Dashboard (Optional)

Falls Sie Grafana nutzen:

**Wichtige Panels:**

1. **Request Rate**
   ```promql
   rate(llm_proxy_requests_total[5m])
   ```

2. **Filter Match Rate**
   ```promql
   rate(llm_proxy_content_filtered_total[5m])
   ```

3. **Response Time P95**
   ```promql
   histogram_quantile(0.95, rate(llm_proxy_request_duration_seconds_bucket[5m]))
   ```

4. **Error Rate**
   ```promql
   rate(llm_proxy_provider_errors_total[5m])
   ```

---

## 🚨 Real-Time Alerting (Optional)

### Webhook für Alerts einrichten

Erstellen Sie einen Webhook-Endpoint der bei kritischen Events benachrichtigt wird:

```bash
# In .env oder config.yaml
ALERT_WEBHOOK_URL=https://your-webhook.com/alerts
ALERT_ON_FILTER_MATCH=true
ALERT_ON_ERROR_RATE_THRESHOLD=0.1  # 10% Error Rate
```

### Beispiel: Slack-Alert bei hoher Filter-Rate

```bash
# Script: scripts/monitoring/check-filter-rate.sh
#!/bin/bash

FILTER_COUNT=$(docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -A -c "
SELECT COUNT(*) FROM filter_matches WHERE created_at > NOW() - INTERVAL '1 hour';
")

if [ "$FILTER_COUNT" -gt 100 ]; then
  curl -X POST https://hooks.slack.com/services/YOUR/WEBHOOK/URL \
    -H 'Content-Type: application/json' \
    -d "{
      \"text\": \"⚠️ Hohe Filter-Rate: ${FILTER_COUNT} PII-Matches in letzter Stunde!\"
    }"
fi
```

**Cron-Job einrichten:**
```bash
# /etc/crontab
*/15 * * * * /opt/llm-proxy/scripts/monitoring/check-filter-rate.sh
```

---

## 🔍 Troubleshooting & Häufige Fragen

### Frage 1: Wie sehe ich ob Filterung wirklich funktioniert?

**Antwort:** 3 Methoden:

1. **Datenbank prüfen:**
   ```sql
   SELECT COUNT(*) FROM filter_matches WHERE created_at > NOW() - INTERVAL '1 hour';
   ```
   Wenn > 0, dann funktioniert die Filterung!

2. **Filter Report:**
   ```bash
   ./scripts/maintenance/filter-report.sh
   ```

3. **Logs prüfen:**
   ```bash
   docker logs llm-proxy-backend | grep -i "filter\|redact"
   ```

### Frage 2: Wie sehe ich welche Inhalte genau gefiltert wurden?

**Antwort:** Filter-Matches Query:

```sql
SELECT 
  request_id,
  pattern,      -- z.B. "EMAIL", "PHONE"
  replacement,  -- z.B. "[EMAIL]", "[PHONE]"
  matched_text, -- Optional: Der originale Text (falls aktiviert)
  created_at
FROM filter_matches
ORDER BY created_at DESC
LIMIT 20;
```

**⚠️ WICHTIG:** `matched_text` enthält die originalen PII-Daten! Verwenden Sie diese Spalte nur für Debugging und **niemals in Logs oder Monitoring-Dashboards anzeigen!**

### Frage 3: Wie viel Prozent meiner Requests werden gefiltert?

**Antwort:**

```sql
SELECT 
  (SELECT COUNT(*) FROM request_logs WHERE created_at > NOW() - INTERVAL '24 hours') as total_requests,
  COUNT(DISTINCT fm.request_id) as filtered_requests,
  ROUND(
    COUNT(DISTINCT fm.request_id)::numeric * 100 / 
    (SELECT COUNT(*) FROM request_logs WHERE created_at > NOW() - INTERVAL '24 hours')
  , 1) || '%' as percentage
FROM filter_matches fm
WHERE fm.created_at > NOW() - INTERVAL '24 hours';
```

### Frage 4: Wie exportiere ich alle Logs?

**Antwort:**

```bash
# Als CSV exportieren
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
COPY (
  SELECT * FROM request_logs
  WHERE created_at > NOW() - INTERVAL '7 days'
  ORDER BY created_at DESC
) TO STDOUT WITH CSV HEADER;
" > request_logs_export.csv

# Als JSON exportieren (via API)
curl -H "X-Admin-API-Key: admin_dev_key_..." \
  "https://llmproxy.aitrail.ch/admin/logs?limit=10000" > logs.json
```

### Frage 5: Wie lange werden Logs gespeichert?

**Antwort:** 
- Request Logs: **Permanent** (bis manuell gelöscht)
- Filter Matches: **Permanent** (bis manuell gelöscht)

**Retention Policy einrichten (optional):**

```sql
-- Lösche Logs älter als 90 Tage
DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL '90 days';
DELETE FROM filter_matches WHERE created_at < NOW() - INTERVAL '90 days';

-- Als Cron-Job (täglich um 3 Uhr nachts)
0 3 * * * docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL '90 days';"
```

### Frage 6: Wie überwache ich die Database-Größe?

**Antwort:**

```bash
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
SELECT 
  pg_size_pretty(pg_database_size('llm_proxy')) as total_size,
  pg_size_pretty(pg_total_relation_size('request_logs')) as request_logs_size,
  pg_size_pretty(pg_total_relation_size('filter_matches')) as filter_matches_size;
"
```

### Frage 7: Performance-Optimierung bei vielen Logs?

**Antwort:**

```sql
-- Indizes prüfen
SELECT schemaname, tablename, indexname 
FROM pg_indexes 
WHERE tablename IN ('request_logs', 'filter_matches');

-- Vacuum ausführen (Performance-Optimierung)
VACUUM ANALYZE request_logs;
VACUUM ANALYZE filter_matches;

-- Auto-Vacuum einschalten (sollte Standard sein)
ALTER TABLE request_logs SET (autovacuum_enabled = true);
ALTER TABLE filter_matches SET (autovacuum_enabled = true);
```

---

## 📚 Zusammenfassung: Alle Monitoring-Methoden

| Methode | Use Case | Echtzeit | Historisch | Komplexität |
|---------|----------|----------|------------|-------------|
| **Quick Check Script** | Schneller Überblick | ❌ | ✅ (1h) | ⭐ Einfach |
| **Filter Report Script** | Detaillierte Analyse | ❌ | ✅ (7-30 Tage) | ⭐ Einfach |
| **Live Monitor Script** | Real-Time Überwachung | ✅ | ❌ | ⭐ Einfach |
| **SQL Queries** | Custom Analysen | ❌ | ✅ (Unbegrenzt) | ⭐⭐ Mittel |
| **Admin UI** | Grafische Oberfläche | ❌ | ✅ (konfigurierbar) | ⭐ Einfach |
| **Admin API** | Programmatischer Zugriff | ❌ | ✅ (via Parameter) | ⭐⭐ Mittel |
| **Prometheus + Grafana** | Professionelles Monitoring | ✅ | ✅ (konfigurierbar) | ⭐⭐⭐ Komplex |
| **Docker Logs** | Debugging | ✅ | ✅ (begrenzt) | ⭐ Einfach |

---

## 🎯 Empfohlener Workflow

### Tägliches Monitoring (5 Minuten)

```bash
# 1. Quick Check
./scripts/maintenance/quick-check.sh

# 2. Falls Auffälligkeiten: Filter Report
./scripts/maintenance/filter-report.sh openweb 1  # Nur heute
```

### Wöchentliches Review (15 Minuten)

```bash
# 1. Wöchentlicher Filter Report
./scripts/maintenance/filter-report.sh openweb 7

# 2. Performance-Check
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
SELECT 
  DATE(created_at) as date,
  COUNT(*) as requests,
  ROUND(AVG(duration_ms)) as avg_duration,
  MAX(duration_ms) as max_duration
FROM request_logs
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;
"

# 3. Error-Check
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
SELECT status_code, COUNT(*) 
FROM request_logs 
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY status_code 
ORDER BY status_code;
"
```

### Bei Problemen / Debugging (Live)

```bash
# 1. Live Monitor starten
./scripts/maintenance/monitor-requests.sh

# 2. Docker Logs in separatem Terminal
docker logs llm-proxy-backend -f | grep -i "error\|filter"

# 3. Database Live-Check
watch -n 5 'docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -t -c "SELECT COUNT(*) FROM filter_matches WHERE created_at > NOW() - INTERVAL '\''1 minute'\'';"'
```

---

## 🔐 Sicherheitshinweise

1. **Matched Text nicht loggen:**
   - Spalte `matched_text` in `filter_matches` enthält originale PII
   - Nur für Debugging verwenden
   - **Niemals in Dashboards/Reports anzeigen!**

2. **Admin API Key schützen:**
   - Niemals in Git committen
   - Nur HTTPS verwenden
   - Regelmäßig rotieren

3. **Database Access beschränken:**
   - Nur von Server aus zugänglich
   - Firewall-Regeln aktiv
   - Starkes Passwort

4. **Log Retention:**
   - Alte Logs regelmäßig löschen (DSGVO)
   - Backup-Strategie haben
   - Verschlüsseltes Backup

---

## 📞 Support & Weitere Informationen

- **Dokumentation:** `docs/guides/`
- **Admin UI:** https://llmproxy.aitrail.ch:3005
- **Prometheus:** http://llmproxy.aitrail.ch:9091/metrics

---

**Last Updated:** 2026-02-06  
**Status:** ✅ Produktionsfertig  
**Version:** 1.0.0
