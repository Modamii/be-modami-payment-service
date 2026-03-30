package dto

import "time"

// CreatePaymentRequest is the request body for POST /payments.
type CreatePaymentRequest struct {
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Method        string `json:"method" binding:"required"`
	Purpose       string `json:"purpose" binding:"required"`
	PurposeRefID  string `json:"purpose_ref_id"`
}

// PaymentResponse represents a payment transaction.
type PaymentResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	OrderRef    string     `json:"order_ref"`
	Amount      int64      `json:"amount"`
	Currency    string     `json:"currency"`
	Method      string     `json:"method"`
	Purpose     string     `json:"purpose"`
	Status      string     `json:"status"`
	PaymentURL  *string    `json:"payment_url,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PaymentStatusResponse is returned for GET /payments/:id/status.
type PaymentStatusResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// PaymentHistoryResponse is returned for GET /payments/history.
type PaymentHistoryResponse struct {
	Items  []*PaymentResponse `json:"items"`
	Total  int                `json:"total"`
	Limit  int                `json:"limit"`
	Offset int                `json:"offset"`
}
