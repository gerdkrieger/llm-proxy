# LLM-Proxy Monitoring & Logging Guide

## 🎯 Übersicht

Umfassendes Monitoring und Logging für LLM-Proxy mit Fokus auf:
- Request/Response Logging
- PDF/Attachment Verarbeitung
- Content Filter Aktivität
- Performance Metrics
- Error Tracking

---

## 📊 Monitoring-Architektur

```
┌────────────────────────────────────────────────────────────────┐
│                    LLM-Proxy Backend                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐        │
│  │   Requests   │  │ Attachments  │  │   Filters    │        │
│  │   Logging    │  │   Logging    │  │   Logging    │        │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘        │
│         │                  │                  │                 │
│         └──────────────────┴──────────────────┘                 │
│                            │                                     │
└────────────────────────────┼─────────────────────────────────────┘
                             │
              ┌──────────────┴──────────────┐
              │                             │
              ▼                             ▼
     ┌────────────────┐           ┌────────────────┐
     │   PostgreSQL   │           │   Prometheus   │
     │   (Structured  │           │   (Metrics)    │
     │    Logging)    │           │                │
     └────────┬───────┘           └────────┬───────┘
              │                             │
              ▼                             ▼
     ┌────────────────┐           ┌────────────────┐
     │   Admin UI     │           │    Grafana     │
     │   (Dashboard)  │           │   (Dashboard)  │
     └────────────────┘           └────────────────┘
```

---

## 🔧 Level 1: Basis-Logging (Bereits implementiert)

### Strukturiertes Logging

```go
// Logging ist bereits vorhanden in:
internal/application/attachment/service.go
internal/application/attachment/redaction_service.go

// Beispiele:
s.logger.Infof("Starting redaction for file: %s (type: %s)", filename, fileType)
s.logger.Infof("OCR extracted %d words from document", len(ocrWords))
s.logger.Infof("Found %d PII matches, proceeding with redaction", len(piiMatches))
s.logger.Infof("Successfully redacted %d locations in document", len(piiMatches))
```

### Logs anzeigen

```bash
# Development
docker compose -f docker-compose.dev.yml logs -f backend

# Production
ssh openweb "docker logs llm-proxy-backend -f"

# Filter auf Attachments
docker logs llm-proxy-backend 2>&1 | grep -i "attachment\|pdf\|ocr\|redact"
```

---

## 📈 Level 2: Database Logging (Bereits implementiert)

### Request Logs Tabelle

```sql
-- Alle Requests werden geloggt
SELECT 
    id,
    client_id,
    model,
    provider,
    input_tokens,
    output_tokens,
    total_cost,
    cached,
    created_at
FROM request_logs
ORDER BY created_at DESC
LIMIT 20;
```

### Filter Matches Tabelle

```sql
-- Alle Filter-Treffer (inkl. Attachments)
SELECT 
    id,
    filter_id,  -- 0 = Attachment Redaction
    original_text,
    replacement,
    filter_type,
    match_count,
    created_at
FROM filter_matches
WHERE filter_id = 0  -- Nur Attachment-Redactions
ORDER BY created_at DESC
LIMIT 20;
```

### Abfragen via Admin API

```bash
# Request Logs
curl -H "X-Admin-API-Key: your_key" \
  http://localhost:8080/admin/logs

# Filter Statistics
curl -H "X-Admin-API-Key: your_key" \
  http://localhost:8080/admin/filters/stats
```

---

## 🎯 Level 3: Erweiterte Monitoring-Endpoints (NEU)

### 1. Attachment Analytics Endpoint

**Neu implementieren:** `/admin/attachments/stats`

```go
// GET /admin/attachments/stats
{
  "total_attachments_processed": 156,
  "pdf_count": 42,
  "image_count": 114,
  "total_redactions": 89,
  "ocr_failures": 3,
  "average_processing_time_ms": 2450,
  "last_24h": {
    "total": 23,
    "redacted": 15,
    "failed": 1
  },
  "by_type": {
    "pdf": {
      "count": 42,
      "redactions": 34,
      "avg_time_ms": 3200
    },
    "image": {
      "count": 114,
      "redactions": 55,
      "avg_time_ms": 1800
    }
  }
}
```

### 2. Real-time Activity Stream

**Neu implementieren:** `/admin/activity/stream` (WebSocket)

```javascript
// Frontend WebSocket Connection
const ws = new WebSocket('ws://localhost:8080/admin/activity/stream');

ws.onmessage = (event) => {
  const activity = JSON.parse(event.data);
  console.log(activity);
  // {
  //   type: "attachment_processed",
  //   timestamp: "2026-02-06T12:34:56Z",
  //   file_type: "pdf",
  //   redactions: 3,
  //   processing_time_ms: 2400
  // }
};
```

### 3. Detailed Request Logs

