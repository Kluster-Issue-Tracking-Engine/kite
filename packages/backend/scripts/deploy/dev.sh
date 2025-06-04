#!/bin/bash

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

# Run the seeder
echo "Running database seeder..."
# Run seeder, ignore local vendor directory
go run -mod=mod cmd/seed/main.go

# Start the main application with auto-reload
echo "Starting server..."
#exec ./server
air -c .air.toml
