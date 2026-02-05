#!/bin/sh
set -e

echo "🔄 Starting LLM-Proxy Development Environment..."

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL..."
until pg_isready -h postgres -p 5432 -U proxy_user > /dev/null 2>&1; do
  echo "   PostgreSQL is unavailable - sleeping"
  sleep 2
done
echo "✅ PostgreSQL is ready!"

# Run database migrations
echo "🔄 Running database migrations..."
if [ -d "/app/migrations" ]; then
  DATABASE_URL="postgres://proxy_user:dev_password_2024_local@postgres:5432/llm_proxy?sslmode=disable"
  
  # Check if migrate tool is available
  if command -v migrate > /dev/null 2>&1; then
    migrate -path /app/migrations -database "$DATABASE_URL" up
    echo "✅ Migrations completed successfully"
  else
    echo "⚠️  Warning: migrate tool not found, skipping migrations"
  fi
else
  echo "⚠️  Warning: migrations directory not found at /app/migrations"
fi

# Start Air for hot-reload
echo "🚀 Starting Air (hot-reload)..."
exec air -c .air.toml
