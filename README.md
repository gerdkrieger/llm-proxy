# 🚀 LLM-Proxy - Enterprise-Grade LLM Gateway

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Ready-326CE5?style=flat&logo=kubernetes)](https://kubernetes.io/)

**LLM-Proxy** is an enterprise-grade API gateway for Large Language Model providers. It provides a unified, OpenAI-compatible interface with advanced features like load balancing, caching, rate limiting, cost tracking, and comprehensive analytics.

---

## 📋 Table of Contents

- [Features](#-features)
- [Architecture](#-architecture)
- [Quick Start](#-quick-start)
- [Installation](#-installation)
- [Configuration](#-configuration)
- [API Documentation](#-api-documentation)
- [Development](#-development)
- [Testing](#-testing)
- [Deployment](#-deployment)
- [Monitoring](#-monitoring)
- [Contributing](#-contributing)
- [License](#-license)

---

## ✨ Features

### Core Functionality
- ✅ **OpenAI-Compatible API** - Drop-in replacement for OpenAI SDK
- ✅ **Multi-Provider Support** - Currently supports Anthropic Claude (extensible architecture)
- ✅ **Load Balancing** - Distribute requests across multiple API keys
- ✅ **Intelligent Caching** - Redis-backed response caching
- ✅ **Streaming Support** - Server-Sent Events (SSE) for real-time responses

### Security & Access Control
- ✅ **OAuth 2.0** - Industry-standard authentication
- ✅ **Scope-Based Authorization** - Read, write, and admin scopes
- ✅ **Rate Limiting** - Configurable per-client limits (RPM/RPD)
- ✅ **Access Lists** - Whitelist/blacklist for clients, IPs, and models
- ✅ **API Key Encryption** - Secure storage of provider keys

### Analytics & Billing
- ✅ **Request Logging** - Comprehensive request/response tracking
- ✅ **Cost Tracking** - Real-time token usage and cost calculation
- ✅ **Usage Analytics** - Daily aggregated statistics
- ✅ **Credit System** - Prepaid billing with manual top-up
- ✅ **Export Functionality** - CSV/JSON export for analytics

### Operations & Monitoring
- ✅ **Prometheus Metrics** - Detailed performance metrics
- ✅ **Grafana Dashboards** - Pre-configured monitoring dashboards
- ✅ **Structured Logging** - JSON logs with request ID tracking
- ✅ **Health Checks** - Liveness and readiness endpoints
- ✅ **Graceful Shutdown** - Zero-downtime deployments

### Admin Interface
- ✅ **Svelte Admin UI** - Modern web interface (separate project)
- ✅ **Client Management** - CRUD operations for OAuth clients
- ✅ **Provider Configuration** - Manage LLM provider API keys
- ✅ **Access Control** - Manage whitelists/blacklists
- ✅ **Billing Dashboard** - Credit management and top-up

---

## 🏗️ Architecture

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Clients   │────────▶│  LLM-Proxy   │────────▶│   Claude    │
│ (OpenAI SDK)│         │  (Gateway)   │         │     API     │
└─────────────┘         └──────────────┘         └─────────────┘
                               │
                               ├─────▶ PostgreSQL (Logging, Analytics)
                               ├─────▶ Redis (Caching, Rate Limiting)
                               └─────▶ Prometheus (Metrics)
```

### Technology Stack
- **Backend**: Go 1.21+
- **Database**: PostgreSQL 14+
- **Cache**: Redis 7+
- **Monitoring**: Prometheus + Grafana
- **HTTP Framework**: Chi Router
- **Authentication**: JWT (RS256)
- **Containerization**: Docker + Docker Compose

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- `golang-migrate` CLI tool
- Claude API key (from [Anthropic Console](https://console.anthropic.com/))

### 1. Clone the Repository
```bash
git clone https://github.com/your-org/llm-proxy.git
cd llm-proxy
```

### 2. Configure Environment
```bash
# Copy example env file
cp .env.example .env

# Edit .env and add your Claude API key
nano .env
```

### 3. Run Setup
```bash
# Install tools (optional, if not already installed)
make install-tools

# Run full setup (starts Docker, runs migrations)
make setup
```

### 4. Start Development Server
```bash
make dev
```

The proxy is now running on `http://localhost:8080`!

### 5. Verify Installation
```bash
# Check health
curl http://localhost:8080/health

# Check services
make health-check
```

---

## 📦 Installation

### Option 1: Using Make (Recommended)
```bash
# Full setup
make setup

# Start development
make dev
```

### Option 2: Manual Setup
```bash
# 1. Install dependencies
go mod download

# 2. Start Docker services
cd deployments/docker
docker-compose up -d
cd ../..

# 3. Wait for services
sleep 10

# 4. Run migrations
migrate -path migrations -database "postgres://proxy_user:dev_password_2024@localhost:5433/llm_proxy?sslmode=disable" up

# 5. Run application
go run cmd/server/main.go
```

### Option 3: Docker Only
```bash
# Build Docker image
make docker-build

# Run in Docker
docker run -d \
  --name llm-proxy \
  -p 8080:8080 \
  --env-file .env \
  llm-proxy:latest
```

---

## ⚙️ Configuration

### Environment Variables

Key environment variables (see `.env.example` for full list):

```bash
# Server
SERVER_PORT=8080

# Database
DB_HOST=localhost
DB_PORT=5433
DB_NAME=llm_proxy
DB_USER=proxy_user
DB_PASSWORD=your_secure_password

# Redis
REDIS_HOST=localhost
REDIS_PORT=6380

# OAuth
OAUTH_JWT_SECRET=your-32-char-secret-key
OAUTH_ACCESS_TOKEN_TTL=3600

# Claude API
CLAUDE_API_KEY=sk-ant-api03-your-key-here

# Admin
ADMIN_API_KEY=admin_your-secure-key
```

### Configuration File

Create `configs/config.yaml`:

```yaml
server:
  port: 8080
  timeout: 300s

cache:
  enabled: true
  ttl: 3600

rate_limiting:
  enabled: true
  default_rpm: 1000
```

---

## 📚 API Documentation

### Client API (OAuth 2.0 Protected)

#### Authentication
```bash
# Get OAuth token (client_credentials flow)
curl -X POST http://localhost:8080/oauth/token \
  -H "Content-Type: application/json" \
  -d '{
    "grant_type": "client_credentials",
    "client_id": "your_client_id",
    "client_secret": "your_client_secret"
  }'
```

#### Chat Completion
```bash
# OpenAI-compatible endpoint
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "max_tokens": 1024
  }'
```

#### List Models
```bash
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### Admin API (API Key Protected)

#### Dashboard Stats
```bash
curl http://localhost:8080/admin/dashboard \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY"
```

#### Create OAuth Client
```bash
curl -X POST http://localhost:8080/admin/clients \
  -H "X-Admin-API-Key: YOUR_ADMIN_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My App",
    "rate_limit_rpm": 1000
  }'
```

---

## 🛠️ Development

### Available Make Commands

```bash
make help              # Show all available commands
make setup             # Initial project setup
make dev               # Start development server
make build             # Build production binary
make test              # Run all tests
make test-coverage     # Generate coverage report
make lint              # Run linter
make fmt               # Format code
make docker-up         # Start Docker services
make docker-down       # Stop Docker services
make migrate-up        # Run database migrations
make migrate-down      # Rollback last migration
make clean             # Clean build artifacts
```

### Project Structure

```
llm-proxy/
├── cmd/server/           # Application entrypoint
├── internal/             # Private application code
│   ├── domain/          # Business logic & models
│   ├── application/     # Use cases
│   ├── infrastructure/  # External services (DB, Redis)
│   └── interfaces/      # HTTP handlers, middleware
├── pkg/                 # Public libraries
├── migrations/          # Database migrations
├── deployments/         # Docker configs
├── tests/               # Test files
└── docs/                # Documentation
```

### Adding a New Feature

1. Create feature branch: `git checkout -b feature/my-feature`
2. Implement in `internal/`
3. Add tests in `tests/`
4. Run checks: `make check test`
5. Commit with conventional commit message
6. Create pull request

---

## 🧪 Testing

### Run All Tests
```bash
make test
```

### Run Specific Test Suite
```bash
# Unit tests only
make test-unit

# Integration tests
make test-integration

# With coverage
make test-coverage
```

### Benchmarks
```bash
make bench
```

---

## 🚢 Deployment

### Docker Compose (Development/Staging)

```bash
# Start all services
cd deployments/docker
docker-compose up -d

# View logs
docker-compose logs -f llm-proxy

# Stop services
docker-compose down
```

### Kubernetes (Production)

```bash
# Apply manifests
kubectl apply -f deployments/kubernetes/base/

# Check status
kubectl get pods -l app=llm-proxy

# View logs
kubectl logs -f deployment/llm-proxy
```

### Environment-Specific Configs

- **Development**: Use `.env` file
- **Staging**: Use Kubernetes ConfigMaps
- **Production**: Use Kubernetes Secrets + Vault

---

## 📊 Monitoring

### Prometheus Metrics

Access metrics at: `http://localhost:9090`

Key metrics:
- `llm_proxy_requests_total` - Total requests
- `llm_proxy_request_duration_seconds` - Request latency
- `llm_proxy_tokens_total` - Token usage
- `llm_proxy_cost_usd_total` - Total costs
- `llm_proxy_cache_hits_total` - Cache performance

### Grafana Dashboards

Access Grafana at: `http://localhost:3001` (admin/admin)

Pre-configured dashboards:
- Overview Dashboard
- Provider Performance
- Billing & Costs
- Infrastructure Metrics

### Logs

```bash
# View application logs
make logs

# Docker logs
docker logs -f llm-proxy-app
```

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Submit a pull request

### Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting (run `make fmt`)
- Add tests for new features
- Update documentation

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/)
- Powered by [Anthropic Claude](https://www.anthropic.com/)
- Monitored by [Prometheus](https://prometheus.io/) & [Grafana](https://grafana.com/)

---

## 📞 Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/llm-proxy/issues)
- **Email**: support@yourcompany.com

---

**Made with ❤️ for Enterprise AI Applications**
