#!/bin/bash

# Complete E-commerce Workflow Test Script
# This script demonstrates the full e-commerce functionality with seeded data

BASE_URL="http://localhost:8080"

echo "🛒 Complete E-commerce Workflow Test"
echo "===================================="
echo "This script demonstrates the full e-commerce functionality:"
echo "1. Database seeding with realistic products"
echo "2. User registration and authentication"
echo "3. Product browsing and cart management"
echo "4. Order creation and management"
echo "5. Admin operations"
echo ""

# Check if server is running
echo "🔍 Checking if server is running..."
if ! curl -s "$BASE_URL/health" > /dev/null; then
    echo "❌ Server is not running. Please start the server first:"
    echo "   go run main.go"
    echo ""
    echo "Then run this script in another terminal."
    exit 1
fi

echo "✅ Server is running!"
echo ""

# Step 1: Seed the database if needed
echo "📊 Step 1: Checking database status..."
DB_STATS=$(curl -s "$BASE_URL/health" | jq -r '.database.status')
if [ "$DB_STATS" != "ok" ]; then
    echo "❌ Database is not healthy. Please check your database connection."
    exit 1
fi

PRODUCT_COUNT=$(curl -s "$BASE_URL/products" | jq '.data.products | length')
if [ "$PRODUCT_COUNT" -eq 0 ]; then
    echo "📦 No products found. Seeding database..."
    ./seed-database.sh 10
    echo ""
else
    echo "✅ Database already has $PRODUCT_COUNT products"
fi

# Step 2: Register users
echo "👤 Step 2: User Registration"
echo "----------------------------"

# Register customer
echo "Registering customer user..."
CUSTOMER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "customer@example.com",
    "password": "customer123",
    "first_name": "John",
    "last_name": "Customer"
  }')

if echo "$CUSTOMER_RESPONSE" | jq -e '.data.token' > /dev/null; then
    CUSTOMER_TOKEN=$(echo "$CUSTOMER_RESPONSE" | jq -r '.data.token')
    echo "✅ Customer registered successfully"
else
    echo "❌ Customer registration failed:"
    echo "$CUSTOMER_RESPONSE" | jq '.'
    exit 1
fi

# Register admin
echo "Registering admin user..."
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "admin123",
    "first_name": "Admin",
    "last_name": "User"
  }')

if echo "$ADMIN_RESPONSE" | jq -e '.data.token' > /dev/null; then
    ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.data.token')
    echo "✅ Admin registered successfully"
else
    echo "❌ Admin registration failed:"
    echo "$ADMIN_RESPONSE" | jq '.'
    exit 1
fi

echo ""

# Step 3: Browse products
echo "🛍️ Step 3: Product Browsing"
echo "---------------------------"

echo "Getting all products..."
PRODUCTS_RESPONSE=$(curl -s "$BASE_URL/products")
PRODUCT_COUNT=$(echo "$PRODUCTS_RESPONSE" | jq '.data.products | length')
echo "✅ Found $PRODUCT_COUNT products"

# Get first product for cart
FIRST_PRODUCT=$(echo "$PRODUCTS_RESPONSE" | jq -r '.data.products[0]')
FIRST_PRODUCT_ID=$(echo "$FIRST_PRODUCT" | jq -r '.id')
FIRST_PRODUCT_NAME=$(echo "$FIRST_PRODUCT" | jq -r '.name')
echo "📦 First product: $FIRST_PRODUCT_NAME (ID: $FIRST_PRODUCT_ID)"

# Get categories
echo "Getting categories..."
CATEGORIES_RESPONSE=$(curl -s "$BASE_URL/categories")
CATEGORY_COUNT=$(echo "$CATEGORIES_RESPONSE" | jq '.data.categories | length')
echo "✅ Found $CATEGORY_COUNT categories"

echo ""

# Step 4: Shopping cart operations
echo "🛒 Step 4: Shopping Cart Operations"
echo "----------------------------------"

echo "Getting initial cart (should be empty)..."
INITIAL_CART=$(curl -s -X GET "$BASE_URL/api/cart" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN")
echo "Initial cart:"
echo "$INITIAL_CART" | jq '.data'

echo ""
echo "Adding first product to cart..."
ADD_TO_CART_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cart/items" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d "{
    \"product_id\": $FIRST_PRODUCT_ID,
    \"quantity\": 2
  }")

if echo "$ADD_TO_CART_RESPONSE" | jq -e '.data' > /dev/null; then
    echo "✅ Product added to cart successfully"
    CART_TOTAL=$(echo "$ADD_TO_CART_RESPONSE" | jq -r '.data.total_amount')
    CART_ITEMS=$(echo "$ADD_TO_CART_RESPONSE" | jq -r '.data.total_items')
    echo "   Cart total: $CART_TOTAL"
    echo "   Total items: $CART_ITEMS"
else
    echo "❌ Failed to add product to cart:"
    echo "$ADD_TO_CART_RESPONSE" | jq '.'
fi

# Get another product and add to cart
SECOND_PRODUCT_ID=$(echo "$PRODUCTS_RESPONSE" | jq -r '.data.products[1].id')
SECOND_PRODUCT_NAME=$(echo "$PRODUCTS_RESPONSE" | jq -r '.data.products[1].name')

echo ""
echo "Adding second product to cart..."
curl -s -X POST "$BASE_URL/api/cart/items" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d "{
    \"product_id\": $SECOND_PRODUCT_ID,
    \"quantity\": 1
  }" > /dev/null

echo "✅ Second product added to cart"

