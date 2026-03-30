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
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// CreditUsecase handles credit wallet business logic.
type CreditUsecase struct {
	walletRepo  repository.CreditWalletRepository
	txRepo      repository.CreditTransactionRepository
	outboxRepo  repository.OutboxRepository
	uow         *storage.UnitOfWork
}

func NewCreditUsecase(
	walletRepo repository.CreditWalletRepository,
	txRepo repository.CreditTransactionRepository,
	outboxRepo repository.OutboxRepository,
	uow *storage.UnitOfWork,
) *CreditUsecase {
	return &CreditUsecase{
		walletRepo: walletRepo,
		txRepo:     txRepo,
		outboxRepo: outboxRepo,
		uow:        uow,
	}
}

// GetBalance returns the credit wallet for a user, creating it if it does not exist.
func (uc *CreditUsecase) GetBalance(ctx context.Context, userID uuid.UUID) (*model.CreditWallet, error) {
	wallet, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err == apperrors.ErrWalletNotFound {
		// Auto-create wallet for new users.
		wallet = &model.CreditWallet{UserID: userID}
		if createErr := uc.walletRepo.Create(ctx, wallet); createErr != nil {
			return nil, createErr
		}
		return wallet, nil
	}
	return wallet, err
}

// GetTransactionHistory returns paginated credit transactions for a user.
func (uc *CreditUsecase) GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int, txType *string) ([]*model.CreditTransaction, int, error) {
	return uc.txRepo.ListByUserID(ctx, userID, limit, offset, txType)
}

// DeductCredit performs an atomic credit deduction inside the provided transaction.
// It uses SELECT FOR UPDATE to prevent race conditions.
func (uc *CreditUsecase) DeductCredit(
	ctx context.Context,
	tx *sqlx.Tx,
	userID uuid.UUID,
	amount int,
	txType, refType string,
	refID *uuid.UUID,
	description, idempotencyKey string,
) (*model.CreditTransaction, error) {
	// Check idempotency.
	if idempotencyKey != "" {
		existing, err := uc.txRepo.GetByIdempotencyKey(ctx, idempotencyKey)
		if err == nil && existing != nil {
			return existing, nil
		}
	}

	wallet, err := uc.walletRepo.GetByUserIDForUpdate(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if err := business.ValidateDeductCredit(wallet.Balance, amount); err != nil {
		return nil, err
	}

	newBalance := wallet.Balance - amount
	if err := uc.walletRepo.UpdateBalance(ctx, tx, userID, newBalance, wallet.Version); err != nil {
		return nil, err
	}

	var ikey *string
	if idempotencyKey != "" {
		ikey = &idempotencyKey
	}
	var rt *string
	if refType != "" {
		rt = &refType
	}

	creditTx := &model.CreditTransaction{
		ID:             uuid.New(),
		UserID:         userID,
		Amount:         -amount,
		Type:           txType,
		RefType:        rt,
		RefID:          refID,
		BalanceAfter:   newBalance,
		Description:    description,
		IdempotencyKey: ikey,
	}
	if err := uc.txRepo.Create(ctx, tx, creditTx); err != nil {
		return nil, err
	}

	// Create outbox event.
	payload, _ := json.Marshal(map[string]interface{}{
		"user_id":       userID,
		"amount":        amount,
		"balance_after": newBalance,
		"tx_id":         creditTx.ID,
	})
	outbox := &model.OutboxEvent{
		AggregateType: "credit_wallet",
		AggregateID:   userID.String(),
		EventType:     domain.EventCreditDeducted,
		Payload:       payload,
	}
	_ = uc.outboxRepo.Create(ctx, tx, outbox)

	return creditTx, nil
}

// CreditWallet credits a user's wallet inside the provided transaction.
func (uc *CreditUsecase) CreditWallet(
	ctx context.Context,
	tx *sqlx.Tx,
	userID uuid.UUID,
	amount int,
	txType, refType string,
	refID *uuid.UUID,
	description string,
) (*model.CreditTransaction, error) {
	if err := business.ValidateCreditAmount(amount); err != nil {
		return nil, err
	}

	wallet, err := uc.walletRepo.GetByUserIDForUpdate(ctx, tx, userID)
	if err == apperrors.ErrWalletNotFound {
		// Create wallet first, then get for update.
		newWallet := &model.CreditWallet{UserID: userID}
		if createErr := uc.walletRepo.Create(ctx, newWallet); createErr != nil {
			return nil, createErr
		}
		wallet, err = uc.walletRepo.GetByUserIDForUpdate(ctx, tx, userID)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	newBalance := wallet.Balance + amount
	if err := uc.walletRepo.UpdateBalance(ctx, tx, userID, newBalance, wallet.Version); err != nil {
		return nil, err
	}

	var rt *string
	if refType != "" {
		rt = &refType
	}

	creditTx := &model.CreditTransaction{
		ID:           uuid.New(),
		UserID:       userID,
		Amount:       amount,
		Type:         txType,
		RefType:      rt,
		RefID:        refID,
		BalanceAfter: newBalance,
		Description:  description,
	}
	if err := uc.txRepo.Create(ctx, tx, creditTx); err != nil {
		return nil, err
	}

	// Create outbox event.
	payload, _ := json.Marshal(map[string]interface{}{
		"user_id":       userID,
		"amount":        amount,
		"balance_after": newBalance,
		"tx_id":         creditTx.ID,
	})
	outbox := &model.OutboxEvent{
		AggregateType: "credit_wallet",
		AggregateID:   userID.String(),
		EventType:     domain.EventCreditAdded,
		Payload:       payload,
	}
	_ = uc.outboxRepo.Create(ctx, tx, outbox)

	return creditTx, nil
}

// AdminAdjust allows an admin to manually adjust a user's credit balance.
func (uc *CreditUsecase) AdminAdjust(ctx context.Context, userID uuid.UUID, amount int, description string) error {
	return uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		if amount > 0 {
			_, err := uc.CreditWallet(ctx, tx, userID, amount, domain.CreditTxAdminAdjust, "", nil, description)
			return err
		}
		_, err := uc.DeductCredit(ctx, tx, userID, -amount, domain.CreditTxAdminAdjust, "", nil, description, "")
		return err
	})
}

// EnsureWalletExists creates a wallet for a user if it does not already exist.
func (uc *CreditUsecase) EnsureWalletExists(ctx context.Context, userID uuid.UUID) error {
	_, err := uc.walletRepo.GetByUserID(ctx, userID)
	if err == apperrors.ErrWalletNotFound {
		wallet := &model.CreditWallet{
			UserID:    userID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		return uc.walletRepo.Create(ctx, wallet)
	}
	return err
}
