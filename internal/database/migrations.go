package database

import (
	"fmt"
	"log"
)

// RunMigrations runs all database migrations
func RunMigrations() error {
	log.Println("Running database migrations...")

	// Create users table
	if err := createUsersTable(); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	// Create categories table
	if err := createCategoriesTable(); err != nil {
		return fmt.Errorf("failed to create categories table: %w", err)
	}

	// Create products table
	if err := createProductsTable(); err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	// Create orders table
	if err := createOrdersTable(); err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	// Create order_items table
	if err := createOrderItemsTable(); err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}

	// Create carts table
	if err := createCartsTable(); err != nil {
		return fmt.Errorf("failed to create carts table: %w", err)
	}

	// Create cart_items table
	if err := createCartItemsTable(); err != nil {
		return fmt.Errorf("failed to create cart_items table: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createUsersTable creates the users table
func createUsersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password VARCHAR(255) NOT NULL,
		first_name VARCHAR(100) NOT NULL,
		last_name VARCHAR(100) NOT NULL,
		role VARCHAR(50) DEFAULT 'customer',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	log.Println("Users table created successfully")
	return nil
}

// createCategoriesTable creates the categories table
func createCategoriesTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS categories (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) UNIQUE NOT NULL,
		description TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create categories table: %w", err)
	}

	log.Println("Categories table created successfully")
	return nil
}

// createProductsTable creates the products table
func createProductsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
		stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
		category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
		image_url VARCHAR(500),
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create products table: %w", err)
	}

	log.Println("Products table created successfully")
	return nil
}

// createOrdersTable creates the orders table
func createOrdersTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS orders (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'shipped', 'delivered', 'cancelled')),
		total_amount DECIMAL(10,2) NOT NULL CHECK (total_amount >= 0),
		shipping_address TEXT NOT NULL,
		payment_method VARCHAR(100) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %w", err)
	}

	log.Println("Orders table created successfully")
	return nil
}

// createOrderItemsTable creates the order_items table
func createOrderItemsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS order_items (
		id SERIAL PRIMARY KEY,
		order_id INTEGER NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
		product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
		quantity INTEGER NOT NULL CHECK (quantity > 0),
		price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create order_items table: %w", err)
	}

	log.Println("Order items table created successfully")
	return nil
}

// createCartsTable creates the carts table
func createCartsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS carts (
		id SERIAL PRIMARY KEY,
		user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id)
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create carts table: %w", err)
	}

	log.Println("Carts table created successfully")
	return nil
}

// createCartItemsTable creates the cart_items table
func createCartItemsTable() error {
	query := `
	CREATE TABLE IF NOT EXISTS cart_items (
		id SERIAL PRIMARY KEY,
		cart_id INTEGER NOT NULL REFERENCES carts(id) ON DELETE CASCADE,
		product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE CASCADE,
		quantity INTEGER NOT NULL CHECK (quantity > 0),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(cart_id, product_id)
	);
	`

	_, err := DB.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create cart_items table: %w", err)
	}

	log.Println("Cart items table created successfully")
	return nil
}

// DropTables drops all tables (for testing/reset purposes)
func DropTables() error {
	log.Println("Dropping all tables...")

	queries := []string{
		"DROP TABLE IF EXISTS users CASCADE;",
		"DROP TABLE IF EXISTS products CASCADE;",
		"DROP TABLE IF EXISTS orders CASCADE;",
		"DROP TABLE IF EXISTS order_items CASCADE;",
	}

	for _, query := range queries {
		_, err := DB.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
	}

	log.Println("All tables dropped successfully")
	return nil
}
