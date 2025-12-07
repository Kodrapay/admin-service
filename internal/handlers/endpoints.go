package handlers

import (
	"log" // Added import
	"github.com/gofiber/fiber/v2"

	"github.com/kodra-pay/admin-service/internal/services"
)

type AdminHandler struct {
	svc *services.AdminService
}

func NewAdminHandler(svc *services.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) ListPendingMerchants(c *fiber.Ctx) error {
	log.Println("AdminHandler: ListPendingMerchants called.") // Added log
	merchants, err := h.svc.ListPendingMerchants(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(merchants)
}

func (h *AdminHandler) ApproveMerchantKYC(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id") // Use c.ParamsInt
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid merchant ID")
	}
	return c.JSON(h.svc.ApproveMerchantKYC(c.Context(), id))
}

func (h *AdminHandler) RejectMerchantKYC(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id") // Use c.ParamsInt
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid merchant ID")
	}
	return c.JSON(h.svc.RejectMerchantKYC(c.Context(), id))
}

func (h *AdminHandler) EnableMerchantKYC(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id") // Use c.ParamsInt
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid merchant ID")
	}
	return c.JSON(h.svc.EnableMerchantKYC(c.Context(), id))
}

func (h *AdminHandler) Transactions(c *fiber.Ctx) error {
	return c.JSON(h.svc.Transactions(c.Context()))
}

func (h *AdminHandler) Stats(c *fiber.Ctx) error {
	return c.JSON(h.svc.Stats(c.Context()))
}

func (h *AdminHandler) ListMerchants(c *fiber.Ctx) error {
	log.Println("AdminHandler: ListMerchants called.") // Added log
	return c.JSON(h.svc.ListMerchants(c.Context()))
}

func (h *AdminHandler) ApproveMerchant(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id") // Use c.ParamsInt
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid merchant ID")
	}
	return c.JSON(h.svc.ApproveMerchant(c.Context(), id))
}

func (h *AdminHandler) SuspendMerchant(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id") // Use c.ParamsInt
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid merchant ID")
	}
	return c.JSON(h.svc.SuspendMerchant(c.Context(), id))
}

// Register registers all admin routes
func (h *AdminHandler) Register(app *fiber.App) {
	admin := app.Group("/admin")
	admin.Get("/merchants", h.ListMerchants)
	admin.Get("/merchants/pending", h.ListPendingMerchants)
	admin.Post("/merchants/:id/approve", h.ApproveMerchant)
	admin.Post("/merchants/:id/suspend", h.SuspendMerchant)
	admin.Post("/merchants/:id/kyc/approve", h.ApproveMerchantKYC)
	admin.Post("/merchants/:id/kyc/reject", h.RejectMerchantKYC)
	admin.Post("/merchants/:id/kyc/enable", h.EnableMerchantKYC)
	admin.Get("/transactions", h.Transactions)
	admin.Get("/stats", h.Stats)
}
