package model

import (
	"time"

	"github.com/google/uuid"
)

// ContactUnlock records that a buyer has unlocked contact info for a product.
type ContactUnlock struct {
	ID          uuid.UUID  `db:"id"`
	BuyerID     uuid.UUID  `db:"buyer_id"`
	ProductID   uuid.UUID  `db:"product_id"`
	SellerID    uuid.UUID  `db:"seller_id"`
	CreditTxID  *uuid.UUID `db:"credit_tx_id"`
	CreatedAt   time.Time  `db:"created_at"`
}
