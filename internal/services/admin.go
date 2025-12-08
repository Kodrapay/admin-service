package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/kodra-pay/admin-service/internal/clients" // Import clients
	"github.com/kodra-pay/admin-service/internal/dto"     // Import dto
	"github.com/kodra-pay/admin-service/internal/repositories"
)

type AdminService struct {
	repo                 *repositories.AdminRepository
	MerchantServiceURL   string
	ComplianceServiceURL string
	TransactionClient    clients.TransactionClient // Add TransactionClient
}

func NewAdminService(repo *repositories.AdminRepository, merchantServiceURL, complianceServiceURL string, txClient clients.TransactionClient) *AdminService {
	return &AdminService{
		repo:                 repo,
		MerchantServiceURL:   merchantServiceURL,
		ComplianceServiceURL: complianceServiceURL,
		TransactionClient:    txClient,
	}
}

func (s *AdminService) ListFraudulentTransactions(ctx context.Context, limit int) (dto.TransactionListResponse, error) {
	return s.TransactionClient.ListFraudulentTransactions(ctx, limit)
}

func (s *AdminService) ListMerchants(ctx context.Context) ([]map[string]interface{}, error) {
	log.Println("AdminService: ListMerchants called.")
	merchants, err := s.repo.ListMerchants(ctx, 200)
	if err != nil {
		log.Printf("ERROR: AdminService.ListMerchants - error from repository: %v", err)
		return nil, err // Return nil slice and error
	}
	log.Printf("DEBUG: AdminService.ListMerchants - merchants retrieved from repo: %d", len(merchants))
	return merchants, nil
}

func (s *AdminService) ListPendingMerchants(ctx context.Context) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/merchants/kyc?kyc_status=pending", s.MerchantServiceURL)
	log.Printf("AdminService: Calling Merchant Service for pending KYC: %s", url)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Printf("AdminService: Failed to create request for pending KYC: %v", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("AdminService: Failed to call Merchant Service for pending KYC: %v", err)
		return nil, fmt.Errorf("failed to call merchant service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read and log response body for non-OK status
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Printf("AdminService: Merchant Service returned non-OK status for pending KYC: %d, failed to read body: %v", resp.StatusCode, readErr)
		} else {
			log.Printf("AdminService: Merchant Service returned non-OK status for pending KYC: %d, body: %s", resp.StatusCode, string(bodyBytes))
		}
		return nil, fmt.Errorf("merchant service returned non-ok status: %d", resp.StatusCode)
	}

	var merchants []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&merchants); err != nil {
		log.Printf("AdminService: Failed to decode response from Merchant Service for pending KYC: %v", err)
		return nil, fmt.Errorf("failed to decode response from merchant service: %w", err)
	}

	log.Printf("AdminService: Successfully retrieved %d pending merchants from Merchant Service", len(merchants))
	return merchants, nil
}

func (s *AdminService) ApproveMerchantKYC(ctx context.Context, id int) map[string]interface{} {
	// Call compliance service to update KYC status
	url := fmt.Sprintf("%s/kyc/update", s.ComplianceServiceURL)
	body := map[string]interface{}{
		"merchant_id":  id,
		"status":       "approved",
		"reviewer_id":  101, // Admin user ID from init-db.sql
		"review_notes": "Approved by admin",
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to create request: %s", err.Error())}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to call compliance service: %s", err.Error())}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("compliance service returned non-ok status: %d", resp.StatusCode)}
	}

	// The compliance service will automatically sync the merchant KYC status
	// Additionally, activate the merchant account after KYC approval
	activateURL := fmt.Sprintf("%s/merchants/%d/status", s.MerchantServiceURL, id)
	activateBody := map[string]string{"status": "active"}
	jsonActivateBody, _ := json.Marshal(activateBody)

	activateReq, err := http.NewRequestWithContext(ctx, "PUT", activateURL, bytes.NewBuffer(jsonActivateBody))
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to create activation request: %s", err.Error())}
	}
	activateReq.Header.Set("Content-Type", "application/json")

	activateResp, err := http.DefaultClient.Do(activateReq)
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to activate merchant: %s", err.Error())}
	}
	defer activateResp.Body.Close()

	if activateResp.StatusCode != http.StatusOK {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("merchant service activation returned non-ok status: %d", activateResp.StatusCode)}
	}

	return map[string]interface{}{"id": id, "status": "approved"}
}

