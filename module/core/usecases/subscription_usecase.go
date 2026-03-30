package usecases

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/modami/be-payment-service/module/core/business"
	"github.com/modami/be-payment-service/module/core/domain"
	"github.com/modami/be-payment-service/module/core/model"
	"github.com/modami/be-payment-service/module/core/repository"
	"github.com/modami/be-payment-service/module/core/storage"
	"github.com/modami/be-payment-service/pkg/statemachine"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// SubscriptionUsecase handles subscription business logic.
type SubscriptionUsecase struct {
	subRepo     repository.SubscriptionRepository
	subEvtRepo  repository.SubscriptionEventRepository
	pkgRepo     repository.PackageRepository
	outboxRepo  repository.OutboxRepository
	creditUC    *CreditUsecase
	paymentUC   *PaymentUsecase
	uow         *storage.UnitOfWork
}

func NewSubscriptionUsecase(
	subRepo repository.SubscriptionRepository,
	subEvtRepo repository.SubscriptionEventRepository,
	pkgRepo repository.PackageRepository,
	outboxRepo repository.OutboxRepository,
	creditUC *CreditUsecase,
	uow *storage.UnitOfWork,
) *SubscriptionUsecase {
	return &SubscriptionUsecase{
		subRepo:    subRepo,
		subEvtRepo: subEvtRepo,
		pkgRepo:    pkgRepo,
		outboxRepo: outboxRepo,
		creditUC:   creditUC,
		uow:        uow,
	}
}

// SetPaymentUsecase injects the payment usecase (avoids circular dep).
func (uc *SubscriptionUsecase) SetPaymentUsecase(p *PaymentUsecase) {
	uc.paymentUC = p
}

// Subscribe creates a pending subscription and a payment transaction for it.
func (uc *SubscriptionUsecase) Subscribe(
	ctx context.Context,
	userID uuid.UUID,
	packageCode, billingCycle, paymentMethod, ipAddr string,
) (*model.Subscription, string, error) {
	// Check for existing active subscription.
	existing, err := uc.subRepo.GetActiveByUserID(ctx, userID)
	if err == nil && existing != nil {
		return nil, "", apperrors.ErrSubscriptionActive
	}

	pkg, err := uc.pkgRepo.GetByCode(ctx, packageCode)
	if err != nil {
		return nil, "", err
	}

	price := business.GetPriceForCycle(pkg, billingCycle)

	sub := &model.Subscription{
		ID:           uuid.New(),
		UserID:       userID,
		PackageID:    pkg.ID,
		BillingCycle: billingCycle,
		PricePaid:    price,
		Status:       domain.SubStatusPending,
		AutoRenew:    true,
	}

	var paymentURL string
	txErr := uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		if err := uc.subRepo.Create(ctx, tx, sub); err != nil {
			return err
		}
		// Record state transition.
		evt := &model.SubscriptionEvent{
			ID:             uuid.New(),
			SubscriptionID: sub.ID,
			ToStatus:       domain.SubStatusPending,
			ActorType:      domain.ActorUser,
			ActorID:        &userID,
		}
		return uc.subEvtRepo.Create(ctx, tx, evt)
	})
	if txErr != nil {
		return nil, "", txErr
	}

	// Create payment (outside inner tx, returns URL).
	if uc.paymentUC != nil && price > 0 {
		_, url, err := uc.paymentUC.CreatePayment(ctx, userID, price, paymentMethod,
			domain.PurposeSubscription, &sub.ID, ipAddr)
		if err != nil {
			return nil, "", err
		}
		paymentURL = url
	}

	return sub, paymentURL, nil
}

// GetCurrentSubscription returns the active subscription for a user.
func (uc *SubscriptionUsecase) GetCurrentSubscription(ctx context.Context, userID uuid.UUID) (*model.Subscription, error) {
	return uc.subRepo.GetActiveByUserID(ctx, userID)
}

// CancelSubscription sets auto_renew=false on the active subscription.
func (uc *SubscriptionUsecase) CancelSubscription(ctx context.Context, userID uuid.UUID) error {
	sub, err := uc.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}

	now := time.Now()
	reason := "user_cancelled"
	sub.AutoRenew = false
	sub.CancelledAt = &now
	sub.CancelReason = &reason

	return uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		if err := uc.subRepo.Update(ctx, tx, sub); err != nil {
			return err
		}
		evt := &model.SubscriptionEvent{
			ID:             uuid.New(),
			SubscriptionID: sub.ID,
			FromStatus:     strPtr(sub.Status),
			ToStatus:       sub.Status, // Status does not change, only auto_renew.
			Reason:         &reason,
			ActorType:      domain.ActorUser,
			ActorID:        &userID,
		}
		return uc.subEvtRepo.Create(ctx, tx, evt)
	})
}

// ReenableAutoRenew enables auto-renewal on the active subscription.
func (uc *SubscriptionUsecase) ReenableAutoRenew(ctx context.Context, userID uuid.UUID) error {
	sub, err := uc.subRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return err
	}
	sub.AutoRenew = true
	sub.CancelledAt = nil
	sub.CancelReason = nil
	return uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		return uc.subRepo.Update(ctx, tx, sub)
	})
}

