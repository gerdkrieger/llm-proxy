#!/bin/sh

# Start Go form backend in background
echo "Starting form backend..."
/usr/local/bin/form-backend &

# Wait a moment for backend to start
sleep 2

# Start nginx in foreground
echo "Starting nginx..."
exec nginx -g "daemon off;"
