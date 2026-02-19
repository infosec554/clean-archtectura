#!/usr/bin/env sh
set -e

HOST="${DB_HOST:-career_db}"
PORT="${DB_PORT:-5432}"
USER="${DB_USER:-postgres}"
export PGPASSWORD="${DB_PASSWORD:-1234}"

ATTEMPTS=60    
SLEEP=2      

echo "⏳ Waiting for PostgreSQL ($HOST:$PORT) to be ready..."

until pg_isready -h "$HOST" -p "$PORT" -U "$USER"; do
  ATTEMPTS=$((ATTEMPTS - 1))
  if [ "$ATTEMPTS" -le 0 ]; then
    echo "❌ PostgreSQL is not ready after $((60*SLEEP)) seconds."
    exit 1
  fi
  sleep "$SLEEP"
done

echo "✅ PostgreSQL is ready!"
exec "$@"
