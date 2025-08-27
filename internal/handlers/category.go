package handlers

import (
	"net/http"
	"strconv"

	"github.com/Code-byme/e-commerce/internal/services"
	"github.com/gin-gonic/gin"
)

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	categoryService *services.CategoryService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{
		categoryService: services.NewCategoryService(),
	}
}

// CreateCategory handles category creation
func (h *CategoryHandler) CreateCategory(c *gin.Context) {
	var req services.CreateCategoryRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Create category
	category, err := h.categoryService.CreateCategory(&req)
	if err != nil {
		if err.Error() == "category with this name already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Category with this name already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create category",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Category created successfully",
		"data":    category,
	})
}

// GetCategory handles retrieving a single category
func (h *CategoryHandler) GetCategory(c *gin.Context) {
	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	// Get category
	category, err := h.categoryService.GetCategory(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": category,
	})
}

// UpdateCategory handles category updates
func (h *CategoryHandler) UpdateCategory(c *gin.Context) {
	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	var req services.UpdateCategoryRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Update category
	category, err := h.categoryService.UpdateCategory(uint(id), &req)
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
			})
			return
		}
		if err.Error() == "category with this name already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Category with this name already exists",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category updated successfully",
		"data":    category,
	})
}

// DeleteCategory handles category deletion
func (h *CategoryHandler) DeleteCategory(c *gin.Context) {
	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	// Delete category
	err = h.categoryService.DeleteCategory(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
			})
			return
		}
		if err.Error() == "cannot delete category with existing products" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot delete category with existing products",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}

// ListCategories handles category listing
func (h *CategoryHandler) ListCategories(c *gin.Context) {
	// Get categories
	categories, err := h.categoryService.ListCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve categories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": categories,
	})
}

// GetCategoryWithProductCount handles retrieving a category with product count
func (h *CategoryHandler) GetCategoryWithProductCount(c *gin.Context) {
	// Parse category ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid category ID",
		})
		return
	}

	// Get category with product count
	category, productCount, err := h.categoryService.GetCategoryWithProductCount(uint(id))
	if err != nil {
		if err.Error() == "category not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Category not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"category":      category,
			"product_count": productCount,
		},
	})
}
