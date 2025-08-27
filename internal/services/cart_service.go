package services

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Code-byme/e-commerce/internal/database"
	"github.com/Code-byme/e-commerce/internal/models"
)

// CartService handles cart operations
type CartService struct {
	db *sql.DB
}

// NewCartService creates a new cart service
func NewCartService() *CartService {
	return &CartService{
		db: database.GetDB(),
	}
}

// AddToCartRequest represents the request to add an item to cart
type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

// UpdateCartItemRequest represents the request to update a cart item
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

// GetOrCreateCart gets the user's cart or creates a new one
func (s *CartService) GetOrCreateCart(userID uint) (*models.Cart, error) {
	// Try to get existing cart
	var cart models.Cart
	err := s.db.QueryRow(
		"SELECT id, user_id, created_at, updated_at FROM carts WHERE user_id = $1",
		userID,
	).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create new cart
			err = s.db.QueryRow(
				"INSERT INTO carts (user_id, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id, user_id, created_at, updated_at",
				userID,
			).Scan(&cart.ID, &cart.UserID, &cart.CreatedAt, &cart.UpdatedAt)

			if err != nil {
				return nil, fmt.Errorf("failed to create cart: %w", err)
			}
		} else {
			return nil, fmt.Errorf("database error: %w", err)
		}
	}

	return &cart, nil
}

// AddToCart adds an item to the user's cart
func (s *CartService) AddToCart(userID uint, req *AddToCartRequest) (*models.CartResponse, error) {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get or create cart
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Validate product exists and is active
	var product models.Product
	err = tx.QueryRow(
		"SELECT id, name, price, stock FROM products WHERE id = $1 AND is_active = true",
		req.ProductID,
	).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found or inactive")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Check if item already exists in cart
	var existingItem models.CartItem
	err = tx.QueryRow(
		"SELECT id, quantity FROM cart_items WHERE cart_id = $1 AND product_id = $2",
		cart.ID, req.ProductID,
	).Scan(&existingItem.ID, &existingItem.Quantity)

	if err != nil {
		if err == sql.ErrNoRows {
			// Item doesn't exist, check stock availability
			if req.Quantity > product.Stock {
				return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)",
					product.Name, product.Stock, req.Quantity)
			}

			// Add new item to cart
			_, err = tx.Exec(
				"INSERT INTO cart_items (cart_id, product_id, quantity, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())",
				cart.ID, req.ProductID, req.Quantity,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to add item to cart: %w", err)
			}
		} else {
			return nil, fmt.Errorf("database error: %w", err)
		}
	} else {
		// Item exists, update quantity
		newQuantity := existingItem.Quantity + req.Quantity
		if newQuantity > product.Stock {
			return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)",
				product.Name, product.Stock, newQuantity)
		}

		_, err = tx.Exec(
			"UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE id = $2",
			newQuantity, existingItem.ID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
	}

	// Update cart timestamp
	_, err = tx.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated cart
	return s.GetCart(userID)
}

// GetCart retrieves the user's cart with items and calculated totals
func (s *CartService) GetCart(userID uint) (*models.CartResponse, error) {
	// Get cart
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	// Get cart items with product details
	itemsQuery := `
		SELECT ci.id, ci.cart_id, ci.product_id, ci.quantity, ci.created_at, ci.updated_at,
		       p.id, p.name, p.description, p.price, p.stock, p.category_id, p.image_url, p.is_active, p.created_at, p.updated_at
		FROM cart_items ci
		JOIN products p ON ci.product_id = p.id
		WHERE ci.cart_id = $1
		ORDER BY ci.created_at ASC
	`

	rows, err := s.db.Query(itemsQuery, cart.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cart items: %w", err)
	}
	defer rows.Close()

	var cartItems []models.CartItem
	var totalItems int
	var totalAmount float64

	for rows.Next() {
		var item models.CartItem
		err := rows.Scan(
			&item.ID, &item.CartID, &item.ProductID, &item.Quantity,
			&item.CreatedAt, &item.UpdatedAt,
			&item.Product.ID, &item.Product.Name, &item.Product.Description, &item.Product.Price,
			&item.Product.Stock, &item.Product.CategoryID, &item.Product.ImageURL, &item.Product.IsActive,
			&item.Product.CreatedAt, &item.Product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cart item: %w", err)
		}

		cartItems = append(cartItems, item)
		totalItems += item.Quantity
		totalAmount += item.Product.Price * float64(item.Quantity)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cart items: %w", err)
	}

	return &models.CartResponse{
		ID:          cart.ID,
		UserID:      cart.UserID,
		CartItems:   cartItems,
		TotalItems:  totalItems,
		TotalAmount: totalAmount,
		CreatedAt:   cart.CreatedAt,
		UpdatedAt:   cart.UpdatedAt,
	}, nil
}

