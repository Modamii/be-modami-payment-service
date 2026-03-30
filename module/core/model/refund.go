package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Refund represents a refund request.
type Refund struct {
	ID               uuid.UUID       `db:"id"`
	PaymentTxID      *uuid.UUID      `db:"payment_tx_id"`
	CreditTxID       *uuid.UUID      `db:"credit_tx_id"`
	UserID           uuid.UUID       `db:"user_id"`
	RefundType       string          `db:"refund_type"`
	Amount           int64           `db:"amount"`
	Reason           *string         `db:"reason"`
	Status           string          `db:"status"`
	RequestedBy      uuid.UUID       `db:"requested_by"`
	ApprovedBy       *uuid.UUID      `db:"approved_by"`
	GatewayRefundID  *string         `db:"gateway_refund_id"`
	GatewayResponse  json.RawMessage `db:"gateway_response"`
	CreatedAt        time.Time       `db:"created_at"`
	UpdatedAt        time.Time       `db:"updated_at"`
}
