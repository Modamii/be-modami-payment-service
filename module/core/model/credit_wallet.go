package model

import (
	"time"

	"github.com/google/uuid"
)

// CreditWallet represents a user's credit balance.
type CreditWallet struct {
	UserID      uuid.UUID `db:"user_id"`
	Balance     int       `db:"balance"`
	TotalEarned int       `db:"total_earned"`
	TotalSpent  int       `db:"total_spent"`
	Version     int64     `db:"version"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
