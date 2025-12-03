package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kodra-pay/admin-service/internal/config"
	"github.com/kodra-pay/admin-service/internal/handlers"
	"github.com/kodra-pay/admin-service/internal/repositories"
	"github.com/kodra-pay/admin-service/internal/services"
)

func Register(app *fiber.App, cfg config.Config, repo *repositories.AdminRepository) {
	health := handlers.NewHealthHandler(cfg.ServiceName)
	health.Register(app)

	svc := services.NewAdminService(repo)
	h := handlers.NewAdminHandler(svc)
	api := app.Group("/admin")
	api.Get("/merchants", h.ListMerchants)
	api.Post("/merchants/:id/approve", h.ApproveMerchant)
	api.Post("/merchants/:id/suspend", h.SuspendMerchant)
	api.Get("/transactions", h.Transactions)
	api.Get("/stats", h.Stats)
}
