# E-commerce Go Backend

A modern e-commerce backend API built with Go and Gin framework.

## Features

- RESTful API design
- User authentication and authorization
- Product management
- Order processing
- Category management
- Health check endpoint
- PostgreSQL database integration

## Project Structure

```
e-commerce-go/
├── cmd/api/           # Application entry point
├── internal/          # Private application code
│   ├── handlers/      # HTTP request handlers
│   ├── models/        # Data models
│   ├── services/      # Business logic
│   └── database/      # Database operations
├── pkg/               # Public packages
│   ├── middleware/    # HTTP middleware
│   └── utils/         # Utility functions
├── main.go           # Main application file
├── go.mod            # Go module file
└── README.md         # Project documentation
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Git

### Database Setup

1. Install PostgreSQL on your system
2. Create a new database for the e-commerce application
3. Set up environment variables for database connection:

```bash
# Database Configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password_here
export DB_NAME=ecommerce
export DB_SSLMODE=disable
```

Or create a `.env` file in the project root:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=ecommerce
DB_SSLMODE=disable
```

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Code-byme/e-commerce.git
cd e-commerce
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run main.go
```

The server will start on `http://localhost:8080`

### Database Seeding

To populate your database with realistic product data for your portfolio, you can use the built-in seeding tool that fetches products from the FakeStore API.

#### Quick Start (Recommended)
```bash
# Seed with 20 products (default)
./seed-database.sh

# Seed with custom number of products
./seed-database.sh 50

# Seed without showing statistics
./seed-database.sh 30 false
```

#### Manual Seeding
```bash
# Show help
go run cmd/seed/main.go -help

# Seed with 20 products (default)
go run cmd/seed/main.go

# Seed with 50 products
go run cmd/seed/main.go -limit 50

# Show database statistics only
go run cmd/seed/main.go -stats

# Seed and show statistics
go run cmd/seed/main.go -limit 30 -stats
```

#### What Gets Seeded
- **Categories**: Electronics, Clothing, Home & Garden, Sports, Books, Toys, Automotive, Health, Jewelry, Food
- **Products**: Realistic products from FakeStore API with:
  - Product names and descriptions
  - Realistic pricing
  - Stock quantities based on popularity
  - Product images
  - Proper category mapping

#### Example Output
```
🛒 E-commerce Database Seeding Tool
==================================
📊 Checking database connection...
🚀 Starting database seeding...
   Products to fetch: 20
   Show statistics: true

Starting database seeding...
Created category: Electronics
Created category: Clothing
Created category: Home & Garden
...
Fetched 20 products from FakeStore API
Inserted product: Fjallraven - Foldsack No. 1 Backpack (Price: $109.95, Stock: 120)
Inserted product: Mens Casual Premium Slim Fit T-Shirts (Price: $22.30, Stock: 259)
...
Successfully inserted 20 new products

Final Database Statistics:
==========================
Database Statistics:
Categories: 10
Products: 20

Categories with product counts:
- Electronics: 6 products
- Clothing: 4 products
- Jewelry: 3 products
...
```

### API Endpoints

#### Public Endpoints
- `GET /health` - Health check endpoint (includes database status)

#### Authentication Endpoints
- `POST /auth/register` - Register a new user
- `POST /auth/login` - Login user and get JWT token

