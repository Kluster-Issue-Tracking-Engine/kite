#!/bin/sh
set -e

# Validate required environment variables
: "${DB_HOST:?Environment variable DB_HOST is required}"
: "${DB_PORT:?Environment variable DB_PORT is required}"
: "${DB_USER:?Environment variable DB_USER is required}"
: "${DB_NAME:?Environment variable DB_NAME is required}"

# Start the main application
echo "Starting server on port ${PORT:-8080}..."
exec ./server
