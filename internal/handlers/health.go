package handlers

import (
	"net/http"

	"github.com/Code-byme/e-commerce/internal/database"
	"github.com/gin-gonic/gin"
)

// HealthCheck handles the health check endpoint
func HealthCheck(c *gin.Context) {
	// Check database health
	dbStatus := "ok"
	if err := database.HealthCheck(); err != nil {
		dbStatus = "error"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "E-commerce API is running",
		"version": "1.0.0",
		"database": gin.H{
			"status": dbStatus,
		},
	})
}
