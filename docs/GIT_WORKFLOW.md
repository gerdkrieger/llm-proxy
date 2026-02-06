# Git Workflow für LLM-Proxy

## Branch-Strategie

Das Projekt verwendet eine **Two-Branch-Strategie**:

```
develop (Development)  →  master (Production/Stable)
```

### Branch-Zwecke

| Branch | Zweck | Commits |
|--------|-------|---------|
| **`develop`** | Aktive Entwicklung | Bei jedem positiven Ergebnis |
| **`master`** | Stabile Releases | Nur via Merge von develop |

---

## Täglicher Workflow

### 1. Feature/Bugfix entwickeln (auf `develop`)

```bash
# Sicherstellen dass du auf develop bist
git checkout develop

# Neueste Änderungen holen (falls Team arbeitet)
git pull origin develop

# Änderungen machen...
# Testen...
# Positives Ergebnis erreicht ✅

# Änderungen committen
git add .
git commit -m "feat: Beschreibung der Änderung"
```

### 2. Bei jedem positiven Ergebnis committen

**Positives Ergebnis** bedeutet:
- ✅ Feature funktioniert
- ✅ Bug gefixt
- ✅ Tests laufen durch
- ✅ Keine Syntax-Fehler
- ✅ Lokale Dev-Environment funktioniert

**Commit-Konvention:**
```bash
feat: Neue Funktion hinzugefügt
fix: Bug behoben
docs: Dokumentation aktualisiert
refactor: Code umstrukturiert
test: Tests hinzugefügt
chore: Build/Config Änderungen
```

### 3. Nach mehreren Commits: Merge nach `master`

**Wann?**
- Am Ende eines Arbeitstages (wenn develop stabil ist)
- Vor einem Deployment
- Wenn ein Milestone erreicht ist

```bash
# Sicherstellen develop ist committed & gepusht
git checkout develop
git status  # Muss clean sein

# Zu master wechseln
git checkout master

# develop in master mergen
git merge develop --no-edit

# Prüfen
git log --oneline -5
```

### 4. Beide Branches pushen

```bash
# develop pushen
git checkout develop
git push origin develop

# master pushen
git checkout master
git push origin master

# Zurück zu develop für weitere Entwicklung
git checkout develop
```

---

## Schnell-Referenz

### Typischer Development-Cycle

```bash
# 1. Auf develop entwickeln
git checkout develop
# ... Code schreiben ...
git add .
git commit -m "feat: Feature X implementiert"

# 2. Weiteres Feature
# ... Code schreiben ...
git add .
git commit -m "fix: Bug Y behoben"

# 3. Wenn stabiler Zustand erreicht → merge nach master
git checkout master
git merge develop --no-edit

# 4. Beide pushen
git push origin develop
git push origin master

# 5. Zurück zu develop
git checkout develop
```

---

## Deployment-Workflow

### Production Deployment

```bash
# 1. Sicherstellen master ist aktuell
git checkout master
git pull origin master

# 2. Deployment ausführen
./deploy.sh

# Nach erfolgreichem Deployment: Tag erstellen (optional)
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### Local Development

```bash
# Auf develop entwickeln
git checkout develop

# Docker Compose starten
docker compose -f docker-compose.dev.yml up -d

# Nach Änderungen: Hot-Reload (Vite + Air)
# Kein Neustart nötig!
```

---

## Branch-Status prüfen

### Sind Branches synchron?

```bash
# Zeige Unterschiede
git log develop..master --oneline   # Was ist in master aber nicht in develop?
git log master..develop --oneline   # Was ist in develop aber nicht in master?

# Grafische Übersicht
git log --oneline --graph --all --decorate -20
```

### Remote vs. Local

```bash
# Lokaler develop vs. remote develop
git fetch
git log develop..origin/develop --oneline

# Status
git status
```

---

## Häufige Situationen

### Situation 1: Versehentlich auf `master` committed

```bash
# Falls noch nicht gepusht:
git checkout master
git log --oneline -3  # Commit-Hash merken

# Cherry-pick nach develop
git checkout develop
git cherry-pick <commit-hash>

# Commit von master entfernen
git checkout master
git reset --hard HEAD~1  # Vorsicht! Nur wenn nicht gepusht!
```

### Situation 2: develop und master sind nicht synchron

```bash
# master ist voraus (sollte nicht passieren)
git checkout develop
git merge master --no-edit

# develop ist voraus (normal)
git checkout master
git merge develop --no-edit
```

### Situation 3: Merge-Konflikt

```bash
# Beim Merge von develop nach master
git checkout master
git merge develop

# Wenn Konflikt:
# 1. Konflikte in Dateien lösen (Editor)
# 2. Resolved Dateien stagen
git add <datei>

# 3. Merge abschließen
git commit --no-edit
```

---

## Best Practices

### ✅ DO

- **Immer auf `develop` entwickeln**
- **Kleine, häufige Commits** (bei jedem positiven Ergebnis)
- **Beschreibende Commit-Messages** (feat/fix/docs/etc.)
- **Vor Merge: Tests laufen lassen**
- **master bleibt immer stabil** (deploybar)

### ❌ DON'T

- **Nicht direkt auf `master` committen** (nur via merge)
- **Keine "WIP" Commits pushen** (erst committen wenn fertig)
- **Keine Merge-Commits rückgängig machen** (auf remote)
- **Kein `git push --force`** (außer in Notfällen, nach Absprache)

---

## GitLab CI/CD Integration

### Automatische Pipeline Trigger

| Branch | Pipeline | Wann |
|--------|----------|------|
| `develop` | Lint + Tests | Bei jedem Push |
| `master` | Lint + Tests + Docker Build | Bei jedem Push |

### Manuelle Jobs

- **`docker:backend`** - Build Backend Image (manual)
- **`docker:admin-ui`** - Build Admin UI Image (manual)
- **`deploy:production`** - Deploy to Production (manual)

**Empfohlen:** Nutze `deploy.sh` statt CI/CD Pipeline für Deployments.

---

## Projekt-spezifische Hinweise

### Lokale Development Environment

```bash
# Services starten
docker compose -f docker-compose.dev.yml up -d

# URLs:
# - Backend: http://localhost:8080
# - Admin UI: http://localhost:3005 (Vite Dev Server)
# - PostgreSQL: localhost:5433
# - Redis: localhost:6380
```

### Production Deployment

```bash
# SSH zu Server
ssh openweb

# Deployment ausführen
cd /opt/llm-proxy/deployments
./deploy.sh
```

---

## Aktueller Branch-Status (2026-02-06)

```
develop: c632585 (sync mit master)
master:  c632585 (Production)

Beide Branches sind synchron ✅
```

**Nächster Schritt:** Entwickle auf `develop`, committe bei jedem positiven Ergebnis.

---

## Hilfreiche Git Aliases (optional)

Füge zu `~/.gitconfig` hinzu:

```ini
[alias]
    st = status
    co = checkout
    br = branch -a
    lg = log --oneline --graph --all --decorate -20
    sync = !git checkout master && git merge develop --no-edit && git checkout develop
    pushall = !git push origin develop && git push origin master
```

Verwendung:
```bash
git st           # status
git co develop   # checkout develop
git lg           # schöner log
git sync         # merge develop nach master
git pushall      # push beide branches
```

---

**Erstellt:** 2026-02-06  
**Autor:** LLM-Proxy Team  
**Letztes Update:** 2026-02-06
