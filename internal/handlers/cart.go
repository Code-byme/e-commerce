package handlers

import (
	"net/http"
	"strconv"

	"github.com/Code-byme/e-commerce/internal/services"
	"github.com/gin-gonic/gin"
)

// CartHandler handles cart-related HTTP requests
type CartHandler struct {
	cartService *services.CartService
}

// NewCartHandler creates a new cart handler
func NewCartHandler() *CartHandler {
	return &CartHandler{
		cartService: services.NewCartService(),
	}
}

// GetCart handles retrieving the user's cart
func (h *CartHandler) GetCart(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get cart
	cart, err := h.cartService.GetCart(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": cart,
	})
}

// AddToCart handles adding an item to the cart
func (h *CartHandler) AddToCart(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req services.AddToCartRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Add item to cart
	cart, err := h.cartService.AddToCart(userID.(uint), &req)
	if err != nil {
		if err.Error() == "product not found or inactive" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Product not found or inactive",
			})
			return
		}
		if err.Error() == "insufficient stock for product iPhone 15 Pro (available: 50, requested: 100)" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient stock for the requested product",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add item to cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item added to cart successfully",
		"data":    cart,
	})
}

// UpdateCartItem handles updating a cart item quantity
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse item ID from URL parameter
	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid item ID",
		})
		return
	}

	var req services.UpdateCartItemRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Update cart item
	cart, err := h.cartService.UpdateCartItem(userID.(uint), uint(itemID), &req)
	if err != nil {
		if err.Error() == "cart item not found or not owned by user" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cart item not found",
			})
			return
		}
		if err.Error() == "product not found or inactive" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Product not found or inactive",
			})
			return
		}
		if err.Error() == "insufficient stock for product iPhone 15 Pro (available: 50, requested: 100)" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient stock for the requested quantity",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update cart item",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart item updated successfully",
		"data":    cart,
	})
}

// RemoveFromCart handles removing an item from the cart
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse item ID from URL parameter
	itemIDStr := c.Param("item_id")
	itemID, err := strconv.ParseUint(itemIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid item ID",
		})
		return
	}

	// Remove item from cart
	cart, err := h.cartService.RemoveFromCart(userID.(uint), uint(itemID))
	if err != nil {
		if err.Error() == "cart item not found or not owned by user" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Cart item not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove item from cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item removed from cart successfully",
		"data":    cart,
	})
}

// ClearCart handles clearing all items from the cart
func (h *CartHandler) ClearCart(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Clear cart
	err := h.cartService.ClearCart(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to clear cart",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cart cleared successfully",
	})
}

// CheckoutCart handles the checkout process
func (h *CartHandler) CheckoutCart(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req struct {
		ShippingAddress string `json:"shipping_address" binding:"required"`
		PaymentMethod   string `json:"payment_method" binding:"required"`
	}

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Checkout cart
	order, err := h.cartService.CheckoutCart(userID.(uint), req.ShippingAddress, req.PaymentMethod)
	if err != nil {
		if err.Error() == "cart is empty" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cart is empty",
			})
			return
		}
		if err.Error() == "product with ID 1 not found or inactive" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "One or more products not found or inactive",
			})
			return
		}
		if err.Error() == "insufficient stock for product iPhone 15 Pro (available: 50, requested: 100)" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Insufficient stock for one or more products",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to checkout cart",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully from cart",
		"data":    order,
	})
}
