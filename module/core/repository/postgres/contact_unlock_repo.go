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

// ContactUnlockRepo implements repository.ContactUnlockRepository.
type ContactUnlockRepo struct {
	db *sqlx.DB
}

func NewContactUnlockRepo(db *sqlx.DB) *ContactUnlockRepo {
	return &ContactUnlockRepo{db: db}
}

func (r *ContactUnlockRepo) Create(ctx context.Context, tx *sqlx.Tx, unlock *model.ContactUnlock) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO contact_unlocks (id, buyer_id, product_id, seller_id, credit_tx_id)
		 VALUES ($1, $2, $3, $4, $5)`,
		unlock.ID, unlock.BuyerID, unlock.ProductID, unlock.SellerID, unlock.CreditTxID)
	return err
}

func (r *ContactUnlockRepo) GetByBuyerAndProduct(ctx context.Context, buyerID, productID uuid.UUID) (*model.ContactUnlock, error) {
	var u model.ContactUnlock
	err := r.db.GetContext(ctx, &u,
		`SELECT id, buyer_id, product_id, seller_id, credit_tx_id, created_at
		 FROM contact_unlocks WHERE buyer_id = $1 AND product_id = $2`,
		buyerID, productID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

func (r *ContactUnlockRepo) ListByBuyerID(ctx context.Context, buyerID uuid.UUID, limit, offset int) ([]*model.ContactUnlock, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM contact_unlocks WHERE buyer_id = $1`, buyerID).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*model.ContactUnlock{}, 0, nil
	}

	var rows []*model.ContactUnlock
	err := r.db.SelectContext(ctx, &rows,
		`SELECT id, buyer_id, product_id, seller_id, credit_tx_id, created_at
		 FROM contact_unlocks WHERE buyer_id = $1
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		buyerID, limit, offset)
	if err != nil {
		return nil, 0, apperrors.ErrInternal
	}
	return rows, total, nil
}
