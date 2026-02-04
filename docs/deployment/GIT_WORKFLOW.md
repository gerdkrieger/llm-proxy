# 📋 Git Workflow Guide - LLM-Proxy

Dokumentation für das automatisierte Git-Update-Script.

---

## 🚀 Quick Start

```bash
./git-update.sh
```

Das Script führt dich interaktiv durch den kompletten Git-Workflow!

---

## 📦 Features

### ✅ Was das Script macht:

- **Automatische Prüfungen**: Git Repo, Branches, Remote
- **Flexible Workflows**: 3 Modi für verschiedene Szenarien
- **Pre-Commit Checks**: Optional Tests & Build ausführen
- **Release Management**: Automatische Version-Tags (Semantic Versioning)
- **Smart Merging**: Develop -> Master Workflow
- **Remote Push**: Automatisch oder manuell
- **Statistiken**: Commit-Details und Zusammenfassung

---

## 🎯 Die 3 Workflow-Modi

### 1️⃣ Simple Mode - Quick Commit & Push

**Wann verwenden?**
- Kleine Änderungen
- Bug fixes
- Dokumentation
- Schnelle Updates

**Was passiert?**
1. Zeigt Änderungen an
2. Erstellt Commit
3. Optional: Push zu Remote

**Beispiel:**
```bash
./git-update.sh
# Wähle: 1) Simple Mode
# Commit Message: "docs: Update README with filter examples"
```

---

### 2️⃣ Release Mode - Mit Release-Tag

**Wann verwenden?**
- Neue Features fertig
- Bugfix-Release
- Stabiler Code der released werden soll

**Was passiert?**
1. Optional: Build & Tests
2. Erstellt Commit
3. Erstellt Release-Tag (v1.2.3)
4. Push Commit + Tag zu Remote

**Beispiel:**
```bash
./git-update.sh
# Wähle: 2) Release Mode
# Build & Tests ausführen? Ja
# Commit Message: "feat: Add content filtering system"
# Release-Typ: 2) MINOR Release (0.X.0)
# Release Notes: "Content filtering with bulk import, admin UI integration"
```

---

### 3️⃣ Full Mode - Develop → Master + Release

**Wann verwenden?**
- Production Release
- Große Features
- Stabiler Merge zu Master

**Was passiert?**
1. Commit auf develop Branch
2. Merge develop → master
3. Erstellt Release-Tag
4. Push beide Branches + Tag

**Beispiel:**
```bash
./git-update.sh
# Wähle: 3) Full Mode
# Commit Message: "feat: Complete content filtering implementation"
# Release-Typ: 2) MINOR Release
# Release Notes: "v1.2.0 - Content Filtering System with Admin UI"
```

---

## 📝 Commit Message Conventions

Das Script verwendet **Conventional Commits** Style:

### Format:
```
<type>: <description>

[optional body]
[optional footer]
```

### Types:

| Type | Verwendung | Beispiel |
|------|------------|----------|
| `feat` | Neue Features | `feat: Add content filtering` |
| `fix` | Bug Fixes | `fix: Resolve API key validation` |
| `docs` | Dokumentation | `docs: Update README` |
| `refactor` | Code Refactoring | `refactor: Improve filter service` |
| `perf` | Performance | `perf: Optimize database queries` |
| `test` | Tests | `test: Add filter integration tests` |
| `chore` | Build/Config | `chore: Update dependencies` |
| `style` | Code Style | `style: Format code with gofmt` |

### Beispiele:
```bash
# Feature
feat: Add bulk import for content filters

# Bug Fix
fix: Resolve admin API key validation issue

# Documentation
docs: Add startup guide and testing report

# Refactoring
refactor: Extract filter validation logic

# Performance
perf: Add caching for filter lookups

# Multiple changes
feat: Add content filtering system

- Implement word, phrase, and regex filters
- Add bulk import API endpoint
- Create Svelte admin UI component
- Add comprehensive documentation
```

---

## 🏷️ Release Versioning (Semantic Versioning)

