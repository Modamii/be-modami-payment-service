package model

import (
	"time"

	"github.com/google/uuid"
)

// SubscriptionEvent records every status transition of a subscription.
type SubscriptionEvent struct {
	ID             uuid.UUID  `db:"id"`
	SubscriptionID uuid.UUID  `db:"subscription_id"`
	FromStatus     *string    `db:"from_status"`
	ToStatus       string     `db:"to_status"`
	Reason         *string    `db:"reason"`
	ActorType      string     `db:"actor_type"`
	ActorID        *uuid.UUID `db:"actor_id"`
	CreatedAt      time.Time  `db:"created_at"`
}
