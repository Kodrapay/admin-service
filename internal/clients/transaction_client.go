package clients

import (
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
