package model

import (
	"time"

	"github.com/google/uuid"
)

// Subscription represents a user's subscription to a package.
type Subscription struct {
	ID               uuid.UUID  `db:"id"`
	UserID           uuid.UUID  `db:"user_id"`
	PackageID        uuid.UUID  `db:"package_id"`
	BillingCycle     string     `db:"billing_cycle"`
	PricePaid        int64      `db:"price_paid"`
	DiscountCode     *string    `db:"discount_code"`
	CreditsAllocated int        `db:"credits_allocated"`
	CreditsUsed      int        `db:"credits_used"`
	Status           string     `db:"status"`
	AutoRenew        bool       `db:"auto_renew"`
	StartDate        *time.Time `db:"start_date"`
	EndDate          *time.Time `db:"end_date"`
	PaymentTxID      *uuid.UUID `db:"payment_tx_id"`
	CancelledAt      *time.Time `db:"cancelled_at"`
	CancelReason     *string    `db:"cancel_reason"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}
