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

// CreditWalletRepo implements repository.CreditWalletRepository using PostgreSQL.
type CreditWalletRepo struct {
	db *sqlx.DB
}

// NewCreditWalletRepo creates a new CreditWalletRepo.
func NewCreditWalletRepo(db *sqlx.DB) *CreditWalletRepo {
	return &CreditWalletRepo{db: db}
}

func (r *CreditWalletRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*model.CreditWallet, error) {
	var w model.CreditWallet
	err := r.db.GetContext(ctx, &w,
		`SELECT user_id, balance, total_earned, total_spent, version, created_at, updated_at
		 FROM credit_wallets WHERE user_id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrWalletNotFound
	}
	return &w, err
}

func (r *CreditWalletRepo) GetByUserIDForUpdate(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (*model.CreditWallet, error) {
	var w model.CreditWallet
	err := tx.GetContext(ctx, &w,
		`SELECT user_id, balance, total_earned, total_spent, version, created_at, updated_at
		 FROM credit_wallets WHERE user_id = $1 FOR UPDATE`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrWalletNotFound
	}
	return &w, err
}

func (r *CreditWalletRepo) Create(ctx context.Context, wallet *model.CreditWallet) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO credit_wallets (user_id, balance, total_earned, total_spent, version)
		 VALUES ($1, $2, $3, $4, $5)`,
		wallet.UserID, wallet.Balance, wallet.TotalEarned, wallet.TotalSpent, wallet.Version)
	return err
}

func (r *CreditWalletRepo) UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, newBalance int, version int64) error {
	var spent, earned int
	if newBalance < 0 {
		return apperrors.ErrInsufficientCredit
	}
	// Compute delta based on direction of change — caller should set totals explicitly.
	// Here we update balance, version, and totals based on the difference.
	row := tx.QueryRowContext(ctx, `SELECT balance FROM credit_wallets WHERE user_id = $1`, userID)
	var currentBalance int
	if err := row.Scan(&currentBalance); err != nil {
		return err
	}
	diff := newBalance - currentBalance
	if diff > 0 {
		earned = diff
	} else {
		spent = -diff
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE credit_wallets
		 SET balance = $1,
		     total_earned = total_earned + $2,
		     total_spent = total_spent + $3,
		     version = version + 1
		 WHERE user_id = $4 AND version = $5`,
		newBalance, earned, spent, userID, version)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("optimistic lock conflict: wallet version mismatch")
	}
	return nil
}