#### Protected Endpoints (require JWT token)
- `GET /api/profile` - Get current user profile
- `POST /api/products` - Create a new product (admin)
- `PUT /api/products/:id` - Update a product (admin)
- `DELETE /api/products/:id` - Delete a product (admin)
- `PATCH /api/products/:id/stock` - Update product stock (admin)
- `POST /api/categories` - Create a new category (admin)
- `PUT /api/categories/:id` - Update a category (admin)
- `DELETE /api/categories/:id` - Delete a category (admin)
- `POST /api/orders` - Create a new order
- `GET /api/orders` - List all orders (admin) or user's orders
- `GET /api/orders/my` - Get current user's orders
- `GET /api/orders/:id` - Get specific order details
- `PUT /api/orders/:id/status` - Update order status (admin)
- `DELETE /api/orders/:id` - Cancel order
- `GET /api/orders/statistics` - Get order statistics (admin)
- `GET /api/cart` - Get user's shopping cart
- `POST /api/cart/items` - Add item to cart
- `PUT /api/cart/items/:item_id` - Update cart item quantity
- `DELETE /api/cart/items/:item_id` - Remove item from cart
- `DELETE /api/cart` - Clear all items from cart
- `POST /api/cart/checkout` - Checkout cart and create order

#### Public Product Endpoints
- `GET /products` - List all products (with filtering and pagination)
- `GET /products/:id` - Get a specific product
- `GET /products/category/:category_id` - Get products by category

#### Public Category Endpoints
- `GET /categories` - List all categories
- `GET /categories/:id` - Get a specific category
- `GET /categories/:id/with-products` - Get category with product count

### Authentication API Usage

#### Register a new user
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

#### Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

#### Access protected endpoint
```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Product & Category API Usage

#### Create a category
```bash
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "Electronics",
    "description": "Electronic devices and gadgets"
  }'
```

#### Create a product
```bash
curl -X POST http://localhost:8080/api/products \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "iPhone 15 Pro",
    "description": "Latest iPhone with advanced features",
    "price": 999.99,
    "stock": 50,
    "category_id": 1,
    "image_url": "https://example.com/iphone15.jpg"
  }'
```

#### List products with filtering
```bash
# All products
curl -X GET "http://localhost:8080/products"

# Search products
curl -X GET "http://localhost:8080/products?search=iPhone"

# Filter by price range
curl -X GET "http://localhost:8080/products?min_price=100&max_price=1000"

# Filter by category
curl -X GET "http://localhost:8080/products?category_id=1"

# Pagination
curl -X GET "http://localhost:8080/products?page=1&limit=10"
```

#### Update product stock
```bash
curl -X PATCH http://localhost:8080/api/products/1/stock \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"quantity": -5}'
```

### Order Management API Usage

#### Create an order
```bash
curl -X POST http://localhost:8080/api/orders \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "shipping_address": "123 Main St, City, State 12345",
    "payment_method": "credit_card",
    "items": [
      {
        "product_id": 1,
        "quantity": 2
      },
      {
        "product_id": 2,
        "quantity": 1
      }
    ]
  }'
```

#### Get user's orders
```bash
curl -X GET "http://localhost:8080/api/orders/my" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get specific order
```bash
curl -X GET "http://localhost:8080/api/orders/1" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Update order status (admin)
```bash
curl -X PUT http://localhost:8080/api/orders/1/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"status": "shipped"}'
```

#### Cancel order
```bash
curl -X DELETE http://localhost:8080/api/orders/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Get order statistics (admin)
```bash
curl -X GET "http://localhost:8080/api/orders/statistics" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Shopping Cart API Usage

#### Get user's cart
```bash
curl -X GET "http://localhost:8080/api/cart" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Add item to cart
```bash
curl -X POST http://localhost:8080/api/cart/items \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "product_id": 1,
    "quantity": 2
  }'
```

#### Update cart item quantity
```bash
curl -X PUT http://localhost:8080/api/cart/items/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"quantity": 3}'
```

#### Remove item from cart
```bash
curl -X DELETE http://localhost:8080/api/cart/items/1 \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Clear cart
```bash
curl -X DELETE http://localhost:8080/api/cart \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### Checkout cart
```bash
curl -X POST http://localhost:8080/api/cart/checkout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "shipping_address": "123 Main St, City, State 12345",
    "payment_method": "credit_card"
  }'
```

## Development

### Adding Dependencies

```bash
go get <package-name>
```

### Running Tests

```bash
go test ./...
```

## License

This project is licensed under the MIT License.
