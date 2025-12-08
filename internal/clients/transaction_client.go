package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kodra-pay/admin-service/internal/dto"
)

// TransactionClient defines the interface for interacting with the Transaction Service.
type TransactionClient interface {
	ListFraudulentTransactions(ctx context.Context, limit int) (dto.TransactionListResponse, error)
	UpdateTransactionStatus(ctx context.Context, reference, status string) error
}

// HTTPTransactionClient is an HTTP implementation of the TransactionClient interface.
type HTTPTransactionClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPTransactionClient creates a new HTTPTransactionClient.
func NewHTTPTransactionClient(baseURL string) *HTTPTransactionClient {
	return &HTTPTransactionClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// ListFraudulentTransactions calls the transaction service to get a list of transactions marked as fraudulent or pending review.
func (c *HTTPTransactionClient) ListFraudulentTransactions(ctx context.Context, limit int) (dto.TransactionListResponse, error) {
	// Construct URL with query parameters for status
	queryParams := url.Values{}
	queryParams.Add("status", "pending_review") // Status for suspected fraudulent transactions
	queryParams.Add("limit", strconv.Itoa(limit))

	url := fmt.Sprintf("%s/transactions?%s", c.baseURL, queryParams.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return dto.TransactionListResponse{}, fmt.Errorf("failed to create http request for fraudulent transactions: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return dto.TransactionListResponse{}, fmt.Errorf("failed to call transaction service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return dto.TransactionListResponse{}, fmt.Errorf("transaction service returned non-ok status: %d, body: %s", resp.StatusCode, respBody)
	}

	var transactionList dto.TransactionListResponse
	if err := json.NewDecoder(resp.Body).Decode(&transactionList); err != nil {
		return dto.TransactionListResponse{}, fmt.Errorf("failed to decode transaction list response: %w", err)
	}

	return transactionList, nil
}

// UpdateTransactionStatus updates a transaction status by reference.
func (c *HTTPTransactionClient) UpdateTransactionStatus(ctx context.Context, reference, status string) error {
	url := fmt.Sprintf("%s/transactions/%s/status", c.baseURL, reference)
	body := map[string]string{"status": status}
	jsonBody, _ := json.Marshal(body)

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create http request to update transaction status: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to call transaction service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("transaction service returned status %d: %s", resp.StatusCode, respBody)
	}
	return nil
}
