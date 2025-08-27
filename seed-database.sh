#!/bin/bash

# Database Seeding Script for E-commerce Portfolio
# This script fetches products from FakeStore API and stores them in your database

echo "🛒 E-commerce Database Seeding Tool"
echo "=================================="

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    exit 1
fi

# Check if database is running (optional check)
echo "📊 Checking database connection..."
if ! go run cmd/seed/main.go -stats &> /dev/null; then
    echo "⚠️  Warning: Could not connect to database. Make sure PostgreSQL is running."
    echo "   You can still try to seed the database..."
    echo ""
fi

# Default values
LIMIT=${1:-20}
SHOW_STATS=${2:-true}

echo "🚀 Starting database seeding..."
echo "   Products to fetch: $LIMIT"
echo "   Show statistics: $SHOW_STATS"
echo ""

# Run the seeding tool
if [ "$SHOW_STATS" = "true" ]; then
    go run cmd/seed/main.go -limit $LIMIT -stats
else
    go run cmd/seed/main.go -limit $LIMIT
fi

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Database seeding completed successfully!"
    echo ""
    echo "🎉 Your e-commerce database now has:"
    echo "   - Realistic product data from FakeStore API"
    echo "   - Proper categories and relationships"
    echo "   - Stock quantities and pricing"
    echo "   - Product images and descriptions"
    echo ""
    echo "🚀 You can now:"
    echo "   - Start your server: go run main.go"
    echo "   - Test the API: ./test-cart.sh"
    echo "   - View products: curl http://localhost:8080/products"
    echo ""
    echo "📚 For more options, run: go run cmd/seed/main.go -help"
else
    echo ""
    echo "❌ Database seeding failed!"
    echo "   Please check your database connection and try again."
    exit 1
fi
