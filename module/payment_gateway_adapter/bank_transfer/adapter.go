package bank_transfer

import (
	"context"
	"fmt"

	gw "github.com/modami/be-payment-service/module/payment_gateway_adapter"
)

// Config holds bank transfer details.
type Config struct {
	BankName      string
	AccountNumber string
	AccountName   string
	Branch        string
}

// Adapter implements PaymentGateway for manual bank transfer.
type Adapter struct {
	cfg Config
}

// New creates a new BankTransfer Adapter.
func New(cfg Config) *Adapter {
	return &Adapter{cfg: cfg}
}

func (a *Adapter) Name() string { return "bank_transfer" }

// CreatePaymentURL returns bank account details as a pseudo-URL / instruction string.
func (a *Adapter) CreatePaymentURL(ctx context.Context, req gw.CreatePaymentRequest) (*gw.PaymentURL, error) {
	instructions := fmt.Sprintf(
		"bank://?bank=%s&account=%s&name=%s&branch=%s&amount=%d&ref=%s",
		a.cfg.BankName,
		a.cfg.AccountNumber,
		a.cfg.AccountName,
		a.cfg.Branch,
		req.Amount,
		req.OrderRef,
	)
	return &gw.PaymentURL{URL: instructions, ExpireAt: 0}, nil
}

// VerifyCallback for bank transfer always returns pending — manual admin approval required.
func (a *Adapter) VerifyCallback(ctx context.Context, params map[string]string) (*gw.PaymentResult, error) {
	return &gw.PaymentResult{
		Success:     false,
		GatewayTxID: params["ref"],
		RawResponse: params,
		FailureReason: "pending_manual_verification",
	}, nil
}

// QueryTransaction returns the status as stored (no live gateway to query).
func (a *Adapter) QueryTransaction(ctx context.Context, txRef string) (*gw.TransactionStatus, error) {
	return &gw.TransactionStatus{
		Status:      "pending",
		GatewayTxID: txRef,
	}, nil
}

// Refund marks the refund for manual processing.
func (a *Adapter) Refund(ctx context.Context, req gw.RefundRequest) (*gw.RefundResult, error) {
	return &gw.RefundResult{
		Success:    false,
		FailureMsg: "manual_refund_required",
	}, nil
}
