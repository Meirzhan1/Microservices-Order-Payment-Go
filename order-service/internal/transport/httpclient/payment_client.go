package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type PaymentClient interface {
	CreatePayment(ctx context.Context, orderID string, amount int64) (*CreatePaymentResponse, error)
}

type HTTPPaymentClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPPaymentClient(baseURL string, client *http.Client) *HTTPPaymentClient {
	return &HTTPPaymentClient{baseURL: baseURL, client: client}
}

type createPaymentRequest struct {
	OrderID string `json:"order_id"`
	Amount  int64  `json:"amount"`
}

type CreatePaymentResponse struct {
	ID            string `json:"id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"`
	DeclineReason string `json:"decline_reason"`
}

func (c *HTTPPaymentClient) CreatePayment(ctx context.Context, orderID string, amount int64) (*CreatePaymentResponse, error) {
	body, err := json.Marshal(createPaymentRequest{OrderID: orderID, Amount: amount})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/payments", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("payment service returned status %d", resp.StatusCode)
	}

	var parsed CreatePaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	return &parsed, nil
}
