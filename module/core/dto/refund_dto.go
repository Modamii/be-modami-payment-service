package dto

import "time"

// CreateRefundRequest is the request body for POST /refunds.
type CreateRefundRequest struct {
	PaymentTxID string `json:"payment_tx_id"`
	CreditTxID  string `json:"credit_tx_id"`
	RefundType  string `json:"refund_type" binding:"required,oneof=gateway credit manual"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	Reason      string `json:"reason"`
}

// RefundResponse represents a refund.
type RefundResponse struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	RefundType  string     `json:"refund_type"`
	Amount      int64      `json:"amount"`
	Reason      *string    `json:"reason,omitempty"`
	Status      string     `json:"status"`
	RequestedBy string     `json:"requested_by"`
	ApprovedBy  *string    `json:"approved_by,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
