package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Code-byme/e-commerce/internal/database"
	"github.com/Code-byme/e-commerce/internal/handlers"
	"github.com/Code-byme/e-commerce/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database connection
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Run database migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Initialize handlers
	authHandler := handlers.NewAuthHandler()
	productHandler := handlers.NewProductHandler()
	categoryHandler := handlers.NewCategoryHandler()
	orderHandler := handlers.NewOrderHandler()
	cartHandler := handlers.NewCartHandler()

	// Public routes
	r.GET("/health", handlers.HealthCheck)

	// Authentication routes
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Public product routes
	products := r.Group("/products")
	{
		products.GET("", productHandler.ListProducts)
		products.GET("/:id", productHandler.GetProduct)
		products.GET("/category/:category_id", productHandler.GetProductsByCategory)
	}

	// Public category routes
	categories := r.Group("/categories")
	{
		categories.GET("", categoryHandler.ListCategories)
		categories.GET("/:id", categoryHandler.GetCategory)
		categories.GET("/:id/with-products", categoryHandler.GetCategoryWithProductCount)
	}

	// Protected routes
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", authHandler.GetProfile)

		// Protected product routes (admin only)
		protected.POST("/products", productHandler.CreateProduct)
		protected.PUT("/products/:id", productHandler.UpdateProduct)
		protected.DELETE("/products/:id", productHandler.DeleteProduct)
		protected.PATCH("/products/:id/stock", productHandler.UpdateStock)

		// Protected category routes (admin only)
		protected.POST("/categories", categoryHandler.CreateCategory)
		protected.PUT("/categories/:id", categoryHandler.UpdateCategory)
		protected.DELETE("/categories/:id", categoryHandler.DeleteCategory)

		// Protected order routes
		protected.POST("/orders", orderHandler.CreateOrder)
		protected.GET("/orders", orderHandler.ListOrders)
		protected.GET("/orders/my", orderHandler.GetUserOrders)
		protected.GET("/orders/:id", orderHandler.GetOrder)
		protected.PUT("/orders/:id/status", orderHandler.UpdateOrderStatus)
		protected.DELETE("/orders/:id", orderHandler.CancelOrder)
		protected.GET("/orders/statistics", orderHandler.GetOrderStatistics)

		// Protected cart routes
		protected.GET("/cart", cartHandler.GetCart)
		protected.POST("/cart/items", cartHandler.AddToCart)
		protected.PUT("/cart/items/:item_id", cartHandler.UpdateCartItem)
		protected.DELETE("/cart/items/:item_id", cartHandler.RemoveFromCart)
		protected.DELETE("/cart", cartHandler.ClearCart)
		protected.POST("/cart/checkout", cartHandler.CheckoutCart)
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Println("Starting e-commerce server on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the server gracefully
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
