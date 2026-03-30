package model

import (
	"encoding/json"
	"time"
)

// OutboxEvent is a transactional outbox entry for reliable Kafka publishing.
type OutboxEvent struct {
	ID            int64           `db:"id"`
	AggregateType string          `db:"aggregate_type"`
	AggregateID   string          `db:"aggregate_id"`
	EventType     string          `db:"event_type"`
	Payload       json.RawMessage `db:"payload"`
	Status        string          `db:"status"`
	RetryCount    int             `db:"retry_count"`
	MaxRetries    int             `db:"max_retries"`
	PublishedAt   *time.Time      `db:"published_at"`
	LastError     *string         `db:"last_error"`
	CreatedAt     time.Time       `db:"created_at"`
}