**Erweitern:** `/admin/logs/detail/{request_id}`

```go
// GET /admin/logs/detail/abc123
{
  "request_id": "abc123",
  "timestamp": "2026-02-06T12:34:56Z",
  "client_id": "client_xyz",
  "model": "claude-3-sonnet",
  "attachments": [
    {
      "type": "pdf",
      "filename": "document.pdf",
      "size_bytes": 245600,
      "ocr_words": 1245,
      "redactions": [
        {
          "type": "EMAIL",
          "original": "john@example.com",
          "replacement": "[EMAIL_ENTFERNT]",
          "location": { "x": 120, "y": 450, "page": 1 }
        },
        {
          "type": "PHONE",
          "original": "0123-456789",
          "replacement": "[TELEFON_ENTFERNT]",
          "location": { "x": 120, "y": 500, "page": 1 }
        }
      ],
      "processing_time_ms": 2400,
      "ocr_success": true
    }
  ],
  "filter_matches": 2,
  "total_tokens": 5600,
  "cost": 0.0234
}
```

---

## 📊 Level 4: Prometheus Metrics (Bereits implementiert)

### Existierende Metrics

```bash
# Metrics Endpoint
curl http://localhost:9090/metrics

# Wichtige Metrics:
# - llm_proxy_requests_total
# - llm_proxy_request_duration_seconds
# - llm_proxy_tokens_total
# - llm_proxy_cost_total
# - llm_proxy_cache_hits_total
# - llm_proxy_cache_misses_total
```

### Neue Attachment Metrics (Vorschlag)

```go
// Zu implementieren in internal/interfaces/api/metrics.go

var (
    attachmentProcessedTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_proxy_attachments_processed_total",
            Help: "Total number of attachments processed",
        },
        []string{"type", "status"}, // type: pdf/image, status: success/failed
    )
    
    attachmentRedactionsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_proxy_attachments_redactions_total",
            Help: "Total number of redactions in attachments",
        },
        []string{"type"}, // type: EMAIL/PHONE/CREDIT_CARD etc.
    )
    
    attachmentProcessingDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "llm_proxy_attachments_processing_duration_seconds",
            Help: "Time spent processing attachments",
            Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
        },
        []string{"type"}, // pdf/image
    )
    
    ocrFailuresTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "llm_proxy_ocr_failures_total",
            Help: "Total number of OCR failures",
        },
    )
)
```

---

## 📺 Level 5: Grafana Dashboard (Vorschlag)

### Dashboard Panels

#### 1. Overview Panel
```
┌─────────────────────────────────────────────────┐
│  Total Requests: 1,234  |  Attachments: 156    │
│  Total Redactions: 89   |  OCR Failures: 3     │
│  Avg Processing: 2.4s   |  Success Rate: 98%   │
└─────────────────────────────────────────────────┘
```

#### 2. Attachment Processing Graph
```
Attachments Processed (Last 24h)
    │
15  │     ▄▄
    │    ▄  ▄
10  │   ▄    ▄
    │  ▄      ▄▄
 5  │ ▄          ▄
    │▄              ▄
 0  └─────────────────────
    0h  6h  12h  18h  24h
```

#### 3. Redaction Heatmap
```
Redaction Types
EMAIL:        ████████████████ 45
PHONE:        ██████████ 23
CREDIT_CARD:  ██████ 12
CONFIDENTIAL: ████ 9
```

#### 4. Processing Time Distribution
```
Processing Time (seconds)
    │
50% │        ████
    │      ██    ██
25% │    ██        ██
    │  ██            ██
 0% └────────────────────
    0s  1s  2s  3s  4s  5s
```

### Grafana Configuration

```yaml
# grafana-dashboard.json (Beispiel)
{
  "dashboard": {
    "title": "LLM-Proxy Monitoring",
    "panels": [
      {
        "title": "Attachments Processed",
        "targets": [{
          "expr": "rate(llm_proxy_attachments_processed_total[5m])"
        }],
        "type": "graph"
      },
      {
        "title": "Redaction Types",
        "targets": [{
          "expr": "llm_proxy_attachments_redactions_total"
        }],
        "type": "pie"
      },
      {
        "title": "OCR Failures",
        "targets": [{
          "expr": "rate(llm_proxy_ocr_failures_total[1h])"
        }],
        "type": "stat"
      }
    ]
  }
}
```

---

## 🚨 Level 6: Alert Rules (Vorschlag)

### Prometheus Alert Rules

