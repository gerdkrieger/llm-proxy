#!/bin/bash
# Rebuild and deploy Admin-UI with correct API fix
set -euo pipefail

echo "🏗️  Rebuilding Admin-UI locally..."
cd "$(dirname "$0")"

# Build new image with current code (includes 784cfaf fix)
docker build -t llm-proxy-admin-ui:fixed -f admin-ui/Dockerfile admin-ui/

echo "📦 Saving image..."
docker save llm-proxy-admin-ui:fixed | gzip > /tmp/admin-ui-fixed.tar.gz

echo "📤 Uploading to server..."
scp /tmp/admin-ui-fixed.tar.gz openweb:/tmp/

echo "🚀 Deploying on server..."
ssh openweb << 'EOF'
  set -e
  echo "Loading image..."
  gunzip < /tmp/admin-ui-fixed.tar.gz | docker load
  
  echo "Stopping old container..."
  docker stop llm-proxy-admin-ui
  docker rm llm-proxy-admin-ui
  
  echo "Starting new container..."
  docker run -d \
    --name llm-proxy-admin-ui \
    --restart unless-stopped \
    --network llm-proxy-network \
    -p 3005:80 \
    llm-proxy-admin-ui:fixed
  
  echo "Waiting for container to be healthy..."
  sleep 10
  
  echo "✅ Deployment complete!"
  docker ps | grep llm-proxy-admin-ui
  
  # Cleanup
  rm /tmp/admin-ui-fixed.tar.gz
EOF

echo "✅ Admin-UI mit API-Fix erfolgreich deployed!"
echo "🌐 Teste: https://llmproxy.aitrail.ch"
