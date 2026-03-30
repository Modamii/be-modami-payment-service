package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/modami/be-payment-service/module/core/model"
)

// CreditWalletRepository defines persistence operations for credit wallets.
type CreditWalletRepository interface {
	GetByUserID(ctx context.Context, userID uuid.UUID) (*model.CreditWallet, error)
	GetByUserIDForUpdate(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID) (*model.CreditWallet, error)
	Create(ctx context.Context, wallet *model.CreditWallet) error
	UpdateBalance(ctx context.Context, tx *sqlx.Tx, userID uuid.UUID, newBalance int, version int64) error
}

// CreditTransactionRepository defines persistence operations for credit transactions.
type CreditTransactionRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, t *model.CreditTransaction) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, txType *string) ([]*model.CreditTransaction, int, error)
	GetByIdempotencyKey(ctx context.Context, key string) (*model.CreditTransaction, error)
}

// ContactUnlockRepository defines persistence operations for contact unlocks.
type ContactUnlockRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, unlock *model.ContactUnlock) error
	GetByBuyerAndProduct(ctx context.Context, buyerID, productID uuid.UUID) (*model.ContactUnlock, error)
	ListByBuyerID(ctx context.Context, buyerID uuid.UUID, limit, offset int) ([]*model.ContactUnlock, int, error)
}

// PackageRepository defines persistence operations for subscription packages.
type PackageRepository interface {
	List(ctx context.Context) ([]*model.Package, error)
	GetByCode(ctx context.Context, code string) (*model.Package, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Package, error)
	Create(ctx context.Context, pkg *model.Package) error
	Update(ctx context.Context, pkg *model.Package) error
}

// SubscriptionRepository defines persistence operations for subscriptions.
type SubscriptionRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, sub *model.Subscription) error
	GetActiveByUserID(ctx context.Context, userID uuid.UUID) (*model.Subscription, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Subscription, error)
	UpdateStatus(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status string) error
	Update(ctx context.Context, tx *sqlx.Tx, sub *model.Subscription) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Subscription, error)
	ListExpiringWithAutoRenew(ctx context.Context) ([]*model.Subscription, error)
}

// SubscriptionEventRepository defines persistence operations for subscription events.
type SubscriptionEventRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, event *model.SubscriptionEvent) error
	ListBySubscriptionID(ctx context.Context, subID uuid.UUID) ([]*model.SubscriptionEvent, error)
}

// PaymentTransactionRepository defines persistence operations for payment transactions.
type PaymentTransactionRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, pt *model.PaymentTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.PaymentTransaction, error)
	GetByOrderRef(ctx context.Context, orderRef string) (*model.PaymentTransaction, error)
	UpdateStatus(ctx context.Context, tx *sqlx.Tx, id uuid.UUID, status, gatewayTxID string, gatewayResponse []byte) error
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.PaymentTransaction, int, error)
}

// RefundRepository defines persistence operations for refunds.
type RefundRepository interface {
	Create(ctx context.Context, refund *model.Refund) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Refund, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Refund, error)
	ListAll(ctx context.Context, limit, offset int) ([]*model.Refund, int, error)
}

// InvoiceRepository defines persistence operations for invoices.
type InvoiceRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, invoice *model.Invoice) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Invoice, error)
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]*model.Invoice, error)
	NextInvoiceNumber(ctx context.Context) (string, error)
}

// OutboxRepository defines persistence operations for outbox events.
type OutboxRepository interface {
	Create(ctx context.Context, tx *sqlx.Tx, event *model.OutboxEvent) error
	ListPending(ctx context.Context, limit int) ([]*model.OutboxEvent, error)
	MarkPublished(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, errMsg string) error
	IncrRetry(ctx context.Context, id int64) error
}
