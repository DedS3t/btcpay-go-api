package btcpay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type PaymentRequestStatus string

const (
	PaymentRequestPending   PaymentRequestStatus = "Pending"
	PaymentRequestCompleted                      = "Completed"
	PaymentRequestExpired                        = "Expired"
)

type PaymentRequest struct {
	PaymentRequestRequest
	Archived bool                 `json:"archived"`
	Created  string               `json:"created"`
	ID       string               `json:"id"`
	Status   PaymentRequestStatus `json:"status"`
}

// Mandatory fields are amount, currency and title.
type PaymentRequestRequest struct {
	AllowCustomPaymentAmounts bool    `json:"allowCustomPaymentAmounts,omitempty"`
	Amount                    float64 `json:"amount"`
	Currency                  string  `json:"currency"`                // ISO 4217 Currency code (BTC, EUR, USD, etc)
	CustomCSSLink             string  `json:"customCSSLink,omitempty"` // URI
	Description               string  `json:"description,omitempty"`   // HTML
	Email                     string  `json:"email,omitempty"`
	EmbeddedCSS               string  `json:"embeddedCSS,omitempty"` // CSS up to 500 bytes
	ExpiryDate                string  `json:"expiryDate,omitempty"`  // RFC3339 date (in contrast to the docs which say int64)
	Title                     string  `json:"title"`                 // required
}

func (req *PaymentRequestRequest) SetExpiryDays(days int) {
	req.ExpiryDate = time.Now().AddDate(0, 0, days).Format(time.RFC3339)
}

func (store *Store) CreatePaymentRequest(request *PaymentRequestRequest) (*PaymentRequest, error) {

	payload, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	resp, err := store.DoRequest(http.MethodPost, "payment-requests", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusUnauthorized: // 401, "Unauthorized" should be "Unauthenticated"
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("response status: %d", resp.StatusCode)
	}

	var paymentRequest = &PaymentRequest{}
	return paymentRequest, json.Unmarshal(body, paymentRequest)
}

func (store *Store) GetPaymentRequest(id string) (*PaymentRequest, error) {

	resp, err := store.DoRequest(http.MethodGet, fmt.Sprintf("payment-requests/%s", id), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		// ok
	case http.StatusUnauthorized: // 401, "Unauthorized" should be "Unauthenticated"
		return nil, ErrUnauthenticated
	case http.StatusForbidden:
		return nil, ErrUnauthorized
	case http.StatusBadRequest:
		return nil, ErrBadRequest
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("response status: %d", resp.StatusCode)
	}

	var paymentRequest = &PaymentRequest{}
	return paymentRequest, json.Unmarshal(body, paymentRequest)
}
