#!/bin/bash
# =============================================================================
# SERVICE STATUS CHECKER
# =============================================================================

echo "=== LLM-Proxy Service Status ==="
echo ""

# Backend
echo -n "🔧 Backend (Port 8080):     "
if lsof -i :8080 2>/dev/null | grep -q "LISTEN"; then
    PID=$(lsof -ti :8080 2>/dev/null | head -1)
    echo "✅ Running (PID: $PID)"
else
    echo "❌ Not Running"
fi

# Frontend
echo -n "🎨 Frontend (Port 5173):    "
if lsof -i :5173 2>/dev/null | grep -q "LISTEN"; then
    echo "✅ Running"
else
    echo "❌ Not Running"
fi

# PostgreSQL
echo -n "🗄️  PostgreSQL (Port 5433): "
if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "llm-proxy-postgres"; then
    if docker ps --filter "name=llm-proxy-postgres" --filter "health=healthy" --format '{{.Names}}' | grep -q "llm-proxy-postgres"; then
        echo "✅ Healthy"
    else
        echo "⚠️  Starting..."
    fi
else
    echo "❌ Not Running"
fi

# Redis
echo -n "⚡ Redis (Port 6380):       "
if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "llm-proxy-redis"; then
    if docker ps --filter "name=llm-proxy-redis" --filter "health=healthy" --format '{{.Names}}' | grep -q "llm-proxy-redis"; then
        echo "✅ Healthy"
    else
        echo "⚠️  Starting..."
    fi
else
    echo "❌ Not Running"
fi

# Prometheus
echo -n "📊 Prometheus (Port 9090):  "
if docker ps --format '{{.Names}}' 2>/dev/null | grep -q "llm-proxy-prometheus"; then
    echo "✅ Running"
else
    echo "❌ Not Running"
fi

echo ""
echo "=== Quick Links ==="
echo "Backend API:    http://localhost:8080"
echo "Admin UI:       http://localhost:5173"
echo "Health Check:   http://localhost:8080/health"
echo "Metrics:        http://localhost:8080/metrics"
echo "Prometheus:     http://localhost:9090"
echo ""
echo "=== Befehle ==="
echo "Backend starten:  ./start-dev.sh"
echo "Backend stoppen:  ./stop-server.sh"
echo "Backend restart:  ./restart-server.sh"
echo "Status prüfen:    ./status.sh"