func (s *AdminService) RejectMerchantKYC(ctx context.Context, id int) map[string]interface{} {
	// Call compliance service to update KYC status
	url := fmt.Sprintf("%s/kyc/update", s.ComplianceServiceURL)
	body := map[string]interface{}{
		"merchant_id":  id,
		"status":       "rejected",
		"reviewer_id":  101, // Admin user ID from init-db.sql
		"review_notes": "Rejected by admin",
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to create request: %s", err.Error())}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to call compliance service: %s", err.Error())}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("compliance service returned non-ok status: %d", resp.StatusCode)}
	}

	// The compliance service will automatically sync the merchant KYC status
	return map[string]interface{}{"id": id, "status": "rejected"}
}

func (s *AdminService) EnableMerchantKYC(ctx context.Context, id int) map[string]interface{} {
	// Update KYC status to pending to allow merchant to proceed with KYC
	url := fmt.Sprintf("%s/merchants/%d/kyc-status", s.MerchantServiceURL, id)
	body := map[string]string{"kyc_status": "pending"}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to create request: %s", err.Error())}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to call merchant service: %s", err.Error())}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("merchant service returned non-ok status: %d", resp.StatusCode)}
	}

	// Also update merchant status to pending if inactive
	statusURL := fmt.Sprintf("%s/merchants/%d/status", s.MerchantServiceURL, id)
	statusBody := map[string]string{"status": "pending"}
	jsonStatusBody, _ := json.Marshal(statusBody)

	statusReq, err := http.NewRequestWithContext(ctx, "PUT", statusURL, bytes.NewBuffer(jsonStatusBody))
	if err != nil {
		log.Printf("Warning: Failed to create status update request: %v", err)
	} else {
		statusReq.Header.Set("Content-Type", "application/json")
		statusResp, err := http.DefaultClient.Do(statusReq)
		if err != nil {
			log.Printf("Warning: Failed to call merchant service for status update: %v", err)
		} else {
			defer statusResp.Body.Close()
			if statusResp.StatusCode != http.StatusOK {
				log.Printf("Warning: Merchant service status update returned: %d", statusResp.StatusCode)
			}
		}
	}

	return map[string]interface{}{"id": id, "status": "enabled"}
}

func (s *AdminService) ApproveMerchant(ctx context.Context, id int) map[string]interface{} {
	// First, update KYC status to completed
	url := fmt.Sprintf("%s/merchants/%d/kyc-status", s.MerchantServiceURL, id)
	body := map[string]string{"kyc_status": "completed"}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to create KYC request: %s", err.Error())}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to update KYC status: %s", err.Error())}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("merchant service KYC update returned status: %d", resp.StatusCode)}
	}

	// Then, update merchant status to active
	statusURL := fmt.Sprintf("%s/merchants/%d/status", s.MerchantServiceURL, id)
	statusBody := map[string]string{"status": "active"}
	jsonStatusBody, _ := json.Marshal(statusBody)

	statusReq, err := http.NewRequestWithContext(ctx, "PUT", statusURL, bytes.NewBuffer(jsonStatusBody))
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to create status request: %s", err.Error())}
	}
	statusReq.Header.Set("Content-Type", "application/json")

	statusResp, err := http.DefaultClient.Do(statusReq)
	if err != nil {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("failed to update status: %s", err.Error())}
	}
	defer statusResp.Body.Close()

	if statusResp.StatusCode != http.StatusOK {
		return map[string]interface{}{"id": id, "status": "error", "message": fmt.Sprintf("merchant service status update returned: %d", statusResp.StatusCode)}
	}

	return map[string]interface{}{"id": id, "status": "active"}
}

func (s *AdminService) SuspendMerchant(ctx context.Context, id int) map[string]interface{} {
	log.Printf("AdminService: Attempting to suspend merchant with ID: %d", id)
	if err := s.repo.UpdateMerchantStatus(ctx, id, "suspended"); err != nil {
		log.Printf("AdminService: Failed to suspend merchant %d: %v", id, err)
		return map[string]interface{}{"id": id, "status": "error", "message": err.Error()}
	}
	log.Printf("AdminService: Successfully suspended merchant with ID: %d", id)
	return map[string]interface{}{"id": id, "status": "suspended"}
}

func (s *AdminService) Transactions(ctx context.Context) ([]map[string]interface{}, error) {
	transactions, err := s.repo.GetTransactions(ctx, 100)
	if err != nil {
		return nil, err
	}
	return transactions, nil
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
