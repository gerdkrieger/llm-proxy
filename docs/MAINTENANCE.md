# LLM-Proxy Docker Maintenance Guide

Wartungs- und Management-Befehle für den LLM-Proxy Docker Stack.

## 📊 Status & Monitoring

### Stack Status anzeigen
```bash
docker compose -f docker-compose.openwebui.yml ps
```

### Detaillierter Status mit Health Checks
```bash
docker compose -f docker-compose.openwebui.yml ps --format "table {{.Name}}\t{{.Status}}\t{{.Ports}}"
```

### Ressourcen-Verbrauch
```bash
docker stats --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
```

### Disk-Space von Containern
```bash
docker system df -v
```

---

## 📝 Logs & Debugging

### Alle Logs anzeigen
```bash
docker compose -f docker-compose.openwebui.yml logs -f
```

### Logs eines bestimmten Services
```bash
docker compose -f docker-compose.openwebui.yml logs -f backend
docker compose -f docker-compose.openwebui.yml logs -f openwebui
docker compose -f docker-compose.openwebui.yml logs -f postgres
docker compose -f docker-compose.openwebui.yml logs -f redis
```

### Letzte 100 Zeilen
```bash
docker compose -f docker-compose.openwebui.yml logs --tail=100 backend
```

### Logs seit bestimmtem Zeitpunkt
```bash
docker compose -f docker-compose.openwebui.yml logs --since 30m backend
docker compose -f docker-compose.openwebui.yml logs --since "2026-02-01T10:00:00" backend
```

### Logs in Datei speichern
```bash
docker compose -f docker-compose.openwebui.yml logs backend > backend-logs.txt
```

---

## 🔄 Container Management

### Einzelnen Service neu starten
```bash
docker compose -f docker-compose.openwebui.yml restart backend
docker compose -f docker-compose.openwebui.yml restart openwebui
```

### Service stoppen
```bash
docker compose -f docker-compose.openwebui.yml stop backend
```

### Service starten
```bash
docker compose -f docker-compose.openwebui.yml start backend
```

### Kompletten Stack neu starten
```bash
docker compose -f docker-compose.openwebui.yml restart
```

### Service neu bauen und starten (nach Code-Änderungen)
```bash
docker compose -f docker-compose.openwebui.yml up -d --build backend
```

### Force recreate (bei Config-Änderungen)
```bash
docker compose -f docker-compose.openwebui.yml up -d --force-recreate backend
```

---

## 🗄️ Datenbank Management

### PostgreSQL Shell öffnen
```bash
docker exec -it llm-proxy-postgres psql -U proxy_user -d llm_proxy
```

### Datenbank Backup erstellen
```bash
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > backup-$(date +%Y%m%d-%H%M%S).sql
```

### Datenbank Backup mit Kompression
```bash
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy | gzip > backup-$(date +%Y%m%d-%H%M%S).sql.gz
```

### Backup wiederherstellen
```bash
cat backup.sql | docker exec -i llm-proxy-postgres psql -U proxy_user -d llm_proxy
```

### Datenbank-Größe prüfen
```bash
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
  SELECT pg_size_pretty(pg_database_size('llm_proxy')) AS size;"
```

### Tabellen-Größen anzeigen
```bash
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "
  SELECT tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
  FROM pg_tables
  WHERE schemaname = 'public'
  ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;"
```

---

## 🔴 Redis Management

### Redis CLI öffnen
```bash
docker exec -it llm-proxy-redis redis-cli
```

### Redis Speicher-Statistiken
```bash
docker exec llm-proxy-redis redis-cli INFO memory
```

### Alle Keys anzeigen
```bash
docker exec llm-proxy-redis redis-cli KEYS '*'
```

### Cache leeren (VORSICHT!)
```bash
docker exec llm-proxy-redis redis-cli FLUSHALL
```

### Redis Backup erstellen
```bash
docker exec llm-proxy-redis redis-cli SAVE
```

---

## 🔧 Updates & Deployments

### Neue Images von Registry pullen
```bash
docker compose -f docker-compose.openwebui.yml pull
```

### Services mit neuen Images aktualisieren
```bash
docker compose -f docker-compose.openwebui.yml up -d --no-build
```

### Lokales Rebuild + Deploy (nach Code-Änderungen)
```bash
docker compose -f docker-compose.openwebui.yml build backend
docker compose -f docker-compose.openwebui.yml up -d backend
```

### Kompletter Stack Update
```bash
docker compose -f docker-compose.openwebui.yml pull
docker compose -f docker-compose.openwebui.yml up -d
```

---

## 🧹 Cleanup & Wartung

### Gestoppte Container entfernen
```bash
docker compose -f docker-compose.openwebui.yml rm
```

### Ungenutzte Images entfernen
```bash
docker image prune -a
```

### Ungenutzte Volumes entfernen (VORSICHT: Datenverlust!)
```bash
docker volume prune
```

### Komplette Docker-Bereinigung
```bash
docker system prune -a --volumes
```

### Logs rotieren (wenn zu groß)
```bash
truncate -s 0 $(docker inspect --format='{{.LogPath}}' llm-proxy-backend)
```

