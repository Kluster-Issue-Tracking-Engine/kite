#!/bin/sh
set -e

# Validate required environment variables
: "${KITE_DB_HOST:?Environment variable KITE_DB_HOST is required}"
: "${KITE_DB_PORT:?Environment variable KITE_DB_PORT is required}"
: "${KITE_DB_USER:?Environment variable KITE_DB_USER is required}"
: "${KITE_DB_NAME:?Environment variable KITE_DB_NAME is required}"

# Start the main application
echo "Starting server on port ${KITE_PORT:-8080}..."
exec ./server