### Format: `vMAJOR.MINOR.PATCH`

```
v1.2.3
│ │ │
│ │ └─ PATCH: Bug fixes, kleine Änderungen
│ └─── MINOR: Neue Features (backwards compatible)
└───── MAJOR: Breaking Changes
```

### Wann welcher Typ?

#### **MAJOR (X.0.0)** - Breaking Changes
- API-Änderungen die nicht kompatibel sind
- Datenbank-Schema Änderungen
- Entfernung von Features

**Beispiel:**
```
v1.0.0 → v2.0.0
- Neue API Authentifizierung (bricht alte Clients)
```

#### **MINOR (0.X.0)** - Neue Features
- Neue Funktionalität
- Neue Endpoints
- Neue UI-Features

**Beispiel:**
```
v1.1.0 → v1.2.0
- Content Filtering System hinzugefügt
- Bulk Import Feature
```

#### **PATCH (0.0.X)** - Bug Fixes
- Bug Fixes
- Kleine Verbesserungen
- Dokumentation

**Beispiel:**
```
v1.2.0 → v1.2.1
- API Key Validierung gefixt
- Typo in Dokumentation korrigiert
```

---

## 🔄 Branch-Strategie

### Mit Develop Branch (Full Mode):

```
develop ──┬──> feature branches
          │
          │   (commits & testing)
          │
          └──> master (merge & release tag)
                 │
                 └──> v1.2.0
```

**Workflow:**
1. Arbeite auf `develop` Branch
2. Commit Changes
3. Teste alles
4. Merge zu `master` (via Script)
5. Erstelle Release-Tag
6. Push beide Branches

### Ohne Develop Branch (Simple/Release Mode):

```
master ──> (commits) ──> v1.2.0
```

**Workflow:**
1. Arbeite direkt auf `master`
2. Commit Changes
3. Optional: Release-Tag
4. Push

---

## 🛠️ Pre-Commit Checks

Das Script bietet optionale Checks vor dem Commit:

### 1. Build Check
```bash
make build
```
- Prüft ob Code kompiliert
- Findet Syntax-Fehler
- Empfohlen vor jedem Release

### 2. Test Check
```bash
go test ./...
```
- Führt alle Tests aus
- Optional, da Tests fehlschlagen können
- Empfohlen für Release Mode

### Wann überspringen?
- Dokumentation-Änderungen
- README Updates
- Kleine Fixes die keine Code-Änderungen betreffen

---

## 📤 Remote Repository Setup

### Remote hinzufügen:

#### GitHub:
```bash
git remote add origin https://github.com/username/llm-proxy.git
git push -u origin master
```

#### GitLab:
```bash
git remote add origin https://gitlab.com/username/llm-proxy.git
git push -u origin master
```

#### Gitea / Self-Hosted:
```bash
git remote add origin https://git.example.com/username/llm-proxy.git
git push -u origin master
```

### Remote prüfen:
```bash
git remote -v
```

### SSH statt HTTPS:
```bash
git remote set-url origin git@github.com:username/llm-proxy.git
```

---

## 📊 Beispiel-Workflows

### Workflow 1: Feature Development

```bash
# 1. Feature entwickeln
vim internal/application/filtering/service.go
vim admin-ui/src/components/Filters.svelte

# 2. Testen
make build
make test

# 3. Commit mit Script
./git-update.sh
# Wähle: 2) Release Mode
# Commit: "feat: Add content filtering admin UI"
# Release: 2) MINOR (v1.2.0)
```

### Workflow 2: Bug Fix

```bash
# 1. Bug fixen
vim internal/interfaces/api/content_filter_handler.go

# 2. Quick Commit
./git-update.sh
# Wähle: 1) Simple Mode
# Commit: "fix: Resolve filter validation issue"
```

### Workflow 3: Production Release

```bash
# 1. Alles auf develop testen
git checkout develop
# ... development work ...

# 2. Full Release zu master
./git-update.sh
# Wähle: 3) Full Mode
# Commit: "feat: Complete v1.2.0 release"
# Release: 2) MINOR (v1.2.0)
# Release Notes: "Content Filtering System - Production Ready"
```