---

## 📊 Health Checks

### Backend Health Check
```bash
curl http://localhost:8080/health
```

### OpenWebUI Health Check
```bash
curl http://localhost:3010/health
```

### PostgreSQL Health Check
```bash
docker exec llm-proxy-postgres pg_isready -U proxy_user -d llm_proxy
```

### Redis Health Check
```bash
docker exec llm-proxy-redis redis-cli PING
```

### Alle Health Checks in einem Befehl
```bash
echo "Backend:"; curl -s http://localhost:8080/health | jq .
echo "OpenWebUI:"; curl -s http://localhost:3010/health | jq .
echo "PostgreSQL:"; docker exec llm-proxy-postgres pg_isready
echo "Redis:"; docker exec llm-proxy-redis redis-cli PING
```

---

## 🔍 Troubleshooting

### Container läuft nicht
```bash
# 1. Status prüfen
docker compose -f docker-compose.openwebui.yml ps

# 2. Logs checken
docker compose -f docker-compose.openwebui.yml logs backend

# 3. Container inspect
docker inspect llm-proxy-backend

# 4. Network prüfen
docker network inspect llm-proxy-network
```

### Service kann nicht erreicht werden
```bash
# 1. Port-Bindings prüfen
docker compose -f docker-compose.openwebui.yml ps

# 2. Firewall prüfen
sudo ufw status
sudo iptables -L -n

# 3. Prozess auf Port prüfen
sudo lsof -i :8080
sudo netstat -tulpn | grep 8080
```

### Hoher Speicherverbrauch
```bash
# 1. Ressourcen anzeigen
docker stats

# 2. PostgreSQL Cache leeren
docker exec llm-proxy-postgres psql -U proxy_user -d llm_proxy -c "DISCARD ALL;"

# 3. Redis Memory Policy prüfen
docker exec llm-proxy-redis redis-cli CONFIG GET maxmemory-policy
```

### Netzwerk-Probleme zwischen Containern
```bash
# 1. Network inspect
docker network inspect llm-proxy-network

# 2. Container zu Network hinzufügen
docker network connect llm-proxy-network <container_name>

# 3. DNS Test
docker exec llm-proxy-backend ping -c 3 postgres
docker exec openwebui ping -c 3 backend
```

---

## 📦 Backup & Restore

### Komplettes Backup (Datenbank + Volumes)
```bash
# 1. PostgreSQL Backup
docker exec llm-proxy-postgres pg_dump -U proxy_user llm_proxy > db-backup.sql

# 2. Volumes sichern
docker run --rm -v llm-proxy-postgres-data:/data -v $(pwd):/backup ubuntu tar czf /backup/postgres-volume.tar.gz /data
docker run --rm -v llm-proxy-redis-data:/data -v $(pwd):/backup ubuntu tar czf /backup/redis-volume.tar.gz /data
docker run --rm -v open-webui:/data -v $(pwd):/backup ubuntu tar czf /backup/openwebui-volume.tar.gz /data
```

### Backup wiederherstellen
```bash
# 1. Stack stoppen
docker compose -f docker-compose.openwebui.yml down

# 2. Volumes wiederherstellen
docker run --rm -v llm-proxy-postgres-data:/data -v $(pwd):/backup ubuntu tar xzf /backup/postgres-volume.tar.gz -C /

# 3. Stack starten
docker compose -f docker-compose.openwebui.yml up -d
```

---

## 🔐 Security

### Secrets rotieren
```bash
# 1. .env Datei editieren (neue API Keys)
vim .env

# 2. Services mit neuen Secrets neu starten
docker compose -f docker-compose.openwebui.yml up -d --force-recreate backend
```

### SSL/TLS Zertifikate erneuern
```bash
# Wenn Reverse Proxy (Nginx/Traefik) verwendet wird
docker compose -f docker-compose.openwebui.yml exec nginx nginx -s reload
```

---

## 📈 Monitoring Zugriff

- **Prometheus Metrics:** http://localhost:9090
- **Grafana Dashboards:** http://localhost:3001
- **Backend Metrics:** http://localhost:8080/metrics
- **OpenWebUI:** http://localhost:3010
- **Admin UI:** http://localhost:3005

---

## 🚨 Notfall-Befehle

### Kompletter Stack-Neustart
```bash
docker compose -f docker-compose.openwebui.yml down
docker compose -f docker-compose.openwebui.yml up -d
```

### Alle Container stoppen
```bash
docker stop $(docker ps -aq)
```

### Factory Reset (VORSICHT: Alle Daten gehen verloren!)
```bash
docker compose -f docker-compose.openwebui.yml down -v
docker system prune -a --volumes
docker compose -f docker-compose.openwebui.yml up -d
```

---

## 📞 Support

Bei Problemen:
1. Logs prüfen: `docker compose -f docker-compose.openwebui.yml logs -f`
2. Health Checks: Siehe Abschnitt "Health Checks"
3. GitHub Issues: [llm-proxy/issues](https://github.com/yourusername/llm-proxy/issues)
