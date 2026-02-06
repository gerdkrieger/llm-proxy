# GitLab CI/CD Container Registry Authentication

## Problem

GitLab CI/CD Pipeline scheiterte beim Push von Docker Images zur GitLab Container Registry mit Fehler:
```
denied: requested access to the resource is denied
```

**Root Cause:** `CI_JOB_TOKEN` hat kein `write_registry` Scope und kann daher nicht zum Pushen von Images verwendet werden.

## Lösung: Deploy Token mit write_registry Scope

Da GitLab.com Free Tier keine **Project Access Tokens** unterstützt (nur Premium/Ultimate), wurde ein **Deploy Token** mit `write_registry` Scope erstellt.

### Erstellter Deploy Token

- **Name:** `ci-registry-deploy`
- **Username:** `gitlab+deploy-token`
- **Scopes:** `read_registry`, `write_registry`
- **Expires:** 2027-12-31
- **Token ID:** 11193856

### CI/CD Variablen

Die folgenden Variablen wurden im Projekt gespeichert:

| Variable | Value | Masked | Protected |
|----------|-------|--------|-----------|
| `REGISTRY_USER` | `gitlab+deploy-token` | No | No |
| `REGISTRY_TOKEN` | `gldt-***` | Yes | No |

### Änderungen in .gitlab-ci.yml

Der Docker Login wurde in allen Jobs angepasst um eine **Fallback-Hierarchie** zu verwenden:

```yaml
if [ -n "${REGISTRY_TOKEN:-}" ]; then
  # Primär: Deploy Token mit write_registry (NEU)
  docker login -u "${REGISTRY_USER}" --password-stdin "${CI_REGISTRY}"
elif [ -n "${CI_DEPLOY_PASSWORD:-}" ]; then
  # Fallback 1: Alter Deploy Token (falls vorhanden)
  docker login -u "${CI_DEPLOY_USER}" --password-stdin "${CI_REGISTRY}"
else
  # Fallback 2: CI_JOB_TOKEN (nur read access)
  docker login -u "${CI_REGISTRY_USER}" --password-stdin "${CI_REGISTRY}"
fi
```

**Betroffene Jobs:**
- `.docker-login` (Template)
- `docker:backend`
- `docker:admin-ui`
- `.deploy-template`

## Status

- ✅ Deploy Token erstellt (per API)
- ✅ CI/CD Variablen gespeichert
- ✅ `.gitlab-ci.yml` angepasst
- ⚠️ **Jobs bleiben `when: manual`** (nur bei Bedarf nutzen)
- ✅ **Primäre Deployment-Methode bleibt `deploy.sh`**

## Verwendung

### Option 1: deploy.sh (Empfohlen)

```bash
./deploy.sh
```

Baut Images lokal und transferiert sie per SSH. **Funktioniert zuverlässig**, keine Registry-Abhängigkeit.

### Option 2: GitLab CI/CD Pipeline (Backup)

1. Gehe zu: GitLab → CI/CD → Pipelines
2. Wähle Pipeline für `master` Branch
3. Klicke **manuell** auf:
   - `docker:backend` Job
   - `docker:admin-ui` Job
4. Images werden in GitLab Container Registry gepusht

**Vorteil:** Kollegen ohne Server-SSH-Zugang können deployen.

## Token Rotation

Deploy Token läuft ab am **2027-12-31**. Zur Erneuerung:

```bash
# 1. Alten Token revoken (optional)
curl --request DELETE \
  --header "PRIVATE-TOKEN: $BOOTSTRAP_TOKEN" \
  "https://gitlab.com/api/v4/projects/78135920/deploy_tokens/11193856"

# 2. Neuen Token erstellen
curl --request POST \
  --header "PRIVATE-TOKEN: $BOOTSTRAP_TOKEN" \
  --header "Content-Type: application/json" \
  --data '{
    "name": "ci-registry-deploy",
    "expires_at": "2029-12-31",
    "username": "gitlab+deploy-token",
    "scopes": ["read_registry", "write_registry"]
  }' \
  "https://gitlab.com/api/v4/projects/78135920/deploy_tokens"

# 3. CI Variable REGISTRY_TOKEN aktualisieren
curl --request PUT \
  --header "PRIVATE-TOKEN: $BOOTSTRAP_TOKEN" \
  --form "value=NEW_TOKEN_VALUE" \
  "https://gitlab.com/api/v4/projects/78135920/variables/REGISTRY_TOKEN"
```

## Referenzen

- [GitLab Deploy Tokens Documentation](https://docs.gitlab.com/ee/user/project/deploy_tokens/)
- [Container Registry Authentication](https://docs.gitlab.com/ee/user/packages/container_registry/authenticate_with_container_registry.html)
- [Project Access Tokens](https://docs.gitlab.com/ee/user/project/settings/project_access_tokens.html) (Premium/Ultimate only)

## Troubleshooting

### Push scheitert weiterhin mit "denied"

1. **Prüfe CI Variables:**
   ```bash
   # Gehe zu: Settings → CI/CD → Variables
   # Verifiziere: REGISTRY_USER und REGISTRY_TOKEN existieren
   ```

2. **Prüfe Deploy Token Status:**
   ```bash
   curl --header "PRIVATE-TOKEN: $TOKEN" \
     "https://gitlab.com/api/v4/projects/78135920/deploy_tokens"
   ```

3. **Test Login manuell:**
   ```bash
   echo "gldt-***" | docker login registry.gitlab.com -u gitlab+deploy-token --password-stdin
   docker pull registry.gitlab.com/krieger-engineering/llm-proxy/backend:latest
   docker tag registry.gitlab.com/krieger-engineering/llm-proxy/backend:latest \
     registry.gitlab.com/krieger-engineering/llm-proxy/backend:test
   docker push registry.gitlab.com/krieger-engineering/llm-proxy/backend:test
   ```

### Pipeline verwendet alten CI_JOB_TOKEN

Stelle sicher dass die Variablen **nicht** als `protected` markiert sind, wenn der Branch nicht protected ist.

Oder: Setze Branch als protected in GitLab → Settings → Repository → Protected Branches.

---

**Erstellt:** 2026-02-06  
**Autor:** LLM-Proxy Team  
**Letztes Update:** 2026-02-06
