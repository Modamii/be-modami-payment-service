package dto

import "time"

// CreditBalanceResponse is returned for GET /credits/balance.
type CreditBalanceResponse struct {
	UserID      string `json:"user_id"`
	Balance     int    `json:"balance"`
	TotalEarned int    `json:"total_earned"`
	TotalSpent  int    `json:"total_spent"`
}

// CreditTransactionResponse represents a single credit transaction.
type CreditTransactionResponse struct {
	ID           string    `json:"id"`
	Amount       int       `json:"amount"`
	Type         string    `json:"type"`
	RefType      *string   `json:"ref_type,omitempty"`
	RefID        *string   `json:"ref_id,omitempty"`
	BalanceAfter int       `json:"balance_after"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreditTransactionListResponse is returned for GET /credits/transactions.
type CreditTransactionListResponse struct {
	Items  []*CreditTransactionResponse `json:"items"`
	Total  int                          `json:"total"`
	Limit  int                          `json:"limit"`
	Offset int                          `json:"offset"`
}

// PurchaseCreditRequest is the request body for POST /credits/purchase.
type PurchaseCreditRequest struct {
	PackageCode    string `json:"package_code" binding:"required"`
	PaymentMethod  string `json:"payment_method" binding:"required"`
	IdempotencyKey string `json:"idempotency_key"`
}

// AdminCreditAdjustRequest is used by admins to manually adjust credits.
type AdminCreditAdjustRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Amount      int    `json:"amount" binding:"required"`
	Description string `json:"description" binding:"required"`
}
