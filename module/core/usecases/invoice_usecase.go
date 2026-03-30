package usecases

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/modami/be-payment-service/module/core/domain"
	"github.com/modami/be-payment-service/module/core/model"
	"github.com/modami/be-payment-service/module/core/repository"
	"github.com/modami/be-payment-service/module/core/storage"
)

// InvoiceUsecase handles invoice operations.
type InvoiceUsecase struct {
	invoiceRepo repository.InvoiceRepository
	uow         *storage.UnitOfWork
}

func NewInvoiceUsecase(invoiceRepo repository.InvoiceRepository, uow *storage.UnitOfWork) *InvoiceUsecase {
	return &InvoiceUsecase{invoiceRepo: invoiceRepo, uow: uow}
}

// CreateInvoice generates an invoice for a payment transaction.
func (uc *InvoiceUsecase) CreateInvoice(
	ctx context.Context,
	tx *sqlx.Tx,
	userID uuid.UUID,
	paymentTxID *uuid.UUID,
	subscriptionID *uuid.UUID,
	amount int64,
	description string,
) (*model.Invoice, error) {
	invoiceNumber, err := uc.invoiceRepo.NextInvoiceNumber(ctx)
	if err != nil {
		return nil, err
	}

	invoice := &model.Invoice{
		ID:            uuid.New(),
		InvoiceNumber: invoiceNumber,
		UserID:        userID,
		PaymentTxID:   paymentTxID,
		SubscriptionID: subscriptionID,
		Subtotal:      amount,
		TaxAmount:     0,
		Total:         amount,
		Status:        domain.InvoiceStatusIssued,
	}
	if description != "" {
		invoice.Description = &description
	}

	if err := uc.invoiceRepo.Create(ctx, tx, invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

// GetInvoice returns an invoice by ID.
func (uc *InvoiceUsecase) GetInvoice(ctx context.Context, id uuid.UUID) (*model.Invoice, error) {
	return uc.invoiceRepo.GetByID(ctx, id)
}

// ListByUser returns all invoices for a user.
func (uc *InvoiceUsecase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Invoice, error) {
	return uc.invoiceRepo.ListByUserID(ctx, userID)
}
