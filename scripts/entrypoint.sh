#!/bin/sh
set -e

echo "🚀 Running database migrations..."
/app/github.com/dev-au/CodeStream migrate up

echo "✅ Migrations complete. Starting application..."
exec /app/github.com/dev-au/CodeStream
