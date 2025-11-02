#!/bin/bash

# Test script for authentication API

BASE_URL="http://localhost:8080"

echo "🧪 Testing Authentication API"
echo "================================"
echo ""

# Test 1: Health check
echo "1. Testing server health..."
curl -s "$BASE_URL/health" | jq . || echo "Failed"
echo ""
echo ""

# Test 2: Register new user
echo "2. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "testpassword123"
  }')

echo "Response:"
echo "$REGISTER_RESPONSE" | jq . || echo "$REGISTER_RESPONSE"
echo ""

# Extract token from response
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.token' 2>/dev/null)

if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
  echo "✅ Registration successful! Token: ${TOKEN:0:50}..."
  echo ""

  # Test 3: Get current user (protected endpoint)
  echo "3. Testing protected endpoint /api/auth/me..."
  ME_RESPONSE=$(curl -s -X GET "$BASE_URL/api/auth/me" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json")

  echo "Response:"
  echo "$ME_RESPONSE" | jq . || echo "$ME_RESPONSE"
  echo ""

  # Test 4: Login
  echo "4. Testing user login..."
  LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
      "email": "test@example.com",
      "password": "testpassword123"
    }')

  echo "Response:"
  echo "$LOGIN_RESPONSE" | jq . || echo "$LOGIN_RESPONSE"
  echo ""
else
  echo "❌ Registration failed!"
  echo "$REGISTER_RESPONSE"
fi

echo ""
echo "================================"
echo "✅ Tests completed!"

