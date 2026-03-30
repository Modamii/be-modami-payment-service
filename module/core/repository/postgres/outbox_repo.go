package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"
	"github.com/modami/be-payment-service/module/core/model"
)

// OutboxRepo implements repository.OutboxRepository.
type OutboxRepo struct {
	db *sqlx.DB
}

func NewOutboxRepo(db *sqlx.DB) *OutboxRepo {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) Create(ctx context.Context, tx *sqlx.Tx, event *model.OutboxEvent) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO outbox_events (aggregate_type, aggregate_id, event_type, payload, status)
		 VALUES ($1, $2, $3, $4, 'pending')`,
		event.AggregateType, event.AggregateID, event.EventType, event.Payload)
	return err
}

func (r *OutboxRepo) ListPending(ctx context.Context, limit int) ([]*model.OutboxEvent, error) {
	var events []*model.OutboxEvent
	err := r.db.SelectContext(ctx, &events,
		`SELECT id, aggregate_type, aggregate_id, event_type, payload, status,
		        retry_count, max_retries, published_at, last_error, created_at
		 FROM outbox_events
		 WHERE status = 'pending' AND retry_count < max_retries
		 ORDER BY id ASC
		 LIMIT $1
		 FOR UPDATE SKIP LOCKED`, limit)
	return events, err
}

func (r *OutboxRepo) MarkPublished(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = 'published', published_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *OutboxRepo) MarkFailed(ctx context.Context, id int64, errMsg string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events SET status = 'failed', last_error = $1 WHERE id = $2`, errMsg, id)
	return err
}

func (r *OutboxRepo) IncrRetry(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE outbox_events SET retry_count = retry_count + 1, last_error = NULL WHERE id = $1`, id)
	return err
}
