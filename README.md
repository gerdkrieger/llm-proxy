# LLM-Proxy

Intelligent multi-provider LLM proxy with content filtering, caching, and OAuth2 authentication.

## 🚀 Quick Start

```bash
# Start all services
./scripts/start-all.sh

# Check status
./scripts/status.sh

# Run tests
./scripts/testing/test_api.sh
```

## 💻 Local Development

### Docker-Based Development (Recommended)

Start the complete development environment with hot-reload using Docker:

```bash
# First time setup
cp .env.example .env.local
# Edit .env.local with your settings (Claude API key, etc.)

# Start all services with hot-reload
make dev-docker

# Or run detached
make dev-docker-up

# View logs
make dev-docker-logs

# Stop services
make dev-docker-down
```

**Services:**
- Backend API: http://localhost:8080 (with Air hot-reload)
- Admin UI: http://localhost:3005 (with Vite hot-reload)
- PostgreSQL: localhost:5433
- Redis: localhost:6380
- Metrics: http://localhost:9090/metrics

**Hot-Reload:**
- Backend: Air watches Go files and automatically rebuilds
- Admin-UI: Vite dev server with instant HMR (Hot Module Replacement)

### Native Development

Run services natively without Docker (requires local PostgreSQL and Redis):

```bash
# Start infrastructure only
make docker-up

# Run backend natively
make dev

# Run admin-ui natively
cd admin-ui && npm run dev
```

**Environment Files:**
- `.env.local` - Local Docker development (Docker service names)
- `.env` - Native development (localhost connections)
- `.env.example` - Template for new setup

### Development Commands

```bash
make help              # Show all available commands
make dev-docker        # Start Docker development (recommended)
make dev               # Start native Go development
make dev-docker-up     # Start Docker dev (detached)
make dev-docker-down   # Stop Docker development
make dev-docker-logs   # View development logs
make dev-docker-clean  # Remove all dev containers and data
make test              # Run tests
make lint              # Run linter
make fmt               # Format code
```

## 📁 Project Structure

```
llm-proxy/
├── admin-ui/              # Admin UI (Svelte)
├── cmd/                   # Main applications
├── internal/              # Internal packages
│   ├── application/       # Business logic
│   ├── domain/            # Domain models
│   ├── infrastructure/    # External services
│   └── interfaces/        # HTTP handlers, middleware
├── pkg/                   # Shared libraries
├── api/                   # API definitions
├── configs/               # Configuration files
│   ├── Caddyfile.example  # Caddy reverse proxy config
│   └── example-filters.csv # Example content filters
├── deployments/           # Deployment configurations
│   └── docker-compose.openwebui.yml
├── migrations/            # Database migrations
│   ├── 001_add_hash_and_stats_columns.sql
│   ├── README.md
│   └── DEPLOYMENT_CHECKLIST.md
├── scripts/               # Utility scripts
│   ├── setup/             # Setup scripts
│   ├── maintenance/       # Maintenance scripts
│   └── testing/           # Test scripts
├── docs/                  # Documentation
│   ├── guides/            # User guides
│   ├── deployment/        # Deployment docs
│   └── sessions/          # Development sessions
├── filter-templates/      # Content filter templates
└── README.md              # This file
```

## 📚 Documentation

### Getting Started
- [Resume Project](docs/RESUME-PROJECT.md) - Start here when resuming work
- [Quick Start Filters](docs/guides/QUICK_START_FILTERS.md) - Get started with content filtering
- [Testing Guide](docs/TESTING.md) - How to run tests

### Guides
- [Admin API](docs/guides/ADMIN_API.md) - API reference
- [Content Filtering](docs/guides/CONTENT_FILTERING.md) - Filtering system
- [Filter Management](docs/guides/FILTER_MANAGEMENT_GUIDE.md) - Managing filters
- [Open WebUI Integration](docs/guides/OPENWEBUI_INTEGRATION_GUIDE.md) - Integration guide
- [Model Management](docs/guides/MODEL_MANAGEMENT_MVP.md) - Managing models
- [Anthropic Credits](docs/guides/ANTHROPIC_CREDITS_GUIDE.md) - Managing API credits
- [Bulk Import](docs/guides/BULK_IMPORT_GUIDE.md) - Bulk importing filters

### Deployment
- [Deployment Guide](docs/deployment/DEPLOYMENT.md) - How to deploy
- [CI/CD Pipeline](docs/deployment/CICD.md) - CI/CD configuration
- [Git Workflow](docs/deployment/GIT_WORKFLOW.md) - Branching strategy
- [Migration Checklist](migrations/DEPLOYMENT_CHECKLIST.md) - Pre-deployment checklist

