package services

import (
	"context"
	"time"

	"github.com/kodra-pay/admin-service/internal/repositories"
)

type AdminService struct {
	repo *repositories.AdminRepository
}

func NewAdminService(repo *repositories.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) ListMerchants(_ context.Context) []map[string]string {
	return []map[string]string{}
}

func (s *AdminService) ApproveMerchant(_ context.Context, id string) map[string]string {
	return map[string]string{"id": id, "status": "approved"}
}

func (s *AdminService) SuspendMerchant(_ context.Context, id string) map[string]string {
	return map[string]string{"id": id, "status": "suspended"}
}

func (s *AdminService) Transactions(_ context.Context) []map[string]string {
	return []map[string]string{}
}

func (s *AdminService) Stats(_ context.Context) map[string]interface{} {
	return map[string]interface{}{
		"total_revenue":    "0",
		"active_merchants": 0,
		"transactions":     0,
		"success_rate":     0.0,
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}
}
