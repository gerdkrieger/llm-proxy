# LLM-PROXY DEPLOYMENT GUIDE

## ⚠️ WICHTIG: Manuelles Deployment verwenden

Die GitLab CI/CD Pipeline ist **DEAKTIVIERT** weil:
- Registry Push scheitert (Permission Probleme)
- Registry Images sind alt/kaputt
- Automatisches Deployment würde Production zerstören

**Verwende IMMER den manuellen Deployment-Prozess unten!**

---

## 🚀 Manuelles Deployment (FUNKTIONIERT 100%)

### 1. Lokale Images bauen

```bash
cd /home/krieger/Sites/golang-projekte/llm-proxy

# Backend Image bauen
docker build -t llm-proxy-backend:latest -f deployments/docker/Dockerfile .

# Admin-UI Image bauen (mit .dockerignore gegen localhost:8080 Bug)
cd admin-ui
docker build -t llm-proxy-admin-ui:latest -f Dockerfile .
cd ..
```

### 2. Images zum Server übertragen

```bash
# Backend übertragen
docker save llm-proxy-backend:latest | ssh openweb "docker load"

# Admin-UI übertragen
docker save llm-proxy-admin-ui:latest | ssh openweb "docker load"
```

### 3. Services auf Production neu starten

```bash
ssh openweb "cd /opt/llm-proxy/deployments && docker compose -f docker-compose.openwebui.yml up -d --force-recreate backend admin-ui"
```

### 4. Health Check

```bash
ssh openweb "docker ps --filter 'name=llm-proxy' && curl -s http://localhost:8080/health"
```

**Erwartetes Ergebnis:**
```
llm-proxy-backend    Up XX seconds (healthy)
llm-proxy-admin-ui   Up XX seconds (healthy)
llm-proxy-postgres   Up XX seconds (healthy)
llm-proxy-redis      Up XX seconds (healthy)

{"status":"ok",...}
```

---

## 🔍 Troubleshooting

### Admin-UI zeigt "localhost:8080" Fehler

**Ursache:** Alte Images ohne .dockerignore

**Fix:**
```bash
# Prüfe ob localhost:8080 im Bundle ist
ssh openweb "docker exec llm-proxy-admin-ui sh -c 'cat /usr/share/nginx/html/assets/index-*.js' | grep -c localhost:8080"

# Wenn > 0: Images neu bauen mit .dockerignore
cd admin-ui
docker build -t llm-proxy-admin-ui:latest -f Dockerfile .
docker save llm-proxy-admin-ui:latest | ssh openweb "docker load"
ssh openweb "cd /opt/llm-proxy/deployments && docker compose -f docker-compose.openwebui.yml up -d --force-recreate admin-ui"
```

### Backend crasht mit DB Auth Fehler

**Ursache:** .env nicht geladen

**Fix:**
```bash
# Prüfe ob .env existiert
ssh openweb "ls -la /opt/llm-proxy/deployments/.env"

# Prüfe DB Credentials
ssh openweb "grep DB_PASSWORD /opt/llm-proxy/deployments/.env"
ssh openweb "docker exec llm-proxy-backend env | grep DATABASE_PASSWORD"

# Wenn unterschiedlich: Container neu starten
ssh openweb "cd /opt/llm-proxy/deployments && docker compose -f docker-compose.openwebui.yml restart backend"
```

---

## ❌ NICHT VERWENDEN

### GitLab Pipeline (DEAKTIVIERT)

Die Pipeline ist auf `when: manual` gesetzt und sollte **NICHT** verwendet werden weil:

1. **docker:backend** - Baut Images aber Registry Push scheitert
2. **docker:admin-ui** - Baut Images aber Registry Push scheitert  
3. **deploy:production** - Würde alte kaputte Registry Images deployen

**Wenn du die Pipeline trotzdem manuell triggerst:**
- Docker Builds laufen aber pushen nicht zum Registry
- deploy:production pulled ALTE KAPUTTE Images aus Registry
- **PRODUCTION GEHT KAPUTT** (localhost:8080 Bug kehrt zurück)

---

## ✅ Produktion URLs

- **Backend API:** https://llmproxy.aitrail.ch
- **Admin UI:** https://llmproxy.aitrail.ch (Port 3005)
- **Health Check:** https://llmproxy.aitrail.ch/health
- **OpenWebUI:** https://chat.aitrail.ch

---

## 📝 Wichtige Dateien

- `.gitlab-ci.yml` - Pipeline (DEAKTIVIERT mit when: manual)
- `admin-ui/.dockerignore` - Verhindert localhost:8080 Bug
- `deployments/.env` - Production Environment Variables (auf Server)
- `deployments/docker-compose.openwebui.yml` - Production Stack

---

## 🎯 Zusammenfassung

**IMMER verwenden:** Manueller Deployment-Prozess (Abschnitt oben)  
**NIEMALS verwenden:** GitLab CI/CD Pipeline

Der manuelle Prozess dauert 5 Minuten und funktioniert 100% zuverlässig.
