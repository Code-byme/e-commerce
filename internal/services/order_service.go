package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Code-byme/e-commerce/internal/database"
	"github.com/Code-byme/e-commerce/internal/models"
)

// OrderService handles order operations
type OrderService struct {
	db *sql.DB
}

// NewOrderService creates a new order service
func NewOrderService() *OrderService {
	return &OrderService{
		db: database.GetDB(),
	}
}

// CreateOrderRequest represents the request to create an order
type CreateOrderRequest struct {
	ShippingAddress string                   `json:"shipping_address" binding:"required"`
	PaymentMethod   string                   `json:"payment_method" binding:"required"`
	Items           []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

// CreateOrderItemRequest represents an item in the order creation request
type CreateOrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

// UpdateOrderStatusRequest represents the request to update order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending confirmed shipped delivered cancelled"`
}

// OrderFilter represents order filtering options
type OrderFilter struct {
	UserID    *uint   `json:"user_id"`
	Status    *string `json:"status"`
	StartDate *string `json:"start_date"`
	EndDate   *string `json:"end_date"`
	Page      int     `json:"page"`
	Limit     int     `json:"limit"`
}

// OrderListResponse represents the paginated order list response
type OrderListResponse struct {
	Orders []models.Order `json:"orders"`
	Total  int            `json:"total"`
	Page   int            `json:"page"`
	Limit  int            `json:"limit"`
	Pages  int            `json:"pages"`
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(userID uint, req *CreateOrderRequest) (*models.Order, error) {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Calculate total amount and validate products
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range req.Items {
		// Get product details
		var product models.Product
		err := tx.QueryRow(
			"SELECT id, name, price, stock FROM products WHERE id = $1 AND is_active = true",
			item.ProductID,
		).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("product with ID %d not found or inactive", item.ProductID)
			}
			return nil, fmt.Errorf("failed to get product: %w", err)
		}

		// Check stock availability
		if product.Stock < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)",
				product.Name, product.Stock, item.Quantity)
		}

		// Calculate item total
		itemTotal := product.Price * float64(item.Quantity)
		totalAmount += itemTotal

		// Create order item
		orderItem := models.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			Price:     product.Price,
		}
		orderItems = append(orderItems, orderItem)
	}

	// Create order
	var order models.Order
	orderQuery := `
		INSERT INTO orders (user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, user_id, status, total_amount, shipping_address, payment_method, created_at, updated_at
	`

	err = tx.QueryRow(
		orderQuery,
		userID, "pending", totalAmount, req.ShippingAddress, req.PaymentMethod,
	).Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
		&order.ShippingAddress, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Create order items
	for _, item := range orderItems {
		itemQuery := `
			INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
			VALUES ($1, $2, $3, $4, NOW(), NOW())
		`
		_, err = tx.Exec(itemQuery, order.ID, item.ProductID, item.Quantity, item.Price)
		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		// Update product stock
		_, err = tx.Exec("UPDATE products SET stock = stock - $1, updated_at = NOW() WHERE id = $2",
			item.Quantity, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to update product stock: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Get order with items
	return s.GetOrder(order.ID)
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(id uint) (*models.Order, error) {
	var order models.Order
	orderQuery := `
		SELECT o.id, o.user_id, o.status, o.total_amount, o.shipping_address, o.payment_method, o.created_at, o.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.created_at, u.updated_at
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`

	err := s.db.QueryRow(orderQuery, id).Scan(
		&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
		&order.ShippingAddress, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt,
		&order.User.ID, &order.User.Email, &order.User.FirstName, &order.User.LastName,
		&order.User.Role, &order.User.CreatedAt, &order.User.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Get order items
	itemsQuery := `
		SELECT oi.id, oi.order_id, oi.product_id, oi.quantity, oi.price, oi.created_at, oi.updated_at,
		       p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, p.is_active, p.created_at, p.updated_at
		FROM order_items oi
		JOIN products p ON oi.product_id = p.id
		WHERE oi.order_id = $1
	`

	rows, err := s.db.Query(itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item models.OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.Quantity, &item.Price,
			&item.CreatedAt, &item.UpdatedAt,
			&item.Product.ID, &item.Product.Name, &item.Product.Description, &item.Product.Price,
			&item.Product.Stock, &item.Product.CategoryID, &item.Product.ImageURL, &item.Product.IsActive,
			&item.Product.CreatedAt, &item.Product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		order.OrderItems = append(order.OrderItems, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return &order, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(id uint, req *UpdateOrderStatusRequest) (*models.Order, error) {
	// Check if order exists
	_, err := s.GetOrder(id)
	if err != nil {
		return nil, err
	}

	// Update order status
	_, err = s.db.Exec(
		"UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2",
		req.Status, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	// Get updated order
	return s.GetOrder(id)
}

// ListOrders retrieves a paginated list of orders with filtering
func (s *OrderService) ListOrders(filter *OrderFilter) (*OrderListResponse, error) {
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

	if filter.UserID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("o.user_id = $%d", argIndex))
		args = append(args, *filter.UserID)
		argIndex++
	}

	if filter.Status != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("o.status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.StartDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("o.created_at >= $%d", argIndex))
		args = append(args, *filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("o.created_at <= $%d", argIndex))
		args = append(args, *filter.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + whereConditions[0]
		for i := 1; i < len(whereConditions); i++ {
			whereClause += " AND " + whereConditions[i]
		}
	}

	// Count total orders
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM orders o %s", whereClause)
	var total int
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count orders: %w", err)
	}

	// Calculate pagination
	offset := (filter.Page - 1) * filter.Limit
	pages := (total + filter.Limit - 1) / filter.Limit

	// Get orders
	query := fmt.Sprintf(`
		SELECT o.id, o.user_id, o.status, o.total_amount, o.shipping_address, o.payment_method, o.created_at, o.updated_at,
		       u.id, u.email, u.first_name, u.last_name, u.role, u.created_at, u.updated_at
		FROM orders o
		JOIN users u ON o.user_id = u.id
		%s
		ORDER BY o.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.Limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID, &order.UserID, &order.Status, &order.TotalAmount,
			&order.ShippingAddress, &order.PaymentMethod, &order.CreatedAt, &order.UpdatedAt,
			&order.User.ID, &order.User.Email, &order.User.FirstName, &order.User.LastName,
			&order.User.Role, &order.User.CreatedAt, &order.User.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return &OrderListResponse{
		Orders: orders,
		Total:  total,
		Page:   filter.Page,
		Limit:  filter.Limit,
		Pages:  pages,
	}, nil
}

// GetUserOrders retrieves orders for a specific user
func (s *OrderService) GetUserOrders(userID uint, page, limit int) (*OrderListResponse, error) {
	filter := &OrderFilter{
		UserID: &userID,
		Page:   page,
		Limit:  limit,
	}
	return s.ListOrders(filter)
}

// CancelOrder cancels an order and restores product stock
func (s *OrderService) CancelOrder(id uint) error {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Check if order exists and can be cancelled
	var order models.Order
	err = tx.QueryRow(
		"SELECT id, status FROM orders WHERE id = $1",
		id,
	).Scan(&order.ID, &order.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("order not found")
		}
		return fmt.Errorf("database error: %w", err)
	}

	if order.Status == "cancelled" {
		return errors.New("order is already cancelled")
	}

	if order.Status == "delivered" {
		return errors.New("cannot cancel delivered order")
	}

	// Update order status to cancelled
	_, err = tx.Exec("UPDATE orders SET status = 'cancelled', updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	// Restore product stock
	_, err = tx.Exec(`
		UPDATE products 
		SET stock = stock + oi.quantity, updated_at = NOW()
		FROM order_items oi
		WHERE oi.order_id = $1 AND oi.product_id = products.id
	`, id)
	if err != nil {
		return fmt.Errorf("failed to restore product stock: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetOrderStatistics retrieves order statistics
func (s *OrderService) GetOrderStatistics() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total orders
	var totalOrders int
	err := s.db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&totalOrders)
	if err != nil {
		return nil, fmt.Errorf("failed to get total orders: %w", err)
	}
	stats["total_orders"] = totalOrders

	// Total revenue
	var totalRevenue float64
	err = s.db.QueryRow("SELECT COALESCE(SUM(total_amount), 0) FROM orders WHERE status != 'cancelled'").Scan(&totalRevenue)
	if err != nil {
		return nil, fmt.Errorf("failed to get total revenue: %w", err)
	}
	stats["total_revenue"] = totalRevenue

	// Orders by status
	statusQuery := `
		SELECT status, COUNT(*) as count
		FROM orders
		GROUP BY status
	`
	rows, err := s.db.Query(statusQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders by status: %w", err)
	}
	defer rows.Close()

	ordersByStatus := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		err := rows.Scan(&status, &count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}
		ordersByStatus[status] = count
	}
	stats["orders_by_status"] = ordersByStatus

	// Recent orders (last 7 days)
	var recentOrders int
	err = s.db.QueryRow("SELECT COUNT(*) FROM orders WHERE created_at >= NOW() - INTERVAL '7 days'").Scan(&recentOrders)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent orders: %w", err)
	}
	stats["recent_orders"] = recentOrders

	return stats, nil
}
