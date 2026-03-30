package model

import (
	"time"

	"github.com/google/uuid"
)

// CreditTransaction records every credit movement.
type CreditTransaction struct {
	ID             uuid.UUID  `db:"id"`
	UserID         uuid.UUID  `db:"user_id"`
	Amount         int        `db:"amount"`
	Type           string     `db:"type"`
	RefType        *string    `db:"ref_type"`
	RefID          *uuid.UUID `db:"ref_id"`
	BalanceAfter   int        `db:"balance_after"`
	Description    string     `db:"description"`
	IdempotencyKey *string    `db:"idempotency_key"`
	CreatedAt      time.Time  `db:"created_at"`
}
