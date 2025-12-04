package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/kodra-pay/admin-service/internal/services"
)

type AdminHandler struct {
	svc *services.AdminService
}

func NewAdminHandler(svc *services.AdminService) *AdminHandler { return &AdminHandler{svc: svc} }

func (h *AdminHandler) ListMerchants(c *fiber.Ctx) error {
	return c.JSON(h.svc.ListMerchants(c.Context()))
}

func (h *AdminHandler) ApproveMerchant(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(h.svc.ApproveMerchant(c.Context(), id))
}

func (h *AdminHandler) SuspendMerchant(c *fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(h.svc.SuspendMerchant(c.Context(), id))
}

func (h *AdminHandler) Transactions(c *fiber.Ctx) error {
	return c.JSON(h.svc.Transactions(c.Context()))
}

func (h *AdminHandler) Stats(c *fiber.Ctx) error {
	return c.JSON(h.svc.Stats(c.Context()))
}

// Register registers all admin routes
func (h *AdminHandler) Register(app *fiber.App) {
	admin := app.Group("/admin")
	admin.Get("/merchants", h.ListMerchants)
	admin.Post("/merchants/:id/approve", h.ApproveMerchant)
	admin.Post("/merchants/:id/suspend", h.SuspendMerchant)
	admin.Get("/transactions", h.Transactions)
	admin.Get("/stats", h.Stats)
}
