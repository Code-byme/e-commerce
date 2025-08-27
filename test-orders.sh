#!/bin/bash

# Test script for order endpoints and complete e-commerce workflow

BASE_URL="http://localhost:8080"

echo "Testing E-commerce Order Management System"
echo "========================================="

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
    echo -e "\n4. Creating categories and products as admin..."
    
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
    
    echo -e "\n5. Customer browsing products..."
    
    # List all products
    echo "All available products:"
    curl -s -X GET "$BASE_URL/products" | jq '.'
    
    # Get specific product
    echo -e "\niPhone details:"
    curl -s -X GET "$BASE_URL/products/$IPHONE_ID" | jq '.'
    
    echo -e "\n6. Customer creating an order..."
    
    if [ "$CUSTOMER_TOKEN" != "null" ] && [ "$CUSTOMER_TOKEN" != "" ]; then
        # Create order with iPhone and MacBook
        ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/orders" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d "{
            \"shipping_address\": \"123 Main St, City, State 12345\",
            \"payment_method\": \"credit_card\",
            \"items\": [
              {
                \"product_id\": $IPHONE_ID,
                \"quantity\": 1
              },
              {
                \"product_id\": $MACBOOK_ID,
                \"quantity\": 1
              }
            ]
          }")
        
        echo "Order created:"
        echo "$ORDER_RESPONSE" | jq '.'
        
        # Extract order ID
        ORDER_ID=$(echo "$ORDER_RESPONSE" | jq -r '.data.id')
        
        echo -e "\n7. Customer viewing their orders..."
        
        # Get customer's orders
        echo "Customer's orders:"
        curl -s -X GET "$BASE_URL/api/orders/my" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Get specific order
        echo -e "\nOrder details:"
        curl -s -X GET "$BASE_URL/api/orders/$ORDER_ID" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        echo -e "\n8. Admin managing orders..."
        
        # Admin viewing all orders
        echo "All orders (admin view):"
        curl -s -X GET "$BASE_URL/api/orders" \
          -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.'
        
        # Update order status
        echo -e "\nUpdating order status to confirmed:"
        curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID/status" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $ADMIN_TOKEN" \
          -d '{"status": "confirmed"}' | jq '.'
        
        # Update order status to shipped
        echo -e "\nUpdating order status to shipped:"
        curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID/status" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $ADMIN_TOKEN" \
          -d '{"status": "shipped"}' | jq '.'
        
        # Update order status to delivered
        echo -e "\nUpdating order status to delivered:"
        curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID/status" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $ADMIN_TOKEN" \
          -d '{"status": "delivered"}' | jq '.'
        
        echo -e "\n9. Order statistics (admin only)..."
        
        # Get order statistics
        echo "Order statistics:"
        curl -s -X GET "$BASE_URL/api/orders/statistics" \
          -H "Authorization: Bearer $ADMIN_TOKEN" | jq '.'
        
        echo -e "\n10. Testing order cancellation..."
        
        # Create another order for cancellation test
        CANCEL_ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/orders" \
          -H "Content-Type: application/json" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" \
          -d "{
            \"shipping_address\": \"456 Oak St, City, State 12345\",
            \"payment_method\": \"paypal\",
            \"items\": [
              {
                \"product_id\": $IPHONE_ID,
                \"quantity\": 2
              }
            ]
          }")
        
        echo "Order for cancellation:"
        echo "$CANCEL_ORDER_RESPONSE" | jq '.'
        
        # Extract cancel order ID
        CANCEL_ORDER_ID=$(echo "$CANCEL_ORDER_RESPONSE" | jq -r '.data.id')
        
        # Cancel the order
        echo -e "\nCancelling order:"
        curl -s -X DELETE "$BASE_URL/api/orders/$CANCEL_ORDER_ID" \
          -H "Authorization: Bearer $CUSTOMER_TOKEN" | jq '.'
        
        # Check updated product stock (should be restored)
        echo -e "\nUpdated iPhone stock (should be restored):"
        curl -s -X GET "$BASE_URL/products/$IPHONE_ID" | jq '.data.stock'
        
    else
        echo -e "\n6. Skipping order creation - no customer token received"
    fi
    
else
    echo -e "\n4. Skipping product creation - no admin token received"
fi

echo -e "\n========================================="
echo "Order Management System testing completed!"
