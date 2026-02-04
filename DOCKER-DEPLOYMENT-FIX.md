# Docker Deployment Fix

**Datum:** 4. Februar 2026  
**Problem:** GitLab CI/CD Deployment schlägt fehl mit Container-Name-Konflikt

---

## 🔴 Problem

Deployment-Fehler beim letzten CI/CD Run:

```
Error response from daemon: Conflict. The container name "/llm-proxy-admin-ui" 
is already in use by container "e17b13de1a4d66544592b4ac127e22d355e7d5f3abb6d92ebf08471d1187a0dd". 
You have to remove (or rename) that container to be able to reuse that name.
```

### Root Cause

1. **`docker compose down` unvollständig:**
   ```
   Network llm-proxy-network  Resource is still in use
   ```
   - Network konnte nicht entfernt werden (Container noch aktiv?)
   - Alte Container wurden nicht vollständig gestoppt/entfernt

2. **Orphaned Container:**
   - Container `llm-proxy-admin-ui` existiert noch
   - Blockiert die Neuerstellung beim `docker compose up`

---

## ✅ Sofortlösung (auf LIVE Server)

### Option 1: Force Remove (Empfohlen)

```bash
# SSH auf LIVE Server
ssh openweb

# Alle LLM-Proxy Container zwangsweise stoppen und entfernen
docker stop llm-proxy-admin-ui llm-proxy-backend llm-proxy-postgres llm-proxy-redis || true
docker rm -f llm-proxy-admin-ui llm-proxy-backend llm-proxy-postgres llm-proxy-redis || true

# Network entfernen (falls notwendig)
docker network rm llm-proxy-network || true

# Danach erneut deployen (über GitLab CI oder manuell)
cd /path/to/llm-proxy
docker compose -f docker-compose.openwebui.yml up -d --build
```

### Option 2: Full Cleanup

```bash
# Alle gestoppten Container entfernen
docker container prune -f

# Alle ungenutzten Networks entfernen
docker network prune -f

# Alte Images aufräumen (optional)
docker image prune -a -f --filter "until=24h"

# Danach neu deployen
docker compose -f docker-compose.openwebui.yml up -d --build
```

---

## 🔧 Langfristige Lösung: GitLab CI Verbesserung

### Problem in `.gitlab-ci.yml`

Aktuell:
```yaml
- docker compose -f docker-compose.openwebui.yml down --remove-orphans || true
- docker compose -f docker-compose.openwebui.yml up -d --build
```

Das `|| true` verschluckt Fehler und führt zu inkonsistentem Zustand.

### Verbesserter Deploy-Step

```yaml
deploy:localhost:
  stage: deploy
  script:
    # ... (vorherige Steps) ...
    
    # Schritt 1: Aggressive Cleanup
    - echo "🧹 Cleaning up existing containers..."
    - docker compose -f docker-compose.openwebui.yml down --remove-orphans --volumes || true
    - docker ps -a | grep llm-proxy | awk '{print $1}' | xargs docker rm -f || true
    - docker network ls | grep llm-proxy | awk '{print $1}' | xargs docker network rm || true
    
    # Schritt 2: Warten auf vollständige Bereinigung
    - sleep 2
    
    # Schritt 3: Images neu bauen (ohne Cache bei Problemen)
    - echo "🏗️ Building images..."
    - docker compose -f docker-compose.openwebui.yml build --no-cache
    
    # Schritt 4: Container starten
    - echo "🚀 Starting services..."
    - docker compose -f docker-compose.openwebui.yml up -d
    
    # Schritt 5: Health Check
    - echo "🏥 Waiting for services..."
    - sleep 10
    - docker compose -f docker-compose.openwebui.yml ps
    - curl -f http://localhost:8080/health || exit 1
```

---

## 📊 Diagnose-Commands

### Container-Status prüfen

```bash
# Alle Container anzeigen (inkl. gestoppte)
docker ps -a | grep llm-proxy

# Container-IDs von LLM-Proxy-Containern
docker ps -a --filter "name=llm-proxy" --format "{{.ID}} {{.Names}} {{.Status}}"

# Welcher Container blockiert das Network?
docker network inspect llm-proxy-network
```

