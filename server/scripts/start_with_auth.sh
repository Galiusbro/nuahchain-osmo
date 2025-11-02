#!/bin/bash

# Script to start server with authentication keys

cd "$(dirname "$0")/../.."

# Generate or use existing keys
if [ -z "$AUTH_MASTER_KEY" ]; then
  export AUTH_MASTER_KEY="$(openssl rand -base64 32)"
  echo "Generated AUTH_MASTER_KEY"
fi

if [ -z "$JWT_SECRET" ]; then
  export JWT_SECRET="$(openssl rand -base64 32)"
  echo "Generated JWT_SECRET"
fi

echo "🚀 Starting server with authentication..."
echo "AUTH_MASTER_KEY: ${AUTH_MASTER_KEY:0:20}..."
echo "JWT_SECRET: ${JWT_SECRET:0:20}..."
echo ""

# Start air with environment variables
exec air

