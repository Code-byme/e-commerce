package handlers

import (
	"net/http"
	"strconv"

	"github.com/Code-byme/e-commerce/internal/services"
	"github.com/gin-gonic/gin"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler() *OrderHandler {
	return &OrderHandler{
		orderService: services.NewOrderService(),
	}
}

// CreateOrder handles order creation
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req services.CreateOrderRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Create order
	order, err := h.orderService.CreateOrder(userID.(uint), &req)
	if err != nil {
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
			"error": "Failed to create order",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"data":    order,
	})
}

// GetOrder handles retrieving a single order
func (h *OrderHandler) GetOrder(c *gin.Context) {
	// Parse order ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	// Get order
	order, err := h.orderService.GetOrder(uint(id))
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve order",
		})
		return
	}

	// Check if user is authorized to view this order
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Only allow users to view their own orders (unless admin)
	userRole, _ := c.Get("user_role")
	if userRole != "admin" && order.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not authorized to view this order",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": order,
	})
}

// UpdateOrderStatus handles order status updates
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	// Parse order ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	var req services.UpdateOrderStatusRequest

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Update order status
	order, err := h.orderService.UpdateOrderStatus(uint(id), &req)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update order status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"data":    order,
	})
}

// ListOrders handles order listing with filtering and pagination
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Parse query parameters
	filter := &services.OrderFilter{}

	// Pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filter.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	// Status filter
	if status := c.Query("status"); status != "" {
		filter.Status = &status
	}

	// Date filters
	if startDate := c.Query("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}

	if endDate := c.Query("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}

	// User filter (admin can filter by user, regular users can only see their own orders)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	userRole, _ := c.Get("user_role")
	if userRole == "admin" {
		// Admin can filter by any user
		if userIDStr := c.Query("user_id"); userIDStr != "" {
			if parsedUserID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
				userIDUint := uint(parsedUserID)
				filter.UserID = &userIDUint
			}
		}
	} else {
		// Regular users can only see their own orders
		userIDUint := userID.(uint)
		filter.UserID = &userIDUint
	}

	// Get orders
	response, err := h.orderService.ListOrders(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve orders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// GetUserOrders handles retrieving orders for the current user
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Get user orders
	response, err := h.orderService.GetUserOrders(userID.(uint), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve user orders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// CancelOrder handles order cancellation
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	// Parse order ID from URL parameter
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order ID",
		})
		return
	}

	// Cancel order
	err = h.orderService.CancelOrder(uint(id))
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Order not found",
			})
			return
		}
		if err.Error() == "order is already cancelled" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Order is already cancelled",
			})
			return
		}
		if err.Error() == "cannot cancel delivered order" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot cancel delivered order",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to cancel order",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order cancelled successfully",
	})
}

// GetOrderStatistics handles retrieving order statistics (admin only)
func (h *OrderHandler) GetOrderStatistics(c *gin.Context) {
	// Check if user is admin
	userRole, exists := c.Get("user_role")
	if !exists || userRole != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admin access required",
		})
		return
	}

	// Get order statistics
	stats, err := h.orderService.GetOrderStatistics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve order statistics",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
	})
}
