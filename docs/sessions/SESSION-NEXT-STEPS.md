# Session Next Steps - Gespeichert für später

**Datum:** 4. Februar 2026  
**Letzte Session-Zusammenfassung:** Siehe `RESUME-PROJECT.md`  
**Status:** Session pausiert, Fortsetzung geplant

---

## 🚨 Dringende Probleme (vor nächster Session beheben)

### 1. Docker Deployment Konflikt ✅ **BEHOBEN**

**Problem:** Letzter GitLab CI/CD Deployment-Job fehlgeschlagen

**Status:** ✅ **BEHOBEN** in Commit `551e54f`

**Was wurde gemacht:**
- Aggressiver Container-Cleanup in `.gitlab-ci.yml` hinzugefügt
- Force-remove aller llm-proxy Container VOR docker compose down
- Network-Cleanup falls blockiert
- 3-Sekunden Wait für sauberen Cleanup

**Nächstes Deployment sollte funktionieren!** 🚀

---

## 📋 Geplante Aufgaben für nächste Session

### Priorität 1: Attachment Redaction Feature vervollständigen

**Status:** 70% implementiert, nicht getestet

**Was fehlt:**
1. ✅ Code ist implementiert (`internal/application/attachment/`)
2. ✅ Tools lokal installiert (Tesseract, ImageMagick, Ghostscript)
3. ❓ Tools auf LIVE Server installieren
4. ❌ Feature testen (PDF mit PII hochladen)
5. ❌ Dokumentation schreiben
6. ❌ Integration in Admin-UI (optional)

**Schritte:**

#### Schritt 1: Tools auf LIVE Server installieren (10 min)
```bash
# Script hochladen
scp scripts/install-redaction-tools.sh openweb:/tmp/

# SSH auf Server
ssh openweb

# Script ausführen (benötigt sudo)
chmod +x /tmp/install-redaction-tools.sh
sudo /tmp/install-redaction-tools.sh

# Verifizieren
which tesseract convert gs pdftoppm
tesseract --version
```

#### Schritt 2: Test-Script erstellen (15 min)

**Datei:** `scripts/test-attachment-redaction.sh`

```bash
#!/bin/bash
# Test des Attachment Redaction Features

API_KEY="admin_dev_key_12345678901234567890123456789012"
BASE_URL="https://llmproxy.aitrail.ch"

# 1. Test-PDF mit PII erstellen (oder manuell hochladen)
# 2. PDF als Base64 kodieren
# 3. Chat-Request mit Attachment senden
# 4. Verifizieren dass PII geschwärzt wurde
```

#### Schritt 3: Dokumentation schreiben (20 min)

**Datei:** `docs/ATTACHMENT-REDACTION-GUIDE.md`

**Inhalt:**
- Feature-Übersicht
- Unterstützte Formate (PDF, Bilder)
- PII-Typen die erkannt werden
- API-Beispiele
- Konfiguration
- Troubleshooting

#### Schritt 4: README aktualisieren (5 min)
- Link zum Redaction-Guide hinzufügen
- Features-Sektion erweitern

**Gesamtaufwand:** ~50 Minuten

---

### Priorität 2: Security Hardening (2-3 Stunden)

**Warum wichtig:** Produktions-System mit API-Keys

#### 2.1 Rate Limiting pro API Key

**Aktuell:** Global Rate Limiting vorhanden  
**Fehlt:** Per-Key-Limiting

**Implementation:**
1. Redis-basierter Counter pro API Key
2. Konfigurierbare Limits (requests/minute, requests/day)
3. HTTP 429 Response bei Überschreitung
4. Admin-UI: Rate-Limit-Status anzeigen

**Dateien:**
- `internal/application/ratelimit/per_key_limiter.go` (neu)
- `internal/interfaces/api/middleware/rate_limit.go` (erweitern)
- `docs/RATE-LIMITING.md` (neu)

#### 2.2 Request Size Limits

**Aktuell:** Keine expliziten Limits  
**Problem:** Large Payloads können Server überlasten

**Implementation:**
1. Gin MaxMultipartMemory konfigurieren
2. Custom Middleware für Request Body Size
3. Separate Limits für Chat vs. Attachments

**Config-Beispiel:**
```yaml
limits:
  max_request_body: 10MB
  max_attachment_size: 5MB
  max_attachments_per_request: 5
```

#### 2.3 Audit Logging

**Aktuell:** Basic Logging vorhanden  
**Fehlt:** Security-Event-Tracking

**Events zu loggen:**
- API-Key-Erstellung/Löschung
- Content-Filter-Änderungen
- Rate-Limit-Violations
- Authentifizierungs-Fehler
- Admin-Aktionen

**Implementation:**
- `internal/application/audit/service.go` (neu)
- Postgres-Tabelle: `audit_logs`
- Admin-UI: Audit-Log-Viewer

---

### Priorität 3: Monitoring & Observability (3-4 Stunden)

#### 3.1 Prometheus Metrics