```yaml
# /etc/prometheus/alerts.yml
groups:
  - name: llm_proxy_alerts
    interval: 1m
    rules:
      # OCR Failures
      - alert: HighOCRFailureRate
        expr: rate(llm_proxy_ocr_failures_total[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High OCR failure rate detected"
          description: "OCR failing at {{ $value }} failures/sec"

      # Attachment Processing Slow
      - alert: SlowAttachmentProcessing
        expr: histogram_quantile(0.95, llm_proxy_attachments_processing_duration_seconds) > 10
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Attachment processing is slow"
          description: "95th percentile is {{ $value }}s (threshold: 10s)"

      # Dependencies Missing
      - alert: OCRDependenciesMissing
        expr: llm_proxy_ocr_failures_total > 0 and increase(llm_proxy_ocr_failures_total[1m]) == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "OCR dependencies might be missing"
          description: "Check if tesseract/pdftoppm/ghostscript are installed"

      # No Attachments Processed (Feature not working?)
      - alert: NoAttachmentsProcessed
        expr: rate(llm_proxy_attachments_processed_total[1h]) == 0 and rate(llm_proxy_requests_total[1h]) > 0
        for: 1h
        labels:
          severity: info
        annotations:
          summary: "No attachments processed recently"
          description: "Either no PDFs sent or feature not working"
```

---

## 💻 Level 7: Admin UI Erweiterungen

### Neues "Monitoring" Tab in Admin UI

```javascript
// admin-ui/src/components/Monitoring.svelte

<script>
  import { onMount } from 'svelte';
  
  let stats = {};
  let recentActivity = [];
  let ws;

  onMount(async () => {
    // Load stats
    const response = await fetch('/admin/attachments/stats', {
      headers: { 'X-Admin-API-Key': apiKey }
    });
    stats = await response.json();

    // Connect to activity stream
    ws = new WebSocket('ws://localhost:8080/admin/activity/stream');
    ws.onmessage = (event) => {
      const activity = JSON.parse(event.data);
      recentActivity = [activity, ...recentActivity].slice(0, 50);
    };
  });
</script>

<div class="monitoring-dashboard">
  <!-- Stats Cards -->
  <div class="stats-grid">
    <div class="stat-card">
      <h3>Attachments Processed</h3>
      <p class="stat-value">{stats.total_attachments_processed || 0}</p>
      <span class="stat-label">Total</span>
    </div>
    
    <div class="stat-card">
      <h3>Redactions</h3>
      <p class="stat-value">{stats.total_redactions || 0}</p>
      <span class="stat-label">PII Removed</span>
    </div>
    
    <div class="stat-card">
      <h3>OCR Failures</h3>
      <p class="stat-value">{stats.ocr_failures || 0}</p>
      <span class="stat-label">Errors</span>
    </div>
    
    <div class="stat-card">
      <h3>Avg Processing Time</h3>
      <p class="stat-value">{stats.average_processing_time_ms || 0}ms</p>
      <span class="stat-label">Performance</span>
    </div>
  </div>

  <!-- Real-time Activity Feed -->
  <div class="activity-feed">
    <h3>Recent Activity</h3>
    {#each recentActivity as activity}
      <div class="activity-item">
        <span class="timestamp">{activity.timestamp}</span>
        <span class="type">{activity.type}</span>
        {#if activity.redactions > 0}
          <span class="redactions">{activity.redactions} redactions</span>
        {/if}
        <span class="time">{activity.processing_time_ms}ms</span>
      </div>
    {/each}
  </div>
</div>
```

---

## 🔍 Debugging & Troubleshooting

### Check 1: Sind Dependencies installiert?

```bash
# Auf Production Server
ssh openweb "which tesseract && which pdftoppm && which gs && which convert"

# Sollte ausgeben:
# /usr/bin/tesseract
# /usr/bin/pdftoppm
# /usr/bin/gs
# /usr/bin/convert
```

**Falls nicht installiert:**

```bash
ssh openweb "sudo apt update && sudo apt install -y tesseract-ocr tesseract-ocr-deu tesseract-ocr-eng poppler-utils ghostscript imagemagick"
```

### Check 2: Läuft Attachment Service?

```bash
# Backend Logs prüfen
ssh openweb "docker logs llm-proxy-backend --tail 100 | grep -i attachment"

# Sollte Logs zeigen wenn PDFs verarbeitet werden
```

### Check 3: Sind Filter aktiv?

```bash
# Filter prüfen
curl -H "X-Admin-API-Key: your_key" \
  https://scrubgate.tech/admin/filters

# Mindestens ein Filter sollte enabled=true haben
```

### Check 4: Test Request senden

```bash
# Test PDF Base64 kodieren
echo "Test PDF" > test.txt
cat test.txt | base64

# Request senden
curl -X POST https://scrubgate.tech/v1/chat/completions \
  -H "Authorization: Bearer your_client_key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{
      "role": "user",
      "content": [
        {"type": "text", "text": "Analyze this"},
        {"type": "image_url", "image_url": {"url": "data:application/pdf;base64,VGVzdCBQREYK"}}
      ]
    }]
  }'

# Dann Logs prüfen
ssh openweb "docker logs llm-proxy-backend --tail 50 | grep -i 'redact\|attachment\|ocr'"
```

