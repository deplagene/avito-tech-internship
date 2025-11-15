#!/bin/sh

set -e

echo "Waiting for database..."
while ! nc -z db 5432; do
  sleep 1
done
echo "Database is ready."

echo "Running migrations..."
migrate -path /migrations -database "$DATABASE_URL" up

exec "$@"
