package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentTransaction represents a payment attempt via a payment gateway.
type PaymentTransaction struct {
	ID              uuid.UUID       `db:"id"`
	UserID          uuid.UUID       `db:"user_id"`
	OrderRef        string          `db:"order_ref"`
	GatewayTxID     *string         `db:"gateway_tx_id"`
	Amount          int64           `db:"amount"`
	Currency        string          `db:"currency"`
	Method          string          `db:"method"`
	Purpose         string          `db:"purpose"`
	PurposeRefID    *uuid.UUID      `db:"purpose_ref_id"`
	Status          string          `db:"status"`
	GatewayResponse json.RawMessage `db:"gateway_response"`
	PaymentURL      *string         `db:"payment_url"`
	IPAddress       *string         `db:"ip_address"`
	ExpiresAt       *time.Time      `db:"expires_at"`
	PaidAt          *time.Time      `db:"paid_at"`
	FailureReason   *string         `db:"failure_reason"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
}