**Metriken hinzufügen:**
```go
// Request Metrics
llm_proxy_requests_total (Counter) - label: provider, model, status
llm_proxy_request_duration_seconds (Histogram)

// Content Filtering
llm_proxy_content_violations_total (Counter) - label: filter_type
llm_proxy_tokens_redacted_total (Counter)

// Cache Metrics
llm_proxy_cache_hits_total (Counter)
llm_proxy_cache_misses_total (Counter)

// Provider Metrics
llm_proxy_provider_errors_total (Counter) - label: provider
llm_proxy_provider_latency_seconds (Histogram) - label: provider
```

**Implementation:**
- `go get github.com/prometheus/client_golang`
- `internal/infrastructure/metrics/prometheus.go` (neu)
- `/metrics` Endpoint in `cmd/server/main.go`

#### 3.2 Grafana Dashboard

**Dashboards erstellen:**
1. **Overview Dashboard**
   - Total Requests (Timeline)
   - Requests by Provider
   - Error Rate
   - P95/P99 Latency

2. **Security Dashboard**
   - Content Violations (Heatmap)
   - Rate Limit Violations
   - Failed Auth Attempts

3. **Performance Dashboard**
   - Provider Latency Comparison
   - Cache Hit Rate
   - Database Query Performance

**Datei:** `monitoring/grafana-dashboard.json` (exportieren)

#### 3.3 Alert Rules

**Alertmanager Config:**
```yaml
alerts:
  - name: HighErrorRate
    expr: rate(llm_proxy_requests_total{status=~"5.."}[5m]) > 0.1
    
  - name: HighLatency
    expr: histogram_quantile(0.95, llm_proxy_request_duration_seconds) > 5
    
  - name: CacheDegraded
    expr: rate(llm_proxy_cache_hits_total[5m]) / rate(llm_proxy_cache_requests_total[5m]) < 0.3
```

---

### Priorität 4: Performance Optimization (2-3 Stunden)

#### 4.1 Database Connection Pooling

**Aktuell:** Defaults von GORM  
**Verbesserung:** Tuning für 1-CPU-Server

```go
// internal/infrastructure/database/postgres.go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(10)      // default: unbegrenzt
sqlDB.SetMaxIdleConns(5)       // default: 2
sqlDB.SetConnMaxLifetime(30m)  // default: unbegrenzt
```

#### 4.2 Redis Connection Pooling

```go
// internal/infrastructure/cache/redis.go
&redis.Options{
    PoolSize:     10,
    MinIdleConns: 2,
    MaxRetries:   3,
}
```

#### 4.3 Cache Warming

**Problem:** Cold Start = langsame erste Requests  
**Lösung:** Startup-Script zum Vorwärmen

```go
// cmd/server/main.go
func warmCaches(ctx context.Context, svc *app.Service) {
    // Lade häufig genutzte Modelle
    svc.ModelService.ListModels(ctx)
    
    // Lade aktive Content-Filter
    svc.FilterService.ListFilters(ctx)
    
    // Lade API-Keys
    svc.APIKeyService.ListKeys(ctx)
}
```

---

### Priorität 5: Admin-UI Enhancements (4-5 Stunden)

#### 5.1 Real-time Dashboard Updates

**Technologie:** Server-Sent Events (SSE) oder WebSocket

**Features:**
- Live Request Counter
- Active Users/Keys
- Current Error Rate
- Provider Health Status

**Implementation:**
- Backend: `/api/v1/stream/metrics` Endpoint
- Frontend: EventSource API oder WebSocket

#### 5.2 Advanced Filtering UI

**Verbesserungen:**
- Bulk-Import von Content-Filtern (CSV/JSON)
- Regex-Testing-Tool (live Preview)
- Filter-Templates (z.B. "E-Mail-Adressen", "Kreditkarten")
- Filter-Gruppen (aktivieren/deaktivieren mehrerer Filter)

#### 5.3 Usage Analytics

**Visualisierungen:**
- Requests per Day (Chart.js Line Chart)
- Top Models (Bar Chart)
- Top Users/Keys (Table)
- Cost per Provider (Stacked Bar Chart)
- Content Violations Timeline (Heatmap)

**Backend:**
- Neue Endpoints in `internal/interfaces/api/analytics_handler.go`
- Aggregation-Queries in PostgreSQL
- Optional: TimescaleDB Extension für bessere Performance

---

## 🛠️ Technische Schulden & Refactoring

### TODO-Kommentare im Code

**Gefunden in letzter Session:**

1. **`internal/application/attachment/service.go`**
   ```go
   // TODO: Implement PDF text extraction using a library
   // TODO: Implement DOCX text extraction
   // TODO: Implement OCR using tesseract
   ```
   **Status:** Teilweise erledigt (OCR in `redaction_service.go`)  
   **Action:** Kommentare aktualisieren oder entfernen

2. **Error Handling verbessern**
   - Konsistente Error-Wrapping mit `fmt.Errorf("...: %w", err)`
   - Custom Error-Types für besseres Handling
   - Structured Logging statt fmt.Printf

3. **Test Coverage erhöhen**
   - Aktuell: Broken Tests entfernt
   - Ziel: >70% Coverage
   - Unit-Tests für neue Features schreiben

---

## 📚 Dokumentation TODO

