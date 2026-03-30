package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/modami/be-payment-service/module/core/domain"
	"github.com/modami/be-payment-service/module/core/model"
	"github.com/modami/be-payment-service/module/core/repository"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// RefundUsecase handles refund operations.
type RefundUsecase struct {
	refundRepo  repository.RefundRepository
	paymentRepo repository.PaymentTransactionRepository
}

func NewRefundUsecase(
	refundRepo repository.RefundRepository,
	paymentRepo repository.PaymentTransactionRepository,
) *RefundUsecase {
	return &RefundUsecase{
		refundRepo:  refundRepo,
		paymentRepo: paymentRepo,
	}
}

// RequestRefund creates a refund request.
func (uc *RefundUsecase) RequestRefund(
	ctx context.Context,
	userID uuid.UUID,
	paymentTxIDStr, creditTxIDStr, refundType, reason string,
	amount int64,
) (*model.Refund, error) {
	refund := &model.Refund{
		ID:          uuid.New(),
		UserID:      userID,
		RefundType:  refundType,
		Amount:      amount,
		Status:      domain.RefundStatusPending,
		RequestedBy: userID,
	}

	if reason != "" {
		refund.Reason = &reason
	}
	if paymentTxIDStr != "" {
		id, err := uuid.Parse(paymentTxIDStr)
		if err != nil {
			return nil, &apperrors.AppError{Code: "INVALID_UUID", Message: "Invalid payment_tx_id", HTTP: 400}
		}
		// Validate payment exists.
		pt, err := uc.paymentRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if pt.Status != domain.PaymentStatusSuccess {
			return nil, apperrors.ErrRefundNotEligible
		}
		refund.PaymentTxID = &id
	}
	if creditTxIDStr != "" {
		id, err := uuid.Parse(creditTxIDStr)
		if err != nil {
			return nil, &apperrors.AppError{Code: "INVALID_UUID", Message: "Invalid credit_tx_id", HTTP: 400}
		}
		refund.CreditTxID = &id
	}

	if err := uc.refundRepo.Create(ctx, refund); err != nil {
		return nil, err
	}
	return refund, nil
}

// GetRefund returns a refund by ID.
func (uc *RefundUsecase) GetRefund(ctx context.Context, id uuid.UUID) (*model.Refund, error) {
	return uc.refundRepo.GetByID(ctx, id)
}

// ListByUser returns all refunds for a user.
func (uc *RefundUsecase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Refund, error) {
	return uc.refundRepo.ListByUserID(ctx, userID)
}

// ListAll returns all refunds (admin).
func (uc *RefundUsecase) ListAll(ctx context.Context, limit, offset int) ([]*model.Refund, int, error) {
	return uc.refundRepo.ListAll(ctx, limit, offset)
}

// ApproveRefund transitions a refund to approved.
func (uc *RefundUsecase) ApproveRefund(ctx context.Context, refundID, adminID uuid.UUID) error {
	return uc.refundRepo.UpdateStatus(ctx, refundID, domain.RefundStatusApproved, &adminID)
}

// RejectRefund transitions a refund to rejected.
func (uc *RefundUsecase) RejectRefund(ctx context.Context, refundID, adminID uuid.UUID) error {
	return uc.refundRepo.UpdateStatus(ctx, refundID, domain.RefundStatusRejected, &adminID)
}
