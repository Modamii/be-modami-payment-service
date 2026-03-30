package storage

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// UnitOfWork wraps a sqlx.Tx for transactional operations.
type UnitOfWork struct {
	db *sqlx.DB
}

// NewUnitOfWork creates a UnitOfWork backed by the given DB.
func NewUnitOfWork(db *sqlx.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// RunInTx executes fn inside a database transaction.
// If fn returns an error the transaction is rolled back, otherwise committed.
func (u *UnitOfWork) RunInTx(ctx context.Context, fn func(tx *sqlx.Tx) error) error {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
