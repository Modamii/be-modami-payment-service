package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
	"github.com/modami/be-payment-service/module/core/model"
)

// SubscriptionRepo implements repository.SubscriptionRepository.
type SubscriptionRepo struct {
	db *sqlx.DB
}

func NewSubscriptionRepo(db *sqlx.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

const subCols = `id, user_id, package_id, billing_cycle, price_paid, discount_code,
	credits_allocated, credits_used, status, auto_renew, start_date, end_date,
	payment_tx_id, cancelled_at, cancel_reason, created_at, updated_at`

func (r *SubscriptionRepo) Create(ctx context.Context, tx *sqlx.Tx, sub *model.Subscription) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO subscriptions (id, user_id, package_id, billing_cycle, price_paid,
		 discount_code, credits_allocated, credits_used, status, auto_renew,
		 start_date, end_date, payment_tx_id, cancelled_at, cancel_reason)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		sub.ID, sub.UserID, sub.PackageID, sub.BillingCycle, sub.PricePaid,
		sub.DiscountCode, sub.CreditsAllocated, sub.CreditsUsed, sub.Status, sub.AutoRenew,
		sub.StartDate, sub.EndDate, sub.PaymentTxID, sub.CancelledAt, sub.CancelReason)
	return err
}

func (r *SubscriptionRepo) GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.GetContext(ctx, &s,
		`SELECT `+subCols+` FROM subscriptions WHERE user_id = $1 AND status = 'active' LIMIT 1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrSubscriptionNotFound
	}
	return &s, err
}

func (r *SubscriptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error) {
	var s model.Subscription
	err := r.db.GetContext(ctx, &s,
		`SELECT `+subCols+` FROM subscriptions WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperrors.ErrNotFound
	}
	return &s, err
}

func (r *SubscriptionRepo) UpdateStatus(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE subscriptions SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *SubscriptionRepo) Update(ctx context.Context, tx *sqlx.Tx, sub *model.Subscription) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE subscriptions SET status=$1, auto_renew=$2, start_date=$3, end_date=$4,
		 payment_tx_id=$5, cancelled_at=$6, cancel_reason=$7, credits_allocated=$8, credits_used=$9
		 WHERE id=$10`,
		sub.Status, sub.AutoRenew, sub.StartDate, sub.EndDate,
		sub.PaymentTxID, sub.CancelledAt, sub.CancelReason,
		sub.CreditsAllocated, sub.CreditsUsed, sub.ID)
	return err
}

func (r *SubscriptionRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Subscription, error) {
	var subs []*model.Subscription
	err := r.db.SelectContext(ctx, &subs,
		`SELECT `+subCols+` FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`, userID)
	return subs, err
}

func (r *SubscriptionRepo) ListExpiringWithAutoRenew(ctx context.Context) ([]*model.Subscription, error) {
	var subs []*model.Subscription
	// Fetch subscriptions expiring within the next 24 hours with auto_renew=true.
	in24h := time.Now().Add(24 * time.Hour)
	err := r.db.SelectContext(ctx, &subs,
		`SELECT `+subCols+` FROM subscriptions
		 WHERE status = 'active' AND auto_renew = TRUE AND end_date <= $1`, in24h)
	return subs, err
}
