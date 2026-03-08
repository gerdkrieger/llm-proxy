# 🛡️ LLM-Proxy

> **Enterprise AI Gateway with Content Filtering & Multi-Provider Support**

Protect your company from data leaks when using AI services like ChatGPT, Claude, and others. LLM-Proxy sits between your employees and AI providers, automatically filtering confidential information before it leaves your network.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

---

## 🌟 **Features**

### 🔒 **Content Filtering**
- Automatic detection and filtering of sensitive data (emails, phone numbers, IPs, etc.)
- Custom filter rules via regex patterns
- Configurable sensitivity levels

### 🌐 **Multi-Provider Support**
- **OpenAI** (GPT-4, GPT-3.5)
- **Anthropic** (Claude 3.5, Claude 3)
- **Azure OpenAI**
- Easy to extend for other providers

### 📊 **Request & Response Logging**
- Full audit trail of all AI interactions
- Filter what was removed from requests
- Cost tracking per user/department
- PostgreSQL storage for compliance

### 🎛️ **Admin Dashboard**
- Modern Svelte-based UI
- User management & API key generation
- Real-time usage statistics
- Filter rule configuration

### 🚀 **Self-Hosted & GDPR Compliant**
- Deploy on your own infrastructure
- Full data sovereignty
- No data leaves your network (except filtered requests to AI providers)
- Swiss-hosted option available

---

## 📸 **Screenshots**

### Landing Page
![Landing Page](docs/screenshots/landing-page.png)

### Admin Dashboard
![Admin Dashboard](docs/screenshots/admin-ui.png)

---

## 🚀 **Quick Start**

### Prerequisites
- Docker & Docker Compose
- (Optional) Domain with SSL certificate

### 1. Clone Repository
```bash
git clone https://github.com/krieger-engineering/llm-proxy.git
cd llm-proxy
```

### 2. Configuration
```bash
cp .env.example .env
# Edit .env with your settings:
# - Database credentials
# - AI provider API keys
# - Filter rules
```

### 3. Start Services
```bash
docker-compose up -d
```

### 4. Access Admin UI
Open http://localhost:3000 in your browser.

**Default credentials:**
- Username: `admin`
- Password: `changeme` (change immediately!)

---

## 🏗️ **Architecture**

```
┌─────────────┐
│   Employee  │
└──────┬──────┘
       │ Request with sensitive data
       ▼
┌─────────────────────────────────┐
│       LLM-Proxy Gateway         │
│  ┌───────────────────────────┐  │
│  │  Content Filter Engine    │  │ ← Removes PII, secrets, etc.
│  └───────────────────────────┘  │
│  ┌───────────────────────────┐  │
│  │   Request Logger          │  │ ← Audit trail (PostgreSQL)
│  └───────────────────────────┘  │
└──────┬──────────────────────────┘
       │ Filtered request
       ▼
┌─────────────────────────────────┐
│   AI Provider (OpenAI, Claude)  │
└─────────────────────────────────┘
```

---

## 🔧 **Configuration**

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://...` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `OPENAI_API_KEY` | Your OpenAI API key | - |
| `ANTHROPIC_API_KEY` | Your Anthropic API key | - |
| `FILTER_MODE` | Filter strictness (`strict`, `moderate`, `permissive`) | `moderate` |
| `LOG_LEVEL` | Logging level (`debug`, `info`, `warn`, `error`) | `info` |

### Custom Filter Rules

Edit `config/filters.yaml`:

```yaml
filters:
  - name: "Email Addresses"
    pattern: '[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}'
    replacement: "[EMAIL_REDACTED]"
    
  - name: "Credit Card Numbers"
    pattern: '\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b'
    replacement: "[CARD_REDACTED]"
```

---

## 🛠️ **Development**

### Project Structure
```
llm-proxy/
├── backend/              # Go backend (API gateway)
│   ├── cmd/             # Main applications
│   ├── internal/        # Internal packages
│   └── pkg/             # Public packages
├── admin-ui/            # Svelte admin dashboard
├── landing-page/        # Marketing website
├── migrations/          # Database migrations
└── docker-compose.yml   # Docker setup
```

### Local Development

**Backend:**
```bash
cd backend
go run cmd/server/main.go
```

**Admin UI:**
```bash
cd admin-ui
npm install
npm run dev
```

### Running Tests
```bash
# Backend tests
cd backend
go test ./...

# Frontend tests
cd admin-ui
npm test
```

---

## 📦 **Deployment**

### Docker Compose (Recommended)
```bash
docker-compose -f docker-compose.prod.yml up -d
```

### Kubernetes
See [docs/kubernetes/](docs/kubernetes/) for Helm charts and manifests.

### Reverse Proxy (Caddy)
```caddyfile
llmproxy.yourdomain.com {
    reverse_proxy localhost:8080
}
```

---

## 🔐 **Security**

- All API requests require authentication (API keys or JWT)
- Rate limiting to prevent abuse
- TLS/SSL encryption enforced
- Content filtering prevents data leaks
- Audit logs for compliance (GDPR Art. 30)

**Security Audit:** Last reviewed 2026-03-08

---

## 📊 **Use Cases**

### 🏢 **SMBs (Small & Medium Businesses)**
Employees use ChatGPT for customer emails → LLM-Proxy filters customer data automatically

### 🏥 **Healthcare**
Doctors use AI for medical documentation → HIPAA-compliant filtering of patient names, SSNs

### ⚖️ **Law Firms**
Lawyers research with AI → Client names and case details removed before sending

### 💻 **Software Companies**
Developers use AI for code review → API keys, database credentials filtered out

---

## 🤝 **Contributing**

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Setup
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 **License**

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

---

## 📞 **Contact & Support**

- **Website:** [https://landing.llmproxy.aitrail.ch/](https://landing.llmproxy.aitrail.ch/)
- **Demo Meeting:** [Book a 30-min demo](https://cal.com/gerd-krieger-plwn35/llm-proxy)
- **Email:** service@aitrail.ch
- **Issues:** [GitHub Issues](https://github.com/krieger-engineering/llm-proxy/issues)

---

## 🗺️ **Roadmap**

- [ ] Support for Google Gemini
- [ ] Browser extension for easy integration
- [ ] Slack/Teams bot integration
- [ ] Advanced analytics dashboard
- [ ] Multi-tenancy support
- [ ] On-premise deployment assistant

---

## ⭐ **Star History**

If you find this project useful, please consider giving it a star! ⭐

---

## 🙏 **Acknowledgments**

- Built with [Go](https://go.dev/), [Svelte](https://svelte.dev/), [PostgreSQL](https://www.postgresql.org/)
- Inspired by the need for GDPR-compliant AI usage in European companies
- Thanks to all contributors!

---

**Made with ❤️ in Zürich, Switzerland 🇨🇭**