// ActivateSubscription transitions a pending subscription to active.
func (uc *SubscriptionUsecase) ActivateSubscription(ctx context.Context, tx *sqlx.Tx, subID, paymentTxID uuid.UUID) error {
	sub, err := uc.subRepo.GetByID(ctx, subID)
	if err != nil {
		return err
	}

	if err := statemachine.ValidateSubscriptionTransition(sub.Status, domain.SubStatusActive); err != nil {
		return err
	}

	pkg, err := uc.pkgRepo.GetByID(ctx, sub.PackageID)
	if err != nil {
		return err
	}

	now := time.Now()
	endDate := business.CalculateEndDate(now, sub.BillingCycle)
	sub.Status = domain.SubStatusActive
	sub.StartDate = &now
	sub.EndDate = &endDate
	sub.PaymentTxID = &paymentTxID
	sub.CreditsAllocated = pkg.CreditsPerMonth

	if err := uc.subRepo.Update(ctx, tx, sub); err != nil {
		return err
	}

	// Allocate credits to user.
	if pkg.CreditsPerMonth > 0 {
		_, err = uc.creditUC.CreditWallet(ctx, tx, sub.UserID, pkg.CreditsPerMonth,
			domain.CreditTxSubscriptionAlloc, "subscription", &subID,
			"Monthly credit allocation for "+pkg.Name)
		if err != nil {
			return err
		}
	}

	// Record state transition event.
	fromStatus := "pending"
	evt := &model.SubscriptionEvent{
		ID:             uuid.New(),
		SubscriptionID: subID,
		FromStatus:     &fromStatus,
		ToStatus:       domain.SubStatusActive,
		ActorType:      domain.ActorSystem,
	}
	if err := uc.subEvtRepo.Create(ctx, tx, evt); err != nil {
		return err
	}

	// Outbox event.
	payload, _ := json.Marshal(map[string]interface{}{
		"subscription_id": subID,
		"user_id":         sub.UserID,
		"package_code":    pkg.Code,
		"end_date":        endDate,
	})
	outboxEvt := &model.OutboxEvent{
		AggregateType: "subscription",
		AggregateID:   subID.String(),
		EventType:     domain.EventSubscriptionActivated,
		Payload:       payload,
	}
	return uc.outboxRepo.Create(ctx, tx, outboxEvt)
}

// UpgradeSubscription cancels current subscription and starts a new one.
func (uc *SubscriptionUsecase) UpgradeSubscription(
	ctx context.Context,
	userID uuid.UUID,
	newPackageCode, billingCycle, paymentMethod, ipAddr string,
) (*model.Subscription, string, error) {
	// Cancel existing if any.
	existing, err := uc.subRepo.GetActiveByUserID(ctx, userID)
	if err == nil && existing != nil {
		_ = uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
			existing.Status = domain.SubStatusCancelled
			now := time.Now()
			existing.CancelledAt = &now
			reason := "upgrade"
			existing.CancelReason = &reason
			return uc.subRepo.Update(ctx, tx, existing)
		})
	}
	return uc.Subscribe(ctx, userID, newPackageCode, billingCycle, paymentMethod, ipAddr)
}

// ProcessExpiredSubscriptions marks active subscriptions past end_date as expired.
func (uc *SubscriptionUsecase) ProcessExpiredSubscriptions(ctx context.Context) error {
	subs, err := uc.subRepo.ListExpiringWithAutoRenew(ctx)
	if err != nil {
		return err
	}
	for _, sub := range subs {
		if sub.EndDate != nil && time.Now().After(*sub.EndDate) {
			_ = uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
				sub.Status = domain.SubStatusExpired
				if err := uc.subRepo.Update(ctx, tx, sub); err != nil {
					return err
				}
				payload, _ := json.Marshal(map[string]interface{}{
					"subscription_id": sub.ID,
					"user_id":         sub.UserID,
				})
				outboxEvt := &model.OutboxEvent{
					AggregateType: "subscription",
					AggregateID:   sub.ID.String(),
					EventType:     domain.EventSubscriptionExpired,
					Payload:       payload,
				}
				return uc.outboxRepo.Create(ctx, tx, outboxEvt)
			})
		}
	}
	return nil
}

// GetHistory returns all subscriptions for a user.
func (uc *SubscriptionUsecase) GetHistory(ctx context.Context, userID uuid.UUID) ([]*model.Subscription, error) {
	return uc.subRepo.ListByUserID(ctx, userID)
}

// ListPackages returns all active packages.
func (uc *SubscriptionUsecase) ListPackages(ctx context.Context) ([]*model.Package, error) {
	return uc.pkgRepo.List(ctx)
}

// GetPackageByCode returns a package by its code.
func (uc *SubscriptionUsecase) GetPackageByCode(ctx context.Context, code string) (*model.Package, error) {
	return uc.pkgRepo.GetByCode(ctx, code)
}

func strPtr(s string) *string { return &s }
