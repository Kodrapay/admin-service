package services

import (
	"context"

	"github.com/kodra-pay/admin-service/internal/repositories"
)

type AdminService struct {
	repo *repositories.AdminRepository
}

func NewAdminService(repo *repositories.AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

func (s *AdminService) ListMerchants(ctx context.Context) []map[string]interface{} {
	merchants, err := s.repo.ListMerchants(ctx)
	if err != nil {
		return []map[string]interface{}{}
	}
	return merchants
}

func (s *AdminService) ApproveMerchant(ctx context.Context, id string) map[string]string {
	err := s.repo.ApproveMerchant(ctx, id)
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": err.Error()}
	}
	return map[string]string{"id": id, "status": "approved"}
}

func (s *AdminService) SuspendMerchant(ctx context.Context, id string) map[string]string {
	err := s.repo.SuspendMerchant(ctx, id)
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": err.Error()}
	}
	return map[string]string{"id": id, "status": "suspended"}
}

func (s *AdminService) Transactions(ctx context.Context) []map[string]interface{} {
	transactions, err := s.repo.GetTransactions(ctx, 100)
	if err != nil {
		return []map[string]interface{}{}
	}
	return transactions
}

func (s *AdminService) Stats(ctx context.Context) map[string]interface{} {
	stats, err := s.repo.GetStats(ctx)
	if err != nil {
		return map[string]interface{}{
			"total_merchants":    0,
			"active_merchants":   0,
			"pending_kyc":        0,
			"total_transactions": 0,
			"total_volume":       0,
			"monthly_volume":     0,
			"success_rate":       0.0,
		}
	}
	return stats
}
