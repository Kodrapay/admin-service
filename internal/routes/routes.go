package routes

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/admin-service/internal/handlers"
	"github.com/kodra-pay/admin-service/internal/repositories"
	"github.com/kodra-pay/admin-service/internal/services"
)

func Register(app *fiber.App, serviceName string) {
	// Health check
	health := handlers.NewHealthHandler(serviceName)
	health.Register(app)

	// Get database URL from environment
	dbURL := os.Getenv("POSTGRES_URL")
	if dbURL == "" {
		dbURL = "postgres://kodrapay:kodrapay_password@localhost:5432/kodrapay?sslmode=disable"
	} else {
		// Add sslmode=disable if not already present
		if !strings.Contains(dbURL, "sslmode=") {
			dbURL = dbURL + "?sslmode=disable"
		}
	}

	// Initialize repository
	repo, err := repositories.NewAdminRepository(dbURL)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v. Using stub implementation.", err)
		return
	}

	// Initialize service
	adminService := services.NewAdminService(repo)

	// Initialize handlers
	adminHandler := handlers.NewAdminHandler(adminService)

	// Register routes
	adminHandler.Register(app)
}