### Workflow 4: Hotfix

```bash
# 1. Kritischen Bug fixen
vim internal/application/oauth/service.go

# 2. Quick Release
./git-update.sh
# Wähle: 2) Release Mode
# Commit: "fix: Critical security issue in OAuth"
# Release: 3) PATCH (v1.2.1)
```

---

## 🔍 Troubleshooting

### Problem: "Kein Git Repository gefunden"
```bash
cd /home/krieger/Sites/golang-projekte/llm-proxy
./git-update.sh
```

### Problem: "Merge Konflikt"
```bash
# Manuell lösen:
git status
git diff
# Konflikte editieren
git add .
git commit
# Script neu starten
```

### Problem: "Tests fehlgeschlagen"
```bash
# Option 1: Trotzdem fortfahren (im Script "j" wählen)
# Option 2: Tests fixen und neu versuchen
# Option 3: Tests überspringen
```

### Problem: "Push rejected"
```bash
# Remote ist ahead - erst pullen:
git pull origin master --rebase
git push origin master
```

---

## 📚 Nützliche Git Commands

### Log & History
```bash
# Letzte Commits
git log --oneline -10

# Mit Graph
git log --oneline --graph --all

# Commit Details
git show <commit-hash>

# Changes zwischen Tags
git log v1.1.0..v1.2.0 --oneline
```

### Tags
```bash
# Alle Tags
git tag -l

# Tag Details
git show v1.2.0

# Tag löschen (lokal)
git tag -d v1.2.0

# Tag löschen (remote)
git push origin :refs/tags/v1.2.0
```

### Branches
```bash
# Alle Branches
git branch -a

# Branch wechseln
git checkout develop

# Neuen Branch erstellen
git checkout -b feature/new-feature

# Branch löschen
git branch -d feature/old-feature
```

### Undo/Reset
```bash
# Letzten Commit rückgängig (behält Änderungen)
git reset --soft HEAD~1

# Änderungen verwerfen
git reset --hard HEAD

# Einzelne Datei zurücksetzen
git checkout -- file.go
```

---

## ⚙️ Script Anpassen

Das Script kann angepasst werden in:
```bash
vim git-update.sh
```

### Wichtige Variablen:
```bash
PROJECT_NAME="LLM-Proxy"
DEFAULT_BRANCH="master"
DEVELOP_BRANCH="develop"
```

### Features hinzufügen:
- Linting vor Commit
- Custom Tests
- Deployment-Trigger
- Slack/Discord Notifications
- CHANGELOG automatisch generieren

---

## 📖 Weitere Ressourcen

- [Conventional Commits](https://www.conventionalcommits.org/)
- [Semantic Versioning](https://semver.org/)
- [Git Flow](https://nvie.com/posts/a-successful-git-branching-model/)
- [GitHub Flow](https://guides.github.com/introduction/flow/)

---

## ✅ Checkliste: Vor großem Release

- [ ] Alle Tests laufen durch
- [ ] Build erfolgreich
- [ ] Dokumentation aktualisiert
- [ ] CHANGELOG.md aktualisiert
- [ ] Version in Code updated (falls relevant)
- [ ] Breaking Changes dokumentiert
- [ ] Migration Guide geschrieben (bei MAJOR)
- [ ] Backup erstellt
- [ ] Team informiert

---

## 🎯 Best Practices

1. **Kleine, häufige Commits** statt große, seltene
2. **Aussagekräftige Commit Messages** schreiben
3. **Tests vor Release** ausführen
4. **Semantic Versioning** konsequent nutzen
5. **Develop Branch** für aktive Entwicklung
6. **Master/Main** nur stabile Versionen
7. **Release Notes** immer schreiben
8. **Tags pushen** nicht vergessen
9. **Branches synchron** halten (pull/push)
10. **Backups** vor großen Änderungen

---

**Happy Coding! 🚀**
