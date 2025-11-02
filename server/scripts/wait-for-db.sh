#!/bin/bash

# Navigate to server directory
cd "$(dirname "$0")/.."

# Check if Docker is available
if ! command -v docker &> /dev/null; then
  echo "Warning: Docker is not installed or not in PATH"
  exit 0
fi

# Check if Docker daemon is running
if ! docker info &> /dev/null; then
  echo "Warning: Docker daemon is not running. Please start Docker Desktop."
  exit 0
fi

# Start PostgreSQL if not running
echo "Starting PostgreSQL..."
docker-compose up -d postgres || {
  echo "Warning: Failed to start PostgreSQL container"
  exit 0
}

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
  if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
    echo "PostgreSQL is ready!"
    exit 0
  fi
  attempt=$((attempt + 1))
  sleep 1
done

echo "Warning: PostgreSQL failed to start after ${max_attempts} seconds"
echo "Server will attempt to connect anyway..."
exit 0
