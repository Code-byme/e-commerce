package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Code-byme/e-commerce/internal/database"
	"github.com/Code-byme/e-commerce/internal/models"
)

// CategoryService handles category operations
type CategoryService struct {
	db *sql.DB
}

// NewCategoryService creates a new category service
func NewCategoryService() *CategoryService {
	return &CategoryService{
		db: database.GetDB(),
	}
}

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(req *CreateCategoryRequest) (*models.Category, error) {
	// Check if category with same name already exists
	var existingCategory models.Category
	err := s.db.QueryRow("SELECT id FROM categories WHERE name = $1", req.Name).Scan(&existingCategory.ID)
	if err == nil {
		return nil, errors.New("category with this name already exists")
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Create category
	var category models.Category
	query := `
		INSERT INTO categories (name, description, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		RETURNING id, name, description, created_at, updated_at
	`

	err = s.db.QueryRow(query, req.Name, req.Description).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(id uint) (*models.Category, error) {
	var category models.Category
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		WHERE id = $1
	`

	err := s.db.QueryRow(query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("category not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &category, nil
}

// UpdateCategory updates an existing category
func (s *CategoryService) UpdateCategory(id uint, req *UpdateCategoryRequest) (*models.Category, error) {
	// Check if category exists
	existingCategory, err := s.GetCategory(id)
	if err != nil {
		return nil, err
	}

	// Check if name is being updated and if it conflicts with existing category
	if req.Name != nil && *req.Name != existingCategory.Name {
		var conflictCategory models.Category
		err := s.db.QueryRow("SELECT id FROM categories WHERE name = $1 AND id != $2", *req.Name, id).Scan(&conflictCategory.ID)
		if err == nil {
			return nil, errors.New("category with this name already exists")
		} else if err != sql.ErrNoRows {
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	// Build update query
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

	if len(updates) == 0 {
		return existingCategory, nil
	}

	updates = append(updates, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE categories SET %s WHERE id = $%d RETURNING id, name, description, created_at, updated_at",
		updates[0], argIndex)

	if len(updates) > 1 {
		query = fmt.Sprintf("UPDATE categories SET %s WHERE id = $%d RETURNING id, name, description, created_at, updated_at",
			updates[0], argIndex)
	}

	var category models.Category
	err = s.db.QueryRow(query, args...).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &category, nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(id uint) error {
	// Check if category exists
	_, err := s.GetCategory(id)
	if err != nil {
		return err
	}

	// Check if category has products
	var productCount int
	err = s.db.QueryRow("SELECT COUNT(*) FROM products WHERE category_id = $1", id).Scan(&productCount)
	if err != nil {
		return fmt.Errorf("failed to check category products: %w", err)
	}

	if productCount > 0 {
		return errors.New("cannot delete category with existing products")
	}

	// Delete category
	_, err = s.db.Exec("DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}

// ListCategories retrieves all categories
func (s *CategoryService) ListCategories() ([]models.Category, error) {
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(
			&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// GetCategoryWithProductCount retrieves a category with product count
func (s *CategoryService) GetCategoryWithProductCount(id uint) (*models.Category, int, error) {
	var category models.Category
	var productCount int

	query := `
		SELECT c.id, c.name, c.description, c.created_at, c.updated_at,
		       COUNT(p.id) as product_count
		FROM categories c
		LEFT JOIN products p ON c.id = p.category_id AND p.is_active = true
		WHERE c.id = $1
		GROUP BY c.id, c.name, c.description, c.created_at, c.updated_at
	`

	err := s.db.QueryRow(query, id).Scan(
		&category.ID, &category.Name, &category.Description, &category.CreatedAt, &category.UpdatedAt,
		&productCount,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, errors.New("category not found")
		}
		return nil, 0, fmt.Errorf("database error: %w", err)
	}

	return &category, productCount, nil
}
