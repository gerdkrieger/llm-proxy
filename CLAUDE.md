# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LLM-Proxy is an enterprise Go proxy server that sits between clients and LLM providers (Claude, OpenAI, Abacus.ai). It provides a unified OpenAI-compatible API (`/v1/chat/completions`, `/v1/models`) while adding content filtering, OAuth 2.0 authentication, weighted load balancing across provider API keys, request logging, and caching.

## Common Commands

### Development
```bash
make dev-docker        # Start full Docker dev stack (Backend+Air, Admin-UI+Vite, Postgres, Redis)
make dev-docker-up     # Same but detached
make dev-docker-down   # Stop dev environment
make dev               # Run Go server natively (requires local Postgres/Redis)
```

### Building
```bash
make build             # Production binary → bin/llm-proxy
go build ./cmd/server  # Quick build check
```

### Testing
```bash
make test                          # All tests with race detection + coverage
make test-unit                     # Unit tests only (go test -short)
make test-integration              # Integration tests (requires running services)
go test -v -race ./internal/...    # Test a specific package subtree
go test -v -run TestName ./path/   # Run a single test
```

### Code Quality
```bash
make lint              # golangci-lint (5min timeout)
make vet               # go vet
make fmt               # gofmt + goimports
make check             # lint + vet combined
```

### Database Migrations
```bash
make migrate-up        # Apply all pending migrations
make migrate-down      # Rollback last migration
# Migrations live in migrations/ and use golang-migrate
# DB connection string is derived from .env (DB_USER, DB_PASSWORD, DB_PORT, DB_NAME)
```

### Docker Services (infrastructure only)
```bash
make docker-up         # Start Postgres, Redis, Prometheus, Grafana
make docker-down       # Stop infrastructure services
```

## Architecture

The project follows **Clean Architecture** with four layers:

```
cmd/server/main.go                  ← Entry point: wires everything together
internal/
  interfaces/                       ← HTTP layer (chi router, handlers, middleware)
    api/router.go                   ← All route definitions
    api/*_handler.go                ← One handler per resource
    middleware/                     ← Auth (OAuth + API key), logging, metrics, CORS
  application/                      ← Business logic services
    filtering/service.go            ← Content filter engine (word/phrase/regex, priority-based)
    caching/service.go              ← Redis caching layer
    oauth/service.go                ← OAuth 2.0 client credentials + JWT
    health/checker.go               ← Provider health checks (5min interval)
    providers/sync.go               ← Model catalog sync
  domain/models/                    ← Data structures (Claude, OpenAI, filter models)
  infrastructure/                   ← External integrations
    database/repositories/          ← PostgreSQL repositories (pgx)
    providers/manager.go            ← Multi-provider load balancer with per-key rate limiting
    providers/claude/               ← Claude API client
    providers/openai/               ← OpenAI API client
    providers/abacus/               ← Abacus.ai API client
    cache/                          ← Redis client
pkg/                                ← Shared utilities (logger, metrics, crypto, SSE, errors)
```

### Key Architectural Concepts

**Provider Manager** (`infrastructure/providers/manager.go`): Central component that routes requests to LLM providers. Supports multiple API keys per provider with weighted load balancing and automatic failover on rate limits.

**Content Filtering** (`application/filtering/service.go`): Filters run on request content before forwarding to providers. Three types: word, phrase, regex. Uses in-memory compiled regex cache with 5-min TTL. Matches are tracked in the database.

**Authentication**: Two parallel auth paths — OAuth 2.0 (client credentials flow with JWT) and API key auth (OpenAI-compatible Bearer tokens). Admin endpoints use a separate `X-Admin-API-Key` header.

**Streaming**: SSE-based streaming for chat completions, handled via `pkg/sse/`.

## Configuration

- Config loaded via Viper from `configs/config.yaml` or environment variables
- Single `.env.example` as template → copy to `.env.local` (dev) or `.env` (production, server-only)
- CORS origins configured via `server.cors_origins` (config) or `SERVER_CORS_ORIGINS` env var
- `ENVIRONMENT=development|production` controls log warnings and defaults
- Encryption key required for storing provider API keys in DB (AES-256-GCM, 32-byte hex)
- Dev ports: API=8080, Admin-UI=3005, Postgres=5433, Redis=6380

## Tech Stack

- **Go 1.25**, chi/v5 router, pgx/v5 (PostgreSQL), go-redis/v9, zerolog, golang-jwt, Prometheus metrics
- **Frontend** (admin-ui/): Svelte + Vite + Tailwind CSS
- **Infrastructure**: PostgreSQL 14, Redis 7, Prometheus, Grafana
- **Dev tooling**: Air (hot-reload), golangci-lint, golang-migrate

## Deployment

Deployment uses `scripts/deployment/deploy.sh` (SSH image transfer). Workflow: build Docker images locally → `docker save | ssh docker load` → restart on server. Only `docker-compose.prod.yml`, `.env`, and monitoring configs exist on the server — no source code.