// UpdateCartItem updates the quantity of a cart item
func (s *CartService) UpdateCartItem(userID uint, itemID uint, req *UpdateCartItemRequest) (*models.CartResponse, error) {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get cart item and verify ownership
	var cartItem models.CartItem
	var cartID uint
	err = tx.QueryRow(`
		SELECT ci.id, ci.cart_id, ci.product_id, ci.quantity
		FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		WHERE ci.id = $1 AND c.user_id = $2
	`, itemID, userID).Scan(&cartItem.ID, &cartID, &cartItem.ProductID, &cartItem.Quantity)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("cart item not found or not owned by user")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check product stock
	var product models.Product
	err = tx.QueryRow(
		"SELECT id, name, price, stock FROM products WHERE id = $1 AND is_active = true",
		cartItem.ProductID,
	).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found or inactive")
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	// Check stock availability
	if req.Quantity > product.Stock {
		return nil, fmt.Errorf("insufficient stock for product %s (available: %d, requested: %d)",
			product.Name, product.Stock, req.Quantity)
	}

	// Update cart item
	_, err = tx.Exec(
		"UPDATE cart_items SET quantity = $1, updated_at = NOW() WHERE id = $2",
		req.Quantity, itemID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}

	// Update cart timestamp
	_, err = tx.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated cart
	return s.GetCart(userID)
}

// RemoveFromCart removes an item from the cart
func (s *CartService) RemoveFromCart(userID uint, itemID uint) (*models.CartResponse, error) {
	// Start a transaction
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get cart item and verify ownership
	var cartID uint
	err = tx.QueryRow(`
		SELECT ci.cart_id
		FROM cart_items ci
		JOIN carts c ON ci.cart_id = c.id
		WHERE ci.id = $1 AND c.user_id = $2
	`, itemID, userID).Scan(&cartID)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("cart item not found or not owned by user")
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Remove cart item
	_, err = tx.Exec("DELETE FROM cart_items WHERE id = $1", itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to remove cart item: %w", err)
	}

	// Update cart timestamp
	_, err = tx.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cartID)
	if err != nil {
		return nil, fmt.Errorf("failed to update cart: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return updated cart
	return s.GetCart(userID)
}

// ClearCart removes all items from the user's cart
func (s *CartService) ClearCart(userID uint) error {
	// Get cart
	cart, err := s.GetOrCreateCart(userID)
	if err != nil {
		return err
	}

	// Remove all cart items
	_, err = s.db.Exec("DELETE FROM cart_items WHERE cart_id = $1", cart.ID)
	if err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	// Update cart timestamp
	_, err = s.db.Exec("UPDATE carts SET updated_at = NOW() WHERE id = $1", cart.ID)
	if err != nil {
		return fmt.Errorf("failed to update cart: %w", err)
	}

	return nil
}

// CheckoutCart converts cart items to order items and clears the cart
func (s *CartService) CheckoutCart(userID uint, shippingAddress, paymentMethod string) (*models.Order, error) {
	// Get cart with items
	cart, err := s.GetCart(userID)
	if err != nil {
		return nil, err
	}

	// Check if cart is empty
	if len(cart.CartItems) == 0 {
		return nil, errors.New("cart is empty")
	}

	// Convert cart items to order items
	var orderItems []CreateOrderItemRequest
	for _, item := range cart.CartItems {
		orderItems = append(orderItems, CreateOrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	// Create order request
	orderReq := &CreateOrderRequest{
		ShippingAddress: shippingAddress,
		PaymentMethod:   paymentMethod,
		Items:           orderItems,
	}

	// Create order using order service
	orderService := NewOrderService()
	order, err := orderService.CreateOrder(userID, orderReq)
	if err != nil {
		return nil, err
	}

	// Clear the cart after successful order creation
	err = s.ClearCart(userID)
	if err != nil {
		// Log error but don't fail the checkout
		fmt.Printf("Warning: Failed to clear cart after checkout: %v\n", err)
	}

	return order, nil
}