---

## 🛠️ Implementation Plan

### Phase 1: Dependencies (SOFORT)

```bash
# 1. Dependencies auf Production installieren
ssh openweb << 'EOF'
sudo apt update
sudo apt install -y \
  tesseract-ocr \
  tesseract-ocr-deu \
  tesseract-ocr-eng \
  poppler-utils \
  ghostscript \
  imagemagick

# Verify
tesseract --version
pdftoppm -v
gs --version
convert -version
EOF

# 2. Backend neu starten
ssh openweb "cd /opt/llm-proxy/deployments && docker compose restart backend"
```

### Phase 2: Monitoring Endpoints (Code-Änderungen)

```go
// File: internal/interfaces/api/monitoring_handler.go (NEU)

func (h *AdminHandler) GetAttachmentStats(w http.ResponseWriter, r *http.Request) {
    // Query database for attachment statistics
    stats := h.attachmentService.GetStatistics(r.Context())
    
    h.respondJSON(w, http.StatusOK, stats)
}

func (h *AdminHandler) StreamActivity(w http.ResponseWriter, r *http.Request) {
    // WebSocket endpoint for real-time activity
    // Implementation with gorilla/websocket
}

// Routes registrieren in router.go
r.Get("/admin/attachments/stats", h.GetAttachmentStats)
r.Get("/admin/activity/stream", h.StreamActivity)
```

### Phase 3: Prometheus Metrics (Code-Änderungen)

```go
// File: internal/application/attachment/service.go
// Metrics hinzufügen in AnalyzeAttachments()

// Am Anfang der Funktion
start := time.Now()
defer func() {
    duration := time.Since(start).Seconds()
    attachmentProcessingDuration.WithLabelValues(fileType).Observe(duration)
}()

// Bei erfolgreicher Redaction
if result.Success {
    attachmentProcessedTotal.WithLabelValues(fileType, "success").Inc()
    attachmentRedactionsTotal.WithLabelValues(piiType).Add(float64(result.TotalRedactions))
}

// Bei Fehler
if err != nil {
    attachmentProcessedTotal.WithLabelValues(fileType, "failed").Inc()
    if strings.Contains(err.Error(), "tesseract") {
        ocrFailuresTotal.Inc()
    }
}
```

### Phase 4: Admin UI Dashboard (Frontend)

```bash
# Neues Monitoring Component
touch admin-ui/src/components/Monitoring.svelte

# In Router einbinden
# admin-ui/src/App.svelte
```

### Phase 5: Grafana & Alerts (Optional)

```bash
# Grafana Container hinzufügen
# docker-compose.yml
```

---

## 📊 Quick Win: Sofort-Lösung

**Minimal-Setup (5 Minuten):**

```bash
# 1. Dependencies installieren
ssh openweb "sudo apt update && sudo apt install -y tesseract-ocr tesseract-ocr-deu tesseract-ocr-eng poppler-utils ghostscript imagemagick"

# 2. Backend neu starten
ssh openweb "cd /opt/llm-proxy/deployments && docker compose restart backend"

# 3. Logs live verfolgen
ssh openweb "docker logs llm-proxy-backend -f" | grep -i "attachment\|redact\|ocr"

# 4. Test PDF senden (siehe oben)

# 5. Filter Matches prüfen
ssh openweb "docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c \"SELECT * FROM filter_matches WHERE filter_id = 0 ORDER BY created_at DESC LIMIT 10;\""
```

---

## 📚 Zusammenfassung

### Existierendes Monitoring (Bereits verfügbar)

✅ **Structured Logging** - In Code implementiert  
✅ **Database Logging** - request_logs, filter_matches  
✅ **Prometheus Metrics** - /metrics endpoint  
✅ **Admin API** - /admin/logs, /admin/filters/stats

### Fehlt aktuell (Muss implementiert werden)

❌ **OCR Dependencies auf Production** - KRITISCH!  
❌ **Attachment-spezifische Metrics** - Prometheus  
❌ **Real-time Activity Stream** - WebSocket  
❌ **Monitoring Dashboard** - Admin UI Tab  
❌ **Alert Rules** - Prometheus Alerts  
❌ **Grafana Dashboards** - Visualisierung

### Priorität

1. **SOFORT:** Dependencies installieren (5 min)
2. **Kurzfristig:** Attachment Metrics hinzufügen (2h)
3. **Mittelfristig:** Admin UI Dashboard (4h)
4. **Langfristig:** Grafana + Alerts (8h)

---

**Last Updated:** 2026-02-06  
**Status:** Dependencies fehlen auf Production!  
**Action Required:** Installation durchführen
