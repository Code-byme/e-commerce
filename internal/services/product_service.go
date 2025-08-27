package services

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Code-byme/e-commerce/internal/database"
	"github.com/Code-byme/e-commerce/internal/models"
)

// ProductService handles product operations
type ProductService struct {
	db *sql.DB
}

// NewProductService creates a new product service
func NewProductService() *ProductService {
	return &ProductService{
		db: database.GetDB(),
	}
}

// CreateProductRequest represents the request to create a product
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Stock       int     `json:"stock" binding:"required,gte=0"`
	CategoryID  uint    `json:"category_id"`
	ImageURL    string  `json:"image_url"`
}

// UpdateProductRequest represents the request to update a product
type UpdateProductRequest struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Price       *float64 `json:"price"`
	Stock       *int     `json:"stock"`
	CategoryID  *uint    `json:"category_id"`
	ImageURL    *string  `json:"image_url"`
	IsActive    *bool    `json:"is_active"`
}

// ProductFilter represents product filtering options
type ProductFilter struct {
	CategoryID *uint    `json:"category_id"`
	MinPrice   *float64 `json:"min_price"`
	MaxPrice   *float64 `json:"max_price"`
	Search     *string  `json:"search"`
	IsActive   *bool    `json:"is_active"`
	Page       int      `json:"page"`
	Limit      int      `json:"limit"`
}

// ProductListResponse represents the paginated product list response
type ProductListResponse struct {
	Products []models.Product `json:"products"`
	Total    int              `json:"total"`
	Page     int              `json:"page"`
	Limit    int              `json:"limit"`
	Pages    int              `json:"pages"`
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(req *CreateProductRequest) (*models.Product, error) {
	// Check if category exists if category_id is provided
	if req.CategoryID > 0 {
		var category models.Category
		err := s.db.QueryRow("SELECT id FROM categories WHERE id = $1", req.CategoryID).Scan(&category.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.New("category not found")
			}
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	// Create product
	var product models.Product
	query := `
		INSERT INTO products (name, description, price, stock, category_id, image_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, name, description, price, stock, category_id, image_url, is_active, created_at, updated_at
	`

	err := s.db.QueryRow(
		query,
		req.Name, req.Description, req.Price, req.Stock, req.CategoryID, req.ImageURL, true,
	).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock,
		&product.CategoryID, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &product, nil
}

// GetProduct retrieves a product by ID
func (s *ProductService) GetProduct(id uint) (*models.Product, error) {
	var product models.Product
	query := `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name, c.description, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock,
		&product.CategoryID, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
		&product.Category.ID, &product.Category.Name, &product.Category.Description,
		&product.Category.CreatedAt, &product.Category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &product, nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(id uint, req *UpdateProductRequest) (*models.Product, error) {
	// Check if product exists
	existingProduct, err := s.GetProduct(id)
	if err != nil {
		return nil, err
	}

	// Check if category exists if category_id is being updated
	if req.CategoryID != nil && *req.CategoryID > 0 {
		var category models.Category
		err := s.db.QueryRow("SELECT id FROM categories WHERE id = $1", *req.CategoryID).Scan(&category.ID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, errors.New("category not found")
			}
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	// Build dynamic update query
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, *req.Name)
		argIndex++
	}

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.Price != nil {
		updates = append(updates, fmt.Sprintf("price = $%d", argIndex))
		args = append(args, *req.Price)
		argIndex++
	}

	if req.Stock != nil {
		updates = append(updates, fmt.Sprintf("stock = $%d", argIndex))
		args = append(args, *req.Stock)
		argIndex++
	}

	if req.CategoryID != nil {
		updates = append(updates, fmt.Sprintf("category_id = $%d", argIndex))
		args = append(args, *req.CategoryID)
		argIndex++
	}

	if req.ImageURL != nil {
		updates = append(updates, fmt.Sprintf("image_url = $%d", argIndex))
		args = append(args, *req.ImageURL)
		argIndex++
	}

	if req.IsActive != nil {
		updates = append(updates, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	if len(updates) == 0 {
		return existingProduct, nil
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE products SET %s WHERE id = $%d RETURNING id, name, description, price, stock, category_id, image_url, is_active, created_at, updated_at",
		strings.Join(updates, ", "), argIndex)

	var product models.Product
	err = s.db.QueryRow(query, args...).Scan(
		&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock,
		&product.CategoryID, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &product, nil
}

// DeleteProduct deletes a product
func (s *ProductService) DeleteProduct(id uint) error {
	// Check if product exists
	_, err := s.GetProduct(id)
	if err != nil {
		return err
	}

	// Soft delete by setting is_active to false
	_, err = s.db.Exec("UPDATE products SET is_active = false, updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// ListProducts retrieves a paginated list of products with filtering
func (s *ProductService) ListProducts(filter *ProductFilter) (*ProductListResponse, error) {
	// Set default values
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	// Build WHERE clause
	whereConditions := []string{}
	args := []interface{}{}
	argIndex := 1

	if filter.CategoryID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.category_id = $%d", argIndex))
		args = append(args, *filter.CategoryID)
		argIndex++
	}

	if filter.MinPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.price >= $%d", argIndex))
		args = append(args, *filter.MinPrice)
		argIndex++
	}

	if filter.MaxPrice != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.price <= $%d", argIndex))
		args = append(args, *filter.MaxPrice)
		argIndex++
	}

	if filter.Search != nil && *filter.Search != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	if filter.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("p.is_active = $%d", argIndex))
		args = append(args, *filter.IsActive)
		argIndex++
	} else {
		// Default to active products only
		whereConditions = append(whereConditions, "p.is_active = true")
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	// Count total products
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM products p %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Calculate pagination
	offset := (filter.Page - 1) * filter.Limit
	pages := (total + filter.Limit - 1) / filter.Limit

	// Get products
	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name, c.description, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.Limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock,
			&product.CategoryID, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
			&product.Category.ID, &product.Category.Name, &product.Category.Description,
			&product.Category.CreatedAt, &product.Category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return &ProductListResponse{
		Products: products,
		Total:    total,
		Page:     filter.Page,
		Limit:    filter.Limit,
		Pages:    pages,
	}, nil
}

// UpdateStock updates product stock
func (s *ProductService) UpdateStock(id uint, quantity int) error {
	_, err := s.db.Exec("UPDATE products SET stock = stock + $1, updated_at = NOW() WHERE id = $2", quantity, id)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}
	return nil
}

// GetProductsByCategory retrieves products by category ID
func (s *ProductService) GetProductsByCategory(categoryID uint, limit int) ([]models.Product, error) {
	if limit <= 0 {
		limit = 10
	}

	query := `
		SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, p.is_active, p.created_at, p.updated_at,
		       c.id, c.name, c.description, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = $1 AND p.is_active = true
		ORDER BY p.created_at DESC
		LIMIT $2
	`

	rows, err := s.db.Query(query, categoryID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query products by category: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		err := rows.Scan(
			&product.ID, &product.Name, &product.Description, &product.Price, &product.Stock,
			&product.CategoryID, &product.ImageURL, &product.IsActive, &product.CreatedAt, &product.UpdatedAt,
			&product.Category.ID, &product.Category.Name, &product.Category.Description,
			&product.Category.CreatedAt, &product.Category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}
