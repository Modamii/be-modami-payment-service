package model

import (
	"time"

	"github.com/google/uuid"
)

// Invoice represents a payment invoice.
type Invoice struct {
	ID             uuid.UUID  `db:"id"`
	InvoiceNumber  string     `db:"invoice_number"`
	UserID         uuid.UUID  `db:"user_id"`
	PaymentTxID    *uuid.UUID `db:"payment_tx_id"`
	SubscriptionID *uuid.UUID `db:"subscription_id"`
	Subtotal       int64      `db:"subtotal"`
	TaxAmount      int64      `db:"tax_amount"`
	Total          int64      `db:"total"`
	Description    *string    `db:"description"`
	Status         string     `db:"status"`
	BillingName    *string    `db:"billing_name"`
	BillingEmail   *string    `db:"billing_email"`
	TaxCode        *string    `db:"tax_code"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at"`
}