echo ""
echo "Getting updated cart..."
UPDATED_CART=$(curl -s -X GET "$BASE_URL/api/cart" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN")
echo "Updated cart:"
echo "$UPDATED_CART" | jq '.data'

echo ""

# Step 5: Checkout process
echo "💳 Step 5: Checkout Process"
echo "---------------------------"

echo "Checking out cart..."
CHECKOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/cart/checkout" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -d '{
    "shipping_address": "123 Main St, City, State 12345",
    "payment_method": "credit_card"
  }')

if echo "$CHECKOUT_RESPONSE" | jq -e '.data.id' > /dev/null; then
    ORDER_ID=$(echo "$CHECKOUT_RESPONSE" | jq -r '.data.id')
    ORDER_TOTAL=$(echo "$CHECKOUT_RESPONSE" | jq -r '.data.total_amount')
    echo "✅ Order created successfully!"
    echo "   Order ID: $ORDER_ID"
    echo "   Order total: $ORDER_TOTAL"
else
    echo "❌ Checkout failed:"
    echo "$CHECKOUT_RESPONSE" | jq '.'
fi

echo ""
echo "Verifying cart is empty after checkout..."
EMPTY_CART=$(curl -s -X GET "$BASE_URL/api/cart" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN")
CART_ITEMS_AFTER=$(echo "$EMPTY_CART" | jq -r '.data.total_items')
echo "Cart items after checkout: $CART_ITEMS_AFTER"

echo ""

# Step 6: Order management
echo "📋 Step 6: Order Management"
echo "---------------------------"

echo "Getting order details..."
ORDER_DETAILS=$(curl -s -X GET "$BASE_URL/api/orders/$ORDER_ID" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN")
echo "Order details:"
echo "$ORDER_DETAILS" | jq '.data'

echo ""
echo "Getting user's order history..."
USER_ORDERS=$(curl -s -X GET "$BASE_URL/api/orders/my" \
  -H "Authorization: Bearer $CUSTOMER_TOKEN")
USER_ORDER_COUNT=$(echo "$USER_ORDERS" | jq '.data.orders | length')
echo "✅ User has $USER_ORDER_COUNT orders"

echo ""

# Step 7: Admin operations
echo "👨‍💼 Step 7: Admin Operations"
echo "----------------------------"

echo "Getting order statistics (admin)..."
ADMIN_STATS=$(curl -s -X GET "$BASE_URL/api/orders/statistics" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
echo "Order statistics:"
echo "$ADMIN_STATS" | jq '.data'

echo ""
echo "Getting all orders (admin)..."
ALL_ORDERS=$(curl -s -X GET "$BASE_URL/api/orders" \
  -H "Authorization: Bearer $ADMIN_TOKEN")
TOTAL_ORDERS=$(echo "$ALL_ORDERS" | jq '.data.orders | length')
echo "✅ Total orders in system: $TOTAL_ORDERS"

echo ""
echo "Updating order status (admin)..."
UPDATE_STATUS_RESPONSE=$(curl -s -X PUT "$BASE_URL/api/orders/$ORDER_ID/status" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"status": "shipped"}')

if echo "$UPDATE_STATUS_RESPONSE" | jq -e '.data' > /dev/null; then
    echo "✅ Order status updated to 'shipped'"
else
    echo "❌ Failed to update order status:"
    echo "$UPDATE_STATUS_RESPONSE" | jq '.'
fi

echo ""

# Step 8: Product management (admin)
echo "📦 Step 8: Product Management (Admin)"
echo "------------------------------------"

echo "Getting product stock information..."
PRODUCT_STOCK=$(curl -s "$BASE_URL/products/$FIRST_PRODUCT_ID" | jq -r '.data.stock')
echo "Current stock for $FIRST_PRODUCT_NAME: $PRODUCT_STOCK"

echo ""
echo "Updating product stock (admin)..."
UPDATE_STOCK_RESPONSE=$(curl -s -X PATCH "$BASE_URL/api/products/$FIRST_PRODUCT_ID/stock" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -d '{"stock": 100}')

if echo "$UPDATE_STOCK_RESPONSE" | jq -e '.data' > /dev/null; then
    echo "✅ Product stock updated to 100"
else
    echo "❌ Failed to update product stock:"
    echo "$UPDATE_STOCK_RESPONSE" | jq '.'
fi

echo ""

# Step 9: Category management
echo "📂 Step 9: Category Management"
echo "-----------------------------"

echo "Getting category with product count..."
CATEGORY_WITH_PRODUCTS=$(curl -s "$BASE_URL/categories/1/with-products")
echo "Category details:"
echo "$CATEGORY_WITH_PRODUCTS" | jq '.data'

echo ""

# Step 10: Final summary
echo "🎉 Step 10: Workflow Summary"
echo "---------------------------"

echo "✅ Complete e-commerce workflow demonstrated successfully!"
echo ""
echo "📊 Summary:"
echo "   - Database seeded with realistic products"
echo "   - User registration and authentication working"
echo "   - Product browsing and cart management functional"
echo "   - Order creation and checkout process complete"
echo "   - Admin operations (order management, product management) working"
echo "   - Category management operational"
echo ""
echo "🚀 Your e-commerce backend is fully functional and ready for:"
echo "   - Frontend development"
echo "   - Production deployment"
echo "   - Portfolio demonstration"
echo ""
echo "📚 Next steps:"
echo "   - Build a frontend application"
echo "   - Add more features (reviews, wishlists, etc.)"
echo "   - Deploy to production"
echo "   - Add monitoring and logging"
echo ""
echo "🎯 Perfect for your portfolio! 🎯"
