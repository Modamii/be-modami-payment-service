package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
	"github.com/modami/be-payment-service/module/core/model"
)

// CreditTransactionRepo implements repository.CreditTransactionRepository.
type CreditTransactionRepo struct {
	db *sqlx.DB
}

func NewCreditTransactionRepo(db *sqlx.DB) *CreditTransactionRepo {
	return &CreditTransactionRepo{db: db}
}

func (r *CreditTransactionRepo) Create(ctx context.Context, tx *sqlx.Tx, t *model.CreditTransaction) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO credit_transactions
		 (id, user_id, amount, type, ref_type, ref_id, balance_after, description, idempotency_key)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		t.ID, t.UserID, t.Amount, t.Type, t.RefType, t.RefID,
		t.BalanceAfter, t.Description, t.IdempotencyKey)
	return err
}

func (r *CreditTransactionRepo) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, txType *string) ([]*model.CreditTransaction, int, error) {
	var total int
	countQ := `SELECT COUNT(*) FROM credit_transactions WHERE user_id = $1`
	args := []interface{}{userID}

	if txType != nil {
		countQ += " AND type = $2"
		args = append(args, *txType)
	}
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	q := `SELECT id, user_id, amount, type, ref_type, ref_id, balance_after, description, idempotency_key, created_at
	      FROM credit_transactions WHERE user_id = $1`
	qArgs := []interface{}{userID}
	if txType != nil {
		q += " AND type = $2"
		qArgs = append(qArgs, *txType)
		q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $3 OFFSET $4")
		qArgs = append(qArgs, limit, offset)
	} else {
		q += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $2 OFFSET $3")
		qArgs = append(qArgs, limit, offset)
	}

	var rows []*model.CreditTransaction
	if err := r.db.SelectContext(ctx, &rows, q, qArgs...); err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

func (r *CreditTransactionRepo) GetByIdempotencyKey(ctx context.Context, key string) (*model.CreditTransaction, error) {
	var t model.CreditTransaction
	err := r.db.GetContext(ctx, &t,
		`SELECT id, user_id, amount, type, ref_type, ref_id, balance_after, description, idempotency_key, created_at
		 FROM credit_transactions WHERE idempotency_key = $1`, key)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &t, err
}
