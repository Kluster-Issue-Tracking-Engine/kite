#!/bin/bash

# Validate required environment variables
: "${DB_HOST:?Environment variable DB_HOST is required}"
: "${DB_PORT:?Environment variable DB_PORT is required}"
: "${DB_USER:?Environment variable DB_USER is required}"
: "${DB_NAME:?Environment variable DB_NAME is required}"

# Wait for database to be ready
echo "Waiting for database to be ready..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -q; do
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
