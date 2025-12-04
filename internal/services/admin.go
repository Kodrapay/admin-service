package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/kodra-pay/admin-service/internal/repositories"
)

type AdminService struct {
	repo               *repositories.AdminRepository
	MerchantServiceURL string
}

func NewAdminService(repo *repositories.AdminRepository, merchantServiceURL string) *AdminService {
	return &AdminService{repo: repo, MerchantServiceURL: merchantServiceURL}
}

func (s *AdminService) ListPendingMerchants(ctx context.Context) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/merchants/kyc?kyc_status=not_started,pending", s.MerchantServiceURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call merchant service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("merchant service returned non-ok status: %d", resp.StatusCode)
	}

	var merchants []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&merchants); err != nil {
		return nil, fmt.Errorf("failed to decode response from merchant service: %w", err)
	}

	return merchants, nil
}

func (s *AdminService) ApproveMerchantKYC(ctx context.Context, id string) map[string]string {
	url := fmt.Sprintf("%s/merchants/%s/kyc-status", s.MerchantServiceURL, id)
	body := map[string]string{"kyc_status": "approved"}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("failed to create request: %s", err.Error())}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("failed to call merchant service: %s", err.Error())}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("merchant service returned non-ok status: %d", resp.StatusCode)}
	}

	// Additionally, activate the merchant account after KYC approval
	activateURL := fmt.Sprintf("%s/merchants/%s/status", s.MerchantServiceURL, id)
	activateBody := map[string]string{"status": "active"}
	jsonActivateBody, _ := json.Marshal(activateBody)

	activateReq, err := http.NewRequestWithContext(ctx, "PUT", activateURL, bytes.NewBuffer(jsonActivateBody))
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("failed to create activation request: %s", err.Error())}
	}
	activateReq.Header.Set("Content-Type", "application/json")

	activateResp, err := http.DefaultClient.Do(activateReq)
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("failed to activate merchant: %s", err.Error())}
	}
	defer activateResp.Body.Close()

	if activateResp.StatusCode != http.StatusOK {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("merchant service activation returned non-ok status: %d", activateResp.StatusCode)}
	}

	return map[string]string{"id": id, "status": "approved"}
}

func (s *AdminService) RejectMerchantKYC(ctx context.Context, id string) map[string]string {
	url := fmt.Sprintf("%s/merchants/%s/kyc-status", s.MerchantServiceURL, id)
	body := map[string]string{"kyc_status": "rejected"}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("failed to create request: %s", err.Error())}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("failed to call merchant service: %s", err.Error())}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]string{"id": id, "status": "error", "message": fmt.Sprintf("merchant service returned non-ok status: %d", resp.StatusCode)}
	}

	return map[string]string{"id": id, "status": "rejected"}
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
