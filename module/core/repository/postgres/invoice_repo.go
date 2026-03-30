package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
	"github.com/modami/be-payment-service/module/core/model"
)

// InvoiceRepo implements repository.InvoiceRepository.
type InvoiceRepo struct {
	db *sqlx.DB
}

func NewInvoiceRepo(db *sqlx.DB) *InvoiceRepo {
	return &InvoiceRepo{db: db}
}

const invoiceCols = `id, invoice_number, user_id, payment_tx_id, subscription_id,
	subtotal, tax_amount, total, description, status,
	billing_name, billing_email, tax_code, created_at, updated_at`

func (r *InvoiceRepo) Create(ctx context.Context, tx *sqlx.Tx, invoice *model.Invoice) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO invoices (id, invoice_number, user_id, payment_tx_id, subscription_id,
		 subtotal, tax_amount, total, description, status, billing_name, billing_email, tax_code)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		invoice.ID, invoice.InvoiceNumber, invoice.UserID, invoice.PaymentTxID, invoice.SubscriptionID,
		invoice.Subtotal, invoice.TaxAmount, invoice.Total, invoice.Description, invoice.Status,
		invoice.BillingName, invoice.BillingEmail, invoice.TaxCode)
	return err
}

func (r *InvoiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	var inv model.Invoice
	err := r.db.GetContext(ctx, &inv,
		`SELECT `+invoiceCols+` FROM invoices WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &inv, err
}

func (r *InvoiceRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Invoice, error) {
	var rows []*model.Invoice
	err := r.db.SelectContext(ctx, &rows,
		`SELECT `+invoiceCols+` FROM invoices WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	return rows, err
}

func (r *InvoiceRepo) NextInvoiceNumber(ctx context.Context) (string, error) {
	// Generate invoice number: INV-YYYYMM-XXXXXX (sequential within month).
	now := time.Now()
	prefix := fmt.Sprintf("INV-%04d%02d-", now.Year(), now.Month())

	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM invoices WHERE invoice_number LIKE $1`, prefix+"%").Scan(&count)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s%06d", prefix, count+1), nil
}
