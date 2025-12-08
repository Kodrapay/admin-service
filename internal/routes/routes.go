package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/kodra-pay/admin-service/internal/clients" // Import clients
	"github.com/kodra-pay/admin-service/internal/config"
	"github.com/kodra-pay/admin-service/internal/handlers"
	"github.com/kodra-pay/admin-service/internal/repositories"
	"github.com/kodra-pay/admin-service/internal/services"
)

func Register(app *fiber.App, serviceName string, merchantServiceURL string) {
	// Health check
	health := handlers.NewHealthHandler(serviceName)
	health.Register(app)

	// Get database URL from environment
	cfg := config.Load(serviceName, "7003") // Load config here

	// Initialize repository
	repo, err := repositories.NewAdminRepository(cfg.PostgresDSN)
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v. Using stub implementation.", err)
		return
	}

	// Initialize clients
	txClient := clients.NewHTTPTransactionClient(cfg.TransactionServiceURL)

	// Initialize service
	adminService := services.NewAdminService(repo, cfg.MerchantServiceURL, cfg.ComplianceServiceURL, txClient)

	// Initialize handlers
	adminHandler := handlers.NewAdminHandler(adminService)

	// Register routes
	adminHandler.Register(app)
}
