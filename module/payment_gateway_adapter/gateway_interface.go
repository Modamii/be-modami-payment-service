package payment_gateway_adapter

import "context"

// CreatePaymentRequest is the input for creating a payment URL.
type CreatePaymentRequest struct {
	OrderRef    string
	Amount      int64
	Currency    string
	Description string
	ReturnURL   string
	IPAddress   string
	UserID      string
	ExtraData   map[string]string
}

// PaymentURL is the result of CreatePaymentURL.
type PaymentURL struct {
	URL      string
	ExpireAt int64
}

// PaymentResult is the result of verifying a gateway callback.
type PaymentResult struct {
	Success       bool
	GatewayTxID   string
	Amount        int64
	FailureReason string
	RawResponse   map[string]string
}

// TransactionStatus represents the current status of a transaction as reported by the gateway.
type TransactionStatus struct {
	Status      string
	GatewayTxID string
	Amount      int64
	PaidAt      *int64
}

// RefundRequest is the input to initiate a refund via the gateway.
type RefundRequest struct {
	GatewayTxID string
	OrderRef    string
	Amount      int64
	Reason      string
}

// RefundResult is the result of a refund operation.
type RefundResult struct {
	Success    bool
	RefundID   string
	FailureMsg string
}

// PaymentGateway is the interface every payment gateway adapter must implement.
type PaymentGateway interface {
	Name() string
	CreatePaymentURL(ctx context.Context, req CreatePaymentRequest) (*PaymentURL, error)
	VerifyCallback(ctx context.Context, params map[string]string) (*PaymentResult, error)
	QueryTransaction(ctx context.Context, txRef string) (*TransactionStatus, error)
	Refund(ctx context.Context, req RefundRequest) (*RefundResult, error)
}
