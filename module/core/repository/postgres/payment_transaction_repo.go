package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
	"github.com/modami/be-payment-service/module/core/model"
)

// PaymentTransactionRepo implements repository.PaymentTransactionRepository.
type PaymentTransactionRepo struct {
	db *sqlx.DB
}

func NewPaymentTransactionRepo(db *sqlx.DB) *PaymentTransactionRepo {
	return &PaymentTransactionRepo{db: db}
}

const ptCols = `id, user_id, order_ref, gateway_tx_id, amount, currency, method,
	purpose, purpose_ref_id, status, gateway_response, payment_url, ip_address,
	expires_at, paid_at, failure_reason, created_at, updated_at`

func (r *PaymentTransactionRepo) Create(ctx context.Context, tx *sqlx.Tx, pt *model.PaymentTransaction) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO payment_transactions
		 (id, user_id, order_ref, gateway_tx_id, amount, currency, method, purpose,
		  purpose_ref_id, status, gateway_response, payment_url, ip_address, expires_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		pt.ID, pt.UserID, pt.OrderRef, pt.GatewayTxID, pt.Amount, pt.Currency,
		pt.Method, pt.Purpose, pt.PurposeRefID, pt.Status, pt.GatewayResponse,
		pt.PaymentURL, pt.IPAddress, pt.ExpiresAt)
	return err
}

func (r *PaymentTransactionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.PaymentTransaction, error) {
	var pt model.PaymentTransaction
	err := r.db.GetContext(ctx, &pt,
		`SELECT `+ptCols+` FROM payment_transactions WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &pt, err
}

func (r *PaymentTransactionRepo) GetByOrderRef(ctx context.Context, orderRef string) (*model.PaymentTransaction, error) {
	var pt model.PaymentTransaction
	err := r.db.GetContext(ctx, &pt,
		`SELECT `+ptCols+` FROM payment_transactions WHERE order_ref = $1`, orderRef)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &pt, err
}

func (r *PaymentTransactionRepo) UpdateStatus(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status, gatewayTxID string, gatewayResponse []byte) error {
	var q string
	var args []interface{}
	if status == "success" {
		q = `UPDATE payment_transactions SET status=$1, gateway_tx_id=$2, gateway_response=$3, paid_at=NOW()
		     WHERE id=$4`
		args = []interface{}{status, gatewayTxID, gatewayResponse, id}
	} else {
		q = `UPDATE payment_transactions SET status=$1, gateway_tx_id=$2, gateway_response=$3 WHERE id=$4`
		args = []interface{}{status, gatewayTxID, gatewayResponse, id}
	}

	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, q, args...)
	} else {
		_, err = r.db.ExecContext(ctx, q, args...)
	}
	return err
}

func (r *PaymentTransactionRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.PaymentTransaction, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM payment_transactions WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	var rows []*model.PaymentTransaction
	err := r.db.SelectContext(ctx, &rows,
		`SELECT `+ptCols+` FROM payment_transactions WHERE user_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	return rows, total, err
}
