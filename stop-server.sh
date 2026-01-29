#!/bin/bash
# =============================================================================
# STOP SERVER SCRIPT
# =============================================================================
# Stoppt alle laufenden LLM-Proxy Server Instanzen
# =============================================================================

echo "🛑 Stoppe LLM-Proxy Server..."

# Finde alle Go-Prozesse die den Server laufen
PIDS=$(pgrep -f "go run cmd/server/main.go" || true)
MAIN_PIDS=$(pgrep -f "llm-proxy.*main" || true)

if [ -z "$PIDS" ] && [ -z "$MAIN_PIDS" ]; then
    echo "✅ Kein Server läuft"
    exit 0
fi

# Stoppe Go-Prozesse
if [ -n "$PIDS" ]; then
    echo "   Stoppe Go-Prozess(e): $PIDS"
    echo "$PIDS" | xargs kill 2>/dev/null || true
    sleep 1
fi

# Stoppe Main-Prozesse
if [ -n "$MAIN_PIDS" ]; then
    echo "   Stoppe Server-Prozess(e): $MAIN_PIDS"
    echo "$MAIN_PIDS" | xargs kill 2>/dev/null || true
    sleep 1
fi

# Erzwinge Beendigung falls nötig
if pgrep -f "llm-proxy.*main" > /dev/null 2>&1; then
    echo "   ⚠️  Erzwinge Beendigung..."
    pkill -9 -f "llm-proxy.*main" 2>/dev/null || true
fi

sleep 1

# Prüfe ob Port frei ist
if lsof -i :8080 | grep -q "LISTEN"; then
    echo "   ❌ Port 8080 ist immer noch belegt"
    lsof -i :8080 | grep "LISTEN"
    exit 1
else
    echo "✅ Server gestoppt, Port 8080 ist frei"
fi
