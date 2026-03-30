package usecases

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/modami/be-payment-service/module/core/domain"
	"github.com/modami/be-payment-service/module/core/model"
	"github.com/modami/be-payment-service/module/core/repository"
	"github.com/modami/be-payment-service/module/core/storage"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// UnlockUsecase handles contact unlock operations.
type UnlockUsecase struct {
	unlockRepo  repository.ContactUnlockRepository
	creditUC    *CreditUsecase
	outboxRepo  repository.OutboxRepository
	uow         *storage.UnitOfWork
}

func NewUnlockUsecase(
	unlockRepo repository.ContactUnlockRepository,
	creditUC *CreditUsecase,
	outboxRepo repository.OutboxRepository,
	uow *storage.UnitOfWork,
) *UnlockUsecase {
	return &UnlockUsecase{
		unlockRepo: unlockRepo,
		creditUC:   creditUC,
		outboxRepo: outboxRepo,
		uow:        uow,
	}
}

// UnlockContact deducts 1 credit and creates a contact_unlock record atomically.
func (uc *UnlockUsecase) UnlockContact(
	ctx context.Context,
	buyerID, productID, sellerID uuid.UUID,
	idempotencyKey string,
) (*model.ContactUnlock, error) {
	// Cannot unlock own product.
	if buyerID == sellerID {
		return nil, apperrors.ErrCannotUnlockOwn
	}

	// Check if already unlocked (outside transaction – cheaper).
	existing, err := uc.unlockRepo.GetByBuyerAndProduct(ctx, buyerID, productID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, apperrors.ErrAlreadyUnlocked
	}

	var unlock *model.ContactUnlock
	txErr := uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		// Deduct 1 credit.
		creditTx, err := uc.creditUC.DeductCredit(
			ctx, tx, buyerID,
			domain.UnlockCost,
			domain.CreditTxUnlock,
			"product",
			&productID,
			"Contact unlock for product "+productID.String(),
			idempotencyKey,
		)
		if err != nil {
			return err
		}

		unlock = &model.ContactUnlock{
			ID:         uuid.New(),
			BuyerID:    buyerID,
			ProductID:  productID,
			SellerID:   sellerID,
			CreditTxID: &creditTx.ID,
		}
		if err := uc.unlockRepo.Create(ctx, tx, unlock); err != nil {
			return err
		}

		// Outbox event.
		payload, _ := json.Marshal(map[string]interface{}{
			"buyer_id":   buyerID,
			"seller_id":  sellerID,
			"product_id": productID,
			"unlock_id":  unlock.ID,
		})
		outboxEvt := &model.OutboxEvent{
			AggregateType: "contact_unlock",
			AggregateID:   unlock.ID.String(),
			EventType:     domain.EventContactUnlocked,
			Payload:       payload,
		}
		return uc.outboxRepo.Create(ctx, tx, outboxEvt)
	})
	if txErr != nil {
		return nil, txErr
	}
	return unlock, nil
}

// CheckUnlock returns true if the buyer has already unlocked the product.
func (uc *UnlockUsecase) CheckUnlock(ctx context.Context, buyerID, productID uuid.UUID) (bool, error) {
	u, err := uc.unlockRepo.GetByBuyerAndProduct(ctx, buyerID, productID)
	if err != nil {
		return false, err
	}
	return u != nil, nil
}

// ListUnlocks returns a paginated list of unlocks for a buyer.
func (uc *UnlockUsecase) ListUnlocks(ctx context.Context, buyerID uuid.UUID, limit, offset int) ([]*model.ContactUnlock, int, error) {
	return uc.unlockRepo.ListByBuyerID(ctx, buyerID, limit, offset)
}
