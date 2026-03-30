package dto

import "time"

// InvoiceResponse represents an invoice.
type InvoiceResponse struct {
	ID            string     `json:"id"`
	InvoiceNumber string     `json:"invoice_number"`
	UserID        string     `json:"user_id"`
	Subtotal      int64      `json:"subtotal"`
	TaxAmount     int64      `json:"tax_amount"`
	Total         int64      `json:"total"`
	Description   *string    `json:"description,omitempty"`
	Status        string     `json:"status"`
	BillingName   *string    `json:"billing_name,omitempty"`
	BillingEmail  *string    `json:"billing_email,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
