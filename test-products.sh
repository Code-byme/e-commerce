#!/bin/bash

# Test script for product and category endpoints

BASE_URL="http://localhost:8080"

echo "Testing E-commerce Product & Category API"
echo "========================================="

# Test health endpoint
echo -e "\n1. Testing health endpoint..."
curl -s -X GET "$BASE_URL/health" | jq '.'

# Test user registration to get admin token
echo -e "\n2. Registering admin user..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "first_name": "Admin",
    "last_name": "User"
  }')

echo "$REGISTER_RESPONSE" | jq '.'

# Extract token from registration response
TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" != "null" ] && [ "$TOKEN" != "" ]; then
    echo -e "\n3. Creating categories..."
    
    # Create Electronics category
    ELECTRONICS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/categories" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{
        "name": "Electronics",
        "description": "Electronic devices and gadgets"
      }')
    
    echo "Electronics category:"
    echo "$ELECTRONICS_RESPONSE" | jq '.'
    
    # Extract Electronics category ID
    ELECTRONICS_ID=$(echo "$ELECTRONICS_RESPONSE" | jq -r '.data.id')
    
    # Create Clothing category
    CLOTHING_RESPONSE=$(curl -s -X POST "$BASE_URL/api/categories" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{
        "name": "Clothing",
        "description": "Fashion and apparel"
      }')
    
    echo "Clothing category:"
    echo "$CLOTHING_RESPONSE" | jq '.'
    
    # Extract Clothing category ID
    CLOTHING_ID=$(echo "$CLOTHING_RESPONSE" | jq -r '.data.id')
    
    echo -e "\n4. Creating products..."
    
    # Create iPhone product
    IPHONE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/products" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{
        \"name\": \"iPhone 15 Pro\",
        \"description\": \"Latest iPhone with advanced features\",
        \"price\": 999.99,
        \"stock\": 50,
        \"category_id\": $ELECTRONICS_ID,
        \"image_url\": \"https://example.com/iphone15.jpg\"
      }")
    
    echo "iPhone product:"
    echo "$IPHONE_RESPONSE" | jq '.'
    
    # Create T-shirt product
    TSHIRT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/products" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d "{
        \"name\": \"Cotton T-Shirt\",
        \"description\": \"Comfortable cotton t-shirt\",
        \"price\": 29.99,
        \"stock\": 100,
        \"category_id\": $CLOTHING_ID,
        \"image_url\": \"https://example.com/tshirt.jpg\"
      }")
    
    echo "T-shirt product:"
    echo "$TSHIRT_RESPONSE" | jq '.'
    
    echo -e "\n5. Testing public product endpoints..."
    
    # List all products
    echo "All products:"
    curl -s -X GET "$BASE_URL/products" | jq '.'
    
    # Get products by category
    echo -e "\nProducts in Electronics category:"
    curl -s -X GET "$BASE_URL/products/category/$ELECTRONICS_ID" | jq '.'
    
    # Search products
    echo -e "\nSearching for 'iPhone':"
    curl -s -X GET "$BASE_URL/products?search=iPhone" | jq '.'
    
    # Filter by price
    echo -e "\nProducts under $50:"
    curl -s -X GET "$BASE_URL/products?max_price=50" | jq '.'
    
    echo -e "\n6. Testing public category endpoints..."
    
    # List all categories
    echo "All categories:"
    curl -s -X GET "$BASE_URL/categories" | jq '.'
    
    # Get category with product count
    echo -e "\nElectronics category with product count:"
    curl -s -X GET "$BASE_URL/categories/$ELECTRONICS_ID/with-products" | jq '.'
    
    echo -e "\n7. Testing protected admin endpoints..."
    
    # Update product stock
    echo "Updating iPhone stock:"
    curl -s -X PATCH "$BASE_URL/api/products/1/stock" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{"quantity": -5}' | jq '.'
    
    # Update product
    echo -e "\nUpdating iPhone price:"
    curl -s -X PUT "$BASE_URL/api/products/1" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $TOKEN" \
      -d '{"price": 899.99}' | jq '.'
    
    # Get updated product
    echo -e "\nUpdated iPhone:"
    curl -s -X GET "$BASE_URL/products/1" | jq '.'
    
else
    echo -e "\n3. Skipping product/category tests - no token received"
fi

echo -e "\n========================================="
echo "Product & Category API testing completed!"
