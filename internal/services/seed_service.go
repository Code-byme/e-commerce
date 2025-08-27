package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Code-byme/e-commerce/internal/database"
)

// SeedService handles seeding the database with sample data
type SeedService struct {
	db *sql.DB
}

// NewSeedService creates a new seed service
func NewSeedService() *SeedService {
	return &SeedService{
		db: database.GetDB(),
	}
}

// FakeStoreProduct represents a product from FakeStore API
type FakeStoreProduct struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Image       string  `json:"image"`
	Rating      struct {
		Rate  float64 `json:"rate"`
		Count int     `json:"count"`
	} `json:"rating"`
}

// SeedCategories creates default categories
func (s *SeedService) SeedCategories() error {
	categories := []struct {
		name        string
		description string
	}{
		{"Electronics", "Electronic devices and gadgets"},
		{"Clothing", "Fashion and apparel"},
		{"Home & Garden", "Home improvement and garden supplies"},
		{"Sports", "Sports equipment and accessories"},
		{"Books", "Books and literature"},
		{"Toys", "Toys and games"},
		{"Automotive", "Automotive parts and accessories"},
		{"Health", "Health and beauty products"},
		{"Jewelry", "Jewelry and accessories"},
		{"Food", "Food and beverages"},
	}

	for _, cat := range categories {
		// Check if category already exists
		var exists bool
		err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM categories WHERE name = $1)", cat.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check category existence: %w", err)
		}

		if !exists {
			_, err = s.db.Exec(
				"INSERT INTO categories (name, description, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())",
				cat.name, cat.description,
			)
			if err != nil {
				return fmt.Errorf("failed to insert category %s: %w", cat.name, err)
			}
			fmt.Printf("Created category: %s\n", cat.name)
		} else {
			fmt.Printf("Category already exists: %s\n", cat.name)
		}
	}

	return nil
}

// GetCategoryIDByName gets category ID by name
func (s *SeedService) GetCategoryIDByName(name string) (uint, error) {
	var id uint
	err := s.db.QueryRow("SELECT id FROM categories WHERE name = $1", name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("category not found: %s", name)
	}
	return id, nil
}

// MapFakeStoreCategory maps FakeStore categories to our categories
func (s *SeedService) MapFakeStoreCategory(fakeCategory string) string {
	categoryMap := map[string]string{
		"electronics":      "Electronics",
		"jewelery":         "Jewelry",
		"men's clothing":   "Clothing",
		"women's clothing": "Clothing",
		"home":             "Home & Garden",
		"sports":           "Sports",
		"books":            "Books",
		"toys":             "Toys",
		"automotive":       "Automotive",
		"health":           "Health",
		"food":             "Food",
	}

	if mapped, exists := categoryMap[strings.ToLower(fakeCategory)]; exists {
		return mapped
	}
	return "Electronics" // Default fallback
}

// SeedProducts fetches products from FakeStore API and stores them
func (s *SeedService) SeedProducts(limit int) error {
	// First, ensure categories exist
	if err := s.SeedCategories(); err != nil {
		return fmt.Errorf("failed to seed categories: %w", err)
	}

	// Fetch products from FakeStore API
	url := fmt.Sprintf("https://fakestoreapi.com/products?limit=%d", limit)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch products from API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	var fakeProducts []FakeStoreProduct
	if err := json.Unmarshal(body, &fakeProducts); err != nil {
		return fmt.Errorf("failed to unmarshal products: %w", err)
	}

	fmt.Printf("Fetched %d products from FakeStore API\n", len(fakeProducts))

	// Insert products into database
	insertedCount := 0
	for _, fakeProduct := range fakeProducts {
		// Map category
		categoryName := s.MapFakeStoreCategory(fakeProduct.Category)
		categoryID, err := s.GetCategoryIDByName(categoryName)
		if err != nil {
			fmt.Printf("Warning: Could not find category %s for product %s, skipping\n", categoryName, fakeProduct.Title)
			continue
		}

		// Check if product already exists (by title)
		var exists bool
		err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE name = $1)", fakeProduct.Title).Scan(&exists)
		if err != nil {
			fmt.Printf("Warning: Could not check product existence for %s: %v\n", fakeProduct.Title, err)
			continue
		}

		if exists {
			fmt.Printf("Product already exists: %s\n", fakeProduct.Title)
			continue
		}

		// Generate realistic stock based on rating count
		stock := fakeProduct.Rating.Count
		if stock == 0 {
			stock = 50 // Default stock
		}
		if stock > 200 {
			stock = 200 // Cap at reasonable amount
		}

		// Insert product
		_, err = s.db.Exec(`
			INSERT INTO products (name, description, price, stock, category_id, image_url, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		`, fakeProduct.Title, fakeProduct.Description, fakeProduct.Price, stock, categoryID, fakeProduct.Image, true)

		if err != nil {
			fmt.Printf("Warning: Failed to insert product %s: %v\n", fakeProduct.Title, err)
			continue
		}

		insertedCount++
		fmt.Printf("Inserted product: %s (Price: $%.2f, Stock: %d)\n", fakeProduct.Title, fakeProduct.Price, stock)
	}

	fmt.Printf("\nSuccessfully inserted %d new products\n", insertedCount)
	return nil
}

// SeedSampleData seeds the database with sample data
func (s *SeedService) SeedSampleData() error {
	fmt.Println("Starting database seeding...")

	// Seed categories and products
	if err := s.SeedProducts(20); err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}

	fmt.Println("Database seeding completed successfully!")
	return nil
}

// GetDatabaseStats returns statistics about the database
func (s *SeedService) GetDatabaseStats() error {
	var categoryCount, productCount int

	// Count categories
	err := s.db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&categoryCount)
	if err != nil {
		return fmt.Errorf("failed to count categories: %w", err)
	}

	// Count products
	err = s.db.QueryRow("SELECT COUNT(*) FROM products").Scan(&productCount)
	if err != nil {
		return fmt.Errorf("failed to count products: %w", err)
	}

	fmt.Printf("\nDatabase Statistics:\n")
	fmt.Printf("Categories: %d\n", categoryCount)
	fmt.Printf("Products: %d\n", productCount)

	// Show categories with product counts
	rows, err := s.db.Query(`
		SELECT c.name, COUNT(p.id) as product_count
		FROM categories c
		LEFT JOIN products p ON c.id = p.category_id
		GROUP BY c.id, c.name
		ORDER BY product_count DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to query category stats: %w", err)
	}
	defer rows.Close()

	fmt.Printf("\nCategories with product counts:\n")
	for rows.Next() {
		var categoryName string
		var productCount int
		if err := rows.Scan(&categoryName, &productCount); err != nil {
			return fmt.Errorf("failed to scan category stats: %w", err)
		}
		fmt.Printf("- %s: %d products\n", categoryName, productCount)
	}

	return nil
}
