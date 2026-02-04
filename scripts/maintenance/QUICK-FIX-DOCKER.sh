#!/bin/bash
# Quick Fix für Docker Container-Konflikt
# Datum: 4. Februar 2026

set -euo pipefail

echo "🔧 LLM-Proxy Docker Quick Fix"
echo "=============================="
echo ""

# SSH Host prüfen
if ! ssh -q -o BatchMode=yes -o ConnectTimeout=5 openweb exit 2>/dev/null; then
    echo "❌ Fehler: Kann nicht auf 'openweb' per SSH zugreifen"
    echo "Bitte SSH-Verbindung prüfen oder direkt auf dem Server ausführen"
    exit 1
fi

echo "✅ SSH-Verbindung zu openweb OK"
echo ""

# Auf Server ausführen
echo "🧹 Entferne blockierende Container..."
ssh openweb << 'EOF'
cd ~
echo "Stoppe und entferne LLM-Proxy Container..."
docker stop llm-proxy-admin-ui llm-proxy-backend llm-proxy-postgres llm-proxy-redis 2>/dev/null || true
docker rm -f llm-proxy-admin-ui llm-proxy-backend llm-proxy-postgres llm-proxy-redis 2>/dev/null || true

echo "Entferne Network (falls blockiert)..."
docker network rm llm-proxy-network 2>/dev/null || true

echo "Warte kurz..."
sleep 2

echo "Status prüfen..."
docker ps -a | grep llm-proxy || echo "✅ Alle LLM-Proxy Container entfernt"
EOF

echo ""
echo "✅ Cleanup abgeschlossen!"
echo ""
echo "Nächster Schritt:"
echo "  1. GitLab CI/CD Pipeline neu starten (Retry-Button)"
echo "  2. ODER manuell deployen:"
echo "     ssh openweb 'cd ~/llm-proxy && docker compose -f docker-compose.openwebui.yml up -d --build'"
echo ""
echo "Status prüfen nach Deployment:"
echo "  curl https://llmproxy.aitrail.ch/health"
echo ""
