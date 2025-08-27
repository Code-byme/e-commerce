#!/bin/bash

# Test script for shopping cart functionality

BASE_URL="http://localhost:8080"

echo "Testing E-commerce Shopping Cart System"
echo "======================================"

# Test health endpoint
echo -e "\n1. Testing health endpoint..."
curl -s -X GET "$BASE_URL/health" | jq '.'

# Register a customer user
echo -e "\n2. Registering customer user..."
CUSTOMER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "customer@example.com",
    "password": "customer123",
    "first_name": "John",
    "last_name": "Customer"
  }')

echo "$CUSTOMER_RESPONSE" | jq '.'

# Extract customer token
CUSTOMER_TOKEN=$(echo "$CUSTOMER_RESPONSE" | jq -r '.data.token')

# Register an admin user
echo -e "\n3. Registering admin user..."
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "first_name": "Admin",
    "last_name": "User"
  }')

echo "$ADMIN_RESPONSE" | jq '.'

# Extract admin token
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.data.token')

if [ "$ADMIN_TOKEN" != "null" ] && [ "$ADMIN_TOKEN" != "" ]; then
    echo -e "\n4. Creating products as admin..."
    
    # Create Electronics category
    ELECTRONICS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/categories" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d '{
        "name": "Electronics",
        "description": "Electronic devices and gadgets"
      }')
    
    echo "Electronics category:"
    echo "$ELECTRONICS_RESPONSE" | jq '.'
    
    # Extract Electronics category ID
    ELECTRONICS_ID=$(echo "$ELECTRONICS_RESPONSE" | jq -r '.data.id')
    
    # Create iPhone product
    IPHONE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/products" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
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
    
    # Extract iPhone product ID
    IPHONE_ID=$(echo "$IPHONE_RESPONSE" | jq -r '.data.id')
    
    # Create MacBook product
    MACBOOK_RESPONSE=$(curl -s -X POST "$BASE_URL/api/products" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d "{
        \"name\": \"MacBook Pro\",
        \"description\": \"Professional laptop for developers\",
        \"price\": 1999.99,
        \"stock\": 25,
        \"category_id\": $ELECTRONICS_ID,
        \"image_url\": \"https://example.com/macbook.jpg\"
      }")
    
    echo "MacBook product:"
    echo "$MACBOOK_RESPONSE" | jq '.'
    
    # Extract MacBook product ID
    MACBOOK_ID=$(echo "$MACBOOK_RESPONSE" | jq -r '.data.id')
    
    # Create AirPods product
    AIRPODS_RESPONSE=$(curl -s -X POST "$BASE_URL/api/products" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $ADMIN_TOKEN" \
      -d "{
        \"name\": \"AirPods Pro\",
        \"description\": \"Wireless earbuds with noise cancellation\",
        \"price\": 249.99,
        \"stock\": 100,
        \"category_id\": $ELECTRONICS_ID,
        \"image_url\": \"https://example.com/airpods.jpg\"
      }")
    
    echo "AirPods product:"
    echo "$AIRPODS_RESPONSE" | jq '.'
    
    # Extract AirPods product ID
    AIRPODS_ID=$(echo "$AIRPODS_RESPONSE" | jq -r '.data.id')
    
    echo -e "\n5. Customer shopping cart operations..."
    
    if [ "$CUSTOMER_TOKEN" != "null" ] && [ "$CUSTOMER_TOKEN" != "" ]; then
        # Get initial cart (should be empty)
        echo "Initial cart (should be empty):"
        curl -s -X GET "$BASE_URL/api/cart" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Add iPhone to cart
        echo -e "\nAdding iPhone to cart:"
        IPHONE_CART_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cart/items" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d "{
            \"product_id\": $IPHONE_ID,
            \"quantity\": 1
          }")
        
        echo "$IPHONE_CART_RESPONSE" | jq '.'
        
        # Add MacBook to cart
        echo -e "\nAdding MacBook to cart:"
        MACBOOK_CART_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cart/items" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d "{
            \"product_id\": $MACBOOK_ID,
            \"quantity\": 1
          }")
        
        echo "$MACBOOK_CART_RESPONSE" | jq '.'
        
        # Add AirPods to cart
        echo -e "\nAdding AirPods to cart:"
        AIRPODS_CART_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cart/items" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d "{
            \"product_id\": $AIRPODS_ID,
            \"quantity\": 2
          }")
        
        echo "$AIRPODS_CART_RESPONSE" | jq '.'
        
        # Get cart with all items
        echo -e "\nCart with all items:"
        CART_RESPONSE=$(curl -s -X GET "$BASE_URL/api/cart" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN")
        
        echo "$CART_RESPONSE" | jq '.'
        
        # Extract cart item IDs for updates
        IPHONE_ITEM_ID=$(echo "$CART_RESPONSE" | jq -r '.data.cart_items[] | select(.product.name == "iPhone 15 Pro") | .id')
        AIRPODS_ITEM_ID=$(echo "$CART_RESPONSE" | jq -r '.data.cart_items[] | select(.product.name == "AirPods Pro") | .id')
        
        # Update iPhone quantity
        echo -e "\nUpdating iPhone quantity to 2:"
        curl -s -X PUT "$BASE_URL/api/cart/items/$IPHONE_ITEM_ID" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d '{"quantity": 2}' | jq '.'
        
        # Update AirPods quantity
        echo -e "\nUpdating AirPods quantity to 1:"
        curl -s -X PUT "$BASE_URL/api/cart/items/$AIRPODS_ITEM_ID" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d '{"quantity": 1}' | jq '.'
        
        # Get updated cart
        echo -e "\nUpdated cart:"
        curl -s -X GET "$BASE_URL/api/cart" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Remove AirPods from cart
        echo -e "\nRemoving AirPods from cart:"
        curl -s -X DELETE "$BASE_URL/api/cart/items/$AIRPODS_ITEM_ID" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Get cart after removal
        echo -e "\nCart after removing AirPods:"
        curl -s -X GET "$BASE_URL/api/cart" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Test adding same product again (should update quantity)
        echo -e "\nAdding iPhone again (should update quantity):"
        curl -s -X POST "$BASE_URL/api/cart/items" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d "{
            \"product_id\": $IPHONE_ID,
            \"quantity\": 1
          }" | jq '.'
        
        # Get final cart before checkout
        echo -e "\nFinal cart before checkout:"
        FINAL_CART_RESPONSE=$(curl -s -X GET "$BASE_URL/api/cart" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN")
        
        echo "$FINAL_CART_RESPONSE" | jq '.'
        
        # Checkout cart
        echo -e "\nChecking out cart:"
        CHECKOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cart/checkout" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d '{
            "shipping_address": "123 Main St, City, State 12345",
            "payment_method": "credit_card"
          }')
        
        echo "$CHECKOUT_RESPONSE" | jq '.'
        
        # Extract order ID
        ORDER_ID=$(echo "$CHECKOUT_RESPONSE" | jq -r '.data.id')
        
        # Verify cart is empty after checkout
        echo -e "\nCart after checkout (should be empty):"
        curl -s -X GET "$BASE_URL/api/cart" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Get order details
        echo -e "\nOrder details:"
        curl -s -X GET "$BASE_URL/api/orders/$ORDER_ID" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Test cart operations with empty cart
        echo -e "\n6. Testing cart operations with empty cart..."
        
        # Try to checkout empty cart
        echo "Trying to checkout empty cart:"
        curl -s -X POST "$BASE_URL/api/cart/checkout" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d '{
            "shipping_address": "456 Oak St, City, State 12345",
            "payment_method": "paypal"
          }" | jq '.'
        
        # Clear cart (should work even if empty)
        echo -e "\nClearing cart:"
        curl -s -X DELETE "$BASE_URL/api/cart" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Test adding product with insufficient stock
        echo -e "\n7. Testing insufficient stock scenario..."
        
        # Try to add more iPhones than available
        echo "Trying to add more iPhones than available:"
        curl -s -X POST "$BASE_URL/api/cart/items" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d "{
            \"product_id\": $IPHONE_ID,
            \"quantity\": 100
          }" | jq '.'
        
    else
        echo -e "\n5. Skipping cart operations - no customer token received"
    fi
    
else
    echo -e "\n4. Skipping product creation - no admin token received"
fi

echo -e "\n======================================"
echo "Shopping Cart System testing completed!"
