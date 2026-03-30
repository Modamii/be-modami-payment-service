package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/modami/be-payment-service/module/core/model"
)

// SubscriptionEventRepo implements repository.SubscriptionEventRepository.
type SubscriptionEventRepo struct {
	db *sqlx.DB
}

func NewSubscriptionEventRepo(db *sqlx.DB) *SubscriptionEventRepo {
	return &SubscriptionEventRepo{db: db}
}

func (r *SubscriptionEventRepo) Create(ctx context.Context, tx *sqlx.Tx, event *model.SubscriptionEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO subscription_events (id, subscription_id, from_status, to_status, reason, actor_type, actor_id)
		 VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		event.ID, event.SubscriptionID, event.FromStatus, event.ToStatus,
		event.Reason, event.ActorType, event.ActorID)
	return err
}

func (r *SubscriptionEventRepo) ListBySubscriptionID(ctx context.Context, subID uuid.UUID) ([]*model.SubscriptionEvent, error) {
	var events []*model.SubscriptionEvent
	err := r.db.SelectContext(ctx, &events,
		`SELECT id, subscription_id, from_status, to_status, reason, actor_type, actor_id, created_at
		 FROM subscription_events WHERE subscription_id = $1 ORDER BY created_at ASC`, subID)
	return events, err
}
