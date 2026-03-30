package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/modami/be-payment-service/module/core/domain"
	"github.com/modami/be-payment-service/module/core/model"
	"github.com/modami/be-payment-service/module/core/repository"
	"github.com/modami/be-payment-service/module/core/storage"
	gw "github.com/modami/be-payment-service/module/payment_gateway_adapter"
	apperrors "github.com/modami/be-payment-service/pkg/errors"
)

// PaymentUsecase handles payment transaction business logic.
type PaymentUsecase struct {
	paymentRepo repository.PaymentTransactionRepository
	outboxRepo  repository.OutboxRepository
	gwSelector  *gw.GatewaySelector
	uow         *storage.UnitOfWork
	subUC       *SubscriptionUsecase
	creditUC    *CreditUsecase
}

func NewPaymentUsecase(
	paymentRepo repository.PaymentTransactionRepository,
	outboxRepo repository.OutboxRepository,
	gwSelector *gw.GatewaySelector,
	uow *storage.UnitOfWork,
) *PaymentUsecase {
	return &PaymentUsecase{
		paymentRepo: paymentRepo,
		outboxRepo:  outboxRepo,
		gwSelector:  gwSelector,
		uow:         uow,
	}
}

// SetSubscriptionUsecase injects the subscription usecase (avoids circular dep).
func (uc *PaymentUsecase) SetSubscriptionUsecase(sub *SubscriptionUsecase) {
	uc.subUC = sub
}

// SetCreditUsecase injects the credit usecase.
func (uc *PaymentUsecase) SetCreditUsecase(credit *CreditUsecase) {
	uc.creditUC = credit
}

// CreatePayment creates a payment transaction and returns the payment URL.
func (uc *PaymentUsecase) CreatePayment(
	ctx context.Context,
	userID uuid.UUID,
	amount int64,
	method, purpose string,
	purposeRefID *uuid.UUID,
	ipAddr string,
) (*model.PaymentTransaction, string, error) {
	gateway, err := uc.gwSelector.Select(method)
	if err != nil {
		return nil, "", err
	}

	orderRef := fmt.Sprintf("PAY-%d-%s", time.Now().UnixMilli(), userID.String()[:8])
	expiresAt := time.Now().Add(domain.PaymentExpiryMinutes * time.Minute)

	pt := &model.PaymentTransaction{
		ID:          uuid.New(),
		UserID:      userID,
		OrderRef:    orderRef,
		Amount:      amount,
		Currency:    "VND",
		Method:      method,
		Purpose:     purpose,
		PurposeRefID: purposeRefID,
		Status:      domain.PaymentStatusPending,
		IPAddress:   &ipAddr,
		ExpiresAt:   &expiresAt,
	}

	var paymentURL string
	txErr := uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		if err := uc.paymentRepo.Create(ctx, tx, pt); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, "", txErr
	}

	// Call gateway to get the payment URL.
	returnURLBase := "http://localhost:8080/api/v1/payments/return/"
	gwReq := gw.CreatePaymentRequest{
		OrderRef:    orderRef,
		Amount:      amount,
		Currency:    "VND",
		Description: purpose,
		ReturnURL:   returnURLBase + method,
		IPAddress:   ipAddr,
		UserID:      userID.String(),
	}
	result, err := gateway.CreatePaymentURL(ctx, gwReq)
	if err != nil {
		return nil, "", apperrors.ErrGatewayUnavailable
	}
	paymentURL = result.URL
	pt.PaymentURL = &paymentURL

	// Update payment_url in DB.
	_ = uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		_, dbErr := tx.ExecContext(ctx,
			`UPDATE payment_transactions SET payment_url = $1 WHERE id = $2`,
			paymentURL, pt.ID)
		return dbErr
	})

	return pt, paymentURL, nil
}

// GetPayment returns a payment transaction by ID.
func (uc *PaymentUsecase) GetPayment(ctx context.Context, id uuid.UUID) (*model.PaymentTransaction, error) {
	return uc.paymentRepo.GetByID(ctx, id)
}

// GetPaymentStatus returns the status of a payment, querying the gateway if still pending.
func (uc *PaymentUsecase) GetPaymentStatus(ctx context.Context, id uuid.UUID) (string, error) {
	pt, err := uc.paymentRepo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}
	if pt.Status != domain.PaymentStatusPending {
		return pt.Status, nil
	}

	// Query gateway for live status.
	gateway, err := uc.gwSelector.Select(pt.Method)
	if err != nil {
		return pt.Status, nil // Return cached status on gateway error.
	}
	txStatus, err := gateway.QueryTransaction(ctx, pt.OrderRef)
	if err != nil {
		return pt.Status, nil
	}

	if txStatus.Status == "success" {
		_ = uc.paymentRepo.UpdateStatus(ctx, nil, id, domain.PaymentStatusSuccess, txStatus.GatewayTxID, nil)
		return domain.PaymentStatusSuccess, nil
	}
	return pt.Status, nil
}