1. **API Reference**
   - OpenAPI/Swagger Spec generieren
   - Interaktive API-Docs (Swagger UI)

2. **Architecture Decision Records (ADR)**
   - Warum Go statt Python/Node?
   - Warum Redis als Cache?
   - Warum keine K8s?

3. **Deployment Guide**
   - Step-by-Step Server-Setup
   - GitLab CI/CD Erklärung
   - Rollback-Prozedur

4. **User Guide**
   - Admin-UI Walkthrough (Screenshots)
   - API-Beispiele in verschiedenen Sprachen (curl, Python, Node.js)
   - Best Practices für API-Key-Management

---

## 🎯 Langfristige Roadmap

### Phase 1: Stabilisierung (2-3 Wochen)
- [x] Code Cleanup
- [ ] Attachment Redaction Feature vervollständigen
- [ ] Docker Deployment Fix
- [ ] Security Hardening
- [ ] Monitoring Setup

### Phase 2: Feature-Erweiterung (4-6 Wochen)
- [ ] Webhook Support für Callbacks
- [ ] Batch Request Processing
- [ ] Streaming Response Improvements
- [ ] Multi-Tenant Support (separate Datenbanken?)

### Phase 3: Skalierung (2-3 Monate)
- [ ] Horizontal Scaling (mehrere Backend-Instanzen)
- [ ] Load Balancing (Traefik/HAProxy)
- [ ] Distributed Caching (Redis Cluster)
- [ ] Database Replication (Read-Replicas)

### Phase 4: Enterprise Features (3-6 Monate)
- [ ] SSO Integration (OAuth2/SAML)
- [ ] Fine-grained Permissions (RBAC)
- [ ] Compliance Features (GDPR, SOC2)
- [ ] White-Label Admin-UI

---

## 📊 Metriken & Erfolgs-KPIs

### Aktueller Status (zu messen):
- [ ] Response Time P95: ??? ms
- [ ] Error Rate: ??? %
- [ ] Cache Hit Rate: ??? %
- [ ] Requests/Day: ???
- [ ] Cost/Request: ???

### Ziele (nach Optimierungen):
- Response Time P95: < 500ms
- Error Rate: < 1%
- Cache Hit Rate: > 60%
- Uptime: > 99.5%

---

## 🔐 Security Checklist

- [ ] Alle API-Keys in Environment Variables (nicht in Code)
- [ ] HTTPS enforced (Caddy configured)
- [ ] SQL Injection Prevention (GORM Prepared Statements)
- [ ] XSS Prevention (Frontend Sanitization)
- [ ] CSRF Protection (für Admin-UI)
- [ ] Rate Limiting (Global ✅, Per-Key ❌)
- [ ] Input Validation (Content-Filter ✅, Size Limits ❌)
- [ ] Audit Logging (❌)
- [ ] Regular Dependency Updates (Dependabot?)
- [ ] Security Headers (X-Frame-Options, CSP, etc.)

---

## 📞 Kontakte & Ressourcen

### GitLab Repository
- **URL:** https://gitlab.com/krieger-engineering/llm-proxy
- **CI/CD:** https://gitlab.com/krieger-engineering/llm-proxy/-/pipelines

### LIVE Server
- **Host:** `ssh openweb` (68.183.208.213)
- **URL:** https://llmproxy.aitrail.ch
- **Admin-UI:** https://llmproxy.aitrail.ch/ (Port 80 → Backend auf 8080)

### Dokumentation
- **Projekt-Docs:** `/docs` Verzeichnis
- **API-Docs:** (noch zu erstellen)
- **Confluence/Wiki:** (falls vorhanden?)

### Support
- **Admin API Key:** `admin_dev_key_12345678901234567890123456789012`

---

## ✅ Session-Abschluss-Checkliste

Vor Beenden der Session:

- [x] Wichtige Erkenntnisse dokumentiert
- [x] Next Steps priorisiert und gespeichert
- [ ] Docker-Problem behoben (MANUELL - siehe DOCKER-DEPLOYMENT-FIX.md)
- [x] Git committed (falls Code-Änderungen)
- [ ] Server-Status geprüft (optional)

---

**Letzte Aktualisierung:** 4. Februar 2026  
**Status:** 🟡 BEREIT FÜR NÄCHSTE SESSION  
**Geschätzter Zeitbedarf (Prio 1):** 50 Minuten  
**Geschätzter Zeitbedarf (Prio 1-3):** 8-10 Stunden  

---

## 🚀 Quick Start für nächste Session

```bash
# 1. Status prüfen
cd /home/krieger/Sites/golang-projekte/llm-proxy
git status
git log --oneline -5

# 2. Dokumentation lesen
cat SESSION-NEXT-STEPS.md
cat DOCKER-DEPLOYMENT-FIX.md

# 3. LIVE Server prüfen
ssh openweb "docker ps | grep llm-proxy"
curl https://llmproxy.aitrail.ch/health

# 4. Mit Priorität 1 starten (Attachment Redaction)
# ODER: Docker-Problem beheben falls Deployment noch blockiert
```

**Viel Erfolg bei der nächsten Session! 🎯**
