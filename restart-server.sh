#!/bin/bash
# =============================================================================
# RESTART SERVER SCRIPT
# =============================================================================
# Stoppt alte Instanzen und startet den Server neu
# =============================================================================

set -e

echo "🔄 Server Neustart..."
echo ""

# Stoppe alte Instanzen
./stop-server.sh

echo ""
echo "🚀 Starte Server..."
./start-dev.sh
