#!/bin/sh
# entrypoint.sh

# Exit immediately if a command exits with a non-zero status.
set -e

# Wait for the database to be ready
echo "Waiting for database..."
while ! nc -z db 5432; do
  sleep 1
done
echo "Database is ready."

# Run migrations
echo "Running migrations..."
migrate -path /migrations -database "$DATABASE_URL" up

# Execute the main command (passed from Dockerfile's CMD)
exec "$@"
