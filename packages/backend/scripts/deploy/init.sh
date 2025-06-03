#!/bin/sh

# Validate required environment variables
: "${KITE_DB_HOST:?Environment variable KITE_DB_HOST is required}"
: "${KITE_DB_PORT:?Environment variable KITE_DB_PORT is required}"
: "${KITE_DB_USER:?Environment variable KITE_DB_USER is required}"
: "${KITE_DB_NAME:?Environment variable KITE_DB_NAME is required}"

# Wait for database to be ready
echo "Waiting for database to be ready..."
until pg_isready -h "$KITE_DB_HOST" -p "$KITE_DB_PORT" -U "$KITE_DB_USER" -d "$KITE_DB_NAME" -q; do
  echo "Database is not ready yet. Waiting 2 seconds..."
  sleep 2
done
echo "Database is ready!"

# Run Atlas migrations with explicit database URL
echo "Running Atlas migrations..."
KITE_DB_SSL_MODE="require"
if [ "$KITE_PROJECT_ENV" = "development" ]; then
  KITE_DB_SSL_MODE="disable"
fi

atlas migrate apply \
  --dir "file://migrations" \
  --url "postgres://$KITE_DB_USER:$KITE_DB_PASSWORD@$KITE_DB_HOST:$KITE_DB_PORT/$KITE_DB_NAME?sslmode=$KITE_DB_SSL_MODE"

# Check if migrations succeeded
# Get exit status of last command
if [ $? -eq 0 ]; then
  echo "Atlas migrations completed successfully!"
else
  echo "Atlas migrations failed!"
  exit 1
fi