### Operations
- [Live Server Commands](docs/LIVE-SERVER-COMMANDS.md) - Common server commands
- [Maintenance](docs/MAINTENANCE.md) - Routine maintenance
- [Troubleshooting](docs/TROUBLESHOOTING.md) - Common issues

## 🛠️ Features

### Core Features
- ✅ Multi-provider LLM routing (OpenAI, Anthropic Claude)
- ✅ OAuth2 authentication with JWT tokens
- ✅ Response caching with Redis
- ✅ Content filtering (profanity, PII, custom patterns)
- ✅ Rate limiting per client
- ✅ Request/response logging
- ✅ Cost tracking and billing

### Admin Features
- ✅ Web-based Admin UI
- ✅ Client management (CRUD)
- ✅ Filter management with live preview
- ✅ Provider status monitoring
- ✅ Usage statistics and analytics
- ✅ Cache management

## 🏗️ Architecture

### Tech Stack
- **Backend**: Go 1.25+
- **Admin UI**: Svelte + Vite
- **Database**: PostgreSQL 14
- **Cache**: Redis 7
- **Reverse Proxy**: Caddy
- **Container**: Docker + Docker Compose

### Components
```
┌─────────────┐
│   Clients   │
└──────┬──────┘
       │
┌──────▼──────────────┐
│   Caddy (HTTPS)     │
└──────┬──────────────┘
       │
┌──────▼──────────────┐      ┌──────────────┐
│   LLM-Proxy         │─────▶│  PostgreSQL  │
│   (Backend)         │      └──────────────┘
└──────┬──────────────┘
       │                      ┌──────────────┐
       ├─────────────────────▶│    Redis     │
       │                      └──────────────┘
       │
┌──────▼──────────────┐      ┌──────────────┐
│   Admin UI          │      │   OpenAI     │
│   (Svelte)          │      │   Anthropic  │
└─────────────────────┘      └──────────────┘
```

## 🔧 Development

### Prerequisites
- Go 1.25+
- Node.js 22+
- Docker & Docker Compose
- PostgreSQL 14
- Redis 7

### Local Development

```bash
# Start development environment
./scripts/start-dev.sh

# Backend API: http://localhost:8080
# Admin UI: http://localhost:5173
# PostgreSQL: localhost:5432
# Redis: localhost:6379
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./internal/application/oauth

# With coverage
go test -cover ./...

# API tests
./scripts/testing/test_api.sh
./scripts/testing/test_admin_api.sh

# Filter tests
./scripts/testing/test-content-filters.sh
```

## 📦 Deployment

### Production Deployment

1. **Check for migrations**
```bash
ls migrations/
```

2. **Run database migrations**
```bash
# See migrations/DEPLOYMENT_CHECKLIST.md
```

3. **Deploy services**
```bash
cd deployments/
docker-compose -f docker-compose.openwebui.yml up -d
```

4. **Verify deployment**
```bash
curl https://llmproxy.aitrail.ch/health
```

See [Deployment Checklist](migrations/DEPLOYMENT_CHECKLIST.md) for complete steps.

## 🐛 Troubleshooting

### Common Issues

**Issue**: "Failed to fetch" in Admin UI
**Solution**: Hard refresh browser (`Ctrl + Shift + R`)

**Issue**: Database column errors
**Solution**: Run migrations from `migrations/` directory

**Issue**: Container won't start
**Solution**: Check logs with `docker logs llm-proxy-backend`

See [Troubleshooting Guide](docs/TROUBLESHOOTING.md) for more solutions.

## 📊 Monitoring

### Health Checks
```bash
# Backend health
curl https://llmproxy.aitrail.ch/health

# Provider status
curl -H "X-Admin-API-Key: YOUR_KEY" \
  https://llmproxy.aitrail.ch/admin/providers/status
```

### Logs
```bash
# Backend logs
docker logs llm-proxy-backend -f

# Admin UI logs
docker logs llm-proxy-admin-ui -f

# All services
docker-compose logs -f
```

## 🤝 Contributing

1. Create feature branch: `git checkout -b feature/my-feature`
2. Make changes and test
3. Run migrations if DB changes
4. Commit with clear message
5. Push and create pull request

See [Git Workflow](docs/deployment/GIT_WORKFLOW.md) for details.

## 📝 License

[Add license information here]

## 🔗 Links

- [Production](https://llmproxy.aitrail.ch)
- [GitLab Repository](https://gitlab.com/krieger-engineering/llm-proxy)
- [GitLab Container Registry](https://gitlab.com/krieger-engineering/llm-proxy/container_registry)

## 📞 Support

For issues and questions:
1. Check [Troubleshooting Guide](docs/TROUBLESHOOTING.md)
2. Check [Session Notes](docs/sessions/) for recent fixes
3. Check GitLab Issues

---

**Last Updated**: February 4, 2026  
**Version**: 1.0.0  
**Status**: Production Ready ✅