// HandleVNPayCallback processes a VNPay IPN/return callback.
func (uc *PaymentUsecase) HandleVNPayCallback(ctx context.Context, params map[string]string) error {
	return uc.handleCallback(ctx, domain.PaymentMethodVNPay, params)
}

// HandleMoMoCallback processes a MoMo IPN callback.
func (uc *PaymentUsecase) HandleMoMoCallback(ctx context.Context, params map[string]string) error {
	return uc.handleCallback(ctx, domain.PaymentMethodMoMo, params)
}

// HandleZaloPayCallback processes a ZaloPay callback.
func (uc *PaymentUsecase) HandleZaloPayCallback(ctx context.Context, params map[string]string) error {
	return uc.handleCallback(ctx, domain.PaymentMethodZaloPay, params)
}

func (uc *PaymentUsecase) handleCallback(ctx context.Context, method string, params map[string]string) error {
	gateway, err := uc.gwSelector.Select(method)
	if err != nil {
		return err
	}

	result, err := gateway.VerifyCallback(ctx, params)
	if err != nil {
		return apperrors.ErrInvalidSignature
	}

	// Determine order ref.
	orderRef := params["vnp_TxnRef"]
	if orderRef == "" {
		orderRef = params["orderId"]
	}
	if orderRef == "" {
		orderRef = params["app_trans_id"]
	}

	pt, err := uc.paymentRepo.GetByOrderRef(ctx, orderRef)
	if err != nil {
		return err
	}

	if pt.Status == domain.PaymentStatusSuccess {
		return nil // Idempotent.
	}

	rawResp, _ := json.Marshal(result.RawResponse)
	status := domain.PaymentStatusFailed
	if result.Success {
		status = domain.PaymentStatusSuccess
	}

	return uc.uow.RunInTx(ctx, func(tx *sqlx.Tx) error {
		if err := uc.paymentRepo.UpdateStatus(ctx, tx, pt.ID, status, result.GatewayTxID, rawResp); err != nil {
			return err
		}
		if result.Success {
			return uc.PostPaymentSuccess(ctx, tx, pt.ID)
		}
		// Emit failure event.
		payload, _ := json.Marshal(map[string]interface{}{
			"payment_tx_id": pt.ID,
			"order_ref":     orderRef,
			"reason":        result.FailureReason,
		})
		outboxEvt := &model.OutboxEvent{
			AggregateType: "payment_transaction",
			AggregateID:   pt.ID.String(),
			EventType:     domain.EventPaymentFailed,
			Payload:       payload,
		}
		return uc.outboxRepo.Create(ctx, tx, outboxEvt)
	})
}

// PostPaymentSuccess is called after a payment succeeds to trigger downstream actions.
func (uc *PaymentUsecase) PostPaymentSuccess(ctx context.Context, tx *sqlx.Tx, paymentTxID uuid.UUID) error {
	pt, err := uc.paymentRepo.GetByID(ctx, paymentTxID)
	if err != nil {
		return err
	}

	switch pt.Purpose {
	case domain.PurposeSubscription:
		if uc.subUC != nil && pt.PurposeRefID != nil {
			if err := uc.subUC.ActivateSubscription(ctx, tx, *pt.PurposeRefID, paymentTxID); err != nil {
				return err
			}
		}
	case domain.PurposeCreditPurchase:
		// Credit purchase — amount in VND needs to map to credits.
		// Simple 1:1000 ratio: 100,000 VND = 100 credits.
		credits := int(pt.Amount / 1000)
		if uc.creditUC != nil {
			_, err := uc.creditUC.CreditWallet(ctx, tx, pt.UserID, credits,
				domain.CreditTxPurchase, "payment_transaction", &paymentTxID,
				fmt.Sprintf("Credit purchase - %d credits", credits))
			if err != nil {
				return err
			}
		}
	}

	// Emit success event.
	payload, _ := json.Marshal(map[string]interface{}{
		"payment_tx_id": paymentTxID,
		"purpose":       pt.Purpose,
		"amount":        pt.Amount,
	})
	outboxEvt := &model.OutboxEvent{
		AggregateType: "payment_transaction",
		AggregateID:   paymentTxID.String(),
		EventType:     domain.EventPaymentSuccess,
		Payload:       payload,
	}
	return uc.outboxRepo.Create(ctx, tx, outboxEvt)
}

// GetHistory returns paginated payment transactions for a user.
func (uc *PaymentUsecase) GetHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*model.PaymentTransaction, int, error) {
	return uc.paymentRepo.ListByUserID(ctx, userID, limit, offset)
}
