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

// RefundRepo implements repository.RefundRepository.
type RefundRepo struct {
	db *sqlx.DB
}

func NewRefundRepo(db *sqlx.DB) *RefundRepo {
	return &RefundRepo{db: db}
}

const refundCols = `id, payment_tx_id, credit_tx_id, user_id, refund_type, amount, reason,
	status, requested_by, approved_by, gateway_refund_id, gateway_response, created_at, updated_at`

func (r *RefundRepo) Create(ctx context.Context, refund *model.Refund) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO refunds (id, payment_tx_id, credit_tx_id, user_id, refund_type, amount,
		 reason, status, requested_by, approved_by)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		refund.ID, refund.PaymentTxID, refund.CreditTxID, refund.UserID, refund.RefundType,
		refund.Amount, refund.Reason, refund.Status, refund.RequestedBy, refund.ApprovedBy)
	return err
}

func (r *RefundRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Refund, error) {
	var ref model.Refund
	err := r.db.GetContext(ctx, &ref,
		`SELECT `+refundCols+` FROM refunds WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &ref, err
}

func (r *RefundRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refunds SET status=$1, approved_by=$2 WHERE id=$3`,
		status, approvedBy, id)
	return err
}

func (r *RefundRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Refund, error) {
	var rows []*model.Refund
	err := r.db.SelectContext(ctx, &rows,
		`SELECT `+refundCols+` FROM refunds WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	return rows, err
}

func (r *RefundRepo) ListAll(ctx context.Context, limit, offset int) ([]*model.Refund, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refunds`).Scan(&total); err != nil {
		return nil, 0, err
	}
	var rows []*model.Refund
	err := r.db.SelectContext(ctx, &rows,
		`SELECT `+refundCols+` FROM refunds ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	return rows, total, err
}