### Log-Analyse

```bash
# Logs vom problematischen Container
docker logs e17b13de1a4d66544592b4ac127e22d355e7d5f3abb6d92ebf08471d1187a0dd

# Compose Logs
docker compose -f docker-compose.openwebui.yml logs --tail=50
```

---

## 🎯 Präventivmaßnahmen

### 1. Deployment-Script erstellen

**Datei:** `scripts/deploy-safe.sh`

```bash
#!/bin/bash
set -euo pipefail

echo "🛑 Stopping all LLM-Proxy services..."
docker compose -f docker-compose.openwebui.yml down --remove-orphans --volumes || true

echo "🧹 Force-removing any orphaned containers..."
docker ps -a | grep llm-proxy | awk '{print $1}' | xargs -r docker rm -f || true

echo "🌐 Removing networks..."
docker network ls | grep llm-proxy | awk '{print $1}' | xargs -r docker network rm || true

echo "⏳ Waiting for cleanup to complete..."
sleep 3

echo "🏗️ Building images..."
docker compose -f docker-compose.openwebui.yml build

echo "🚀 Starting services..."
docker compose -f docker-compose.openwebui.yml up -d

echo "⏳ Waiting for services to be healthy..."
sleep 10

echo "🏥 Health check..."
docker compose -f docker-compose.openwebui.yml ps
curl -f http://localhost:8080/health && echo "✅ Deployment successful!" || echo "❌ Health check failed!"
```

### 2. GitLab CI anpassen

```yaml
deploy:localhost:
  stage: deploy
  script:
    # ... (setup steps) ...
    
    # Deployment-Script nutzen
    - chmod +x scripts/deploy-safe.sh
    - ./scripts/deploy-safe.sh
```

---

## 📝 Was beim letzten Deployment passiert ist

**Timestamp:** Job #XXXXX (siehe GitLab Log)

### Erfolgreiche Schritte:
✅ Docker Images gebaut (Backend + Admin-UI)  
✅ Backend, Postgres, Redis gestoppt und entfernt  
✅ Backend, Postgres, Redis neu erstellt  

### Fehlgeschlagener Schritt:
❌ Admin-UI Container konnte nicht erstellt werden  
**Grund:** Alter Container mit gleichem Namen existiert noch  

### Workaround (manuell angewendet):
```bash
# Auf LIVE Server:
docker rm -f e17b13de1a4d66544592b4ac127e22d355e7d5f3abb6d92ebf08471d1187a0dd
docker compose -f docker-compose.openwebui.yml up -d
```

---

## 🔍 Weitere Untersuchung notwendig

### Fragen:
1. **Warum konnte das Network nicht entfernt werden?**
   - Möglicherweise andere Container im gleichen Network?
   - Prüfen: `docker network inspect llm-proxy-network`

2. **Läuft ein zweiter Container außerhalb von Compose?**
   - Manuell gestartete Container?
   - Alte Deployments?

3. **GitLab Runner Container-Cleanup?**
   - Läuft der Runner selbst in einem Container?
   - Kann er andere Container überhaupt stoppen?

### Debug-Commands:
```bash
# Alle laufenden Container
docker ps

# Alle Container (inkl. gestoppte) mit Details
docker ps -a --format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Image}}"

# Networks mit Details
docker network ls
docker network inspect llm-proxy-network
```

---

## ✅ Nächste Schritte

**Priorität:** HOCH (blockiert Deployments)

1. **Sofort:** Manuell Container entfernen (siehe Option 1)
2. **Kurzfristig:** `scripts/deploy-safe.sh` erstellen und testen
3. **Mittelfristig:** `.gitlab-ci.yml` mit robustem Cleanup aktualisieren
4. **Monitoring:** Container-Status ins Monitoring aufnehmen

---

**Status:** 🔴 OFFEN (Deployment blockiert)  
**Owner:** DevOps / @krieger  
**Letzte Aktualisierung:** 4. Februar 2026
