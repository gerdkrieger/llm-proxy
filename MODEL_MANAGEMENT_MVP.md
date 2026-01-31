# 🎯 Model Management - MVP Implementierung

## ✅ Was bereits gemacht wurde:

1. **Database Migration** (`000005_add_provider_models`)
   - Tabelle: `provider_models` 
   - Felder: provider_id, model_id, model_name, enabled
   - ✅ Bereit für Migration

2. **Repository** (`provider_model_repository.go`)
   - CRUD Operations für Model-Settings
   - IsModelEnabled() Funktion
   - ✅ Implementiert

## 📝 Was noch zu tun ist (Schnell-Implementierung):

### Backend (15-20 Min):

```go
// 1. Neue Endpoints in provider_management_handler.go:

GET  /admin/providers/{id}/models
→ Gibt alle bekannten Models zurück + enabled Status aus DB

POST /admin/providers/{id}/models/configure
→ Body: {"enabled_models": ["model-id-1", "model-id-2"]}
→ Updated DB für diese Models

// 2. Hardcoded Model-Listen:
var CLAUDE_MODELS = []string{
    "claude-3-5-sonnet-20241022",
    "claude-3-5-haiku-20241022", 
    "claude-3-opus-20240229",
    "claude-3-sonnet-20240229",
    "claude-3-haiku-20240307",
    // ... alle anderen
}

var OPENAI_MODELS = []string{
    "gpt-4-turbo",
    "gpt-4",
    "gpt-4-32k",
    "gpt-3.5-turbo",
    // ... alle anderen
}
```

### Frontend (10-15 Min):

```svelte
<!-- In Providers.svelte: -->

<button on:click={() => openModelManagement(provider)}>
  Manage Models ({enabledCount}/{totalCount})
</button>

<!-- Model Management Modal -->
<div class="model-list">
  {#each models as model}
    <label>
      <input type="checkbox" bind:checked={model.enabled} />
      {model.name}
    </label>
  {/each}
</div>

<button on:click={saveModelConfiguration}>
  Save Configuration
</button>
```

### Integration in Chat Handler (5 Min):

```go
// In chat_handler.go vor dem Request:
isEnabled, err := h.providerModelRepo.IsModelEnabled(ctx, provider, modelID)
if !isEnabled {
    return error("Model not available")
}
```

## 🚀 Soll ich das jetzt implementieren?

**Option A: Vollständige Implementierung (30-40 Min)**
- Alle oben genannten Schritte
- Getestet und funktionsfähig
- Du kannst es sofort nutzen

**Option B: Nur Backend (15 Min)**
- Endpoints + DB-Integration
- Du kannst via API testen
- Frontend machst du später selbst

**Option C: Dokumentation Only**
- Ich gebe dir Code-Beispiele
- Du implementierst es selbst
- Ich helfe bei Fragen

Was möchtest du? 🎯
