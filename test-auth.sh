#!/bin/bash

# Test script for authentication endpoints

BASE_URL="http://localhost:8080"

echo "Testing E-commerce Authentication API"
echo "====================================="

# Test health endpoint
echo -e "\n1. Testing health endpoint..."
curl -s -X GET "$BASE_URL/health" | jq '.'

# Test user registration
echo -e "\n2. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }')

echo "$REGISTER_RESPONSE" | jq '.'

# Extract token from registration response
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
    echo -e "\n3. Testing protected profile endpoint..."
    curl -s -X GET "$BASE_URL/api/profile" \
      -H "Authorization: Bearer $TOKEN" | jq '.'
else
    echo -e "\n3. Skipping profile test - no token received"
fi

# Test user login
echo -e "\n4. Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }')

echo "$LOGIN_RESPONSE" | jq '.'

# Extract token from login response
LOGIN_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')

if [ "$LOGIN_TOKEN" != "null" ] && [ "$LOGIN_TOKEN" != "" ]; then
    echo -e "\n5. Testing protected profile endpoint with login token..."
    curl -s -X GET "$BASE_URL/api/profile" \
      -H "Authorization: Bearer $LOGIN_TOKEN" | jq '.'
else
    echo -e "\n5. Skipping profile test - no login token received"
fi

# Test invalid login
echo -e "\n6. Testing invalid login..."
curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "wrongpassword"
  }' | jq '.'

echo -e "\n====================================="
echo "Authentication API testing completed!"
